package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type RefreshFunc func(token *auth.Token) error

type Client struct {
	baseURL    string
	token      *auth.Token
	refreshFn  RefreshFunc
	httpClient *http.Client
}

func NewClient(baseURL string, token *auth.Token, refreshFn RefreshFunc) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		refreshFn:  refreshFn,
		httpClient: &http.Client{Timeout: 30 * time.Second},
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

func (c *Client) do(method, path string, body []byte) (*Response, error) {
	if c.token.NeedsRefresh() && c.refreshFn != nil {
		if err := c.refreshFn(c.token); err != nil {
			return nil, fmt.Errorf("토큰 갱신 실패: %w", err)
		}
	}
	return c.doWithRetry(method, path, body, false)
}

func (c *Client) doWithRetry(method, path string, body []byte, retried401 bool) (*Response, error) {
	const maxRateLimitRetries = 3

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
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		switch {
		case resp.StatusCode == 401 && !retried401 && c.refreshFn != nil:
			if err := c.refreshFn(c.token); err != nil {
				return nil, fmt.Errorf("토큰 갱신 실패: %w", err)
			}
			return c.doWithRetry(method, path, body, true)

		case resp.StatusCode == 429 && attempt < maxRateLimitRetries:
			waitDuration := parseRateLimitReset(resp.Header, attempt)
			time.Sleep(waitDuration)
			continue

		case resp.StatusCode >= 400:
			apiErr := &APIError{StatusCode: resp.StatusCode}
			json.Unmarshal(respBody, apiErr)
			return nil, apiErr

		default:
			return &Response{StatusCode: resp.StatusCode, Body: respBody}, nil
		}
	}

	return nil, &APIError{StatusCode: 429, Code: "RATE_LIMIT_EXCEEDED", Description: "최대 재시도 횟수 초과"}
}

// UploadFile uploads a file to a pre-signed URL using PUT.
// No Authorization header is sent (the URL is pre-signed).
func (c *Client) UploadFile(uploadURL string, filePath string) error {
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

	resp, err := http.DefaultClient.Do(req)
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
