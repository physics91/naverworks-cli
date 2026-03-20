package cmd

import (
	"bufio"
	"encoding/json"
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
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		if cfg.BotID == "" {
			return fmt.Errorf("bot_id가 설정되지 않았습니다. nw-cli config set bot_id <id>")
		}

		client := buildAPIClient(cfg, token)
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
			resp, err = bot.SendTextToUser(cfg.BotID, to, text)
		} else {
			resp, err = bot.SendTextToChannel(cfg.BotID, channel, text)
		}
		if err != nil {
			return err
		}

		if len(resp.Body) == 0 || strings.TrimSpace(string(resp.Body)) == "" {
			fmt.Println("{}")
		} else {
			output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		}
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
		all, _ := cmd.Flags().GetBool("all")

		if all {
			var allMembers []json.RawMessage
			for {
				resp, err := bot.ListChannelMembers(cfg.BotID, args[0], cursor, count)
				if err != nil {
					return err
				}
				var page struct {
					Members          []json.RawMessage `json:"members"`
					ResponseMetaData struct {
						NextCursor string `json:"nextCursor"`
					} `json:"responseMetaData"`
				}
				json.Unmarshal(resp.Body, &page)
				allMembers = append(allMembers, page.Members...)
				if page.ResponseMetaData.NextCursor == "" {
					break
				}
				cursor = page.ResponseMetaData.NextCursor
			}
			formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"userId"}, "members")
			merged := map[string]interface{}{"members": allMembers}
			data, _ := json.Marshal(merged)
			formatter.PrintRaw(data)
			return nil
		}

		resp, err := bot.ListChannelMembers(cfg.BotID, args[0], cursor, count)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"userId"}, "members").PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	botSendCmd.Flags().String("to", "", "수신자 userId")
	botSendCmd.Flags().String("channel", "", "채널 ID")
	botSendCmd.Flags().String("text", "", "메시지 텍스트 (- 이면 stdin)")

	botChannelMembersCmd.Flags().String("cursor", "", "페이지네이션 커서")
	botChannelMembersCmd.Flags().Int("count", 0, "페이지 크기")
	botChannelMembersCmd.Flags().Bool("all", false, "전체 페이지 자동 순회")

	botCmd.AddCommand(botSendCmd, botGetChannelCmd, botChannelMembersCmd)
	rootCmd.AddCommand(botCmd)
}
