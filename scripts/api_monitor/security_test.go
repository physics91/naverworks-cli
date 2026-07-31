package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The title comes from untrusted remote HTML and is embedded in a GitHub issue
// body, so it must not be able to inject newlines or control characters.
func TestNormalizeReleaseNoteTitle_StripsControlCharacters(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "plain", raw: "  릴리즈 노트 - 네이버웍스 ", want: "릴리즈 노트"},
		{name: "html entity", raw: "A &amp; B - 네이버웍스", want: "A & B"},
		{name: "newline injection", raw: "Title\n\n## Injected\n- item", want: `Title \#\# Injected - item`},
		{name: "carriage return", raw: "Title\r\nSecond", want: "Title Second"},
		{name: "null byte", raw: "Ti\x00tle", want: "Title"},
		{name: "markdown link injection", raw: "[click](https://evil.example)", want: `\[click\]\(https://evil.example\)`},
		{name: "emphasis and html", raw: "**bold** <img src=x>", want: `\*\*bold\*\* \<img src=x\>`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeReleaseNoteTitle(tc.raw)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
			if strings.ContainsAny(got, "\n\r") {
				t.Errorf("title still contains newlines: %q", got)
			}
		})
	}
}

func TestNormalizeReleaseNoteTitle_TruncatesWithoutBreakingRunes(t *testing.T) {
	raw := strings.Repeat("가", maxReleaseNoteTitleLen+50)
	got := normalizeReleaseNoteTitle(raw)

	runes := []rune(got)
	if len(runes) != maxReleaseNoteTitleLen+1 { // +1 for the ellipsis
		t.Errorf("expected %d runes, got %d", maxReleaseNoteTitleLen+1, len(runes))
	}
	for _, r := range got {
		if r == '\uFFFD' {
			t.Fatalf("truncation produced an invalid rune: %q", got)
		}
	}
}

func TestFetchText_RejectsOversizedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chunk := strings.Repeat("x", 1<<20)
		for written := 0; written <= maxFetchSize; written += len(chunk) {
			if _, err := fmt.Fprint(w, chunk); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	_, err := fetchText(&http.Client{Timeout: 30 * time.Second}, srv.URL)
	if err == nil {
		t.Fatal("expected oversized response to be rejected")
	}
	if !strings.Contains(err.Error(), "응답 크기 초과") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchText_AllowsNormalResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<html><title>ok</title></html>")
	}))
	defer srv.Close()

	body, err := fetchText(&http.Client{Timeout: 30 * time.Second}, srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(body, "<title>ok</title>") {
		t.Fatalf("unexpected body: %s", body)
	}
}
