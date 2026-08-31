package cmd

import (
	"net/http"
	"strings"
	"testing"

	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
)

func TestJourneyDriveSearch(t *testing.T) {
	harness := clitest.NewHarness(t)
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
	installJourneyFixture(t, harness.HomeDir(), "task/search/token.json", harness.TokenPath())
	responseBody := `{"files":[{"fileId":"file-1","fileName":"Weekly Notes.docx","fileType":"DOC","modifiedTime":"2026-06-03T10:30:00+09:00"}]}`
	server := harness.StartScriptedServer([]clitest.ResponseScript{{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       responseBody,
	}})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	result, err := harness.Run([]string{
		"drive", "search", "weekly & notes", "--user-id", "externalKey:user/one",
		"--query-filters", "fileName,content", "--parent-file-id", "folder/one",
		"--file-types", "DOC,FOLDER", "--start-time", "2026-06-03T10:00:00+09:00",
		"--end-time", "2026-06-03T11:00:00+09:00", "--drive-type-filters", "MY_DRIVE,CHANNEL_FOLDER",
		"--order-by", "modifiedTime desc", "--count", "2",
	}, newRootCommandRunner(t))
	if err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "drive search failed: %v", err)
	}
	if result.Stderr != "" {
		clitest.Fatalf(t, clitest.UXContractFailure, "drive search stderr = %q", result.Stderr)
	}
	assertNormalizedJSON(t, result.Stdout, []byte(responseBody))

	logs := harness.RequestLogs()
	if len(logs) != 1 {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "request log count = %d, want 1", len(logs))
	}
	wantQuery := "count=2&driveTypeFilters=MY_DRIVE%2CCHANNEL_FOLDER&endTime=2026-06-03T11%3A00%3A00%2B09%3A00&fileTypes=DOC%2CFOLDER&orderBy=modifiedTime+desc&parentFileId=folder%2Fone&query=weekly+%26+notes&queryFilters=fileName%2Ccontent&startTime=2026-06-03T10%3A00%3A00%2B09%3A00"
	if logs[0].Method != http.MethodGet || logs[0].Path != "/users/externalKey:user/one/drive/search" || logs[0].RawQuery != wantQuery {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "request = %s %s?%s", logs[0].Method, logs[0].Path, logs[0].RawQuery)
	}
	expectedAuthorization := "Bearer " + "task-" + "journey-token"
	if logs[0].Headers["Authorization"] != expectedAuthorization {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "authorization = %q", logs[0].Headers["Authorization"])
	}
}

func TestJourneyDriveSearchTableOutput(t *testing.T) {
	harness := clitest.NewHarness(t)
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
	installJourneyFixture(t, harness.HomeDir(), "task/search/token.json", harness.TokenPath())
	server := harness.StartScriptedServer([]clitest.ResponseScript{{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"files":[{"fileId":"file-1","fileName":"Weekly Notes.docx","fileType":"DOC","modifiedTime":"2026-06-03T10:30:00+09:00"}]}`,
	}})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	result, err := harness.Run([]string{"drive", "search", "weekly", "--user-id", "user-1", "--output", "table"}, newRootCommandRunner(t))
	if err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "drive search table failed: %v", err)
	}
	for _, want := range []string{"fileId", "fileName", "fileType", "modifiedTime", "file-1", "Weekly Notes.docx", "DOC"} {
		if !strings.Contains(result.Stdout, want) {
			clitest.Fatalf(t, clitest.UXContractFailure, "stdout = %q, want %q", result.Stdout, want)
		}
	}
}

func TestJourneyMonitoringDownloadMessagesChannelID(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		wantQuery string
	}{
		{
			name:      "specified",
			channelID: "channel/one",
			wantQuery: "channelId=channel%2Fone&endTime=2026-06-03T11%3A00%3A00%2B09%3A00&startTime=2026-06-03T10%3A00%3A00%2B09%3A00",
		},
		{
			name:      "blank omitted",
			channelID: " ",
			wantQuery: "endTime=2026-06-03T11%3A00%3A00%2B09%3A00&startTime=2026-06-03T10%3A00%3A00%2B09%3A00",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			harness := clitest.NewHarness(t)
			installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
			installJourneyFixture(t, harness.HomeDir(), "directory/list-users/token.json", harness.TokenPath())
			server := harness.StartScriptedServer([]clitest.ResponseScript{{
				StatusCode: http.StatusFound,
				Headers:    map[string]string{"Location": "https://download.example.com/messages.csv"},
			}})
			defer server.Close()
			setAPIBaseURL(t, server.URL)

			result, err := harness.Run([]string{
				"monitoring", "download-messages", "--start-time", "2026-06-03T10:00:00+09:00",
				"--end-time", "2026-06-03T11:00:00+09:00", "--channel-id", test.channelID,
			}, newRootCommandRunner(t))
			if err != nil {
				clitest.Fatalf(t, clitest.SetupFailure, "monitoring download failed: %v", err)
			}
			assertNormalizedJSON(t, result.Stdout, []byte(`{"download_url":"https://download.example.com/messages.csv"}`))
			logs := harness.RequestLogs()
			if len(logs) != 1 || logs[0].Path != "/monitoring/message-contents/download" || logs[0].RawQuery != test.wantQuery {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "request logs = %#v", logs)
			}
			expectedAuthorization := "Bearer " + "directory-" + "journey-token"
			if logs[0].Headers["Authorization"] != expectedAuthorization {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "authorization = %q", logs[0].Headers["Authorization"])
			}
		})
	}
}

func TestJourneyDriveMonitoringRejectInvalidInputBeforeAuth(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "blank drive query", args: []string{"drive", "search", "  "}, want: "query는 비어 있을 수 없습니다"},
		{name: "missing monitoring start", args: []string{"monitoring", "download-messages", "--end-time", "2026-06-03T11:00:00+09:00"}, want: "--start-time과 --end-time은 필수입니다"},
		{name: "missing monitoring end", args: []string{"monitoring", "download-messages", "--start-time", "2026-06-03T10:00:00+09:00"}, want: "--start-time과 --end-time은 필수입니다"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setupTestEnv(t)
			_, err := runCLI(t, test.args...)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestJourneyDriveSearchPropagatesAPIError(t *testing.T) {
	harness := clitest.NewHarness(t)
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
	installJourneyFixture(t, harness.HomeDir(), "task/search/token.json", harness.TokenPath())
	server := harness.StartScriptedServer([]clitest.ResponseScript{{
		StatusCode: http.StatusForbidden,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"code":"FORBIDDEN","description":"file.read scope required"}`,
	}})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	_, err := harness.Run([]string{"drive", "search", "weekly", "--user-id", "user-1"}, newRootCommandRunner(t))
	if err == nil || !strings.Contains(err.Error(), "FORBIDDEN") {
		t.Fatalf("error = %v, want FORBIDDEN API error", err)
	}
}

func TestJourneyDriveSearchRejectsServiceAccount(t *testing.T) {
	homeDir := setupTestEnv(t)
	writeTestConfig(t, homeDir)

	_, err := runCLI(t, "drive", "search", "weekly", "--user-id", "user-1")
	if err == nil || !strings.Contains(err.Error(), "구성원 계정 Access Token만 지원합니다") {
		t.Fatalf("error = %v, want service-account rejection", err)
	}
}
