package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/authdoctor"
	"github.com/physics91/naverworks-cli/internal/config"
	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
)

func TestAuthDoctor_LocalOnlyUsesEnvFallback(t *testing.T) {
	h := clitest.NewHarness(t)
	t.Setenv("NW_PROFILE", "ci")
	t.Setenv("NW_CLIENT_ID", "env-client")
	t.Setenv("NW_CLIENT_SECRET", "env-secret")

	result, err := h.Run([]string{"auth", "doctor"}, newRootCommandRunner(t))
	if err != nil {
		t.Fatalf("auth doctor failed: %v", err)
	}
	if result.Stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.Stderr)
	}
	if got := len(h.RequestLogs()); got != 0 {
		t.Fatalf("request logs = %d, want 0", got)
	}

	payload := decodeDoctorResult(t, result.Stdout)
	if payload.Profile.Selected != "ci" {
		t.Fatalf("profile.selected = %q, want %q", payload.Profile.Selected, "ci")
	}
	if payload.Profile.Source != "env" {
		t.Fatalf("profile.source = %q, want %q", payload.Profile.Source, "env")
	}
	if payload.VerificationScope != authdoctor.VerificationScopeLocalOnly {
		t.Fatalf("verification_scope = %q, want %q", payload.VerificationScope, authdoctor.VerificationScopeLocalOnly)
	}

	checks := indexDoctorChecks(payload.Checks)
	if got := checks["config.client_credentials"].Status; got != authdoctor.StatusPass {
		t.Fatalf("config.client_credentials status = %q, want %q", got, authdoctor.StatusPass)
	}
	if got := checks["token.present"].Status; got != authdoctor.StatusFail {
		t.Fatalf("token.present status = %q, want %q", got, authdoctor.StatusFail)
	}
}

func TestAuthDoctorVerifyRemoteDoesNotMutateToken(t *testing.T) {
	h := clitest.NewHarness(t)
	writeDoctorProfileConfig(t, h.HomeDir(), &config.ProfileConfig{
		CurrentProfile: "default",
		Profiles: map[string]*config.Config{
			"default": {
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				Scope:        "openid profile directory",
			},
		},
	})
	tokenPath := tokenPathForHome(h.HomeDir())
	if err := os.MkdirAll(filepath.Dir(tokenPath), 0700); err != nil {
		t.Fatalf("mkdir token dir failed: %v", err)
	}
	tokenBytes := []byte(`{"auth_method":"oauth","access_token":"access-token","token_type":"Bearer","expires_at":"2099-01-01T00:00:00Z","scope":"openid profile directory"}`)
	if err := os.WriteFile(tokenPath, tokenBytes, 0600); err != nil {
		t.Fatalf("write token failed: %v", err)
	}

	server := h.StartScriptedServer([]clitest.ResponseScript{
		{StatusCode: 200, Body: `{"users":[]}`},
	})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	result, err := h.Run([]string{"auth", "doctor", "--verify-remote"}, newRootCommandRunner(t))
	if err != nil {
		t.Fatalf("auth doctor --verify-remote failed: %v", err)
	}
	if result.Stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.Stderr)
	}

	payload := decodeDoctorResult(t, result.Stdout)
	if payload.VerificationScope != authdoctor.VerificationScopeRemoteOptIn {
		t.Fatalf("verification_scope = %q, want %q", payload.VerificationScope, authdoctor.VerificationScopeRemoteOptIn)
	}
	checks := indexDoctorChecks(payload.Checks)
	if got := checks["scope.directory_remote_probe"].Status; got != authdoctor.StatusPass {
		t.Fatalf("scope.directory_remote_probe status = %q, want %q", got, authdoctor.StatusPass)
	}

	after, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("read token after doctor failed: %v", err)
	}
	if string(after) != string(tokenBytes) {
		t.Fatalf("token file changed\nafter: %s\nwant: %s", string(after), string(tokenBytes))
	}
	if got := len(h.RequestLogs()); got != 1 {
		t.Fatalf("request log count = %d, want 1", got)
	}
	if got := h.RequestLogs()[0].Path; got != "/users" {
		t.Fatalf("request path = %q, want %q", got, "/users")
	}
}

func TestAuthDoctorVerifyRemoteChecksScimWhenConfigured(t *testing.T) {
	h := clitest.NewHarness(t)
	writeDoctorProfileConfig(t, h.HomeDir(), &config.ProfileConfig{
		CurrentProfile: "default",
		Profiles: map[string]*config.Config{
			"default": {
				ClientID:        "test-client",
				ClientSecret:    "test-secret",
				ScimAccessToken: "scim-token",
				Scope:           "openid profile directory",
			},
		},
	})
	writeDoctorToken(t, h.HomeDir(), &auth.Token{
		AuthMethod:  auth.AuthMethodOAuth,
		AccessToken: "access-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
		Scope:       "openid profile directory",
	})

	apiServer := h.StartScriptedServer([]clitest.ResponseScript{
		{StatusCode: 200, Body: `{"users":[]}`},
	})
	defer apiServer.Close()
	setAPIBaseURL(t, apiServer.URL)

	scimServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Users" {
			t.Fatalf("SCIM path = %q, want %q", r.URL.Path, "/Users")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"Resources":[]}`))
	}))
	defer scimServer.Close()
	setSCIMBaseURL(t, scimServer.URL)

	result, err := h.Run([]string{"auth", "doctor", "--verify-remote"}, newRootCommandRunner(t))
	if err != nil {
		t.Fatalf("auth doctor --verify-remote failed: %v", err)
	}
	payload := decodeDoctorResult(t, result.Stdout)
	checks := indexDoctorChecks(payload.Checks)
	if got := checks["scim.endpoint_reachable"].Status; got != authdoctor.StatusPass {
		t.Fatalf("scim.endpoint_reachable status = %q, want %q", got, authdoctor.StatusPass)
	}
}

func decodeDoctorResult(t *testing.T, stdout string) authdoctor.Result {
	t.Helper()

	var payload authdoctor.Result
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("doctor output is not valid JSON: %v\noutput: %q", err, stdout)
	}
	return payload
}

func indexDoctorChecks(checks []authdoctor.Check) map[string]authdoctor.Check {
	out := make(map[string]authdoctor.Check, len(checks))
	for _, check := range checks {
		out[check.CheckID] = check
	}
	return out
}

func writeDoctorProfileConfig(t *testing.T, homeDir string, pc *config.ProfileConfig) {
	t.Helper()

	cfgPath := filepath.Join(homeDir, ".config", "naverworks", "config.json")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
		t.Fatalf("mkdir config dir failed: %v", err)
	}
	data, err := json.Marshal(pc)
	if err != nil {
		t.Fatalf("marshal profile config failed: %v", err)
	}
	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		t.Fatalf("write profile config failed: %v", err)
	}
}

func writeDoctorToken(t *testing.T, homeDir string, token *auth.Token) {
	t.Helper()

	tokenPath := filepath.Join(homeDir, ".config", "naverworks", "token.json")
	if err := os.MkdirAll(filepath.Dir(tokenPath), 0700); err != nil {
		t.Fatalf("mkdir token dir failed: %v", err)
	}
	data, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("marshal token failed: %v", err)
	}
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatalf("write token failed: %v", err)
	}
}
