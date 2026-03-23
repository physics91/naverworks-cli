package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}

func setupTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", tmpDir)
	}
	t.Setenv("NW_PROFILE", "")
	t.Setenv("NW_CLIENT_ID", "")
	t.Setenv("NW_CLIENT_SECRET", "")
	t.Setenv("NW_SERVICE_ACCOUNT_ID", "")
	t.Setenv("NW_PRIVATE_KEY_PATH", "")
	t.Setenv("NW_DOMAIN_ID", "")
	t.Setenv("NW_BOT_ID", "")
	t.Setenv("NW_SCOPE", "")
	t.Setenv("NW_DEFAULT_CALENDAR_USER_ID", "")
	t.Setenv("NW_SCIM_ACCESS_TOKEN", "")
	return tmpDir
}

func writeTestConfig(t *testing.T, tmpDir string) {
	t.Helper()
	cfgDir := filepath.Join(tmpDir, ".config", "naverworks")
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	// Legacy flat format matching config.Load / auth.TokenStore.Load
	cfgData := `{"client_id":"test","client_secret":"test","bot_id":"bot1"}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(cfgData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	tokenData := `{"auth_method":"jwt","access_token":"test-token","token_type":"Bearer","expires_at":"2099-01-01T00:00:00Z"}`
	if err := os.WriteFile(filepath.Join(cfgDir, "token.json"), []byte(tokenData), 0600); err != nil {
		t.Fatalf("failed to write token: %v", err)
	}
}

func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var stdout string
	var cmdErr error
	stdout = captureStdout(t, func() {
		rootCmd.SetArgs(args)
		cmdErr = rootCmd.Execute()
		rootCmd.SetArgs(nil)
	})
	return stdout, cmdErr
}

func TestSmoke_Version(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "version")
	if err != nil {
		t.Fatalf("version failed: %v", err)
	}
	var v map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &v); err != nil {
		t.Fatalf("version output is not valid JSON: %v\noutput: %q", err, out)
	}
	for _, key := range []string{"version", "commit", "build_date"} {
		if _, ok := v[key]; !ok {
			t.Errorf("version JSON missing %q key", key)
		}
	}
}

func TestSmoke_Help(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "--help")
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if !strings.Contains(out, "naverworks") {
		t.Error("help output missing 'naverworks'")
	}
}

func TestSmoke_ConfigGetInvalidKey(t *testing.T) {
	setupTestEnv(t)
	_, err := runCLI(t, "config", "get", "no_such_key")
	if err == nil {
		t.Fatal("expected error for invalid config key")
	}
	if !strings.Contains(err.Error(), "유효하지 않은 설정 키") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_AuthStatus_NotLoggedIn(t *testing.T) {
	setupTestEnv(t)
	_, err := runCLI(t, "auth", "status")
	if err == nil {
		t.Fatal("expected error when no token exists")
	}
}

func TestSmoke_BotSend_MissingTarget(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "bot", "send", "--text", "hi")
	if err == nil {
		t.Fatal("expected error when neither --to nor --channel specified")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("expected --to mention in error, got: %v", err)
	}
}

func TestSmoke_BotSend_ConflictingFlags(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "bot", "send", "--to", "u1", "--channel", "c1", "--text", "hi")
	if err == nil {
		t.Fatal("expected error for conflicting --to and --channel")
	}
	if !strings.Contains(err.Error(), "동시에 지정할 수 없습니다") {
		t.Errorf("expected conflict error, got: %v", err)
	}
}
