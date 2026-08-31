package cmd

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
)

func TestJourneyModernizedAPIAllPaginationContracts(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		path        string
		responseKey string
	}{
		{
			name:        "directory search users",
			args:        []string{"directory", "search-users", "integration", "--count", "2", "--all"},
			path:        "/users/search",
			responseKey: "users",
		},
		{
			name:        "board search posts",
			args:        []string{"board", "search-posts", "integration", "--count", "2", "--all"},
			path:        "/boards/posts/search",
			responseKey: "posts",
		},
		{
			name:        "task search",
			args:        []string{"task", "search", "integration", "--user-id", "user-1", "--count", "2", "--all"},
			path:        "/users/user-1/tasks/search",
			responseKey: "tasks",
		},
		{
			name:        "approval admin documents",
			args:        []string{"approval", "list-all", "--from", "2026-01-01", "--until", "2026-01-20", "--count", "2", "--all"},
			path:        "/business-support/approval/documents",
			responseKey: "documents",
		},
		{
			name:        "channel folder files",
			args:        []string{"drive", "channel", "files", "cf1", "--count", "2", "--all"},
			path:        "/channel-folders/cf1/files",
			responseKey: "files",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			harness := clitest.NewHarness(t)
			installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
			installJourneyFixture(t, harness.HomeDir(), "modernized-api/token.json", harness.TokenPath())

			server := harness.StartScriptedServer([]clitest.ResponseScript{
				{
					StatusCode: http.StatusOK,
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body:       `{"` + test.responseKey + `":[{"integrationId":"page-1"}],"responseMetaData":{"nextCursor":"next"}}`,
				},
				{
					StatusCode: http.StatusOK,
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body:       `{"` + test.responseKey + `":[{"integrationId":"page-2"}]}`,
				},
			})
			defer server.Close()
			setAPIBaseURL(t, server.URL)

			result, err := harness.Run(test.args, newRootCommandRunner(t))
			if err != nil {
				clitest.Fatalf(t, clitest.SetupFailure, "%s failed: %v", test.name, err)
			}
			if result.Stderr != "" {
				clitest.Fatalf(t, clitest.UXContractFailure, "%s stderr = %q, want empty", test.name, result.Stderr)
			}

			var payload map[string]json.RawMessage
			if err := json.Unmarshal([]byte(result.Stdout), &payload); err != nil {
				clitest.Fatalf(t, clitest.ResponseHandlingFailure, "%s stdout is not valid JSON: %v\noutput: %q", test.name, err, result.Stdout)
			}
			if len(payload) != 1 {
				clitest.Fatalf(t, clitest.ResponseHandlingFailure, "%s response keys = %v, want only %q", test.name, payload, test.responseKey)
			}
			var items []map[string]any
			if err := json.Unmarshal(payload[test.responseKey], &items); err != nil {
				clitest.Fatalf(t, clitest.ResponseHandlingFailure, "%s response key %q is not an item array: %v", test.name, test.responseKey, err)
			}
			if len(items) != 2 {
				clitest.Fatalf(t, clitest.ResponseHandlingFailure, "%s merged item count = %d, want 2", test.name, len(items))
			}
			for index, want := range []string{"page-1", "page-2"} {
				if got := items[index]["integrationId"]; got != want {
					clitest.Fatalf(t, clitest.ResponseHandlingFailure, "%s item %d integrationId = %v, want %q", test.name, index+1, got, want)
				}
			}

			logs := harness.RequestLogs()
			if len(logs) != 2 {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request count = %d, want one request per page", test.name, len(logs))
			}
			for index, log := range logs {
				if log.Method != http.MethodGet || log.Path != test.path {
					clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request %d = %s %s", test.name, index+1, log.Method, log.Path)
				}
				if got := log.Headers["Authorization"]; got != "Bearer modernized-api-journey-token" {
					clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request %d authorization = %q", test.name, index+1, got)
				}
				query, err := url.ParseQuery(log.RawQuery)
				if err != nil {
					clitest.Fatalf(t, clitest.RequestShapeFailure, "%s query %q: %v", test.name, log.RawQuery, err)
				}
				if query.Get("count") != "2" {
					clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request %d count = %q", test.name, index+1, query.Get("count"))
				}
				wantCursor := ""
				if index == 1 {
					wantCursor = "next"
				}
				if query.Get("cursor") != wantCursor {
					clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request %d cursor = %q, want %q", test.name, index+1, query.Get("cursor"), wantCursor)
				}
			}
		})
	}
}

func TestJourneyModernizedAPIAllPaginationRejectsRepeatedCursor(t *testing.T) {
	harness := clitest.NewHarness(t)
	installJourneyFixture(t, harness.HomeDir(), "directory/list-users/config.json", harness.ConfigPath())
	installJourneyFixture(t, harness.HomeDir(), "modernized-api/token.json", harness.TokenPath())
	server := harness.StartScriptedServer([]clitest.ResponseScript{
		{StatusCode: http.StatusOK, Body: `{"files":[{"integrationId":"page-1"}],"responseMetaData":{"nextCursor":"repeat"}}`},
		{StatusCode: http.StatusOK, Body: `{"files":[{"integrationId":"page-2"}],"responseMetaData":{"nextCursor":"repeat"}}`},
	})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	_, err := harness.Run([]string{"drive", "channel", "files", "cf1", "--count", "2", "--all"}, newRootCommandRunner(t))
	if err == nil || !strings.Contains(err.Error(), "next cursor 순환 감지") {
		clitest.Fatalf(t, clitest.ResponseHandlingFailure, "error = %v, want repeated cursor rejection", err)
	}
	if logs := harness.RequestLogs(); len(logs) != 2 {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "request count = %d, want bounded two requests", len(logs))
	} else {
		for index, log := range logs {
			if got := log.Headers["Authorization"]; got != "Bearer modernized-api-journey-token" {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "request %d authorization = %q", index+1, got)
			}
		}
	}
}
