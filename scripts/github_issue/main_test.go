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

func TestFindOwnedOpenIssue(t *testing.T) {
	var gotPath, gotState, gotPageSize string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotState = r.URL.Query().Get("state")
		gotPageSize = r.URL.Query().Get("per_page")
		_ = json.NewEncoder(w).Encode([]issue{
			{Number: 10, Title: "other", State: "open", Labels: []issueLabel{{Name: "ops-error"}}, HTMLURL: "https://example.com/10"},
			{Number: 29, Title: "[Ops Error] weekly health check failed", State: "open", Labels: []issueLabel{{Name: "ops-error"}}, HTMLURL: "https://example.com/29"},
		})
	}))
	defer server.Close()

	client := server.Client()
	// Rewrite requests to the test server by using a custom transport via absolute URL helper.
	// findOpenIssueByExactTitle hits api.github.com; inject via roundTrip rewrite.
	client = &http.Client{Transport: rewriteHost{base: server.URL, next: http.DefaultTransport}}

	found, err := findOwnedOpenIssue(client, "physics91/naverworks-cli", "token", "[Ops Error] weekly health check failed", []string{"ops-error"})
	if err != nil {
		t.Fatalf("findOwnedOpenIssue: %v", err)
	}
	if found == nil || found.Number != 29 {
		t.Fatalf("expected issue #29, got %+v", found)
	}
	if gotPath != "/repos/physics91/naverworks-cli/issues" || gotState != "open" || gotPageSize != "100" {
		t.Fatalf("unexpected list request: path=%q state=%q per_page=%q", gotPath, gotState, gotPageSize)
	}
}

func TestFindOwnedOpenIssuePaginatesCompleteRepositoryList(t *testing.T) {
	var pages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		pages = append(pages, page)
		if page == "1" {
			items := make([]issue, 100)
			for i := range items {
				items[i] = issue{Number: i + 1, Title: "other", State: "open"}
			}
			// Repository issue listings include pull requests. Even an exact-title
			// pull request must not become the automation target.
			items[0] = issue{
				Number:      1,
				Title:       "[Ops Error] weekly health check failed",
				State:       "open",
				Labels:      []issueLabel{{Name: "ops-error"}},
				PullRequest: json.RawMessage(`{"url":"https://example.com/pulls/1"}`),
			}
			_ = json.NewEncoder(w).Encode(items)
			return
		}
		_ = json.NewEncoder(w).Encode([]issue{{
			Number:  101,
			Title:   "[Ops Error] weekly health check failed",
			State:   "open",
			Labels:  []issueLabel{{Name: "ops-error"}},
			HTMLURL: "https://example.com/101",
		}})
	}))
	defer server.Close()

	client := &http.Client{Transport: rewriteHost{base: server.URL, next: http.DefaultTransport}}
	found, err := findOwnedOpenIssue(client, "physics91/naverworks-cli", "token", "[Ops Error] weekly health check failed", []string{"ops-error"})
	if err != nil {
		t.Fatalf("findOwnedOpenIssue: %v", err)
	}
	if found == nil || found.Number != 101 {
		t.Fatalf("expected second-page issue #101, got %+v", found)
	}
	if got := strings.Join(pages, ","); got != "1,2" {
		t.Fatalf("expected complete pagination, got pages %s", got)
	}
}

func TestFindOwnedOpenIssueRejectsLabelMismatch(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		_ = json.NewEncoder(w).Encode([]issue{{
			Number: 29,
			Title:  "[Ops Error] weekly health check failed",
			State:  "open",
			Labels: []issueLabel{{Name: "human-owned"}},
		}})
	}))
	defer server.Close()

	client := &http.Client{Transport: rewriteHost{base: server.URL, next: http.DefaultTransport}}
	_, err := findOwnedOpenIssue(client, "physics91/naverworks-cli", "token", "[Ops Error] weekly health check failed", []string{"ops-error"})
	if err == nil || !strings.Contains(err.Error(), "ownership mismatch") {
		t.Fatalf("expected ownership mismatch, got %v", err)
	}
	if requests != 1 {
		t.Fatalf("expected one read-only request, got %d", requests)
	}
}

func TestFindOwnedOpenIssueRejectsDuplicateTargets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]issue{
			{Number: 29, Title: "[Ops Error] weekly health check failed", State: "open", Labels: []issueLabel{{Name: "ops-error"}}},
			{Number: 31, Title: "[Ops Error] weekly health check failed", State: "open", Labels: []issueLabel{{Name: "ops-error"}}},
		})
	}))
	defer server.Close()

	client := &http.Client{Transport: rewriteHost{base: server.URL, next: http.DefaultTransport}}
	_, err := findOwnedOpenIssue(client, "physics91/naverworks-cli", "token", "[Ops Error] weekly health check failed", []string{"ops-error"})
	if err == nil || !strings.Contains(err.Error(), "ambiguous automation target") {
		t.Fatalf("expected ambiguous target error, got %v", err)
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

func TestReconcileIssueLifecycle(t *testing.T) {
	var current *issue
	var mutations []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/physics91/naverworks-cli/issues":
			items := []issue(nil)
			if current != nil && current.State == "open" {
				items = append(items, *current)
			}
			_ = json.NewEncoder(w).Encode(items)
		case r.Method == http.MethodPost && r.URL.Path == "/repos/physics91/naverworks-cli/issues":
			mutations = append(mutations, "create")
			current = &issue{Number: 29, Title: "[Ops Error] weekly health check failed", State: "open", Labels: []issueLabel{{Name: "ops-error"}}, HTMLURL: "https://example.com/29"}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(current)
		case r.Method == http.MethodPatch && r.URL.Path == "/repos/physics91/naverworks-cli/issues/29":
			var payload struct {
				Body  *string `json:"body"`
				State string  `json:"state"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			if payload.State == "closed" {
				mutations = append(mutations, "close")
				current.State = "closed"
			} else {
				mutations = append(mutations, "update")
			}
			_ = json.NewEncoder(w).Encode(current)
		default:
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client := &http.Client{Transport: rewriteHost{base: server.URL, next: http.DefaultTransport}}
	labels := []string{"ops-error"}

	action, _, err := reconcileIssue(client, "physics91/naverworks-cli", "token", "[Ops Error] weekly health check failed", "first", labels, "open")
	if err != nil || action != "created" {
		t.Fatalf("create reconcile: action=%q err=%v", action, err)
	}
	action, _, err = reconcileIssue(client, "physics91/naverworks-cli", "token", "[Ops Error] weekly health check failed", "second", labels, "open")
	if err != nil || action != "updated" {
		t.Fatalf("update reconcile: action=%q err=%v", action, err)
	}
	action, _, err = reconcileIssue(client, "physics91/naverworks-cli", "token", "[Ops Error] weekly health check failed", "", labels, "closed")
	if err != nil || action != "closed" {
		t.Fatalf("close reconcile: action=%q err=%v", action, err)
	}
	action, closed, err := reconcileIssue(client, "physics91/naverworks-cli", "token", "[Ops Error] weekly health check failed", "", labels, "closed")
	if err != nil || action != "unchanged" || closed != nil {
		t.Fatalf("idempotent close reconcile: action=%q issue=%+v err=%v", action, closed, err)
	}
	if got := strings.Join(mutations, ","); got != "create,update,close" {
		t.Fatalf("unexpected mutation sequence: %s", got)
	}
}

func TestInvalidRepositoryFailsBeforeRequest(t *testing.T) {
	requests := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		return nil, nil
	})}
	_, _, err := reconcileIssue(client, "physics91/naverworks-cli/other", "token", "title", "body", []string{"ops-error"}, "open")
	if err == nil || !strings.Contains(err.Error(), "invalid repo") {
		t.Fatalf("expected invalid repo error, got %v", err)
	}
	if requests != 0 {
		t.Fatalf("invalid repository must not trigger requests, got %d", requests)
	}
}

func TestWeeklyHealthWorkflowReconcilesOpsIssue(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "weekly-health.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(raw)
	for _, required := range []string{
		"Close recovered weekly health ops issue",
		`--title "[Ops Error] weekly health check failed"`,
		`--labels "ops-error"`,
		`--state closed`,
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("workflow missing %q", required)
		}
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
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
