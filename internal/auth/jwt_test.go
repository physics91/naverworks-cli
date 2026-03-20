package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func generateTestKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key gen failed: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "private.pem")
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	os.WriteFile(path, pemData, 0600)
	return path
}

func TestBuildJWTAssertion(t *testing.T) {
	keyPath := generateTestKey(t)
	assertion, err := BuildJWTAssertion("client-id", "sa@example.com", keyPath)
	if err != nil {
		t.Fatalf("build assertion failed: %v", err)
	}
	if assertion == "" {
		t.Error("expected non-empty assertion")
	}
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		t.Errorf("expected 3 parts in JWT, got %d", len(parts))
	}
}

func TestBuildJWTAssertion_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.pem")
	os.WriteFile(path, []byte("not a pem"), 0600)

	_, err := BuildJWTAssertion("client-id", "sa@example.com", path)
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestBuildJWTAssertion_FileNotFound(t *testing.T) {
	_, err := BuildJWTAssertion("client-id", "sa@example.com", "/nonexistent.pem")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestCheckKeyPermissions_Unix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only test")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "key.pem")
	os.WriteFile(path, []byte("test"), 0644)

	warning := CheckKeyPermissions(path)
	if warning == "" {
		t.Error("expected warning for 0644 permissions")
	}

	os.Chmod(path, 0600)
	warning = CheckKeyPermissions(path)
	if warning != "" {
		t.Errorf("expected no warning for 0600, got %q", warning)
	}
}
