package cmd

import (
	"fmt"
	"time"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/spf13/cobra"
)

const (
	apiBaseURL  = "https://www.worksapis.com/v1.0"
	scimBaseURL = "https://www.worksapis.com/scim/v2"
)

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
		store := auth.NewTokenStore(auth.DefaultTokenPath())

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
		return nil, fmt.Errorf("scim_access_token이 설정되지 않았습니다. nw-cli config set scim_access_token <token>")
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
