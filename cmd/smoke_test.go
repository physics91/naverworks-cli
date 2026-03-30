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

	"github.com/spf13/cobra"
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
	for _, sub := range []string{
		"list", "get", "create", "update", "patch", "delete",
		"regenerate-secret", "send",
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
	for _, sub := range []string{
		"register", "list", "update", "patch", "delete",
		"add-members", "list-members", "remove-member",
	} {
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
		if !strings.Contains(out, sub) {
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
	for _, sub := range []string{
		"create", "list", "get", "delete",
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
	if !strings.Contains(err.Error(), "파일 읽기 실패") {
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
	for _, sub := range []string{"create-calendar", "get-calendar", "update-calendar", "delete-calendar",
		"get-personal", "update-personal", "remove-user", "update-event", "delete-event", "default"} {
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
	for _, sub := range []string{
		"create", "update", "delete",
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
	for _, sub := range []string{
		"send", "get", "delete", "list-folders", "get-folder", "list",
		"update", "unread-count", "get-attachment", "list-favorite-folders",
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
		if !strings.Contains(out, sub) {
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

// ─── Task Smoke Tests ───

func TestSmoke_TaskHelp(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "task", "--help")
	if err != nil {
		t.Fatalf("task --help failed: %v", err)
	}
	for _, sub := range []string{
		"list", "get", "create", "update", "delete", "list-categories",
		"create-category", "get-category", "update-category", "delete-category",
		"move", "complete", "incomplete", "complete-assignee", "incomplete-assignee",
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
	for _, sub := range []string{
		"list", "list-user", "get", "create", "update", "full-update", "delete", "force-delete",
		"upload-photo", "get-photo", "delete-photo",
		"list-tags", "list-user-tags",
		"custom-property", "tag",
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
		if !strings.Contains(out, sub) {
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
	for _, sub := range []string{"create", "get", "update", "patch", "delete", "create-user-tags"} {
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
	for _, sub := range []string{
		"create", "delete", "list-posts", "get-post",
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
	for _, sub := range []string{
		"status", "clock-in", "clock-out",
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
	for _, sub := range []string{
		"list", "list-all", "get", "list-categories", "get-category", "list-forms",
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
		if !strings.Contains(out, sub) {
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
		if !strings.Contains(out, sub) {
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
		if !strings.Contains(out, sub) {
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
		if !strings.Contains(out, sub) {
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
		if !strings.Contains(out, sub) {
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
		if !strings.Contains(out, sub) {
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
		if !strings.Contains(out, sub) {
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
		if !strings.Contains(out, sub) {
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
		if !strings.Contains(out, sub) {
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
		// Existing
		"info", "list", "get", "download", "upload", "mkdir", "delete",
		"trash-list", "trash-restore",
		// Task 5-1: File operations
		"copy", "rename", "move", "protect", "unprotect", "lock", "unlock",
		// Task 5-2: Revisions
		"revision",
		// Task 5-3: Trash delete, link, share
		"trash-delete", "link-setting", "link", "share",
		// Existing groups
		"shared", "group", "shared-folder",
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
		if !strings.Contains(out, sub) {
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
		if !strings.Contains(out, sub) {
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
	for _, sub := range []string{"get", "create", "update", "delete", "list-sub-folders"} {
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
