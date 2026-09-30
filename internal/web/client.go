package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

const taskOrigin = "https://task.worksmobile.com"
const maxResponseSize = 8 << 20

type Client struct {
	ctx     context.Context
	close   func()
	room    Room
	project string
	userID  string
}

// Open reuses only this CLI profile's session. Cookies stay inside Chrome;
// credentials are never copied into Go requests or an API token store.
func (b Browser) Open(ctx context.Context) (*Client, error) {
	if _, err := os.Stat(b.SessionDir); err != nil {
		return nil, loginRequired()
	}
	executable, err := browserExecutable(b.ExecPath)
	if err != nil {
		return nil, err
	}
	if err := prepareSession(b.SessionDir, b.Profile); err != nil {
		return nil, err
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	ctx, timeout := context.WithTimeout(ctx, 2*time.Minute)
	allocator, cancelAllocator := chromedp.NewExecAllocator(ctx, b.options(executable, true)...)
	page, cancelPage := chromedp.NewContext(allocator)
	c := &Client{ctx: page}
	c.close = func() {
		_ = chromedp.Cancel(page)
		cancelPage()
		cancelAllocator()
		timeout()
		stop()
	}
	if err := chromedp.Run(page, chromedp.Navigate(EntryURL)); err != nil {
		c.Close()
		return nil, browserStartError(executable, err)
	}
	readyCtx, cancelReady := context.WithTimeout(page, 12*time.Second)
	defer cancelReady()
	if err := chromedp.Run(readyCtx, chromedp.Poll(authenticatedPage, nil)); err != nil {
		c.Close()
		return nil, loginRequired()
	}
	return c, nil
}

func loginRequired() error {
	return fmt.Errorf("브라우저 세션이 없거나 만료됐습니다. naverworks auth login --method browser로 로그인하세요")
}

func (c *Client) Close() { c.close() }

// request is deliberately not exported: domain methods choose fixed service
// hosts and paths. Redirects and automatic retries are forbidden, including
// after an ambiguous write failure.
func (c *Client) request(origin, method, path, contentType, body string) ([]byte, error) {
	if err := validateRequest(origin, method, path); err != nil {
		return nil, err
	}
	input, err := json.Marshal(map[string]any{
		"origin": origin, "method": method, "path": path, "contentType": contentType, "body": body,
		"headers": c.requestHeaders(origin),
	})
	if err != nil {
		return nil, err
	}
	script := `(async () => {
const p = ` + string(input) + `;
if (location.origin !== p.origin) return {error:'origin'};
try {
 const opts = {method:p.method, credentials:'same-origin', redirect:'error', cache:'no-store', headers:p.headers};
 if (p.body) { opts.body=p.body; opts.headers['Content-Type']=p.contentType; }
 // Match the service's Axios CSRF convention, without exporting the token.
 const csrf = document.cookie.split('; ').find(s=>s.startsWith('XSRF-TOKEN='));
 if (csrf) opts.headers['X-XSRF-TOKEN']=decodeURIComponent(csrf.slice('XSRF-TOKEN='.length));
 const r = await fetch(p.path, opts);
 if (r.status === 401 || r.status === 403) return {status:r.status};
 if (!r.body) return {status:r.status, type:r.headers.get('content-type')||'', serviceError:!!r.headers.get('task-error'), text:''};
 const reader = r.body.getReader(), decoder = new TextDecoder();
 let text = '', size = 0;
 for (;;) {
  const part = await reader.read(); if (part.done) break;
  size += part.value.byteLength;
  if (size > ` + fmt.Sprint(maxResponseSize) + `) { await reader.cancel(); return {error:'size'}; }
  text += decoder.decode(part.value, {stream:true});
 }
 text += decoder.decode();
 return {status:r.status, type:r.headers.get('content-type')||'', serviceError:!!r.headers.get('task-error'), text};
} catch (_) { return {error:'network'}; }
})()`
	var result struct {
		Status       int
		Type         string
		Text         string
		Error        string
		ServiceError bool
	}
	err = chromedp.Run(c.ctx, chromedp.Evaluate(script, &result, awaitPromise))
	if err != nil || result.Error != "" {
		if result.Error == "size" {
			return nil, fmt.Errorf("웹 응답 크기가 제한을 초과했습니다")
		}
		return nil, fmt.Errorf("웹 요청 결과를 확인하지 못했습니다. 변경 요청은 재실행 전에 웹에서 반영 여부를 확인하세요")
	}
	if result.Status == 401 || result.Status == 403 {
		return nil, fmt.Errorf("웹 접근이 거부됐습니다 (HTTP %d). 방 접근 권한을 확인하고 필요하면 naverworks auth login --method browser로 다시 로그인하세요", result.Status)
	}
	if result.Status < 200 || result.Status >= 300 {
		return nil, fmt.Errorf("웹 요청 실패 (HTTP %d)", result.Status)
	}
	if result.ServiceError {
		return nil, fmt.Errorf("할 일 서비스가 요청 오류를 반환했습니다. 변경 요청은 웹에서 반영 여부를 확인하세요")
	}
	if strings.TrimSpace(result.Text) != "" && !strings.Contains(result.Type, "json") && !strings.HasPrefix(result.Type, "text/plain") && result.Status != 204 {
		return nil, fmt.Errorf("예상하지 못한 웹 응답입니다. 브라우저 로그인 상태와 서비스 변경 여부를 확인하세요")
	}
	return []byte(result.Text), nil
}

func (c *Client) requestHeaders(origin string) map[string]string {
	headers := map[string]string{"Accept": "application/json, text/plain, */*"}
	if origin == taskOrigin {
		// These are the task web client's required version and identity headers.
		// UserInfo bootstraps the identity; it never comes from command input.
		headers["WM-TASK-CLIENT-VERSION"] = "4.0.0"
		headers["WM-TASK-SERVICE-VERSION"] = "2"
		headers["WM-TASK-CLIENT-TYPE"] = "WEB_TALK"
		if positiveNumber(c.userID) {
			headers["task-web-client-user-id"] = c.userID
		}
	}
	return headers
}

func awaitPromise(p *runtime.EvaluateParams) *runtime.EvaluateParams {
	return p.WithAwaitPromise(true).WithReturnByValue(true)
}

func validateRequest(origin, method, path string) error {
	if origin != strings.TrimSuffix(EntryURL, "/") && origin != taskOrigin {
		return fmt.Errorf("허용되지 않은 웹 서비스입니다")
	}
	u, err := url.Parse(path)
	if err != nil || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || u.IsAbs() || u.Host != "" || u.Fragment != "" || strings.ContainsAny(path, "\\\r\n") {
		return fmt.Errorf("유효하지 않은 웹 요청 경로입니다")
	}
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
		return nil
	default:
		return fmt.Errorf("허용되지 않은 웹 요청 방식입니다")
	}
}

func (c *Client) get(origin, path string, query url.Values) ([]byte, error) {
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.request(origin, "GET", path, "", "")
}

func (c *Client) postJSON(origin, path string, body any) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.request(origin, "POST", path, "application/json", string(data))
}

func responseJSON(body []byte, out any) error {
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("웹 응답 형식이 변경됐습니다")
	}
	return nil
}
