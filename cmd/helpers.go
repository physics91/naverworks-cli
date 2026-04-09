package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	apiBaseURL  = "https://www.worksapis.com/v1.0"
	scimBaseURL = "https://www.worksapis.com/scim/v2"
)

func loadConfigAndToken() (*config.Config, *auth.Token, string, error) {
	profile, name, err := loadActiveConfig()
	if err != nil {
		return nil, nil, "", err
	}

	store := auth.NewProfileTokenStore(auth.DefaultTokenPath(), name)
	token, err := store.Load()
	if err != nil {
		return nil, nil, "", err
	}
	if token == nil {
		return nil, nil, "", fmt.Errorf("로그인되어 있지 않습니다. naverworks auth login을 실행하세요")
	}
	return profile, token, name, nil
}

func refreshJWTTokenFromAssertion(cfg *config.Config, t *auth.Token, store *auth.ProfileTokenStore) error {
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

func buildAPIClient(cfg *config.Config, token *auth.Token, activeProfileName string) *api.Client {
	refreshFn := func(t *auth.Token) error {
		store := auth.NewProfileTokenStore(auth.DefaultTokenPath(), activeProfileName)

		if t.RefreshToken != "" {
			if err := auth.RefreshAccessToken(authBaseURL, cfg.ClientID, cfg.ClientSecret, t); err == nil {
				return store.Save(t)
			}
		}

		if t.AuthMethod == auth.AuthMethodJWT {
			return refreshJWTTokenFromAssertion(cfg, t, store)
		}

		return fmt.Errorf("토큰 갱신 불가")
	}
	return api.NewClient(apiBaseURL, token, refreshFn)
}

func newAPIClient() (*api.Client, *config.Config, *auth.Token, error) {
	cfg, token, name, err := loadConfigAndToken()
	if err != nil {
		return nil, nil, nil, err
	}
	return buildAPIClient(cfg, token, name), cfg, token, nil
}

func newAPIClientWithUser(cmd *cobra.Command) (*api.Client, string, error) {
	client, cfg, token, err := newAPIClient()
	if err != nil {
		return nil, "", err
	}
	userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
	if err != nil {
		return nil, "", err
	}
	return client, userID, nil
}

func newSvc[T any](constructor func(*api.Client) *T) (*T, error) {
	client, _, _, err := newAPIClient()
	if err != nil {
		return nil, err
	}
	return constructor(client), nil
}

func buildScimClient(cfg *config.Config) (*api.Client, error) {
	if cfg.ScimAccessToken == "" {
		return nil, fmt.Errorf("scim_access_token이 설정되지 않았습니다. naverworks config set scim_access_token <token>")
	}
	token := &auth.Token{
		AuthMethod:  auth.AuthMethodSCIM,
		AccessToken: cfg.ScimAccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
	}
	return api.NewClient(scimBaseURL, token, nil), nil
}

func resolveUserID(cmd *cobra.Command, defaultUID string, authMethod auth.AuthMethod) (string, error) {
	userID, _ := cmd.Flags().GetString("user-id")
	if userID == "" {
		userID = defaultUID
	}
	if userID == "" {
		return "", fmt.Errorf("--user-id를 지정하거나 config에서 기본값을 설정하세요")
	}
	if userID == "me" && authMethod == auth.AuthMethodJWT {
		return "", fmt.Errorf("JWT 모드에서는 --user-id me를 사용할 수 없습니다. 명시적 userId를 지정하세요")
	}
	return userID, nil
}

func paginateAndPrint(fetch api.FetchFunc, key string, formatter *output.Formatter) error {
	items, err := api.PaginateAll(fetch, key)
	if err != nil {
		return err
	}
	merged, err := json.Marshal(map[string]json.RawMessage{key: items})
	if err != nil {
		return fmt.Errorf("결과 직렬화 실패: %w", err)
	}
	formatter.PrintRaw(merged)
	return nil
}

func printResponse(resp *api.Response) {
	if len(resp.Body) == 0 || strings.TrimSpace(string(resp.Body)) == "" {
		fmt.Println("{}")
	} else {
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
	}
}

func printBody(body []byte) {
	output.NewFormatter(outputFormat, os.Stdout).PrintRaw(body)
}

func fetchAndPrint(fn func(*api.Client) (*api.Response, error)) error {
	client, _, _, err := newAPIClient()
	if err != nil {
		return err
	}
	resp, err := fn(client)
	if err != nil {
		return err
	}
	printBody(resp.Body)
	return nil
}

func runListCmd(cmd *cobra.Command, columns []string, itemKey string, fetch func(string, int) (*api.Response, error)) error {
	cursor, _ := cmd.Flags().GetString("cursor")
	count, _ := cmd.Flags().GetInt("count")
	all, _ := cmd.Flags().GetBool("all")

	formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable(columns, itemKey)

	if all {
		return paginateAndPrint(func(c string) (*api.Response, error) {
			return fetch(c, count)
		}, itemKey, formatter)
	}

	resp, err := fetch(cursor, count)
	if err != nil {
		return err
	}
	formatter.PrintRaw(resp.Body)
	return nil
}

func loadActiveConfig() (*config.Config, string, error) {
	pc, err := config.LoadProfileConfig(config.DefaultPath())
	if err != nil {
		return nil, "", err
	}
	profile, name, err := pc.ActiveProfile(profileName)
	if err != nil {
		return nil, "", err
	}
	profile.ApplyEnvOverrides()
	return profile, name, nil
}

func requireTitleBodyPost(cmd *cobra.Command) (map[string]interface{}, error) {
	title, _ := cmd.Flags().GetString("title")
	body, _ := cmd.Flags().GetString("body")
	if title == "" {
		return nil, fmt.Errorf("--title은 필수입니다")
	}
	if body == "" {
		return nil, fmt.Errorf("--body는 필수입니다")
	}
	return map[string]interface{}{"title": title, "body": body}, nil
}

func parseOptionalJSONData(cmd *cobra.Command) (map[string]interface{}, error) {
	data, _ := cmd.Flags().GetString("data")
	if data == "" {
		return nil, nil
	}
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(data), &body); err != nil {
		return nil, fmt.Errorf("--data JSON 파싱 실패: %w", err)
	}
	return body, nil
}

func printDownloadURL(downloadURL string) {
	result, err := json.Marshal(map[string]string{"download_url": downloadURL})
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"error":{"code":"marshal_error","description":"%s"}}`+"\n", err)
		return
	}
	printBody(result)
}

func addListFlags(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
		c.Flags().Bool("all", false, "전체 페이지 자동 순회")
	}
}

const maxStdinSize int64 = 1 << 20 // 1MB

// readStdinLimited reads from the given reader up to maxBytes.
// Returns an explicit error if the input exceeds maxBytes.
func readStdinLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("stdin 읽기 실패: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("stdin 입력이 너무 큽니다 (최대 %d bytes)", maxBytes)
	}
	return data, nil
}

// readJSONFlagRaw reads the --json flag value as raw bytes.
// If the value is "-", it reads from stdin (limited to maxStdinSize).
// It validates that the content is valid JSON.
func readJSONFlagRaw(cmd *cobra.Command) ([]byte, error) {
	val, _ := cmd.Flags().GetString("json")
	if val == "" {
		return nil, fmt.Errorf("--json 플래그가 필요합니다")
	}

	var data []byte
	if val == "-" {
		var err error
		data, err = readStdinLimited(os.Stdin, maxStdinSize)
		if err != nil {
			return nil, err
		}
	} else {
		data = []byte(val)
	}

	if !json.Valid(data) {
		return nil, fmt.Errorf("유효하지 않은 JSON입니다")
	}
	return data, nil
}

// readJSONFlag reads the --json flag value and parses it into a map.
// If the value is "-", it reads from stdin.
func readJSONFlag(cmd *cobra.Command) (map[string]interface{}, error) {
	data, err := readJSONFlagRaw(cmd)
	if err != nil {
		return nil, err
	}
	var body map[string]interface{}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
	}
	return body, nil
}

const maxDefaultFileSize int64 = 10 << 20 // 10MB

// readFileFlagWithLimit reads the file from the given flag, rejecting files
// larger than maxBytes. Only regular files are accepted (no FIFO, devices, etc.).
// File type is checked via os.Stat before opening to avoid blocking on FIFOs.
// Returns file contents and base filename.
func readFileFlagWithLimit(cmd *cobra.Command, flagName string, maxBytes int64) ([]byte, string, error) {
	filePath, _ := cmd.Flags().GetString(flagName)
	if filePath == "" {
		return nil, "", fmt.Errorf("--%s 플래그가 필요합니다", flagName)
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("파일 접근 실패 (%s): %w", filePath, err)
	}
	if !info.Mode().IsRegular() {
		return nil, "", fmt.Errorf("일반 파일만 허용합니다: %s", filePath)
	}
	if info.Size() > maxBytes {
		return nil, "", fmt.Errorf("파일 크기 초과: %s (%d bytes, 최대 %d bytes)", filePath, info.Size(), maxBytes)
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("파일 열기 실패 (%s): %w", filePath, err)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		return nil, "", fmt.Errorf("파일 읽기 실패 (%s): %w", filePath, err)
	}
	if int64(len(data)) > maxBytes {
		return nil, "", fmt.Errorf("파일 크기 초과: %s (%d bytes, 최대 %d bytes)", filePath, len(data), maxBytes)
	}
	return data, filepath.Base(filePath), nil
}

// readFileFlag reads the file path from the given flag and returns the file
// contents along with the base file name. Uses default 10MB limit.
func readFileFlag(cmd *cobra.Command, flagName string) ([]byte, string, error) {
	return readFileFlagWithLimit(cmd, flagName, maxDefaultFileSize)
}

// validateFromUntil parses from/until RFC3339 strings, validates the range
// (until >= from, max 31 days), and returns the parsed times.
func validateFromUntil(from, until string) (time.Time, time.Time, error) {
	fromTime, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("--from 형식 오류 (RFC3339): %w", err)
	}
	untilTime, err := time.Parse(time.RFC3339, until)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("--until 형식 오류 (RFC3339): %w", err)
	}
	if untilTime.Before(fromTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("--from이 --until보다 이후입니다")
	}
	if untilTime.Sub(fromTime) > 31*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("--from과 --until 간격은 최대 31일입니다")
	}
	return fromTime, untilTime, nil
}

// statFileForUpload returns the base file name and size for a local file path.
// Used by presigned-upload commands across drive, bot, contact, directory, and approval.
func statFileForUpload(localPath string) (fileName string, fileSize int64, err error) {
	stat, err := os.Stat(localPath)
	if err != nil {
		return "", 0, fmt.Errorf("파일 정보 조회 실패: %w", err)
	}
	return filepath.Base(localPath), stat.Size(), nil
}

// doUploadFromResponse extracts the uploadUrl from a JSON response body
// and uploads the local file to it.
func doUploadFromResponse(client *api.Client, respBody []byte, localPath string) error {
	var result struct {
		UploadURL string `json:"uploadUrl"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("업로드 URL 파싱 실패: %w", err)
	}
	if result.UploadURL == "" {
		return fmt.Errorf("업로드 URL을 받지 못했습니다")
	}
	return client.UploadFile(result.UploadURL, localPath)
}

func resolveOrCreateProfile(pc *config.ProfileConfig) (*config.Config, string) {
	_, name, err := pc.ActiveProfile(profileName)
	if err != nil {
		name = profileName
		if name == "" {
			name = pc.CurrentProfile
			if name == "" {
				name = "default"
			}
		}
		pc.EnsureProfile(name)
	}
	return pc.Profiles[name], name
}
