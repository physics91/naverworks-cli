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
		{
			name:       "ApprovalService.ListUserDocuments",
			call:       func(c *Client) { NewApprovalService(c).ListUserDocuments("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/users/u1/documents",
		},
		{
			name:       "ApprovalService.ListDocuments",
			call:       func(c *Client) { NewApprovalService(c).ListDocuments("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/documents",
		},
		{
			name:       "ApprovalService.GetDocument",
			call:       func(c *Client) { NewApprovalService(c).GetDocument("doc1") },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/documents/doc1",
		},
		{
			name:       "ApprovalService.ListCategories",
			call:       func(c *Client) { NewApprovalService(c).ListCategories("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/categories",
		},
		{
			name:       "ApprovalService.GetCategory",
			call:       func(c *Client) { NewApprovalService(c).GetCategory("cat1") },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/categories/cat1",
		},
		{
			name:       "ApprovalService.ListDocumentForms",
			call:       func(c *Client) { NewApprovalService(c).ListDocumentForms("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/document-forms",
		},
		{
			name:       "ApprovalService.PatchCategory",
			call:       func(c *Client) { NewApprovalService(c).PatchCategory("cat1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/business-support/approval/categories/cat1",
		},
		{
			name:       "ApprovalService.DeleteCategory",
			call:       func(c *Client) { NewApprovalService(c).DeleteCategory("cat1") },
			wantMethod: "DELETE",
			wantPath:   "/business-support/approval/categories/cat1",
		},
		{
			name:       "ApprovalService.CreateUserDocument",
			call:       func(c *Client) { NewApprovalService(c).CreateUserDocument("u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/approval/users/u1/documents",
		},
		{
			name:       "ApprovalService.CreateImportedDocument",
			call:       func(c *Client) { NewApprovalService(c).CreateImportedDocument([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/approval/imported-documents",
		},
		{
			name:       "ApprovalService.CreateDocumentLink",
			call:       func(c *Client) { NewApprovalService(c).CreateDocumentLink("u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/approval/users/u1/documents/create-document-link",
		},
		{
			name:       "ApprovalService.GetDocumentForm",
			call:       func(c *Client) { NewApprovalService(c).GetDocumentForm("df1") },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/document-forms/df1",
		},
		{
			name:       "ApprovalService.CreateUserDocumentAttachment",
			call:       func(c *Client) { NewApprovalService(c).CreateUserDocumentAttachment("u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/approval/users/u1/documents/attachments",
		},
		{
			name:       "ApprovalService.CreateImportedDocumentAttachment",
			call:       func(c *Client) { NewApprovalService(c).CreateImportedDocumentAttachment([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/approval/imported-documents/attachments",
		},
		{
			name:       "ApprovalService.CreateLinkageCode",
			call:       func(c *Client) { NewApprovalService(c).CreateLinkageCode([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/approval/linkage-codes",
		},
		{
			name:       "ApprovalService.GetLinkageCode",
			call:       func(c *Client) { NewApprovalService(c).GetLinkageCode("k1") },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/linkage-codes/k1",
		},
		{
			name:       "ApprovalService.PatchLinkageCode",
			call:       func(c *Client) { NewApprovalService(c).PatchLinkageCode("k1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/business-support/approval/linkage-codes/k1",
		},
		{
			name:       "ApprovalService.CreateLinkageCodeItem",
			call:       func(c *Client) { NewApprovalService(c).CreateLinkageCodeItem("k1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/approval/linkage-codes/k1/linkage-code-items",
		},
		{
			name:       "ApprovalService.ListLinkageCodeItems",
			call:       func(c *Client) { NewApprovalService(c).ListLinkageCodeItems("k1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/linkage-codes/k1/linkage-code-items",
		},
		{
			name:       "ApprovalService.GetLinkageCodeItem",
			call:       func(c *Client) { NewApprovalService(c).GetLinkageCodeItem("k1", "item1") },
			wantMethod: "GET",
			wantPath:   "/business-support/approval/linkage-codes/k1/linkage-code-items/item1",
		},
		{
			name:       "ApprovalService.PatchLinkageCodeItem",
			call:       func(c *Client) { NewApprovalService(c).PatchLinkageCodeItem("k1", "item1", []byte(`{}`)) },
			wantMethod: "PATCH",
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
		{
			name:       "AttendanceService.GetStatus",
			call:       func(c *Client) { NewAttendanceService(c).GetStatus("u1") },
			wantMethod: "GET",
			wantPath:   "/business-support/attendance/users/u1/status",
		},
		{
			name:       "AttendanceService.ClockIn",
			call:       func(c *Client) { NewAttendanceService(c).ClockIn("u1", "2024-01-01", "09:00") },
			wantMethod: "POST",
			wantPath:   "/business-support/attendance/users/u1/clock-in",
		},
		{
			name:       "AttendanceService.ClockOut",
			call:       func(c *Client) { NewAttendanceService(c).ClockOut("u1", "2024-01-01", "18:00") },
			wantMethod: "POST",
			wantPath:   "/business-support/attendance/users/u1/clock-out",
		},
		{
			name:       "AttendanceService.ListAbsences",
			call:       func(c *Client) { NewAttendanceService(c).ListAbsences("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/attendance/absences",
		},
		{
			name:       "AttendanceService.ListAnnualLeaves",
			call:       func(c *Client) { NewAttendanceService(c).ListAnnualLeaves("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/attendance/annual-leaves",
		},
		{
			name:       "AttendanceService.ListTimecards",
			call:       func(c *Client) { NewAttendanceService(c).ListTimecards("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/attendance/timecards",
		},
		{
			name:       "AttendanceService.GetTimecard",
			call:       func(c *Client) { NewAttendanceService(c).GetTimecard("tc1") },
			wantMethod: "GET",
			wantPath:   "/business-support/attendance/timecards/tc1",
		},
		{
			name:       "AttendanceService.PatchTimecard",
			call:       func(c *Client) { NewAttendanceService(c).PatchTimecard("tc1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/business-support/attendance/timecards/tc1",
		},
		{
			name:       "AttendanceService.AdjustAnnualLeave",
			call:       func(c *Client) { NewAttendanceService(c).AdjustAnnualLeave([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/attendance/annual-leaves/adjust",
		},
		{
			name:       "AttendanceService.ListAbsenceSchedules",
			call:       func(c *Client) { NewAttendanceService(c).ListAbsenceSchedules("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/attendance/absence-schedule",
		},
		{
			name:       "AttendanceService.CreateAbsence",
			call:       func(c *Client) { NewAttendanceService(c).CreateAbsence([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/attendance/absences",
		},
		{
			name:       "AttendanceService.GetAbsence",
			call:       func(c *Client) { NewAttendanceService(c).GetAbsence("abs1") },
			wantMethod: "GET",
			wantPath:   "/business-support/attendance/absences/abs1",
		},
		{
			name:       "AttendanceService.DeleteAbsence",
			call:       func(c *Client) { NewAttendanceService(c).DeleteAbsence("abs1") },
			wantMethod: "DELETE",
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
		{
			name:       "ContactService.ListContacts",
			call:       func(c *Client) { NewContactService(c).ListContacts("", 10) },
			wantMethod: "GET",
			wantPath:   "/contacts",
		},
		{
			name:       "ContactService.ListUserContacts",
			call:       func(c *Client) { NewContactService(c).ListUserContacts("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/contacts",
		},
		{
			name:       "ContactService.GetContact",
			call:       func(c *Client) { NewContactService(c).GetContact("c1") },
			wantMethod: "GET",
			wantPath:   "/contacts/c1",
		},
		{
			name:       "ContactService.CreateContact",
			call:       func(c *Client) { NewContactService(c).CreateContact(map[string]interface{}{"name": "x"}) },
			wantMethod: "POST",
			wantPath:   "/contacts",
		},
		{
			name:       "ContactService.UpdateContact",
			call:       func(c *Client) { NewContactService(c).UpdateContact("c1", map[string]interface{}{"name": "x"}) },
			wantMethod: "PATCH",
			wantPath:   "/contacts/c1",
		},
		{
			name:       "ContactService.DeleteContact",
			call:       func(c *Client) { NewContactService(c).DeleteContact("c1") },
			wantMethod: "DELETE",
			wantPath:   "/contacts/c1",
		},
		{
			name:       "ContactService.ListTags",
			call:       func(c *Client) { NewContactService(c).ListTags("", 10) },
			wantMethod: "GET",
			wantPath:   "/contact-tags",
		},
		{
			name:       "ContactService.ListUserTags",
			call:       func(c *Client) { NewContactService(c).ListUserTags("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/contact-tags",
		},
		{
			name:       "ContactService.ForceDeleteContact",
			call:       func(c *Client) { NewContactService(c).ForceDeleteContact("c1") },
			wantMethod: "DELETE",
			wantPath:   "/contacts/c1/forcedelete",
		},
		{
			name:       "ContactService.DeletePhoto",
			call:       func(c *Client) { NewContactService(c).DeletePhoto("c1") },
			wantMethod: "DELETE",
			wantPath:   "/contacts/c1/photo",
		},
		{
			name:       "ContactService.CreateCustomProperty",
			call:       func(c *Client) { NewContactService(c).CreateCustomProperty(map[string]interface{}{"name": "x"}) },
			wantMethod: "POST",
			wantPath:   "/contacts/custom-properties",
		},
		{
			name:       "ContactService.ListCustomProperties",
			call:       func(c *Client) { NewContactService(c).ListCustomProperties("", 10) },
			wantMethod: "GET",
			wantPath:   "/contacts/custom-properties",
		},
		{
			name:       "ContactService.GetCustomProperty",
			call:       func(c *Client) { NewContactService(c).GetCustomProperty("cp1") },
			wantMethod: "GET",
			wantPath:   "/contacts/custom-properties/cp1",
		},
		{
			name:       "ContactService.PatchCustomProperty",
			call:       func(c *Client) { NewContactService(c).PatchCustomProperty("cp1", map[string]interface{}{"name": "y"}) },
			wantMethod: "PATCH",
			wantPath:   "/contacts/custom-properties/cp1",
		},
		{
			name:       "ContactService.DeleteCustomProperty",
			call:       func(c *Client) { NewContactService(c).DeleteCustomProperty("cp1") },
			wantMethod: "DELETE",
			wantPath:   "/contacts/custom-properties/cp1",
		},
		{
			name:       "ContactService.CreateTag",
			call:       func(c *Client) { NewContactService(c).CreateTag(map[string]interface{}{"name": "t"}) },
			wantMethod: "POST",
			wantPath:   "/contact-tags",
		},
		{
			name:       "ContactService.GetTag",
			call:       func(c *Client) { NewContactService(c).GetTag("tag1") },
			wantMethod: "GET",
			wantPath:   "/contact-tags/tag1",
		},
		{
			name:       "ContactService.UpdateTag",
			call:       func(c *Client) { NewContactService(c).UpdateTag("tag1", map[string]interface{}{"name": "t2"}) },
			wantMethod: "PUT",
			wantPath:   "/contact-tags/tag1",
		},
		{
			name:       "ContactService.PatchTag",
			call:       func(c *Client) { NewContactService(c).PatchTag("tag1", map[string]interface{}{"name": "t3"}) },
			wantMethod: "PATCH",
			wantPath:   "/contact-tags/tag1",
		},
		{
			name:       "ContactService.CreateUserTags",
			call:       func(c *Client) { NewContactService(c).CreateUserTags("u1", map[string]interface{}{"tagIds": []string{"t1"}}) },
			wantMethod: "POST",
			wantPath:   "/users/u1/contact-tags",
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

		// ─── Mail (remaining) ───
		{
			name:       "MailService.SendMail",
			call:       func(c *Client) { NewMailService(c).SendMail("u1", map[string]interface{}{"to": "x"}) },
			wantMethod: "POST",
			wantPath:   "/users/u1/mail",
		},
		{
			name:       "MailService.GetMail",
			call:       func(c *Client) { NewMailService(c).GetMail("u1", "m1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/mail/m1",
		},
		{
			name:       "MailService.DeleteMail",
			call:       func(c *Client) { NewMailService(c).DeleteMail("u1", "m1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/mail/m1",
		},
		{
			name:       "MailService.ListFolders",
			call:       func(c *Client) { NewMailService(c).ListFolders("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/mail/mailfolders",
		},
		{
			name:       "MailService.GetFolder",
			call:       func(c *Client) { NewMailService(c).GetFolder("u1", "f1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/mail/mailfolders/f1",
		},
		{
			name:       "MailService.ListMails",
			call:       func(c *Client) { NewMailService(c).ListMails("u1", "f1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/mail/mailfolders/f1/children",
		},
		{
			name:       "MailService.GetUnreadCount",
			call:       func(c *Client) { NewMailService(c).GetUnreadCount("u1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/mail/unread-count",
		},
		{
			name:       "MailService.GetAttachment",
			call:       func(c *Client) { NewMailService(c).GetAttachment("u1", "m1", "a1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/mail/m1/attachments/a1",
		},
		{
			name:       "MailService.ListFavoriteContactsFolders",
			call:       func(c *Client) { NewMailService(c).ListFavoriteContactsFolders("u1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/mail/mailfolders/favorite-contacts",
		},
		{
			name:       "MailService.CreateMailFolder",
			call:       func(c *Client) { NewMailService(c).CreateMailFolder("u1", map[string]interface{}{"name": "f"}) },
			wantMethod: "POST",
			wantPath:   "/users/u1/mail/mailfolders",
		},
		{
			name:       "MailService.UpdateMailFolder",
			call:       func(c *Client) { NewMailService(c).UpdateMailFolder("u1", "f1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/users/u1/mail/mailfolders/f1",
		},
		{
			name:       "MailService.DeleteMailFolder",
			call:       func(c *Client) { NewMailService(c).DeleteMailFolder("u1", "f1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/mail/mailfolders/f1",
		},
		{
			name:       "MailService.CreateFilter",
			call:       func(c *Client) { NewMailService(c).CreateFilter("u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/mail/filters",
		},
		{
			name:       "MailService.GetFilter",
			call:       func(c *Client) { NewMailService(c).GetFilter("u1", "flt1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/mail/filters/flt1",
		},
		{
			name:       "MailService.DeleteFilter",
			call:       func(c *Client) { NewMailService(c).DeleteFilter("u1", "flt1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/mail/filters/flt1",
		},
		{
			name:       "MailService.CreateImapMigration",
			call:       func(c *Client) { NewMailService(c).CreateImapMigration("u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/mail/migration/imap",
		},
		{
			name:       "MailService.GetImapMigration",
			call:       func(c *Client) { NewMailService(c).GetImapMigration("u1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/mail/migration/imap",
		},
		{
			name:       "MailService.DeleteImapMigration",
			call:       func(c *Client) { NewMailService(c).DeleteImapMigration("u1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/mail/migration/imap",
		},
		{
			name:       "MailService.CreatePop3Migration",
			call:       func(c *Client) { NewMailService(c).CreatePop3Migration("u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/mail/migration/pop3",
		},
		{
			name:       "MailService.CreateForwarding",
			call:       func(c *Client) { NewMailService(c).CreateForwarding("u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/mail/settings/forwarding",
		},

		// ─── Board (remaining) ───
		{
			name:       "BoardService.ListBoards",
			call:       func(c *Client) { NewBoardService(c).ListBoards("", 10) },
			wantMethod: "GET",
			wantPath:   "/boards",
		},
		{
			name:       "BoardService.GetBoard",
			call:       func(c *Client) { NewBoardService(c).GetBoard("b1") },
			wantMethod: "GET",
			wantPath:   "/boards/b1",
		},
		{
			name:       "BoardService.ListPosts",
			call:       func(c *Client) { NewBoardService(c).ListPosts("b1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/boards/b1/posts",
		},
		{
			name:       "BoardService.GetPost",
			call:       func(c *Client) { NewBoardService(c).GetPost("b1", "p1") },
			wantMethod: "GET",
			wantPath:   "/boards/b1/posts/p1",
		},
		{
			name:       "BoardService.CreatePost",
			call:       func(c *Client) { NewBoardService(c).CreatePost("b1", map[string]interface{}{"body": "hi"}) },
			wantMethod: "POST",
			wantPath:   "/boards/b1/posts",
		},
		{
			name:       "BoardService.UpdatePost",
			call:       func(c *Client) { NewBoardService(c).UpdatePost("b1", "p1", map[string]interface{}{"body": "up"}) },
			wantMethod: "PUT",
			wantPath:   "/boards/b1/posts/p1",
		},
		{
			name:       "BoardService.DeletePost",
			call:       func(c *Client) { NewBoardService(c).DeletePost("b1", "p1") },
			wantMethod: "DELETE",
			wantPath:   "/boards/b1/posts/p1",
		},
		{
			name:       "BoardService.ListComments",
			call:       func(c *Client) { NewBoardService(c).ListComments("b1", "p1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/boards/b1/posts/p1/comments",
		},
		{
			name:       "BoardService.UpdateBoard",
			call:       func(c *Client) { NewBoardService(c).UpdateBoard("b1", map[string]interface{}{"name": "x"}) },
			wantMethod: "PUT",
			wantPath:   "/boards/b1",
		},
		{
			name:       "BoardService.DeleteBoard",
			call:       func(c *Client) { NewBoardService(c).DeleteBoard("b1") },
			wantMethod: "DELETE",
			wantPath:   "/boards/b1",
		},
		{
			name:       "BoardService.ListPostReaders",
			call:       func(c *Client) { NewBoardService(c).ListPostReaders("b1", "p1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/boards/b1/posts/p1/readers",
		},
		{
			name:       "BoardService.ListRecentPosts",
			call:       func(c *Client) { NewBoardService(c).ListRecentPosts("", 10) },
			wantMethod: "GET",
			wantPath:   "/boards/recent/posts",
		},
		{
			name:       "BoardService.ListMyPosts",
			call:       func(c *Client) { NewBoardService(c).ListMyPosts("", 10) },
			wantMethod: "GET",
			wantPath:   "/boards/my/posts",
		},
		{
			name:       "BoardService.ListMustPosts",
			call:       func(c *Client) { NewBoardService(c).ListMustPosts("", 10) },
			wantMethod: "GET",
			wantPath:   "/boards/must/posts",
		},
		{
			name:       "BoardService.CreatePostAttachment",
			call:       func(c *Client) { NewBoardService(c).CreatePostAttachment("b1", "p1", map[string]interface{}{"fileName": "a.txt"}) },
			wantMethod: "POST",
			wantPath:   "/boards/b1/posts/p1/attachments",
		},
		{
			name:       "BoardService.ListPostAttachments",
			call:       func(c *Client) { NewBoardService(c).ListPostAttachments("b1", "p1") },
			wantMethod: "GET",
			wantPath:   "/boards/b1/posts/p1/attachments",
		},
		{
			name:       "BoardService.GetPostAttachment",
			call:       func(c *Client) { NewBoardService(c).GetPostAttachment("b1", "p1", "a1") },
			wantMethod: "GET",
			wantPath:   "/boards/b1/posts/p1/attachments/a1",
		},
		{
			name:       "BoardService.DeletePostAttachment",
			call:       func(c *Client) { NewBoardService(c).DeletePostAttachment("b1", "p1", "a1") },
			wantMethod: "DELETE",
			wantPath:   "/boards/b1/posts/p1/attachments/a1",
		},
		{
			name:       "BoardService.GetComment",
			call:       func(c *Client) { NewBoardService(c).GetComment("b1", "p1", "c1") },
			wantMethod: "GET",
			wantPath:   "/boards/b1/posts/p1/comments/c1",
		},
		{
			name:       "BoardService.UpdateComment",
			call:       func(c *Client) { NewBoardService(c).UpdateComment("b1", "p1", "c1", map[string]interface{}{"body": "up"}) },
			wantMethod: "PUT",
			wantPath:   "/boards/b1/posts/p1/comments/c1",
		},
		{
			name:       "BoardService.DeleteComment",
			call:       func(c *Client) { NewBoardService(c).DeleteComment("b1", "p1", "c1") },
			wantMethod: "DELETE",
			wantPath:   "/boards/b1/posts/p1/comments/c1",
		},
		{
			name:       "BoardService.CreateCommentAttachment",
			call:       func(c *Client) { NewBoardService(c).CreateCommentAttachment("b1", "p1", "c1", map[string]interface{}{"fileName": "a.txt"}) },
			wantMethod: "POST",
			wantPath:   "/boards/b1/posts/p1/comments/c1/attachments",
		},
		{
			name:       "BoardService.ListCommentAttachments",
			call:       func(c *Client) { NewBoardService(c).ListCommentAttachments("b1", "p1", "c1") },
			wantMethod: "GET",
			wantPath:   "/boards/b1/posts/p1/comments/c1/attachments",
		},
		{
			name:       "BoardService.GetCommentAttachment",
			call:       func(c *Client) { NewBoardService(c).GetCommentAttachment("b1", "p1", "c1", "a1") },
			wantMethod: "GET",
			wantPath:   "/boards/b1/posts/p1/comments/c1/attachments/a1",
		},

		// ─── Note (remaining) ───
		{
			name:       "NoteService.CreateNote",
			call:       func(c *Client) { NewNoteService(c).CreateNote("g1") },
			wantMethod: "POST",
			wantPath:   "/groups/g1/note",
		},
		{
			name:       "NoteService.DeleteNote",
			call:       func(c *Client) { NewNoteService(c).DeleteNote("g1") },
			wantMethod: "DELETE",
			wantPath:   "/groups/g1/note",
		},
		{
			name:       "NoteService.ListPosts",
			call:       func(c *Client) { NewNoteService(c).ListPosts("g1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/groups/g1/note/posts",
		},
		{
			name:       "NoteService.GetPost",
			call:       func(c *Client) { NewNoteService(c).GetPost("g1", "p1") },
			wantMethod: "GET",
			wantPath:   "/groups/g1/note/posts/p1",
		},
		{
			name:       "NoteService.CreatePost",
			call:       func(c *Client) { NewNoteService(c).CreatePost("g1", map[string]interface{}{"title": "x"}) },
			wantMethod: "POST",
			wantPath:   "/groups/g1/note/posts",
		},
		{
			name:       "NoteService.UpdatePost",
			call:       func(c *Client) { NewNoteService(c).UpdatePost("g1", "p1", map[string]interface{}{"title": "y"}) },
			wantMethod: "PUT",
			wantPath:   "/groups/g1/note/posts/p1",
		},
		{
			name:       "NoteService.DeletePost",
			call:       func(c *Client) { NewNoteService(c).DeletePost("g1", "p1") },
			wantMethod: "DELETE",
			wantPath:   "/groups/g1/note/posts/p1",
		},
		{
			name:       "NoteService.ListPostAttachments",
			call:       func(c *Client) { NewNoteService(c).ListPostAttachments("g1", "p1") },
			wantMethod: "GET",
			wantPath:   "/groups/g1/note/posts/p1/attachments",
		},
		{
			name:       "NoteService.GetPostAttachment",
			call:       func(c *Client) { NewNoteService(c).GetPostAttachment("g1", "p1", "a1") },
			wantMethod: "GET",
			wantPath:   "/groups/g1/note/posts/p1/attachments/a1",
		},
		{
			name:       "NoteService.DeletePostAttachment",
			call:       func(c *Client) { NewNoteService(c).DeletePostAttachment("g1", "p1", "a1") },
			wantMethod: "DELETE",
			wantPath:   "/groups/g1/note/posts/p1/attachments/a1",
		},

		// ─── HR (remaining) ───
		{
			name:       "HRService.ListExtensionProperties",
			call:       func(c *Client) { NewHRService(c).ListExtensionProperties("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/human-resource/extension-properties",
		},
		{
			name:       "HRService.GetUserExtensionProperties",
			call:       func(c *Client) { NewHRService(c).GetUserExtensionProperties("u1") },
			wantMethod: "GET",
			wantPath:   "/business-support/human-resource/user/u1/extension-properties",
		},
		{
			name:       "HRService.ListLeaveOfAbsences",
			call:       func(c *Client) { NewHRService(c).ListLeaveOfAbsences("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/human-resource/leave-of-absences",
		},
		{
			name:       "HRService.ListOnLeaveUsers",
			call:       func(c *Client) { NewHRService(c).ListOnLeaveUsers("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/human-resource/on-leave-users",
		},
		{
			name:       "HRService.GetExtensionProperty",
			call:       func(c *Client) { NewHRService(c).GetExtensionProperty("ep1") },
			wantMethod: "GET",
			wantPath:   "/business-support/human-resource/extension-properties/ep1",
		},
		{
			name:       "HRService.PatchExtensionProperty",
			call:       func(c *Client) { NewHRService(c).PatchExtensionProperty("ep1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/business-support/human-resource/extension-properties/ep1",
		},
		{
			name:       "HRService.DeleteExtensionProperty",
			call:       func(c *Client) { NewHRService(c).DeleteExtensionProperty("ep1") },
			wantMethod: "DELETE",
			wantPath:   "/business-support/human-resource/extension-properties/ep1",
		},
		{
			name:       "HRService.GetUserExtensionProperty",
			call:       func(c *Client) { NewHRService(c).GetUserExtensionProperty("u1", "ep1") },
			wantMethod: "GET",
			wantPath:   "/business-support/human-resource/user/u1/extension-properties/ep1",
		},
		{
			name:       "HRService.PatchUserExtensionProperty",
			call:       func(c *Client) { NewHRService(c).PatchUserExtensionProperty("u1", "ep1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/business-support/human-resource/user/u1/extension-properties/ep1",
		},
		{
			name:       "HRService.CreateLeaveOfAbsence",
			call:       func(c *Client) { NewHRService(c).CreateLeaveOfAbsence([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/business-support/human-resource/leave-of-absences",
		},
		{
			name:       "HRService.GetLeaveOfAbsence",
			call:       func(c *Client) { NewHRService(c).GetLeaveOfAbsence("loa1") },
			wantMethod: "GET",
			wantPath:   "/business-support/human-resource/leave-of-absences/loa1",
		},
		{
			name:       "HRService.PatchLeaveOfAbsence",
			call:       func(c *Client) { NewHRService(c).PatchLeaveOfAbsence("loa1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/business-support/human-resource/leave-of-absences/loa1",
		},

		// ─── Task (remaining) ───
		{
			name:       "TaskService.ListTasks",
			call:       func(c *Client) { NewTaskService(c).ListTasks("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/tasks",
		},
		{
			name:       "TaskService.GetTask",
			call:       func(c *Client) { NewTaskService(c).GetTask("t1") },
			wantMethod: "GET",
			wantPath:   "/tasks/t1",
		},
		{
			name:       "TaskService.CreateTask",
			call:       func(c *Client) { NewTaskService(c).CreateTask("u1", map[string]interface{}{"title": "x"}) },
			wantMethod: "POST",
			wantPath:   "/users/u1/tasks",
		},
		{
			name:       "TaskService.UpdateTask",
			call:       func(c *Client) { NewTaskService(c).UpdateTask("t1", map[string]interface{}{"title": "y"}) },
			wantMethod: "PATCH",
			wantPath:   "/tasks/t1",
		},
		{
			name:       "TaskService.DeleteTask",
			call:       func(c *Client) { NewTaskService(c).DeleteTask("t1") },
			wantMethod: "DELETE",
			wantPath:   "/tasks/t1",
		},
		{
			name:       "TaskService.ListCategories",
			call:       func(c *Client) { NewTaskService(c).ListCategories("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/task-categories",
		},
		{
			name:       "TaskService.CreateCategory",
			call:       func(c *Client) { NewTaskService(c).CreateCategory("u1", map[string]interface{}{"name": "c"}) },
			wantMethod: "POST",
			wantPath:   "/users/u1/task-categories",
		},
		{
			name:       "TaskService.GetCategory",
			call:       func(c *Client) { NewTaskService(c).GetCategory("u1", "cat1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/task-categories/cat1",
		},
		{
			name:       "TaskService.PatchCategory",
			call:       func(c *Client) { NewTaskService(c).PatchCategory("u1", "cat1", map[string]interface{}{"name": "c2"}) },
			wantMethod: "PATCH",
			wantPath:   "/users/u1/task-categories/cat1",
		},
		{
			name:       "TaskService.DeleteCategory",
			call:       func(c *Client) { NewTaskService(c).DeleteCategory("u1", "cat1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/task-categories/cat1",
		},
		{
			name:       "TaskService.IncompleteTask",
			call:       func(c *Client) { NewTaskService(c).IncompleteTask("t1") },
			wantMethod: "POST",
			wantPath:   "/tasks/t1/incomplete",
		},
		{
			name:       "TaskService.CompleteAssigneeTask",
			call:       func(c *Client) { NewTaskService(c).CompleteAssigneeTask("t1", "u1") },
			wantMethod: "POST",
			wantPath:   "/tasks/t1/assignees/u1/complete",
		},
		{
			name:       "TaskService.IncompleteAssigneeTask",
			call:       func(c *Client) { NewTaskService(c).IncompleteAssigneeTask("t1", "u1") },
			wantMethod: "POST",
			wantPath:   "/tasks/t1/assignees/u1/incomplete",
		},

		// ─── Audit (remaining) ───
		{
			name:       "AuditService.ListPolicyGroups",
			call:       func(c *Client) { NewAuditService(c).ListPolicyGroups("", 10) },
			wantMethod: "GET",
			wantPath:   "/audits/policy-groups",
		},
		{
			name:       "AuditService.GetPolicyGroup",
			call:       func(c *Client) { NewAuditService(c).GetPolicyGroup("pg1") },
			wantMethod: "GET",
			wantPath:   "/audits/policy-groups/pg1",
		},
		{
			name:       "AuditService.UpdatePolicyGroup",
			call:       func(c *Client) { NewAuditService(c).UpdatePolicyGroup("pg1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/audits/policy-groups/pg1",
		},
		{
			name:       "AuditService.DeletePolicyGroup",
			call:       func(c *Client) { NewAuditService(c).DeletePolicyGroup("pg1") },
			wantMethod: "DELETE",
			wantPath:   "/audits/policy-groups/pg1",
		},
		{
			name:       "AuditService.AddPolicyGroupMembers",
			call:       func(c *Client) { NewAuditService(c).AddPolicyGroupMembers("pg1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/audits/policy-groups/pg1/members",
		},
		{
			name:       "AuditService.ListPolicyGroupMembers",
			call:       func(c *Client) { NewAuditService(c).ListPolicyGroupMembers("pg1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/audits/policy-groups/pg1/members",
		},

		// ─── BusinessPlace (remaining) ───
		{
			name:       "BusinessPlaceService.ListBusinessPlaces",
			call:       func(c *Client) { NewBusinessPlaceService(c).ListBusinessPlaces("", 10) },
			wantMethod: "GET",
			wantPath:   "/business-support/business-places",
		},
		{
			name:       "BusinessPlaceService.GetBusinessPlace",
			call:       func(c *Client) { NewBusinessPlaceService(c).GetBusinessPlace("bp1") },
			wantMethod: "GET",
			wantPath:   "/business-support/business-places/bp1",
		},
		{
			name:       "BusinessPlaceService.UpdateBusinessPlace",
			call:       func(c *Client) { NewBusinessPlaceService(c).UpdateBusinessPlace("bp1", map[string]interface{}{"name": "x"}) },
			wantMethod: "PATCH",
			wantPath:   "/business-support/business-places/bp1",
		},

		// ─── Calendar ───
		{
			name:       "CalendarService.ListCalendars",
			call:       func(c *Client) { NewCalendarService(c).ListCalendars("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/calendar-personals",
		},
		{
			name:       "CalendarService.GetDefaultCalendar",
			call:       func(c *Client) { NewCalendarService(c).GetDefaultCalendar("u1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/calendar",
		},
		{
			name:       "CalendarService.ListEvents",
			call:       func(c *Client) { NewCalendarService(c).ListEvents("u1", "cal1", "2024-01-01", "2024-01-31") },
			wantMethod: "GET",
			wantPath:   "/users/u1/calendars/cal1/events",
		},
		{
			name:       "CalendarService.GetEvent",
			call:       func(c *Client) { NewCalendarService(c).GetEvent("u1", "cal1", "ev1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/calendars/cal1/events/ev1",
		},
		{
			name:       "CalendarService.CreateEvent",
			call:       func(c *Client) { NewCalendarService(c).CreateEvent("u1", "cal1", map[string]interface{}{"summary": "x"}) },
			wantMethod: "POST",
			wantPath:   "/users/u1/calendars/cal1/events",
		},
		{
			name:       "CalendarService.CreateCalendar",
			call:       func(c *Client) { NewCalendarService(c).CreateCalendar(map[string]interface{}{"name": "c"}) },
			wantMethod: "POST",
			wantPath:   "/calendars",
		},
		{
			name:       "CalendarService.GetCalendar",
			call:       func(c *Client) { NewCalendarService(c).GetCalendar("cal1") },
			wantMethod: "GET",
			wantPath:   "/calendars/cal1",
		},
		{
			name:       "CalendarService.PatchCalendar",
			call:       func(c *Client) { NewCalendarService(c).PatchCalendar("cal1", map[string]interface{}{"name": "c2"}) },
			wantMethod: "PATCH",
			wantPath:   "/calendars/cal1",
		},
		{
			name:       "CalendarService.DeleteCalendar",
			call:       func(c *Client) { NewCalendarService(c).DeleteCalendar("cal1") },
			wantMethod: "DELETE",
			wantPath:   "/calendars/cal1",
		},
		{
			name:       "CalendarService.GetCalendarPersonal",
			call:       func(c *Client) { NewCalendarService(c).GetCalendarPersonal("u1", "cal1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/calendar-personals/cal1",
		},
		{
			name:       "CalendarService.PatchCalendarPersonal",
			call:       func(c *Client) { NewCalendarService(c).PatchCalendarPersonal("u1", "cal1", map[string]interface{}{"color": "red"}) },
			wantMethod: "PATCH",
			wantPath:   "/users/u1/calendar-personals/cal1",
		},
		{
			name:       "CalendarService.RemoveUserFromCalendar",
			call:       func(c *Client) { NewCalendarService(c).RemoveUserFromCalendar("u1", "cal1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/calendars/cal1",
		},
		{
			name:       "CalendarService.UpdateEvent",
			call:       func(c *Client) { NewCalendarService(c).UpdateEvent("u1", "cal1", "ev1", map[string]interface{}{"summary": "y"}) },
			wantMethod: "PUT",
			wantPath:   "/users/u1/calendars/cal1/events/ev1",
		},
		{
			name:       "CalendarService.DeleteEvent",
			call:       func(c *Client) { NewCalendarService(c).DeleteEvent("u1", "cal1", "ev1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/calendars/cal1/events/ev1",
		},
		{
			name:       "CalendarService.CreateDefaultEvent",
			call:       func(c *Client) { NewCalendarService(c).CreateDefaultEvent("u1", map[string]interface{}{"summary": "x"}) },
			wantMethod: "POST",
			wantPath:   "/users/u1/calendar/events",
		},
		{
			name:       "CalendarService.ListDefaultEvents",
			call:       func(c *Client) { NewCalendarService(c).ListDefaultEvents("u1", "2024-01-01", "2024-01-31") },
			wantMethod: "GET",
			wantPath:   "/users/u1/calendar/events",
		},
		{
			name:       "CalendarService.GetDefaultEvent",
			call:       func(c *Client) { NewCalendarService(c).GetDefaultEvent("u1", "ev1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/calendar/events/ev1",
		},
		{
			name:       "CalendarService.UpdateDefaultEvent",
			call:       func(c *Client) { NewCalendarService(c).UpdateDefaultEvent("u1", "ev1", map[string]interface{}{"summary": "y"}) },
			wantMethod: "PUT",
			wantPath:   "/users/u1/calendar/events/ev1",
		},
		{
			name:       "CalendarService.DeleteDefaultEvent",
			call:       func(c *Client) { NewCalendarService(c).DeleteDefaultEvent("u1", "ev1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/calendar/events/ev1",
		},

		// ─── Security ───
		{
			name:       "SecurityService.GetExternalBrowser",
			call:       func(c *Client) { NewSecurityService(c).GetExternalBrowser() },
			wantMethod: "GET",
			wantPath:   "/security/external-browser",
		},
		{
			name:       "SecurityService.EnableExternalBrowser",
			call:       func(c *Client) { NewSecurityService(c).EnableExternalBrowser() },
			wantMethod: "POST",
			wantPath:   "/security/external-browser/enable",
		},
		{
			name:       "SecurityService.DisableExternalBrowser",
			call:       func(c *Client) { NewSecurityService(c).DisableExternalBrowser() },
			wantMethod: "POST",
			wantPath:   "/security/external-browser/disable",
		},

		// ─── Directory ───
		{
			name:       "DirectoryService.ListUsers",
			call:       func(c *Client) { NewDirectoryService(c).ListUsers("", 10) },
			wantMethod: "GET",
			wantPath:   "/users",
		},
		{
			name:       "DirectoryService.GetUser",
			call:       func(c *Client) { NewDirectoryService(c).GetUser("u1") },
			wantMethod: "GET",
			wantPath:   "/users/u1",
		},
		{
			name:       "DirectoryService.CreateUser",
			call:       func(c *Client) { NewDirectoryService(c).CreateUser([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users",
		},
		{
			name:       "DirectoryService.UpdateUser",
			call:       func(c *Client) { NewDirectoryService(c).UpdateUser("u1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/users/u1",
		},
		{
			name:       "DirectoryService.PatchUser",
			call:       func(c *Client) { NewDirectoryService(c).PatchUser("u1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/users/u1",
		},
		{
			name:       "DirectoryService.DeleteUser",
			call:       func(c *Client) { NewDirectoryService(c).DeleteUser("u1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1",
		},
		{
			name:       "DirectoryService.ForceDeleteUser",
			call:       func(c *Client) { NewDirectoryService(c).ForceDeleteUser("u1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/forcedelete",
		},
		{
			name:       "DirectoryService.UndeleteUser",
			call:       func(c *Client) { NewDirectoryService(c).UndeleteUser("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/undelete",
		},
		{
			name:       "DirectoryService.SuspendUser",
			call:       func(c *Client) { NewDirectoryService(c).SuspendUser("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/suspend",
		},
		{
			name:       "DirectoryService.UnsuspendUser",
			call:       func(c *Client) { NewDirectoryService(c).UnsuspendUser("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/unsuspend",
		},
		{
			name:       "DirectoryService.ForceLogoutUser",
			call:       func(c *Client) { NewDirectoryService(c).ForceLogoutUser("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/force-logout",
		},
		{
			name:       "DirectoryService.MoveUser",
			call:       func(c *Client) { NewDirectoryService(c).MoveUser("u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/move",
		},
		{
			name:       "DirectoryService.SetLeaveOfAbsence",
			call:       func(c *Client) { NewDirectoryService(c).SetLeaveOfAbsence("u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/set-leave-of-absence",
		},
		{
			name:       "DirectoryService.ClearLeaveOfAbsence",
			call:       func(c *Client) { NewDirectoryService(c).ClearLeaveOfAbsence("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/clear-leave-of-absence",
		},
		{
			name:       "DirectoryService.CreateUserPhoto",
			call:       func(c *Client) { NewDirectoryService(c).CreateUserPhoto("u1", map[string]interface{}{"fileName": "photo.jpg"}) },
			wantMethod: "POST",
			wantPath:   "/users/u1/photo",
		},
		{
			name:       "DirectoryService.DeleteUserPhoto",
			call:       func(c *Client) { NewDirectoryService(c).DeleteUserPhoto("u1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/photo",
		},
		{
			name:       "DirectoryService.CreateProfileStatus",
			call:       func(c *Client) { NewDirectoryService(c).CreateProfileStatus("u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/user-profile-statuses",
		},
		{
			name:       "DirectoryService.ListProfileStatuses",
			call:       func(c *Client) { NewDirectoryService(c).ListProfileStatuses("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/user-profile-statuses",
		},
		{
			name:       "DirectoryService.GetProfileStatus",
			call:       func(c *Client) { NewDirectoryService(c).GetProfileStatus("u1", "ps1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/user-profile-statuses/ps1",
		},
		{
			name:       "DirectoryService.UpdateProfileStatus",
			call:       func(c *Client) { NewDirectoryService(c).UpdateProfileStatus("u1", "ps1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/users/u1/user-profile-statuses/ps1",
		},
		{
			name:       "DirectoryService.PatchProfileStatus",
			call:       func(c *Client) { NewDirectoryService(c).PatchProfileStatus("u1", "ps1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/users/u1/user-profile-statuses/ps1",
		},
		{
			name:       "DirectoryService.DeleteProfileStatus",
			call:       func(c *Client) { NewDirectoryService(c).DeleteProfileStatus("u1", "ps1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/user-profile-statuses/ps1",
		},
		{
			name:       "DirectoryService.AddAliasEmail",
			call:       func(c *Client) { NewDirectoryService(c).AddAliasEmail("u1", "alias@test.com") },
			wantMethod: "POST",
			wantPath:   "/users/u1/alias-emails/alias@test.com",
		},
		{
			name:       "DirectoryService.DeleteAliasEmail",
			call:       func(c *Client) { NewDirectoryService(c).DeleteAliasEmail("u1", "alias@test.com") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/alias-emails/alias@test.com",
		},
		{
			name:       "DirectoryService.SendInvitationEmail",
			call:       func(c *Client) { NewDirectoryService(c).SendInvitationEmail("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/send-invitation-email",
		},
		{
			name:       "DirectoryService.SendInvitationEmailToAll",
			call:       func(c *Client) { NewDirectoryService(c).SendInvitationEmailToAll() },
			wantMethod: "POST",
			wantPath:   "/users/send-invitation-email",
		},
		{
			name:       "DirectoryService.LinkAllUsersToWorks",
			call:       func(c *Client) { NewDirectoryService(c).LinkAllUsersToWorks() },
			wantMethod: "POST",
			wantPath:   "/users/link-to-works",
		},
		{
			name:       "DirectoryService.LinkUserToWorks",
			call:       func(c *Client) { NewDirectoryService(c).LinkUserToWorks("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/link-to-works",
		},
		{
			name:       "DirectoryService.UnlinkUserToWorks",
			call:       func(c *Client) { NewDirectoryService(c).UnlinkUserToWorks("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/unlink-to-works",
		},
		{
			name:       "DirectoryService.LinkAllUsersToLine",
			call:       func(c *Client) { NewDirectoryService(c).LinkAllUsersToLine() },
			wantMethod: "POST",
			wantPath:   "/users/link-to-line",
		},
		{
			name:       "DirectoryService.LinkUserToLine",
			call:       func(c *Client) { NewDirectoryService(c).LinkUserToLine("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/link-to-line",
		},
		{
			name:       "DirectoryService.UnlinkUserToLine",
			call:       func(c *Client) { NewDirectoryService(c).UnlinkUserToLine("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/unlink-to-line",
		},
		{
			name:       "DirectoryService.GetLinkUrl",
			call:       func(c *Client) { NewDirectoryService(c).GetLinkUrl("u1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/link-url",
		},
		{
			name:       "DirectoryService.ResetLinkUrl",
			call:       func(c *Client) { NewDirectoryService(c).ResetLinkUrl("u1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/link-url/reset",
		},
		{
			name:       "DirectoryService.UpsertUserExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).UpsertUserExternalKeys([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/external-keys",
		},
		{
			name:       "DirectoryService.ListUserExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).ListUserExternalKeys("", 10) },
			wantMethod: "GET",
			wantPath:   "/users/external-keys",
		},
		{
			name:       "DirectoryService.CreateUserCustomProperty",
			call:       func(c *Client) { NewDirectoryService(c).CreateUserCustomProperty([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/users/custom-properties",
		},
		{
			name:       "DirectoryService.ListUserCustomProperties",
			call:       func(c *Client) { NewDirectoryService(c).ListUserCustomProperties("", 10) },
			wantMethod: "GET",
			wantPath:   "/directory/users/custom-properties",
		},
		{
			name:       "DirectoryService.GetUserCustomProperty",
			call:       func(c *Client) { NewDirectoryService(c).GetUserCustomProperty("cp1") },
			wantMethod: "GET",
			wantPath:   "/directory/users/custom-properties/cp1",
		},
		{
			name:       "DirectoryService.PatchUserCustomProperty",
			call:       func(c *Client) { NewDirectoryService(c).PatchUserCustomProperty("cp1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/directory/users/custom-properties/cp1",
		},
		{
			name:       "DirectoryService.DeleteUserCustomProperty",
			call:       func(c *Client) { NewDirectoryService(c).DeleteUserCustomProperty("cp1") },
			wantMethod: "DELETE",
			wantPath:   "/directory/users/custom-properties/cp1",
		},
		{
			name:       "DirectoryService.ListGroups",
			call:       func(c *Client) { NewDirectoryService(c).ListGroups("", 10) },
			wantMethod: "GET",
			wantPath:   "/groups",
		},
		{
			name:       "DirectoryService.GetGroup",
			call:       func(c *Client) { NewDirectoryService(c).GetGroup("g1") },
			wantMethod: "GET",
			wantPath:   "/groups/g1",
		},
		{
			name:       "DirectoryService.CreateGroup",
			call:       func(c *Client) { NewDirectoryService(c).CreateGroup([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/groups",
		},
		{
			name:       "DirectoryService.UpdateGroup",
			call:       func(c *Client) { NewDirectoryService(c).UpdateGroup("g1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/groups/g1",
		},
		{
			name:       "DirectoryService.PatchGroup",
			call:       func(c *Client) { NewDirectoryService(c).PatchGroup("g1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/groups/g1",
		},
		{
			name:       "DirectoryService.DeleteGroup",
			call:       func(c *Client) { NewDirectoryService(c).DeleteGroup("g1") },
			wantMethod: "DELETE",
			wantPath:   "/groups/g1",
		},
		{
			name:       "DirectoryService.ListGroupMembers",
			call:       func(c *Client) { NewDirectoryService(c).ListGroupMembers("g1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/groups/g1/members",
		},
		{
			name:       "DirectoryService.AddGroupMembers",
			call:       func(c *Client) { NewDirectoryService(c).AddGroupMembers("g1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/groups/g1/members",
		},
		{
			name:       "DirectoryService.RemoveGroupMember",
			call:       func(c *Client) { NewDirectoryService(c).RemoveGroupMember("g1", "m1") },
			wantMethod: "DELETE",
			wantPath:   "/groups/g1/members/m1",
		},
		{
			name:       "DirectoryService.ListGroupAdministrators",
			call:       func(c *Client) { NewDirectoryService(c).ListGroupAdministrators("g1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/groups/g1/administrators",
		},
		{
			name:       "DirectoryService.AddGroupAdministrator",
			call:       func(c *Client) { NewDirectoryService(c).AddGroupAdministrator("g1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/groups/g1/administrators",
		},
		{
			name:       "DirectoryService.RemoveGroupAdministrator",
			call:       func(c *Client) { NewDirectoryService(c).RemoveGroupAdministrator("g1", "u1") },
			wantMethod: "DELETE",
			wantPath:   "/groups/g1/administrators/u1",
		},
		{
			name:       "DirectoryService.UpsertGroupExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).UpsertGroupExternalKeys([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/groups/external-keys",
		},
		{
			name:       "DirectoryService.ListGroupExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).ListGroupExternalKeys("", 10) },
			wantMethod: "GET",
			wantPath:   "/groups/external-keys",
		},
		{
			name:       "DirectoryService.ListOrgUnits",
			call:       func(c *Client) { NewDirectoryService(c).ListOrgUnits("", 10) },
			wantMethod: "GET",
			wantPath:   "/orgunits",
		},
		{
			name:       "DirectoryService.GetOrgUnit",
			call:       func(c *Client) { NewDirectoryService(c).GetOrgUnit("ou1") },
			wantMethod: "GET",
			wantPath:   "/orgunits/ou1",
		},
		{
			name:       "DirectoryService.CreateOrgUnit",
			call:       func(c *Client) { NewDirectoryService(c).CreateOrgUnit([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/orgunits",
		},
		{
			name:       "DirectoryService.UpdateOrgUnit",
			call:       func(c *Client) { NewDirectoryService(c).UpdateOrgUnit("ou1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/orgunits/ou1",
		},
		{
			name:       "DirectoryService.PatchOrgUnit",
			call:       func(c *Client) { NewDirectoryService(c).PatchOrgUnit("ou1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/orgunits/ou1",
		},
		{
			name:       "DirectoryService.DeleteOrgUnit",
			call:       func(c *Client) { NewDirectoryService(c).DeleteOrgUnit("ou1") },
			wantMethod: "DELETE",
			wantPath:   "/orgunits/ou1",
		},
		{
			name:       "DirectoryService.MoveOrgUnit",
			call:       func(c *Client) { NewDirectoryService(c).MoveOrgUnit("ou1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/orgunits/ou1/move",
		},
		{
			name:       "DirectoryService.ListOrgUnitMembers",
			call:       func(c *Client) { NewDirectoryService(c).ListOrgUnitMembers("ou1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/orgunits/ou1/members",
		},
		{
			name:       "DirectoryService.CreateOrgUnitAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).CreateOrgUnitAccessRestrict("ou1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/orgunits/ou1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.GetOrgUnitAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).GetOrgUnitAccessRestrict("ou1") },
			wantMethod: "GET",
			wantPath:   "/orgunits/ou1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.UpdateOrgUnitAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).UpdateOrgUnitAccessRestrict("ou1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/orgunits/ou1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.DeleteOrgUnitAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).DeleteOrgUnitAccessRestrict("ou1") },
			wantMethod: "DELETE",
			wantPath:   "/orgunits/ou1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.UpsertOrgUnitExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).UpsertOrgUnitExternalKeys([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/orgunits/external-keys",
		},
		{
			name:       "DirectoryService.ListOrgUnitExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).ListOrgUnitExternalKeys("", 10) },
			wantMethod: "GET",
			wantPath:   "/orgunits/external-keys",
		},
		{
			name:       "DirectoryService.ListLevels",
			call:       func(c *Client) { NewDirectoryService(c).ListLevels("", 10) },
			wantMethod: "GET",
			wantPath:   "/directory/levels",
		},
		{
			name:       "DirectoryService.ListPositions",
			call:       func(c *Client) { NewDirectoryService(c).ListPositions("", 10) },
			wantMethod: "GET",
			wantPath:   "/directory/positions",
		},
		{
			name:       "DirectoryService.ListUserTypes",
			call:       func(c *Client) { NewDirectoryService(c).ListUserTypes("", 10) },
			wantMethod: "GET",
			wantPath:   "/directory/user-types",
		},
		{
			name:       "DirectoryService.ListEmploymentTypes",
			call:       func(c *Client) { NewDirectoryService(c).ListEmploymentTypes("", 10) },
			wantMethod: "GET",
			wantPath:   "/directory/employment-types",
		},
		{
			name:       "DirectoryService.CreatePosition",
			call:       func(c *Client) { NewDirectoryService(c).CreatePosition([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/positions",
		},
		{
			name:       "DirectoryService.GetPosition",
			call:       func(c *Client) { NewDirectoryService(c).GetPosition("pos1") },
			wantMethod: "GET",
			wantPath:   "/directory/positions/pos1",
		},
		{
			name:       "DirectoryService.UpdatePosition",
			call:       func(c *Client) { NewDirectoryService(c).UpdatePosition("pos1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/directory/positions/pos1",
		},
		{
			name:       "DirectoryService.PatchPosition",
			call:       func(c *Client) { NewDirectoryService(c).PatchPosition("pos1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/directory/positions/pos1",
		},
		{
			name:       "DirectoryService.DeletePosition",
			call:       func(c *Client) { NewDirectoryService(c).DeletePosition("pos1") },
			wantMethod: "DELETE",
			wantPath:   "/directory/positions/pos1",
		},
		{
			name:       "DirectoryService.EnablePositions",
			call:       func(c *Client) { NewDirectoryService(c).EnablePositions() },
			wantMethod: "POST",
			wantPath:   "/directory/positions/enable",
		},
		{
			name:       "DirectoryService.DisablePositions",
			call:       func(c *Client) { NewDirectoryService(c).DisablePositions() },
			wantMethod: "POST",
			wantPath:   "/directory/positions/disable",
		},
		{
			name:       "DirectoryService.UpsertPositionExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).UpsertPositionExternalKeys([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/positions/external-keys",
		},
		{
			name:       "DirectoryService.ListPositionExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).ListPositionExternalKeys() },
			wantMethod: "GET",
			wantPath:   "/directory/positions/external-keys",
		},
		{
			name:       "DirectoryService.CreateLevel",
			call:       func(c *Client) { NewDirectoryService(c).CreateLevel([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/levels",
		},
		{
			name:       "DirectoryService.GetLevel",
			call:       func(c *Client) { NewDirectoryService(c).GetLevel("lv1") },
			wantMethod: "GET",
			wantPath:   "/directory/levels/lv1",
		},
		{
			name:       "DirectoryService.UpdateLevel",
			call:       func(c *Client) { NewDirectoryService(c).UpdateLevel("lv1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/directory/levels/lv1",
		},
		{
			name:       "DirectoryService.PatchLevel",
			call:       func(c *Client) { NewDirectoryService(c).PatchLevel("lv1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/directory/levels/lv1",
		},
		{
			name:       "DirectoryService.DeleteLevel",
			call:       func(c *Client) { NewDirectoryService(c).DeleteLevel("lv1") },
			wantMethod: "DELETE",
			wantPath:   "/directory/levels/lv1",
		},
		{
			name:       "DirectoryService.EnableLevels",
			call:       func(c *Client) { NewDirectoryService(c).EnableLevels() },
			wantMethod: "POST",
			wantPath:   "/directory/levels/enable",
		},
		{
			name:       "DirectoryService.DisableLevels",
			call:       func(c *Client) { NewDirectoryService(c).DisableLevels() },
			wantMethod: "POST",
			wantPath:   "/directory/levels/disable",
		},
		{
			name:       "DirectoryService.UpsertLevelExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).UpsertLevelExternalKeys([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/levels/external-keys",
		},
		{
			name:       "DirectoryService.ListLevelExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).ListLevelExternalKeys() },
			wantMethod: "GET",
			wantPath:   "/directory/levels/external-keys",
		},
		{
			name:       "DirectoryService.CreateEmploymentType",
			call:       func(c *Client) { NewDirectoryService(c).CreateEmploymentType([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/employment-types",
		},
		{
			name:       "DirectoryService.GetEmploymentType",
			call:       func(c *Client) { NewDirectoryService(c).GetEmploymentType("et1") },
			wantMethod: "GET",
			wantPath:   "/directory/employment-types/et1",
		},
		{
			name:       "DirectoryService.UpdateEmploymentType",
			call:       func(c *Client) { NewDirectoryService(c).UpdateEmploymentType("et1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/directory/employment-types/et1",
		},
		{
			name:       "DirectoryService.PatchEmploymentType",
			call:       func(c *Client) { NewDirectoryService(c).PatchEmploymentType("et1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/directory/employment-types/et1",
		},
		{
			name:       "DirectoryService.DeleteEmploymentType",
			call:       func(c *Client) { NewDirectoryService(c).DeleteEmploymentType("et1") },
			wantMethod: "DELETE",
			wantPath:   "/directory/employment-types/et1",
		},
		{
			name:       "DirectoryService.EnableEmploymentTypes",
			call:       func(c *Client) { NewDirectoryService(c).EnableEmploymentTypes() },
			wantMethod: "POST",
			wantPath:   "/directory/employment-types/enable",
		},
		{
			name:       "DirectoryService.DisableEmploymentTypes",
			call:       func(c *Client) { NewDirectoryService(c).DisableEmploymentTypes() },
			wantMethod: "POST",
			wantPath:   "/directory/employment-types/disable",
		},
		{
			name:       "DirectoryService.UpsertEmploymentTypeExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).UpsertEmploymentTypeExternalKeys([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/employment-types/external-keys",
		},
		{
			name:       "DirectoryService.ListEmploymentTypeExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).ListEmploymentTypeExternalKeys() },
			wantMethod: "GET",
			wantPath:   "/directory/employment-types/external-keys",
		},
		{
			name:       "DirectoryService.CreateEmploymentTypeAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).CreateEmploymentTypeAccessRestrict("et1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/employment-types/et1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.GetEmploymentTypeAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).GetEmploymentTypeAccessRestrict("et1") },
			wantMethod: "GET",
			wantPath:   "/directory/employment-types/et1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.UpdateEmploymentTypeAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).UpdateEmploymentTypeAccessRestrict("et1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/directory/employment-types/et1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.DeleteEmploymentTypeAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).DeleteEmploymentTypeAccessRestrict("et1") },
			wantMethod: "DELETE",
			wantPath:   "/directory/employment-types/et1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.CreateUserType",
			call:       func(c *Client) { NewDirectoryService(c).CreateUserType([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/user-types",
		},
		{
			name:       "DirectoryService.GetUserType",
			call:       func(c *Client) { NewDirectoryService(c).GetUserType("ut1") },
			wantMethod: "GET",
			wantPath:   "/directory/user-types/ut1",
		},
		{
			name:       "DirectoryService.UpdateUserType",
			call:       func(c *Client) { NewDirectoryService(c).UpdateUserType("ut1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/directory/user-types/ut1",
		},
		{
			name:       "DirectoryService.PatchUserType",
			call:       func(c *Client) { NewDirectoryService(c).PatchUserType("ut1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/directory/user-types/ut1",
		},
		{
			name:       "DirectoryService.DeleteUserType",
			call:       func(c *Client) { NewDirectoryService(c).DeleteUserType("ut1") },
			wantMethod: "DELETE",
			wantPath:   "/directory/user-types/ut1",
		},
		{
			name:       "DirectoryService.EnableUserTypes",
			call:       func(c *Client) { NewDirectoryService(c).EnableUserTypes() },
			wantMethod: "POST",
			wantPath:   "/directory/user-types/enable",
		},
		{
			name:       "DirectoryService.DisableUserTypes",
			call:       func(c *Client) { NewDirectoryService(c).DisableUserTypes() },
			wantMethod: "POST",
			wantPath:   "/directory/user-types/disable",
		},
		{
			name:       "DirectoryService.UpsertUserTypeExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).UpsertUserTypeExternalKeys([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/user-types/external-keys",
		},
		{
			name:       "DirectoryService.ListUserTypeExternalKeys",
			call:       func(c *Client) { NewDirectoryService(c).ListUserTypeExternalKeys() },
			wantMethod: "GET",
			wantPath:   "/directory/user-types/external-keys",
		},
		{
			name:       "DirectoryService.CreateUserTypeAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).CreateUserTypeAccessRestrict("ut1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/user-types/ut1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.GetUserTypeAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).GetUserTypeAccessRestrict("ut1") },
			wantMethod: "GET",
			wantPath:   "/directory/user-types/ut1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.UpdateUserTypeAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).UpdateUserTypeAccessRestrict("ut1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/directory/user-types/ut1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.DeleteUserTypeAccessRestrict",
			call:       func(c *Client) { NewDirectoryService(c).DeleteUserTypeAccessRestrict("ut1") },
			wantMethod: "DELETE",
			wantPath:   "/directory/user-types/ut1/orgunit-access-restrict",
		},
		{
			name:       "DirectoryService.CreateDirectoryProfileStatus",
			call:       func(c *Client) { NewDirectoryService(c).CreateDirectoryProfileStatus([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/profile-statuses",
		},
		{
			name:       "DirectoryService.ListDirectoryProfileStatuses",
			call:       func(c *Client) { NewDirectoryService(c).ListDirectoryProfileStatuses("", 10) },
			wantMethod: "GET",
			wantPath:   "/directory/profile-statuses",
		},
		{
			name:       "DirectoryService.GetDirectoryProfileStatus",
			call:       func(c *Client) { NewDirectoryService(c).GetDirectoryProfileStatus("dps1") },
			wantMethod: "GET",
			wantPath:   "/directory/profile-statuses/dps1",
		},
		{
			name:       "DirectoryService.UpdateDirectoryProfileStatus",
			call:       func(c *Client) { NewDirectoryService(c).UpdateDirectoryProfileStatus("dps1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/directory/profile-statuses/dps1",
		},
		{
			name:       "DirectoryService.PatchDirectoryProfileStatus",
			call:       func(c *Client) { NewDirectoryService(c).PatchDirectoryProfileStatus("dps1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/directory/profile-statuses/dps1",
		},
		{
			name:       "DirectoryService.DeleteDirectoryProfileStatus",
			call:       func(c *Client) { NewDirectoryService(c).DeleteDirectoryProfileStatus("dps1") },
			wantMethod: "DELETE",
			wantPath:   "/directory/profile-statuses/dps1",
		},
		{
			name:       "DirectoryService.EnableDirectoryProfileStatuses",
			call:       func(c *Client) { NewDirectoryService(c).EnableDirectoryProfileStatuses() },
			wantMethod: "POST",
			wantPath:   "/directory/profile-statuses/enable",
		},
		{
			name:       "DirectoryService.DisableDirectoryProfileStatuses",
			call:       func(c *Client) { NewDirectoryService(c).DisableDirectoryProfileStatuses() },
			wantMethod: "POST",
			wantPath:   "/directory/profile-statuses/disable",
		},
		{
			name:       "DirectoryService.CreateCustomField",
			call:       func(c *Client) { NewDirectoryService(c).CreateCustomField([]byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/directory/custom-fields",
		},
		{
			name:       "DirectoryService.ListCustomFields",
			call:       func(c *Client) { NewDirectoryService(c).ListCustomFields("", 10) },
			wantMethod: "GET",
			wantPath:   "/directory/custom-fields",
		},
		{
			name:       "DirectoryService.GetCustomField",
			call:       func(c *Client) { NewDirectoryService(c).GetCustomField("cf1") },
			wantMethod: "GET",
			wantPath:   "/directory/custom-fields/cf1",
		},
		{
			name:       "DirectoryService.PatchCustomField",
			call:       func(c *Client) { NewDirectoryService(c).PatchCustomField("cf1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/directory/custom-fields/cf1",
		},
		{
			name:       "DirectoryService.DeleteCustomField",
			call:       func(c *Client) { NewDirectoryService(c).DeleteCustomField("cf1") },
			wantMethod: "DELETE",
			wantPath:   "/directory/custom-fields/cf1",
		},

		// ─── Drive (personal) ───
		{
			name:       "DriveService.GetDriveInfo",
			call:       func(c *Client) { NewDriveService(c).GetDriveInfo("u1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive",
		},
		{
			name:       "DriveService.ListFiles",
			call:       func(c *Client) { NewDriveService(c).ListFiles("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/files",
		},
		{
			name:       "DriveService.ListFolderChildren",
			call:       func(c *Client) { NewDriveService(c).ListFolderChildren("u1", "f1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/files/f1/children",
		},
		{
			name:       "DriveService.GetFile",
			call:       func(c *Client) { NewDriveService(c).GetFile("u1", "f1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/files/f1",
		},
		{
			name:       "DriveService.CreateUploadURL",
			call:       func(c *Client) { NewDriveService(c).CreateUploadURL("u1", map[string]interface{}{"fileName": "a.txt"}, 100) },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files",
		},
		{
			name:       "DriveService.CreateUploadURLInFolder",
			call:       func(c *Client) { NewDriveService(c).CreateUploadURLInFolder("u1", "f1", map[string]interface{}{"fileName": "a.txt"}, 100) },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1",
		},
		{
			name:       "DriveService.CreateFolder",
			call:       func(c *Client) { NewDriveService(c).CreateFolder("u1", "myfolder") },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/createfolder",
		},
		{
			name:       "DriveService.CreateFolderInParent",
			call:       func(c *Client) { NewDriveService(c).CreateFolderInParent("u1", "p1", "myfolder") },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/p1/createfolder",
		},
		{
			name:       "DriveService.DeleteFile",
			call:       func(c *Client) { NewDriveService(c).DeleteFile("u1", "f1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/drive/files/f1",
		},
		{
			name:       "DriveService.ListTrashFiles",
			call:       func(c *Client) { NewDriveService(c).ListTrashFiles("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/trash-files",
		},
		{
			name:       "DriveService.RestoreTrashFile",
			call:       func(c *Client) { NewDriveService(c).RestoreTrashFile("u1", "f1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/trash-files/f1/restore",
		},
		{
			name:       "DriveService.CopyFile",
			call:       func(c *Client) { NewDriveService(c).CopyFile("u1", "f1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1/copy",
		},
		{
			name:       "DriveService.RenameFile",
			call:       func(c *Client) { NewDriveService(c).RenameFile("u1", "f1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1/rename",
		},
		{
			name:       "DriveService.MoveFile",
			call:       func(c *Client) { NewDriveService(c).MoveFile("u1", "f1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1/move",
		},
		{
			name:       "DriveService.ProtectFile",
			call:       func(c *Client) { NewDriveService(c).ProtectFile("u1", "f1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1/protect",
		},
		{
			name:       "DriveService.UnprotectFile",
			call:       func(c *Client) { NewDriveService(c).UnprotectFile("u1", "f1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1/unprotect",
		},
		{
			name:       "DriveService.LockFile",
			call:       func(c *Client) { NewDriveService(c).LockFile("u1", "f1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1/lock",
		},
		{
			name:       "DriveService.UnlockFile",
			call:       func(c *Client) { NewDriveService(c).UnlockFile("u1", "f1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1/unlock",
		},
		{
			name:       "DriveService.ListRevisions",
			call:       func(c *Client) { NewDriveService(c).ListRevisions("u1", "f1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/files/f1/revisions",
		},
		{
			name:       "DriveService.GetRevision",
			call:       func(c *Client) { NewDriveService(c).GetRevision("u1", "f1", "r1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/files/f1/revisions/r1",
		},
		{
			name:       "DriveService.RestoreRevision",
			call:       func(c *Client) { NewDriveService(c).RestoreRevision("u1", "f1", "r1") },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1/revisions/r1/restore",
		},
		{
			name:       "DriveService.DeleteTrashFile",
			call:       func(c *Client) { NewDriveService(c).DeleteTrashFile("u1", "f1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/drive/trash-files/f1",
		},
		{
			name:       "DriveService.GetLinkSetting",
			call:       func(c *Client) { NewDriveService(c).GetLinkSetting("u1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/link-setting",
		},
		{
			name:       "DriveService.GetLink",
			call:       func(c *Client) { NewDriveService(c).GetLink("u1", "f1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/files/f1/link",
		},
		{
			name:       "DriveService.CreateLink",
			call:       func(c *Client) { NewDriveService(c).CreateLink("u1", "f1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1/link",
		},
		{
			name:       "DriveService.PatchLink",
			call:       func(c *Client) { NewDriveService(c).PatchLink("u1", "f1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/users/u1/drive/files/f1/link",
		},
		{
			name:       "DriveService.DeleteLink",
			call:       func(c *Client) { NewDriveService(c).DeleteLink("u1", "f1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/drive/files/f1/link",
		},
		{
			name:       "DriveService.GetShare",
			call:       func(c *Client) { NewDriveService(c).GetShare("u1", "f1") },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/files/f1/share",
		},
		{
			name:       "DriveService.CreateShare",
			call:       func(c *Client) { NewDriveService(c).CreateShare("u1", "f1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/files/f1/share",
		},
		{
			name:       "DriveService.PatchShare",
			call:       func(c *Client) { NewDriveService(c).PatchShare("u1", "f1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/users/u1/drive/files/f1/share",
		},
		{
			name:       "DriveService.DeleteShare",
			call:       func(c *Client) { NewDriveService(c).DeleteShare("u1", "f1") },
			wantMethod: "DELETE",
			wantPath:   "/users/u1/drive/files/f1/share",
		},
		{
			name:       "DriveService.ListShareSubFolders",
			call:       func(c *Client) { NewDriveService(c).ListShareSubFolders("u1", "f1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/files/f1/share-sub-folders",
		},
		{
			name:       "DriveService.ListSharedFolders",
			call:       func(c *Client) { NewDriveService(c).ListSharedFolders("u1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/sharedfolders",
		},
		{
			name:       "DriveService.ListSharedFolderFiles",
			call:       func(c *Client) { NewDriveService(c).ListSharedFolderFiles("u1", "sf1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/sharedfolders/sf1/files",
		},

		// ─── GroupFolder (drive_group.go) ───
		{name: "GroupFolderService.GetFolder", call: func(c *Client) { NewGroupFolderService(c).GetFolder("g1") }, wantMethod: "GET", wantPath: "/groups/g1/folder"},
		{name: "GroupFolderService.ListFiles", call: func(c *Client) { NewGroupFolderService(c).ListFiles("g1", "", 10) }, wantMethod: "GET", wantPath: "/groups/g1/folder/files"},
		{name: "GroupFolderService.ListFolderChildren", call: func(c *Client) { NewGroupFolderService(c).ListFolderChildren("g1", "f1", "", 10) }, wantMethod: "GET", wantPath: "/groups/g1/folder/files/f1/children"},
		{name: "GroupFolderService.GetFile", call: func(c *Client) { NewGroupFolderService(c).GetFile("g1", "f1") }, wantMethod: "GET", wantPath: "/groups/g1/folder/files/f1"},
		{name: "GroupFolderService.CreateFolder", call: func(c *Client) { NewGroupFolderService(c).CreateFolder("g1") }, wantMethod: "POST", wantPath: "/groups/g1/folder"},
		{name: "GroupFolderService.DeleteFolder", call: func(c *Client) { NewGroupFolderService(c).DeleteFolder("g1") }, wantMethod: "DELETE", wantPath: "/groups/g1/folder"},
		{name: "GroupFolderService.CreateFolderInRoot", call: func(c *Client) { NewGroupFolderService(c).CreateFolderInRoot("g1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/createfolder"},
		{name: "GroupFolderService.CreateSubFolder", call: func(c *Client) { NewGroupFolderService(c).CreateSubFolder("g1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/createfolder"},
		{name: "GroupFolderService.DeleteFile", call: func(c *Client) { NewGroupFolderService(c).DeleteFile("g1", "f1") }, wantMethod: "DELETE", wantPath: "/groups/g1/folder/files/f1"},
		{name: "GroupFolderService.CreateUploadURL", call: func(c *Client) { NewGroupFolderService(c).CreateUploadURL("g1", "f1", map[string]interface{}{"fileName": "a.txt"}, 100) }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1"},
		{name: "GroupFolderService.CreateRootUploadURL", call: func(c *Client) { NewGroupFolderService(c).CreateRootUploadURL("g1", map[string]interface{}{"fileName": "a.txt"}, 100) }, wantMethod: "POST", wantPath: "/groups/g1/folder/files"},
		{name: "GroupFolderService.CopyFile", call: func(c *Client) { NewGroupFolderService(c).CopyFile("g1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/copy"},
		{name: "GroupFolderService.RenameFile", call: func(c *Client) { NewGroupFolderService(c).RenameFile("g1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/rename"},
		{name: "GroupFolderService.MoveFile", call: func(c *Client) { NewGroupFolderService(c).MoveFile("g1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/move"},
		{name: "GroupFolderService.ProtectFile", call: func(c *Client) { NewGroupFolderService(c).ProtectFile("g1", "f1") }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/protect"},
		{name: "GroupFolderService.UnprotectFile", call: func(c *Client) { NewGroupFolderService(c).UnprotectFile("g1", "f1") }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/unprotect"},
		{name: "GroupFolderService.LockFile", call: func(c *Client) { NewGroupFolderService(c).LockFile("g1", "f1") }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/lock"},
		{name: "GroupFolderService.UnlockFile", call: func(c *Client) { NewGroupFolderService(c).UnlockFile("g1", "f1") }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/unlock"},
		{name: "GroupFolderService.ListRevisions", call: func(c *Client) { NewGroupFolderService(c).ListRevisions("g1", "f1", "", 10) }, wantMethod: "GET", wantPath: "/groups/g1/folder/files/f1/revisions"},
		{name: "GroupFolderService.GetRevision", call: func(c *Client) { NewGroupFolderService(c).GetRevision("g1", "f1", "r1") }, wantMethod: "GET", wantPath: "/groups/g1/folder/files/f1/revisions/r1"},
		{name: "GroupFolderService.RestoreRevision", call: func(c *Client) { NewGroupFolderService(c).RestoreRevision("g1", "f1", "r1") }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/revisions/r1/restore"},
		{name: "GroupFolderService.ListTrashFiles", call: func(c *Client) { NewGroupFolderService(c).ListTrashFiles("g1", "", 10) }, wantMethod: "GET", wantPath: "/groups/g1/folder/trash-files"},
		{name: "GroupFolderService.RestoreTrashFile", call: func(c *Client) { NewGroupFolderService(c).RestoreTrashFile("g1", "f1") }, wantMethod: "POST", wantPath: "/groups/g1/folder/trash-files/f1/restore"},
		{name: "GroupFolderService.DeleteTrashFile", call: func(c *Client) { NewGroupFolderService(c).DeleteTrashFile("g1", "f1") }, wantMethod: "DELETE", wantPath: "/groups/g1/folder/trash-files/f1"},
		{name: "GroupFolderService.GetLinkSetting", call: func(c *Client) { NewGroupFolderService(c).GetLinkSetting("g1") }, wantMethod: "GET", wantPath: "/groups/g1/folder/link-setting"},
		{name: "GroupFolderService.GetLink", call: func(c *Client) { NewGroupFolderService(c).GetLink("g1", "f1") }, wantMethod: "GET", wantPath: "/groups/g1/folder/files/f1/link"},
		{name: "GroupFolderService.CreateLink", call: func(c *Client) { NewGroupFolderService(c).CreateLink("g1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/link"},
		{name: "GroupFolderService.PatchLink", call: func(c *Client) { NewGroupFolderService(c).PatchLink("g1", "f1", []byte(`{}`)) }, wantMethod: "PATCH", wantPath: "/groups/g1/folder/files/f1/link"},
		{name: "GroupFolderService.DeleteLink", call: func(c *Client) { NewGroupFolderService(c).DeleteLink("g1", "f1") }, wantMethod: "DELETE", wantPath: "/groups/g1/folder/files/f1/link"},
		{name: "GroupFolderService.ListPermissions", call: func(c *Client) { NewGroupFolderService(c).ListPermissions("g1", "f1") }, wantMethod: "GET", wantPath: "/groups/g1/folder/files/f1/permissions"},
		{name: "GroupFolderService.CreatePermission", call: func(c *Client) { NewGroupFolderService(c).CreatePermission("g1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/groups/g1/folder/files/f1/permissions"},
		{name: "GroupFolderService.GetPermission", call: func(c *Client) { NewGroupFolderService(c).GetPermission("g1", "f1", "pm1") }, wantMethod: "GET", wantPath: "/groups/g1/folder/files/f1/permissions/pm1"},
		{name: "GroupFolderService.PatchPermission", call: func(c *Client) { NewGroupFolderService(c).PatchPermission("g1", "f1", "pm1", []byte(`{}`)) }, wantMethod: "PATCH", wantPath: "/groups/g1/folder/files/f1/permissions/pm1"},
		{name: "GroupFolderService.DeletePermission", call: func(c *Client) { NewGroupFolderService(c).DeletePermission("g1", "f1", "pm1") }, wantMethod: "DELETE", wantPath: "/groups/g1/folder/files/f1/permissions/pm1"},
		{name: "GroupFolderService.DeleteAllPermissions", call: func(c *Client) { NewGroupFolderService(c).DeleteAllPermissions("g1", "f1") }, wantMethod: "DELETE", wantPath: "/groups/g1/folder/files/f1/permissions"},

		// ─── SharedDrive (drive_shared.go) ───
		{name: "SharedDriveService.ListDrives", call: func(c *Client) { NewSharedDriveService(c).ListDrives("", 10) }, wantMethod: "GET", wantPath: "/sharedrives"},
		{name: "SharedDriveService.GetDrive", call: func(c *Client) { NewSharedDriveService(c).GetDrive("d1") }, wantMethod: "GET", wantPath: "/sharedrives/d1"},
		{name: "SharedDriveService.ListFiles", call: func(c *Client) { NewSharedDriveService(c).ListFiles("d1", "", 10) }, wantMethod: "GET", wantPath: "/sharedrives/d1/files"},
		{name: "SharedDriveService.ListFolderChildren", call: func(c *Client) { NewSharedDriveService(c).ListFolderChildren("d1", "f1", "", 10) }, wantMethod: "GET", wantPath: "/sharedrives/d1/files/f1/children"},
		{name: "SharedDriveService.GetFile", call: func(c *Client) { NewSharedDriveService(c).GetFile("d1", "f1") }, wantMethod: "GET", wantPath: "/sharedrives/d1/files/f1"},
		{name: "SharedDriveService.CreateUploadURL", call: func(c *Client) { NewSharedDriveService(c).CreateUploadURL("d1", map[string]interface{}{"fileName": "a.txt"}, 100) }, wantMethod: "POST", wantPath: "/sharedrives/d1/files"},
		{name: "SharedDriveService.CreateDrive", call: func(c *Client) { NewSharedDriveService(c).CreateDrive([]byte(`{}`)) }, wantMethod: "POST", wantPath: "/sharedrives"},
		{name: "SharedDriveService.PatchDrive", call: func(c *Client) { NewSharedDriveService(c).PatchDrive("d1", []byte(`{}`)) }, wantMethod: "PATCH", wantPath: "/sharedrives/d1"},
		{name: "SharedDriveService.DeleteDrive", call: func(c *Client) { NewSharedDriveService(c).DeleteDrive("d1") }, wantMethod: "DELETE", wantPath: "/sharedrives/d1"},
		{name: "SharedDriveService.CreateFolderInRoot", call: func(c *Client) { NewSharedDriveService(c).CreateFolderInRoot("d1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/createfolder"},
		{name: "SharedDriveService.CreateSubFolder", call: func(c *Client) { NewSharedDriveService(c).CreateSubFolder("d1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/createfolder"},
		{name: "SharedDriveService.DeleteFile", call: func(c *Client) { NewSharedDriveService(c).DeleteFile("d1", "f1") }, wantMethod: "DELETE", wantPath: "/sharedrives/d1/files/f1"},
		{name: "SharedDriveService.CreateUploadURLInFolder", call: func(c *Client) { NewSharedDriveService(c).CreateUploadURLInFolder("d1", "f1", map[string]interface{}{"fileName": "a.txt"}, 100) }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1"},
		{name: "SharedDriveService.CopyFile", call: func(c *Client) { NewSharedDriveService(c).CopyFile("d1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/copy"},
		{name: "SharedDriveService.RenameFile", call: func(c *Client) { NewSharedDriveService(c).RenameFile("d1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/rename"},
		{name: "SharedDriveService.MoveFile", call: func(c *Client) { NewSharedDriveService(c).MoveFile("d1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/move"},
		{name: "SharedDriveService.ProtectFile", call: func(c *Client) { NewSharedDriveService(c).ProtectFile("d1", "f1") }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/protect"},
		{name: "SharedDriveService.UnprotectFile", call: func(c *Client) { NewSharedDriveService(c).UnprotectFile("d1", "f1") }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/unprotect"},
		{name: "SharedDriveService.LockFile", call: func(c *Client) { NewSharedDriveService(c).LockFile("d1", "f1") }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/lock"},
		{name: "SharedDriveService.UnlockFile", call: func(c *Client) { NewSharedDriveService(c).UnlockFile("d1", "f1") }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/unlock"},
		{name: "SharedDriveService.ListRevisions", call: func(c *Client) { NewSharedDriveService(c).ListRevisions("d1", "f1", "", 10) }, wantMethod: "GET", wantPath: "/sharedrives/d1/files/f1/revisions"},
		{name: "SharedDriveService.GetRevision", call: func(c *Client) { NewSharedDriveService(c).GetRevision("d1", "f1", "r1") }, wantMethod: "GET", wantPath: "/sharedrives/d1/files/f1/revisions/r1"},
		{name: "SharedDriveService.RestoreRevision", call: func(c *Client) { NewSharedDriveService(c).RestoreRevision("d1", "f1", "r1") }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/revisions/r1/restore"},
		{name: "SharedDriveService.ListTrashFiles", call: func(c *Client) { NewSharedDriveService(c).ListTrashFiles("d1", "", 10) }, wantMethod: "GET", wantPath: "/sharedrives/d1/trash-files"},
		{name: "SharedDriveService.RestoreTrashFile", call: func(c *Client) { NewSharedDriveService(c).RestoreTrashFile("d1", "f1") }, wantMethod: "POST", wantPath: "/sharedrives/d1/trash-files/f1/restore"},
		{name: "SharedDriveService.DeleteTrashFile", call: func(c *Client) { NewSharedDriveService(c).DeleteTrashFile("d1", "f1") }, wantMethod: "DELETE", wantPath: "/sharedrives/d1/trash-files/f1"},
		{name: "SharedDriveService.GetLinkSetting", call: func(c *Client) { NewSharedDriveService(c).GetLinkSetting("d1") }, wantMethod: "GET", wantPath: "/sharedrives/d1/link-setting"},
		{name: "SharedDriveService.GetLink", call: func(c *Client) { NewSharedDriveService(c).GetLink("d1", "f1") }, wantMethod: "GET", wantPath: "/sharedrives/d1/files/f1/link"},
		{name: "SharedDriveService.CreateLink", call: func(c *Client) { NewSharedDriveService(c).CreateLink("d1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/link"},
		{name: "SharedDriveService.PatchLink", call: func(c *Client) { NewSharedDriveService(c).PatchLink("d1", "f1", []byte(`{}`)) }, wantMethod: "PATCH", wantPath: "/sharedrives/d1/files/f1/link"},
		{name: "SharedDriveService.DeleteLink", call: func(c *Client) { NewSharedDriveService(c).DeleteLink("d1", "f1") }, wantMethod: "DELETE", wantPath: "/sharedrives/d1/files/f1/link"},
		{name: "SharedDriveService.ListPermissions", call: func(c *Client) { NewSharedDriveService(c).ListPermissions("d1") }, wantMethod: "GET", wantPath: "/sharedrives/d1/permissions"},
		{name: "SharedDriveService.CreatePermission", call: func(c *Client) { NewSharedDriveService(c).CreatePermission("d1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/sharedrives/d1/permissions"},
		{name: "SharedDriveService.GetPermission", call: func(c *Client) { NewSharedDriveService(c).GetPermission("d1", "pm1") }, wantMethod: "GET", wantPath: "/sharedrives/d1/permissions/pm1"},
		{name: "SharedDriveService.PatchPermission", call: func(c *Client) { NewSharedDriveService(c).PatchPermission("d1", "pm1", []byte(`{}`)) }, wantMethod: "PATCH", wantPath: "/sharedrives/d1/permissions/pm1"},
		{name: "SharedDriveService.DeletePermission", call: func(c *Client) { NewSharedDriveService(c).DeletePermission("d1", "pm1") }, wantMethod: "DELETE", wantPath: "/sharedrives/d1/permissions/pm1"},
		{name: "SharedDriveService.DeleteAllPermissions", call: func(c *Client) { NewSharedDriveService(c).DeleteAllPermissions("d1") }, wantMethod: "DELETE", wantPath: "/sharedrives/d1/permissions"},
		{name: "SharedDriveService.EnablePermissions", call: func(c *Client) { NewSharedDriveService(c).EnablePermissions("d1") }, wantMethod: "POST", wantPath: "/sharedrives/d1/permissions/enable"},
		{name: "SharedDriveService.DisablePermissions", call: func(c *Client) { NewSharedDriveService(c).DisablePermissions("d1") }, wantMethod: "POST", wantPath: "/sharedrives/d1/permissions/disable"},
		{name: "SharedDriveService.ListFilePermissions", call: func(c *Client) { NewSharedDriveService(c).ListFilePermissions("d1", "f1") }, wantMethod: "GET", wantPath: "/sharedrives/d1/files/f1/permissions"},
		{name: "SharedDriveService.CreateFilePermission", call: func(c *Client) { NewSharedDriveService(c).CreateFilePermission("d1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/permissions"},
		{name: "SharedDriveService.GetFilePermission", call: func(c *Client) { NewSharedDriveService(c).GetFilePermission("d1", "f1", "pm1") }, wantMethod: "GET", wantPath: "/sharedrives/d1/files/f1/permissions/pm1"},
		{name: "SharedDriveService.PatchFilePermission", call: func(c *Client) { NewSharedDriveService(c).PatchFilePermission("d1", "f1", "pm1", []byte(`{}`)) }, wantMethod: "PATCH", wantPath: "/sharedrives/d1/files/f1/permissions/pm1"},
		{name: "SharedDriveService.DeleteFilePermission", call: func(c *Client) { NewSharedDriveService(c).DeleteFilePermission("d1", "f1", "pm1") }, wantMethod: "DELETE", wantPath: "/sharedrives/d1/files/f1/permissions/pm1"},
		{name: "SharedDriveService.DeleteAllFilePermissions", call: func(c *Client) { NewSharedDriveService(c).DeleteAllFilePermissions("d1", "f1") }, wantMethod: "DELETE", wantPath: "/sharedrives/d1/files/f1/permissions"},
		{name: "SharedDriveService.EnableFilePermissions", call: func(c *Client) { NewSharedDriveService(c).EnableFilePermissions("d1", "f1") }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/permissions/enable"},
		{name: "SharedDriveService.DisableFilePermissions", call: func(c *Client) { NewSharedDriveService(c).DisableFilePermissions("d1", "f1") }, wantMethod: "POST", wantPath: "/sharedrives/d1/files/f1/permissions/disable"},

		// ─── SharedFolder (drive_sharedfolder.go) ───
		{name: "SharedFolderService.GetFolder", call: func(c *Client) { NewSharedFolderService(c).GetFolder("u1", "sf1") }, wantMethod: "GET", wantPath: "/users/u1/drive/sharedfolders/sf1"},
		{name: "SharedFolderService.LeaveFolder", call: func(c *Client) { NewSharedFolderService(c).LeaveFolder("u1", "sf1") }, wantMethod: "DELETE", wantPath: "/users/u1/drive/sharedfolders/sf1"},
		{name: "SharedFolderService.ListMembers", call: func(c *Client) { NewSharedFolderService(c).ListMembers("u1", "sf1") }, wantMethod: "GET", wantPath: "/users/u1/drive/sharedfolders/sf1/members"},
		{name: "SharedFolderService.ListFiles", call: func(c *Client) { NewSharedFolderService(c).ListFiles("u1", "sf1", "", 10) }, wantMethod: "GET", wantPath: "/users/u1/drive/sharedfolders/sf1/files"},
		{name: "SharedFolderService.CreateUploadURLInRoot", call: func(c *Client) { NewSharedFolderService(c).CreateUploadURLInRoot("u1", "sf1", map[string]interface{}{"fileName": "a.txt"}, 100) }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files"},
		{name: "SharedFolderService.CreateFolderInRoot", call: func(c *Client) { NewSharedFolderService(c).CreateFolderInRoot("u1", "sf1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/createfolder"},
		{name: "SharedFolderService.CreateSubFolder", call: func(c *Client) { NewSharedFolderService(c).CreateSubFolder("u1", "sf1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/createfolder"},
		{name: "SharedFolderService.GetFile", call: func(c *Client) { NewSharedFolderService(c).GetFile("u1", "sf1", "f1") }, wantMethod: "GET", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1"},
		{name: "SharedFolderService.ListFolderChildren", call: func(c *Client) { NewSharedFolderService(c).ListFolderChildren("u1", "sf1", "f1", "", 10) }, wantMethod: "GET", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/children"},
		{name: "SharedFolderService.DeleteFile", call: func(c *Client) { NewSharedFolderService(c).DeleteFile("u1", "sf1", "f1") }, wantMethod: "DELETE", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1"},
		{name: "SharedFolderService.CreateUploadURL", call: func(c *Client) { NewSharedFolderService(c).CreateUploadURL("u1", "sf1", "f1", map[string]interface{}{"fileName": "a.txt"}, 100) }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1"},
		{name: "SharedFolderService.CopyFile", call: func(c *Client) { NewSharedFolderService(c).CopyFile("u1", "sf1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/copy"},
		{name: "SharedFolderService.RenameFile", call: func(c *Client) { NewSharedFolderService(c).RenameFile("u1", "sf1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/rename"},
		{name: "SharedFolderService.MoveFile", call: func(c *Client) { NewSharedFolderService(c).MoveFile("u1", "sf1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/move"},
		{name: "SharedFolderService.ProtectFile", call: func(c *Client) { NewSharedFolderService(c).ProtectFile("u1", "sf1", "f1") }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/protect"},
		{name: "SharedFolderService.UnprotectFile", call: func(c *Client) { NewSharedFolderService(c).UnprotectFile("u1", "sf1", "f1") }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/unprotect"},
		{name: "SharedFolderService.LockFile", call: func(c *Client) { NewSharedFolderService(c).LockFile("u1", "sf1", "f1") }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/lock"},
		{name: "SharedFolderService.UnlockFile", call: func(c *Client) { NewSharedFolderService(c).UnlockFile("u1", "sf1", "f1") }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/unlock"},
		{name: "SharedFolderService.ListRevisions", call: func(c *Client) { NewSharedFolderService(c).ListRevisions("u1", "sf1", "f1", "", 10) }, wantMethod: "GET", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/revisions"},
		{name: "SharedFolderService.GetRevision", call: func(c *Client) { NewSharedFolderService(c).GetRevision("u1", "sf1", "f1", "r1") }, wantMethod: "GET", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/revisions/r1"},
		{name: "SharedFolderService.RestoreRevision", call: func(c *Client) { NewSharedFolderService(c).RestoreRevision("u1", "sf1", "f1", "r1") }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/revisions/r1/restore"},
		{name: "SharedFolderService.GetLinkSetting", call: func(c *Client) { NewSharedFolderService(c).GetLinkSetting("u1", "sf1") }, wantMethod: "GET", wantPath: "/users/u1/drive/sharedfolders/sf1/link-setting"},
		{name: "SharedFolderService.GetLink", call: func(c *Client) { NewSharedFolderService(c).GetLink("u1", "sf1", "f1") }, wantMethod: "GET", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/link"},
		{name: "SharedFolderService.CreateLink", call: func(c *Client) { NewSharedFolderService(c).CreateLink("u1", "sf1", "f1", []byte(`{}`)) }, wantMethod: "POST", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/link"},
		{name: "SharedFolderService.PatchLink", call: func(c *Client) { NewSharedFolderService(c).PatchLink("u1", "sf1", "f1", []byte(`{}`)) }, wantMethod: "PATCH", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/link"},
		{name: "SharedFolderService.DeleteLink", call: func(c *Client) { NewSharedFolderService(c).DeleteLink("u1", "sf1", "f1") }, wantMethod: "DELETE", wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/link"},

		// ─── SCIM (remaining) ───
		{
			name:       "ScimService.GetUser",
			call:       func(c *Client) { NewScimService(c).GetUser("u1") },
			wantMethod: "GET",
			wantPath:   "/Users/u1",
		},
		{
			name:       "ScimService.CreateUser",
			call:       func(c *Client) { NewScimService(c).CreateUser(map[string]interface{}{"userName": "test"}) },
			wantMethod: "POST",
			wantPath:   "/Users",
		},
		{
			name:       "ScimService.UpdateUser",
			call:       func(c *Client) { NewScimService(c).UpdateUser("u1", map[string]interface{}{"userName": "test2"}) },
			wantMethod: "PUT",
			wantPath:   "/Users/u1",
		},
		{
			name:       "ScimService.PatchUser",
			call:       func(c *Client) { NewScimService(c).PatchUser("u1", map[string]interface{}{"Operations": []string{}}) },
			wantMethod: "PATCH",
			wantPath:   "/Users/u1",
		},
		{
			name:       "ScimService.ListGroups",
			call:       func(c *Client) { NewScimService(c).ListGroups(0, 0, "") },
			wantMethod: "GET",
			wantPath:   "/Groups",
		},
		{
			name:       "ScimService.GetGroup",
			call:       func(c *Client) { NewScimService(c).GetGroup("g1") },
			wantMethod: "GET",
			wantPath:   "/Groups/g1",
		},
		{
			name:       "ScimService.UpdateGroup",
			call:       func(c *Client) { NewScimService(c).UpdateGroup("g1", map[string]interface{}{"displayName": "test2"}) },
			wantMethod: "PUT",
			wantPath:   "/Groups/g1",
		},
		{
			name:       "ScimService.PatchGroup",
			call:       func(c *Client) { NewScimService(c).PatchGroup("g1", map[string]interface{}{"Operations": []string{}}) },
			wantMethod: "PATCH",
			wantPath:   "/Groups/g1",
		},
		{
			name:       "ScimService.DeleteGroup",
			call:       func(c *Client) { NewScimService(c).DeleteGroup("g1") },
			wantMethod: "DELETE",
			wantPath:   "/Groups/g1",
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
