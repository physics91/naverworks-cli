//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/physics91/naverworks-cli/internal/fileutil"
)

// config.json holds client_secret and scim_access_token, so a world-readable
// file must be repaired before the secrets are read.
func TestLoadProfileConfig_RepairsInsecurePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	body := `{"current_profile":"default","profiles":{"default":{"client_secret":"secret-value"}}}`
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}

	var warnings strings.Builder
	original := fileutil.SecureFileWarnWriter
	fileutil.SecureFileWarnWriter = &warnings
	defer func() { fileutil.SecureFileWarnWriter = original }()

	pc, err := LoadProfileConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got := pc.Profiles["default"].ClientSecret; got != "secret-value" {
		t.Fatalf("expected secret to load, got %q", got)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected config file tightened to 0600, got %04o", perm)
	}
	if !strings.Contains(warnings.String(), "0644") {
		t.Errorf("expected warning mentioning original mode, got %q", warnings.String())
	}
}

func TestLoad_RepairsInsecurePermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"client_secret":"secret-value"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}

	var warnings strings.Builder
	original := fileutil.SecureFileWarnWriter
	fileutil.SecureFileWarnWriter = &warnings
	defer func() { fileutil.SecureFileWarnWriter = original }()

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.ClientSecret != "secret-value" {
		t.Fatalf("expected secret to load, got %q", cfg.ClientSecret)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected config file tightened to 0600, got %04o", perm)
	}
}

func TestLoadProfileConfig_MissingFileStillReturnsDefaults(t *testing.T) {
	pc, err := LoadProfileConfig(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatalf("expected missing config to be tolerated, got %v", err)
	}
	if pc.CurrentProfile != "default" || pc.Profiles["default"] == nil {
		t.Fatalf("expected default profile scaffold, got %+v", pc)
	}
}
