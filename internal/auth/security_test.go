package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/fileutil"
)

func writeKeyWithPerm(t *testing.T, perm os.FileMode) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key gen failed: %v", err)
	}
	path := filepath.Join(t.TempDir(), "private.pem")
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err := os.WriteFile(path, pemData, perm); err != nil {
		t.Fatalf("write key failed: %v", err)
	}
	// os.WriteFile is subject to the process umask; force the intended mode.
	if err := os.Chmod(path, perm); err != nil {
		t.Fatalf("chmod key failed: %v", err)
	}
	return path
}

// BuildJWTAssertion is the single choke point for every signing path (login and
// automatic refresh alike), so it must refuse an over-permissive key itself
// rather than relying on the caller to have validated it.
func TestBuildJWTAssertion_RejectsInsecureKeyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission check")
	}

	keyPath := writeKeyWithPerm(t, 0644)
	if _, err := BuildJWTAssertion("client-id", "sa@example.com", keyPath); err == nil {
		t.Fatal("expected BuildJWTAssertion to reject a 0644 private key")
	} else if !strings.Contains(err.Error(), "0600") {
		t.Fatalf("expected 0600 guidance, got %v", err)
	}

	if err := os.Chmod(keyPath, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildJWTAssertion("client-id", "sa@example.com", keyPath); err != nil {
		t.Fatalf("expected 0600 key to be accepted, got %v", err)
	}
}

// Tokens written by an older version, restored from a backup, or copied with a
// permissive umask must not be read as-is: the mode is repaired before the
// secret is used.
func TestProfileTokenStore_Load_RepairsInsecurePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission check")
	}

	path := filepath.Join(t.TempDir(), "token.json")
	data, err := json.Marshal(profileTokenFile{Tokens: map[string]*Token{
		"default": {
			AuthMethod:  AuthMethodOAuth,
			AccessToken: "secret-access-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(time.Hour),
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	// os.WriteFile is subject to the process umask; force the intended mode.
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}

	var warnings strings.Builder
	original := fileutil.SecureFileWarnWriter
	fileutil.SecureFileWarnWriter = &warnings
	defer func() { fileutil.SecureFileWarnWriter = original }()

	token, err := NewProfileTokenStore(path, "default").Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if token == nil || token.AccessToken != "secret-access-token" {
		t.Fatalf("expected token to load, got %+v", token)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected token file to be tightened to 0600, got %04o", perm)
	}
	if !strings.Contains(warnings.String(), "0644") {
		t.Errorf("expected a warning mentioning the original mode, got %q", warnings.String())
	}
}

func TestTokenStore_Load_RepairsInsecurePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission check")
	}

	path := filepath.Join(t.TempDir(), "token.json")
	data, err := json.Marshal(&Token{
		AuthMethod:  AuthMethodOAuth,
		AccessToken: "secret-access-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0640); err != nil {
		t.Fatal(err)
	}
	// os.WriteFile is subject to the process umask; force the intended mode.
	if err := os.Chmod(path, 0640); err != nil {
		t.Fatal(err)
	}

	var warnings strings.Builder
	original := fileutil.SecureFileWarnWriter
	fileutil.SecureFileWarnWriter = &warnings
	defer func() { fileutil.SecureFileWarnWriter = original }()

	if _, err := NewTokenStore(path).Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected token file to be tightened to 0600, got %04o", perm)
	}
}
