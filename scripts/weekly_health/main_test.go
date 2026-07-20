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
	if rep.IssueTitle != "[Weekly Health] Dependency updates available" {
		t.Fatalf("unexpected title: %q", rep.IssueTitle)
	}
}

func TestBuildReportIgnoresIndirectOnlyUpdates(t *testing.T) {
	rep := buildReport(time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC), []checkStatus{
		{Name: "test-full", Status: "success"},
		{Name: "go-vet", Status: "success"},
		{Name: "build", Status: "success"},
		{Name: "module-updates", Status: "success"},
	}, []dependencyUpdate{
		{Path: "github.com/cpuguy83/go-md2man/v2", Current: "v2.0.6", Latest: "v2.0.7", Type: "indirect"},
		{Path: "gopkg.in/check.v1", Current: "v0.0.0-old", Latest: "v1.0.0", Type: "indirect"},
	})

	if rep.HasIssue || rep.ShouldFailWorkflow || rep.IssueTitle != "" {
		t.Fatalf("indirect-only updates must not open an issue, got %+v", rep)
	}
	if len(rep.OutdatedModules) != 2 {
		t.Fatalf("expected outdated modules retained for reporting, got %+v", rep.OutdatedModules)
	}
}

func TestRenderMarkdownSeparatesDirectAndIndirect(t *testing.T) {
	rep := report{
		GeneratedAt: time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Checks:      []checkStatus{{Name: "build", Status: "success"}},
		OutdatedModules: []dependencyUpdate{
			{Path: "golang.org/x/term", Current: "v0.42.0", Latest: "v0.43.0", Type: "direct"},
			{Path: "gopkg.in/check.v1", Current: "v0.0.0-old", Latest: "v1.0.0", Type: "indirect"},
		},
		HasIssue:   true,
		IssueTitle: "[Weekly Health] Dependency updates available",
	}
	md := renderMarkdown(rep)
	if !strings.Contains(md, "#### Direct") || !strings.Contains(md, "#### Indirect (informational)") {
		t.Fatalf("expected separated sections, got:\n%s", md)
	}
	if !strings.Contains(md, "golang.org/x/term") || !strings.Contains(md, "gopkg.in/check.v1") {
		t.Fatalf("expected both modules listed, got:\n%s", md)
	}
}
