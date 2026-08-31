package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
)

var channelFolderCommandByAPIMethod = map[string]string{
	"ListChannelFolders":     "list",
	"GetChannelFolder":       "get",
	"CreateRootUploadURL":    "upload",
	"ListFiles":              "files",
	"CreateFolderInRoot":     "mkdir",
	"ListFolderChildren":     "files",
	"GetFile":                "get-file",
	"DeleteFile":             "delete",
	"CreateSubFolder":        "mkdir",
	"CopyFile":               "copy",
	"RenameFile":             "rename",
	"MoveFile":               "move",
	"ProtectFile":            "protect",
	"UnprotectFile":          "unprotect",
	"LockFile":               "lock",
	"UnlockFile":             "unlock",
	"CreateUploadURL":        "upload",
	"GetDownloadURL":         "download",
	"CreatePermission":       "permission create",
	"ListPermissions":        "permission list",
	"GetPermission":          "permission get",
	"PatchPermission":        "permission update",
	"DeletePermission":       "permission delete",
	"DeleteAllPermissions":   "permission delete-all",
	"EnablePermissions":      "permission enable",
	"DisablePermissions":     "permission disable",
	"ListRevisions":          "revision list",
	"GetRevision":            "revision get",
	"RestoreRevision":        "revision restore",
	"GetRevisionDownloadURL": "revision download",
	"GetLinkSetting":         "link-setting",
	"GetLink":                "link get",
	"CreateLink":             "link create",
	"PatchLink":              "link update",
	"DeleteLink":             "link delete",
	"ListTrashFiles":         "trash-list",
	"RestoreTrashFile":       "trash-restore",
	"DeleteTrashFile":        "trash-delete",
}

var channelFolderReadAPIMethods = map[string]struct{}{
	"ListChannelFolders":     {},
	"GetChannelFolder":       {},
	"ListFiles":              {},
	"ListFolderChildren":     {},
	"GetFile":                {},
	"GetDownloadURL":         {},
	"ListPermissions":        {},
	"GetPermission":          {},
	"ListRevisions":          {},
	"GetRevision":            {},
	"GetRevisionDownloadURL": {},
	"GetLinkSetting":         {},
	"GetLink":                {},
	"ListTrashFiles":         {},
}

func TestChannelFolderOfficialEndpointsHaveRegisteredCommands(t *testing.T) {
	var snapshot struct {
		ExpectedEndpointCount int `json:"expected_endpoint_count"`
		Endpoints             []struct {
			APIMethod string `json:"api_method"`
			CLIStatus string `json:"cli_status"`
		} `json:"endpoints"`
	}
	raw, err := os.ReadFile(filepath.Join("..", "docs", "baselines", "channel-folder-api-coverage.json"))
	if err != nil {
		t.Fatalf("read channel folder baseline: %v", err)
	}
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatalf("decode channel folder baseline: %v", err)
	}
	if snapshot.ExpectedEndpointCount != 38 || len(snapshot.Endpoints) != 38 || len(channelFolderCommandByAPIMethod) != 38 {
		t.Fatalf("coverage counts = expected:%d endpoints:%d commands:%d", snapshot.ExpectedEndpointCount, len(snapshot.Endpoints), len(channelFolderCommandByAPIMethod))
	}

	seen := make(map[string]struct{}, len(snapshot.Endpoints))
	for _, endpoint := range snapshot.Endpoints {
		commandPath, ok := channelFolderCommandByAPIMethod[endpoint.APIMethod]
		if !ok {
			t.Fatalf("API method %s has no CLI command", endpoint.APIMethod)
		}
		if _, duplicate := seen[endpoint.APIMethod]; duplicate {
			t.Fatalf("duplicate API method %s", endpoint.APIMethod)
		}
		seen[endpoint.APIMethod] = struct{}{}
		if endpoint.CLIStatus != "implemented:TASK-009" {
			t.Fatalf("API method %s CLI status = %q", endpoint.APIMethod, endpoint.CLIStatus)
		}

		command, remaining, err := driveChannelCmd.Find(strings.Fields(commandPath))
		if err != nil || command == driveChannelCmd || len(remaining) != 0 {
			t.Fatalf("API method %s command %q not registered: command=%v remaining=%v err=%v", endpoint.APIMethod, commandPath, command, remaining, err)
		}
		if _, readOnly := channelFolderReadAPIMethods[endpoint.APIMethod]; readOnly {
			for _, want := range []string{"file.read", "group.folder.read"} {
				if !strings.Contains(command.Long, want) {
					t.Fatalf("read command %q help missing %q", commandPath, want)
				}
			}
		} else {
			for _, want := range []string{"file", "group.folder"} {
				if !strings.Contains(command.Long, want) {
					t.Fatalf("write command %q help missing %q", commandPath, want)
				}
			}
		}
	}
}

func TestJourneyChannelFolderCommandFamilies(t *testing.T) {
	tests := []struct {
		apiMethod  string
		args       []string
		method     string
		path       string
		rawQuery   string
		body       string
		statusCode int
		headers    map[string]string
		response   string
		stdoutWant string
	}{
		{apiMethod: "ListChannelFolders", args: []string{"list"}, method: http.MethodGet, path: "/channel-folders", response: `{"channelFolders":[{"channelFolderId":"cf1","name":"일반","createdTime":"2026-08-31T10:00:00+09:00"}]}`, stdoutWant: "channelFolderId"},
		{apiMethod: "GetChannelFolder", args: []string{"get", "cf1"}, method: http.MethodGet, path: "/channel-folders/cf1", response: `{}`},
		{apiMethod: "ListFiles", args: []string{"files", "cf1", "--count", "2"}, method: http.MethodGet, path: "/channel-folders/cf1/files", rawQuery: "count=2", response: `{"files":[]}`},
		{apiMethod: "CreateFolderInRoot", args: []string{"mkdir", "cf1", "--json", `{"fileName":"new-folder"}`}, method: http.MethodPost, path: "/channel-folders/cf1/files/createfolder", body: `{"fileName":"new-folder"}`, response: `{}`},
		{apiMethod: "ListFolderChildren", args: []string{"files", "cf1", "--folder", "folder1", "--cursor", "next", "--count", "2"}, method: http.MethodGet, path: "/channel-folders/cf1/files/folder1/children", rawQuery: "count=2&cursor=next", response: `{"files":[]}`},
		{apiMethod: "GetFile", args: []string{"get-file", "cf1", "file1"}, method: http.MethodGet, path: "/channel-folders/cf1/files/file1", response: `{}`},
		{apiMethod: "DeleteFile", args: []string{"delete", "cf1", "file1"}, method: http.MethodDelete, path: "/channel-folders/cf1/files/file1", response: `{}`},
		{apiMethod: "CreateSubFolder", args: []string{"mkdir", "cf1", "--parent", "folder1", "--json", `{"fileName":"new-folder"}`}, method: http.MethodPost, path: "/channel-folders/cf1/files/folder1/createfolder", body: `{"fileName":"new-folder"}`, response: `{}`},
		{apiMethod: "CopyFile", args: []string{"copy", "cf1", "file1", "--json", `{"destinationFileId":"folder2"}`}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/copy", body: `{"destinationFileId":"folder2"}`, response: `{}`},
		{apiMethod: "RenameFile", args: []string{"rename", "cf1", "file1", "--json", `{"fileName":"renamed.txt"}`}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/rename", body: `{"fileName":"renamed.txt"}`, response: `{}`},
		{apiMethod: "MoveFile", args: []string{"move", "cf1", "file1", "--json", `{"destinationFileId":"folder2"}`}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/move", body: `{"destinationFileId":"folder2"}`, response: `{}`},
		{apiMethod: "ProtectFile", args: []string{"protect", "cf1", "file1"}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/protect", response: `{}`},
		{apiMethod: "UnprotectFile", args: []string{"unprotect", "cf1", "file1"}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/unprotect", response: `{}`},
		{apiMethod: "LockFile", args: []string{"lock", "cf1", "file1"}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/lock", response: `{}`},
		{apiMethod: "UnlockFile", args: []string{"unlock", "cf1", "file1"}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/unlock", response: `{}`},
		{apiMethod: "GetDownloadURL", args: []string{"download", "cf1", "file1"}, method: http.MethodGet, path: "/channel-folders/cf1/files/file1/download", statusCode: http.StatusFound, headers: map[string]string{"Location": "https://download.example.com/file"}, stdoutWant: "download.example.com"},
		{apiMethod: "CreatePermission", args: []string{"permission", "create", "cf1", "file1", "--json", `{"userId":"user1","type":"READ"}`}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/permissions", body: `{"userId":"user1","type":"READ"}`, response: `{}`},
		{apiMethod: "ListPermissions", args: []string{"permission", "list", "cf1", "file1"}, method: http.MethodGet, path: "/channel-folders/cf1/files/file1/permissions", response: `{"permissions":[]}`},
		{apiMethod: "GetPermission", args: []string{"permission", "get", "cf1", "file1", "permission1"}, method: http.MethodGet, path: "/channel-folders/cf1/files/file1/permissions/permission1", response: `{}`},
		{apiMethod: "PatchPermission", args: []string{"permission", "update", "cf1", "file1", "permission1", "--json", `{"type":"READ"}`}, method: http.MethodPatch, path: "/channel-folders/cf1/files/file1/permissions/permission1", body: `{"type":"READ"}`, response: `{}`},
		{apiMethod: "DeletePermission", args: []string{"permission", "delete", "cf1", "file1", "permission1"}, method: http.MethodDelete, path: "/channel-folders/cf1/files/file1/permissions/permission1", response: `{}`},
		{apiMethod: "DeleteAllPermissions", args: []string{"permission", "delete-all", "cf1", "file1"}, method: http.MethodDelete, path: "/channel-folders/cf1/files/file1/permissions", response: `{}`},
		{apiMethod: "EnablePermissions", args: []string{"permission", "enable", "cf1", "file1"}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/permissions/enable", response: `{}`},
		{apiMethod: "DisablePermissions", args: []string{"permission", "disable", "cf1", "file1"}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/permissions/disable", response: `{}`},
		{apiMethod: "ListRevisions", args: []string{"revision", "list", "cf1", "file1", "--cursor", "next", "--count", "2"}, method: http.MethodGet, path: "/channel-folders/cf1/files/file1/revisions", rawQuery: "count=2&cursor=next", response: `{"revisions":[]}`},
		{apiMethod: "GetRevision", args: []string{"revision", "get", "cf1", "file1", "rev1"}, method: http.MethodGet, path: "/channel-folders/cf1/files/file1/revisions/rev1", response: `{}`},
		{apiMethod: "RestoreRevision", args: []string{"revision", "restore", "cf1", "file1", "rev1"}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/revisions/rev1/restore", response: `{}`},
		{apiMethod: "GetRevisionDownloadURL", args: []string{"revision", "download", "cf1", "file1", "rev1"}, method: http.MethodGet, path: "/channel-folders/cf1/files/file1/revisions/rev1/download", statusCode: http.StatusFound, headers: map[string]string{"Location": "https://download.example.com/revision"}, stdoutWant: "download.example.com"},
		{apiMethod: "GetLinkSetting", args: []string{"link-setting", "cf1"}, method: http.MethodGet, path: "/channel-folders/cf1/link-setting", response: `{}`},
		{apiMethod: "GetLink", args: []string{"link", "get", "cf1", "file1"}, method: http.MethodGet, path: "/channel-folders/cf1/files/file1/link", response: `{}`},
		{apiMethod: "CreateLink", args: []string{"link", "create", "cf1", "file1", "--json", `{"scope":"DOMAIN"}`}, method: http.MethodPost, path: "/channel-folders/cf1/files/file1/link", body: `{"scope":"DOMAIN"}`, response: `{}`},
		{apiMethod: "PatchLink", args: []string{"link", "update", "cf1", "file1", "--json", `{"scope":"DOMAIN"}`}, method: http.MethodPatch, path: "/channel-folders/cf1/files/file1/link", body: `{"scope":"DOMAIN"}`, response: `{}`},
		{apiMethod: "DeleteLink", args: []string{"link", "delete", "cf1", "file1"}, method: http.MethodDelete, path: "/channel-folders/cf1/files/file1/link", response: `{}`},
		{apiMethod: "ListTrashFiles", args: []string{"trash-list", "cf1", "--cursor", "next", "--count", "2"}, method: http.MethodGet, path: "/channel-folders/cf1/trash-files", rawQuery: "count=2&cursor=next", response: `{"trashFiles":[]}`},
		{apiMethod: "RestoreTrashFile", args: []string{"trash-restore", "cf1", "trash1"}, method: http.MethodPost, path: "/channel-folders/cf1/trash-files/trash1/restore", response: `{}`},
		{apiMethod: "DeleteTrashFile", args: []string{"trash-delete", "cf1", "trash1"}, method: http.MethodDelete, path: "/channel-folders/cf1/trash-files/trash1", response: `{}`},
	}

	covered := map[string]struct{}{
		"CreateRootUploadURL": {},
		"CreateUploadURL":     {},
	}
	for _, test := range tests {
		if _, duplicate := covered[test.apiMethod]; duplicate {
			t.Fatalf("duplicate request case for %s", test.apiMethod)
		}
		covered[test.apiMethod] = struct{}{}
		t.Run(test.apiMethod, func(t *testing.T) {
			harness := clitest.NewHarness(t)
			installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
			installJourneyFixture(t, harness.HomeDir(), "task/search/token.json", harness.TokenPath())
			statusCode := test.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}
			server := harness.StartScriptedServer([]clitest.ResponseScript{{
				StatusCode: statusCode,
				Headers:    test.headers,
				Body:       test.response,
			}})
			defer server.Close()
			setAPIBaseURL(t, server.URL)

			args := append([]string{"drive", "channel"}, test.args...)
			result, err := harness.Run(args, newRootCommandRunner(t))
			if err != nil {
				clitest.Fatalf(t, clitest.SetupFailure, "%s failed: %v", test.apiMethod, err)
			}
			if result.Stderr != "" {
				clitest.Fatalf(t, clitest.UXContractFailure, "%s stderr = %q", test.apiMethod, result.Stderr)
			}
			if test.stdoutWant != "" && !strings.Contains(result.Stdout, test.stdoutWant) {
				clitest.Fatalf(t, clitest.UXContractFailure, "%s stdout = %q, want %q", test.apiMethod, result.Stdout, test.stdoutWant)
			}
			logs := harness.RequestLogs()
			if len(logs) != 1 {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request count = %d", test.apiMethod, len(logs))
			}
			if logs[0].Method != test.method || logs[0].Path != test.path || logs[0].RawQuery != test.rawQuery || logs[0].Body != test.body {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request = %s %s?%s body=%s", test.apiMethod, logs[0].Method, logs[0].Path, logs[0].RawQuery, logs[0].Body)
			}
		})
	}

	if len(covered) != len(channelFolderCommandByAPIMethod) {
		t.Fatalf("actual request coverage = %d, want %d", len(covered), len(channelFolderCommandByAPIMethod))
	}
	for apiMethod := range channelFolderCommandByAPIMethod {
		if _, ok := covered[apiMethod]; !ok {
			t.Fatalf("API method %s has no actual request test", apiMethod)
		}
	}
}

func TestJourneyChannelFolderUploadsUsingOfficialHost(t *testing.T) {
	tests := []struct {
		apiMethod string
		folder    string
		path      string
	}{
		{apiMethod: "CreateRootUploadURL", path: "/v1.0/channel-folders/cf1/files"},
		{apiMethod: "CreateUploadURL", folder: "folder1", path: "/v1.0/channel-folders/cf1/files/folder1"},
	}

	for _, test := range tests {
		t.Run(test.apiMethod, func(t *testing.T) {
			harness := clitest.NewHarness(t)
			installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
			installJourneyFixture(t, harness.HomeDir(), "task/search/token.json", harness.TokenPath())
			filePath := filepath.Join(harness.HomeDir(), "report.txt")
			fileBody := []byte("channel report")
			if err := os.WriteFile(filePath, fileBody, 0o600); err != nil {
				t.Fatalf("write upload fixture: %v", err)
			}

			var createBody map[string]interface{}
			var uploadedBody string
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("read request body: %v", err)
				}
				switch {
				case r.Method == http.MethodPost && r.Host == "www.worksapis.com" && r.URL.Path == test.path:
					if err := json.Unmarshal(body, &createBody); err != nil {
						t.Fatalf("decode upload URL request: %v", err)
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, `{"uploadUrl":"https://apis-storage.worksmobile.com/upload"}`)
				case r.Method == http.MethodPut && r.Host == "apis-storage.worksmobile.com" && r.URL.Path == "/upload":
					uploadedBody = string(body)
					w.WriteHeader(http.StatusOK)
				default:
					t.Fatalf("unexpected request: %s %s host=%s body=%s", r.Method, r.URL.Path, r.Host, body)
				}
			}))
			defer server.Close()

			installHTTPSIntercept(t, server.Listener.Addr().String())
			t.Setenv("HTTPS_PROXY", "")
			t.Setenv("HTTP_PROXY", "")
			t.Setenv("ALL_PROXY", "")
			setAPIBaseURL(t, "https://www.worksapis.com/v1.0")

			args := []string{"drive", "channel", "upload", "cf1", "--file", filePath}
			if test.folder != "" {
				args = append(args, "--folder", test.folder)
			}
			result, err := harness.Run(args, newRootCommandRunner(t))
			if err != nil {
				t.Fatalf("%s failed: %v", test.apiMethod, err)
			}
			if strings.Contains(result.Stdout, "uploadUrl") || strings.Contains(result.Stdout, "apis-storage.worksmobile.com") {
				t.Fatalf("upload URL leaked in stdout: %s", result.Stdout)
			}
			if !strings.Contains(result.Stdout, `"uploaded": true`) && !strings.Contains(result.Stdout, `"uploaded":true`) {
				t.Fatalf("stdout missing upload marker: %s", result.Stdout)
			}
			if createBody["fileName"] != "report.txt" || createBody["fileSize"] != float64(len(fileBody)) {
				t.Fatalf("create upload body = %#v", createBody)
			}
			if uploadedBody != string(fileBody) {
				t.Fatalf("uploaded body = %q, want %q", uploadedBody, fileBody)
			}
		})
	}
}

func TestJourneyChannelFolderListRejectsPositionalArguments(t *testing.T) {
	harness := clitest.NewHarness(t)
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
	installJourneyFixture(t, harness.HomeDir(), "task/search/token.json", harness.TokenPath())
	server := harness.StartScriptedServer(nil)
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	_, err := harness.Run([]string{"drive", "channel", "list", "unexpected"}, newRootCommandRunner(t))
	if err == nil {
		t.Fatalf("error = %v, want no-args rejection", err)
	}
	if len(harness.RequestLogs()) != 0 {
		clitest.Fatalf(t, clitest.SideEffectFailure, "invalid list arguments sent network requests: %#v", harness.RequestLogs())
	}
}

func TestJourneyChannelFolderTableOutput(t *testing.T) {
	harness := clitest.NewHarness(t)
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
	installJourneyFixture(t, harness.HomeDir(), "task/search/token.json", harness.TokenPath())
	server := harness.StartScriptedServer([]clitest.ResponseScript{{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"files":[{"fileId":"file1","fileName":"보고서.pdf","fileType":"DOC","modifiedTime":"2026-08-31T10:00:00+09:00"}]}`,
	}})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	result, err := harness.Run([]string{"drive", "channel", "files", "cf1", "--output", "table"}, newRootCommandRunner(t))
	if err != nil {
		t.Fatalf("channel files table failed: %v", err)
	}
	for _, want := range []string{"fileId", "fileName", "fileType", "modifiedTime", "file1", "보고서.pdf"} {
		if !strings.Contains(result.Stdout, want) {
			t.Fatalf("stdout = %q, want %q", result.Stdout, want)
		}
	}
}

func TestJourneyChannelFolderRejectsServiceAccountBeforeNetwork(t *testing.T) {
	homeDir := setupTestEnv(t)
	writeTestConfig(t, homeDir)

	_, err := runCLI(t, "drive", "channel", "list")
	if err == nil || !strings.Contains(err.Error(), "구성원 계정 Access Token만 지원합니다") {
		t.Fatalf("error = %v, want service-account rejection", err)
	}
}

func TestJourneyChannelFolderUploadDryRunDoesNotSendNetworkRequest(t *testing.T) {
	harness := clitest.NewHarness(t)
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
	installJourneyFixture(t, harness.HomeDir(), "task/search/token.json", harness.TokenPath())
	filePath := filepath.Join(harness.HomeDir(), "report.txt")
	if err := os.WriteFile(filePath, []byte("channel report"), 0o600); err != nil {
		t.Fatalf("write upload fixture: %v", err)
	}
	server := harness.StartScriptedServer(nil)
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	result, err := harness.Run([]string{"drive", "channel", "upload", "cf1", "--folder", "folder1", "--file", filePath, "--dry-run"}, newRootCommandRunner(t))
	if err != nil {
		t.Fatalf("channel upload dry-run failed: %v", err)
	}
	if len(harness.RequestLogs()) != 0 {
		clitest.Fatalf(t, clitest.SideEffectFailure, "dry-run sent network requests: %#v", harness.RequestLogs())
	}
	for _, want := range []string{"dry_run", "presigned_upload", "report.txt"} {
		if !strings.Contains(result.Stdout, want) {
			clitest.Fatalf(t, clitest.UXContractFailure, "stdout = %q, want %q", result.Stdout, want)
		}
	}
}

func TestJourneyChannelFolderPropagatesAPIError(t *testing.T) {
	harness := clitest.NewHarness(t)
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
	installJourneyFixture(t, harness.HomeDir(), "task/search/token.json", harness.TokenPath())
	server := harness.StartScriptedServer([]clitest.ResponseScript{{
		StatusCode: http.StatusForbidden,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"code":"FORBIDDEN","description":"group.folder.read scope required"}`,
	}})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	_, err := harness.Run([]string{"drive", "channel", "list"}, newRootCommandRunner(t))
	if err == nil || !strings.Contains(err.Error(), "FORBIDDEN") {
		t.Fatalf("error = %v, want FORBIDDEN", err)
	}
}
