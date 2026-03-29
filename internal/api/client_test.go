package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
