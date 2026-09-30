package web

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissingBrowserFailsBeforeCreatingLoginSession(t *testing.T) {
	dir, err := SessionDir(filepath.Join(t.TempDir(), "config.json"), "default")
	if err != nil {
		t.Fatal(err)
	}
	err = (Browser{SessionDir: dir, Profile: "default", ExecPath: filepath.Join(t.TempDir(), "missing-browser")}).Login(context.Background())
	if err == nil || !strings.Contains(err.Error(), "NW_BROWSER_PATH") {
		t.Fatalf("missing browser was not reported: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(dir)); !os.IsNotExist(err) {
		t.Fatalf("missing browser created a credential directory: %v", err)
	}
}

func TestBrowserDiscoveryReturnsInstallGuidanceWhenNoneFound(t *testing.T) {
	_, err := findBrowserExecutable([]string{filepath.Join(t.TempDir(), "missing-chrome"), filepath.Join(t.TempDir(), "missing-chromium")})
	if err == nil || !strings.Contains(err.Error(), "자동 설치하지 않습니다") || !strings.Contains(err.Error(), "NW_BROWSER_PATH") {
		t.Fatalf("missing browser discovery: %v", err)
	}
}
