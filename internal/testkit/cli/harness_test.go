package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestHarnessCreatesIsolatedEnv(t *testing.T) {
	h := NewHarness(t)

	tmpDir := h.HomeDir()
	if tmpDir == "" {
		t.Fatal("expected fake home directory")
	}
	if _, err := os.Stat(tmpDir); err != nil {
		t.Fatalf("expected fake home directory to exist: %v", err)
	}

	configDir := filepath.Join(tmpDir, ".config", "naverworks")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("failed to create config dir inside fake home: %v", err)
	}

	if got := os.Getenv("HOME"); got != tmpDir {
		t.Fatalf("HOME = %q, want %q", got, tmpDir)
	}
	if runtime.GOOS == "windows" {
		if got := os.Getenv("APPDATA"); got != tmpDir {
			t.Fatalf("APPDATA = %q, want %q", got, tmpDir)
		}
	}

	for _, key := range []string{
		"NW_PROFILE",
		"NW_CLIENT_ID",
		"NW_CLIENT_SECRET",
		"NW_SERVICE_ACCOUNT_ID",
		"NW_PRIVATE_KEY_PATH",
		"NW_DOMAIN_ID",
		"NW_BOT_ID",
		"NW_SCOPE",
		"NW_DEFAULT_CALENDAR_USER_ID",
		"NW_SCIM_ACCESS_TOKEN",
	} {
		if got := os.Getenv(key); got != "" {
			t.Fatalf("%s = %q, want empty string", key, got)
		}
	}

	result, err := h.Capture(func() error {
		fmt.Fprint(os.Stdout, "hello-stdout")
		fmt.Fprint(os.Stderr, "hello-stderr")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected capture error: %v", err)
	}
	if result.Stdout != "hello-stdout" {
		t.Fatalf("stdout = %q, want %q", result.Stdout, "hello-stdout")
	}
	if result.Stderr != "hello-stderr" {
		t.Fatalf("stderr = %q, want %q", result.Stderr, "hello-stderr")
	}
}

func TestHarnessRecordsRequests(t *testing.T) {
	h := NewHarness(t)
	server := h.StartScriptedServer([]ResponseScript{
		{
			StatusCode: http.StatusCreated,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"status":"created"}`,
		},
	})
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/users?cursor=next", strings.NewReader(`{"name":"tester"}`))
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
	if got := resp.StatusCode; got != http.StatusCreated {
		t.Fatalf("status = %d, want %d", got, http.StatusCreated)
	}
	if got := string(body); got != `{"status":"created"}` {
		t.Fatalf("body = %q, want %q", got, `{"status":"created"}`)
	}

	logs := h.RequestLogs()
	if len(logs) != 1 {
		t.Fatalf("request log count = %d, want 1", len(logs))
	}
	if logs[0].Method != http.MethodPost {
		t.Fatalf("method = %q, want %q", logs[0].Method, http.MethodPost)
	}
	if logs[0].Path != "/users" {
		t.Fatalf("path = %q, want %q", logs[0].Path, "/users")
	}
	if logs[0].RawQuery != "cursor=next" {
		t.Fatalf("raw query = %q, want %q", logs[0].RawQuery, "cursor=next")
	}
	if logs[0].Body != `{"name":"tester"}` {
		t.Fatalf("body = %q, want %q", logs[0].Body, `{"name":"tester"}`)
	}
	if logs[0].Headers["Content-Type"] != "application/json" {
		t.Fatalf("content-type = %q, want %q", logs[0].Headers["Content-Type"], "application/json")
	}
}

func TestHarnessFailureCategories(t *testing.T) {
	tests := []struct {
		name     string
		category FailureCategory
		message  string
		want     string
	}{
		{
			name:     "setup failure",
			category: SetupFailure,
			message:  "missing config fixture",
			want:     "SetupFailure: missing config fixture",
		},
		{
			name:     "request shape failure",
			category: RequestShapeFailure,
			message:  "unexpected request body",
			want:     "RequestShapeFailure: unexpected request body",
		},
		{
			name:     "response handling failure",
			category: ResponseHandlingFailure,
			message:  "missing pagination cursor",
			want:     "ResponseHandlingFailure: missing pagination cursor",
		},
		{
			name:     "side effect failure",
			category: SideEffectFailure,
			message:  "config file not updated",
			want:     "SideEffectFailure: config file not updated",
		},
		{
			name:     "ux contract failure",
			category: UXContractFailure,
			message:  "stdout mismatch",
			want:     "UXContractFailure: stdout mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatFailure(tt.category, "%s", tt.message); got != tt.want {
				t.Fatalf("formatted failure = %q, want %q", got, tt.want)
			}
		})
	}
}
