package cmd

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

	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
)

func TestLoginJWT_RejectsInsecurePrivateKeyPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission check")
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key gen failed: %v", err)
	}

	keyPath := filepath.Join(t.TempDir(), "private.pem")
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err := os.WriteFile(keyPath, pemData, 0644); err != nil {
		t.Fatalf("write key failed: %v", err)
	}

	cfg := &config.Config{
		ClientID:         "client-id",
		ClientSecret:     "client-secret",
		ServiceAccountID: "service-account",
		PrivateKeyPath:   keyPath,
	}
	store := auth.NewProfileTokenStore(filepath.Join(t.TempDir(), "token.json"), "default")

	err = loginJWT(cfg, store)
	if err == nil {
		t.Fatal("expected insecure key permission error")
	}
	if !strings.Contains(err.Error(), "0600") {
		t.Fatalf("expected 0600 guidance, got %v", err)
	}
}
