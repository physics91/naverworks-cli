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
	authBaseURL       = "https://auth.worksmobile.com/oauth2/v2.0"
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
			if name, err := fetchUserName(token.AccessToken); err == nil && name != "" {
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
	for _, s := range splitFields(scope) {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func splitFields(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
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
