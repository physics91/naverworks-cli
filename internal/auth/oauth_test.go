package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildAuthorizationURL(t *testing.T) {
	url := BuildAuthorizationURL("https://auth.example.com", "client-id", "http://localhost:8484/callback", "test-state", "openid profile bot")
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	for _, param := range []string{"client_id=client-id", "state=test-state", "response_type=code"} {
		if !strings.Contains(url, param) {
			t.Errorf("URL missing %q: %s", param, url)
		}
	}
}

func TestExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
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

func TestFindAvailablePort(t *testing.T) {
	port, err := FindAvailablePort(8484, 8494)
	if err != nil {
		t.Fatalf("find port failed: %v", err)
	}
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
