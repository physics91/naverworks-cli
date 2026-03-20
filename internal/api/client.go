package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
