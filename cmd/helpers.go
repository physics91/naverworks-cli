package cmd

import (
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
)

const apiBaseURL = "https://www.worksapis.com/v1.0"

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
