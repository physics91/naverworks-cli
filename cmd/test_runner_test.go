package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
)

func TestSharedCLIRunnerVersion(t *testing.T) {
	h := clitest.NewHarness(t)

	result, err := h.Run([]string{"version"}, newRootCommandRunner(t))
	if err != nil {
		t.Fatalf("version run failed: %v", err)
	}
	if result.Stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.Stderr)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(result.Stdout)), &payload); err != nil {
		t.Fatalf("version output is not valid JSON: %v\noutput: %q", err, result.Stdout)
	}
	for _, key := range []string{"version", "commit", "build_date"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("version output missing %q", key)
		}
	}
}

func newRootCommandRunner(t *testing.T) func([]string) error {
	t.Helper()

	return func(args []string) error {
		resetCommandTreeFlags(rootCmd)
		rootCmd.SetArgs(args)
		defer func() {
			rootCmd.SetArgs(nil)
			resetCommandTreeFlags(rootCmd)
		}()
		return rootCmd.Execute()
	}
}

func TestBuildAPIClientUsesOverriddenBaseURL(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"users":[]}`))
	}))
	defer server.Close()

	setAPIBaseURL(t, server.URL)

	cfg := &config.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}
	token := &auth.Token{
		AuthMethod:  auth.AuthMethodJWT,
		AccessToken: "token",
		TokenType:   "Bearer",
		ExpiresAt:   mustParseRFC3339(t, "2099-01-01T00:00:00Z"),
	}

	client := buildAPIClient(cfg, token, "default")
	_, err := api.NewDirectoryService(client).ListUsers("", 20)
	if err != nil {
		t.Fatalf("directory list-users failed: %v", err)
	}
	if gotPath != "/users" {
		t.Fatalf("request path = %q, want %q", gotPath, "/users")
	}
}

func mustParseRFC3339(t *testing.T, value string) (ts time.Time) {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q failed: %v", value, err)
	}
	return parsed
}
