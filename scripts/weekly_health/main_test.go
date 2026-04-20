package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadModuleUpdates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "modules.json")
	content := strings.Join([]string{
		`{"Path":"github.com/physics91/naverworks-cli","Main":true}`,
		`{"Path":"golang.org/x/term","Version":"v0.42.0","Update":{"Version":"v0.43.0"}}`,
		`{"Path":"golang.org/x/sys","Version":"v0.43.0","Indirect":true,"Update":{"Version":"v0.44.0"}}`,
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	updates, err := readModuleUpdates(path)
	if err != nil {
		t.Fatalf("readModuleUpdates failed: %v", err)
	}
	if len(updates) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(updates))
	}
	if updates[0].Path != "golang.org/x/term" || updates[0].Type != "direct" {
		t.Fatalf("unexpected first update: %+v", updates[0])
	}
	if updates[1].Path != "golang.org/x/sys" || updates[1].Type != "indirect" {
		t.Fatalf("unexpected second update: %+v", updates[1])
	}
}

func TestBuildReportClassifiesOpsError(t *testing.T) {
	rep := buildReport(time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC), []checkStatus{
		{Name: "test-full", Status: "failure"},
	}, nil)

	if !rep.HasIssue || !rep.ShouldFailWorkflow {
		t.Fatalf("expected failing report, got %+v", rep)
	}
	if rep.IssueTitle != "[Ops Error] weekly health check failed" {
		t.Fatalf("unexpected title: %q", rep.IssueTitle)
	}
}

func TestBuildReportClassifiesDependencyUpdate(t *testing.T) {
	rep := buildReport(time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC), []checkStatus{
		{Name: "test-full", Status: "success"},
		{Name: "go-vet", Status: "success"},
		{Name: "build", Status: "success"},
		{Name: "module-updates", Status: "success"},
	}, []dependencyUpdate{{Path: "golang.org/x/term", Current: "v0.42.0", Latest: "v0.43.0", Type: "direct"}})

	if !rep.HasIssue || rep.ShouldFailWorkflow {
		t.Fatalf("expected non-failing issue report, got %+v", rep)
	}
	if rep.IssueTitle != "[Weekly Health] 2026-04-20: Dependency updates available" {
		t.Fatalf("unexpected title: %q", rep.IssueTitle)
	}
}
