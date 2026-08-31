package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type issue struct {
	Number  int          `json:"number"`
	Title   string       `json:"title"`
	State   string       `json:"state"`
	Labels  []issueLabel `json:"labels"`
	HTMLURL string       `json:"html_url"`
	// GitHub's repository issues endpoint also returns pull requests.
	PullRequest json.RawMessage `json:"pull_request,omitempty"`
}

type issueLabel struct {
	Name string `json:"name"`
}

func main() {
	var (
		repo     = flag.String("repo", os.Getenv("GITHUB_REPOSITORY"), "GitHub repository in owner/repo form")
		token    = flag.String("token", "", "GitHub token (기본: GITHUB_TOKEN)")
		title    = flag.String("title", "", "issue title")
		bodyFile = flag.String("body-file", "", "path to issue body markdown")
		labels   = flag.String("labels", "", "comma-separated labels")
		state    = flag.String("state", "open", "desired issue state: open or closed")
	)
	flag.Parse()

	resolvedToken := strings.TrimSpace(*token)
	if resolvedToken == "" {
		resolvedToken = strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	}

	if strings.TrimSpace(*repo) == "" || resolvedToken == "" {
		fatalf("--repo와 GITHUB_TOKEN(또는 --token)이 필요합니다")
	}
	if _, _, err := parseRepository(*repo); err != nil {
		fatalf("저장소 형식 오류: %v", err)
	}
	if strings.TrimSpace(*title) == "" {
		fatalf("--title이 필요합니다")
	}
	desiredState := strings.ToLower(strings.TrimSpace(*state))
	if desiredState != "open" && desiredState != "closed" {
		fatalf("--state는 open 또는 closed여야 합니다")
	}
	expectedLabels, err := normalizeLabels(splitLabels(*labels))
	if err != nil {
		fatalf("--labels 오류: %v", err)
	}
	if len(expectedLabels) == 0 {
		fatalf("자동화 소유권 확인을 위해 --labels가 필요합니다")
	}

	var body string
	if desiredState == "open" {
		if strings.TrimSpace(*bodyFile) == "" {
			fatalf("--state=open에서는 --body-file이 필요합니다")
		}
		rawBody, err := os.ReadFile(*bodyFile)
		if err != nil {
			fatalf("이슈 본문 읽기 실패: %v", err)
		}
		body = string(rawBody)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	action, reconciled, err := reconcileIssue(
		client,
		*repo,
		resolvedToken,
		*title,
		body,
		expectedLabels,
		desiredState,
	)
	if err != nil {
		fatalf("이슈 reconcile 실패: %v", err)
	}
	if reconciled == nil {
		fmt.Printf("%s: 일치하는 열린 이슈 없음\n", action)
		return
	}
	fmt.Printf("%s %s\n", action, reconciled.HTMLURL)
}

func reconcileIssue(client *http.Client, repo, token, title, body string, labels []string, desiredState string) (string, *issue, error) {
	existing, err := findOwnedOpenIssue(client, repo, token, title, labels)
	if err != nil {
		return "", nil, err
	}

	switch desiredState {
	case "open":
		if existing == nil {
			created, err := createIssue(client, repo, token, title, body, labels)
			return "created", created, err
		}
		updated, err := updateIssueBody(client, repo, token, existing.Number, body)
		return "updated", updated, err
	case "closed":
		if existing == nil {
			return "unchanged", nil, nil
		}
		closed, err := updateIssueState(client, repo, token, existing.Number, "closed")
		return "closed", closed, err
	default:
		return "", nil, fmt.Errorf("invalid desired state: %s", desiredState)
	}
}

func findOwnedOpenIssue(client *http.Client, repo, token, title string, expectedLabels []string) (*issue, error) {
	owner, name, err := parseRepository(repo)
	if err != nil {
		return nil, err
	}
	expectedLabels, err = normalizeLabels(expectedLabels)
	if err != nil {
		return nil, err
	}
	if len(expectedLabels) == 0 {
		return nil, fmt.Errorf("at least one ownership label is required")
	}

	exactMatches := make([]issue, 0, 1)
	for page := 1; ; page++ {
		endpoint := fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/issues?state=open&per_page=100&page=%d",
			owner,
			name,
			page,
		)
		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		setGitHubHeaders(req, token)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var candidates []issue
		decodeErr := json.NewDecoder(resp.Body).Decode(&candidates)
		closeErr := resp.Body.Close()
		if decodeErr != nil {
			return nil, decodeErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		for _, candidate := range candidates {
			if len(candidate.PullRequest) != 0 {
				continue
			}
			if candidate.Title == title && strings.EqualFold(candidate.State, "open") {
				exactMatches = append(exactMatches, candidate)
			}
		}
		if len(candidates) < 100 {
			break
		}
	}
	if len(exactMatches) == 0 {
		return nil, nil
	}
	if len(exactMatches) > 1 {
		return nil, fmt.Errorf("ambiguous automation target: %d open issues have the exact title", len(exactMatches))
	}
	actualLabels := make([]string, 0, len(exactMatches[0].Labels))
	for _, label := range exactMatches[0].Labels {
		actualLabels = append(actualLabels, label.Name)
	}
	actualLabels, err = normalizeLabels(actualLabels)
	if err != nil {
		return nil, fmt.Errorf("invalid labels on issue #%d: %w", exactMatches[0].Number, err)
	}
	if !equalStrings(actualLabels, expectedLabels) {
		return nil, fmt.Errorf(
			"automation ownership mismatch on issue #%d: expected labels %v, observed %v",
			exactMatches[0].Number,
			expectedLabels,
			actualLabels,
		)
	}
	return &exactMatches[0], nil
}

func createIssue(client *http.Client, repo, token, title, body string, labels []string) (*issue, error) {
	owner, name, err := parseRepository(repo)
	if err != nil {
		return nil, err
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

	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, name)
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

func updateIssueBody(client *http.Client, repo, token string, number int, body string) (*issue, error) {
	return updateIssue(client, repo, token, number, &body, "")
}

func updateIssueState(client *http.Client, repo, token string, number int, state string) (*issue, error) {
	if state != "open" && state != "closed" {
		return nil, fmt.Errorf("invalid issue state: %s", state)
	}
	return updateIssue(client, repo, token, number, nil, state)
}

func updateIssue(client *http.Client, repo, token string, number int, body *string, state string) (*issue, error) {
	if number <= 0 {
		return nil, fmt.Errorf("invalid issue number: %d", number)
	}
	owner, name, err := parseRepository(repo)
	if err != nil {
		return nil, err
	}

	payload := struct {
		Body  *string `json:"body,omitempty"`
		State string  `json:"state,omitempty"`
	}{Body: body, State: state}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", owner, name, number)
	req, err := http.NewRequest(http.MethodPatch, endpoint, bytes.NewReader(data))
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
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var updated issue
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		return nil, err
	}
	return &updated, nil
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

var repositoryPartRE = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

func parseRepository(repo string) (string, string, error) {
	repo = strings.TrimSpace(repo)
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo: %s", repo)
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." || !repositoryPartRE.MatchString(part) {
			return "", "", fmt.Errorf("invalid repo: %s", repo)
		}
	}
	return parts[0], parts[1], nil
}

func normalizeLabels(labels []string) ([]string, error) {
	seen := make(map[string]struct{}, len(labels))
	normalized := make([]string, 0, len(labels))
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		key := strings.ToLower(label)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("duplicate label: %s", label)
		}
		seen[key] = struct{}{}
		normalized = append(normalized, key)
	}
	sort.Strings(normalized)
	return normalized, nil
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
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
