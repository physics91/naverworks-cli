package api

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestClient_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_Post_WithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json, got %q", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(201)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	resp, err := client.Post("/test", []byte(`{"text":"hello"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	resp, err := client.Put("/test", []byte(`{"name":"updated"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	resp, err := client.Patch("/test", []byte(`{"name":"patched"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	resp, err := client.Delete("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 204 {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestClient_401_Retry(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count == 1 {
			w.WriteHeader(401)
			w.Write([]byte(`{"code":"UNAUTHORIZED"}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "old-token", RefreshToken: "rt", ExpiresAt: time.Now().Add(1 * time.Hour)}
	refreshCalled := false
	refreshFn := func(t *auth.Token) error {
		refreshCalled = true
		t.AccessToken = "new-token"
		return nil
	}
	client := NewClient(server.URL, token, refreshFn)
	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}
	if !refreshCalled {
		t.Error("expected refresh to be called")
	}
}

func TestClient_429_Backoff(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count <= 2 {
			w.Header().Set("RateLimit-Reset", "0")
			w.WriteHeader(429)
			w.Write([]byte(`{"code":"RATE_LIMIT"}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after backoff, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"code":"INVALID","description":"bad request"}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	_, err := client.Get("/test")
	if err == nil {
		t.Fatal("expected error for 400")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Code != "INVALID" {
		t.Errorf("expected INVALID, got %q", apiErr.Code)
	}
}

func TestClient_Post_RetryPreservesBody(t *testing.T) {
	var callCount int32
	var lastBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		body := make([]byte, 1024)
		n, _ := r.Body.Read(body)
		lastBody = string(body[:n])
		if count == 1 {
			w.WriteHeader(401)
			w.Write([]byte(`{"code":"UNAUTHORIZED"}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, func(t *auth.Token) error {
		t.AccessToken = "new"
		return nil
	})
	_, err := client.Post("/test", []byte(`{"msg":"hello"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lastBody != `{"msg":"hello"}` {
		t.Errorf("body not preserved on retry, got %q", lastBody)
	}
}

func TestDecodeAPIError_ValidJSON(t *testing.T) {
	body := []byte(`{"code":"INVALID","description":"bad request"}`)
	apiErr := DecodeAPIError(400, body)
	if apiErr.Code != "INVALID" {
		t.Errorf("expected INVALID, got %q", apiErr.Code)
	}
	if apiErr.Description != "bad request" {
		t.Errorf("expected 'bad request', got %q", apiErr.Description)
	}
}

func TestDecodeAPIError_InvalidJSON(t *testing.T) {
	body := []byte(`not json`)
	apiErr := DecodeAPIError(500, body)
	if apiErr.StatusCode != 500 {
		t.Errorf("expected 500, got %d", apiErr.StatusCode)
	}
	if apiErr.Code != "UNKNOWN" {
		t.Errorf("expected UNKNOWN, got %q", apiErr.Code)
	}
}

func TestDecodeAPIError_EmptyBody(t *testing.T) {
	apiErr := DecodeAPIError(502, nil)
	if apiErr.Code != "UNKNOWN" {
		t.Errorf("expected UNKNOWN, got %q", apiErr.Code)
	}
}

func TestDecodeAPIError_OAuthStyle(t *testing.T) {
	body := []byte(`{"error":"invalid_grant","error_description":"token expired"}`)
	apiErr := DecodeAPIError(401, body)
	if apiErr.Code != "invalid_grant" {
		t.Errorf("expected invalid_grant, got %q", apiErr.Code)
	}
}

func TestGetDownloadURL_MissingLocation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(302)
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	_, err := client.GetDownloadURL("/test")
	if err == nil {
		t.Fatal("expected error for 302 without Location")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Code != "MISSING_REDIRECT_LOCATION" {
		t.Errorf("expected MISSING_REDIRECT_LOCATION, got %q", apiErr.Code)
	}
}

func TestClient_OversizedResponse_DoWithRetry(t *testing.T) {
	oversizedBody := strings.Repeat("x", maxAPIResponseSize+1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(oversizedBody))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	_, err := client.Get("/test")
	if err == nil {
		t.Fatal("expected error for oversized response")
	}
	expected := fmt.Sprintf("API 응답 크기 초과: > %d bytes", maxAPIResponseSize)
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestClient_OversizedResponse_GetDownloadURL(t *testing.T) {
	oversizedBody := strings.Repeat("x", maxAPIResponseSize+1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(oversizedBody))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	_, err := client.GetDownloadURL("/test")
	if err == nil {
		t.Fatal("expected error for oversized response")
	}
	expected := fmt.Sprintf("API 응답 크기 초과: > %d bytes", maxAPIResponseSize)
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestClient_GetDownloadURL_Success(t *testing.T) {
	expectedLocation := "https://cdn.example.com/file/download?token=abc123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Location", expectedLocation)
		w.WriteHeader(302)
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	url, err := client.GetDownloadURL("/files/download")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != expectedLocation {
		t.Errorf("expected %q, got %q", expectedLocation, url)
	}
}

func TestClient_GetDownloadURL_301(t *testing.T) {
	expectedLocation := "https://cdn.example.com/file/redirect"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", expectedLocation)
		w.WriteHeader(301)
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	url, err := client.GetDownloadURL("/files/redirect")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != expectedLocation {
		t.Errorf("expected %q, got %q", expectedLocation, url)
	}
}

func TestValidatePresignedUploadURL_AllowsKnownStorageHost(t *testing.T) {
	_, err := validatePresignedUploadURL("https://kr.object.ncloudstorage.com/upload?signature=test")
	if err != nil {
		t.Fatalf("known storage host should pass: %v", err)
	}
}

func TestValidatePresignedUploadURL_RejectsLoopbackHost(t *testing.T) {
	_, err := validatePresignedUploadURL("https://127.0.0.1:8443/upload")
	if err == nil {
		t.Fatal("loopback upload host should be rejected")
	}
	if !strings.Contains(err.Error(), "허용되지 않는 업로드 호스트") {
		t.Fatalf("expected disallowed upload host error, got: %v", err)
	}
}

func TestValidatePresignedUploadURL_RejectsUnknownHost(t *testing.T) {
	_, err := validatePresignedUploadURL("https://evil.example.com/upload")
	if err == nil {
		t.Fatal("unknown upload host should be rejected")
	}
	if !strings.Contains(err.Error(), "허용되지 않는 업로드 호스트") {
		t.Fatalf("expected disallowed upload host error, got: %v", err)
	}
}

func TestClient_UploadFile_RejectsLoopbackUploadURL(t *testing.T) {
	tmp, err := os.CreateTemp("", "upload-file-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString("hello"); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	client := NewClient("https://www.worksapis.com/v1.0", &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(time.Hour)}, nil)
	err = client.UploadFile("https://127.0.0.1:8443/upload", tmp.Name())
	if err == nil {
		t.Fatal("expected loopback upload URL to be rejected")
	}
	if !strings.Contains(err.Error(), "허용되지 않는 업로드 호스트") {
		t.Fatalf("expected disallowed upload host error, got: %v", err)
	}
}

func newUploadTestClient(t *testing.T, handler http.HandlerFunc) (*Client, string) {
	t.Helper()

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	t.Setenv("NW_UPLOAD_ALLOWED_HOSTS", "example.com")

	client := NewClient("https://www.worksapis.com/v1.0", &auth.Token{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil)

	transport := server.Client().Transport.(*http.Transport).Clone()
	dialer := &net.Dialer{}
	targetAddr := server.Listener.Addr().String()
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if addr == "example.com:443" {
			addr = targetAddr
		}
		return dialer.DialContext(ctx, network, addr)
	}
	client.uploadClient = &http.Client{
		Timeout:   10 * time.Minute,
		Transport: transport,
	}

	return client, "https://example.com/upload"
}

func TestClient_UploadFileFromOffset_SendsContentRange(t *testing.T) {
	var contentRange string
	var contentLength int64
	var body []byte
	client, uploadURL := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		contentRange = r.Header.Get("Content-Range")
		contentLength = r.ContentLength
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	tmp, err := os.CreateTemp("", "upload-offset-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString("hello world"); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	if err := client.UploadFileFromOffset(uploadURL, tmp.Name(), 6); err != nil {
		t.Fatalf("UploadFileFromOffset failed: %v", err)
	}

	if contentRange != "bytes 6-10/11" {
		t.Fatalf("expected Content-Range bytes 6-10/11, got %q", contentRange)
	}
	if contentLength != 5 {
		t.Fatalf("expected Content-Length 5, got %d", contentLength)
	}
	if string(body) != "world" {
		t.Fatalf("expected resumed body 'world', got %q", string(body))
	}
}

func TestClient_UploadFileFromOffset_ZeroOffsetOmitsContentRange(t *testing.T) {
	var contentRange string
	var contentLength int64
	var body []byte
	client, uploadURL := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		contentRange = r.Header.Get("Content-Range")
		contentLength = r.ContentLength
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})

	tmp, err := os.CreateTemp("", "upload-zero-offset-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString("hello"); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	if err := client.UploadFileFromOffset(uploadURL, tmp.Name(), 0); err != nil {
		t.Fatalf("UploadFileFromOffset failed: %v", err)
	}

	if contentRange != "" {
		t.Fatalf("expected empty Content-Range, got %q", contentRange)
	}
	if contentLength != 5 {
		t.Fatalf("expected Content-Length 5, got %d", contentLength)
	}
	if string(body) != "hello" {
		t.Fatalf("expected full body 'hello', got %q", string(body))
	}
}

func TestClient_UploadFileFromOffset_RejectsInvalidOffset(t *testing.T) {
	tmp, err := os.CreateTemp("", "upload-invalid-offset-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString("hello"); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		offset int64
	}{
		{name: "negative", offset: -1},
		{name: "equal to file size", offset: 5},
		{name: "greater than file size", offset: 6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var requestCount int32
			client, uploadURL := newUploadTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&requestCount, 1)
				w.WriteHeader(http.StatusOK)
			})

			err := client.UploadFileFromOffset(uploadURL, tmp.Name(), tc.offset)
			if err == nil {
				t.Fatal("expected invalid offset error")
			}
			if !strings.Contains(err.Error(), "offset") {
				t.Fatalf("expected offset error, got %v", err)
			}
			if atomic.LoadInt32(&requestCount) != 0 {
				t.Fatalf("expected no upload request, got %d", requestCount)
			}
		})
	}
}

func TestClient_UploadMultipart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// Verify Content-Type starts with multipart/form-data
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "multipart/form-data") {
			t.Errorf("expected Content-Type starting with multipart/form-data, got %q", ct)
		}
		// Verify Authorization header
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %q", r.Header.Get("Authorization"))
		}
		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("failed to parse multipart form: %v", err)
			w.WriteHeader(400)
			return
		}
		// Verify field name and file content
		file, header, err := r.FormFile("fileData")
		if err != nil {
			t.Errorf("failed to get form file 'fileData': %v", err)
			w.WriteHeader(400)
			return
		}
		defer file.Close()
		if header.Filename != "test.txt" {
			t.Errorf("expected filename 'test.txt', got %q", header.Filename)
		}
		content := make([]byte, 1024)
		n, _ := file.Read(content)
		if string(content[:n]) != "hello multipart" {
			t.Errorf("expected file content 'hello multipart', got %q", string(content[:n]))
		}

		w.WriteHeader(200)
		w.Write([]byte(`{"fileId":"f1"}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	resp, err := client.UploadMultipart("/upload", "fileData", "test.txt", []byte("hello multipart"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if string(resp.Body) != `{"fileId":"f1"}` {
		t.Errorf("unexpected body: %s", string(resp.Body))
	}
}

func TestClient_DownloadFile(t *testing.T) {
	binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Disposition", `attachment; filename="test.png"`)
		w.WriteHeader(200)
		w.Write(binaryData)
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	data, headers, err := client.DownloadFile("/files/test.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != len(binaryData) {
		t.Errorf("expected %d bytes, got %d", len(binaryData), len(data))
	}
	for i, b := range binaryData {
		if data[i] != b {
			t.Errorf("byte %d: expected 0x%02X, got 0x%02X", i, b, data[i])
		}
	}
	if headers.Get("Content-Type") != "image/png" {
		t.Errorf("expected Content-Type image/png, got %q", headers.Get("Content-Type"))
	}
	if headers.Get("Content-Disposition") != `attachment; filename="test.png"` {
		t.Errorf("unexpected Content-Disposition: %q", headers.Get("Content-Disposition"))
	}
}

func TestClient_ExactMaxSizeResponse_Allowed(t *testing.T) {
	// A response exactly at maxAPIResponseSize should succeed (not exceed).
	exactBody := strings.Repeat("y", maxAPIResponseSize)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(exactBody))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error for exact-max-size response: %v", err)
	}
	if len(resp.Body) != maxAPIResponseSize {
		t.Errorf("expected body length %d, got %d", maxAPIResponseSize, len(resp.Body))
	}
}
