package cmd

import (
	"encoding/json"
	"fmt"
	"os"
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

func buildAPIClient(cfg *config.Config, token *auth.Token, activeProfileName string) *api.Client {
	refreshFn := func(t *auth.Token) error {
		store := auth.NewProfileTokenStore(auth.DefaultTokenPath(), activeProfileName)

		if t.AuthMethod == "oauth" && t.RefreshToken != "" {
			if err := auth.RefreshAccessToken(authBaseURL, cfg.ClientID, cfg.ClientSecret, t); err == nil {
				return store.Save(t)
			}
		}

		if t.AuthMethod == "jwt" {
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

func buildScimClient(cfg *config.Config) (*api.Client, error) {
	if cfg.ScimAccessToken == "" {
		return nil, fmt.Errorf("scim_access_token이 설정되지 않았습니다. naverworks config set scim_access_token <token>")
	}
	token := &auth.Token{
		AuthMethod:  "scim",
		AccessToken: cfg.ScimAccessToken,
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(365 * 24 * time.Hour),
	}
	return api.NewClient(scimBaseURL, token, nil), nil
}

func resolveUserID(cmd *cobra.Command, defaultUID string, authMethod string) (string, error) {
	userID, _ := cmd.Flags().GetString("user-id")
	if userID == "" {
		userID = defaultUID
	}
	if userID == "" {
		return "", fmt.Errorf("--user-id를 지정하거나 config에서 기본값을 설정하세요")
	}
	if userID == "me" && authMethod == "jwt" {
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
	result, _ := json.Marshal(map[string]string{"download_url": downloadURL})
	printBody(result)
}

func addListFlags(cmds ...*cobra.Command) {
	for _, c := range cmds {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
		c.Flags().Bool("all", false, "전체 페이지 자동 순회")
	}
}
