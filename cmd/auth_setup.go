package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// nonTTYErrorMessage returns the error message shown when auth setup
// is run in a non-interactive environment.
func nonTTYErrorMessage() string {
	return "auth setup은 대화형 터미널에서만 실행 가능합니다.\n" +
		"자동화 환경에서는 NW_CLIENT_SECRET 환경변수 또는 " +
		"'naverworks config set client_secret --stdin'을 사용하세요"
}

// promptSecret reads a secret value from the terminal without echoing.
// It handles SIGINT/SIGTERM to restore terminal state before exit.
func promptSecret(fd int, label string) (string, error) {
	oldState, err := term.GetState(fd)
	if err != nil {
		return "", fmt.Errorf("터미널 상태 저장 실패: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			term.Restore(fd, oldState)
			os.Exit(130) // 128 + SIGINT(2)
		case <-done:
			return
		}
	}()
	defer func() {
		signal.Stop(sigCh)
		close(done)
	}()
	defer term.Restore(fd, oldState)

	fmt.Fprintf(os.Stderr, "  %s: ", label)
	pass, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("비밀번호 읽기 실패: %w", err)
	}
	return strings.TrimSpace(string(pass)), nil
}

var authSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "대화형 인증 설정",
	Long:  "질문-답변 형식으로 네이버웍스 인증에 필요한 설정을 구성합니다.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fd := int(os.Stdin.Fd())
		if !term.IsTerminal(fd) {
			return errors.New(nonTTYErrorMessage())
		}

		reader := bufio.NewReader(os.Stdin)
		path := config.DefaultPath()
		pc, err := config.LoadProfileConfig(path)
		if err != nil {
			pc = config.NewProfileConfig()
			pc.EnsureProfile("default")
		}

		cfg, name := resolveOrCreateProfile(pc)

		fmt.Println("네이버웍스 CLI 인증 설정을 시작합니다.")
		fmt.Println()

		// Step 1: 인증 방식 선택
		defaultAuthMethod := "oauth"
		if cfg.ServiceAccountID != "" {
			defaultAuthMethod = "jwt"
		}
		authMethod := prompt(reader, "인증 방식을 선택하세요 [oauth/jwt]", defaultAuthMethod)
		authMethod = strings.ToLower(strings.TrimSpace(authMethod))
		if authMethod != "oauth" && authMethod != "jwt" {
			return fmt.Errorf("유효하지 않은 인증 방식: %s (oauth 또는 jwt)", authMethod)
		}

		// Step 2: Client ID
		cfg.ClientID = prompt(reader, "Client ID", cfg.ClientID)
		if cfg.ClientID == "" {
			return fmt.Errorf("Client ID는 필수입니다")
		}

		// Step 3: Client Secret (마스킹 입력)
		secretHint := "(비어 있음)"
		if cfg.ClientSecret != "" {
			secretHint = "(설정됨, 엔터로 유지)"
		}
		newSecret, err := promptSecret(fd, fmt.Sprintf("Client Secret %s", secretHint))
		if err != nil {
			return err
		}
		if newSecret != "" {
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

func init() {
	authCmd.AddCommand(authSetupCmd)
}
