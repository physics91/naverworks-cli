package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindOpenIssueByExactTitle(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		_ = json.NewEncoder(w).Encode(searchResponse{
			Items: []issue{
				{Number: 10, Title: "other", HTMLURL: "https://example.com/10"},
				{Number: 26, Title: "[Weekly Health] Dependency updates available", HTMLURL: "https://example.com/26"},
			},
		})
	}))
	defer server.Close()

	client := server.Client()
	// Rewrite requests to the test server by using a custom transport via absolute URL helper.
	// findOpenIssueByExactTitle hits api.github.com; inject via roundTrip rewrite.
	client = &http.Client{Transport: rewriteHost{base: server.URL, next: http.DefaultTransport}}

	found, err := findOpenIssueByExactTitle(client, "physics91/naverworks-cli", "token", "[Weekly Health] Dependency updates available")
	if err != nil {
		t.Fatalf("findOpenIssueByExactTitle: %v", err)
	}
	if found == nil || found.Number != 26 {
		t.Fatalf("expected issue #26, got %+v", found)
	}
	if !strings.Contains(gotQuery, `repo:physics91/naverworks-cli`) {
		t.Fatalf("unexpected search query: %q", gotQuery)
	}
}

func TestUpdateIssueBody(t *testing.T) {
	var method, path, body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		_ = json.NewEncoder(w).Encode(issue{
			Number:  26,
			Title:   "[Weekly Health] Dependency updates available",
			HTMLURL: "https://example.com/26",
		})
	}))
	defer server.Close()

	client := &http.Client{Transport: rewriteHost{base: server.URL, next: http.DefaultTransport}}
	updated, err := updateIssueBody(client, "physics91/naverworks-cli", "token", 26, "## refreshed body\n")
	if err != nil {
		t.Fatalf("updateIssueBody: %v", err)
	}
	if updated == nil || updated.Number != 26 {
		t.Fatalf("unexpected updated issue: %+v", updated)
	}
	if method != http.MethodPatch {
		t.Fatalf("expected PATCH, got %s", method)
	}
	if path != "/repos/physics91/naverworks-cli/issues/26" {
		t.Fatalf("unexpected path: %s", path)
	}
	if !strings.Contains(body, "refreshed body") {
		t.Fatalf("unexpected patch body: %s", body)
	}
}

func TestCreateIssue(t *testing.T) {
	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(issue{
			Number:  99,
			Title:   "[Weekly Health] Dependency updates available",
			HTMLURL: "https://example.com/99",
		})
	}))
	defer server.Close()

	client := &http.Client{Transport: rewriteHost{base: server.URL, next: http.DefaultTransport}}
	created, err := createIssue(client, "physics91/naverworks-cli", "token", "[Weekly Health] Dependency updates available", "body", []string{"health-check"})
	if err != nil {
		t.Fatalf("createIssue: %v", err)
	}
	if created == nil || created.Number != 99 {
		t.Fatalf("unexpected created issue: %+v", created)
	}
	if method != http.MethodPost {
		t.Fatalf("expected POST, got %s", method)
	}
	if path != "/repos/physics91/naverworks-cli/issues" {
		t.Fatalf("unexpected path: %s", path)
	}
}

func TestSplitLabels(t *testing.T) {
	got := splitLabels(" health-check , ,ops-error ")
	if len(got) != 2 || got[0] != "health-check" || got[1] != "ops-error" {
		t.Fatalf("unexpected labels: %#v", got)
	}
	if splitLabels("") != nil {
		t.Fatalf("expected nil for empty labels")
	}
}

// Ensure body file read path used by main stays trivial to exercise in isolation.
func TestBodyFileReadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.md")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != "hello" {
		t.Fatalf("body file read failed: %v %q", err, raw)
	}
}

// rewriteHost sends api.github.com requests to the httptest server.
type rewriteHost struct {
	base string
	next http.RoundTripper
}

func (r rewriteHost) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	u := *req.URL
	baseURL, err := http.NewRequest(http.MethodGet, r.base, nil)
	if err != nil {
		return nil, err
	}
	u.Scheme = baseURL.URL.Scheme
	u.Host = baseURL.URL.Host
	clone.URL = &u
	clone.Host = u.Host
	return r.next.RoundTrip(clone)
}
