package cmd

import (
	"net/http"
	"testing"

	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
)

func TestJourneyDirectoryListUsers(t *testing.T) {
	h := clitest.NewHarness(t)
	installJourneyFixture(t, h.HomeDir(), "directory/list-users/config.json", h.ConfigPath())
	installJourneyFixture(t, h.HomeDir(), "directory/list-users/token.json", h.TokenPath())

	server := h.StartScriptedServer([]clitest.ResponseScript{
		{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: string(readJourneyFixture(t, "directory/list-users/api-response.json")),
		},
	})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	result, err := h.Run([]string{"directory", "list-users", "--count", "20"}, newRootCommandRunner(t))
	if err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "directory list-users failed: %v", err)
	}
	if result.Stderr != "" {
		clitest.Fatalf(t, clitest.UXContractFailure, "directory list-users stderr = %q, want empty", result.Stderr)
	}
	assertNormalizedJSON(t, result.Stdout, readJourneyFixture(t, "directory/list-users/expected-stdout.json"))

	logs := h.RequestLogs()
	if len(logs) != 1 {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "directory list-users request log count = %d, want 1", len(logs))
	}
	if logs[0].Method != http.MethodGet {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "directory list-users method = %q, want %q", logs[0].Method, http.MethodGet)
	}
	if logs[0].Path != "/users" {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "directory list-users path = %q, want %q", logs[0].Path, "/users")
	}
	if logs[0].RawQuery != "count=20" {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "directory list-users raw query = %q, want %q", logs[0].RawQuery, "count=20")
	}
	if logs[0].Headers["Authorization"] != "Bearer directory-journey-token" {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "directory list-users authorization = %q, want %q", logs[0].Headers["Authorization"], "Bearer directory-journey-token")
	}
}

func TestJourneyDirectorySearchAndUserMembership(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		responseBody string
		path         string
		rawQuery     string
	}{
		{
			name:         "search users",
			args:         []string{"directory", "search-users", "kim & lee", "--count", "2", "--domain-id", "10000001", "--order-by", "userName desc"},
			responseBody: `{"users":[{"userId":"u1","email":"u1@example.com"}]}`,
			path:         "/users/search",
			rawQuery:     "count=2&domainId=10000001&orderBy=userName+desc&query=kim+%26+lee",
		},
		{
			name:         "search groups",
			args:         []string{"directory", "search-groups", "platform", "--count", "2", "--order-by", "groupName asc"},
			responseBody: `{"groups":[{"groupId":"g1","groupName":"Platform"}]}`,
			path:         "/groups/search",
			rawQuery:     "count=2&orderBy=groupName+asc&query=platform",
		},
		{
			name:         "search org units",
			args:         []string{"directory", "search-orgunits", "engineering", "--count", "2", "--order-by", "orgUnitName desc"},
			responseBody: `{"orgUnits":[{"orgUnitId":"o1","orgUnitName":"Engineering"}]}`,
			path:         "/orgunits/search",
			rawQuery:     "count=2&orderBy=orgUnitName+desc&query=engineering",
		},
		{
			name:         "list user groups",
			args:         []string{"directory", "list-user-groups", "user-1", "--count", "2", "--membership-type", "DIRECT"},
			responseBody: `{"groups":[{"groupId":"g1","groupName":"Platform","isGroupMember":true}]}`,
			path:         "/users/user-1/groups",
			rawQuery:     "count=2&membershipType=DIRECT",
		},
		{
			name:         "list user org units",
			args:         []string{"directory", "list-user-orgunits", "user-1", "--count", "2", "--membership-type", "ALL"},
			responseBody: `{"orgUnits":[{"orgUnitId":"o1","orgUnitName":"Engineering","isPrimaryOrgUnit":true}]}`,
			path:         "/users/user-1/orgunits",
			rawQuery:     "count=2&membershipType=ALL",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := clitest.NewHarness(t)
			installJourneyFixture(t, h.HomeDir(), "directory/list-users/config.json", h.ConfigPath())
			installJourneyFixture(t, h.HomeDir(), "directory/list-users/token.json", h.TokenPath())

			server := h.StartScriptedServer([]clitest.ResponseScript{{
				StatusCode: http.StatusOK,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       test.responseBody,
			}})
			defer server.Close()
			setAPIBaseURL(t, server.URL)

			result, err := h.Run(test.args, newRootCommandRunner(t))
			if err != nil {
				clitest.Fatalf(t, clitest.SetupFailure, "%s failed: %v", test.name, err)
			}
			if result.Stderr != "" {
				clitest.Fatalf(t, clitest.UXContractFailure, "%s stderr = %q, want empty", test.name, result.Stderr)
			}
			assertNormalizedJSON(t, result.Stdout, []byte(test.responseBody))

			logs := h.RequestLogs()
			if len(logs) != 1 {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request log count = %d, want 1", test.name, len(logs))
			}
			if logs[0].Method != http.MethodGet || logs[0].Path != test.path || logs[0].RawQuery != test.rawQuery {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "%s request = %s %s?%s, want GET %s?%s", test.name, logs[0].Method, logs[0].Path, logs[0].RawQuery, test.path, test.rawQuery)
			}
			expectedAuthorization := "Bearer " + "directory-journey-token"
			if logs[0].Headers["Authorization"] != expectedAuthorization {
				clitest.Fatalf(t, clitest.RequestShapeFailure, "%s authorization = %q", test.name, logs[0].Headers["Authorization"])
			}
		})
	}
}

func TestJourneyDirectoryProfileStatusDelegatesPreserved(t *testing.T) {
	h := clitest.NewHarness(t)
	installJourneyFixture(t, h.HomeDir(), "directory/list-users/config.json", h.ConfigPath())
	installJourneyFixture(t, h.HomeDir(), "directory/list-users/token.json", h.TokenPath())

	responseBody := `{"userProfileStatuses":[{"id":"status-1","name":"휴가","delegates":[{"service":"MAIL","delegateUserId":"user-2"},{"service":"CALENDAR","delegateUserId":"user-3"}],"extension":{"preserved":true}}]}`
	server := h.StartScriptedServer([]clitest.ResponseScript{{
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       responseBody,
	}})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	result, err := h.Run([]string{"directory", "profile-status", "list", "user-1", "--count", "1"}, newRootCommandRunner(t))
	if err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "directory profile-status list failed: %v", err)
	}
	if result.Stderr != "" {
		clitest.Fatalf(t, clitest.UXContractFailure, "directory profile-status list stderr = %q, want empty", result.Stderr)
	}
	assertNormalizedJSON(t, result.Stdout, []byte(responseBody))

	logs := h.RequestLogs()
	if len(logs) != 1 {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "profile-status request log count = %d, want 1", len(logs))
	}
	if logs[0].Path != "/users/user-1/user-profile-statuses" || logs[0].RawQuery != "count=1" {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "profile-status request = %s?%s", logs[0].Path, logs[0].RawQuery)
	}
}

func TestJourneyBotSendText(t *testing.T) {
	h := clitest.NewHarness(t)
	installJourneyFixture(t, h.HomeDir(), "bot/send-text/config.json", h.ConfigPath())
	installJourneyFixture(t, h.HomeDir(), "bot/send-text/token.json", h.TokenPath())

	server := h.StartScriptedServer([]clitest.ResponseScript{
		{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: string(readJourneyFixture(t, "bot/send-text/api-response.json")),
		},
	})
	defer server.Close()
	setAPIBaseURL(t, server.URL)

	result, err := h.Run([]string{"bot", "send", "--to", "user-123", "--text", "hello journey"}, newRootCommandRunner(t))
	if err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "bot send failed: %v", err)
	}
	if result.Stderr != "" {
		clitest.Fatalf(t, clitest.UXContractFailure, "bot send stderr = %q, want empty", result.Stderr)
	}
	assertNormalizedJSON(t, result.Stdout, readJourneyFixture(t, "bot/send-text/expected-stdout.json"))

	logs := h.RequestLogs()
	if len(logs) != 1 {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "bot send request log count = %d, want 1", len(logs))
	}
	if logs[0].Method != http.MethodPost {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "bot send method = %q, want %q", logs[0].Method, http.MethodPost)
	}
	if logs[0].Path != "/bots/bot-journey/users/user-123/messages" {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "bot send path = %q, want %q", logs[0].Path, "/bots/bot-journey/users/user-123/messages")
	}
	if logs[0].Headers["Authorization"] != "Bearer bot-journey-token" {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "bot send authorization = %q, want %q", logs[0].Headers["Authorization"], "Bearer bot-journey-token")
	}
	if logs[0].Headers["Content-Type"] != "application/json" {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "bot send content-type = %q, want %q", logs[0].Headers["Content-Type"], "application/json")
	}
	assertNormalizedJSON(t, logs[0].Body, readJourneyFixture(t, "bot/send-text/expected-request-body.json"))
}
