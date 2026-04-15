package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

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
