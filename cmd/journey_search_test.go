package cmd

import (
	"net/http"
	"strings"
	"testing"

	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
)

func TestJourneyDomainSearchCommands(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		defaultUser  string
		tokenFixture string
		expectedAuth string
		responseBody string
		path         string
		rawQuery     string
	}{
		{
			name:         "board posts",
			args:         []string{"board", "search-posts", "weekly & notes", "--board-ids", "1,2", "--has-attachment", "--writer-id", "user/one", "--start-time", "2026-01-01", "--end-time", "2026-04-28", "--count", "2"},
			responseBody: `{"posts":[{"postId":1,"title":"Weekly Notes","boardName":"Notice","snippet":"weekly"}]}`,
			path:         "/boards/posts/search",
			rawQuery:     "boardIds=1%2C2&count=2&endTime=2026-04-28&hasAttachment=true&query=weekly+%26+notes&startTime=2026-01-01&writerId=user%2Fone",
		},
		{
			name:         "note posts",
			args:         []string{"note", "search-posts", "group/one", "weekly & notes", "--count", "2"},
			responseBody: `{"posts":[{"postId":1,"title":"Weekly Notes","snippet":"weekly"}]}`,
			path:         "/groups/group/one/note/posts/search",
			rawQuery:     "count=2&query=weekly+%26+notes",
		},
		{
			name:         "calendar events",
			args:         []string{"calendar", "search-events", "weekly & meeting", "--user-id", "externalKey:user/one", "--query-filters", "summary,attendee", "--start-time", "2026-06-03T10:00:00+09:00", "--end-time", "2026-06-03T11:00:00+09:00", "--count", "2"},
			responseBody: `{"events":[{"eventComponents":[{"eventId":"event-1","summary":"Weekly Meeting"}]}]}`,
			path:         "/users/externalKey:user/one/calendars/events/search",
			rawQuery:     "count=2&endTime=2026-06-03T11%3A00%3A00%2B09%3A00&query=weekly+%26+meeting&queryFilters=summary%2Cattendee&startTime=2026-06-03T10%3A00%3A00%2B09%3A00",
		},
		{
			name: "tasks",
			args: []string{
				"task", "search", "weekly & meeting", "--user-id", "externalKey:user/one",
				"--assignor-id", "assignor/one", "--assignee-id", "assignee/one",
				"--start-time", "2026-06-03T10:00:00+09:00", "--end-time", "2026-06-03T11:00:00+09:00",
				"--status", "TODO", "--has-due-date=false", "--has-attachment",
				"--order-by", "createdTime desc", "--count", "2",
			},
			tokenFixture: "task/search/token.json",
			expectedAuth: "task-journey-token",
			responseBody: `{"tasks":[{"taskId":"task-1","title":"Weekly Meeting","status":"TODO","dueDate":null}]}`,
			path:         "/users/externalKey:user/one/tasks/search",
			rawQuery:     "assigneeId=assignee%2Fone&assignorId=assignor%2Fone&count=2&endTime=2026-06-03T11%3A00%3A00%2B09%3A00&hasAttachment=true&hasDueDate=false&orderBy=createdTime+desc&query=weekly+%26+meeting&startTime=2026-06-03T10%3A00%3A00%2B09%3A00&status=TODO",
		},
		{
			name:         "approval admin documents",
			args:         []string{"approval", "list-all", "--from", "2026-01-01", "--until", "2026-01-20", "--document-form-id", "form/one", "--type", "approved", "--order-by", "createdTime desc", "--count", "10"},
			responseBody: `{"documents":[{"approvalDocumentId":10001,"title":"Annual leave","status":"APPROVED"}]}`,
			path:         "/business-support/approval/documents",
			rawQuery:     "count=10&documentFormId=form%2Fone&fromDateTime=2026-01-01&orderBy=createdTime+desc&type=approved&untilDateTime=2026-01-20",
		},
		{
			name:         "contacts with configured user",
			args:         []string{"contact", "search", "kim & lee", "--query-filters", "contactName,emails", "--order-by", "name desc", "--count", "2"},
			defaultUser:  "default-user",
			responseBody: `{"contacts":[{"contactId":"contact-1","contactName":{"lastName":"Kim"},"emails":[{"primary":true,"email":"kim@example.com"}]}]}`,
			path:         "/users/default-user/contacts/search",
			rawQuery:     "count=2&orderBy=name+desc&query=kim+%26+lee&queryFilters=contactName%2Cemails",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			harness := clitest.NewHarness(t)
			if test.defaultUser != "" {
				t.Setenv("NW_DEFAULT_CALENDAR_USER_ID", test.defaultUser)
			}
			installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
			tokenFixture := test.tokenFixture
			if tokenFixture == "" {
				tokenFixture = "directory/list-users/token.json"
			}
			installJourneyFixture(t, harness.HomeDir(), tokenFixture, harness.TokenPath())

			server := harness.StartScriptedServer([]clitest.ResponseScript{{
				StatusCode: http.StatusOK,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       test.responseBody,
			}})
			defer server.Close()
			setAPIBaseURL(t, server.URL)

			result, err := harness.Run(test.args, newRootCommandRunner(t))
			if err != nil {
				clitest.Fatalf(t, clitest.SetupFailure, "%s failed: %v", test.name, err)
			}
			if result.Stderr != "" {
				clitest.Fatalf(t, clitest.UXContractFailure, "%s stderr = %q, want empty", test.name, result.Stderr)
			}
			assertNormalizedJSON(t, result.Stdout, []byte(test.responseBody))

			logs := harness.RequestLogs()
			if len(logs) != 1 {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request log count = %d, want 1", test.name, len(logs))
			}
			if logs[0].Method != http.MethodGet || logs[0].Path != test.path || logs[0].RawQuery != test.rawQuery {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request = %s %s?%s, want GET %s?%s", test.name, logs[0].Method, logs[0].Path, logs[0].RawQuery, test.path, test.rawQuery)
			}
			expectedToken := test.expectedAuth
			if expectedToken == "" {
				expectedToken = "directory-journey-token"
			}
			expectedAuthorization := "Bearer " + expectedToken
			if logs[0].Headers["Authorization"] != expectedAuthorization {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "%s authorization = %q", test.name, logs[0].Headers["Authorization"])
			}
		})
	}
}

func TestJourneySearchCommandsRejectInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "board query required", args: []string{"board", "search-posts"}, want: "accepts 1 arg"},
		{name: "note query required", args: []string{"note", "search-posts", "group-1"}, want: "accepts 2 arg"},
		{name: "contact query required", args: []string{"contact", "search"}, want: "accepts 1 arg"},
		{name: "calendar filter requires query", args: []string{"calendar", "search-events", "--query-filters", "summary"}, want: "query를 지정하세요"},
		{name: "board blank query rejected", args: []string{"board", "search-posts", "   "}, want: "query는 비어 있을 수 없습니다"},
		{name: "note blank query rejected", args: []string{"note", "search-posts", "group-1", ""}, want: "query는 비어 있을 수 없습니다"},
		{name: "contact blank query rejected", args: []string{"contact", "search", "\t"}, want: "query는 비어 있을 수 없습니다"},
		{name: "calendar blank query rejected", args: []string{"calendar", "search-events", " "}, want: "query는 비어 있을 수 없습니다"},
		{name: "calendar blank query rejects filters", args: []string{"calendar", "search-events", " ", "--query-filters", "summary"}, want: "query는 비어 있을 수 없습니다"},
		{name: "task search condition required", args: []string{"task", "search"}, want: "중 하나가 필요합니다"},
		{name: "task status alone is insufficient", args: []string{"task", "search", "--status", "TODO"}, want: "중 하나가 필요합니다"},
		{name: "task blank query rejected", args: []string{"task", "search", " "}, want: "query는 비어 있을 수 없습니다"},
		{name: "approval from required", args: []string{"approval", "list-all", "--until", "2026-01-20"}, want: "--from과 --until은 필수입니다"},
		{name: "approval until required", args: []string{"approval", "list-all", "--from", "2026-01-01"}, want: "--from과 --until은 필수입니다"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setupTestEnv(t)
			_, err := runCLI(t, test.args...)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %q, want substring %q", err, test.want)
			}
		})
	}
}

func TestJourneyTaskSearchRejectsServiceAccount(t *testing.T) {
	homeDir := setupTestEnv(t)
	writeTestConfig(t, homeDir)

	_, err := runCLI(t, "task", "search", "weekly", "--user-id", "user-1")
	if err == nil || !strings.Contains(err.Error(), "구성원 계정 Access Token만 지원합니다") {
		t.Fatalf("error = %v, want service-account rejection", err)
	}
}

func TestJourneyDomainSearchTableOutput(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		tokenFixture string
		responseBody string
		want         []string
	}{
		{
			name:         "calendar nested event fields",
			args:         []string{"calendar", "search-events", "weekly", "--user-id", "user-1", "--output", "table"},
			responseBody: `{"events":[{"eventComponents":[{"eventId":"event-1","summary":"Weekly Meeting","start":{"dateTime":"2026-06-03T10:00:00","timeZone":"Asia/Seoul"},"end":{"dateTime":"2026-06-03T11:00:00","timeZone":"Asia/Seoul"}}]}]}`,
			want:         []string{"eventId", "summary", "start", "end", "event-1", "Weekly Meeting", "2026-06-03T10:00:00", "Asia/Seoul"},
		},
		{
			name:         "contact nested name and email fields",
			args:         []string{"contact", "search", "홍길동", "--user-id", "user-1", "--output", "table"},
			responseBody: `{"contacts":[{"contactId":"contact-1","contactName":{"lastName":"홍","firstName":"길동"},"emails":[{"primary":true,"email":"hong@example.com"}]}]}`,
			want:         []string{"contactId", "lastName", "firstName", "email", "contact-1", "홍", "길동", "hong@example.com"},
		},
		{
			name:         "task fields",
			args:         []string{"task", "search", "weekly", "--user-id", "user-1", "--output", "table"},
			tokenFixture: "task/search/token.json",
			responseBody: `{"tasks":[{"taskId":"task-1","title":"Weekly Meeting","status":"TODO","dueDate":"2026-06-03T11:00:00+09:00"}]}`,
			want:         []string{"taskId", "title", "status", "dueDate", "task-1", "Weekly Meeting", "TODO", "2026-06-03T11:00:00+09:00"},
		},
		{
			name:         "approval document fields",
			args:         []string{"approval", "list-all", "--from", "2026-01-01", "--until", "2026-01-20", "--output", "table"},
			responseBody: `{"documents":[{"approvalDocumentId":10001,"title":"Annual leave","status":"APPROVED"}]}`,
			want:         []string{"approvalDocumentId", "title", "10001", "Annual leave"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			harness := clitest.NewHarness(t)
			installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
			tokenFixture := test.tokenFixture
			if tokenFixture == "" {
				tokenFixture = "directory/list-users/token.json"
			}
			installJourneyFixture(t, harness.HomeDir(), tokenFixture, harness.TokenPath())
			server := harness.StartScriptedServer([]clitest.ResponseScript{{
				StatusCode: http.StatusOK,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       test.responseBody,
			}})
			defer server.Close()
			setAPIBaseURL(t, server.URL)

			result, err := harness.Run(test.args, newRootCommandRunner(t))
			if err != nil {
				clitest.Fatalf(t, clitest.SetupFailure, "%s failed: %v", test.name, err)
			}
			for _, want := range test.want {
				if !strings.Contains(result.Stdout, want) {
					clitest.Fatalf(t, clitest.UXContractFailure, "%s stdout = %q, want substring %q", test.name, result.Stdout, want)
				}
			}
		})
	}
}

func TestJourneyBoardSearchPropagatesAPIError(t *testing.T) {
	harness := clitest.NewHarness(t)
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/token.json", harness.TokenPath())
	server := harness.StartScriptedServer([]clitest.ResponseScript{{
		StatusCode: http.StatusForbidden,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"code":"FORBIDDEN","description":"scope required"}`,
	}})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	_, err := harness.Run([]string{"board", "search-posts", "weekly"}, newRootCommandRunner(t))
	if err == nil || !strings.Contains(err.Error(), "FORBIDDEN") {
		t.Fatalf("error = %v, want FORBIDDEN API error", err)
	}
}

func TestJourneyApprovalAdminListPropagatesAPIError(t *testing.T) {
	harness := clitest.NewHarness(t)
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/token.json", harness.TokenPath())
	server := harness.StartScriptedServer([]clitest.ResponseScript{{
		StatusCode: http.StatusForbidden,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"code":"FORBIDDEN","description":"administrator role required"}`,
	}})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	_, err := harness.Run([]string{"approval", "list-all", "--from", "2026-01-01", "--until", "2026-01-20"}, newRootCommandRunner(t))
	if err == nil || !strings.Contains(err.Error(), "FORBIDDEN") {
		t.Fatalf("error = %v, want FORBIDDEN API error", err)
	}
}
