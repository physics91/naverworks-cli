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
