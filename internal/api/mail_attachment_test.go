package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestMailService_GetAttachment_AllowsLargeAttachmentPayload(t *testing.T) {
	largeBody := strings.Repeat("x", maxAPIResponseSize+1024)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(largeBody))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)

	resp, err := NewMailService(client).GetAttachment("u1", "m1", "a1")
	if err != nil {
		t.Fatalf("unexpected error for valid large attachment payload: %v", err)
	}
	if len(resp.Body) != len(largeBody) {
		t.Fatalf("body length = %d, want %d", len(resp.Body), len(largeBody))
	}
}
