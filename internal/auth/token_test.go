package auth

import (
	"path/filepath"
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
