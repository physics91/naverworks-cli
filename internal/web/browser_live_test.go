package web

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This opt-in check uses real Chrome and the real NAVER WORKS login service.
// It cannot accept authenticated room features; those require a logged-in user.
func TestLiveBrowserUnauthenticated(t *testing.T) {
	if os.Getenv("NW_RUN_LIVE_BROWSER_TEST") != "1" {
		t.Skip("real Chrome/NAVER WORKS check requires NW_RUN_LIVE_BROWSER_TEST=1")
	}
	dir, err := SessionDir(filepath.Join(t.TempDir(), "config.json"), "live-unauthenticated")
	if err != nil {
		t.Fatal(err)
	}
	// Allow chromedp's 20-second WebSocket startup timeout to surface instead
	// of masking an executable-selection failure with our own earlier deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	err = (Browser{SessionDir: dir, Profile: "live-unauthenticated"}).Login(ctx)
	if err == nil {
		t.Fatal("reported success for a browser with no authenticated session")
	}
	if !strings.Contains(err.Error(), "브라우저 로그인을 완료하지 못했습니다") {
		t.Fatalf("browser did not reach the login wait: %v", err)
	}
}
