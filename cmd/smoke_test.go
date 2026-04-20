package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// containsCommand checks if helpOutput contains cmdName as an actual Cobra
// subcommand (indented with leading whitespace followed by the command name
// and a space or newline). This avoids false positives when a short name
// like "list" matches inside longer names like "list-members".
func containsCommand(helpOutput, cmdName string) bool {
	return strings.Contains(helpOutput, "  "+cmdName+" ") ||
		strings.Contains(helpOutput, "  "+cmdName+"\n") ||
		strings.Contains(helpOutput, "\t"+cmdName+" ") ||
		strings.Contains(helpOutput, "\t"+cmdName+"\n")
}

func setupTestEnv(t *testing.T) string {
	t.Helper()
	homeDir := clitest.NewHarness(t).HomeDir()

	origAPIBaseURL := apiBaseURL
	origAuthBaseURL := authBaseURL
	origSCIMBaseURL := scimBaseURL
	t.Cleanup(func() {
		apiBaseURL = origAPIBaseURL
		authBaseURL = origAuthBaseURL
		scimBaseURL = origSCIMBaseURL
	})

	return homeDir
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
	result, err := clitest.FromCurrentEnv(t).Run(args, newRootCommandRunner(t))
	return result.Stdout, err
}

func resetFlagSet(fs *pflag.FlagSet) {
	fs.VisitAll(func(f *pflag.Flag) {
		_ = fs.Set(f.Name, f.DefValue)
		f.Changed = false
	})
}

func resetCommandTreeFlags(cmd *cobra.Command) {
	resetFlagSet(cmd.Flags())
	resetFlagSet(cmd.PersistentFlags())
	for _, child := range cmd.Commands() {
		resetCommandTreeFlags(child)
	}
}

func setAPIBaseURL(t *testing.T, url string) {
	t.Helper()

	orig := apiBaseURL
	apiBaseURL = url
	t.Cleanup(func() {
		apiBaseURL = orig
	})
}

func setAuthBaseURL(t *testing.T, url string) {
	t.Helper()

	orig := authBaseURL
	authBaseURL = url
	t.Cleanup(func() {
		authBaseURL = orig
	})
}

func setSCIMBaseURL(t *testing.T, url string) {
	t.Helper()

	orig := scimBaseURL
	scimBaseURL = url
	t.Cleanup(func() {
		scimBaseURL = orig
	})
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

func TestSmoke_AuthHelpIncludesDoctor(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "auth", "--help")
	if err != nil {
		t.Fatalf("auth --help failed: %v", err)
	}
	for _, sub := range []string{"login", "status", "logout", "refresh", "setup", "doctor"} {
		if !containsCommand(out, sub) {
			t.Errorf("auth --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_AuthDoctor_NoConfig(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "auth", "doctor")
	if err != nil {
		t.Fatalf("auth doctor failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("auth doctor output is not valid JSON: %v\noutput: %q", err, out)
	}
	if _, ok := payload["checks"]; !ok {
		t.Fatalf("auth doctor output missing checks: %v", payload)
	}
}

func TestSmoke_BotSend_DryRunWithoutToken(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "--dry-run", "bot", "send", "--bot-id", "bot1", "--to", "u1", "--text", "hi")
	if err != nil {
		t.Fatalf("bot send --dry-run failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("dry-run output is not valid JSON: %v\noutput: %q", err, out)
	}
	if payload["method"] != "POST" {
		t.Fatalf("expected POST method, got %#v", payload["method"])
	}
	if payload["dry_run"] != true {
		t.Fatalf("expected dry_run=true, got %#v", payload["dry_run"])
	}
}

func TestSmoke_DirectoryListUsers_DryRunWithoutToken(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "--dry-run", "directory", "list-users", "--count", "1")
	if err != nil {
		t.Fatalf("directory list-users --dry-run failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("dry-run output is not valid JSON: %v\noutput: %q", err, out)
	}
	if payload["method"] != "GET" {
		t.Fatalf("expected GET method, got %#v", payload["method"])
	}
	if payload["path"] != "/users?count=1" {
		t.Fatalf("expected /users?count=1 path, got %#v", payload["path"])
	}
}

func TestSmoke_SCIMListUsers_WithScimTokenOnly(t *testing.T) {
	tmpDir := setupTestEnv(t)
	cfgDir := filepath.Join(tmpDir, ".config", "naverworks")
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	cfgData := `{"scim_access_token":"scim-token"}`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte(cfgData), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"Resources":[]}`))
	}))
	defer server.Close()
	setSCIMBaseURL(t, server.URL)

	out, err := runCLI(t, "scim", "list-users")
	if err != nil {
		t.Fatalf("scim list-users failed: %v", err)
	}
	if authHeader != "Bearer scim-token" {
		t.Fatalf("expected SCIM token auth, got %q", authHeader)
	}
	if !strings.Contains(out, `"Resources"`) {
		t.Fatalf("expected SCIM list output, got %q", out)
	}
}

func TestSmoke_SCIMListUsers_DryRunWithoutToken(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "--dry-run", "scim", "list-users")
	if err != nil {
		t.Fatalf("scim list-users --dry-run failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("dry-run output is not valid JSON: %v\noutput: %q", err, out)
	}
	if payload["method"] != "GET" {
		t.Fatalf("expected GET method, got %#v", payload["method"])
	}
}

func TestSmoke_SCIMListUsers_DryRunUsesCurrentProfile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	cfgDir := filepath.Join(tmpDir, ".config", "naverworks")
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	profiles := &config.ProfileConfig{
		CurrentProfile: "work",
		Profiles: map[string]*config.Config{
			"default": {},
			"work":    {},
		},
	}
	cfgData, err := json.Marshal(profiles)
	if err != nil {
		t.Fatalf("failed to marshal profile config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.json"), cfgData, 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	out, err := runCLI(t, "--dry-run", "scim", "list-users")
	if err != nil {
		t.Fatalf("scim list-users --dry-run failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("dry-run output is not valid JSON: %v\noutput: %q", err, out)
	}
	if payload["profile"] != "work" {
		t.Fatalf("expected current_profile in preview output, got %#v", payload["profile"])
	}
}

func TestSmoke_DirectoryListUsers_DryRunAllShowsPreviewPlan(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "--dry-run", "directory", "list-users", "--all")
	if err != nil {
		t.Fatalf("directory list-users --dry-run --all failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("dry-run output is not valid JSON: %v\noutput: %q", err, out)
	}
	if payload["method"] != "GET" {
		t.Fatalf("expected GET method, got %#v", payload["method"])
	}
	if payload["path"] != "/users" {
		t.Fatalf("expected /users path, got %#v", payload["path"])
	}
	pagination, ok := payload["pagination"].(map[string]any)
	if !ok || pagination["all"] != true {
		t.Fatalf("expected pagination preview metadata, got %#v", payload["pagination"])
	}
}

func TestSmoke_DirectoryListUsers_GenerateInputPlanOutAll(t *testing.T) {
	setupTestEnv(t)
	planFile := filepath.Join(t.TempDir(), "plan.json")
	out, err := runCLI(t, "--generate-input", "--plan-out", planFile, "directory", "list-users", "--all")
	if err != nil {
		t.Fatalf("directory list-users --generate-input --plan-out --all failed: %v", err)
	}
	if strings.TrimSpace(out) != "{}" {
		t.Fatalf("expected pure generated input, got %q", out)
	}
	data, err := os.ReadFile(planFile)
	if err != nil {
		t.Fatalf("expected plan file to be written: %v", err)
	}
	if !strings.Contains(string(data), `"pagination"`) {
		t.Fatalf("expected pagination metadata in plan file: %s", data)
	}
}

func TestSmoke_DriveDownload_DryRunShowsPreviewPlan(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "--dry-run", "drive", "download", "file1", "--user-id", "u1")
	if err != nil {
		t.Fatalf("drive download --dry-run failed: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &payload); err != nil {
		t.Fatalf("dry-run output is not valid JSON: %v\noutput: %q", err, out)
	}
	if payload["method"] != "GET" {
		t.Fatalf("expected GET method, got %#v", payload["method"])
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

func TestSmoke_BotSend_TextAndJsonConflict(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "bot", "send", "--to", "u1", "--channel", "", "--text", "hi", "--json", `{"content":{"type":"text"}}`)
	if err == nil {
		t.Fatal("expected error for conflicting --text and --json")
	}
	if !strings.Contains(err.Error(), "--text와 --json은 동시에 지정할 수 없습니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_BotSend_NeitherTextNorJson(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "bot", "send", "--to", "u1", "--channel", "", "--text", "", "--json", "")
	if err == nil {
		t.Fatal("expected error when neither --text nor --json specified")
	}
	if !strings.Contains(err.Error(), "--text 또는 --json 중 하나를 지정하세요") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_BotHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "bot", "--help")
	if err != nil {
		t.Fatalf("bot --help failed: %v", err)
	}
	// Short ambiguous names — use containsCommand for exact matching
	for _, sub := range []string{"list", "get", "create", "update", "patch", "delete", "send"} {
		if !containsCommand(out, sub) {
			t.Errorf("bot --help missing subcommand %q", sub)
		}
	}
	// Longer unique names — strings.Contains is fine
	for _, sub := range []string{
		"regenerate-secret",
		"create-attachment", "get-attachment",
		"get-channel", "channel-members", "create-channel", "leave-channel",
		"domain", "persistent-menu", "richmenu",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("bot --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_BotDomainHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "bot", "domain", "--help")
	if err != nil {
		t.Fatalf("bot domain --help failed: %v", err)
	}
	for _, sub := range []string{"register", "list", "update", "patch", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("bot domain --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{"add-members", "list-members", "remove-member"} {
		if !strings.Contains(out, sub) {
			t.Errorf("bot domain --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_BotPersistentMenuHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "bot", "persistent-menu", "--help")
	if err != nil {
		t.Fatalf("bot persistent-menu --help failed: %v", err)
	}
	for _, sub := range []string{"set", "get", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("bot persistent-menu --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_BotRichMenuHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "bot", "richmenu", "--help")
	if err != nil {
		t.Fatalf("bot richmenu --help failed: %v", err)
	}
	for _, sub := range []string{"create", "list", "get", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("bot richmenu --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"set-image", "get-image",
		"set-user", "get-user", "delete-user",
		"set-default", "get-default", "delete-default",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("bot richmenu --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_BotCreate_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "bot", "create")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_BotDomainRemoveMember_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "bot", "domain", "remove-member", "d1")
	if err == nil {
		t.Fatal("expected error when userId arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_BotRichMenuSetImage_MissingFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "bot", "richmenu", "set-image", "rm1")
	if err == nil {
		t.Fatal("expected error when --file is missing")
	}
	if !strings.Contains(err.Error(), "--file 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_BotCreateAttachment_MissingFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "bot", "create-attachment")
	if err == nil {
		t.Fatal("expected error when --file is missing")
	}
	if !strings.Contains(err.Error(), "--file 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_ReadJSONFlagRaw_Missing(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("json", "", "JSON body")

	_, err := readJSONFlagRaw(cmd)
	if err == nil {
		t.Fatal("expected error when --json is empty")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_ReadJSONFlagRaw_InvalidJSON(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("json", "", "JSON body")
	_ = cmd.Flags().Set("json", "{invalid")

	_, err := readJSONFlagRaw(cmd)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "유효하지 않은 JSON") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_ReadJSONFlagRaw_Valid(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("json", "", "JSON body")
	_ = cmd.Flags().Set("json", `{"key":"value"}`)

	data, err := readJSONFlagRaw(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `{"key":"value"}` {
		t.Errorf("unexpected data: %s", data)
	}
}

func TestSmoke_ReadJSONFlag_Valid(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("json", "", "JSON body")
	_ = cmd.Flags().Set("json", `{"title":"hello","count":3}`)

	body, err := readJSONFlag(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body["title"] != "hello" {
		t.Errorf("expected title=hello, got %v", body["title"])
	}
	if body["count"] != float64(3) {
		t.Errorf("expected count=3, got %v", body["count"])
	}
}

func TestSmoke_ReadJSONFlag_InvalidJSON(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("json", "", "JSON body")
	_ = cmd.Flags().Set("json", "not-json")

	_, err := readJSONFlag(cmd)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSmoke_ReadFileFlag_Missing(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("file", "", "파일 경로")

	_, _, err := readFileFlag(cmd, "file")
	if err == nil {
		t.Fatal("expected error when --file is empty")
	}
	if !strings.Contains(err.Error(), "--file 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_ReadFileFlag_NotFound(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("file", "", "파일 경로")
	_ = cmd.Flags().Set("file", "/nonexistent/path/to/file.txt")

	_, _, err := readFileFlag(cmd, "file")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "파일 접근 실패") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_ReadFileFlag_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-upload.txt")
	content := []byte("hello naverworks")
	if err := os.WriteFile(testFile, content, 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("file", "", "파일 경로")
	_ = cmd.Flags().Set("file", testFile)

	data, name, err := readFileFlag(cmd, "file")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "hello naverworks" {
		t.Errorf("expected file content 'hello naverworks', got %q", string(data))
	}
	if name != "test-upload.txt" {
		t.Errorf("expected file name 'test-upload.txt', got %q", name)
	}
}

func TestSmoke_CalendarCreateCalendar_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "calendar", "create-calendar")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_CalendarDeleteEvent_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "calendar", "delete-event")
	if err == nil {
		t.Fatal("expected error when args are missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_CalendarDeleteEvent_OneArg(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "calendar", "delete-event", "cal1")
	if err == nil {
		t.Fatal("expected error when only one arg is provided")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_CalendarDefaultDeleteEvent_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "calendar", "default", "delete-event")
	if err == nil {
		t.Fatal("expected error when eventId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_CalendarGetCalendar_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "calendar", "get-calendar")
	if err == nil {
		t.Fatal("expected error when calendarId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_CalendarDefaultHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "calendar", "default", "--help")
	if err != nil {
		t.Fatalf("calendar default --help failed: %v", err)
	}
	// All names here are hyphenated and unique — strings.Contains is fine
	for _, sub := range []string{"list-events", "get-event", "create-event", "update-event", "delete-event"} {
		if !strings.Contains(out, sub) {
			t.Errorf("calendar default --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_CalendarHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "calendar", "--help")
	if err != nil {
		t.Fatalf("calendar --help failed: %v", err)
	}
	for _, sub := range []string{"default"} {
		if !containsCommand(out, sub) {
			t.Errorf("calendar --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{"create-calendar", "get-calendar", "update-calendar", "delete-calendar",
		"get-personal", "update-personal", "remove-user", "update-event", "delete-event"} {
		if !strings.Contains(out, sub) {
			t.Errorf("calendar --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_ReadJSONFlagRaw_Stdin(t *testing.T) {
	// Simulate stdin by replacing os.Stdin with a pipe
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}

	input := `{"from":"stdin"}`
	go func() {
		w.Write([]byte(input))
		w.Close()
	}()
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	cmd := &cobra.Command{}
	cmd.Flags().String("json", "", "JSON body")
	_ = cmd.Flags().Set("json", "-")

	data, err := readJSONFlagRaw(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != input {
		t.Errorf("expected %q, got %q", input, string(data))
	}
}

// ─── Board Smoke Tests ───

func TestSmoke_BoardHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "board", "--help")
	if err != nil {
		t.Fatalf("board --help failed: %v", err)
	}
	for _, sub := range []string{"create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("board --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"list-readers", "list-recent", "list-my", "list-must",
		"create-attachment", "list-attachments", "get-attachment", "delete-attachment",
		"create-comment", "get-comment", "update-comment", "delete-comment",
		"create-comment-attachment", "list-comment-attachments", "get-comment-attachment", "delete-comment-attachment",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("board --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_BoardCreate_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "board", "create")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_BoardDeleteComment_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "board", "delete-comment")
	if err == nil {
		t.Fatal("expected error when args are missing")
	}
	if !strings.Contains(err.Error(), "accepts 3 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_BoardGetCommentAttachment_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "board", "get-comment-attachment", "b1", "p1")
	if err == nil {
		t.Fatal("expected error when not enough args")
	}
	if !strings.Contains(err.Error(), "accepts 4 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_BoardCreateAttachment_MissingFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "board", "create-attachment", "b1", "p1")
	if err == nil {
		t.Fatal("expected error when --file is missing")
	}
	if !strings.Contains(err.Error(), "--file 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Mail Smoke Tests ───

func TestSmoke_MailHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "mail", "--help")
	if err != nil {
		t.Fatalf("mail --help failed: %v", err)
	}
	for _, sub := range []string{"send", "get", "delete", "list", "update"} {
		if !containsCommand(out, sub) {
			t.Errorf("mail --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"list-folders", "get-folder",
		"unread-count", "get-attachment", "list-favorite-folders",
		"create-folder", "update-folder", "delete-folder",
		"filter", "migration", "forwarding",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("mail --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_MailFilterHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "mail", "filter", "--help")
	if err != nil {
		t.Fatalf("mail filter --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("mail filter --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_MailMigrationHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "mail", "migration", "--help")
	if err != nil {
		t.Fatalf("mail migration --help failed: %v", err)
	}
	for _, sub := range []string{"create-imap", "get-imap", "delete-imap", "create-pop3"} {
		if !strings.Contains(out, sub) {
			t.Errorf("mail migration --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_MailUpdate_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "mail", "update", "mail1", "--user-id", "testuser")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_MailGetAttachment_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "mail", "get-attachment", "mail1")
	if err == nil {
		t.Fatal("expected error when attachmentId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_MailGet_HasThreadsFlag(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "mail", "get", "--help")
	if err != nil {
		t.Fatalf("mail get --help failed: %v", err)
	}
	if !strings.Contains(out, "--has-threads") {
		t.Errorf("mail get --help missing --has-threads flag; got: %s", out)
	}
}

// ─── Task Smoke Tests ───

func TestSmoke_TaskHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "task", "--help")
	if err != nil {
		t.Fatalf("task --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete", "move", "complete", "incomplete"} {
		if !containsCommand(out, sub) {
			t.Errorf("task --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"list-categories", "create-category", "get-category", "update-category", "delete-category",
		"complete-assignee", "incomplete-assignee",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("task --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_TaskCreateCategory_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "task", "create-category", "--user-id", "testuser")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_TaskMove_MissingCategory(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "task", "move", "task1", "--user-id", "testuser")
	if err == nil {
		t.Fatal("expected error when --category is missing")
	}
	if !strings.Contains(err.Error(), "--category는 필수입니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_TaskCompleteAssignee_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "task", "complete-assignee", "task1")
	if err == nil {
		t.Fatal("expected error when userId arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Contact Smoke Tests ───

func TestSmoke_ContactHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "contact", "--help")
	if err != nil {
		t.Fatalf("contact --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete", "tag"} {
		if !containsCommand(out, sub) {
			t.Errorf("contact --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"list-user", "full-update", "force-delete",
		"upload-photo", "get-photo", "delete-photo",
		"list-tags", "list-user-tags",
		"custom-property",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("contact --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_ContactCustomPropertyHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "contact", "custom-property", "--help")
	if err != nil {
		t.Fatalf("contact custom-property --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("contact custom-property --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_ContactTagHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "contact", "tag", "--help")
	if err != nil {
		t.Fatalf("contact tag --help failed: %v", err)
	}
	for _, sub := range []string{"create", "get", "update", "patch", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("contact tag --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{"create-user-tags"} {
		if !strings.Contains(out, sub) {
			t.Errorf("contact tag --help missing subcommand %q", sub)
		}
	}
}

// ─── Note Smoke Tests ───

func TestSmoke_NoteHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "note", "--help")
	if err != nil {
		t.Fatalf("note --help failed: %v", err)
	}
	for _, sub := range []string{"create", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("note --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"list-posts", "get-post",
		"create-post", "update-post", "delete-post",
		"patch-post",
		"create-attachment", "list-attachments", "get-attachment", "delete-attachment",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("note --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_NotePatchPost_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "note", "patch-post", "g1", "p1")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_NoteCreateAttachment_MissingFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "note", "create-attachment", "g1", "p1")
	if err == nil {
		t.Fatal("expected error when --file is missing")
	}
	if !strings.Contains(err.Error(), "--file 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Attendance Smoke Tests ───

func TestSmoke_AttendanceHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "attendance", "--help")
	if err != nil {
		t.Fatalf("attendance --help failed: %v", err)
	}
	for _, sub := range []string{"status"} {
		if !containsCommand(out, sub) {
			t.Errorf("attendance --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"clock-in", "clock-out",
		"list-absences", "list-annual-leaves",
		"create-timecard", "list-timecards", "get-timecard", "update-timecard",
		"adjust-annual-leave",
		"list-absence-schedules",
		"create-absence", "get-absence", "update-absence", "delete-absence",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("attendance --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_AttendanceCreateTimecard_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "attendance", "create-timecard")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_AttendanceGetTimecard_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "attendance", "get-timecard")
	if err == nil {
		t.Fatal("expected error when timecardId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_AttendanceDeleteAbsence_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "attendance", "delete-absence")
	if err == nil {
		t.Fatal("expected error when absenceId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── HR Smoke Tests ───

func TestSmoke_HRHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "hr", "--help")
	if err != nil {
		t.Fatalf("hr --help failed: %v", err)
	}
	for _, sub := range []string{
		"list-extension-properties", "create-extension-property", "get-extension-property",
		"update-extension-property", "delete-extension-property",
		"get-user-properties", "get-user-property", "update-user-property",
		"list-leave-types", "create-leave-of-absence", "get-leave-of-absence",
		"update-leave-of-absence", "delete-leave-of-absence", "list-on-leave",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("hr --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_HRCreateExtensionProperty_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "hr", "create-extension-property")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_HRGetUserProperty_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "hr", "get-user-property", "user1")
	if err == nil {
		t.Fatal("expected error when second arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Audit Smoke Tests ───

func TestSmoke_AuditHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "audit", "--help")
	if err != nil {
		t.Fatalf("audit --help failed: %v", err)
	}
	for _, sub := range []string{
		"download-logs", "list-policy-groups",
		"create-policy-group", "get-policy-group", "update-policy-group", "delete-policy-group",
		"add-policy-members", "list-policy-members", "remove-policy-member",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("audit --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_AuditRemovePolicyMember_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "audit", "remove-policy-member", "pg1")
	if err == nil {
		t.Fatal("expected error when userId arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Approval Smoke Tests ───

func TestSmoke_ApprovalHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "approval", "--help")
	if err != nil {
		t.Fatalf("approval --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get"} {
		if !containsCommand(out, sub) {
			t.Errorf("approval --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"list-all", "list-categories", "get-category", "list-forms",
		"create-category", "update-category", "delete-category",
		"create-document", "create-imported-document", "create-document-link",
		"get-form", "upload-attachment", "upload-imported-attachment",
		"linkage-code", "linkage-code-item",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("approval --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_ApprovalLinkageCodeHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "approval", "linkage-code", "--help")
	if err != nil {
		t.Fatalf("approval linkage-code --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update"} {
		if !containsCommand(out, sub) {
			t.Errorf("approval linkage-code --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_ApprovalLinkageCodeItemHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "approval", "linkage-code-item", "--help")
	if err != nil {
		t.Fatalf("approval linkage-code-item --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("approval linkage-code-item --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_ApprovalUploadAttachment_MissingFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "approval", "upload-attachment", "--user-id", "testuser")
	if err == nil {
		t.Fatal("expected error when --file is missing")
	}
	if !strings.Contains(err.Error(), "--file 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Security Smoke Tests ───

func TestSmoke_SecurityHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "security", "--help")
	if err != nil {
		t.Fatalf("security --help failed: %v", err)
	}
	for _, sub := range []string{"get-external-browser", "enable-external-browser", "disable-external-browser"} {
		if !strings.Contains(out, sub) {
			t.Errorf("security --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_SecurityCommandRegistration(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "--help")
	if err != nil {
		t.Fatalf("root --help failed: %v", err)
	}
	if !strings.Contains(out, "security") {
		t.Error("root --help missing 'security' command")
	}
}

func TestSmoke_ContactFullUpdate_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "contact", "full-update", "c1")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Directory Smoke Tests ───

func TestSmoke_DirectoryHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "directory", "--help")
	if err != nil {
		t.Fatalf("directory --help failed: %v", err)
	}
	for _, sub := range []string{
		// Existing read
		"list-users", "get-user", "list-groups", "get-group", "list-orgunits", "get-orgunit",
		"list-levels", "list-positions", "list-user-types", "list-employment-types",
		// Task 4-1: User CUD
		"create-user", "update-user", "patch-user", "delete-user", "force-delete-user",
		"undelete-user", "suspend-user", "unsuspend-user", "force-logout-user",
		"move-user", "set-leave", "clear-leave",
		// Task 4-2: User Profile
		"upload-photo", "get-photo", "delete-photo", "profile-status",
		// Task 4-3: Email + Invitations + Links
		"add-alias-email", "delete-alias-email", "send-invitation", "send-invitation-all",
		"link-to-works", "link-all-to-works", "unlink-to-works",
		"link-to-line", "link-all-to-line", "unlink-to-line",
		"get-link-url", "reset-link-url",
		// Task 4-4: External Keys + Custom Properties
		"upsert-external-keys", "list-external-keys", "user-custom-property",
		// Task 4-5: Group CUD
		"create-group", "update-group", "patch-group", "delete-group",
		"list-group-members", "add-group-members", "remove-group-member",
		"list-group-admins", "add-group-admin", "remove-group-admin",
		"upsert-group-external-keys", "list-group-external-keys",
		// Task 4-6: OrgUnit CUD
		"create-orgunit", "update-orgunit", "patch-orgunit", "delete-orgunit",
		"move-orgunit", "list-orgunit-members", "orgunit-access-restrict",
		"upsert-orgunit-external-keys", "list-orgunit-external-keys",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("directory --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DirectoryProfileStatusHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "directory", "profile-status", "--help")
	if err != nil {
		t.Fatalf("directory profile-status --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "patch", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("directory profile-status --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DirectoryUserCustomPropertyHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "directory", "user-custom-property", "--help")
	if err != nil {
		t.Fatalf("directory user-custom-property --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("directory user-custom-property --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DirectoryOrgUnitAccessRestrictHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "directory", "orgunit-access-restrict", "--help")
	if err != nil {
		t.Fatalf("directory orgunit-access-restrict --help failed: %v", err)
	}
	for _, sub := range []string{"create", "get", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("directory orgunit-access-restrict --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DirectoryCreateUser_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "create-user")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryDeleteUser_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "delete-user")
	if err == nil {
		t.Fatal("expected error when userId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryRemoveGroupMember_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "remove-group-member", "g1")
	if err == nil {
		t.Fatal("expected error when memberId arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryAddAliasEmail_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "add-alias-email", "u1")
	if err == nil {
		t.Fatal("expected error when email arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryProfileStatusGet_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "profile-status", "get", "u1")
	if err == nil {
		t.Fatal("expected error when second arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryUploadPhoto_MissingFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "upload-photo", "u1")
	if err == nil {
		t.Fatal("expected error when --file is missing")
	}
	if !strings.Contains(err.Error(), "--file 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryRemoveGroupAdmin_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "remove-group-admin", "g1")
	if err == nil {
		t.Fatal("expected error when userId arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryHelp_Phase4(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "directory", "--help")
	if err != nil {
		t.Fatalf("directory --help failed: %v", err)
	}
	for _, sub := range []string{
		// Task 4-7: Positions
		"get-position", "create-position", "update-position", "patch-position",
		"delete-position", "enable-positions", "disable-positions",
		"upsert-position-external-keys", "list-position-external-keys",
		// Task 4-8: Levels
		"get-level", "create-level", "update-level", "patch-level",
		"delete-level", "enable-levels", "disable-levels",
		"upsert-level-external-keys", "list-level-external-keys",
		// Task 4-9: Employment Types
		"get-employment-type", "create-employment-type", "update-employment-type", "patch-employment-type",
		"delete-employment-type", "enable-employment-types", "disable-employment-types",
		"upsert-employment-type-external-keys", "list-employment-type-external-keys",
		"employment-type-access-restrict",
		// Task 4-10: User Types
		"get-user-type", "create-user-type", "update-user-type", "patch-user-type",
		"delete-user-type", "enable-user-types", "disable-user-types",
		"upsert-user-type-external-keys", "list-user-type-external-keys",
		"user-type-access-restrict",
		// Task 4-11: Profile Statuses Def
		"profile-status-def",
		// Task 4-12: Custom Fields
		"custom-field",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("directory --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DirectoryEmploymentTypeAccessRestrictHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "directory", "employment-type-access-restrict", "--help")
	if err != nil {
		t.Fatalf("directory employment-type-access-restrict --help failed: %v", err)
	}
	for _, sub := range []string{"create", "get", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("directory employment-type-access-restrict --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DirectoryUserTypeAccessRestrictHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "directory", "user-type-access-restrict", "--help")
	if err != nil {
		t.Fatalf("directory user-type-access-restrict --help failed: %v", err)
	}
	for _, sub := range []string{"create", "get", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("directory user-type-access-restrict --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DirectoryProfileStatusDefHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "directory", "profile-status-def", "--help")
	if err != nil {
		t.Fatalf("directory profile-status-def --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "patch", "delete", "enable", "disable"} {
		if !containsCommand(out, sub) {
			t.Errorf("directory profile-status-def --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DirectoryCustomFieldHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "directory", "custom-field", "--help")
	if err != nil {
		t.Fatalf("directory custom-field --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("directory custom-field --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DirectoryCreatePosition_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "create-position")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryDeletePosition_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "delete-position")
	if err == nil {
		t.Fatal("expected error when positionId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryDeleteLevel_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "delete-level")
	if err == nil {
		t.Fatal("expected error when levelId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryCreateEmploymentType_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "create-employment-type")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DirectoryCustomFieldCreate_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "directory", "custom-field", "create")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Drive Phase 5 Smoke Tests ───

func TestSmoke_DriveHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "--help")
	if err != nil {
		t.Fatalf("drive --help failed: %v", err)
	}
	for _, sub := range []string{
		"info", "list", "get", "download", "upload", "mkdir", "delete",
		"copy", "rename", "move", "protect", "unprotect", "lock", "unlock",
		"revision", "link", "share", "shared", "group",
	} {
		if !containsCommand(out, sub) {
			t.Errorf("drive --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"trash-list", "trash-restore", "trash-delete",
		"link-setting", "shared-folder",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("drive --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveRevisionHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "revision", "--help")
	if err != nil {
		t.Fatalf("drive revision --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "restore", "download"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive revision --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveLinkHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "link", "--help")
	if err != nil {
		t.Fatalf("drive link --help failed: %v", err)
	}
	for _, sub := range []string{"get", "create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive link --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveShareHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "share", "--help")
	if err != nil {
		t.Fatalf("drive share --help failed: %v", err)
	}
	for _, sub := range []string{"get", "create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive share --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{"list-sub-folders"} {
		if !strings.Contains(out, sub) {
			t.Errorf("drive share --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveCopy_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "copy", "f1", "--user-id", "testuser")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DriveRevisionGet_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "revision", "get", "f1")
	if err == nil {
		t.Fatal("expected error when revisionId arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 2 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Drive Group (Phase 5 Tasks 5-4 ~ 5-7) Smoke Tests ───

func TestSmoke_DriveGroupHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "group", "--help")
	if err != nil {
		t.Fatalf("drive group --help failed: %v", err)
	}
	for _, sub := range []string{
		"list", "get", "mkdir", "delete", "upload", "download",
		"copy", "rename", "move", "protect", "unprotect", "lock", "unlock",
		"revision", "link", "permission",
	} {
		if !containsCommand(out, sub) {
			t.Errorf("drive group --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"get-folder", "create-folder", "delete-folder",
		"trash-list", "trash-restore", "trash-delete",
		"link-setting",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("drive group --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveGroupRevisionHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "group", "revision", "--help")
	if err != nil {
		t.Fatalf("drive group revision --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "restore", "download"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive group revision --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveGroupLinkHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "group", "link", "--help")
	if err != nil {
		t.Fatalf("drive group link --help failed: %v", err)
	}
	for _, sub := range []string{"get", "create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive group link --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveGroupPermissionHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "group", "permission", "--help")
	if err != nil {
		t.Fatalf("drive group permission --help failed: %v", err)
	}
	for _, sub := range []string{"list", "create", "get", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive group permission --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{"delete-all"} {
		if !strings.Contains(out, sub) {
			t.Errorf("drive group permission --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveGroupCopy_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "group", "copy", "g1", "f1")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DriveGroupRevisionGet_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "group", "revision", "get", "g1", "f1")
	if err == nil {
		t.Fatal("expected error when revisionId arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 3 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Drive Shared (Phase 5 Tasks 5-8 ~ 5-11) Smoke Tests ───

func TestSmoke_DriveSharedHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "shared", "--help")
	if err != nil {
		t.Fatalf("drive shared --help failed: %v", err)
	}
	for _, sub := range []string{
		"list", "get", "download", "upload", "mkdir", "delete",
		"copy", "rename", "move", "protect", "unprotect", "lock", "unlock",
		"revision", "link", "permission",
	} {
		if !containsCommand(out, sub) {
			t.Errorf("drive shared --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"list-drives", "get-drive",
		"create-drive", "update-drive", "delete-drive",
		"trash-list", "trash-restore", "trash-delete",
		"link-setting", "file-permission",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("drive shared --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveSharedRevisionHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "shared", "revision", "--help")
	if err != nil {
		t.Fatalf("drive shared revision --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "restore", "download"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive shared revision --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveSharedLinkHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "shared", "link", "--help")
	if err != nil {
		t.Fatalf("drive shared link --help failed: %v", err)
	}
	for _, sub := range []string{"get", "create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive shared link --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveSharedPermissionHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "shared", "permission", "--help")
	if err != nil {
		t.Fatalf("drive shared permission --help failed: %v", err)
	}
	for _, sub := range []string{"list", "create", "get", "update", "delete", "enable", "disable"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive shared permission --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{"delete-all"} {
		if !strings.Contains(out, sub) {
			t.Errorf("drive shared permission --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveSharedFilePermissionHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "shared", "file-permission", "--help")
	if err != nil {
		t.Fatalf("drive shared file-permission --help failed: %v", err)
	}
	for _, sub := range []string{"list", "create", "get", "update", "delete", "enable", "disable"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive shared file-permission --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{"delete-all"} {
		if !strings.Contains(out, sub) {
			t.Errorf("drive shared file-permission --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveSharedCopy_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "shared", "copy", "d1", "f1")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DriveSharedRevisionGet_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "shared", "revision", "get", "d1", "f1")
	if err == nil {
		t.Fatal("expected error when revisionId arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 3 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DriveSharedUpload_MissingFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "shared", "upload", "d1")
	if err == nil {
		t.Fatal("expected error when --file is missing")
	}
	if !strings.Contains(err.Error(), "--file 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Drive SharedFolder (Phase 5 Tasks 5-12 ~ 5-14) Smoke Tests ───

func TestSmoke_DriveSharedFolderHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "shared-folder", "--help")
	if err != nil {
		t.Fatalf("drive shared-folder --help failed: %v", err)
	}
	for _, sub := range []string{
		"list", "files", "get", "leave",
		"mkdir", "delete", "upload", "download",
		"copy", "rename", "move", "protect", "unprotect", "lock", "unlock",
		"revision", "link",
	} {
		if !containsCommand(out, sub) {
			t.Errorf("drive shared-folder --help missing subcommand %q", sub)
		}
	}
	for _, sub := range []string{
		"list-members", "list-files", "get-file",
		"link-setting",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("drive shared-folder --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveSharedFolderRevisionHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "shared-folder", "revision", "--help")
	if err != nil {
		t.Fatalf("drive shared-folder revision --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "restore", "download"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive shared-folder revision --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveSharedFolderLinkHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "drive", "shared-folder", "link", "--help")
	if err != nil {
		t.Fatalf("drive shared-folder link --help failed: %v", err)
	}
	for _, sub := range []string{"get", "create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("drive shared-folder link --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_DriveSharedFolderCopy_MissingJSON(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "shared-folder", "copy", "sf1", "f1", "--user-id", "testuser")
	if err == nil {
		t.Fatal("expected error when --json is missing")
	}
	if !strings.Contains(err.Error(), "--json 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DriveSharedFolderRevisionGet_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "shared-folder", "revision", "get", "sf1", "f1")
	if err == nil {
		t.Fatal("expected error when revisionId arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 3 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DriveSharedFolderUpload_MissingFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "shared-folder", "upload", "sf1", "--user-id", "testuser")
	if err == nil {
		t.Fatal("expected error when --file is missing")
	}
	if !strings.Contains(err.Error(), "--file 플래그가 필요합니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_DriveUpload_ResumeFlag(t *testing.T) {
	setupTestEnv(t)
	for _, path := range [][]string{
		{"drive", "upload", "--help"},
		{"drive", "shared", "upload", "--help"},
		{"drive", "group", "upload", "--help"},
		{"drive", "shared-folder", "upload", "--help"},
	} {
		out, err := runCLI(t, path...)
		if err != nil {
			t.Fatalf("%v failed: %v", path, err)
		}
		if !strings.Contains(out, "--resume") {
			t.Errorf("%v missing --resume flag in help output", path)
		}
	}
}

// ─── SCIM Smoke Tests ───

func TestSmoke_ScimHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "scim", "--help")
	if err != nil {
		t.Fatalf("scim --help failed: %v", err)
	}
	for _, sub := range []string{
		"list-users", "get-user", "create-user", "update-user", "patch-user", "delete-user",
		"list-groups", "get-group", "create-group", "update-group", "patch-group", "delete-group",
	} {
		if !strings.Contains(out, sub) {
			t.Errorf("scim --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_ScimCreateUser_MissingData(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	t.Setenv("NW_SCIM_ACCESS_TOKEN", "test-scim-token")
	_, err := runCLI(t, "scim", "create-user")
	if err == nil {
		t.Fatal("expected error when --data is missing")
	}
	if !strings.Contains(err.Error(), "--data는 필수입니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_ScimGetUser_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "scim", "get-user")
	if err == nil {
		t.Fatal("expected error when id arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_ScimDeleteGroup_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "scim", "delete-group")
	if err == nil {
		t.Fatal("expected error when id arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Business Place Smoke Tests ───

func TestSmoke_BusinessPlaceHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "business-place", "--help")
	if err != nil {
		t.Fatalf("business-place --help failed: %v", err)
	}
	for _, sub := range []string{"list", "get", "create", "update", "delete"} {
		if !containsCommand(out, sub) {
			t.Errorf("business-place --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_BusinessPlaceCreate_MissingName(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "business-place", "create")
	if err == nil {
		t.Fatal("expected error when --name is missing")
	}
	if !strings.Contains(err.Error(), "--name은 필수입니다") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_BusinessPlaceGet_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "business-place", "get")
	if err == nil {
		t.Fatal("expected error when businessPlaceId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_BusinessPlaceDelete_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "business-place", "delete")
	if err == nil {
		t.Fatal("expected error when businessPlaceId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Form Smoke Tests ───

func TestSmoke_FormHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "form", "--help")
	if err != nil {
		t.Fatalf("form --help failed: %v", err)
	}
	for _, sub := range []string{"list-responses", "download-attachment"} {
		if !strings.Contains(out, sub) {
			t.Errorf("form --help missing subcommand %q", sub)
		}
	}
}

func TestSmoke_FormListResponses_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "form", "list-responses")
	if err == nil {
		t.Fatal("expected error when formId is missing")
	}
	if !strings.Contains(err.Error(), "accepts 1 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSmoke_FormDownloadAttachment_MissingArgs(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "form", "download-attachment", "f1", "r1")
	if err == nil {
		t.Fatal("expected error when attachmentId arg is missing")
	}
	if !strings.Contains(err.Error(), "accepts 3 arg(s)") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── Flag Branch Path Tests (INFO 3) ───

func TestSmoke_DriveSharedFolderMkdir_WithParent(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "shared-folder", "mkdir", "sf1", "--parent", "file1", "--json", `{"fileName":"test"}`, "--user-id", "testuser")
	// Should fail on API call (not logged in), NOT on flag/arg parsing
	if err != nil && strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("unexpected flag error: %v", err)
	}
}

func TestSmoke_DriveSharedFolderUpload_WithFolder(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "shared-folder", "upload", "sf1", "--folder", "file1", "--file", "/tmp/test", "--user-id", "testuser")
	if err != nil && strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("unexpected flag error: %v", err)
	}
}

func TestSmoke_DriveSharedMkdir_WithParent(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "shared", "mkdir", "driveId1", "--parent", "file1", "--json", `{"fileName":"test"}`)
	if err != nil && strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("unexpected flag error: %v", err)
	}
}

func TestSmoke_DriveGroupMkdir_WithParent(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)
	_, err := runCLI(t, "drive", "group", "mkdir", "groupId1", "--parent", "file1", "--json", `{"fileName":"test"}`)
	if err != nil && strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("unexpected flag error: %v", err)
	}
}

func TestResolveUserID(t *testing.T) {
	tests := []struct {
		name       string
		flagValue  string
		defaultUID string
		authMethod auth.AuthMethod
		wantUID    string
		wantErr    bool
	}{
		{"flag value used", "user1", "default1", auth.AuthMethodOAuth, "user1", false},
		{"default used when flag empty", "", "default1", auth.AuthMethodOAuth, "default1", false},
		{"error when both empty", "", "", auth.AuthMethodOAuth, "", true},
		{"me rejected in jwt mode", "me", "", auth.AuthMethodJWT, "", true},
		{"me allowed in oauth mode", "me", "", auth.AuthMethodOAuth, "me", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("user-id", tt.flagValue, "")

			uid, err := resolveUserID(cmd, tt.defaultUID, tt.authMethod)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if uid != tt.wantUID {
				t.Errorf("got %q, want %q", uid, tt.wantUID)
			}
		})
	}
}

func TestRequireTitleBodyPost(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		body    string
		wantErr bool
	}{
		{"both set", "t1", "b1", false},
		{"missing title", "", "b1", true},
		{"missing body", "t1", "", true},
		{"both missing", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("title", tt.title, "")
			cmd.Flags().String("body", tt.body, "")
			_, err := requireTitleBodyPost(cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestParseOptionalJSONData(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantNil bool
		wantErr bool
	}{
		{"empty", "", true, false},
		{"valid json", `{"key":"val"}`, false, false},
		{"invalid json", `{bad`, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("data", tt.data, "")
			result, err := parseOptionalJSONData(cmd)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil && result != nil {
				t.Error("expected nil result")
			}
			if !tt.wantNil && result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestResolveBotID(t *testing.T) {
	tests := []struct {
		name    string
		flagVal string
		cfgVal  string
		wantID  string
		wantErr bool
	}{
		{"flag value", "flag-bot", "cfg-bot", "flag-bot", false},
		{"config value", "", "cfg-bot", "cfg-bot", false},
		{"both empty", "", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().String("bot-id", tt.flagVal, "")
			id, err := resolveBotID(cmd, tt.cfgVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if !tt.wantErr && id != tt.wantID {
				t.Errorf("got %q, want %q", id, tt.wantID)
			}
		})
	}
}

func TestResolveOrCreateProfile(t *testing.T) {
	origProfileName := profileName
	profileName = ""
	t.Cleanup(func() {
		profileName = origProfileName
	})
	pc := &config.ProfileConfig{
		CurrentProfile: "default",
		Profiles: map[string]*config.Config{
			"default": {ClientID: "cid1"},
		},
	}
	cfg, name := resolveOrCreateProfile(pc)
	if name != "default" {
		t.Errorf("expected default, got %s", name)
	}
	if cfg.ClientID != "cid1" {
		t.Errorf("expected cid1, got %s", cfg.ClientID)
	}
}

func TestResolveOrCreateProfile_UsesEnvSelectionForMissingProfile(t *testing.T) {
	origProfileName := profileName
	profileName = ""
	t.Cleanup(func() {
		profileName = origProfileName
	})
	t.Setenv("NW_PROFILE", "work")

	pc := &config.ProfileConfig{
		CurrentProfile: "default",
		Profiles: map[string]*config.Config{
			"default": {ClientID: "cid1"},
		},
	}
	cfg, name := resolveOrCreateProfile(pc)
	if name != "work" {
		t.Fatalf("expected work, got %s", name)
	}
	if cfg == nil {
		t.Fatal("expected created profile config")
	}
	if pc.Profiles["work"] == nil {
		t.Fatal("expected missing env-selected profile to be created")
	}
}

// --- parseReminders tests ---

func TestParseReminders_Valid(t *testing.T) {
	tests := []struct {
		input   []string
		method  string
		trigger string
	}{
		{[]string{"DISPLAY:-PT10M"}, "DISPLAY", "-PT10M"},
		{[]string{"email:-P1D"}, "EMAIL", "-P1D"},
		{[]string{"display:PT15M"}, "DISPLAY", "PT15M"},
	}
	for _, tt := range tests {
		reminders, err := parseReminders(tt.input)
		if err != nil {
			t.Fatalf("parseReminders(%v) unexpected error: %v", tt.input, err)
		}
		if len(reminders) != 1 {
			t.Fatalf("expected 1 reminder, got %d", len(reminders))
		}
		if reminders[0]["method"] != tt.method {
			t.Errorf("expected method %q, got %q", tt.method, reminders[0]["method"])
		}
		if reminders[0]["trigger"] != tt.trigger {
			t.Errorf("expected trigger %q, got %q", tt.trigger, reminders[0]["trigger"])
		}
	}
}

func TestParseReminders_Multiple(t *testing.T) {
	reminders, err := parseReminders([]string{"DISPLAY:-PT10M", "EMAIL:-P1D"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reminders) != 2 {
		t.Fatalf("expected 2 reminders, got %d", len(reminders))
	}
	if reminders[0]["method"] != "DISPLAY" || reminders[1]["method"] != "EMAIL" {
		t.Errorf("unexpected methods: %v, %v", reminders[0]["method"], reminders[1]["method"])
	}
}

func TestParseReminders_InvalidFormat(t *testing.T) {
	badInputs := [][]string{
		{"DISPLAY"},
		{""},
		{":PT10M"},
		{"DISPLAY:"},
	}
	for _, input := range badInputs {
		_, err := parseReminders(input)
		if err == nil {
			t.Errorf("parseReminders(%v) expected error, got nil", input)
		}
	}
}

func TestParseReminders_InvalidMethod(t *testing.T) {
	_, err := parseReminders([]string{"PUSH:-PT10M"})
	if err == nil {
		t.Fatal("expected error for invalid method")
	}
	if !strings.Contains(err.Error(), "DISPLAY 또는 EMAIL") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSmoke_CalendarCreateEvent_ReminderFlag(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "calendar", "create-event", "--help")
	if err != nil {
		t.Fatalf("calendar create-event --help failed: %v", err)
	}
	if !strings.Contains(out, "--reminder") {
		t.Error("calendar create-event --help missing --reminder flag")
	}
	for _, flag := range []string{"--visibility", "--transparency"} {
		if !strings.Contains(out, flag) {
			t.Errorf("calendar create-event --help missing %s flag", flag)
		}
	}
}

func TestSmoke_MailSend_Flags(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "mail", "send", "--help")
	if err != nil {
		t.Fatalf("mail send --help failed: %v", err)
	}
	for _, flag := range []string{"--to", "--cc", "--bcc", "--subject", "--body", "--content-type"} {
		if !strings.Contains(out, flag) {
			t.Errorf("mail send --help missing %s flag", flag)
		}
	}
}

func TestSmoke_TaskCreate_Flags(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "task", "create", "--help")
	if err != nil {
		t.Fatalf("task create --help failed: %v", err)
	}
	for _, flag := range []string{"--title", "--description", "--due-date"} {
		if !strings.Contains(out, flag) {
			t.Errorf("task create --help missing %s flag", flag)
		}
	}
}

func TestSmoke_TaskUpdate_Flags(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "task", "update", "--help")
	if err != nil {
		t.Fatalf("task update --help failed: %v", err)
	}
	for _, flag := range []string{"--title", "--description", "--due-date"} {
		if !strings.Contains(out, flag) {
			t.Errorf("task update --help missing %s flag", flag)
		}
	}
}

func TestSmoke_ContactCreate_Flags(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "contact", "create", "--help")
	if err != nil {
		t.Fatalf("contact create --help failed: %v", err)
	}
	for _, flag := range []string{"--name", "--email", "--phone"} {
		if !strings.Contains(out, flag) {
			t.Errorf("contact create --help missing %s flag", flag)
		}
	}
}

func TestSmoke_ContactUpdate_Flags(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "contact", "update", "--help")
	if err != nil {
		t.Fatalf("contact update --help failed: %v", err)
	}
	for _, flag := range []string{"--name", "--email", "--phone"} {
		if !strings.Contains(out, flag) {
			t.Errorf("contact update --help missing %s flag", flag)
		}
	}
}
