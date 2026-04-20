package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/fileutil"
	"github.com/physics91/naverworks-cli/internal/httputil"
)

type Response struct {
	StatusCode int
	Body       []byte
}

const maxRateLimitRetries = 3
const maxAPIResponseSize = 10 << 20 // 10MB

var errRateLimitExceeded = &APIError{StatusCode: 429, Code: "RATE_LIMIT_EXCEEDED", Description: "최대 재시도 횟수 초과"}
var rateLimitSleep = time.Sleep

var defaultAllowedPresignedUploadHostSuffixes = []string{
	"worksapis.com",
	"worksmobile.com",
	"ncloudstorage.com",
	"amazonaws.com",
	"amazonaws.com.cn",
	"cloudfront.net",
}

type RefreshFunc func(token *auth.Token) error

type PreviewOptions struct {
	DryRun        bool
	PlanOutPath   string
	GenerateInput bool
	Profile       string
}

type Client struct {
	baseURL          string
	token            *auth.Token
	refreshFn        RefreshFunc
	httpClient       *http.Client
	noRedirectClient *http.Client
	uploadClient     *http.Client
	preview          *PreviewOptions
}

func NewClient(baseURL string, token *auth.Token, refreshFn RefreshFunc) *Client {
	transport := httputil.NewSecureTransport()
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		refreshFn:  refreshFn,
		httpClient: &http.Client{Timeout: 30 * time.Second, Transport: transport},
		noRedirectClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		uploadClient: &http.Client{Timeout: 10 * time.Minute, Transport: transport},
	}
}

func (c *Client) WithPreview(opts PreviewOptions) *Client {
	opts.PlanOutPath = strings.TrimSpace(opts.PlanOutPath)
	if opts.PlanOutPath != "" {
		opts.DryRun = true
	}
	if !opts.enabled() {
		c.preview = nil
		return c
	}
	copied := opts
	c.preview = &copied
	return c
}

func (c *Client) PreviewEnabled() bool {
	return c.preview != nil && c.preview.enabled()
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
	if resp, err, handled := c.previewJSONRequest(method, path, body); handled {
		return resp, err
	}
	if err := c.refreshIfNeeded(); err != nil {
		return nil, err
	}
	return c.doWithRetry(method, path, body, false)
}

func (c *Client) doWithRetry(method, path string, body []byte, retried401 bool) (*Response, error) {
	return c.doWithRetryAndMaxResponseSize(method, path, body, retried401, maxAPIResponseSize)
}

func (c *Client) GetWithMaxResponseSize(path string, maxResponseSize int64) (*Response, error) {
	if resp, err, handled := c.previewJSONRequest(http.MethodGet, path, nil); handled {
		return resp, err
	}
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

		case resp.StatusCode == 429:
			if attempt == maxRateLimitRetries {
				return nil, errRateLimitExceeded
			}
			waitDuration := parseRateLimitReset(resp.Header, attempt)
			rateLimitSleep(waitDuration)
			continue

		case resp.StatusCode >= 400:
			return nil, DecodeAPIError(resp.StatusCode, respBody)

		default:
			return &Response{StatusCode: resp.StatusCode, Body: respBody}, nil
		}
	}

	return nil, errRateLimitExceeded
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
	if resp, err, handled := c.previewJSONRequest(http.MethodGet, path, nil); handled {
		if err != nil {
			return "", err
		}
		return string(resp.Body), nil
	}
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
		body, err := readResponseBodyWithLimit(resp.Body, maxAPIResponseSize)
		if err != nil {
			return "", err
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

		case resp.StatusCode == 429:
			if attempt == maxRateLimitRetries {
				return "", errRateLimitExceeded
			}
			waitDuration := parseRateLimitReset(resp.Header, attempt)
			rateLimitSleep(waitDuration)
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

	return "", errRateLimitExceeded
}

// UploadFile uploads a file to a pre-signed URL using PUT.
// No Authorization header is sent (the URL is pre-signed).
func (c *Client) UploadFile(uploadURL string, filePath string) error {
	return c.UploadFileFromOffset(uploadURL, filePath, 0)
}

// UploadFileFromOffset uploads a file to a pre-signed URL starting at the given offset.
// When offset > 0, the request sends only the remaining bytes and includes Content-Range.
func (c *Client) UploadFileFromOffset(uploadURL string, filePath string, offset int64) error {
	if err, handled := c.previewUploadFromOffset(uploadURL, filePath, offset); handled {
		return err
	}
	if _, err := validatePresignedUploadURL(uploadURL); err != nil {
		return err
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("파일 정보 조회 실패: %w", err)
	}
	if !stat.Mode().IsRegular() {
		return fmt.Errorf("일반 파일만 허용합니다: %s", filePath)
	}
	size := stat.Size()
	if offset < 0 || offset > size || (size > 0 && offset == size) {
		return fmt.Errorf("유효하지 않은 업로드 offset: %d (file size: %d)", offset, size)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("파일 열기 실패: %w", err)
	}
	defer f.Close()
	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return fmt.Errorf("업로드 offset 이동 실패: %w", err)
		}
	}

	req, err := http.NewRequest("PUT", uploadURL, f)
	if err != nil {
		return fmt.Errorf("업로드 요청 생성 실패: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = size - offset
	if offset > 0 {
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", offset, size-1, size))
	}

	resp, err := c.uploadClient.Do(req)
	if err != nil {
		return fmt.Errorf("업로드 네트워크 에러: %w", err)
	}

	if resp.StatusCode >= 400 {
		body, err := readResponseBodyWithLimit(resp.Body, maxAPIResponseSize)
		if err != nil {
			return err
		}
		return fmt.Errorf("업로드 실패 (HTTP %d): %s", resp.StatusCode, string(body))
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
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
	if resp, err, handled := c.previewMultipart(path, fieldName, fileName, data); handled {
		return resp, err
	}
	if err := c.refreshIfNeeded(); err != nil {
		return nil, err
	}
	return c.uploadMultipartWithRetry(path, fieldName, fileName, data, false)
}

func (c *Client) uploadMultipartWithRetry(path, fieldName, fileName string, data []byte, retried401 bool) (*Response, error) {
	// Build multipart body once before the retry loop
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
	contentType := writer.FormDataContentType()
	bodyBytes := body.Bytes()

	for attempt := 0; attempt <= maxRateLimitRetries; attempt++ {
		req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("요청 생성 실패: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
		req.Header.Set("Content-Type", contentType)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("네트워크 에러: %w", err)
		}
		respBody, err := readResponseBodyWithLimit(resp.Body, maxAPIResponseSize)
		if err != nil {
			return nil, err
		}

		switch {
		case resp.StatusCode == 401 && !retried401 && c.refreshFn != nil:
			if err := c.refreshFn(c.token); err != nil {
				return nil, fmt.Errorf("토큰 갱신 실패: %w", err)
			}
			return c.uploadMultipartWithRetry(path, fieldName, fileName, data, true)

		case resp.StatusCode == 429:
			if attempt == maxRateLimitRetries {
				return nil, errRateLimitExceeded
			}
			waitDuration := parseRateLimitReset(resp.Header, attempt)
			rateLimitSleep(waitDuration)
			continue

		case resp.StatusCode >= 400:
			return nil, DecodeAPIError(resp.StatusCode, respBody)

		default:
			return &Response{StatusCode: resp.StatusCode, Body: respBody}, nil
		}
	}

	return nil, errRateLimitExceeded
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
		respBody, err := readResponseBodyWithLimit(resp.Body, maxAPIResponseSize)
		if err != nil {
			return nil, nil, err
		}

		switch {
		case resp.StatusCode == 401 && !retried401 && c.refreshFn != nil:
			if err := c.refreshFn(c.token); err != nil {
				return nil, nil, fmt.Errorf("토큰 갱신 실패: %w", err)
			}
			return c.downloadFileWithRetry(path, true)

		case resp.StatusCode == 429:
			if attempt == maxRateLimitRetries {
				return nil, nil, errRateLimitExceeded
			}
			waitDuration := parseRateLimitReset(resp.Header, attempt)
			rateLimitSleep(waitDuration)
			continue

		case resp.StatusCode >= 400:
			return nil, nil, DecodeAPIError(resp.StatusCode, respBody)

		default:
			return respBody, resp.Header, nil
		}
	}

	return nil, nil, errRateLimitExceeded
}

func validatePresignedUploadURL(uploadURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(uploadURL)
	if err != nil {
		return nil, fmt.Errorf("업로드 URL 파싱 실패: %w", err)
	}
	if parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("업로드 URL이 HTTPS가 아닙니다: %s", parsedURL.Scheme)
	}
	if parsedURL.Host == "" || parsedURL.Hostname() == "" {
		return nil, fmt.Errorf("업로드 URL 호스트가 비어 있습니다")
	}
	if parsedURL.User != nil {
		return nil, fmt.Errorf("업로드 URL에 사용자 정보가 포함되어 있습니다")
	}

	host := strings.ToLower(parsedURL.Hostname())
	if _, err := netip.ParseAddr(host); err == nil {
		return nil, fmt.Errorf("허용되지 않는 업로드 호스트: %s", host)
	}

	if !isAllowedPresignedUploadHost(host) {
		return nil, fmt.Errorf("허용되지 않는 업로드 호스트: %s", host)
	}

	return parsedURL, nil
}

func isAllowedPresignedUploadHost(host string) bool {
	for _, suffix := range allowedPresignedUploadHostSuffixes() {
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}
	return false
}

func allowedPresignedUploadHostSuffixes() []string {
	raw := os.Getenv("NW_UPLOAD_ALLOWED_HOSTS")
	if strings.TrimSpace(raw) == "" {
		return defaultAllowedPresignedUploadHostSuffixes
	}

	parts := strings.Split(raw, ",")
	allowed := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		allowed = append(allowed, part)
	}
	if len(allowed) == 0 {
		return defaultAllowedPresignedUploadHostSuffixes
	}
	return allowed
}

func (o PreviewOptions) enabled() bool {
	return o.DryRun || o.GenerateInput || strings.TrimSpace(o.PlanOutPath) != ""
}

func (c *Client) previewJSONRequest(method, path string, body []byte) (*Response, error, bool) {
	if !c.PreviewEnabled() {
		return nil, nil, false
	}
	plan, err := c.buildJSONPreviewPlan(method, path, body)
	if err != nil {
		return nil, err, true
	}
	resp, err := c.previewResponse(plan, body)
	return resp, err, true
}

func (c *Client) previewMultipart(path, fieldName, fileName string, data []byte) (*Response, error, bool) {
	if !c.PreviewEnabled() {
		return nil, nil, false
	}
	plan := map[string]interface{}{
		"dry_run":      true,
		"method":       http.MethodPost,
		"path":         path,
		"url":          c.baseURL + path,
		"content_type": "multipart/form-data",
		"multipart": map[string]interface{}{
			"field_name": fieldName,
			"file_name":  fileName,
			"file_size":  len(data),
		},
	}
	if profile := strings.TrimSpace(c.preview.Profile); profile != "" {
		plan["profile"] = profile
	}
	input := map[string]interface{}{
		"field_name": fieldName,
		"file_name":  fileName,
		"file_size":  len(data),
	}
	resp, err := c.previewResponse(plan, mustMarshalJSON(input))
	return resp, err, true
}

func (c *Client) previewUploadFromOffset(uploadURL, filePath string, offset int64) (error, bool) {
	if !c.PreviewEnabled() {
		return nil, false
	}
	plan := map[string]interface{}{
		"dry_run":      true,
		"method":       http.MethodPut,
		"url":          uploadURL,
		"content_type": "application/octet-stream",
		"upload": map[string]interface{}{
			"file_path": filePath,
			"offset":    offset,
		},
	}
	if profile := strings.TrimSpace(c.preview.Profile); profile != "" {
		plan["profile"] = profile
	}
	_, err := c.previewResponse(plan, mustMarshalJSON(plan["upload"]))
	return err, true
}

func (c *Client) previewResponse(plan map[string]interface{}, inputBody []byte) (*Response, error) {
	if err := c.writePreviewPlan(plan); err != nil {
		return nil, err
	}
	if c.preview != nil && c.preview.GenerateInput {
		return &Response{StatusCode: http.StatusOK, Body: previewInputBody(inputBody)}, nil
	}
	body, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("dry-run 결과 직렬화 실패: %w", err)
	}
	return &Response{StatusCode: http.StatusOK, Body: body}, nil
}

func (c *Client) buildJSONPreviewPlan(method, path string, body []byte) (map[string]interface{}, error) {
	plan := map[string]interface{}{
		"dry_run":      true,
		"method":       method,
		"path":         path,
		"url":          c.baseURL + path,
		"content_type": "application/json",
	}
	if profile := strings.TrimSpace(c.preview.Profile); profile != "" {
		plan["profile"] = profile
	}
	if len(body) == 0 {
		plan["body"] = map[string]interface{}{}
		return plan, nil
	}
	plan["body_size"] = len(body)
	if json.Valid(body) {
		var payload interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, fmt.Errorf("dry-run 본문 파싱 실패: %w", err)
		}
		plan["body"] = payload
		return plan, nil
	}
	plan["body_text"] = string(body)
	return plan, nil
}

func (c *Client) writePreviewPlan(plan map[string]interface{}) error {
	if c.preview == nil || strings.TrimSpace(c.preview.PlanOutPath) == "" {
		return nil
	}
	if err := fileutil.WriteSecureJSON(c.preview.PlanOutPath, plan); err != nil {
		return fmt.Errorf("plan 파일 저장 실패: %w", err)
	}
	return nil
}

func previewInputBody(body []byte) []byte {
	if len(body) == 0 {
		return []byte("{}")
	}
	if json.Valid(body) {
		var payload interface{}
		if err := json.Unmarshal(body, &payload); err == nil {
			if data, err := json.Marshal(payload); err == nil {
				return data
			}
		}
	}
	return mustMarshalJSON(map[string]interface{}{"input_text": string(body)})
}

func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return data
}
