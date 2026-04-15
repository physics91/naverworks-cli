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

// TestDriveService_CreateUploadURL_Resume verifies that the Zero-Change API
// layer faithfully forwards the caller-supplied "resume" field in the request
// body. This proves that adding --resume support in cmd/drive.go does not
// require any changes inside internal/api/drive.go.
func TestDriveService_CreateUploadURL_Resume(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, _ := io.ReadAll(r.Body)
		capturedBody = string(buf)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)

	svc := NewDriveService(client)
	body := map[string]interface{}{"fileName": "a.txt", "resume": true}
	if _, err := svc.CreateUploadURL("u1", body, 123); err != nil {
		t.Fatalf("CreateUploadURL failed: %v", err)
	}

	if !strings.Contains(capturedBody, `"resume":true`) {
		t.Errorf("expected resume=true in request body, got %s", capturedBody)
	}
	if !strings.Contains(capturedBody, `"fileName":"a.txt"`) {
		t.Errorf("expected fileName in request body, got %s", capturedBody)
	}
	if !strings.Contains(capturedBody, `"fileSize":123`) {
		t.Errorf("expected fileSize=123 in request body, got %s", capturedBody)
	}
}
