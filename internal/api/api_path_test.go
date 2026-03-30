package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestAPIEndpointPaths(t *testing.T) {
	tests := []struct {
		name       string
		call       func(client *Client)
		wantMethod string
		wantPath   string
	}{
		{
			name: "SharedFolderService.GetFolder",
			call: func(c *Client) {
				svc := NewSharedFolderService(c)
				svc.GetFolder("u1", "sf1")
			},
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/sharedfolders/sf1",
		},
		{
			name: "SharedFolderService.CreateUploadURLInRoot",
			call: func(c *Client) {
				svc := NewSharedFolderService(c)
				svc.CreateUploadURLInRoot("u1", "sf1", map[string]interface{}{"fileName": "test.txt"}, 1024)
			},
			wantMethod: "POST",
			wantPath:   "/users/u1/drive/sharedfolders/sf1/files",
		},
		{
			name: "SharedFolderService.ListFolderChildren",
			call: func(c *Client) {
				svc := NewSharedFolderService(c)
				svc.ListFolderChildren("u1", "sf1", "folder1", "", 0)
			},
			wantMethod: "GET",
			wantPath:   "/users/u1/drive/sharedfolders/sf1/files/folder1/children",
		},
		{
			name: "SharedFolderService.PatchLink",
			call: func(c *Client) {
				svc := NewSharedFolderService(c)
				svc.PatchLink("u1", "sf1", "f1", []byte(`{}`))
			},
			wantMethod: "PATCH",
			wantPath:   "/users/u1/drive/sharedfolders/sf1/files/f1/link",
		},
		{
			name: "GroupFolderService.CopyFile",
			call: func(c *Client) {
				svc := NewGroupFolderService(c)
				svc.CopyFile("g1", "f1", []byte(`{}`))
			},
			wantMethod: "POST",
			wantPath:   "/groups/g1/folder/files/f1/copy",
		},
		{
			name: "SharedDriveService.CreateDrive",
			call: func(c *Client) {
				svc := NewSharedDriveService(c)
				svc.CreateDrive([]byte(`{"name":"test"}`))
			},
			wantMethod: "POST",
			wantPath:   "/sharedrives",
		},
		{
			name: "SharedDriveService.EnablePermissions",
			call: func(c *Client) {
				svc := NewSharedDriveService(c)
				svc.EnablePermissions("d1")
			},
			wantMethod: "POST",
			wantPath:   "/sharedrives/d1/permissions/enable",
		},
		{
			name: "SecurityService.EnableExternalBrowser",
			call: func(c *Client) {
				svc := NewSecurityService(c)
				svc.EnableExternalBrowser()
			},
			wantMethod: "POST",
			wantPath:   "/security/external-browser/enable",
		},
		{
			name: "CalendarService.CreateCalendar",
			call: func(c *Client) {
				svc := NewCalendarService(c)
				svc.CreateCalendar(map[string]interface{}{"calendarName": "test"})
			},
			wantMethod: "POST",
			wantPath:   "/calendars",
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

// TestAPIEndpointPaths_DownloadURL tests endpoints that use GetDownloadURL
// (which uses noRedirectClient and expects 302 responses).
func TestAPIEndpointPaths_DownloadURL(t *testing.T) {
	tests := []struct {
		name     string
		call     func(client *Client)
		wantPath string
	}{
		{
			name: "SharedFolderService.GetRevisionDownloadURL",
			call: func(c *Client) {
				svc := NewSharedFolderService(c)
				svc.GetRevisionDownloadURL("u1", "sf1", "f1", "rev1")
			},
			wantPath: "/users/u1/drive/sharedfolders/sf1/files/f1/revisions/rev1/download",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				// Return 302 with Location to satisfy GetDownloadURL
				w.Header().Set("Location", "https://example.com/download")
				w.WriteHeader(302)
			}))
			defer srv.Close()

			token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
			client := NewClient(srv.URL, token, nil)

			tt.call(client)

			if gotPath != tt.wantPath {
				t.Errorf("path = %s, want %s", gotPath, tt.wantPath)
			}
		})
	}
}
