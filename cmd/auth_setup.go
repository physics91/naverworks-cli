package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const setupBotListCount = 100

type setupBotOption struct {
	BotID string
	Label string
}

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
				if err := loginJWT(cfg, store); err != nil {
					return err
				}
			} else {
				fmt.Println("OAuth 인증을 시작합니다. 브라우저가 열립니다...")
				if err := loginOAuth(cfg, store); err != nil {
					return err
				}
			}
			return runPostLoginBotSelection(
				reader,
				os.Stdout,
				cfg,
				func() ([]setupBotOption, error) {
					return fetchSetupBotsForProfile(cfg, name, store)
				},
				func() error {
					return pc.Save(path)
				},
			)
		}

		fmt.Println()
		fmt.Println("설정 완료. naverworks auth login 으로 로그인하세요.")
		return nil
	},
}

func prompt(reader *bufio.Reader, question string, defaultVal string) string {
	return promptWithWriter(reader, os.Stdout, question, defaultVal)
}

func promptWithWriter(reader *bufio.Reader, out io.Writer, question string, defaultVal string) string {
	if defaultVal != "" {
		fmt.Fprintf(out, "  %s [%s]: ", question, defaultVal)
	} else {
		fmt.Fprintf(out, "  %s: ", question)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

func shouldReselectBot(existingBotID string, answer string) bool {
	if strings.TrimSpace(existingBotID) == "" {
		return true
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func needsBotConfigSave(before string, after string) bool {
	return strings.TrimSpace(before) != strings.TrimSpace(after)
}

func parseSetupBots(body []byte) ([]setupBotOption, error) {
	var payload struct {
		Bots []struct {
			BotID   string `json:"botId"`
			BotName string `json:"botName"`
		} `json:"bots"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("bot 목록 파싱 실패: %w", err)
	}

	options := make([]setupBotOption, 0, len(payload.Bots))
	for _, bot := range payload.Bots {
		if strings.TrimSpace(bot.BotID) == "" {
			continue
		}
		label := strings.TrimSpace(bot.BotName)
		if label == "" {
			label = bot.BotID
		}
		options = append(options, setupBotOption{
			BotID: bot.BotID,
			Label: label,
		})
	}

	return options, nil
}

func fetchSetupBots(fetch func() (*api.Response, error)) ([]setupBotOption, error) {
	resp, err := fetch()
	if err != nil {
		return nil, err
	}
	return parseSetupBots(resp.Body)
}

func fetchSetupBotsForProfile(cfg *config.Config, profileName string, store *auth.ProfileTokenStore) ([]setupBotOption, error) {
	token, err := store.Load()
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, fmt.Errorf("로그인 토큰을 찾을 수 없습니다")
	}

	client := buildAPIClient(cfg, token, profileName)
	botSvc := api.NewBotService(client)
	return fetchSetupBots(func() (*api.Response, error) {
		return botSvc.ListBots("", setupBotListCount)
	})
}

func chooseSetupBot(reader *bufio.Reader, out io.Writer, bots []setupBotOption, existingBotID string) (string, bool, error) {
	if len(bots) == 0 {
		manualBotID := promptWithWriter(reader, out, "Bot ID 직접 입력 (엔터로 건너뛰기)", "")
		if manualBotID == "" {
			return existingBotID, false, nil
		}
		return manualBotID, needsBotConfigSave(existingBotID, manualBotID), nil
	}

	if len(bots) == 1 {
		fmt.Fprintf(out, "Bot 자동 선택: %s (%s)\n", bots[0].Label, bots[0].BotID)
		return bots[0].BotID, needsBotConfigSave(existingBotID, bots[0].BotID), nil
	}

	fmt.Fprintln(out, "조회된 Bot 목록:")
	for idx, bot := range bots {
		fmt.Fprintf(out, "  %d) %s (%s)\n", idx+1, bot.Label, bot.BotID)
	}

	for {
		fmt.Fprint(out, "  번호 선택, m=직접 입력, Enter=건너뛰기: ")
		input, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", false, fmt.Errorf("bot 선택 입력 실패: %w", err)
		}
		input = strings.TrimSpace(input)

		switch {
		case input == "":
			return existingBotID, false, nil
		case strings.EqualFold(input, "m"):
			manualBotID := promptWithWriter(reader, out, "Bot ID 직접 입력 (엔터로 건너뛰기)", "")
			if manualBotID == "" {
				return existingBotID, false, nil
			}
			return manualBotID, needsBotConfigSave(existingBotID, manualBotID), nil
		default:
			choice, convErr := strconv.Atoi(input)
			if convErr != nil || choice < 1 || choice > len(bots) {
				fmt.Fprintln(out, "다시 입력하세요. 유효한 번호를 고르거나 m/Enter를 사용하면 됩니다")
				continue
			}
			selected := bots[choice-1].BotID
			return selected, needsBotConfigSave(existingBotID, selected), nil
		}
	}
}

func runPostLoginBotSelection(
	reader *bufio.Reader,
	out io.Writer,
	cfg *config.Config,
	fetch func() ([]setupBotOption, error),
	save func() error,
) error {
	currentBotID := cfg.BotID
	if strings.TrimSpace(currentBotID) != "" {
		answer := promptWithWriter(reader, out, "현재 Bot ID가 설정되어 있음. 다시 선택할까? [y/N]", "N")
		if !shouldReselectBot(currentBotID, answer) {
			return nil
		}
	}

	bots, err := fetch()
	if err != nil {
		fmt.Fprintf(out, "경고: Bot 목록 조회 실패: %v\n", err)
		fmt.Fprintln(out, "bot scope가 포함됐는지 확인하고, 필요하면 Bot ID를 직접 입력하면 됨")
		manualBotID := promptWithWriter(reader, out, "Bot ID 직접 입력 (엔터로 건너뛰기)", "")
		if manualBotID == "" {
			return nil
		}
		if !needsBotConfigSave(currentBotID, manualBotID) {
			return nil
		}
		cfg.BotID = manualBotID
		if err := save(); err != nil {
			return fmt.Errorf("로그인은 성공했지만 bot_id 저장 실패: %w", err)
		}
		fmt.Fprintf(out, "Bot ID 저장됨: %s\n", cfg.BotID)
		return nil
	}

	selectedBotID, changed, err := chooseSetupBot(reader, out, bots, currentBotID)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}

	cfg.BotID = selectedBotID
	if err := save(); err != nil {
		return fmt.Errorf("로그인은 성공했지만 bot_id 저장 실패: %w", err)
	}
	fmt.Fprintf(out, "Bot ID 저장됨: %s\n", cfg.BotID)
	return nil
}

func init() {
	authCmd.AddCommand(authSetupCmd)
}
