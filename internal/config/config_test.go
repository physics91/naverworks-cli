package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ClientID != "" {
		t.Errorf("expected empty client_id, got %q", cfg.ClientID)
	}
}

func TestSetAndGet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg, _ := Load(path)
	if err := cfg.Set("client_id", "test-id"); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	cfg2, _ := Load(path)
	if cfg2.ClientID != "test-id" {
		t.Errorf("expected test-id, got %q", cfg2.ClientID)
	}
}

func TestEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg, _ := Load(path)
	cfg.Set("client_id", "from-file")

	os.Setenv("NW_CLIENT_ID", "from-env")
	defer os.Unsetenv("NW_CLIENT_ID")

	cfg.ApplyEnvOverrides()
	if cfg.ClientID != "from-env" {
		t.Errorf("expected from-env, got %q", cfg.ClientID)
	}
}

func TestGetMasked(t *testing.T) {
	cfg := &Config{}
	cfg.Set("client_secret", "super-secret")

	val := cfg.GetMasked("client_secret")
	if val != "****" {
		t.Errorf("expected ****, got %q", val)
	}
}

func TestSetInvalidKey(t *testing.T) {
	cfg := &Config{}
	err := cfg.Set("invalid_key", "value")
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestIsValidKey(t *testing.T) {
	if !IsValidKey("client_id") {
		t.Error("expected client_id to be valid")
	}
	if IsValidKey("invalid") {
		t.Error("expected invalid to be invalid")
	}
}
