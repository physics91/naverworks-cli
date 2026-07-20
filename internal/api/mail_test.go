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
	var capturedMethod, capturedPath, capturedBody, capturedContentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		capturedPath = r.URL.Path
		capturedContentType = r.Header.Get("Content-Type")
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
	if capturedMethod != http.MethodPatch {
		t.Fatalf("expected PATCH, got %s", capturedMethod)
	}
	if capturedPath != "/users/u1/mail/m1" {
		t.Fatalf("unexpected path: %s", capturedPath)
	}
	if !strings.Contains(capturedContentType, "application/json") {
		t.Fatalf("expected JSON content type, got %q", capturedContentType)
	}
	if !strings.Contains(capturedBody, `"folderId":42`) {
		t.Fatalf("expected numeric folderId, got %q", capturedBody)
	}

	if _, err := svc.MoveMail("u1", "m1", "inbox"); err == nil {
		t.Fatal("expected error for non-integer folderId")
	} else if !strings.Contains(err.Error(), "folderId는 정수여야 합니다") {
		t.Fatalf("unexpected error: %v", err)
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
