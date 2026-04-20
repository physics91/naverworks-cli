package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	defaultPageURL    = "https://naver.worksmobile.com/release-notes/?page=1"
	defaultMaxEntries = 40
)

var (
	releaseNoteSlugRE = regexp.MustCompile(`(?:https://naver\.worksmobile\.com)?/release-notes/([A-Za-z0-9_-]+)/`)
	ogTitleRE         = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:title["'][^>]+content=["']([^"']+)["']`)
	titleTagRE        = regexp.MustCompile(`(?is)<title>(.*?)</title>`)
)

type baseline struct {
	KnownSlugs []string `json:"known_slugs"`
}

type entry struct {
	Slug  string `json:"slug"`
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
}

type report struct {
	CheckedAt          string   `json:"checked_at"`
	PageURL            string   `json:"page_url"`
	NewEntries         []entry  `json:"new_entries"`
	HasIssue           bool     `json:"has_issue"`
	ShouldFailWorkflow bool     `json:"should_fail_workflow"`
	IssueTitle         string   `json:"issue_title,omitempty"`
	IssueLabels        []string `json:"issue_labels,omitempty"`
}

type issueSearchResponse struct {
	Items []struct {
		Title string `json:"title"`
	} `json:"items"`
}

func main() {
	var (
		baselinePath = flag.String("baseline", "", "baseline JSON path")
		pageURL      = flag.String("page-url", defaultPageURL, "release note page URL")
		maxEntries   = flag.Int("max-entries", defaultMaxEntries, "max entries to inspect from the page")
		outputJSON   = flag.String("output-json", "", "write JSON report to file")
		outputMD     = flag.String("output-markdown", "", "write Markdown issue body to file")
		githubRepo   = flag.String("github-repo", os.Getenv("GITHUB_REPOSITORY"), "GitHub repository in owner/repo form")
		githubToken  = flag.String("github-token", os.Getenv("GITHUB_TOKEN"), "GitHub token used to suppress already-tracked entries")
	)
	flag.Parse()

	if strings.TrimSpace(*baselinePath) == "" {
		fatalf("--baseline 플래그가 필요합니다")
	}

	knownSlugs, err := loadBaseline(*baselinePath)
	if err != nil {
		fatalf("baseline 로드 실패: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	pageHTML, err := fetchText(client, *pageURL)
	if err != nil {
		fatalf("릴리즈 노트 페이지 조회 실패: %v", err)
	}

	slugs := extractReleaseNoteSlugs(pageHTML)
	if *maxEntries > 0 && len(slugs) > *maxEntries {
		slugs = slugs[:*maxEntries]
	}

	newEntries := make([]entry, 0, len(slugs))
	for _, slug := range slugs {
		if _, ok := knownSlugs[slug]; ok {
			continue
		}

		note := entry{
			Slug: slug,
			URL:  releaseNoteURL(slug),
		}
		if title, err := fetchReleaseNoteTitle(client, note.URL); err == nil {
			note.Title = title
		}

		if strings.TrimSpace(*githubRepo) != "" && strings.TrimSpace(*githubToken) != "" {
			tracked, err := hasTrackedIssue(client, *githubRepo, *githubToken, note.URL)
			if err != nil {
				fatalf("기존 이슈 조회 실패: %v", err)
			}
			if tracked {
				continue
			}
		}

		newEntries = append(newEntries, note)
	}

	reportDate := time.Now().UTC().Format("2006-01-02")
	rep := report{
		CheckedAt:          time.Now().UTC().Format(time.RFC3339),
		PageURL:            *pageURL,
		NewEntries:         newEntries,
		HasIssue:           len(newEntries) > 0,
		ShouldFailWorkflow: false,
	}
	if rep.HasIssue {
		rep.IssueTitle = fmt.Sprintf("[API Change] %d new release note(s) detected (%s)", len(newEntries), reportDate)
		rep.IssueLabels = []string{"api-monitor"}
	}

	if err := writeReportJSON(*outputJSON, rep); err != nil {
		fatalf("JSON 리포트 저장 실패: %v", err)
	}
	if err := writeMarkdown(*outputMD, renderIssueMarkdown(*baselinePath, rep)); err != nil {
		fatalf("Markdown 리포트 저장 실패: %v", err)
	}

	sort.Slice(rep.NewEntries, func(i, j int) bool {
		return rep.NewEntries[i].Slug < rep.NewEntries[j].Slug
	})
	fmt.Printf("new release notes: %d\n", len(rep.NewEntries))
}

func loadBaseline(path string) (map[string]struct{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var b baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	result := make(map[string]struct{}, len(b.KnownSlugs))
	for _, slug := range b.KnownSlugs {
		slug = strings.TrimSpace(slug)
		if slug == "" {
			continue
		}
		result[slug] = struct{}{}
	}
	return result, nil
}

func extractReleaseNoteSlugs(pageHTML string) []string {
	matches := releaseNoteSlugRE.FindAllStringSubmatch(pageHTML, -1)
	seen := make(map[string]struct{}, len(matches))
	slugs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		slug := strings.TrimSpace(match[1])
		switch {
		case slug == "", slug == "release-notes", strings.HasPrefix(slug, "page"):
			continue
		}
		if _, ok := seen[slug]; ok {
			continue
		}
		seen[slug] = struct{}{}
		slugs = append(slugs, slug)
	}
	return slugs
}

func releaseNoteURL(slug string) string {
	return "https://naver.worksmobile.com/release-notes/" + slug + "/"
}

func fetchReleaseNoteTitle(client *http.Client, pageURL string) (string, error) {
	body, err := fetchText(client, pageURL)
	if err != nil {
		return "", err
	}
	if match := ogTitleRE.FindStringSubmatch(body); len(match) > 1 {
		return normalizeReleaseNoteTitle(match[1]), nil
	}
	if match := titleTagRE.FindStringSubmatch(body); len(match) > 1 {
		return normalizeReleaseNoteTitle(match[1]), nil
	}
	return "", fmt.Errorf("title not found")
}

func normalizeReleaseNoteTitle(raw string) string {
	title := html.UnescapeString(strings.TrimSpace(raw))
	title = strings.TrimSuffix(title, " - 네이버웍스")
	return strings.TrimSpace(title)
}

func fetchText(client *http.Client, pageURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "naverworks-cli-api-monitor/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func hasTrackedIssue(client *http.Client, repo, token, noteURL string) (bool, error) {
	query := fmt.Sprintf(`repo:%s is:issue "%s"`, repo, noteURL)
	endpoint := "https://api.github.com/search/issues?q=" + url.QueryEscape(query) + "&per_page=5"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "naverworks-cli-api-monitor/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return false, fmt.Errorf("GitHub search failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload issueSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return false, err
	}
	return len(payload.Items) > 0, nil
}

func renderIssueMarkdown(baselinePath string, rep report) string {
	var b strings.Builder
	b.WriteString("## API 변경 감지\n\n")
	b.WriteString(fmt.Sprintf("- 검사 시각: `%s`\n", rep.CheckedAt))
	b.WriteString(fmt.Sprintf("- 소스 페이지: %s\n", rep.PageURL))
	b.WriteString(fmt.Sprintf("- baseline: `%s`\n\n", baselinePath))

	if len(rep.NewEntries) == 0 {
		b.WriteString("새로 감지된 릴리즈 노트가 없습니다.\n")
		return b.String()
	}

	b.WriteString("다음 릴리즈 노트가 baseline에 없고, 기존 GitHub 이슈에도 아직 추적되지 않았습니다.\n\n")
	for i, note := range rep.NewEntries {
		title := note.Title
		if title == "" {
			title = note.Slug
		}
		b.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, title))
		b.WriteString(fmt.Sprintf("   - URL: %s\n", note.URL))
		b.WriteString(fmt.Sprintf("   - slug: `%s`\n", note.Slug))
	}

	b.WriteString("\n## 권장 조치\n\n")
	b.WriteString("- 변경 내용을 확인하고 필요한 CLI 반영 범위를 판별합니다.\n")
	b.WriteString("- 처리 완료 후 baseline 파일을 갱신해 같은 공지가 반복 감지되지 않도록 합니다.\n")
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
