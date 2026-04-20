package auth

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBuildAuthorizationURL(t *testing.T) {
	u := BuildAuthorizationURL("https://auth.example.com", "client-id", "http://localhost:8484/callback", "test-state", "openid profile bot")
	if u == "" {
		t.Fatal("expected non-empty URL")
	}
	for _, param := range []string{"client_id=client-id", "state=test-state", "response_type=code"} {
		if !strings.Contains(u, param) {
			t.Errorf("URL missing %q: %s", param, u)
		}
	}
}

func TestExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "at-new",
			"refresh_token": "rt-new",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "openid profile bot",
		})
	}))
	defer server.Close()

	token, err := ExchangeCode(server.URL, "client-id", "client-secret", "auth-code", "http://localhost:8484/callback")
	if err != nil {
		t.Fatalf("exchange failed: %v", err)
	}
	if token.AccessToken != "at-new" {
		t.Errorf("expected at-new, got %q", token.AccessToken)
	}
	if token.AuthMethod != "oauth" {
		t.Errorf("expected oauth, got %q", token.AuthMethod)
	}
}

func TestExchangeCode_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_grant",
			"error_description": "bad code",
		})
	}))
	defer server.Close()

	_, err := ExchangeCode(server.URL, "client-id", "client-secret", "bad-code", "http://localhost:8484/callback")
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	if !strings.Contains(err.Error(), "invalid_grant") {
		t.Errorf("expected invalid_grant in error, got %v", err)
	}
}

func TestRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "at-refreshed",
			"refresh_token": "rt-refreshed",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "bot",
		})
	}))
	defer server.Close()

	token := &Token{RefreshToken: "rt-old", AuthMethod: "oauth"}
	err := RefreshAccessToken(server.URL, "client-id", "client-secret", token)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if token.AccessToken != "at-refreshed" {
		t.Errorf("expected at-refreshed, got %q", token.AccessToken)
	}
}

func TestRevokeToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	err := RevokeToken(server.URL, "client-id", "client-secret", "some-token", "access_token")
	if err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
}

func TestFindAvailableListener(t *testing.T) {
	ln, port, err := FindAvailableListener(8484, 8494)
	if err != nil {
		t.Fatalf("find listener failed: %v", err)
	}
	defer ln.Close()
	if port < 8484 || port > 8494 {
		t.Errorf("port %d out of range", port)
	}
}

func TestHasScope(t *testing.T) {
	if !HasScope("openid profile bot", "openid") {
		t.Error("expected openid to be found")
	}
	if HasScope("bot directory", "openid") {
		t.Error("expected openid to not be found")
	}
}

func TestGenerateState(t *testing.T) {
	state, err := GenerateState()
	if err != nil {
		t.Fatalf("generate state failed: %v", err)
	}
	if len(state) != 32 {
		t.Errorf("expected 32 hex chars, got %d", len(state))
	}
}

func TestWaitForCallback_IgnoresStateMismatchUntilValidCodeArrives(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	go func() {
		time.Sleep(50 * time.Millisecond)
		resp, _ := http.Get(fmt.Sprintf("http://127.0.0.1:%d/callback?state=wrong&code=test", port))
		if resp != nil {
			resp.Body.Close()
		}

		time.Sleep(50 * time.Millisecond)
		resp, _ = http.Get(fmt.Sprintf("http://127.0.0.1:%d/callback?state=expected-state&code=good", port))
		if resp != nil {
			resp.Body.Close()
		}
	}()

	code, err := WaitForCallback(ln, "expected-state", 5*time.Second)
	if err != nil {
		t.Fatalf("expected callback to continue waiting after state mismatch: %v", err)
	}
	if code != "good" {
		t.Fatalf("expected good code, got %q", code)
	}
}

func TestRequestToken_EmptyAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token_type": "Bearer",
			"expires_in": 3600,
		})
	}))
	defer server.Close()

	_, err := ExchangeCode(server.URL, "id", "secret", "code", "http://localhost/cb")
	if err == nil {
		t.Fatal("expected error for empty access_token")
	}
}
