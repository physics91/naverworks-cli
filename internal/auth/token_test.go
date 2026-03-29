package auth

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestTokenStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	store := NewTokenStore(path)

	token := &Token{
		AuthMethod:   "oauth",
		AccessToken:  "at-123",
		RefreshToken: "rt-456",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "bot directory calendar",
	}
	if err := store.Save(token); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.AccessToken != "at-123" {
		t.Errorf("expected at-123, got %q", loaded.AccessToken)
	}
	if loaded.AuthMethod != "oauth" {
		t.Errorf("expected oauth, got %q", loaded.AuthMethod)
	}
}

func TestTokenStore_LoadNotExist(t *testing.T) {
	store := NewTokenStore("/nonexistent/token.json")
	token, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != nil {
		t.Error("expected nil token")
	}
}

func TestToken_IsExpired(t *testing.T) {
	token := &Token{ExpiresAt: time.Now().Add(-1 * time.Minute)}
	if !token.IsExpired() {
		t.Error("expected expired")
	}

	token2 := &Token{ExpiresAt: time.Now().Add(1 * time.Hour)}
	if token2.IsExpired() {
		t.Error("expected not expired")
	}
}

func TestToken_NeedsRefresh(t *testing.T) {
	token := &Token{ExpiresAt: time.Now().Add(30 * time.Second)}
	if !token.NeedsRefresh() {
		t.Error("expected needs refresh (within 60s buffer)")
	}

	token2 := &Token{ExpiresAt: time.Now().Add(5 * time.Minute)}
	if token2.NeedsRefresh() {
		t.Error("expected no refresh needed")
	}
}

func TestTokenStore_Delete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	store := NewTokenStore(path)

	store.Save(&Token{AccessToken: "test"})
	if err := store.Delete(); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	token, _ := store.Load()
	if token != nil {
		t.Error("expected nil after delete")
	}
}

func TestProfileTokenStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	store := NewProfileTokenStore(path, "work")
	token := &Token{
		AuthMethod:  "oauth",
		AccessToken: "work-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	if err := store.Save(token); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.AccessToken != "work-token" {
		t.Errorf("expected work-token, got %q", loaded.AccessToken)
	}
}

func TestProfileTokenStore_MultipleProfiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	store1 := NewProfileTokenStore(path, "work")
	store1.Save(&Token{AuthMethod: "oauth", AccessToken: "work-tk", ExpiresAt: time.Now().Add(1 * time.Hour)})

	store2 := NewProfileTokenStore(path, "staging")
	store2.Save(&Token{AuthMethod: "jwt", AccessToken: "stg-tk", ExpiresAt: time.Now().Add(1 * time.Hour)})

	t1, _ := store1.Load()
	t2, _ := store2.Load()

	if t1.AccessToken != "work-tk" {
		t.Errorf("work profile: expected work-tk, got %q", t1.AccessToken)
	}
	if t2.AccessToken != "stg-tk" {
		t.Errorf("staging profile: expected stg-tk, got %q", t2.AccessToken)
	}
}

func TestProfileTokenStore_MigrateLegacy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	// Save in legacy format
	legacy := &Token{AuthMethod: "oauth", AccessToken: "legacy-tk", ExpiresAt: time.Now().Add(1 * time.Hour)}
	legacyStore := NewTokenStore(path)
	legacyStore.Save(legacy)

	// Load with profile store — auto migration
	profileStore := NewProfileTokenStore(path, "default")
	loaded, err := profileStore.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.AccessToken != "legacy-tk" {
		t.Errorf("expected legacy-tk, got %q", loaded.AccessToken)
	}
}

func TestProfileTokenStore_DeleteProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	store := NewProfileTokenStore(path, "temp")
	store.Save(&Token{AuthMethod: "jwt", AccessToken: "tmp", ExpiresAt: time.Now().Add(1 * time.Hour)})
	store.Delete()

	loaded, _ := store.Load()
	if loaded != nil {
		t.Error("expected nil after delete")
	}
}

func TestProfileTokenStore_DeleteNonDefault_PreservesLegacy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	// Save legacy format
	legacy := &Token{AuthMethod: "oauth", AccessToken: "default-tk", ExpiresAt: time.Now().Add(1 * time.Hour)}
	legacyStore := NewTokenStore(path)
	legacyStore.Save(legacy)

	// Delete staging profile — should preserve default token
	stagingStore := NewProfileTokenStore(path, "staging")
	stagingStore.Delete()

	defaultStore := NewProfileTokenStore(path, "default")
	loaded, _ := defaultStore.Load()
	if loaded == nil || loaded.AccessToken != "default-tk" {
		t.Error("default token should be preserved after deleting non-default profile")
	}
}

func TestWriteSecureJSON_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "token.json")

	token := &Token{AccessToken: "new-file-test"}
	if err := writeSecureJSON(path, token); err != nil {
		t.Fatalf("writeSecureJSON failed: %v", err)
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

func TestWriteSecureJSON_OverwriteExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	// Write initial file
	if err := writeSecureJSON(path, &Token{AccessToken: "first"}); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// Overwrite
	if err := writeSecureJSON(path, &Token{AccessToken: "second"}); err != nil {
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

func TestWriteSecureJSON_NoTempFileLeftOver(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	if err := writeSecureJSON(path, &Token{AccessToken: "cleanup-test"}); err != nil {
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

func TestWriteSecureJSON_DirPermissionsPreserved(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission checks not applicable on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	if err := writeSecureJSON(path, &Token{AccessToken: "perm-test"}); err != nil {
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

func TestWriteSecureJSON_OriginalPreservedOnMarshalError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("atomic write not used on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	// Write initial file
	if err := writeSecureJSON(path, &Token{AccessToken: "original"}); err != nil {
		t.Fatalf("initial write failed: %v", err)
	}

	// Attempt to write something that will fail serialization (channel is not JSON-serializable)
	err := writeSecureJSON(path, make(chan int))
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
