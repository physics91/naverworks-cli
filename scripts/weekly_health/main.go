package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type moduleRecord struct {
	Path     string `json:"Path"`
	Version  string `json:"Version"`
	Indirect bool   `json:"Indirect"`
	Main     bool   `json:"Main"`
	Update   *struct {
		Version string `json:"Version"`
	} `json:"Update"`
}

type dependencyUpdate struct {
	Path    string `json:"path"`
	Current string `json:"current"`
	Latest  string `json:"latest"`
	Type    string `json:"type"`
}

type checkStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type report struct {
	GeneratedAt        string             `json:"generated_at"`
	Checks             []checkStatus      `json:"checks"`
	OutdatedModules    []dependencyUpdate `json:"outdated_modules"`
	HasIssue           bool               `json:"has_issue"`
	ShouldFailWorkflow bool               `json:"should_fail_workflow"`
	IssueTitle         string             `json:"issue_title,omitempty"`
	IssueLabels        []string           `json:"issue_labels,omitempty"`
}

func main() {
	var (
		modulesPath   = flag.String("modules", "", "path to go list -m -u -json all output")
		testStatus    = flag.String("test-status", "skipped", "test step outcome")
		vetStatus     = flag.String("vet-status", "skipped", "vet step outcome")
		buildStatus   = flag.String("build-status", "skipped", "build step outcome")
		modulesStatus = flag.String("modules-status", "skipped", "module scan step outcome")
		outputJSON    = flag.String("output-json", "", "write JSON report to file")
		outputMD      = flag.String("output-markdown", "", "write Markdown issue body to file")
	)
	flag.Parse()

	checks := []checkStatus{
		{Name: "test-full", Status: normalizeStatus(*testStatus)},
		{Name: "go-vet", Status: normalizeStatus(*vetStatus)},
		{Name: "build", Status: normalizeStatus(*buildStatus)},
		{Name: "module-updates", Status: normalizeStatus(*modulesStatus)},
	}

	var updates []dependencyUpdate
	if normalizeStatus(*modulesStatus) == "success" && strings.TrimSpace(*modulesPath) != "" {
		var err error
		updates, err = readModuleUpdates(*modulesPath)
		if err != nil {
			fatalf("의존성 업데이트 파싱 실패: %v", err)
		}
	}

	rep := buildReport(time.Now().UTC(), checks, updates)
	if err := writeReportJSON(*outputJSON, rep); err != nil {
		fatalf("JSON 리포트 저장 실패: %v", err)
	}
	if err := writeMarkdown(*outputMD, renderMarkdown(rep)); err != nil {
		fatalf("Markdown 리포트 저장 실패: %v", err)
	}

	fmt.Printf("has issue: %t\n", rep.HasIssue)
}

func normalizeStatus(status string) string {
	status = strings.TrimSpace(strings.ToLower(status))
	if status == "" {
		return "unknown"
	}
	return status
}

func readModuleUpdates(path string) ([]dependencyUpdate, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var updates []dependencyUpdate
	dec := json.NewDecoder(f)
	for {
		var record moduleRecord
		err := dec.Decode(&record)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if record.Main || record.Update == nil || record.Update.Version == "" {
			continue
		}
		updateType := "direct"
		if record.Indirect {
			updateType = "indirect"
		}
		updates = append(updates, dependencyUpdate{
			Path:    record.Path,
			Current: record.Version,
			Latest:  record.Update.Version,
			Type:    updateType,
		})
	}

	sort.Slice(updates, func(i, j int) bool {
		if updates[i].Type != updates[j].Type {
			return updates[i].Type < updates[j].Type
		}
		return updates[i].Path < updates[j].Path
	})
	return updates, nil
}

func buildReport(now time.Time, checks []checkStatus, updates []dependencyUpdate) report {
	rep := report{
		GeneratedAt:     now.Format(time.RFC3339),
		Checks:          checks,
		OutdatedModules: updates,
	}

	date := now.Format("2006-01-02")
	if hasCheckFailures(checks) {
		rep.HasIssue = true
		rep.ShouldFailWorkflow = true
		rep.IssueTitle = "[Ops Error] weekly health check failed"
		rep.IssueLabels = []string{"ops-error"}
		return rep
	}
	if len(updates) > 0 {
		rep.HasIssue = true
		rep.ShouldFailWorkflow = false
		rep.IssueTitle = fmt.Sprintf("[Weekly Health] %s: Dependency updates available", date)
		rep.IssueLabels = []string{"health-check"}
	}
	return rep
}

func hasCheckFailures(checks []checkStatus) bool {
	for _, check := range checks {
		if check.Status != "success" {
			return true
		}
	}
	return false
}

func renderMarkdown(rep report) string {
	var b strings.Builder
	date := ""
	if t, err := time.Parse(time.RFC3339, rep.GeneratedAt); err == nil {
		date = t.Format("2006-01-02")
	}

	b.WriteString(fmt.Sprintf("## Weekly Health Check — %s\n\n", date))
	b.WriteString(fmt.Sprintf("- 생성 시각: `%s`\n\n", rep.GeneratedAt))

	b.WriteString("### Command Status\n\n")
	b.WriteString("| Check | Status |\n")
	b.WriteString("| --- | --- |\n")
	for _, check := range rep.Checks {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", check.Name, check.Status))
	}
	b.WriteString("\n")

	if hasCheckFailures(rep.Checks) {
		b.WriteString("### Failure Summary\n\n")
		for _, check := range rep.Checks {
			if check.Status == "success" {
				continue
			}
			b.WriteString(fmt.Sprintf("- `%s` 단계가 `%s` 상태로 종료되었습니다.\n", check.Name, check.Status))
		}
		b.WriteString("\n로그는 workflow artifact에서 확인할 수 있습니다.\n")
		return b.String()
	}

	b.WriteString("### Dependency Updates\n\n")
	if len(rep.OutdatedModules) == 0 {
		b.WriteString("업데이트 가능한 Go 모듈이 없습니다.\n")
		return b.String()
	}

	b.WriteString("| Module | Current | Latest | Type |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, update := range rep.OutdatedModules {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` |\n", update.Path, update.Current, update.Latest, update.Type))
	}
	return b.String()
}

func writeReportJSON(path string, rep report) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}

func writeMarkdown(path, content string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
