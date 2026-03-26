package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Bot 메시지 관리",
}

var botSendCmd = &cobra.Command{
	Use:   "send",
	Short: "메시지 전송",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		botID, err := resolveBotID(cmd, cfg.BotID)
		if err != nil {
			return err
		}

		client := buildAPIClient(cfg, token, name)
		bot := api.NewBotService(client)

		to, _ := cmd.Flags().GetString("to")
		channel, _ := cmd.Flags().GetString("channel")
		text, _ := cmd.Flags().GetString("text")

		if text == "-" {
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			text = strings.Join(lines, "\n")
		}

		if to == "" && channel == "" {
			return fmt.Errorf("--to 또는 --channel 중 하나를 지정하세요")
		}
		if to != "" && channel != "" {
			return fmt.Errorf("--to와 --channel은 동시에 지정할 수 없습니다")
		}
		if text == "" {
			return fmt.Errorf("--text를 지정하세요")
		}

		var resp *api.Response
		if to != "" {
			resp, err = bot.SendTextToUser(botID, to, text)
		} else {
			resp, err = bot.SendTextToChannel(botID, channel, text)
		}
		if err != nil {
			return err
		}

		printResponse(resp)
		return nil
	},
}

var botGetChannelCmd = &cobra.Command{
	Use:   "get-channel <channelId>",
	Short: "채널 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		botID, err := resolveBotID(cmd, cfg.BotID)
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		bot := api.NewBotService(client)

		resp, err := bot.GetChannel(botID, args[0])
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
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		botID, err := resolveBotID(cmd, cfg.BotID)
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		bot := api.NewBotService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		if all {
			formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"userId"}, "members")
			return paginateAndPrint(func(c string) (*api.Response, error) {
				return bot.ListChannelMembers(botID, args[0], c, count)
			}, "members", formatter)
		}

		resp, err := bot.ListChannelMembers(botID, args[0], cursor, count)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"userId"}, "members").PrintRaw(resp.Body)
		return nil
	},
}

func resolveBotID(cmd *cobra.Command, cfgBotID string) (string, error) {
	flagBotID, _ := cmd.Flags().GetString("bot-id")
	if flagBotID != "" {
		return flagBotID, nil
	}
	if cfgBotID != "" {
		return cfgBotID, nil
	}
	return "", fmt.Errorf("bot_id가 설정되지 않았습니다. --bot-id 플래그 또는 naverworks config set bot_id <id>")
}

func init() {
	botCmd.PersistentFlags().String("bot-id", "", "Bot ID (config 기본값 오버라이드)")

	botSendCmd.Flags().String("to", "", "수신자 userId")
	botSendCmd.Flags().String("channel", "", "채널 ID")
	botSendCmd.Flags().String("text", "", "메시지 텍스트 (- 이면 stdin)")

	botChannelMembersCmd.Flags().String("cursor", "", "페이지네이션 커서")
	botChannelMembersCmd.Flags().Int("count", 0, "페이지 크기")
	botChannelMembersCmd.Flags().Bool("all", false, "전체 페이지 자동 순회")

	botCmd.AddCommand(botSendCmd, botGetChannelCmd, botChannelMembersCmd)
	rootCmd.AddCommand(botCmd)
}
