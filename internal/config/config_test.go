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

func TestProfileConfig_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// v0.1.0 CLI에서는 SetCurrentProfile을 직접 호출하지 않음.
	// 이 테스트는 저장 구조의 정합성을 검증.
	cfg := NewProfileConfig()
	cfg.SetCurrentProfile("work")
	profile := cfg.EnsureProfile("work")
	profile.ClientID = "work-id"
	profile.BotID = "work-bot"

	if err := cfg.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadProfileConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.CurrentProfile != "work" {
		t.Errorf("expected current_profile=work, got %q", loaded.CurrentProfile)
	}
	p, _, err := loaded.ActiveProfile("")
	if err != nil {
		t.Fatalf("active profile: %v", err)
	}
	if p.ClientID != "work-id" {
		t.Errorf("expected client_id=work-id, got %q", p.ClientID)
	}
}

func TestProfileConfig_FlagOverride(t *testing.T) {
	cfg := NewProfileConfig()
	cfg.SetCurrentProfile("default")
	cfg.EnsureProfile("default").ClientID = "def-id"
	cfg.EnsureProfile("staging").ClientID = "stg-id"

	p, name, _ := cfg.ActiveProfile("staging")
	if name != "staging" {
		t.Errorf("expected name=staging, got %q", name)
	}
	if p.ClientID != "stg-id" {
		t.Errorf("expected stg-id, got %q", p.ClientID)
	}
}

func TestProfileConfig_MigrateLegacy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// 레거시 형식으로 저장
	legacy := &Config{ClientID: "legacy-id", BotID: "legacy-bot"}
	if err := legacy.Save(path); err != nil {
		t.Fatalf("legacy save: %v", err)
	}

	// 프로필 형식으로 로드 — 자동 마이그레이션
	loaded, err := LoadProfileConfig(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	p, name, _ := loaded.ActiveProfile("")
	if name != "default" {
		t.Errorf("expected migrated to default, got %q", name)
	}
	if p.ClientID != "legacy-id" {
		t.Errorf("expected legacy-id, got %q", p.ClientID)
	}
}

func TestProfileConfig_EnvOverride(t *testing.T) {
	t.Setenv("NW_PROFILE", "env-profile")

	cfg := NewProfileConfig()
	cfg.EnsureProfile("env-profile").ClientID = "env-id"

	p, name, _ := cfg.ActiveProfile("")
	if name != "env-profile" {
		t.Errorf("expected env-profile from NW_PROFILE, got %q", name)
	}
	if p.ClientID != "env-id" {
		t.Errorf("expected env-id, got %q", p.ClientID)
	}
}
