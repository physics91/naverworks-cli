package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

// TestScimClient_UsesCorrectBaseURLAndToken verifies that a Client configured
// with SCIM base URL sends requests to the correct path with the correct
// Authorization header. In production, SCIM client creation lives in
// cmd/helpers.go (buildScimClient) which is hard to unit-test directly, so
// this test validates the underlying mechanism.
func TestScimClient_UsesCorrectBaseURLAndToken(t *testing.T) {
	var gotAuth, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.WriteHeader(200)
		w.Write([]byte(`{"totalResults":0,"Resources":[]}`))
	}))
	defer srv.Close()

	// Simulate what buildScimClient does: create a client with SCIM base URL
	// and a long-lived token (no refresh function).
	scimToken := &auth.Token{
		AccessToken: "scim-long-lived-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
	}
	client := NewClient(srv.URL, scimToken, nil)
	svc := NewScimService(client)

	_, err := svc.ListUsers(1, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer scim-long-lived-token" {
		t.Errorf("expected SCIM token in Authorization header, got %q", gotAuth)
	}
	if gotPath != "/Users" {
		t.Errorf("expected /Users path, got %q", gotPath)
	}
}

// TestServiceMethods verifies HTTP method and path for representative CRUD
// methods across domains that previously had 0% test coverage.
func TestServiceMethods(t *testing.T) {
	tests := []struct {
		name       string
		call       func(client *Client)
		wantMethod string
		wantPath   string
	}{
		// ─── Approval ───
		{
			name:       "ApprovalService.CreateCategory",
			call:       func(c *Client) { NewApprovalService(c).CreateCategory([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/approval/categories",
		},
		{
			name:       "ApprovalService.ListLinkageCodes",
			call:       func(c *Client) { NewApprovalService(c).ListLinkageCodes("", 0) },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/linkage-codes",
		},
		{
			name:       "ApprovalService.DeleteLinkageCodeItem",
			call:       func(c *Client) { NewApprovalService(c).DeleteLinkageCodeItem("k1", "item1") },
			wantMethod: "DELETE",
			wantPath:   "/business-support/approval/linkage-codes/k1/linkage-code-items/item1",
		},

		// ─── Mail ───
		{
			name:       "MailService.PatchMail",
			call:       func(c *Client) { NewMailService(c).PatchMail("u1", "m1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/users/u1/mail/m1",
		},
		{
			name:       "MailService.ListFilters",
			call:       func(c *Client) { NewMailService(c).ListFilters("u1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/mail/filters",
		},
		{
			name:       "MailService.DeleteForwarding",
			call:       func(c *Client) { NewMailService(c).DeleteForwarding("u1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/mail/settings/forwarding",
		},

		// ─── Board ───
		{
			name:       "BoardService.CreateBoard",
			call:       func(c *Client) { NewBoardService(c).CreateBoard(map[string]interface{}{"name": "b"}) },
			wantMethod: "POST",
			wantPath:   "/boards",
		},
		{
			name:       "BoardService.CreateComment",
			call:       func(c *Client) { NewBoardService(c).CreateComment("b1", "p1", map[string]interface{}{"body": "hi"}) },
			wantMethod: "POST",
			wantPath:   "/boards/b1/posts/p1/comments",
		},
		{
			name:       "BoardService.DeleteCommentAttachment",
			call:       func(c *Client) { NewBoardService(c).DeleteCommentAttachment("b1", "p1", "c1", "a1") },
			wantMethod: "DELETE",
			wantPath:   "/boards/b1/posts/p1/comments/c1/attachments/a1",
		},

		// ─── Attendance ───
		{
			name:       "AttendanceService.CreateTimecard",
			call:       func(c *Client) { NewAttendanceService(c).CreateTimecard([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/attendance/timecards",
		},
		{
			name:       "AttendanceService.PatchAbsence",
			call:       func(c *Client) { NewAttendanceService(c).PatchAbsence("abs1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/business-support/attendance/absences/abs1",
		},

		// ─── Contact ───
		{
			name: "ContactService.FullUpdateContact",
			call: func(c *Client) {
				NewContactService(c).FullUpdateContact("c1", map[string]interface{}{"name": "x"})
			},
			wantMethod: "PUT",
			wantPath:   "/contacts/c1",
		},
		{
			name: "ContactService.CreatePhoto",
			call: func(c *Client) {
				NewContactService(c).CreatePhoto("c1", map[string]interface{}{"fileName": "photo.jpg"})
			},
			wantMethod: "POST",
			wantPath:   "/contacts/c1/photo",
		},
		{
			name:       "ContactService.DeleteTag",
			call:       func(c *Client) { NewContactService(c).DeleteTag("tag1") },
			wantMethod: "DELETE",
			wantPath:   "/contact-tags/tag1",
		},

		// ─── Note ───
		{
			name: "NoteService.PatchPost",
			call: func(c *Client) {
				NewNoteService(c).PatchPost("g1", "p1", map[string]interface{}{"title": "x"})
			},
			wantMethod: "PATCH",
			wantPath:   "/groups/g1/note/posts/p1",
		},
		{
			name: "NoteService.CreatePostAttachment",
			call: func(c *Client) {
				NewNoteService(c).CreatePostAttachment("g1", "p1", map[string]interface{}{"fileName": "a.txt"})
			},
			wantMethod: "POST",
			wantPath:   "/groups/g1/note/posts/p1/attachments",
		},

		// ─── HR ───
		{
			name:       "HRService.CreateExtensionProperty",
			call:       func(c *Client) { NewHRService(c).CreateExtensionProperty([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/human-resource/extension-properties",
		},
		{
			name:       "HRService.DeleteLeaveOfAbsence",
			call:       func(c *Client) { NewHRService(c).DeleteLeaveOfAbsence("loa1") },
			wantMethod: "DELETE",
			wantPath:   "/business-support/human-resource/leave-of-absences/loa1",
		},

		// ─── Task ───
		{
			name:       "TaskService.CompleteTask",
			call:       func(c *Client) { NewTaskService(c).CompleteTask("t1") },
			wantMethod: "POST",
			wantPath:   "/tasks/t1/complete",
		},
		{
			name: "TaskService.MoveTask",
			call: func(c *Client) {
				NewTaskService(c).MoveTask("u1", "t1", map[string]interface{}{"toCategoryId": "cat1"})
			},
			wantMethod: "POST",
			wantPath:   "/users/u1/tasks/t1/move",
		},

		// ─── Audit ───
		{
			name:       "AuditService.CreatePolicyGroup",
			call:       func(c *Client) { NewAuditService(c).CreatePolicyGroup([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/audits/policy-groups",
		},
		{
			name:       "AuditService.RemovePolicyGroupMember",
			call:       func(c *Client) { NewAuditService(c).RemovePolicyGroupMember("pg1", "u1") },
			wantMethod: "DELETE",
			wantPath:   "/audits/policy-groups/pg1/members/u1",
		},

		// ─── BusinessPlace ───
		{
			name: "BusinessPlaceService.CreateBusinessPlace",
			call: func(c *Client) {
				NewBusinessPlaceService(c).CreateBusinessPlace(map[string]interface{}{"businessPlaceName": "test"})
			},
			wantMethod: "POST",
			wantPath:   "/business-support/business-places",
		},
		{
			name:       "BusinessPlaceService.DeleteBusinessPlace",
			call:       func(c *Client) { NewBusinessPlaceService(c).DeleteBusinessPlace("bp1") },
			wantMethod: "DELETE",
			wantPath:   "/business-support/business-places/bp1",
		},

		// ─── Form ───
		{
			name:       "FormService.ListResponses",
			call:       func(c *Client) { NewFormService(c).ListResponses("f1", "", 0) },
			wantMethod: "GET",
			wantPath:   "/forms/f1/responses",
		},

		// ─── SCIM ───
		{
			name:       "ScimService.ListUsers",
			call:       func(c *Client) { NewScimService(c).ListUsers(0, 0, "") },
			wantMethod: "GET",
			wantPath:   "/Users",
		},
		{
			name: "ScimService.CreateGroup",
			call: func(c *Client) {
				NewScimService(c).CreateGroup(map[string]interface{}{"displayName": "test"})
			},
			wantMethod: "POST",
			wantPath:   "/Groups",
		},
		{
			name:       "ScimService.DeleteUser",
			call:       func(c *Client) { NewScimService(c).DeleteUser("u1") },
			wantMethod: "DELETE",
			wantPath:   "/Users/u1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				w.WriteHeader(200)
				w.Write([]byte("{}"))
			}))
			defer srv.Close()

			token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
			client := NewClient(srv.URL, token, nil)

			tt.call(client)

			if gotMethod != tt.wantMethod {
				t.Errorf("method = %s, want %s", gotMethod, tt.wantMethod)
			}
			if gotPath != tt.wantPath {
				t.Errorf("path = %s, want %s", gotPath, tt.wantPath)
			}
		})
	}
}

func TestTaskService_MoveTask_SendsExpectedBody(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		gotBody = string(body)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(srv.URL, token, nil)

	_, err := NewTaskService(client).MoveTask("u1", "t1", map[string]interface{}{"toCategoryId": "cat1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody != `{"toCategoryId":"cat1"}` {
		t.Fatalf("body = %s, want %s", gotBody, `{"toCategoryId":"cat1"}`)
	}
}
