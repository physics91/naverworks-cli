package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/physics91/naverworks-cli/internal/fileutil"
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

func TestConfigPathFromDir(t *testing.T) {
	t.Run("absolute config dir", func(t *testing.T) {
		got, err := configPathFromDir("/tmp/naverworks-test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join("/tmp/naverworks-test", "naverworks", "config.json")
		if got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
	})

	t.Run("empty config dir rejected", func(t *testing.T) {
		_, err := configPathFromDir("")
		if err == nil {
			t.Fatal("expected error for empty config dir")
		}
		if !strings.Contains(err.Error(), "비어") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("relative config dir rejected", func(t *testing.T) {
		_, err := configPathFromDir(".config")
		if err == nil {
			t.Fatal("expected error for relative config dir")
		}
		if !strings.Contains(err.Error(), "절대 경로") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDefaultPathOrError_PropagatesLookupFailure(t *testing.T) {
	_, err := defaultPathOrError(func() (string, error) {
		return "", errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected lookup error")
	}
	if !strings.Contains(err.Error(), "설정 디렉토리 조회 실패") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSaveSecureJSON_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.json")

	cfg := &Config{ClientID: "new-file-test"}
	if err := fileutil.WriteSecureJSON(path, cfg); err != nil {
		t.Fatalf("WriteSecureJSON failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !strings.Contains(string(data), "new-file-test") {
		t.Error("expected file to contain new-file-test")
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat failed: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Errorf("expected file perm 0600, got %04o", perm)
		}
	}
}

func TestSaveSecureJSON_OverwriteExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Write initial
	if err := fileutil.WriteSecureJSON(path, &Config{ClientID: "first"}); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// Overwrite
	if err := fileutil.WriteSecureJSON(path, &Config{ClientID: "second"}); err != nil {
		t.Fatalf("overwrite failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !strings.Contains(string(data), "second") {
		t.Error("expected file to contain second")
	}
	if strings.Contains(string(data), "first") {
		t.Error("expected file to not contain first")
	}
}

func TestSaveSecureJSON_NoTempFileLeftOver(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := fileutil.WriteSecureJSON(path, &Config{ClientID: "cleanup-test"}); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".naverworks-") && strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("leftover temp file found: %s", e.Name())
		}
	}
}

func TestSaveSecureJSON_DirPermissionsPreserved(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission checks not applicable on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	if err := fileutil.WriteSecureJSON(path, &Config{ClientID: "perm-test"}); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	dirInfo, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir failed: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0700 {
		t.Errorf("expected dir perm 0700, got %04o", perm)
	}
}

func TestSaveSecureJSON_OriginalPreservedOnMarshalError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("atomic write not used on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Write initial file
	if err := fileutil.WriteSecureJSON(path, &Config{ClientID: "original"}); err != nil {
		t.Fatalf("initial write failed: %v", err)
	}

	// Attempt to write something that will fail serialization
	err := fileutil.WriteSecureJSON(path, make(chan int))
	if err == nil {
		t.Fatal("expected serialization error")
	}

	// Original file should be untouched
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !strings.Contains(string(data), "original") {
		t.Error("original file content should be preserved after serialization failure")
	}
}
