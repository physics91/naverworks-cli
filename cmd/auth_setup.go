package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/spf13/cobra"
)

var authSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "대화형 인증 설정",
	Long:  "질문-답변 형식으로 네이버웍스 인증에 필요한 설정을 구성합니다.",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)
		path := config.DefaultPath()
		pc, err := config.LoadProfileConfig(path)
		if err != nil {
			pc = config.NewProfileConfig()
			pc.EnsureProfile("default")
		}

		_, name, err := pc.ActiveProfile(profileName)
		if err != nil {
			// Profile doesn't exist yet, create it
			name = profileName
			if name == "" {
				name = pc.CurrentProfile
				if name == "" {
					name = "default"
				}
			}
			pc.EnsureProfile(name)
		}
		cfg := pc.Profiles[name]

		fmt.Println("네이버웍스 CLI 인증 설정을 시작합니다.")
		fmt.Println()

		// Step 1: 인증 방식 선택
		authMethod := prompt(reader, "인증 방식을 선택하세요 [oauth/jwt]", defaultVal(cfg.ServiceAccountID != "", "jwt", "oauth"))
		authMethod = strings.ToLower(strings.TrimSpace(authMethod))
		if authMethod != "oauth" && authMethod != "jwt" {
			return fmt.Errorf("유효하지 않은 인증 방식: %s (oauth 또는 jwt)", authMethod)
		}

		// Step 2: Client ID
		cfg.ClientID = prompt(reader, "Client ID", cfg.ClientID)
		if cfg.ClientID == "" {
			return fmt.Errorf("Client ID는 필수입니다")
		}

		// Step 3: Client Secret (마스킹 표시)
		currentSecret := ""
		if cfg.ClientSecret != "" {
			currentSecret = "****"
		}
		newSecret := prompt(reader, "Client Secret (입력하면 덮어씀, 빈 값이면 유지)", currentSecret)
		if newSecret != "" && newSecret != "****" {
			cfg.ClientSecret = newSecret
		}
		if cfg.ClientSecret == "" {
			return fmt.Errorf("Client Secret은 필수입니다")
		}

		// Step 4: JWT 전용 설정
		if authMethod == "jwt" {
			cfg.ServiceAccountID = prompt(reader, "Service Account ID", cfg.ServiceAccountID)
			if cfg.ServiceAccountID == "" {
				return fmt.Errorf("JWT 인증에는 Service Account ID가 필수입니다")
			}

			cfg.PrivateKeyPath = prompt(reader, "Private Key 파일 경로", cfg.PrivateKeyPath)
			if cfg.PrivateKeyPath == "" {
				return fmt.Errorf("JWT 인증에는 Private Key 경로가 필수입니다")
			}
		}

		// Step 5: Bot ID (선택)
		cfg.BotID = prompt(reader, "Bot ID (선택, 엔터로 건너뛰기)", cfg.BotID)

		// Step 6: Scope (선택)
		defaultScope := defaultOAuthScope
		if authMethod == "jwt" {
			defaultScope = defaultJWTScope
		}
		if cfg.Scope == "" {
			cfg.Scope = ""
		}
		scopeInput := prompt(reader, fmt.Sprintf("Scope (기본값: %s)", defaultScope), cfg.Scope)
		if scopeInput != "" {
			cfg.Scope = scopeInput
		}

		// Step 7: Calendar User ID (선택)
		cfg.DefaultCalendarUserID = prompt(reader, "기본 Calendar User ID (선택, OAuth면 'me' 가능)", cfg.DefaultCalendarUserID)

		// 저장
		if err := pc.Save(path); err != nil {
			return fmt.Errorf("설정 저장 실패: %w", err)
		}

		fmt.Println()
		fmt.Printf("설정이 저장되었습니다: %s\n", path)
		fmt.Println()

		// Step 8: 바로 로그인 할지
		doLogin := prompt(reader, "지금 바로 로그인하시겠습니까? [Y/n]", "Y")
		doLogin = strings.ToLower(strings.TrimSpace(doLogin))
		if doLogin == "" || doLogin == "y" || doLogin == "yes" {
			fmt.Println()
			store := auth.NewProfileTokenStore(auth.DefaultTokenPath(), name)
			if authMethod == "jwt" {
				fmt.Println("JWT 인증을 시작합니다...")
				return loginJWT(cfg, store)
			}
			fmt.Println("OAuth 인증을 시작합니다. 브라우저가 열립니다...")
			return loginOAuth(cfg, store)
		}

		fmt.Println()
		fmt.Println("설정 완료. naverworks auth login 으로 로그인하세요.")
		return nil
	},
}

func prompt(reader *bufio.Reader, question string, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", question, defaultVal)
	} else {
		fmt.Printf("  %s: ", question)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

func defaultVal(condition bool, ifTrue, ifFalse string) string {
	if condition {
		return ifTrue
	}
	return ifFalse
}

func init() {
	authCmd.AddCommand(authSetupCmd)
}
