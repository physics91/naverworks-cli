//go:build !windows

package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
)

// ──────────────────────────────────────────────────────────────────────
// Test 1: Atomic config write via CLI
// ──────────────────────────────────────────────────────────────────────

func TestE2E_ConfigSave_AtomicWrite(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)

	// Use CLI to modify config
	_, err := runCLI(t, "config", "set", "bot_id", "new-bot-id")
	if err != nil {
		t.Fatalf("config set failed: %v", err)
	}

	// Read back the config file and verify the value was persisted
	cfgPath := filepath.Join(tmpDir, ".config", "naverworks", "config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// The config is now in profile format after save
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("config file is not valid JSON: %v\ncontent: %s", err, string(data))
	}

	// Verify bot_id was updated — check in profile format
	if profiles, ok := raw["profiles"]; ok {
		var profs map[string]map[string]interface{}
		if err := json.Unmarshal(profiles, &profs); err != nil {
			t.Fatalf("failed to parse profiles: %v", err)
		}
		found := false
		for _, prof := range profs {
			if botID, ok := prof["bot_id"]; ok && botID == "new-bot-id" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("bot_id not updated to 'new-bot-id' in any profile\ncontent: %s", string(data))
		}
	} else {
		// Legacy flat format fallback check
		var flat map[string]interface{}
		json.Unmarshal(data, &flat)
		if flat["bot_id"] != "new-bot-id" {
			t.Errorf("bot_id not updated to 'new-bot-id'\ncontent: %s", string(data))
		}
	}

	// Verify no leftover temp files in config directory
	cfgDir := filepath.Dir(cfgPath)
	entries, err := os.ReadDir(cfgDir)
	if err != nil {
		t.Fatalf("failed to read config dir: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".naverworks-") && strings.HasSuffix(entry.Name(), ".tmp") {
			t.Errorf("leftover temp file found: %s", entry.Name())
		}
	}

	// Verify file permissions are 0600 (non-Windows)
	info, err := os.Stat(cfgPath)
	if err != nil {
		t.Fatalf("failed to stat config file: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected config file permissions 0600, got %04o", perm)
	}

	// Verify directory permissions are 0700
	dirInfo, err := os.Stat(cfgDir)
	if err != nil {
		t.Fatalf("failed to stat config dir: %v", err)
	}
	dirPerm := dirInfo.Mode().Perm()
	if dirPerm != 0700 {
		t.Errorf("expected config dir permissions 0700, got %04o", dirPerm)
	}
}

// ──────────────────────────────────────────────────────────────────────
// Test 2: Concurrent atomic writes don't corrupt config
// ──────────────────────────────────────────────────────────────────────

func TestE2E_ConfigSave_AtomicWrite_Concurrent(t *testing.T) {
	tmpDir := setupTestEnv(t)
	writeTestConfig(t, tmpDir)

	const iterations = 10
	var wg sync.WaitGroup
	errs := make([]error, iterations)

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			value := fmt.Sprintf("bot-%d", idx)
			_, err := runCLI(t, "config", "set", "bot_id", value)
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	// Check that no goroutine returned a fatal error
	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d returned error: %v", i, err)
		}
	}

	// Verify the config file is valid JSON (not corrupted)
	cfgPath := filepath.Join(tmpDir, ".config", "naverworks", "config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("failed to read config after concurrent writes: %v", err)
	}
	if !json.Valid(data) {
		t.Fatalf("config file is corrupt after concurrent writes:\n%s", string(data))
	}

	// Verify it can be parsed back
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("config file not parseable after concurrent writes: %v", err)
	}

	// Verify no leftover temp files
	cfgDir := filepath.Dir(cfgPath)
	entries, err := os.ReadDir(cfgDir)
	if err != nil {
		t.Fatalf("failed to read config dir: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".naverworks-") && strings.HasSuffix(entry.Name(), ".tmp") {
			t.Errorf("leftover temp file after concurrent writes: %s", entry.Name())
		}
	}
}

// ──────────────────────────────────────────────────────────────────────
// Test 3: Oversized API response rejection via service layer
// ──────────────────────────────────────────────────────────────────────

func TestE2E_API_OversizedResponse_ViaBot(t *testing.T) {
	// Create a mock server that returns a >10MB response body
	const maxAPIResponseSize = 10 << 20 // 10MB — mirrors internal/api/client.go
	oversizedBody := strings.Repeat("x", maxAPIResponseSize+1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(oversizedBody))
	}))
	defer server.Close()

	// Create a Client pointing at the mock server
	token := &auth.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	client := api.NewClient(server.URL, token, nil)

	// Test through BotService.SendTextToUser — a real service method
	botSvc := api.NewBotService(client)
	_, err := botSvc.SendTextToUser("bot1", "user1", "hello")
	if err == nil {
		t.Fatal("expected error for oversized API response via BotService")
	}
	expectedMsg := fmt.Sprintf("API 응답 크기 초과: > %d bytes", maxAPIResponseSize)
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestE2E_API_OversizedResponse_ViaDirectory(t *testing.T) {
	const maxAPIResponseSize = 10 << 20
	oversizedBody := strings.Repeat("x", maxAPIResponseSize+1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(oversizedBody))
	}))
	defer server.Close()

	token := &auth.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	client := api.NewClient(server.URL, token, nil)

	// Test through DirectoryService.ListUsers — a real GET-based service method
	dirSvc := api.NewDirectoryService(client)
	_, err := dirSvc.ListUsers("", 100)
	if err == nil {
		t.Fatal("expected error for oversized API response via DirectoryService")
	}
	expectedMsg := fmt.Sprintf("API 응답 크기 초과: > %d bytes", maxAPIResponseSize)
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestE2E_API_NormalResponse_Allowed(t *testing.T) {
	// Verify that normal-sized responses still pass through
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"users":[]}`))
	}))
	defer server.Close()

	token := &auth.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	client := api.NewClient(server.URL, token, nil)

	dirSvc := api.NewDirectoryService(client)
	resp, err := dirSvc.ListUsers("", 100)
	if err != nil {
		t.Fatalf("unexpected error for normal response: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ──────────────────────────────────────────────────────────────────────
// Test 4: GitHub Actions workflow files use SHA-pinned actions
// ──────────────────────────────────────────────────────────────────────

func TestE2E_WorkflowFiles_SHA_Pinned(t *testing.T) {
	workflowDir := filepath.Join("..", ".github", "workflows")

	// Verify the workflow directory exists
	if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
		t.Skip("workflow directory not found (running outside repo root?)")
	}

	expectedFiles := []string{"ci.yml", "release.yml"}
	usesLineRe := regexp.MustCompile(`^\s*-?\s*uses:\s*(.+)$`)
	shaRe := regexp.MustCompile(`@[0-9a-f]{40}\b`)
	versionCommentRe := regexp.MustCompile(`#\s*v\d+`)

	for _, filename := range expectedFiles {
		t.Run(filename, func(t *testing.T) {
			path := filepath.Join(workflowDir, filename)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read %s: %v", filename, err)
			}

			lines := strings.Split(string(data), "\n")
			usesCount := 0

			for lineNum, line := range lines {
				matches := usesLineRe.FindStringSubmatch(line)
				if matches == nil {
					continue
				}
				usesCount++
				actionRef := matches[1]

				// Verify SHA pinning: must contain @<40-hex-chars>
				if !shaRe.MatchString(actionRef) {
					t.Errorf("%s:%d: action not SHA-pinned: %s", filename, lineNum+1, strings.TrimSpace(line))
				}

				// Verify version comment: must contain # v<digits>
				if !versionCommentRe.MatchString(actionRef) {
					t.Errorf("%s:%d: missing version comment (# vN): %s", filename, lineNum+1, strings.TrimSpace(line))
				}
			}

			if usesCount == 0 {
				t.Errorf("%s: no 'uses:' lines found", filename)
			}
			t.Logf("%s: verified %d SHA-pinned actions", filename, usesCount)
		})
	}
}

func TestE2E_WorkflowFiles_SHA_NoDuplicates(t *testing.T) {
	// Additional: verify all workflow files in the directory are checked
	workflowDir := filepath.Join("..", ".github", "workflows")
	if _, err := os.Stat(workflowDir); os.IsNotExist(err) {
		t.Skip("workflow directory not found")
	}

	err := filepath.WalkDir(workflowDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".yml") && !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read %s: %v", path, err)
			return nil
		}

		usesRe := regexp.MustCompile(`uses:\s*(\S+)`)
		shaRe := regexp.MustCompile(`@[0-9a-f]{40}\b`)

		for lineNum, line := range strings.Split(string(data), "\n") {
			matches := usesRe.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			actionRef := matches[1]
			if !shaRe.MatchString(actionRef) {
				t.Errorf("%s:%d: action not SHA-pinned: %s", d.Name(), lineNum+1, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk workflow dir: %v", err)
	}
}
