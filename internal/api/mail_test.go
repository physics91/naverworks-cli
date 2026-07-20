package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestMailService_MoveMail_FolderIDEncoding(t *testing.T) {
	var capturedPath string
	var capturedBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		capturedBody = string(raw)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(srv.URL, token, nil)
	svc := NewMailService(client)

	if _, err := svc.MoveMail("u1", "m1", "42"); err != nil {
		t.Fatalf("MoveMail numeric folder failed: %v", err)
	}
	if capturedPath != "/users/u1/mail/m1" {
		t.Fatalf("unexpected path: %s", capturedPath)
	}
	if !strings.Contains(capturedBody, `"folderId":42`) {
		t.Fatalf("expected numeric folderId, got %q", capturedBody)
	}

	if _, err := svc.MoveMail("u1", "m1", "inbox"); err != nil {
		t.Fatalf("MoveMail string folder failed: %v", err)
	}
	if !strings.Contains(capturedBody, `"folderId":"inbox"`) {
		t.Fatalf("expected string folderId, got %q", capturedBody)
	}
}

func TestMailService_GetMail_HasThreads(t *testing.T) {
	var capturedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.RequestURI()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(srv.URL, token, nil)
	svc := NewMailService(client)

	if _, err := svc.GetMail("u1", "m1", true); err != nil {
		t.Fatalf("GetMail(hasThreads=true) failed: %v", err)
	}
	if !strings.Contains(capturedPath, "hasThreads=true") {
		t.Errorf("expected hasThreads=true in path, got %q", capturedPath)
	}

	if _, err := svc.GetMail("u1", "m1", false); err != nil {
		t.Fatalf("GetMail(hasThreads=false) failed: %v", err)
	}
	if strings.Contains(capturedPath, "hasThreads") {
		t.Errorf("expected no hasThreads when false, got %q", capturedPath)
	}
}
