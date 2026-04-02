package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

type Response struct {
	StatusCode int
	Body       []byte
}

const maxRateLimitRetries = 3
const maxAPIResponseSize = 10 << 20 // 10MB

type RefreshFunc func(token *auth.Token) error

type Client struct {
	baseURL          string
	token            *auth.Token
	refreshFn        RefreshFunc
	httpClient       *http.Client
	noRedirectClient *http.Client
}

func NewClient(baseURL string, token *auth.Token, refreshFn RefreshFunc) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		refreshFn:  refreshFn,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		noRedirectClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (c *Client) Get(path string) (*Response, error) {
	return c.do("GET", path, nil)
}

func (c *Client) Post(path string, body []byte) (*Response, error) {
	return c.do("POST", path, body)
}

func (c *Client) Put(path string, body []byte) (*Response, error) {
	return c.do("PUT", path, body)
}

func (c *Client) Patch(path string, body []byte) (*Response, error) {
	return c.do("PATCH", path, body)
}

func (c *Client) Delete(path string) (*Response, error) {
	return c.do("DELETE", path, nil)
}

func (c *Client) refreshIfNeeded() error {
	if c.token.NeedsRefresh() && c.refreshFn != nil {
		if err := c.refreshFn(c.token); err != nil {
			return fmt.Errorf("토큰 갱신 실패: %w", err)
		}
	}
	return nil
}

func (c *Client) do(method, path string, body []byte) (*Response, error) {
	if err := c.refreshIfNeeded(); err != nil {
		return nil, err
	}
	return c.doWithRetry(method, path, body, false)
}

func (c *Client) doWithRetry(method, path string, body []byte, retried401 bool) (*Response, error) {
	return c.doWithRetryAndMaxResponseSize(method, path, body, retried401, maxAPIResponseSize)
}

func (c *Client) GetWithMaxResponseSize(path string, maxResponseSize int64) (*Response, error) {
	if err := c.refreshIfNeeded(); err != nil {
		return nil, err
	}
	return c.doWithRetryAndMaxResponseSize("GET", path, nil, false, maxResponseSize)
}

func (c *Client) doWithRetryAndMaxResponseSize(method, path string, body []byte, retried401 bool, maxResponseSize int64) (*Response, error) {
	for attempt := 0; attempt <= maxRateLimitRetries; attempt++ {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("요청 생성 실패: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("네트워크 에러: %w", err)
		}
		respBody, err := readResponseBodyWithLimit(resp.Body, maxResponseSize)
		if err != nil {
			return nil, err
		}

		switch {
		case resp.StatusCode == 401 && !retried401 && c.refreshFn != nil:
			if err := c.refreshFn(c.token); err != nil {
				return nil, fmt.Errorf("토큰 갱신 실패: %w", err)
			}
			return c.doWithRetryAndMaxResponseSize(method, path, body, true, maxResponseSize)

		case resp.StatusCode == 429 && attempt < maxRateLimitRetries:
			waitDuration := parseRateLimitReset(resp.Header, attempt)
			time.Sleep(waitDuration)
			continue

		case resp.StatusCode >= 400:
			return nil, DecodeAPIError(resp.StatusCode, respBody)

		default:
			return &Response{StatusCode: resp.StatusCode, Body: respBody}, nil
		}
	}

	return nil, &APIError{StatusCode: 429, Code: "RATE_LIMIT_EXCEEDED", Description: "최대 재시도 횟수 초과"}
}

func readResponseBodyWithLimit(body io.ReadCloser, maxResponseSize int64) ([]byte, error) {
	respBody, err := io.ReadAll(io.LimitReader(body, maxResponseSize+1))
	body.Close()
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}
	if int64(len(respBody)) > maxResponseSize {
		return nil, fmt.Errorf("API 응답 크기 초과: > %d bytes", maxResponseSize)
	}
	return respBody, nil
}

// GetDownloadURL calls the API endpoint without following redirects,
// returning the Location header URL for download endpoints that return 302.
// Includes token refresh and 401/429 retry logic (same as doWithRetry).
func (c *Client) GetDownloadURL(path string) (string, error) {
	if err := c.refreshIfNeeded(); err != nil {
		return "", err
	}
	return c.getDownloadURLWithRetry(path, false)
}

func (c *Client) getDownloadURLWithRetry(path string, retried401 bool) (string, error) {
	for attempt := 0; attempt <= maxRateLimitRetries; attempt++ {
		req, err := http.NewRequest("GET", c.baseURL+path, nil)
		if err != nil {
			return "", fmt.Errorf("요청 생성 실패: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)

		resp, err := c.noRedirectClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("네트워크 에러: %w", err)
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIResponseSize+1))
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("응답 읽기 실패: %w", err)
		}
		if int64(len(body)) > maxAPIResponseSize {
			return "", fmt.Errorf("API 응답 크기 초과: > %d bytes", maxAPIResponseSize)
		}

		switch {
		case resp.StatusCode == 301 || resp.StatusCode == 302:
			location := resp.Header.Get("Location")
			if location != "" {
				return location, nil
			}
			return "", &APIError{
				StatusCode:  resp.StatusCode,
				Code:        "MISSING_REDIRECT_LOCATION",
				Description: "리다이렉트 응답에 Location 헤더가 없습니다",
			}

		case resp.StatusCode == 401 && !retried401 && c.refreshFn != nil:
			if err := c.refreshFn(c.token); err != nil {
				return "", fmt.Errorf("토큰 갱신 실패: %w", err)
			}
			return c.getDownloadURLWithRetry(path, true)

		case resp.StatusCode == 429 && attempt < maxRateLimitRetries:
			waitDuration := parseRateLimitReset(resp.Header, attempt)
			time.Sleep(waitDuration)
			continue

		case resp.StatusCode >= 400:
			return "", DecodeAPIError(resp.StatusCode, body)

		default:
			var result struct {
				DownloadURL string `json:"downloadUrl"`
			}
			if json.Unmarshal(body, &result) == nil && result.DownloadURL != "" {
				return result.DownloadURL, nil
			}
			return string(body), nil
		}
	}

	return "", &APIError{StatusCode: 429, Code: "RATE_LIMIT_EXCEEDED", Description: "최대 재시도 횟수 초과"}
}

// UploadFile uploads a file to a pre-signed URL using PUT.
// No Authorization header is sent (the URL is pre-signed).
func (c *Client) UploadFile(uploadURL string, filePath string) error {
	parsedURL, err := url.Parse(uploadURL)
	if err != nil {
		return fmt.Errorf("업로드 URL 파싱 실패: %w", err)
	}
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("업로드 URL이 HTTPS가 아닙니다: %s", parsedURL.Scheme)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("파일 정보 조회 실패: %w", err)
	}

	req, err := http.NewRequest("PUT", uploadURL, f)
	if err != nil {
		return fmt.Errorf("업로드 요청 생성 실패: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = stat.Size()

	uploadClient := &http.Client{Timeout: 10 * time.Minute}
	resp, err := uploadClient.Do(req)
	if err != nil {
		return fmt.Errorf("업로드 네트워크 에러: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("업로드 실패 (HTTP %d): %s", resp.StatusCode, string(body))
	}
	return nil
}

func marshalBody(body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("요청 데이터 직렬화 실패: %w", err)
	}
	return data, nil
}

func (c *Client) doJSON(method, path string, body interface{}) (*Response, error) {
	data, err := marshalBody(body)
	if err != nil {
		return nil, err
	}
	return c.do(method, path, data)
}

func (c *Client) PostJSON(path string, body interface{}) (*Response, error) {
	return c.doJSON("POST", path, body)
}

func (c *Client) PutJSON(path string, body interface{}) (*Response, error) {
	return c.doJSON("PUT", path, body)
}

func (c *Client) PatchJSON(path string, body interface{}) (*Response, error) {
	return c.doJSON("PATCH", path, body)
}

func parseRateLimitReset(header http.Header, attempt int) time.Duration {
	for _, key := range []string{"RateLimit-Reset", "X-RateLimit-Reset"} {
		if val := header.Get(key); val != "" {
			if seconds, err := strconv.Atoi(val); err == nil && seconds > 0 {
				return time.Duration(seconds) * time.Second
			}
		}
	}
	return time.Duration(1<<uint(attempt)) * time.Second
}

// UploadMultipart sends a multipart/form-data POST request to the given API
// path. fieldName is the form field name, fileName is the original file name,
// and data is the file content. Auth header and token refresh / retry logic
// are handled identically to other Client methods.
func (c *Client) UploadMultipart(path, fieldName, fileName string, data []byte) (*Response, error) {
	if err := c.refreshIfNeeded(); err != nil {
		return nil, err
	}
	return c.uploadMultipartWithRetry(path, fieldName, fileName, data, false)
}

func (c *Client) uploadMultipartWithRetry(path, fieldName, fileName string, data []byte, retried401 bool) (*Response, error) {
	for attempt := 0; attempt <= maxRateLimitRetries; attempt++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile(fieldName, fileName)
		if err != nil {
			return nil, fmt.Errorf("multipart 파트 생성 실패: %w", err)
		}
		if _, err := part.Write(data); err != nil {
			return nil, fmt.Errorf("multipart 파트 쓰기 실패: %w", err)
		}
		if err := writer.Close(); err != nil {
			return nil, fmt.Errorf("multipart writer 닫기 실패: %w", err)
		}

		req, err := http.NewRequest("POST", c.baseURL+path, body)
		if err != nil {
			return nil, fmt.Errorf("요청 생성 실패: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("네트워크 에러: %w", err)
		}
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIResponseSize+1))
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("응답 읽기 실패: %w", err)
		}
		if int64(len(respBody)) > maxAPIResponseSize {
			return nil, fmt.Errorf("API 응답 크기 초과: > %d bytes", maxAPIResponseSize)
		}

		switch {
		case resp.StatusCode == 401 && !retried401 && c.refreshFn != nil:
			if err := c.refreshFn(c.token); err != nil {
				return nil, fmt.Errorf("토큰 갱신 실패: %w", err)
			}
			return c.uploadMultipartWithRetry(path, fieldName, fileName, data, true)

		case resp.StatusCode == 429 && attempt < maxRateLimitRetries:
			waitDuration := parseRateLimitReset(resp.Header, attempt)
			time.Sleep(waitDuration)
			continue

		case resp.StatusCode >= 400:
			return nil, DecodeAPIError(resp.StatusCode, respBody)

		default:
			return &Response{StatusCode: resp.StatusCode, Body: respBody}, nil
		}
	}

	return nil, &APIError{StatusCode: 429, Code: "RATE_LIMIT_EXCEEDED", Description: "최대 재시도 횟수 초과"}
}

// DownloadFile performs a GET request and returns the raw response body along
// with HTTP headers. This is useful for binary file downloads where the
// caller needs access to Content-Type, Content-Disposition, etc.
func (c *Client) DownloadFile(path string) ([]byte, http.Header, error) {
	if err := c.refreshIfNeeded(); err != nil {
		return nil, nil, err
	}
	return c.downloadFileWithRetry(path, false)
}

func (c *Client) downloadFileWithRetry(path string, retried401 bool) ([]byte, http.Header, error) {
	for attempt := 0; attempt <= maxRateLimitRetries; attempt++ {
		req, err := http.NewRequest("GET", c.baseURL+path, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("요청 생성 실패: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("네트워크 에러: %w", err)
		}
		respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxAPIResponseSize+1))
		resp.Body.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("응답 읽기 실패: %w", err)
		}
		if int64(len(respBody)) > maxAPIResponseSize {
			return nil, nil, fmt.Errorf("API 응답 크기 초과: > %d bytes", maxAPIResponseSize)
		}

		switch {
		case resp.StatusCode == 401 && !retried401 && c.refreshFn != nil:
			if err := c.refreshFn(c.token); err != nil {
				return nil, nil, fmt.Errorf("토큰 갱신 실패: %w", err)
			}
			return c.downloadFileWithRetry(path, true)

		case resp.StatusCode == 429 && attempt < maxRateLimitRetries:
			waitDuration := parseRateLimitReset(resp.Header, attempt)
			time.Sleep(waitDuration)
			continue

		case resp.StatusCode >= 400:
			return nil, nil, DecodeAPIError(resp.StatusCode, respBody)

		default:
			return respBody, resp.Header, nil
		}
	}

	return nil, nil, &APIError{StatusCode: 429, Code: "RATE_LIMIT_EXCEEDED", Description: "최대 재시도 횟수 초과"}
}
