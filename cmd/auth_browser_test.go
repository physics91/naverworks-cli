package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/physics91/naverworks-cli/internal/web"
)

func TestBrowserLoginRejectsConflictingOptions(t *testing.T) {
	setupTestEnv(t)
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--method", "browser", "--jwt"}, "함께 사용할 수 없습니다"},
		{[]string{"--method", "browser", "--callback-port", "0"}, "callback-port"},
		{[]string{"--method", "browser", "--dry-run"}, "미리보기"},
		{[]string{"--method", "browser", "--generate-input"}, "미리보기"},
		{[]string{"--method", "unknown"}, "유효하지 않은 인증 방식"},
	} {
		args := append([]string{"auth", "login"}, tc.args...)
		_, err := runCLI(t, args...)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%v: got %v; want %q", args, err, tc.want)
		}
	}
}

func TestBrowserStatusAndLogoutRejectUnsupportedMethodsAndPreview(t *testing.T) {
	setupTestEnv(t)
	for _, action := range []string{"status", "logout"} {
		for _, extra := range [][]string{{"--method", "unknown"}, {"--method", "browser", "--dry-run"}, {"--method", "browser", "--generate-input"}} {
			if _, err := runCLI(t, append([]string{"auth", action}, extra...)...); err == nil {
				t.Fatalf("%s accepted invalid flags: %v", action, extra)
			}
		}
	}
}

func TestBrowserLogoutCLIWorksWithoutBrowserAndPreservesAPIAuth(t *testing.T) {
	home := setupTestEnv(t)
	writeTestConfig(t, home)
	t.Setenv("NW_BROWSER_PATH", filepath.Join(t.TempDir(), "missing-browser"))
	dir, err := web.SessionDir(config.DefaultPath(), "default")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "owner.json"), []byte(`{"kind":"browser-sessions","profile":"default"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := runCLI(t, "auth", "logout", "--method", "browser", "--dry-run"); err == nil {
		t.Fatal("preview accepted for local credential deletion")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal("preview deleted local credentials")
	}
	out, err := runCLI(t, "auth", "logout", "--method", "browser")
	if err != nil || !strings.Contains(out, `"local_session_deleted": true`) {
		t.Fatalf("browser logout failed: %s, %v", out, err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("session remains after CLI logout: %v", err)
	}
	if _, err := runCLI(t, "auth", "status"); err != nil {
		t.Fatalf("browser logout damaged API authentication: %v", err)
	}
	if _, err := runCLI(t, "auth", "logout", "--dry-run"); err == nil {
		t.Fatal("API logout accepted preview and could delete a token")
	}
	if _, err := runCLI(t, "auth", "status"); err != nil {
		t.Fatalf("preview logout damaged API authentication: %v", err)
	}
}
