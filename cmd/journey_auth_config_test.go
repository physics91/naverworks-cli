package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/physics91/naverworks-cli/internal/config"
	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
)

func TestJourneyAuthStatus(t *testing.T) {
	h := clitest.NewHarness(t)
	installJourneyFixture(t, h.HomeDir(), "auth/status/config.json", config.DefaultPath())
	installJourneyFixture(t, h.HomeDir(), "auth/status/token.json", tokenPathForHome(h.HomeDir()))

	result, err := h.Run([]string{"auth", "status"}, newRootCommandRunner(t))
	if err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "auth status run failed: %v", err)
	}
	if result.Stderr != "" {
		clitest.Fatalf(t, clitest.UXContractFailure, "auth status stderr = %q, want empty", result.Stderr)
	}
	if got := len(h.RequestLogs()); got != 0 {
		clitest.Fatalf(t, clitest.RequestShapeFailure, "auth status request log count = %d, want 0", got)
	}

	expected := readJourneyFixture(t, "auth/status/expected-stdout.json")
	assertNormalizedJSON(t, result.Stdout, expected)
}

func TestJourneyConfigLifecycle(t *testing.T) {
	h := clitest.NewHarness(t)
	installJourneyFixture(t, h.HomeDir(), "config/set-get-list/initial-config.json", config.DefaultPath())

	result, err := h.Run([]string{"config", "set", "bot_id", "bot-updated"}, newRootCommandRunner(t))
	if err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "config set failed: %v", err)
	}
	if result.Stdout != "" {
		clitest.Fatalf(t, clitest.UXContractFailure, "config set stdout = %q, want empty", result.Stdout)
	}
	if result.Stderr != "" {
		clitest.Fatalf(t, clitest.UXContractFailure, "config set stderr = %q, want empty", result.Stderr)
	}

	getResult, err := h.Run([]string{"config", "get", "bot_id"}, newRootCommandRunner(t))
	if err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "config get failed: %v", err)
	}
	if strings.TrimSpace(getResult.Stdout) != "bot-updated" {
		clitest.Fatalf(t, clitest.UXContractFailure, "config get stdout = %q, want %q", strings.TrimSpace(getResult.Stdout), "bot-updated")
	}

	listResult, err := h.Run([]string{"config", "list"}, newRootCommandRunner(t))
	if err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "config list failed: %v", err)
	}
	if listResult.Stderr != "" {
		clitest.Fatalf(t, clitest.UXContractFailure, "config list stderr = %q, want empty", listResult.Stderr)
	}
	expected := readJourneyFixture(t, "config/set-get-list/expected-list.json")
	assertNormalizedJSON(t, listResult.Stdout, expected)

	cfgPath := filepath.Join(h.HomeDir(), ".config", "naverworks", "config.json")
	pc, err := config.LoadProfileConfig(cfgPath)
	if err != nil {
		clitest.Fatalf(t, clitest.SideEffectFailure, "load saved profile config failed: %v", err)
	}
	profile, name, err := pc.ActiveProfile("")
	if err != nil {
		clitest.Fatalf(t, clitest.SideEffectFailure, "active profile lookup failed: %v", err)
	}
	if name != "default" {
		clitest.Fatalf(t, clitest.SideEffectFailure, "active profile = %q, want %q", name, "default")
	}
	if profile.BotID != "bot-updated" {
		clitest.Fatalf(t, clitest.SideEffectFailure, "saved bot_id = %q, want %q", profile.BotID, "bot-updated")
	}
}

func installJourneyFixture(t *testing.T, homeDir, fixturePath, targetPath string) {
	t.Helper()

	data := readJourneyFixture(t, fixturePath)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0700); err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "mkdir %s failed: %v", filepath.Dir(targetPath), err)
	}
	if err := os.WriteFile(targetPath, data, 0600); err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "write fixture %s failed: %v", targetPath, err)
	}
}

func readJourneyFixture(t *testing.T, relativePath string) []byte {
	t.Helper()

	candidates := []string{
		filepath.Join("testdata", "journey", relativePath),
		filepath.Join("..", "testdata", "journey", relativePath),
	}

	var lastErr error
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			return data
		}
		lastErr = err
	}

	clitest.Fatalf(t, clitest.SetupFailure, "read fixture %s failed: %v", relativePath, lastErr)
	return nil
}

func assertNormalizedJSON(t *testing.T, gotText string, wantJSON []byte) {
	t.Helper()

	var gotValue any
	if err := json.Unmarshal([]byte(strings.TrimSpace(gotText)), &gotValue); err != nil {
		clitest.Fatalf(t, clitest.UXContractFailure, "actual output is not valid JSON: %v\noutput: %q", err, gotText)
	}

	var wantValue any
	if err := json.Unmarshal(wantJSON, &wantValue); err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "expected fixture is not valid JSON: %v\nfixture: %q", err, string(wantJSON))
	}

	gotNormalized, err := json.MarshalIndent(gotValue, "", "  ")
	if err != nil {
		clitest.Fatalf(t, clitest.UXContractFailure, "normalize actual JSON failed: %v", err)
	}
	wantNormalized, err := json.MarshalIndent(wantValue, "", "  ")
	if err != nil {
		clitest.Fatalf(t, clitest.SetupFailure, "normalize expected JSON failed: %v", err)
	}

	if string(gotNormalized) != string(wantNormalized) {
		clitest.Fatalf(t, clitest.UXContractFailure, "json mismatch\nactual:\n%s\nexpected:\n%s", string(gotNormalized), string(wantNormalized))
	}
}

func tokenPathForHome(homeDir string) string {
	return filepath.Join(homeDir, ".config", "naverworks", "token.json")
}
