package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type issue struct {
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
}

type searchResponse struct {
	Items []issue `json:"items"`
}

func main() {
	var (
		repo     = flag.String("repo", os.Getenv("GITHUB_REPOSITORY"), "GitHub repository in owner/repo form")
		token    = flag.String("token", "", "GitHub token (기본: GITHUB_TOKEN)")
		title    = flag.String("title", "", "issue title")
		bodyFile = flag.String("body-file", "", "path to issue body markdown")
		labels   = flag.String("labels", "", "comma-separated labels")
	)
	flag.Parse()

	resolvedToken := strings.TrimSpace(*token)
	if resolvedToken == "" {
		resolvedToken = strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	}

	if strings.TrimSpace(*repo) == "" || resolvedToken == "" {
		fatalf("--repo와 GITHUB_TOKEN(또는 --token)이 필요합니다")
	}
	if strings.TrimSpace(*title) == "" {
		fatalf("--title이 필요합니다")
	}
	if strings.TrimSpace(*bodyFile) == "" {
		fatalf("--body-file이 필요합니다")
	}

	body, err := os.ReadFile(*bodyFile)
	if err != nil {
		fatalf("이슈 본문 읽기 실패: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	existing, err := findOpenIssueByExactTitle(client, *repo, resolvedToken, *title)
	if err != nil {
		fatalf("기존 이슈 조회 실패: %v", err)
	}
	if existing != nil {
		fmt.Printf("existing %s\n", existing.HTMLURL)
		return
	}

	created, err := createIssue(client, *repo, resolvedToken, *title, string(body), splitLabels(*labels))
	if err != nil {
		fatalf("이슈 생성 실패: %v", err)
	}
	fmt.Printf("created %s\n", created.HTMLURL)
}

func findOpenIssueByExactTitle(client *http.Client, repo, token, title string) (*issue, error) {
	query := fmt.Sprintf(`repo:%s is:issue is:open "%s"`, repo, title)
	endpoint := "https://api.github.com/search/issues?q=" + url.QueryEscape(query) + "&per_page=10"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	setGitHubHeaders(req, token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	for _, candidate := range payload.Items {
		if candidate.Title == title {
			return &candidate, nil
		}
	}
	return nil, nil
}

func createIssue(client *http.Client, repo, token, title, body string, labels []string) (*issue, error) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo: %s", repo)
	}

	payload := struct {
		Title  string   `json:"title"`
		Body   string   `json:"body"`
		Labels []string `json:"labels,omitempty"`
	}{
		Title:  title,
		Body:   body,
		Labels: labels,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", parts[0], parts[1])
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	setGitHubHeaders(req, token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var created issue
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, err
	}
	return &created, nil
}

func splitLabels(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		labels = append(labels, part)
	}
	return labels
}

func setGitHubHeaders(req *http.Request, token string) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "naverworks-cli-github-issue/1.0")
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
