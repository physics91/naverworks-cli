# naverworks-cli 구현 계획

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 네이버웍스 REST API v1.0을 래핑하는 Go CLI 도구(nw-cli)를 구현한다. v0.1은 bot, directory, calendar 3개 서비스를 지원한다.

**Architecture:** cobra 기반 서브커맨드 구조. `internal/config`로 설정 관리, `internal/auth`로 JWT/OAuth 토큰 관리, `internal/api`로 HTTP 클라이언트(401 재시도 1회, 429 백오프 최대 3회), `internal/output`으로 JSON/테이블 출력. 각 서비스 커맨드가 `loadConfigAndToken()` + `buildAPIClient()`를 호출하는 per-command 초기화 방식을 사용한다 (PersistentPreRunE 대신 — auth/config/version 건너뛰기 복잡도 회피).

**에러 출력 계약:** `main.go`에서 `rootCmd.Execute()` 에러를 받아 `APIError` 타입이면 `{"error":{"code":"...","description":"..."}}` JSON을 stderr에 출력, 그 외 에러는 `{"error":{"code":"CLI_ERROR","description":"..."}}` 형식으로 stderr에 출력 후 exit 1.

**HTTP 재시도 계약:** POST 등 body가 있는 요청은 `[]byte`로 버퍼링하여 재시도 시 재사용. 401 재시도는 요청당 최대 1회. 429 재시도는 `RateLimit-Reset`(초) 대기 후 최대 3회, 헤더 없으면 지수 백오프(1초, 2초, 4초).

**JWT 갱신 계약:** JWT 토큰 만료 60초 전 또는 401 시, refresh_token이 있으면 refresh 시도, 없거나 실패하면 assertion으로 재발급.

**Windows 보안:** config.json/token.json 저장 시 Windows에서는 `icacls`로 현재 사용자 전용 ACL 설정. private_key 권한 검사도 Windows ACE 확인 포함.

**Tech Stack:** Go 1.22+, cobra, crypto/rsa (JWT), net/http, encoding/json

**Requirements:** `docs/requirements/2026-03-20-naverworks-cli.md`

---

## Task 1: Go 모듈 초기화 + cobra 스켈레톤

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `cmd/version.go`

**Step 1: Go 모듈 초기화**

Run: `go mod init github.com/physics91/naverworks-cli`
Expected: `go.mod` 생성

**Step 2: cobra 의존성 추가**

Run: `go get github.com/spf13/cobra@latest`

**Step 3: main.go 작성**

```go
// main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/physics91/naverworks-cli/cmd"
	"github.com/physics91/naverworks-cli/internal/api"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// NFR-04: 에러를 JSON 형식으로 stderr에 출력
		errObj := map[string]map[string]string{
			"error": {"code": "CLI_ERROR", "description": err.Error()},
		}
		if apiErr, ok := err.(*api.APIError); ok {
			errObj["error"]["code"] = apiErr.Code
			errObj["error"]["description"] = apiErr.Description
		}
		data, _ := json.Marshal(errObj)
		fmt.Fprintln(os.Stderr, string(data))
		os.Exit(1)
	}
}
```

**Step 4: cmd/root.go 작성**

```go
// cmd/root.go
package cmd

import (
	"github.com/spf13/cobra"
)

var outputFormat string

var rootCmd = &cobra.Command{
	Use:   "nw-cli",
	Short: "네이버웍스 CLI",
	Long:  "네이버웍스 REST API v1.0 명령줄 도구",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "json", "출력 형식 (json|table)")
}

func Execute() error {
	return rootCmd.Execute()
}
```

**Step 5: cmd/version.go 작성**

```go
// cmd/version.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "버전 정보 출력",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(`{"version":"%s","commit":"%s","build_date":"%s"}`+"\n", version, commit, buildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
```

**Step 6: 빌드 및 실행 테스트**

Run: `go build -o nw-cli . && ./nw-cli version`
Expected: `{"version":"dev","commit":"none","build_date":"unknown"}`

Run: `./nw-cli --help`
Expected: cobra 기본 도움말 출력

**Step 7: Makefile 작성**

```makefile
# Makefile
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X github.com/physics91/naverworks-cli/cmd.version=$(VERSION) \
	-X github.com/physics91/naverworks-cli/cmd.commit=$(COMMIT) \
	-X github.com/physics91/naverworks-cli/cmd.buildDate=$(BUILD_DATE)

.PHONY: build test clean

build:
	go build -ldflags "$(LDFLAGS)" -o nw-cli .

test:
	go test ./... -v

clean:
	rm -f nw-cli
```

**Step 8: 커밋**

```bash
git add go.mod go.sum main.go cmd/root.go cmd/version.go Makefile
git commit -m "feat: Go 모듈 초기화 및 cobra 스켈레톤 구성"
```

---

## Task 2: 설정 관리 (`internal/config`)

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `cmd/config_cmd.go`

**Step 1: 실패하는 테스트 작성**

```go
// internal/config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ClientID != "" {
		t.Errorf("expected empty client_id, got %q", cfg.ClientID)
	}
}

func TestSetAndGet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg, _ := Load(path)
	if err := cfg.Set("client_id", "test-id"); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	cfg2, _ := Load(path)
	if cfg2.ClientID != "test-id" {
		t.Errorf("expected test-id, got %q", cfg2.ClientID)
	}
}

func TestEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg, _ := Load(path)
	cfg.Set("client_id", "from-file")

	os.Setenv("NW_CLIENT_ID", "from-env")
	defer os.Unsetenv("NW_CLIENT_ID")

	cfg.ApplyEnvOverrides()
	if cfg.ClientID != "from-env" {
		t.Errorf("expected from-env, got %q", cfg.ClientID)
	}
}

func TestGetMasked(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg, _ := Load(path)
	cfg.Set("client_secret", "super-secret")
	cfg.Save(path)

	val := cfg.GetMasked("client_secret")
	if val != "****" {
		t.Errorf("expected ****, got %q", val)
	}
}

func TestSetInvalidKey(t *testing.T) {
	cfg := &Config{}
	err := cfg.Set("invalid_key", "value")
	if err == nil {
		t.Error("expected error for invalid key")
	}
}
```

**Step 2: 테스트 실패 확인**

Run: `go test ./internal/config/ -v`
Expected: FAIL (컴파일 에러)

**Step 3: 구현**

```go
// internal/config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	ClientID               string `json:"client_id,omitempty"`
	ClientSecret           string `json:"client_secret,omitempty"`
	ServiceAccountID       string `json:"service_account_id,omitempty"`
	PrivateKeyPath         string `json:"private_key_path,omitempty"`
	DomainID               string `json:"domain_id,omitempty"`
	BotID                  string `json:"bot_id,omitempty"`
	Scope                  string `json:"scope,omitempty"`
	DefaultCalendarUserID  string `json:"default_calendar_user_id,omitempty"`
}

var validKeys = map[string]bool{
	"client_id": true, "client_secret": true, "service_account_id": true,
	"private_key_path": true, "domain_id": true, "bot_id": true,
	"scope": true, "default_calendar_user_id": true,
}

var sensitiveKeys = map[string]bool{
	"client_secret": true,
}

var envMap = map[string]string{
	"client_id":                "NW_CLIENT_ID",
	"client_secret":            "NW_CLIENT_SECRET",
	"service_account_id":       "NW_SERVICE_ACCOUNT_ID",
	"private_key_path":         "NW_PRIVATE_KEY_PATH",
	"domain_id":                "NW_DOMAIN_ID",
	"bot_id":                   "NW_BOT_ID",
	"scope":                    "NW_SCOPE",
	"default_calendar_user_id": "NW_DEFAULT_CALENDAR_USER_ID",
}

func DefaultPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "naverworks", "config.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "naverworks", "config.json")
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("config 읽기 실패: %w", err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config 파싱 실패: %w", err)
	}
	return cfg, nil
}

func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("config 직렬화 실패: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) Set(key, value string) error {
	if !validKeys[key] {
		return fmt.Errorf("유효하지 않은 설정 키: %s", key)
	}
	switch key {
	case "client_id":
		c.ClientID = value
	case "client_secret":
		c.ClientSecret = value
	case "service_account_id":
		c.ServiceAccountID = value
	case "private_key_path":
		c.PrivateKeyPath = value
	case "domain_id":
		c.DomainID = value
	case "bot_id":
		c.BotID = value
	case "scope":
		c.Scope = value
	case "default_calendar_user_id":
		c.DefaultCalendarUserID = value
	}
	return nil
}

func (c *Config) Get(key string) (string, error) {
	if !validKeys[key] {
		return "", fmt.Errorf("유효하지 않은 설정 키: %s", key)
	}
	switch key {
	case "client_id":
		return c.ClientID, nil
	case "client_secret":
		return c.ClientSecret, nil
	case "service_account_id":
		return c.ServiceAccountID, nil
	case "private_key_path":
		return c.PrivateKeyPath, nil
	case "domain_id":
		return c.DomainID, nil
	case "bot_id":
		return c.BotID, nil
	case "scope":
		return c.Scope, nil
	case "default_calendar_user_id":
		return c.DefaultCalendarUserID, nil
	}
	return "", nil
}

func (c *Config) GetMasked(key string) string {
	val, _ := c.Get(key)
	if sensitiveKeys[key] && val != "" {
		return "****"
	}
	return val
}

func (c *Config) ApplyEnvOverrides() {
	for key, envVar := range envMap {
		if val := os.Getenv(envVar); val != "" {
			c.Set(key, val)
		}
	}
}
```

**Step 4: 테스트 통과 확인**

Run: `go test ./internal/config/ -v`
Expected: PASS

**Step 5: config 커맨드 작성**

```go
// cmd/config_cmd.go
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "설정 관리",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> [value]",
	Short: "설정값 저장",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}

		key := args[0]
		var value string
		useStdin, _ := cmd.Flags().GetBool("stdin")
		if useStdin {
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				value = scanner.Text()
			}
		} else if len(args) == 2 {
			value = args[1]
		} else {
			return fmt.Errorf("값을 지정하세요: nw-cli config set %s <value>", key)
		}

		if err := cfg.Set(key, value); err != nil {
			return err
		}
		return cfg.Save(path)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "설정값 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		cfg.ApplyEnvOverrides()
		fmt.Println(cfg.GetMasked(args[0]))
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "전체 설정 목록",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		cfg.ApplyEnvOverrides()

		masked := map[string]string{
			"client_id":                cfg.GetMasked("client_id"),
			"client_secret":            cfg.GetMasked("client_secret"),
			"service_account_id":       cfg.GetMasked("service_account_id"),
			"private_key_path":         cfg.GetMasked("private_key_path"),
			"domain_id":               cfg.GetMasked("domain_id"),
			"bot_id":                  cfg.GetMasked("bot_id"),
			"scope":                   cfg.GetMasked("scope"),
			"default_calendar_user_id": cfg.GetMasked("default_calendar_user_id"),
		}
		data, _ := json.MarshalIndent(masked, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	configSetCmd.Flags().Bool("stdin", false, "stdin에서 값 읽기")
	configCmd.AddCommand(configSetCmd, configGetCmd, configListCmd)
	rootCmd.AddCommand(configCmd)
}
```

**Step 6: 빌드 및 통합 테스트**

Run: `go build -o nw-cli . && ./nw-cli config set client_id test123 && ./nw-cli config get client_id`
Expected: `test123`

**Step 7: 커밋**

```bash
git add internal/config/ cmd/config_cmd.go
git commit -m "feat: 설정 관리 기능 구현 (config set/get/list)"
```

---

## Task 3: 토큰 저장소 (`internal/auth/token.go`)

**Files:**
- Create: `internal/auth/token.go`
- Create: `internal/auth/token_test.go`

**Step 1: 실패하는 테스트 작성**

```go
// internal/auth/token_test.go
package auth

import (
	"path/filepath"
	"testing"
	"time"
)

func TestTokenStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	store := NewTokenStore(path)

	token := &Token{
		AuthMethod:   "oauth",
		AccessToken:  "at-123",
		RefreshToken: "rt-456",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scope:        "bot directory calendar",
	}
	if err := store.Save(token); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.AccessToken != "at-123" {
		t.Errorf("expected at-123, got %q", loaded.AccessToken)
	}
	if loaded.AuthMethod != "oauth" {
		t.Errorf("expected oauth, got %q", loaded.AuthMethod)
	}
}

func TestTokenStore_LoadNotExist(t *testing.T) {
	store := NewTokenStore("/nonexistent/token.json")
	token, err := store.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != nil {
		t.Error("expected nil token")
	}
}

func TestToken_IsExpired(t *testing.T) {
	token := &Token{ExpiresAt: time.Now().Add(-1 * time.Minute)}
	if !token.IsExpired() {
		t.Error("expected expired")
	}

	token2 := &Token{ExpiresAt: time.Now().Add(1 * time.Hour)}
	if token2.IsExpired() {
		t.Error("expected not expired")
	}
}

func TestToken_NeedsRefresh(t *testing.T) {
	token := &Token{ExpiresAt: time.Now().Add(30 * time.Second)}
	if !token.NeedsRefresh() {
		t.Error("expected needs refresh (within 60s buffer)")
	}

	token2 := &Token{ExpiresAt: time.Now().Add(5 * time.Minute)}
	if token2.NeedsRefresh() {
		t.Error("expected no refresh needed")
	}
}

func TestTokenStore_Delete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	store := NewTokenStore(path)

	store.Save(&Token{AccessToken: "test"})
	if err := store.Delete(); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	token, _ := store.Load()
	if token != nil {
		t.Error("expected nil after delete")
	}
}
```

**Step 2: 테스트 실패 확인**

Run: `go test ./internal/auth/ -v`
Expected: FAIL

**Step 3: 구현**

```go
// internal/auth/token.go
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const refreshBuffer = 60 * time.Second

type Token struct {
	AuthMethod       string    `json:"auth_method"`
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token,omitempty"`
	TokenType        string    `json:"token_type"`
	ExpiresAt        time.Time `json:"expires_at"`
	Scope            string    `json:"scope"`
	ServiceAccountID string    `json:"service_account_id,omitempty"`
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *Token) NeedsRefresh() bool {
	return time.Now().After(t.ExpiresAt.Add(-refreshBuffer))
}

type TokenStore struct {
	path string
}

func DefaultTokenPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "naverworks", "token.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "naverworks", "token.json")
}

func NewTokenStore(path string) *TokenStore {
	return &TokenStore{path: path}
}

func (s *TokenStore) Load() (*Token, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("토큰 읽기 실패: %w", err)
	}
	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("토큰 파싱 실패: %w", err)
	}
	return &token, nil
}

func (s *TokenStore) Save(token *Token) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("토큰 직렬화 실패: %w", err)
	}
	return os.WriteFile(s.path, data, 0600)
}

func (s *TokenStore) Delete() error {
	err := os.Remove(s.path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("토큰 삭제 실패: %w", err)
	}
	return nil
}
```

**Step 4: 테스트 통과 확인**

Run: `go test ./internal/auth/ -v`
Expected: PASS

**Step 5: 커밋**

```bash
git add internal/auth/
git commit -m "feat: 토큰 저장소 구현 (저장/로드/삭제/만료 판정)"
```

---

## Task 4: HTTP 클라이언트 (`internal/api/client.go`)

**Files:**
- Create: `internal/api/client.go`
- Create: `internal/api/client_test.go`

**Step 1: 실패하는 테스트 작성**

```go
// internal/api/client_test.go
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestClient_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	token := &auth.Token{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	client := NewClient(server.URL, token, nil)
	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClient_401_Retry(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count == 1 {
			w.WriteHeader(401)
			w.Write([]byte(`{"code":"UNAUTHORIZED"}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	token := &auth.Token{
		AccessToken:  "old-token",
		RefreshToken: "rt",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}
	refreshCalled := false
	refreshFn := func(t *auth.Token) error {
		refreshCalled = true
		t.AccessToken = "new-token"
		return nil
	}
	client := NewClient(server.URL, token, refreshFn)
	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}
	if !refreshCalled {
		t.Error("expected refresh to be called")
	}
}

func TestClient_429_Backoff(t *testing.T) {
	var callCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		if count <= 2 {
			w.Header().Set("RateLimit-Reset", "1")
			w.WriteHeader(429)
			w.Write([]byte(`{"code":"RATE_LIMIT"}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after backoff, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestClient_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"code":"INVALID","description":"bad request"}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	_, err := client.Get("/test")
	if err == nil {
		t.Fatal("expected error for 400")
	}
	var apiErr *APIError
	if !isAPIError(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Code != "INVALID" {
		t.Errorf("expected INVALID, got %q", apiErr.Code)
	}
}

func isAPIError(err error, target **APIError) bool {
	e, ok := err.(*APIError)
	if ok {
		*target = e
	}
	return ok
}
```

**Step 2: 테스트 실패 확인**

Run: `go test ./internal/api/ -v`
Expected: FAIL

**Step 3: 구현**

```go
// internal/api/client.go
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

type APIError struct {
	StatusCode  int    `json:"-"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API 에러 %d: %s - %s", e.StatusCode, e.Code, e.Description)
}

type Response struct {
	StatusCode int
	Body       []byte
}

type RefreshFunc func(token *auth.Token) error

type Client struct {
	baseURL    string
	token      *auth.Token
	refreshFn  RefreshFunc
	httpClient *http.Client
}

func NewClient(baseURL string, token *auth.Token, refreshFn RefreshFunc) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		refreshFn:  refreshFn,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) Get(path string) (*Response, error) {
	return c.do("GET", path, nil)
}

func (c *Client) Post(path string, body io.Reader) (*Response, error) {
	return c.do("POST", path, body)
}

func (c *Client) do(method, path string, body io.Reader) (*Response, error) {
	if c.token.NeedsRefresh() && c.refreshFn != nil {
		if err := c.refreshFn(c.token); err != nil {
			return nil, fmt.Errorf("토큰 갱신 실패: %w", err)
		}
	}

	return c.doWithRetry(method, path, body, false)
}

func (c *Client) doWithRetry(method, path string, body io.Reader, retried401 bool) (*Response, error) {
	const maxRateLimitRetries = 3

	for attempt := 0; attempt <= maxRateLimitRetries; attempt++ {
		req, err := http.NewRequest(method, c.baseURL+path, body)
		if err != nil {
			return nil, fmt.Errorf("요청 생성 실패: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.token.AccessToken)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("네트워크 에러: %w", err)
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		switch {
		case resp.StatusCode == 401 && !retried401 && c.refreshFn != nil:
			if err := c.refreshFn(c.token); err != nil {
				return nil, fmt.Errorf("토큰 갱신 실패: %w", err)
			}
			return c.doWithRetry(method, path, body, true)

		case resp.StatusCode == 429 && attempt < maxRateLimitRetries:
			waitSeconds := parseRateLimitReset(resp.Header)
			time.Sleep(time.Duration(waitSeconds) * time.Second)
			continue

		case resp.StatusCode >= 400:
			apiErr := &APIError{StatusCode: resp.StatusCode}
			json.Unmarshal(respBody, apiErr)
			return nil, apiErr

		default:
			return &Response{StatusCode: resp.StatusCode, Body: respBody}, nil
		}
	}

	return nil, &APIError{StatusCode: 429, Code: "RATE_LIMIT_EXCEEDED", Description: "최대 재시도 횟수 초과"}
}

func parseRateLimitReset(header http.Header) int {
	for _, key := range []string{"RateLimit-Reset", "X-RateLimit-Reset"} {
		if val := header.Get(key); val != "" {
			if seconds, err := strconv.Atoi(val); err == nil && seconds > 0 {
				return seconds
			}
		}
	}
	return 1
}
```

**Step 4: 테스트 통과 확인**

Run: `go test ./internal/api/ -v -count=1`
Expected: PASS

**Step 5: 커밋**

```bash
git add internal/api/
git commit -m "feat: HTTP 클라이언트 구현 (401 재시도, 429 백오프)"
```

---

## Task 5: JWT 인증 (`internal/auth/jwt.go`)

**Files:**
- Create: `internal/auth/jwt.go`
- Create: `internal/auth/jwt_test.go`

**Step 1: 실패하는 테스트 작성**

```go
// internal/auth/jwt_test.go
package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func generateTestKey(t *testing.T) (string, *rsa.PrivateKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key gen failed: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "private.pem")
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	os.WriteFile(path, pemData, 0600)
	return path, key
}

func TestBuildJWTAssertion(t *testing.T) {
	keyPath, _ := generateTestKey(t)
	assertion, err := BuildJWTAssertion("client-id", "sa@example.com", keyPath)
	if err != nil {
		t.Fatalf("build assertion failed: %v", err)
	}
	if assertion == "" {
		t.Error("expected non-empty assertion")
	}
	// JWT는 3개의 dot-separated 파트
	parts := 0
	for _, c := range assertion {
		if c == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("expected 2 dots in JWT, got %d", parts)
	}
}

func TestBuildJWTAssertion_InvalidKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.pem")
	os.WriteFile(path, []byte("not a pem"), 0600)

	_, err := BuildJWTAssertion("client-id", "sa@example.com", path)
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestBuildJWTAssertion_FileNotFound(t *testing.T) {
	_, err := BuildJWTAssertion("client-id", "sa@example.com", "/nonexistent.pem")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestCheckKeyPermissions_Unix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "key.pem")
	os.WriteFile(path, []byte("test"), 0644)

	warning := CheckKeyPermissions(path)
	if warning == "" {
		t.Error("expected warning for 0644 permissions")
	}

	os.Chmod(path, 0600)
	warning = CheckKeyPermissions(path)
	if warning != "" {
		t.Errorf("expected no warning for 0600, got %q", warning)
	}
}
```

**Step 2: 테스트 실패 확인**

Run: `go test ./internal/auth/ -run TestBuild -v`
Expected: FAIL

**Step 3: 구현**

```go
// internal/auth/jwt.go
package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"runtime"
	"time"
)

func BuildJWTAssertion(clientID, serviceAccountID, privateKeyPath string) (string, error) {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("private key 파일 읽기 실패: %w", err)
	}

	key, err := parsePrivateKey(keyData)
	if err != nil {
		return "", err
	}

	now := time.Now()
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	payload := map[string]interface{}{
		"iss": clientID,
		"sub": serviceAccountID,
		"iat": now.Unix(),
		"exp": now.Add(1 * time.Hour).Unix(),
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := headerB64 + "." + payloadB64

	hash := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("JWT 서명 실패: %w", err)
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)
	return signingInput + "." + signatureB64, nil
}

func parsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("유효하지 않은 PEM 형식입니다. RSA PRIVATE KEY 또는 PRIVATE KEY 블록이 필요합니다")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("PKCS8 키 파싱 실패: %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("RSA 키가 아닙니다")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("지원하지 않는 PEM 블록 타입: %s", block.Type)
	}
}

func CheckKeyPermissions(path string) string {
	if runtime.GOOS == "windows" {
		return ""
	}
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		return fmt.Sprintf("경고: %s 파일 권한이 %04o입니다. 0600을 권장합니다", path, perm)
	}
	return ""
}
```

**Step 4: 테스트 통과 확인**

Run: `go test ./internal/auth/ -v`
Expected: PASS

**Step 5: 커밋**

```bash
git add internal/auth/jwt.go internal/auth/jwt_test.go
git commit -m "feat: JWT assertion 생성 및 키 권한 검증 구현"
```

---

## Task 6: OAuth 플로우 (`internal/auth/oauth.go`)

**Files:**
- Create: `internal/auth/oauth.go`
- Create: `internal/auth/oauth_test.go`

**Step 1: 실패하는 테스트 작성**

```go
// internal/auth/oauth_test.go
package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildAuthorizationURL(t *testing.T) {
	url := BuildAuthorizationURL("https://auth.example.com", "client-id", "http://localhost:8484/callback", "test-state", "openid profile bot")
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	// URL에 필수 파라미터 포함 확인
	for _, param := range []string{"client_id=client-id", "state=test-state", "response_type=code", "scope=openid"} {
		if !containsStr(url, param) {
			t.Errorf("URL missing %q: %s", param, url)
		}
	}
}

func TestExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "at-new",
			"refresh_token": "rt-new",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "openid profile bot",
		})
	}))
	defer server.Close()

	token, err := ExchangeCode(server.URL, "client-id", "client-secret", "auth-code", "http://localhost:8484/callback")
	if err != nil {
		t.Fatalf("exchange failed: %v", err)
	}
	if token.AccessToken != "at-new" {
		t.Errorf("expected at-new, got %q", token.AccessToken)
	}
	if token.AuthMethod != "oauth" {
		t.Errorf("expected oauth, got %q", token.AuthMethod)
	}
}

func TestRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "at-refreshed",
			"refresh_token": "rt-refreshed",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "bot",
		})
	}))
	defer server.Close()

	token := &Token{RefreshToken: "rt-old", AuthMethod: "oauth"}
	err := RefreshAccessToken(server.URL, "client-id", "client-secret", token)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if token.AccessToken != "at-refreshed" {
		t.Errorf("expected at-refreshed, got %q", token.AccessToken)
	}
}

func TestRevokeToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	err := RevokeToken(server.URL, "client-id", "client-secret", "some-token", "access_token")
	if err != nil {
		t.Fatalf("revoke failed: %v", err)
	}
}

func TestFindAvailablePort(t *testing.T) {
	port, err := FindAvailablePort(8484, 8494)
	if err != nil {
		t.Fatalf("find port failed: %v", err)
	}
	if port < 8484 || port > 8494 {
		t.Errorf("port %d out of range", port)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

**Step 2: 테스트 실패 확인**

Run: `go test ./internal/auth/ -run TestBuildAuth -v`
Expected: FAIL

**Step 3: 구현**

```go
// internal/auth/oauth.go
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func GenerateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func BuildAuthorizationURL(authBaseURL, clientID, redirectURI, state, scope string) string {
	params := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"state":         {state},
		"scope":         {scope},
		"response_type": {"code"},
	}
	return authBaseURL + "/authorize?" + params.Encode()
}

func ExchangeCode(authBaseURL, clientID, clientSecret, code, redirectURI string) (*Token, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}
	return requestToken(authBaseURL+"/token", data)
}

func RefreshAccessToken(authBaseURL, clientID, clientSecret string, token *Token) error {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {token.RefreshToken},
	}
	newToken, err := requestToken(authBaseURL+"/token", data)
	if err != nil {
		return err
	}
	token.AccessToken = newToken.AccessToken
	token.RefreshToken = newToken.RefreshToken
	token.ExpiresAt = newToken.ExpiresAt
	token.Scope = newToken.Scope
	return nil
}

func RequestJWTToken(authBaseURL, clientID, clientSecret, assertion, scope string) (*Token, error) {
	data := url.Values{
		"grant_type":    {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"assertion":     {assertion},
		"scope":         {scope},
	}
	token, err := requestToken(authBaseURL+"/token", data)
	if err != nil {
		return nil, err
	}
	token.AuthMethod = "jwt"
	return token, nil
}

func RevokeToken(authBaseURL, clientID, clientSecret, token, tokenTypeHint string) error {
	data := url.Values{
		"client_id":       {clientID},
		"client_secret":   {clientSecret},
		"token":           {token},
		"token_type_hint": {tokenTypeHint},
	}
	resp, err := http.PostForm(authBaseURL+"/revoke", data)
	if err != nil {
		return fmt.Errorf("revoke 요청 실패: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("revoke 실패: HTTP %d", resp.StatusCode)
	}
	return nil
}

func FindAvailablePort(start, end int) (int, error) {
	for port := start; port <= end; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("사용 가능한 포트를 찾을 수 없습니다 (%d-%d)", start, end)
}

func WaitForCallback(port int, expectedState string, timeout time.Duration) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != expectedState {
			errCh <- fmt.Errorf("state 불일치: expected %q, got %q", expectedState, state)
			w.WriteHeader(400)
			fmt.Fprint(w, "인증 실패: state 불일치")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("인증 코드가 없습니다: %s", r.URL.Query().Get("error"))
			w.WriteHeader(400)
			fmt.Fprint(w, "인증 실패")
			return
		}
		w.WriteHeader(200)
		fmt.Fprint(w, "인증 완료! 이 창을 닫아도 됩니다.")
		codeCh <- code
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go server.ListenAndServe()
	defer server.Close()

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("인증 타임아웃 (%v)", timeout)
	}
}

func requestToken(tokenURL string, data url.Values) (*Token, error) {
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("토큰 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("토큰 응답 파싱 실패: %w", err)
	}
	if result.Error != "" {
		return nil, fmt.Errorf("토큰 발급 실패: %s - %s", result.Error, result.ErrorDesc)
	}

	return &Token{
		AuthMethod:   "oauth",
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
		Scope:        result.Scope,
	}, nil
}

func HasScope(scopeStr string, target string) bool {
	for _, s := range strings.Fields(scopeStr) {
		if s == target {
			return true
		}
	}
	return false
}
```

**Step 4: 테스트 통과 확인**

Run: `go test ./internal/auth/ -v`
Expected: PASS

**Step 5: 커밋**

```bash
git add internal/auth/oauth.go internal/auth/oauth_test.go
git commit -m "feat: OAuth 플로우 구현 (코드 교환, 토큰 갱신, revoke, 로컬 콜백 서버)"
```

---

## Task 7: 출력 포맷터 (`internal/output`)

**Files:**
- Create: `internal/output/formatter.go`
- Create: `internal/output/formatter_test.go`

**Step 1: 실패하는 테스트 작성**

```go
// internal/output/formatter_test.go
package output

import (
	"bytes"
	"testing"
)

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("json", &buf)
	data := map[string]string{"name": "test", "id": "123"}
	f.Print(data)

	got := buf.String()
	if got == "" {
		t.Error("expected non-empty output")
	}
	if !containsStr(got, `"name"`) {
		t.Errorf("expected JSON with name field, got %s", got)
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("table", &buf)
	f.PrintTable([]string{"ID", "Name"}, [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	})

	got := buf.String()
	if !containsStr(got, "Alice") {
		t.Errorf("expected Alice in table, got %s", got)
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

**Step 2: 테스트 실패 확인**

Run: `go test ./internal/output/ -v`
Expected: FAIL

**Step 3: 구현**

```go
// internal/output/formatter.go
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

type Formatter struct {
	format string
	writer io.Writer
}

func NewFormatter(format string, writer io.Writer) *Formatter {
	return &Formatter{format: format, writer: writer}
}

func (f *Formatter) Print(data interface{}) {
	encoded, _ := json.MarshalIndent(data, "", "  ")
	fmt.Fprintln(f.writer, string(encoded))
}

func (f *Formatter) PrintRaw(data []byte) {
	var pretty interface{}
	if err := json.Unmarshal(data, &pretty); err == nil {
		f.Print(pretty)
	} else {
		fmt.Fprintln(f.writer, string(data))
	}
}

func (f *Formatter) PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(f.writer, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	fmt.Fprintln(w, strings.Repeat("-\t", len(headers)))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}
```

**Step 4: 테스트 통과 확인**

Run: `go test ./internal/output/ -v`
Expected: PASS

**Step 5: 커밋**

```bash
git add internal/output/
git commit -m "feat: JSON/테이블 출력 포맷터 구현"
```

---

## Task 8: auth 커맨드 (`cmd/auth.go`)

**Files:**
- Create: `cmd/auth.go`

> Note: per-command 초기화 방식을 사용하므로 root.go는 수정하지 않는다. 각 서비스 커맨드가 `loadConfigAndToken()` + `buildAPIClient()`를 직접 호출한다.

**Step 1: cmd/auth.go 작성**

```go
// cmd/auth.go
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/spf13/cobra"
)

const (
	authBaseURL     = "https://auth.worksmobile.com/oauth2/v2.0"
	defaultOAuthScope = "openid profile bot directory calendar"
	defaultJWTScope   = "bot directory calendar"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "인증 관리",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "로그인",
	RunE: func(cmd *cobra.Command, args []string) error {
		useJWT, _ := cmd.Flags().GetBool("jwt")
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return err
		}
		cfg.ApplyEnvOverrides()

		store := auth.NewTokenStore(auth.DefaultTokenPath())

		if useJWT {
			return loginJWT(cfg, store)
		}
		return loginOAuth(cfg, store)
	},
}

func loginJWT(cfg *config.Config, store *auth.TokenStore) error {
	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.ServiceAccountID == "" || cfg.PrivateKeyPath == "" {
		return fmt.Errorf("JWT 인증에 필요한 설정이 누락되었습니다: client_id, client_secret, service_account_id, private_key_path")
	}

	if warning := auth.CheckKeyPermissions(cfg.PrivateKeyPath); warning != "" {
		fmt.Fprintln(os.Stderr, warning)
	}

	scope := cfg.Scope
	if scope == "" {
		scope = defaultJWTScope
	}

	assertion, err := auth.BuildJWTAssertion(cfg.ClientID, cfg.ServiceAccountID, cfg.PrivateKeyPath)
	if err != nil {
		return err
	}

	token, err := auth.RequestJWTToken(authBaseURL, cfg.ClientID, cfg.ClientSecret, assertion, scope)
	if err != nil {
		return err
	}
	token.ServiceAccountID = cfg.ServiceAccountID

	return store.Save(token)
}

func loginOAuth(cfg *config.Config, store *auth.TokenStore) error {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return fmt.Errorf("OAuth 인증에 필요한 설정이 누락되었습니다: client_id, client_secret")
	}

	port, err := auth.FindAvailablePort(8484, 8494)
	if err != nil {
		return err
	}
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

	scope := cfg.Scope
	if scope == "" {
		scope = defaultOAuthScope
	}

	state := auth.GenerateState()
	authURL := auth.BuildAuthorizationURL(authBaseURL, cfg.ClientID, redirectURI, state, scope)

	if err := openBrowser(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "브라우저를 열 수 없습니다. 아래 URL을 직접 열어주세요:\n%s\n", authURL)
	}

	code, err := auth.WaitForCallback(port, state, 120*time.Second)
	if err != nil {
		return err
	}

	token, err := auth.ExchangeCode(authBaseURL, cfg.ClientID, cfg.ClientSecret, code, redirectURI)
	if err != nil {
		return err
	}

	return store.Save(token)
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "인증 상태 확인",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := auth.NewTokenStore(auth.DefaultTokenPath())
		token, err := store.Load()
		if err != nil {
			return err
		}
		if token == nil {
			return fmt.Errorf("로그인되어 있지 않습니다. nw-cli auth login을 실행하세요")
		}

		status := map[string]interface{}{
			"auth_method": token.AuthMethod,
			"expires_at":  token.ExpiresAt.Format(time.RFC3339),
			"scopes":      splitScopes(token.Scope),
		}

		if token.AuthMethod == "jwt" {
			status["service_account_id"] = token.ServiceAccountID
		} else if auth.HasScope(token.Scope, "openid") && auth.HasScope(token.Scope, "profile") {
			if name, err := fetchUserName(token.AccessToken); err == nil {
				status["user_name"] = name
			}
		}

		data, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "로그아웃",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load(config.DefaultPath())
		cfg.ApplyEnvOverrides()
		store := auth.NewTokenStore(auth.DefaultTokenPath())
		token, err := store.Load()
		if err != nil {
			return err
		}
		if token == nil {
			return fmt.Errorf("로그인되어 있지 않습니다")
		}

		if token.AuthMethod == "oauth" && cfg.ClientID != "" && cfg.ClientSecret != "" {
			if token.RefreshToken != "" {
				if err := auth.RevokeToken(authBaseURL, cfg.ClientID, cfg.ClientSecret, token.RefreshToken, "refresh_token"); err != nil {
					fmt.Fprintf(os.Stderr, "경고: refresh token revoke 실패: %v\n", err)
				}
			}
			if err := auth.RevokeToken(authBaseURL, cfg.ClientID, cfg.ClientSecret, token.AccessToken, "access_token"); err != nil {
				fmt.Fprintf(os.Stderr, "경고: access token revoke 실패: %v\n", err)
			}
		}

		return store.Delete()
	},
}

func init() {
	authLoginCmd.Flags().Bool("jwt", false, "JWT Service Account 인증")
	authCmd.AddCommand(authLoginCmd, authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	return fmt.Errorf("지원하지 않는 OS")
}

func splitScopes(scope string) []string {
	if scope == "" {
		return []string{}
	}
	result := []string{}
	current := ""
	for _, c := range scope {
		if c == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func fetchUserName(accessToken string) (string, error) {
	req, _ := http.NewRequest("GET", authBaseURL+"/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var info struct {
		Name string `json:"name"`
	}
	json.NewDecoder(resp.Body).Decode(&info)
	return info.Name, nil
}
```

**Step 2: 빌드 확인**

Run: `go build -o nw-cli .`
Expected: 성공

**Step 3: 커밋**

```bash
git add cmd/auth.go
git commit -m "feat: auth 커맨드 구현 (login/status/logout, JWT/OAuth)"
```

---

## Task 9: Bot 서비스 API + 커맨드

**Files:**
- Create: `internal/api/bot.go`
- Create: `internal/api/bot_test.go`
- Create: `cmd/bot.go`

**Step 1: 실패하는 테스트 작성**

```go
// internal/api/bot_test.go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestBotService_SendTextToUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bots/123/users/user1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(201)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	bot := NewBotService(client)

	resp, err := bot.SendTextToUser("123", "user1", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestBotService_SendTextToChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bots/123/channels/ch1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(201)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	bot := NewBotService(client)

	resp, err := bot.SendTextToChannel("123", "ch1", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}
```

**Step 2: 테스트 실패 확인**

Run: `go test ./internal/api/ -run TestBot -v`
Expected: FAIL

**Step 3: API 서비스 구현**

```go
// internal/api/bot.go
package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

type BotService struct {
	client *Client
}

func NewBotService(client *Client) *BotService {
	return &BotService{client: client}
}

func (s *BotService) SendTextToUser(botID, userID, text string) (*Response, error) {
	body := map[string]interface{}{
		"content": map[string]interface{}{
			"type": "text",
			"text": text,
		},
	}
	data, _ := json.Marshal(body)
	return s.client.Post(fmt.Sprintf("/bots/%s/users/%s/messages", botID, userID), strings.NewReader(string(data)))
}

func (s *BotService) SendTextToChannel(botID, channelID, text string) (*Response, error) {
	body := map[string]interface{}{
		"content": map[string]interface{}{
			"type": "text",
			"text": text,
		},
	}
	data, _ := json.Marshal(body)
	return s.client.Post(fmt.Sprintf("/bots/%s/channels/%s/messages", botID, channelID), strings.NewReader(string(data)))
}

func (s *BotService) GetChannel(botID, channelID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s/channels/%s", botID, channelID))
}

func (s *BotService) ListChannelMembers(botID, channelID, cursor string, count int) (*Response, error) {
	path := fmt.Sprintf("/bots/%s/channels/%s/members", botID, channelID)
	params := []string{}
	if cursor != "" {
		params = append(params, "cursor="+cursor)
	}
	if count > 0 {
		params = append(params, fmt.Sprintf("count=%d", count))
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}
	return s.client.Get(path)
}
```

**Step 4: 테스트 통과 확인**

Run: `go test ./internal/api/ -run TestBot -v`
Expected: PASS

**Step 5: cmd/bot.go 작성**

```go
// cmd/bot.go
package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

const apiBaseURL = "https://www.worksapis.com/v1.0"

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Bot 메시지 관리",
}

var botSendCmd = &cobra.Command{
	Use:   "send",
	Short: "메시지 전송",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		if cfg.BotID == "" {
			return fmt.Errorf("bot_id가 설정되지 않았습니다. nw-cli config set bot_id <id>")
		}

		client := buildAPIClient(cfg, token)
		bot := api.NewBotService(client)
		formatter := output.NewFormatter(outputFormat, os.Stdout)

		to, _ := cmd.Flags().GetString("to")
		channel, _ := cmd.Flags().GetString("channel")
		text, _ := cmd.Flags().GetString("text")

		if text == "-" {
			scanner := bufio.NewScanner(os.Stdin)
			text = ""
			for scanner.Scan() {
				if text != "" {
					text += "\n"
				}
				text += scanner.Text()
			}
		}

		if to == "" && channel == "" {
			return fmt.Errorf("--to 또는 --channel 중 하나를 지정하세요")
		}
		if text == "" {
			return fmt.Errorf("--text를 지정하세요")
		}

		var resp *api.Response
		if to != "" {
			resp, err = bot.SendTextToUser(cfg.BotID, to, text)
		} else {
			resp, err = bot.SendTextToChannel(cfg.BotID, channel, text)
		}
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var botGetChannelCmd = &cobra.Command{
	Use:   "get-channel <channelId>",
	Short: "채널 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		if cfg.BotID == "" {
			return fmt.Errorf("bot_id가 설정되지 않았습니다")
		}
		client := buildAPIClient(cfg, token)
		bot := api.NewBotService(client)

		resp, err := bot.GetChannel(cfg.BotID, args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var botChannelMembersCmd = &cobra.Command{
	Use:   "channel-members <channelId>",
	Short: "채널 멤버 목록",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		if cfg.BotID == "" {
			return fmt.Errorf("bot_id가 설정되지 않았습니다")
		}
		client := buildAPIClient(cfg, token)
		bot := api.NewBotService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")

		resp, err := bot.ListChannelMembers(cfg.BotID, args[0], cursor, count)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	botSendCmd.Flags().String("to", "", "수신자 userId")
	botSendCmd.Flags().String("channel", "", "채널 ID")
	botSendCmd.Flags().String("text", "", "메시지 텍스트 (- 이면 stdin)")

	botChannelMembersCmd.Flags().String("cursor", "", "페이지네이션 커서")
	botChannelMembersCmd.Flags().Int("count", 0, "페이지 크기")

	botCmd.AddCommand(botSendCmd, botGetChannelCmd, botChannelMembersCmd)
	rootCmd.AddCommand(botCmd)
}

func loadConfigAndToken() (*config.Config, *auth.Token, error) {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return nil, nil, err
	}
	cfg.ApplyEnvOverrides()

	store := auth.NewTokenStore(auth.DefaultTokenPath())
	token, err := store.Load()
	if err != nil {
		return nil, nil, err
	}
	if token == nil {
		return nil, nil, fmt.Errorf("로그인되어 있지 않습니다. nw-cli auth login을 실행하세요")
	}
	return cfg, token, nil
}

func buildAPIClient(cfg *config.Config, token *auth.Token) *api.Client {
	refreshFn := func(t *auth.Token) error {
		if t.AuthMethod == "oauth" && t.RefreshToken != "" {
			return auth.RefreshAccessToken(authBaseURL, cfg.ClientID, cfg.ClientSecret, t)
		}
		return fmt.Errorf("토큰 갱신 불가")
	}
	return api.NewClient(apiBaseURL, token, refreshFn)
}
```

**Step 6: 빌드 확인**

Run: `go build -o nw-cli .`
Expected: 성공

**Step 7: 커밋**

```bash
git add internal/api/bot.go internal/api/bot_test.go cmd/bot.go
git commit -m "feat: Bot 서비스 API 및 커맨드 구현 (send/get-channel/channel-members)"
```

---

## Task 10: Directory 서비스 API + 커맨드

**Files:**
- Create: `internal/api/directory.go`
- Create: `internal/api/directory_test.go`
- Create: `cmd/directory.go`

**Step 1: 실패하는 테스트 작성**

```go
// internal/api/directory_test.go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestDirectoryService_ListUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"users":[]}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	dir := NewDirectoryService(client)

	resp, err := dir.ListUsers("", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDirectoryService_GetUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/user1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"userId":"user1"}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	dir := NewDirectoryService(client)

	resp, err := dir.GetUser("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
```

**Step 2: 테스트 실패 확인 → 구현**

```go
// internal/api/directory.go
package api

import (
	"fmt"
	"strings"
)

type DirectoryService struct {
	client *Client
}

func NewDirectoryService(client *Client) *DirectoryService {
	return &DirectoryService{client: client}
}

func (s *DirectoryService) ListUsers(cursor string, count int) (*Response, error) {
	return s.client.Get("/users" + buildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetUser(userID string) (*Response, error) {
	return s.client.Get("/users/" + userID)
}

func (s *DirectoryService) ListGroups(cursor string, count int) (*Response, error) {
	return s.client.Get("/groups" + buildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetGroup(groupID string) (*Response, error) {
	return s.client.Get("/groups/" + groupID)
}

func buildPaginationQuery(cursor string, count int) string {
	params := []string{}
	if cursor != "" {
		params = append(params, "cursor="+cursor)
	}
	if count > 0 {
		params = append(params, fmt.Sprintf("count=%d", count))
	}
	if len(params) > 0 {
		return "?" + strings.Join(params, "&")
	}
	return ""
}
```

**Step 3: cmd/directory.go 작성**

```go
// cmd/directory.go
package cmd

import (
	"os"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var directoryCmd = &cobra.Command{
	Use:   "directory",
	Short: "디렉토리 관리 (사용자, 그룹)",
}

var dirListUsersCmd = &cobra.Command{
	Use:   "list-users",
	Short: "사용자 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")

		resp, err := dir.ListUsers(cursor, count)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var dirGetUserCmd = &cobra.Command{
	Use:   "get-user <userId>",
	Short: "사용자 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		resp, err := dir.GetUser(args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var dirListGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "그룹 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")

		resp, err := dir.ListGroups(cursor, count)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var dirGetGroupCmd = &cobra.Command{
	Use:   "get-group <groupId>",
	Short: "그룹 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		resp, err := dir.GetGroup(args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	for _, cmd := range []*cobra.Command{dirListUsersCmd, dirListGroupsCmd} {
		cmd.Flags().String("cursor", "", "페이지네이션 커서")
		cmd.Flags().Int("count", 0, "페이지 크기")
	}
	directoryCmd.AddCommand(dirListUsersCmd, dirGetUserCmd, dirListGroupsCmd, dirGetGroupCmd)
	rootCmd.AddCommand(directoryCmd)
}
```

**Step 4: 테스트 통과 + 빌드 확인**

Run: `go test ./internal/api/ -v && go build -o nw-cli .`
Expected: PASS + 빌드 성공

**Step 5: 커밋**

```bash
git add internal/api/directory.go internal/api/directory_test.go cmd/directory.go
git commit -m "feat: Directory 서비스 API 및 커맨드 구현 (list-users/get-user/list-groups/get-group)"
```

---

## Task 11: Calendar 서비스 API + 커맨드

**Files:**
- Create: `internal/api/calendar.go`
- Create: `internal/api/calendar_test.go`
- Create: `cmd/calendar.go`

**Step 1: 실패하는 테스트 작성**

```go
// internal/api/calendar_test.go
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestCalendarService_ListCalendars(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/user1/calendar-personals" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"calendarPersonals":[]}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	cal := NewCalendarService(client)

	resp, err := cal.ListCalendars("user1", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCalendarService_GetDefaultCalendar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/user1/calendar" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"calendarId":"default"}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	cal := NewCalendarService(client)

	resp, err := cal.GetDefaultCalendar("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCalendarService_ListEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/user1/calendars/cal1/events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		from := r.URL.Query().Get("fromDateTime")
		until := r.URL.Query().Get("untilDateTime")
		if from == "" || until == "" {
			t.Error("expected fromDateTime and untilDateTime")
		}
		w.Write([]byte(`{"events":[]}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	cal := NewCalendarService(client)

	resp, err := cal.ListEvents("user1", "cal1", "2026-03-01T00:00:00Z", "2026-03-31T23:59:59Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
```

**Step 2: 테스트 실패 확인 → 구현**

```go
// internal/api/calendar.go
package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type CalendarService struct {
	client *Client
}

func NewCalendarService(client *Client) *CalendarService {
	return &CalendarService{client: client}
}

func (s *CalendarService) ListCalendars(userID, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/calendar-personals", userID) + buildPaginationQuery(cursor, count))
}

func (s *CalendarService) GetDefaultCalendar(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/calendar", userID))
}

func (s *CalendarService) ListEvents(userID, calendarID, from, until string) (*Response, error) {
	params := url.Values{
		"fromDateTime":  {from},
		"untilDateTime": {until},
	}
	return s.client.Get(fmt.Sprintf("/users/%s/calendars/%s/events?%s", userID, calendarID, params.Encode()))
}

func (s *CalendarService) GetEvent(userID, calendarID, eventID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/calendars/%s/events/%s", userID, calendarID, eventID))
}

func (s *CalendarService) CreateEvent(userID, calendarID string, event map[string]interface{}) (*Response, error) {
	body := map[string]interface{}{
		"eventComponents": []interface{}{event},
	}
	data, _ := json.Marshal(body)
	return s.client.Post(fmt.Sprintf("/users/%s/calendars/%s/events", userID, calendarID), strings.NewReader(string(data)))
}
```

**Step 3: cmd/calendar.go 작성**

```go
// cmd/calendar.go
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "캘린더 관리",
}

func resolveCalendarUserID(cmd *cobra.Command, token_authMethod string, defaultUID string) (string, error) {
	userID, _ := cmd.Flags().GetString("user-id")
	if userID == "" {
		userID = defaultUID
	}
	if userID == "" {
		return "", fmt.Errorf("--user-id를 지정하거나 config set default_calendar_user_id를 설정하세요")
	}
	if userID == "me" && token_authMethod == "jwt" {
		return "", fmt.Errorf("JWT 모드에서는 --user-id me를 사용할 수 없습니다. 명시적 userId를 지정하세요")
	}
	return userID, nil
}

var calListCalendarsCmd = &cobra.Command{
	Use:   "list-calendars",
	Short: "캘린더 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveCalendarUserID(cmd, token.AuthMethod, cfg.DefaultCalendarUserID)
		if err != nil {
			return err
		}

		client := buildAPIClient(cfg, token)
		cal := api.NewCalendarService(client)

		useDefault, _ := cmd.Flags().GetBool("default")
		if useDefault {
			resp, err := cal.GetDefaultCalendar(userID)
			if err != nil {
				return err
			}
			output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
			return nil
		}

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		resp, err := cal.ListCalendars(userID, cursor, count)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var calListEventsCmd = &cobra.Command{
	Use:   "list-events",
	Short: "일정 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveCalendarUserID(cmd, token.AuthMethod, cfg.DefaultCalendarUserID)
		if err != nil {
			return err
		}

		calendarID, _ := cmd.Flags().GetString("calendar-id")
		from, _ := cmd.Flags().GetString("from")
		until, _ := cmd.Flags().GetString("until")

		if calendarID == "" || from == "" || until == "" {
			return fmt.Errorf("--calendar-id, --from, --until은 필수입니다")
		}

		fromTime, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return fmt.Errorf("--from 형식 오류 (RFC3339): %w", err)
		}
		untilTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			return fmt.Errorf("--until 형식 오류 (RFC3339): %w", err)
		}
		if untilTime.Sub(fromTime) > 31*24*time.Hour {
			return fmt.Errorf("--from과 --until 간격은 최대 31일입니다")
		}

		client := buildAPIClient(cfg, token)
		cal := api.NewCalendarService(client)

		resp, err := cal.ListEvents(userID, calendarID, from, until)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var calGetEventCmd = &cobra.Command{
	Use:   "get-event",
	Short: "일정 상세 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveCalendarUserID(cmd, token.AuthMethod, cfg.DefaultCalendarUserID)
		if err != nil {
			return err
		}

		calendarID, _ := cmd.Flags().GetString("calendar-id")
		eventID, _ := cmd.Flags().GetString("event-id")
		if calendarID == "" || eventID == "" {
			return fmt.Errorf("--calendar-id와 --event-id는 필수입니다")
		}

		client := buildAPIClient(cfg, token)
		cal := api.NewCalendarService(client)

		resp, err := cal.GetEvent(userID, calendarID, eventID)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var calCreateEventCmd = &cobra.Command{
	Use:   "create-event",
	Short: "일정 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveCalendarUserID(cmd, token.AuthMethod, cfg.DefaultCalendarUserID)
		if err != nil {
			return err
		}

		calendarID, _ := cmd.Flags().GetString("calendar-id")
		title, _ := cmd.Flags().GetString("title")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		description, _ := cmd.Flags().GetString("description")
		location, _ := cmd.Flags().GetString("location")
		isAllDay, _ := cmd.Flags().GetBool("is-all-day")

		if calendarID == "" || title == "" || start == "" || end == "" {
			return fmt.Errorf("--calendar-id, --title, --start, --end는 필수입니다")
		}

		event := map[string]interface{}{
			"summary": title,
		}
		if isAllDay {
			event["startDate"] = start[:10]
			event["endDate"] = end[:10]
			event["isAllDay"] = true
		} else {
			event["startDateTime"] = start
			event["endDateTime"] = end
		}
		if description != "" {
			event["description"] = description
		}
		if location != "" {
			event["location"] = location
		}

		client := buildAPIClient(cfg, token)
		cal := api.NewCalendarService(client)

		resp, err := cal.CreateEvent(userID, calendarID, event)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	for _, cmd := range []*cobra.Command{calListCalendarsCmd, calListEventsCmd, calGetEventCmd, calCreateEventCmd} {
		cmd.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	calListCalendarsCmd.Flags().Bool("default", false, "기본 캘린더만 조회")
	calListCalendarsCmd.Flags().String("cursor", "", "페이지네이션 커서")
	calListCalendarsCmd.Flags().Int("count", 0, "페이지 크기")

	calListEventsCmd.Flags().String("calendar-id", "", "캘린더 ID (필수)")
	calListEventsCmd.Flags().String("from", "", "시작 시간 RFC3339 (필수)")
	calListEventsCmd.Flags().String("until", "", "종료 시간 RFC3339 (필수)")

	calGetEventCmd.Flags().String("calendar-id", "", "캘린더 ID (필수)")
	calGetEventCmd.Flags().String("event-id", "", "이벤트 ID (필수)")

	calCreateEventCmd.Flags().String("calendar-id", "", "캘린더 ID (필수)")
	calCreateEventCmd.Flags().String("title", "", "일정 제목 (필수)")
	calCreateEventCmd.Flags().String("start", "", "시작 시간 RFC3339 (필수)")
	calCreateEventCmd.Flags().String("end", "", "종료 시간 RFC3339 (필수)")
	calCreateEventCmd.Flags().String("description", "", "설명")
	calCreateEventCmd.Flags().String("location", "", "장소")
	calCreateEventCmd.Flags().Bool("is-all-day", false, "종일 일정")

	calendarCmd.AddCommand(calListCalendarsCmd, calListEventsCmd, calGetEventCmd, calCreateEventCmd)
	rootCmd.AddCommand(calendarCmd)
}
```

**Step 4: 테스트 통과 + 빌드 확인**

Run: `go test ./internal/api/ -v && go build -o nw-cli .`
Expected: PASS + 빌드 성공

**Step 5: 커밋**

```bash
git add internal/api/calendar.go internal/api/calendar_test.go cmd/calendar.go
git commit -m "feat: Calendar 서비스 API 및 커맨드 구현 (list-calendars/list-events/get-event/create-event)"
```

---

## Task 12: 전체 통합 테스트 + 최종 빌드

**Step 1: 전체 테스트 실행**

Run: `go test ./... -v`
Expected: 모든 테스트 PASS

**Step 2: 빌드 및 CLI 동작 확인**

Run: `make build && ./nw-cli --help`
Expected: 모든 서브커맨드 표시 (auth, config, bot, directory, calendar, version)

Run: `./nw-cli version`
Expected: 빌드 정보 JSON

Run: `./nw-cli auth --help`
Expected: login, status, logout 서브커맨드

Run: `./nw-cli bot --help`
Expected: send, get-channel, channel-members 서브커맨드

Run: `./nw-cli directory --help`
Expected: list-users, get-user, list-groups, get-group 서브커맨드

Run: `./nw-cli calendar --help`
Expected: list-calendars, list-events, get-event, create-event 서브커맨드

**Step 3: 바이너리 크기 확인**

Run: `ls -lh nw-cli`
Expected: < 15MB

**Step 4: 최종 커밋**

```bash
git add -A
git commit -m "chore: 전체 통합 테스트 및 빌드 확인 완료 (v0.1)"
```

---

## 태스크 의존성 요약

```
Task 1 (스켈레톤) → Task 2 (config) → Task 3 (token store)
                                              ↓
Task 4 (HTTP client) ← Task 3 ← Task 5 (JWT) + Task 6 (OAuth)
                                              ↓
Task 7 (output) → Task 8 (auth cmd) → Task 9 (bot) → Task 10 (directory) → Task 11 (calendar) → Task 12 (통합)
```

**순차 실행이 필요한 경로:** 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10 → 11 → 12

**병렬 가능:** Task 5와 6은 Task 3 이후 병렬 가능. Task 9, 10, 11은 Task 8 이후 병렬 가능.

---

## 정오표: Codex 리뷰 반영 사항

아래 수정사항은 각 Task의 코드를 구현할 때 반드시 적용해야 한다.

### E1. HTTP client body 재사용 (Task 4 수정)

`Client.do()`와 `Client.Post()`에서 `io.Reader` 대신 `[]byte`를 받도록 변경한다:

```go
// internal/api/client.go 수정
func (c *Client) Post(path string, body []byte) (*Response, error) {
	return c.doWithRetry("POST", path, body, false)
}

func (c *Client) doWithRetry(method, path string, body []byte, retried401 bool) (*Response, error) {
	// 재시도 시 bytes.NewReader(body)로 매번 새 Reader 생성
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	// ...
}
```

429 지수 백오프도 수정:
```go
// 헤더 없을 때 지수 백오프
func backoffDuration(attempt int) time.Duration {
	return time.Duration(1<<uint(attempt)) * time.Second // 1s, 2s, 4s
}
```

### E2. JWT 토큰 갱신 경로 (Task 8 수정)

`buildAPIClient()`의 `refreshFn`을 JWT도 지원하도록 수정:

```go
func buildAPIClient(cfg *config.Config, token *auth.Token) *api.Client {
	refreshFn := func(t *auth.Token) error {
		store := auth.NewTokenStore(auth.DefaultTokenPath())
		if t.AuthMethod == "oauth" && t.RefreshToken != "" {
			if err := auth.RefreshAccessToken(authBaseURL, cfg.ClientID, cfg.ClientSecret, t); err == nil {
				return store.Save(t)
			}
		}
		if t.AuthMethod == "jwt" {
			// refresh_token 시도, 실패하면 assertion 재발급
			if t.RefreshToken != "" {
				if err := auth.RefreshAccessToken(authBaseURL, cfg.ClientID, cfg.ClientSecret, t); err == nil {
					return store.Save(t)
				}
			}
			assertion, err := auth.BuildJWTAssertion(cfg.ClientID, cfg.ServiceAccountID, cfg.PrivateKeyPath)
			if err != nil {
				return err
			}
			scope := cfg.Scope
			if scope == "" {
				scope = defaultJWTScope
			}
			newToken, err := auth.RequestJWTToken(authBaseURL, cfg.ClientID, cfg.ClientSecret, assertion, scope)
			if err != nil {
				return err
			}
			t.AccessToken = newToken.AccessToken
			t.RefreshToken = newToken.RefreshToken
			t.ExpiresAt = newToken.ExpiresAt
			return store.Save(t)
		}
		return fmt.Errorf("토큰 갱신 불가")
	}
	return api.NewClient(apiBaseURL, token, refreshFn)
}
```

### E3. Windows ACL 보안 (Task 2, 3 수정)

파일 저장 시 OS별 권한 설정. 새 파일 `internal/platform/permissions.go`를 생성한다:

```go
// internal/platform/permissions.go
package platform

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// SetSecurePermissions는 파일을 현재 사용자만 접근 가능하게 설정한다.
func SetSecurePermissions(path string) error {
	if runtime.GOOS == "windows" {
		return setWindowsACL(path)
	}
	return os.Chmod(path, 0600)
}

func setWindowsACL(path string) error {
	user := os.Getenv("USERNAME")
	if user == "" {
		return fmt.Errorf("USERNAME 환경변수가 비어 있습니다")
	}
	// 상속 제거 + 현재 사용자에게만 Full Control
	cmd := exec.Command("icacls", path, "/inheritance:r", "/grant:r", user+":F")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("icacls 실패: %s - %w", string(out), err)
	}
	return nil
}

// CheckKeyPermissions는 private key 파일 권한이 안전한지 검사한다.
// 안전하지 않으면 경고 문자열을 반환하고, 안전하면 빈 문자열을 반환한다.
func CheckKeyPermissions(path string) string {
	if runtime.GOOS == "windows" {
		return checkWindowsKeyPermissions(path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		return fmt.Sprintf("경고: %s 파일 권한이 %04o입니다. 0600을 권장합니다", path, perm)
	}
	return ""
}

func checkWindowsKeyPermissions(path string) string {
	out, err := exec.Command("icacls", path).Output()
	if err != nil {
		return "" // icacls 실행 불가 시 검사 건너뜀
	}
	lines := strings.Split(string(out), "\n")
	user := os.Getenv("USERNAME")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "Successfully") {
			continue
		}
		// 파일 경로만 있는 라인 건너뛰기
		if !strings.Contains(trimmed, ":") {
			continue
		}
		// 현재 사용자의 ACE가 아닌 라인이 있으면 경고
		if !strings.Contains(strings.ToLower(trimmed), strings.ToLower(user)) {
			return fmt.Sprintf("경고: %s에 현재 사용자(%s) 외의 접근 권한이 설정되어 있습니다", path, user)
		}
	}
	return ""
}
```

`config.Save()`, `TokenStore.Save()` 모두:
1. `os.MkdirAll(dir, 0700)` 후 `platform.SetSecurePermissions(dir)` — 부모 디렉토리 ACL 설정
2. `os.WriteFile(path, data, 0600)` 후 `platform.SetSecurePermissions(path)` — 파일 ACL 설정

```go
// config.Save() 예시
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}
	platform.SetSecurePermissions(dir) // 부모 디렉토리도 보안 설정
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("config 직렬화 실패: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	return platform.SetSecurePermissions(path) // 파일 보안 설정
}
```

`auth/jwt.go`의 `CheckKeyPermissions()`를 제거하고 `platform.CheckKeyPermissions()`로 대체.

### E4. config 마스킹 수정 (Task 2 수정)

`sensitiveKeys`에 `private_key_path` 추가하지 않는다 — 요구사항은 "private_key_path의 내용"을 마스킹하라는 것이지 경로 자체가 아니다. 대신 `config get`에서 잘못된 키 에러를 반환하도록 수정:

```go
var configGetCmd = &cobra.Command{
	// ...
	RunE: func(cmd *cobra.Command, args []string) error {
		// ...
		val := cfg.GetMasked(args[0])
		if !config.IsValidKey(args[0]) {
			return fmt.Errorf("유효하지 않은 설정 키: %s", args[0])
		}
		fmt.Println(val)
		return nil
	},
}
```

`config` 패키지에 `IsValidKey()` 추가:
```go
func IsValidKey(key string) bool {
	return validKeys[key]
}
```

### E5. bot send 빈 응답 처리 (Task 9 수정)

요구사항: "HTTP 201 상태와 빈 객체를 출력한다"

```go
// cmd/bot.go 수정 — send 명령의 응답 출력 부분
if len(resp.Body) == 0 || strings.TrimSpace(string(resp.Body)) == "" {
	fmt.Println("{}")
} else {
	formatter.PrintRaw(resp.Body)
}
```

빈 body 시 `{}`를 출력하고 exit 0으로 성공을 표현한다.

### E6. create-event 요청 바디 구조 (Task 11 수정)

SDK의 실제 구조에 맞게 수정:

```go
func (s *CalendarService) CreateEvent(userID, calendarID string, event map[string]interface{}) (*Response, error) {
	body := map[string]interface{}{
		"eventComponents": []interface{}{event},
	}
	data, _ := json.Marshal(body)
	return s.client.Post(
		fmt.Sprintf("/users/%s/calendars/%s/events", userID, calendarID),
		data,
	)
}
```

`cmd/calendar.go`의 `calCreateEventCmd`에서 이벤트 구성:

```go
event := map[string]interface{}{
	"summary": title,
}
if isAllDay {
	event["start"] = map[string]string{"date": start[:10]}
	event["end"] = map[string]string{"date": end[:10]}
	event["isAllDay"] = true
} else {
	event["start"] = map[string]string{"dateTime": start}
	event["end"] = map[string]string{"dateTime": end}
}
```

### E7. --all 페이지네이션 자동 순회 (Task 9, 10 수정)

cursor 기반 페이지네이션 대상 명령에만 `--all` 플래그 추가: `directory list-users`, `directory list-groups`, `calendar list-calendars`, `bot channel-members`. `calendar list-events`는 시간 범위 기반이므로 제외. cursor를 URL 인코딩:

```go
// internal/api/pagination.go (새 파일)
package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func BuildPaginationQuery(cursor string, count int) string {
	params := url.Values{}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	if count > 0 {
		params.Set("count", fmt.Sprintf("%d", count))
	}
	if len(params) > 0 {
		return "?" + params.Encode()
	}
	return ""
}

func ExtractNextCursor(body []byte) string {
	var resp struct {
		ResponseMetaData struct {
			NextCursor string `json:"nextCursor"`
		} `json:"responseMetaData"`
	}
	json.Unmarshal(body, &resp)
	return resp.ResponseMetaData.NextCursor
}
```

각 목록 커맨드에서 `--all` 시 루프. 합산 결과는 원래 응답의 데이터 키에 모든 결과를 합치고 `responseMetaData`를 제거한 형태로 출력:

```go
// 예시: directory list-users에서 --all 사용 시
all, _ := cmd.Flags().GetBool("all")
if all {
	var allUsers []json.RawMessage
	cursor := ""
	for {
		resp, err := dir.ListUsers(cursor, count)
		if err != nil { return err }

		var page struct {
			Users            []json.RawMessage `json:"users"`
			ResponseMetaData struct {
				NextCursor string `json:"nextCursor"`
			} `json:"responseMetaData"`
		}
		json.Unmarshal(resp.Body, &page)
		allUsers = append(allUsers, page.Users...)

		if page.ResponseMetaData.NextCursor == "" { break }
		cursor = page.ResponseMetaData.NextCursor
	}
	// 합산 결과를 원래 키로 출력
	merged := map[string]interface{}{"users": allUsers}
	formatter.Print(merged)
	return nil
}
```

다른 대상 커맨드(`list-groups` → `"groups"`, `list-calendars` → `"calendarPersonals"`, `channel-members` → `"members"`)도 동일 패턴 적용.

### E8. --output table 지원 (Task 7 수정)

`Formatter`에 `commandKey`를 추가하여 테이블 컬럼을 결정한다:

```go
// internal/output/formatter.go 수정
type Formatter struct {
	format     string
	writer     io.Writer
	columns    []string  // 테이블 모드에서 사용할 컬럼 키
	dataKey    string    // JSON 응답에서 배열을 추출할 키 (예: "users", "groups")
}

func NewFormatter(format string, writer io.Writer) *Formatter {
	return &Formatter{format: format, writer: writer}
}

// WithTable은 테이블 모드를 위한 컬럼과 데이터 키를 설정한다.
func (f *Formatter) WithTable(columns []string, dataKey string) *Formatter {
	f.columns = columns
	f.dataKey = dataKey
	return f
}

func (f *Formatter) PrintRaw(data []byte) {
	if f.format == "table" && len(f.columns) > 0 {
		f.printAsTable(data)
		return
	}
	// 기존 JSON pretty-print
	var pretty interface{}
	if err := json.Unmarshal(data, &pretty); err == nil {
		f.Print(pretty)
	} else {
		fmt.Fprintln(f.writer, string(data))
	}
}

func (f *Formatter) printAsTable(data []byte) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		fmt.Fprintln(f.writer, string(data))
		return
	}
	arrayData, ok := raw[f.dataKey]
	if !ok {
		// 단건 응답이면 JSON 그대로 출력 (테이블 불가)
		fmt.Fprintln(f.writer, string(data))
		return
	}
	var items []map[string]interface{}
	json.Unmarshal(arrayData, &items)

	rows := make([][]string, 0, len(items))
	for _, item := range items {
		row := make([]string, len(f.columns))
		for i, col := range f.columns {
			if v, ok := item[col]; ok {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		rows = append(rows, row)
	}
	f.PrintTable(f.columns, rows)
}
```

각 커맨드 호출부를 수정:
```go
// cmd/directory.go 예시
formatter := output.NewFormatter(outputFormat, os.Stdout).
	WithTable([]string{"userId", "userName", "email"}, "users")
formatter.PrintRaw(resp.Body)

// cmd/bot.go 예시 (channel-members)
formatter := output.NewFormatter(outputFormat, os.Stdout).
	WithTable([]string{"userId"}, "members")
formatter.PrintRaw(resp.Body)

// cmd/calendar.go 예시 (list-events)
formatter := output.NewFormatter(outputFormat, os.Stdout).
	WithTable([]string{"eventId", "summary", "start", "end"}, "events")
formatter.PrintRaw(resp.Body)
```

### E9. calendar list-events from > until 검증 (Task 11 수정)

```go
if untilTime.Before(fromTime) {
	return fmt.Errorf("--from이 --until보다 이후입니다")
}
if untilTime.Sub(fromTime) > 31*24*time.Hour {
	return fmt.Errorf("--from과 --until 간격은 최대 31일입니다")
}
```

### E10. TDD 보강 — 커맨드 레벨 테스트 추가 (각 Task에 적용)

Task 8~11에서 cobra 커맨드 레벨 테스트를 추가한다. 예시:

```go
// cmd/bot_test.go
func TestBotSend_MissingBotID(t *testing.T) {
	rootCmd.SetArgs([]string{"bot", "send", "--to", "user1", "--text", "hello"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for missing bot_id")
	}
}

func TestCalendarListEvents_FromAfterUntil(t *testing.T) {
	rootCmd.SetArgs([]string{"calendar", "list-events",
		"--calendar-id", "cal1",
		"--user-id", "user1",
		"--from", "2026-04-01T00:00:00Z",
		"--until", "2026-03-01T00:00:00Z"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error for from > until")
	}
}
```
