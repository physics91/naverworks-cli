package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

const roomTestID = "09d3455a-841c-81c9-a3d0-6551e63ccb92"

func TestRoomPreviewsNeverLaunchBrowserOrLoadAPIToken(t *testing.T) {
	setupTestEnv(t)
	t.Setenv("NW_BROWSER_PATH", filepath.Join(t.TempDir(), "does-not-exist"))
	for _, args := range [][]string{
		{"note", "list-posts"}, {"note", "get-post", "4070000000208811422"},
		{"note", "create-post", "--title", "test", "--body", "<p>test</p>"},
		{"note", "update-post", "4070000000208811422", "--title", "test", "--body", "<p>test</p>"},
		{"note", "delete-post", "4070000000208811422"},
		{"task", "list"}, {"task", "get", "t-12d36f03-9450-4c8d-8263-ca14793d76ac"},
		{"task", "create", "--title", "test"},
		{"task", "update", "t-12d36f03-9450-4c8d-8263-ca14793d76ac", "--description", "updated"},
		{"task", "delete", "t-12d36f03-9450-4c8d-8263-ca14793d76ac"},
		{"task", "complete", "t-12d36f03-9450-4c8d-8263-ca14793d76ac"},
		{"task", "incomplete", "t-12d36f03-9450-4c8d-8263-ca14793d76ac"}, {"task", "list-categories"},
	} {
		args = append(args, "--room-id", roomTestID, "--dry-run")
		out, err := runCLI(t, args...)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatal(err)
		}
		if result["transport"] != "browser" || result["room_id"] != roomTestID || result["resolved"] != false {
			t.Fatalf("incorrect room preview: %s", out)
		}
	}
	if _, err := os.Stat(filepath.Join(testConfigDir(t), "browser")); !os.IsNotExist(err) {
		t.Fatalf("preview created a credential directory: %v", err)
	}
}

func TestAllPagesHonorsRequestedInitialCursor(t *testing.T) {
	setupTestEnv(t)
	cmd := &cobra.Command{Use: "test"}
	addListFlags(cmd)
	_ = cmd.Flags().Set("cursor", "requested-page")
	_ = cmd.Flags().Set("count", "2")
	_ = cmd.Flags().Set("all", "true")
	var first string
	err := runListCmd(cmd, []string{"id"}, "items", func(cursor string, count int) (*api.Response, error) {
		first = cursor
		return &api.Response{Body: []byte(`{"items":[]}`)}, nil
	})
	if err != nil || first != "requested-page" {
		t.Fatalf("initial cursor was ignored: %q, %v", first, err)
	}
}

func TestRoomCategoryPreviewRejectsInvalidPaging(t *testing.T) {
	setupTestEnv(t)
	for _, extra := range [][]string{{"--count", "101"}, {"--count", "-1"}, {"--cursor", "invalid"}} {
		args := append([]string{"task", "list-categories", "--room-id", roomTestID, "--dry-run"}, extra...)
		if _, err := runCLI(t, args...); err == nil {
			t.Fatalf("category paging accepted invalid input: %v", extra)
		}
	}
}

func TestRoomPreviewShowsAllAndInitialCursor(t *testing.T) {
	setupTestEnv(t)
	out, err := runCLI(t, "note", "list-posts", "--room-id", roomTestID, "--all", "--cursor", "3", "--count", "2", "--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	var plan struct {
		Pagination struct {
			AllRequested  bool   `json:"all"`
			InitialCursor string `json:"start_cursor"`
			Count         int
		}
	}
	if json.Unmarshal([]byte(out), &plan) != nil || !plan.Pagination.AllRequested || plan.Pagination.InitialCursor != "3" || plan.Pagination.Count != 2 {
		t.Fatalf("preview omitted actual paging behavior: %s", out)
	}
}

func TestRoomRejectsMalformedTaskUUIDBeforeBrowserLaunch(t *testing.T) {
	setupTestEnv(t)
	_, err := runCLI(t, "task", "get", "t-------------------------------------", "--room-id", roomTestID, "--dry-run")
	if err == nil {
		t.Fatal("accepted a malformed UUID")
	}
}

func TestRoomRejectsAmbiguousTargetAndUnscopedTaskData(t *testing.T) {
	setupTestEnv(t)
	for _, args := range [][]string{
		{"note", "list-posts", "group1", "--room-id", roomTestID},
		{"task", "list", "--room-id", "459730455"},
		{"task", "list", "--room-id", roomTestID, "--user-id", "me"},
		{"task", "create", "--room-id", roomTestID, "--data", `{"title":"test","projectId":"other"}`},
		{"task", "create", "--room-id", roomTestID, "--data", `{"title":"test"} trailing`},
		{"task", "create", "--room-id", roomTestID, "--data", `{"title":"test"}`, "--title", "different"},
		{"task", "create", "--room-id", roomTestID, "--title", "test", "--due-date", "invalid"},
		{"task", "create", "--title", "test", "--assignee-ids", "100"},
		{"task", "list", "--room-id", ""},
	} {
		args = append(args, "--dry-run")
		if _, err := runCLI(t, args...); err == nil {
			t.Fatalf("invalid command accepted: %v", args)
		}
	}
}

func TestRoomDispatchPreservesGroupAPI(t *testing.T) {
	home := setupTestEnv(t)
	writeTestConfig(t, home)
	t.Setenv("NW_BROWSER_PATH", "missing-browser")
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"posts":[]}`))
	}))
	defer server.Close()
	setAPIBaseURL(t, server.URL)
	_, err := runCLI(t, "note", "list-posts", "group1")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || paths[0] != "/groups/group1/note/posts" {
		t.Fatalf("API request changed: %v", paths)
	}
}

func TestRoomGenerateInputAndPlanFileDoNotOpenSession(t *testing.T) {
	setupTestEnv(t)
	t.Setenv("NW_BROWSER_PATH", "missing-browser")
	path := filepath.Join(t.TempDir(), "plan.json")
	out, err := runCLI(t, "task", "create", "--room-id", roomTestID, "--title", "test", "--generate-input", "--plan-out", path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"resolution_required": true`) {
		t.Fatalf("input concealed unresolved values: %s", out)
	}
	data, err := os.ReadFile(path)
	if err != nil || !json.Valid(data) {
		t.Fatalf("plan file: %v", err)
	}
}
