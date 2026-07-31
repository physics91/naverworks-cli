//go:build !windows

package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/physics91/naverworks-cli/internal/config"
)

// The documented way to set a secret is `pass show ... | naverworks config set
// client_secret --stdin`. When the producing command fails it writes nothing,
// and storing that empty value would wipe a working credential.
func TestConfigSet_Stdin_RejectsEmptyInput(t *testing.T) {
	tests := []struct {
		name  string
		stdin string
	}{
		{name: "no output at all", stdin: ""},
		{name: "trailing newline only", stdin: "\n"},
		{name: "crlf only", stdin: "\r\n"},
		{name: "whitespace only", stdin: "   \t \n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setupTestEnv(t)

			// Seed an existing secret that must survive the failed attempt.
			cfgPath := testConfigPath(t)
			pc := config.NewProfileConfig()
			pc.EnsureProfile("default").ClientSecret = "EXISTING_SECRET"
			if err := pc.Save(cfgPath); err != nil {
				t.Fatalf("seed config failed: %v", err)
			}

			withStdin(t, tc.stdin)

			_, err := runCLI(t, "config", "set", "client_secret", "--stdin")
			if err == nil {
				t.Fatal("expected empty stdin to be rejected")
			}
			if !strings.Contains(err.Error(), "빈 값") {
				t.Errorf("expected an empty-value error, got: %v", err)
			}

			reloaded, err := config.LoadProfileConfig(cfgPath)
			if err != nil {
				t.Fatalf("reload config failed: %v", err)
			}
			if got := reloaded.Profiles["default"].ClientSecret; got != "EXISTING_SECRET" {
				t.Errorf("existing secret was overwritten: %q", got)
			}
		})
	}
}

func TestConfigSet_Stdin_AcceptsPipedValue(t *testing.T) {
	setupTestEnv(t)
	cfgPath := testConfigPath(t)

	withStdin(t, "PIPED_SECRET\n")

	if _, err := runCLI(t, "config", "set", "client_secret", "--stdin"); err != nil {
		t.Fatalf("piped value should be accepted: %v", err)
	}

	pc, err := config.LoadProfileConfig(cfgPath)
	if err != nil {
		t.Fatalf("reload config failed: %v", err)
	}
	if got := pc.Profiles["default"].ClientSecret; got != "PIPED_SECRET" {
		t.Errorf("stored secret = %q, want %q", got, "PIPED_SECRET")
	}
}

// Clearing a value stays possible through an explicit empty argument.
func TestConfigSet_ExplicitEmptyArgumentStillClears(t *testing.T) {
	setupTestEnv(t)
	cfgPath := testConfigPath(t)

	pc := config.NewProfileConfig()
	pc.EnsureProfile("default").ClientSecret = "EXISTING_SECRET"
	if err := pc.Save(cfgPath); err != nil {
		t.Fatalf("seed config failed: %v", err)
	}

	if _, err := runCLI(t, "config", "set", "client_secret", ""); err != nil {
		t.Fatalf("explicit empty argument should be accepted: %v", err)
	}

	reloaded, err := config.LoadProfileConfig(cfgPath)
	if err != nil {
		t.Fatalf("reload config failed: %v", err)
	}
	if got := reloaded.Profiles["default"].ClientSecret; got != "" {
		t.Errorf("secret should have been cleared, got %q", got)
	}
}

func withStdin(t *testing.T, content string) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	if _, err := w.WriteString(content); err != nil {
		t.Fatalf("write stdin failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close stdin writer failed: %v", err)
	}

	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = orig
		_ = r.Close()
	})
}
