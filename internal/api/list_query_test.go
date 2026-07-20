package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestBuildListQuery(t *testing.T) {
	got := BuildListQuery("c1", 20, map[string]string{
		"status":     "OPEN",
		"empty":      "",
		"categoryId": "cat1",
	})
	if !strings.HasPrefix(got, "?") {
		t.Fatalf("expected query prefix, got %q", got)
	}
	if !strings.Contains(got, "cursor=c1") || !strings.Contains(got, "count=20") {
		t.Fatalf("missing pagination: %q", got)
	}
	if !strings.Contains(got, "status=OPEN") || !strings.Contains(got, "categoryId=cat1") {
		t.Fatalf("missing extras: %q", got)
	}
	if strings.Contains(got, "empty=") {
		t.Fatalf("empty values should be omitted: %q", got)
	}
	if BuildPaginationQuery("x", 5) != BuildListQuery("x", 5, nil) {
		t.Fatal("BuildPaginationQuery should match BuildListQuery without extras")
	}
}

func TestTaskService_ListTasks_Filters(t *testing.T) {
	var uri string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri = r.URL.RequestURI()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tasks":[]}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(time.Hour)}, nil)
	_, err := NewTaskService(client).ListTasks("u1", "cur", 10, TaskListOptions{
		CategoryID:       "cat9",
		Status:           "PROGRESS",
		SearchFilterType: "ASSIGNEE",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"categoryId=cat9", "status=PROGRESS", "searchFilterType=ASSIGNEE", "cursor=cur", "count=10"} {
		if !strings.Contains(uri, want) {
			t.Fatalf("missing %q in %q", want, uri)
		}
	}
}

func TestApprovalService_ListDocuments_Filters(t *testing.T) {
	var uri string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uri = r.URL.RequestURI()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"documents":[]}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(time.Hour)}, nil)
	_, err := NewApprovalService(client).ListDocuments("n", 50, DocumentListOptions{
		FromDateTime:   "2026-01-01T00:00:00Z",
		UntilDateTime:  "2026-01-31T00:00:00Z",
		DocumentFormID: "form1",
		OrderBy:        "createdTime",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"fromDateTime=2026-01-01T00%3A00%3A00Z",
		"untilDateTime=2026-01-31T00%3A00%3A00Z",
		"documentFormId=form1",
		"orderBy=createdTime",
		"cursor=n",
		"count=50",
	} {
		if !strings.Contains(uri, want) {
			t.Fatalf("missing %q in %q", want, uri)
		}
	}
}
