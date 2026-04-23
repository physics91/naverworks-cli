package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/physics91/naverworks-cli/internal/config"
	clitest "github.com/physics91/naverworks-cli/internal/testkit/cli"
)

func TestLoadActiveConfig_UsesEnvOverridesForMissingExplicitProfile(t *testing.T) {
	setupTestEnv(t)
	t.Setenv("NW_PROFILE", "ci")
	t.Setenv("NW_CLIENT_ID", "env-client")
	t.Setenv("NW_CLIENT_SECRET", "env-secret")

	cfg, name, err := loadActiveConfig()
	if err != nil {
		t.Fatalf("loadActiveConfig failed: %v", err)
	}
	if name != "ci" {
		t.Fatalf("profile name = %q, want %q", name, "ci")
	}
	if cfg.ClientID != "env-client" {
		t.Fatalf("client_id = %q, want %q", cfg.ClientID, "env-client")
	}
	if cfg.ClientSecret != "env-secret" {
		t.Fatalf("client_secret = %q, want %q", cfg.ClientSecret, "env-secret")
	}
}

func TestLoadActiveConfig_MissingExplicitProfileWithoutEnvConfigStillErrors(t *testing.T) {
	setupTestEnv(t)
	t.Setenv("NW_PROFILE", "ci")

	_, name, err := loadActiveConfig()
	if err == nil {
		t.Fatal("expected missing profile error")
	}
	if name != "ci" {
		t.Fatalf("profile name = %q, want %q", name, "ci")
	}
}

func TestSelectedProfileName_TrimsCurrentProfileWhenNoOverrides(t *testing.T) {
	originalProfileName := profileName
	profileName = ""
	t.Cleanup(func() {
		profileName = originalProfileName
	})

	pc := &config.ProfileConfig{CurrentProfile: " work "}

	name, explicit := selectedProfileName(pc)
	if name != "work" {
		t.Fatalf("name = %q, want work", name)
	}
	if explicit {
		t.Fatal("expected current profile to remain implicit")
	}
}

func TestSelectedProfileName_IgnoresWhitespaceOnlyOverrides(t *testing.T) {
	originalProfileName := profileName
	profileName = "   "
	t.Cleanup(func() {
		profileName = originalProfileName
	})
	t.Setenv("NW_PROFILE", "   ")

	pc := &config.ProfileConfig{CurrentProfile: " work "}

	name, explicit := selectedProfileName(pc)
	if name != "work" {
		t.Fatalf("name = %q, want work", name)
	}
	if explicit {
		t.Fatal("expected whitespace-only overrides to stay implicit")
	}
}

func TestJourneyConfigSet_StdinLargeValue(t *testing.T) {
	h := clitest.NewHarness(t)
	largeValue := strings.Repeat("a", 70000)

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	defer func() {
		os.Stdin = oldStdin
		_ = r.Close()
	}()

	go func() {
		_, _ = io.WriteString(w, largeValue)
		_ = w.Close()
	}()
	os.Stdin = r

	result, err := h.Run([]string{"config", "set", "client_id", "--stdin"}, newRootCommandRunner(t))
	if err != nil {
		t.Fatalf("config set --stdin failed: %v", err)
	}
	if result.Stdout != "" {
		t.Fatalf("stdout = %q, want empty", result.Stdout)
	}
	if result.Stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.Stderr)
	}

	pc, err := config.LoadProfileConfig(config.DefaultPath())
	if err != nil {
		t.Fatalf("LoadProfileConfig failed: %v", err)
	}
	profile, name, err := pc.ActiveProfile("")
	if err != nil {
		t.Fatalf("ActiveProfile failed: %v", err)
	}
	if name != "default" {
		t.Fatalf("active profile = %q, want %q", name, "default")
	}
	if profile.ClientID != largeValue {
		t.Fatalf("saved client_id length = %d, want %d", len(profile.ClientID), len(largeValue))
	}
}

func TestJourneyConfigSet_StdinOversizeRejected(t *testing.T) {
	h := clitest.NewHarness(t)
	tooLargeValue := strings.Repeat("a", int(maxStdinSize)+1)

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	defer func() {
		os.Stdin = oldStdin
		_ = r.Close()
	}()

	go func() {
		_, _ = io.WriteString(w, tooLargeValue)
		_ = w.Close()
	}()
	os.Stdin = r

	_, err = h.Run([]string{"config", "set", "client_id", "--stdin"}, newRootCommandRunner(t))
	if err == nil {
		t.Fatal("expected oversized stdin error")
	}
	if !strings.Contains(err.Error(), "stdin 읽기 실패") || !strings.Contains(err.Error(), "너무 큽니다") {
		t.Fatalf("unexpected error: %v", err)
	}
}
