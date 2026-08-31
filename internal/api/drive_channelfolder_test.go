package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

type channelFolderCoverageSnapshot struct {
	SchemaVersion         string                          `json:"schema_version"`
	CapturedAt            string                          `json:"captured_at"`
	SourceURL             string                          `json:"source_url"`
	PathSourcePolicy      string                          `json:"path_source_policy"`
	PathSourceURLs        []string                        `json:"path_source_urls"`
	AuthenticationSubject string                          `json:"authentication_subject"`
	Scopes                []string                        `json:"scopes"`
	ExpectedEndpointCount int                             `json:"expected_endpoint_count"`
	Endpoints             []channelFolderCoverageEndpoint `json:"endpoints"`
}

type channelFolderCoverageEndpoint struct {
	Method       string `json:"method"`
	Path         string `json:"path"`
	APIMethod    string `json:"api_method"`
	SourceAnchor string `json:"source_anchor"`
	CLIStatus    string `json:"cli_status"`
}

type channelFolderCoverageCase struct {
	method       string
	pathTemplate string
	wantPath     string
	apiMethod    string
	call         func(*ChannelFolderService) error
}

func channelFolderCoverageCases() []channelFolderCoverageCase {
	responseError := func(_ *Response, err error) error { return err }
	return []channelFolderCoverageCase{
		{http.MethodGet, "/channel-folders", "/channel-folders", "ListChannelFolders", func(s *ChannelFolderService) error { return responseError(s.ListChannelFolders("", 0)) }},
		{http.MethodGet, "/channel-folders/{channelFolderId}", "/channel-folders/cf1", "GetChannelFolder", func(s *ChannelFolderService) error { return responseError(s.GetChannelFolder("cf1")) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files", "/channel-folders/cf1/files", "CreateRootUploadURL", func(s *ChannelFolderService) error {
			return responseError(s.CreateRootUploadURL("cf1", map[string]interface{}{"fileName": "report.txt"}, 10))
		}},
		{http.MethodGet, "/channel-folders/{channelFolderId}/files", "/channel-folders/cf1/files", "ListFiles", func(s *ChannelFolderService) error { return responseError(s.ListFiles("cf1", "", 0)) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/createfolder", "/channel-folders/cf1/files/createfolder", "CreateFolderInRoot", func(s *ChannelFolderService) error { return responseError(s.CreateFolderInRoot("cf1", []byte(`{}`))) }},
		{http.MethodGet, "/channel-folders/{channelFolderId}/files/{fileId}/children", "/channel-folders/cf1/files/f1/children", "ListFolderChildren", func(s *ChannelFolderService) error { return responseError(s.ListFolderChildren("cf1", "f1", "", 0)) }},
		{http.MethodGet, "/channel-folders/{channelFolderId}/files/{fileId}", "/channel-folders/cf1/files/f1", "GetFile", func(s *ChannelFolderService) error { return responseError(s.GetFile("cf1", "f1")) }},
		{http.MethodDelete, "/channel-folders/{channelFolderId}/files/{fileId}", "/channel-folders/cf1/files/f1", "DeleteFile", func(s *ChannelFolderService) error { return responseError(s.DeleteFile("cf1", "f1")) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/createfolder", "/channel-folders/cf1/files/f1/createfolder", "CreateSubFolder", func(s *ChannelFolderService) error {
			return responseError(s.CreateSubFolder("cf1", "f1", []byte(`{}`)))
		}},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/copy", "/channel-folders/cf1/files/f1/copy", "CopyFile", func(s *ChannelFolderService) error { return responseError(s.CopyFile("cf1", "f1", []byte(`{}`))) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/rename", "/channel-folders/cf1/files/f1/rename", "RenameFile", func(s *ChannelFolderService) error { return responseError(s.RenameFile("cf1", "f1", []byte(`{}`))) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/move", "/channel-folders/cf1/files/f1/move", "MoveFile", func(s *ChannelFolderService) error { return responseError(s.MoveFile("cf1", "f1", []byte(`{}`))) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/protect", "/channel-folders/cf1/files/f1/protect", "ProtectFile", func(s *ChannelFolderService) error { return responseError(s.ProtectFile("cf1", "f1")) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/unprotect", "/channel-folders/cf1/files/f1/unprotect", "UnprotectFile", func(s *ChannelFolderService) error { return responseError(s.UnprotectFile("cf1", "f1")) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/lock", "/channel-folders/cf1/files/f1/lock", "LockFile", func(s *ChannelFolderService) error { return responseError(s.LockFile("cf1", "f1")) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/unlock", "/channel-folders/cf1/files/f1/unlock", "UnlockFile", func(s *ChannelFolderService) error { return responseError(s.UnlockFile("cf1", "f1")) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}", "/channel-folders/cf1/files/f1", "CreateUploadURL", func(s *ChannelFolderService) error {
			return responseError(s.CreateUploadURL("cf1", "f1", map[string]interface{}{"fileName": "report.txt"}, 10))
		}},
		{http.MethodGet, "/channel-folders/{channelFolderId}/files/{fileId}/download", "/channel-folders/cf1/files/f1/download", "GetDownloadURL", func(s *ChannelFolderService) error { _, err := s.GetDownloadURL("cf1", "f1"); return err }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/permissions", "/channel-folders/cf1/files/f1/permissions", "CreatePermission", func(s *ChannelFolderService) error {
			return responseError(s.CreatePermission("cf1", "f1", []byte(`{}`)))
		}},
		{http.MethodGet, "/channel-folders/{channelFolderId}/files/{fileId}/permissions", "/channel-folders/cf1/files/f1/permissions", "ListPermissions", func(s *ChannelFolderService) error { return responseError(s.ListPermissions("cf1", "f1")) }},
		{http.MethodGet, "/channel-folders/{channelFolderId}/files/{fileId}/permissions/{permissionId}", "/channel-folders/cf1/files/f1/permissions/p1", "GetPermission", func(s *ChannelFolderService) error { return responseError(s.GetPermission("cf1", "f1", "p1")) }},
		{http.MethodPatch, "/channel-folders/{channelFolderId}/files/{fileId}/permissions/{permissionId}", "/channel-folders/cf1/files/f1/permissions/p1", "PatchPermission", func(s *ChannelFolderService) error {
			return responseError(s.PatchPermission("cf1", "f1", "p1", []byte(`{}`)))
		}},
		{http.MethodDelete, "/channel-folders/{channelFolderId}/files/{fileId}/permissions/{permissionId}", "/channel-folders/cf1/files/f1/permissions/p1", "DeletePermission", func(s *ChannelFolderService) error { return responseError(s.DeletePermission("cf1", "f1", "p1")) }},
		{http.MethodDelete, "/channel-folders/{channelFolderId}/files/{fileId}/permissions", "/channel-folders/cf1/files/f1/permissions", "DeleteAllPermissions", func(s *ChannelFolderService) error { return responseError(s.DeleteAllPermissions("cf1", "f1")) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/permissions/enable", "/channel-folders/cf1/files/f1/permissions/enable", "EnablePermissions", func(s *ChannelFolderService) error { return responseError(s.EnablePermissions("cf1", "f1")) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/permissions/disable", "/channel-folders/cf1/files/f1/permissions/disable", "DisablePermissions", func(s *ChannelFolderService) error { return responseError(s.DisablePermissions("cf1", "f1")) }},
		{http.MethodGet, "/channel-folders/{channelFolderId}/files/{fileId}/revisions", "/channel-folders/cf1/files/f1/revisions", "ListRevisions", func(s *ChannelFolderService) error { return responseError(s.ListRevisions("cf1", "f1", "", 0)) }},
		{http.MethodGet, "/channel-folders/{channelFolderId}/files/{fileId}/revisions/{revisionId}", "/channel-folders/cf1/files/f1/revisions/r1", "GetRevision", func(s *ChannelFolderService) error { return responseError(s.GetRevision("cf1", "f1", "r1")) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/revisions/{revisionId}/restore", "/channel-folders/cf1/files/f1/revisions/r1/restore", "RestoreRevision", func(s *ChannelFolderService) error { return responseError(s.RestoreRevision("cf1", "f1", "r1")) }},
		{http.MethodGet, "/channel-folders/{channelFolderId}/files/{fileId}/revisions/{revisionId}/download", "/channel-folders/cf1/files/f1/revisions/r1/download", "GetRevisionDownloadURL", func(s *ChannelFolderService) error { _, err := s.GetRevisionDownloadURL("cf1", "f1", "r1"); return err }},
		{http.MethodGet, "/channel-folders/{channelFolderId}/link-setting", "/channel-folders/cf1/link-setting", "GetLinkSetting", func(s *ChannelFolderService) error { return responseError(s.GetLinkSetting("cf1")) }},
		{http.MethodGet, "/channel-folders/{channelFolderId}/files/{fileId}/link", "/channel-folders/cf1/files/f1/link", "GetLink", func(s *ChannelFolderService) error { return responseError(s.GetLink("cf1", "f1")) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/files/{fileId}/link", "/channel-folders/cf1/files/f1/link", "CreateLink", func(s *ChannelFolderService) error { return responseError(s.CreateLink("cf1", "f1", []byte(`{}`))) }},
		{http.MethodPatch, "/channel-folders/{channelFolderId}/files/{fileId}/link", "/channel-folders/cf1/files/f1/link", "PatchLink", func(s *ChannelFolderService) error { return responseError(s.PatchLink("cf1", "f1", []byte(`{}`))) }},
		{http.MethodDelete, "/channel-folders/{channelFolderId}/files/{fileId}/link", "/channel-folders/cf1/files/f1/link", "DeleteLink", func(s *ChannelFolderService) error { return responseError(s.DeleteLink("cf1", "f1")) }},
		{http.MethodGet, "/channel-folders/{channelFolderId}/trash-files", "/channel-folders/cf1/trash-files", "ListTrashFiles", func(s *ChannelFolderService) error { return responseError(s.ListTrashFiles("cf1", "", 0)) }},
		{http.MethodPost, "/channel-folders/{channelFolderId}/trash-files/{trashFileId}/restore", "/channel-folders/cf1/trash-files/t1/restore", "RestoreTrashFile", func(s *ChannelFolderService) error { return responseError(s.RestoreTrashFile("cf1", "t1")) }},
		{http.MethodDelete, "/channel-folders/{channelFolderId}/trash-files/{trashFileId}", "/channel-folders/cf1/trash-files/t1", "DeleteTrashFile", func(s *ChannelFolderService) error { return responseError(s.DeleteTrashFile("cf1", "t1")) }},
	}
}

func loadChannelFolderCoverageSnapshot(t *testing.T) channelFolderCoverageSnapshot {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "docs", "baselines", "channel-folder-api-coverage.json"))
	if err != nil {
		t.Fatalf("read channel-folder coverage snapshot: %v", err)
	}
	var snapshot channelFolderCoverageSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatalf("decode channel-folder coverage snapshot: %v", err)
	}
	return snapshot
}

func TestChannelFolderOfficialCoverage(t *testing.T) {
	snapshot := loadChannelFolderCoverageSnapshot(t)
	cases := channelFolderCoverageCases()
	if snapshot.SchemaVersion != "channel-folder-api-coverage-v1" || snapshot.CapturedAt != "2026-08-31" {
		t.Fatalf("snapshot identity = %q %q", snapshot.SchemaVersion, snapshot.CapturedAt)
	}
	if snapshot.SourceURL != "https://developers.worksmobile.com/kr/docs/drive" || snapshot.AuthenticationSubject != "member-access-token-only" {
		t.Fatalf("snapshot source/auth = %q %q", snapshot.SourceURL, snapshot.AuthenticationSubject)
	}
	wantPathSources := []string{
		"https://developers.worksmobile.com/kr/docs/channel-folder-file-list",
		"https://developers.worksmobile.com/kr/docs/channel-folder-file-permission-create",
		"https://developers.worksmobile.com/kr/docs/channel-folder-file-revision-list",
		"https://developers.worksmobile.com/kr/docs/channel-folder-trash-file-list",
	}
	if snapshot.PathSourcePolicy != "individual-endpoint-pages-override-overview-table" || len(snapshot.PathSourceURLs) != len(wantPathSources) {
		t.Fatalf("snapshot path sources = %q %#v", snapshot.PathSourcePolicy, snapshot.PathSourceURLs)
	}
	for index, want := range wantPathSources {
		if snapshot.PathSourceURLs[index] != want {
			t.Fatalf("snapshot path source %d = %q, want %q", index, snapshot.PathSourceURLs[index], want)
		}
	}
	if len(snapshot.Scopes) != 2 || snapshot.Scopes[0] != "file.read" || snapshot.Scopes[1] != "group.folder.read" {
		t.Fatalf("snapshot scopes = %#v", snapshot.Scopes)
	}
	if snapshot.ExpectedEndpointCount != 38 || len(snapshot.Endpoints) != snapshot.ExpectedEndpointCount || len(cases) != snapshot.ExpectedEndpointCount {
		t.Fatalf("coverage counts: expected=%d snapshot=%d cases=%d", snapshot.ExpectedEndpointCount, len(snapshot.Endpoints), len(cases))
	}

	seen := make(map[string]struct{}, snapshot.ExpectedEndpointCount)
	allowedAnchors := map[string]struct{}{
		"manage-channelfolder":                   {},
		"manage-channelfolder-rootfolder":        {},
		"manage-channelfolder-file-folder":       {},
		"manage-channelfolder-folder-permission": {},
		"manage-channelfolder-version":           {},
		"manage-link-channel":                    {},
		"manage-channellfolder-trashfile":        {},
	}
	usedAnchors := make(map[string]struct{}, len(allowedAnchors))
	for index, test := range cases {
		endpoint := snapshot.Endpoints[index]
		if endpoint.Method != test.method || endpoint.Path != test.pathTemplate || endpoint.APIMethod != test.apiMethod {
			t.Fatalf("coverage row %d = %s %s %s, want %s %s %s", index+1, endpoint.Method, endpoint.Path, endpoint.APIMethod, test.method, test.pathTemplate, test.apiMethod)
		}
		if _, ok := allowedAnchors[endpoint.SourceAnchor]; !ok || endpoint.CLIStatus != "pending:TASK-009" {
			t.Fatalf("coverage row %d lacks source/CLI status: %#v", index+1, endpoint)
		}
		usedAnchors[endpoint.SourceAnchor] = struct{}{}
		key := endpoint.Method + " " + endpoint.Path
		if _, exists := seen[key]; exists {
			t.Fatalf("duplicate official endpoint: %s", key)
		}
		seen[key] = struct{}{}

		t.Run(test.apiMethod, func(t *testing.T) {
			var gotMethod, gotPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				gotMethod = request.Method
				gotPath = request.URL.Path
				if test.apiMethod == "GetDownloadURL" || test.apiMethod == "GetRevisionDownloadURL" {
					w.Header().Set("Location", "https://download.example.com/file")
					w.WriteHeader(http.StatusFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			defer server.Close()

			client := NewClient(server.URL, &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(time.Hour)}, nil)
			if err := test.call(NewChannelFolderService(client)); err != nil {
				t.Fatalf("%s failed: %v", test.apiMethod, err)
			}
			if gotMethod != test.method || gotPath != test.wantPath {
				t.Fatalf("request = %s %s, want %s %s", gotMethod, gotPath, test.method, test.wantPath)
			}
		})
	}
	if len(usedAnchors) != len(allowedAnchors) {
		t.Fatalf("official source anchors used = %#v", usedAnchors)
	}
}

func TestChannelFolderServiceEscapingPaginationAndUploadBody(t *testing.T) {
	t.Run("escaping and pagination", func(t *testing.T) {
		var gotPath, gotQuery string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			gotPath = request.URL.EscapedPath()
			gotQuery = request.URL.RawQuery
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(time.Hour)}, nil)
		_, err := NewChannelFolderService(client).ListFolderChildren("channel/one", "file/one", "cursor/one", 25)
		if err != nil {
			t.Fatalf("ListFolderChildren failed: %v", err)
		}
		if gotPath != "/channel-folders/channel%2Fone/files/file%2Fone/children" || gotQuery != "count=25&cursor=cursor%2Fone" {
			t.Fatalf("request = %s?%s", gotPath, gotQuery)
		}
	})

	t.Run("root upload file size", func(t *testing.T) {
		var gotBody map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			defer request.Body.Close()
			raw, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			if err := json.Unmarshal(raw, &gotBody); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client := NewClient(server.URL, &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(time.Hour)}, nil)
		_, err := NewChannelFolderService(client).CreateRootUploadURL("cf1", map[string]interface{}{"fileName": "report.txt"}, 1234)
		if err != nil {
			t.Fatalf("CreateRootUploadURL failed: %v", err)
		}
		if gotBody["fileName"] != "report.txt" || gotBody["fileSize"] != float64(1234) {
			t.Fatalf("upload body = %#v", gotBody)
		}
	})
}
