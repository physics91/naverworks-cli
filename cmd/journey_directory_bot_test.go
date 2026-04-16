package cmd

import (
	"net/http"
	"path/filepath"
	"testing"

	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
)

func TestJourneyDirectoryListUsers(t *testing.T) {
	h := clitest.NewHarness(t)
	installJourneyFixture(t, h.HomeDir(), "directory/list-users/config.json", configPathForHome(h.HomeDir()))
	installJourneyFixture(t, h.HomeDir(), "directory/list-users/token.json", tokenPathForHome(h.HomeDir()))

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

func TestJourneyBotSendText(t *testing.T) {
	h := clitest.NewHarness(t)
	installJourneyFixture(t, h.HomeDir(), "bot/send-text/config.json", configPathForHome(h.HomeDir()))
	installJourneyFixture(t, h.HomeDir(), "bot/send-text/token.json", tokenPathForHome(h.HomeDir()))

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

func configPathForHome(homeDir string) string {
	return filepath.Join(homeDir, ".config", "naverworks", "config.json")
}
