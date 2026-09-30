package web

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testRoomID = "09d3455a-841c-81c9-a3d0-6551e63ccb92"

func TestTaskRequestsCarryLoggedInIdentityOnlyToTaskOrigin(t *testing.T) {
	c := &Client{userID: "123"}
	headers := c.requestHeaders(taskOrigin)
	if headers["task-web-client-user-id"] != c.userID || headers["WM-TASK-SERVICE-VERSION"] != "2" {
		t.Fatal("task requests omit the authenticated client identity or service version")
	}
	if c.requestHeaders(talkOrigin)["task-web-client-user-id"] != "" {
		t.Fatal("task identity leaked to the note service")
	}
	if (&Client{}).requestHeaders(taskOrigin)["task-web-client-user-id"] != "" {
		t.Fatal("unauthenticated bootstrap sent a user identity")
	}
}

func TestNoteWritePreviewKeepsContextInFormOnly(t *testing.T) {
	for _, action := range []string{"create-post", "update-post", "delete-post"} {
		plan, err := Plan(Operation{Domain: "note", Action: action, RoomID: testRoomID, ResourceID: "4070000000208811422", Title: "test", Body: "body"})
		if err != nil {
			t.Fatal(err)
		}
		if len(plan["query_template"].(url.Values)) != 0 {
			t.Fatalf("%s preview adds a query the real form POST does not send", action)
		}
		if plan["body"].(map[string]any)["channelNo"] != "${channelNo}" {
			t.Fatal("form lost the room context")
		}
	}
}

func TestRequestBoundaryRejectsForeignHostsAndRedirectPaths(t *testing.T) {
	for _, tc := range []struct{ origin, method, path string }{
		{"https://evil.example", "GET", "/posts"},
		{"https://talk.worksmobile.com.evil.example", "GET", "/posts"},
		{talkOrigin, "GET", "https://evil.example/posts"},
		{talkOrigin, "GET", "//evil.example/posts"},
		{talkOrigin, "GET", "/\\evil.example/posts"},
		{talkOrigin, "GET", "/posts#fragment"},
		{talkOrigin, "CONNECT", "/posts"},
		{talkOrigin, "GET", "/posts\r\nheader:value"},
	} {
		if err := validateRequest(tc.origin, tc.method, tc.path); err == nil {
			t.Fatalf("accepted unsafe request: %+v", tc)
		}
	}
	if err := validateRequest(taskOrigin, "POST", "/rd/1/v2/project/p1/task/t1/taskStatus"); err != nil {
		t.Fatal(err)
	}
}

func TestMissingBrowserSessionDoesNotLaunchOrCreateCredentials(t *testing.T) {
	dir, err := SessionDir(filepath.Join(t.TempDir(), "config.json"), "default")
	if err != nil {
		t.Fatal(err)
	}
	_, err = (Browser{SessionDir: dir, Profile: "default", ExecPath: "missing-browser"}).Open(context.Background())
	if err == nil {
		t.Fatal("missing session accepted")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("session was created: %v", err)
	}
}

func TestNoteIDPreservesInt64BeyondJavaScriptPrecision(t *testing.T) {
	const id = "4070000000208811422"
	var p notePost
	if err := json.Unmarshal([]byte(`{"postNo":4070000000208811422,"postNoString":"4070000000208811422","title":"test"}`), &p); err != nil {
		t.Fatal(err)
	}
	post, err := p.normalized()
	if err != nil {
		t.Fatal(err)
	}
	if post.ID != id || p.PostNo.String() != id {
		t.Fatal("post ID lost precision")
	}
	p.PostNoString = ""
	post, err = p.normalized()
	if err != nil || post.ID != id {
		t.Fatalf("numeric fallback lost precision: %v", err)
	}
}

func TestTaskCursorCannotCrossRooms(t *testing.T) {
	data, _ := json.Marshal(taskCursor{RoomID: testRoomID, CategoryID: "category1", After: "opaque", Read: 2})
	encoded := base64.RawURLEncoding.EncodeToString(data)
	got, err := parseTaskCursor(encoded, testRoomID)
	if err != nil || got.After != "opaque" {
		t.Fatalf("valid cursor rejected: %v", err)
	}
	if _, err := parseTaskCursor(encoded, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Fatal("foreign room cursor accepted")
	}
	if _, err := parseTaskCursor("not-base64", testRoomID); err == nil {
		t.Fatal("malformed cursor accepted")
	}
}

func TestTaskInputRejectsInvalidAssignmentsAndDates(t *testing.T) {
	str := func(s string) *string { return &s }
	for _, in := range []TaskInput{
		{Title: str(" ")},
		{Title: str("test"), DueDate: str("2026-02-30")},
		{Title: str("test"), AssigneeIDs: []string{}},
		{Title: str("test"), AssigneeIDs: []string{"../user"}},
		{Title: str("test"), AssigneeIDs: []string{"123", "123"}},
		{Title: str("test"), StatusOption: str("SOMEONE")},
	} {
		if err := in.Validate(true); err == nil {
			t.Fatalf("invalid input accepted: %+v", in)
		}
	}
}

func TestTaskPreviewUsesCursorCategoryAndRejectsConflictingFilter(t *testing.T) {
	data, _ := json.Marshal(taskCursor{RoomID: testRoomID, CategoryID: "second-category", After: "opaque", Read: 2})
	op := Operation{Domain: "task", Action: "list", RoomID: testRoomID, Cursor: base64.RawURLEncoding.EncodeToString(data)}
	plan, err := Plan(op)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(plan["path_template"].(string), "/category/second-category/tasks") || plan["query_template"].(url.Values).Get("cursor") != "opaque" {
		t.Fatal("preview ignored the category encoded in the cursor")
	}
	op.CategoryID = "other-category"
	if _, err := Plan(op); err == nil {
		t.Fatal("conflicting cursor and category filter accepted")
	}
}

func TestOfflinePlanDeclaresRuntimeResolutionAndOwnStatusDefault(t *testing.T) {
	plan, err := Plan(Operation{Domain: "task", Action: "complete", RoomID: testRoomID, ResourceID: "t-12d36f03-9450-4c8d-8263-ca14793d76ac"})
	if err != nil {
		t.Fatal(err)
	}
	if plan["resolved"] != false {
		t.Fatal("offline plan claimed to resolve the live room")
	}
	if plan["body"].(map[string]any)["forceChangeTotalStatus"] != false {
		t.Fatal("completion changed all assignees without an explicit option")
	}
}
