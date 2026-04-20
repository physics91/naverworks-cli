package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Bot 메시지 관리",
}

type botServiceIDCall func(*api.BotService, string) (*api.Response, error)
type botServiceBodyCall func(*api.BotService, []byte) (*api.Response, error)
type botServiceIDBodyCall func(*api.BotService, string, []byte) (*api.Response, error)
type botServiceListCall func(*api.BotService, string, int) (*api.Response, error)
type botScopedCall func(*api.BotService, string) (*api.Response, error)
type botScopedIDCall func(*api.BotService, string, string) (*api.Response, error)
type botScopedTwoIDCall func(*api.BotService, string, string, string) (*api.Response, error)
type botScopedBodyCall func(*api.BotService, string, []byte) (*api.Response, error)
type botScopedIDBodyCall func(*api.BotService, string, string, []byte) (*api.Response, error)
type botScopedListCall func(*api.BotService, string, string, int) (*api.Response, error)
type botScopedIDListCall func(*api.BotService, string, string, string, int) (*api.Response, error)

// Keep these wrappers local so cmd/helpers.go does not grow a bot-only helper family.
func printBotBody(resp *api.Response) {
	printBody(resp.Body)
}

func botServiceIDRunE(call botServiceIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBotService)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0])
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func botServiceBodyRunE(call botServiceBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBotService)
		if err != nil {
			return err
		}
		data, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := call(svc, data)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func botServiceIDBodyRunE(call botServiceIDBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBotService)
		if err != nil {
			return err
		}
		data, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0], data)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func botServiceListRunListE(columns []string, itemKey string, call botServiceListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewBotService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(svc, cursor, count)
		})
	}
}

func botScopedRunE(call botScopedCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}
		resp, err := call(bot, botID)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func botScopedIDRunE(call botScopedIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}
		resp, err := call(bot, botID, args[0])
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func botScopedTwoIDRunE(call botScopedTwoIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}
		resp, err := call(bot, botID, args[0], args[1])
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func botScopedBodyRunE(call botScopedBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}
		data, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := call(bot, botID, data)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func botScopedIDBodyRunE(call botScopedIDBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}
		data, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := call(bot, botID, args[0], data)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func botScopedListRunListE(columns []string, itemKey string, call botScopedListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(bot, botID, cursor, count)
		})
	}
}

func botScopedIDRunListE(columns []string, itemKey string, call botScopedIDListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(bot, botID, args[0], cursor, count)
		})
	}
}

// ─── Bot CRUD (Task 3-1) ───

var botListCmd = &cobra.Command{
	Use:   "list",
	Short: "Bot 목록 조회",
	RunE:  botServiceListRunListE([]string{"botId", "botName"}, "bots", (*api.BotService).ListBots),
}

var botGetCmd = &cobra.Command{
	Use:   "get <botId>",
	Short: "Bot 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  botServiceIDRunE((*api.BotService).GetBot, printBotBody),
}

var botCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Bot 생성",
	RunE:  botServiceBodyRunE((*api.BotService).CreateBot, printResponse),
}

var botUpdateCmd = &cobra.Command{
	Use:   "update <botId>",
	Short: "Bot 수정 (전체)",
	Args:  cobra.ExactArgs(1),
	RunE:  botServiceIDBodyRunE((*api.BotService).UpdateBot, printResponse),
}

var botPatchCmd = &cobra.Command{
	Use:   "patch <botId>",
	Short: "Bot 수정 (부분)",
	Args:  cobra.ExactArgs(1),
	RunE:  botServiceIDBodyRunE((*api.BotService).PatchBot, printResponse),
}

var botDeleteCmd = &cobra.Command{
	Use:   "delete <botId>",
	Short: "Bot 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  botServiceIDRunE((*api.BotService).DeleteBot, printResponse),
}

var botRegenerateSecretCmd = &cobra.Command{
	Use:   "regenerate-secret <botId>",
	Short: "Bot Secret 재생성",
	Args:  cobra.ExactArgs(1),
	RunE:  botServiceIDRunE((*api.BotService).RegenerateSecret, printResponse),
}

// ─── Messages (Task 3-2) ───

var botSendCmd = &cobra.Command{
	Use:   "send",
	Short: "메시지 전송",
	RunE: func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}

		to, _ := cmd.Flags().GetString("to")
		channel, _ := cmd.Flags().GetString("channel")
		text, _ := cmd.Flags().GetString("text")
		jsonStr, _ := cmd.Flags().GetString("json")

		if text == "-" {
			data, err := readStdinLimited(os.Stdin, maxStdinSize)
			if err != nil {
				return err
			}
			text = strings.TrimRight(string(data), "\n")
		}

		if to == "" && channel == "" {
			return fmt.Errorf("--to 또는 --channel 중 하나를 지정하세요")
		}
		if to != "" && channel != "" {
			return fmt.Errorf("--to와 --channel은 동시에 지정할 수 없습니다")
		}
		if text != "" && jsonStr != "" {
			return fmt.Errorf("--text와 --json은 동시에 지정할 수 없습니다")
		}
		if text == "" && jsonStr == "" {
			return fmt.Errorf("--text 또는 --json 중 하나를 지정하세요")
		}

		var resp *api.Response
		if jsonStr != "" {
			data, err := readJSONFlagRaw(cmd)
			if err != nil {
				return err
			}
			if to != "" {
				resp, err = bot.SendMessageToUser(botID, to, data)
			} else {
				resp, err = bot.SendMessageToChannel(botID, channel, data)
			}
		} else {
			if to != "" {
				resp, err = bot.SendTextToUser(botID, to, text)
			} else {
				resp, err = bot.SendTextToChannel(botID, channel, text)
			}
		}
		if err != nil {
			return err
		}

		printResponse(resp)
		return nil
	},
}

// ─── Attachments (Task 3-2) ───

var botCreateAttachmentCmd = &cobra.Command{
	Use:   "create-attachment",
	Short: "Bot 첨부파일 생성 (presigned URL)",
	RunE: func(cmd *cobra.Command, args []string) error {
		bot, client, botID, err := newBotSvcWithClient(cmd)
		if err != nil {
			return err
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file 플래그가 필요합니다")
		}
		fileName, fileSize, err := statFileForUpload(filePath)
		if err != nil {
			return err
		}

		uploadBody := map[string]interface{}{
			"fileName": fileName,
			"fileSize": fileSize,
		}
		resp, err := bot.CreateAttachment(botID, uploadBody)
		if err != nil {
			return err
		}

		respBody, err := doUploadFromResponse(client, resp.Body, filePath)
		if err != nil {
			return err
		}
		printBody(respBody)
		return nil
	},
}

var botGetAttachmentCmd = &cobra.Command{
	Use:   "get-attachment <fileId>",
	Short: "Bot 첨부파일 다운로드 URL 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}
		downloadURL, err := bot.GetAttachmentDownloadUrl(botID, args[0])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

// ─── Channels (Task 3-3) ───

var botGetChannelCmd = &cobra.Command{
	Use:   "get-channel <channelId>",
	Short: "채널 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDRunE((*api.BotService).GetChannel, printBotBody),
}

var botChannelMembersCmd = &cobra.Command{
	Use:   "channel-members <channelId>",
	Short: "채널 멤버 목록",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDRunListE([]string{"userId"}, "members", (*api.BotService).ListChannelMembers),
}

var botCreateChannelCmd = &cobra.Command{
	Use:   "create-channel",
	Short: "채널 생성",
	RunE:  botScopedBodyRunE((*api.BotService).CreateChannel, printResponse),
}

var botLeaveChannelCmd = &cobra.Command{
	Use:   "leave-channel <channelId>",
	Short: "채널 나가기",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDRunE((*api.BotService).LeaveChannel, printResponse),
}

// ─── Domain (Task 3-4) ───

var botDomainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Bot 도메인 관리",
}

var botDomainRegisterCmd = &cobra.Command{
	Use:   "register <domainId>",
	Short: "도메인 등록",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDBodyRunE((*api.BotService).RegisterDomain, printResponse),
}

var botDomainListCmd = &cobra.Command{
	Use:   "list",
	Short: "도메인 목록 조회",
	RunE:  botScopedListRunListE([]string{"domainId"}, "domains", (*api.BotService).ListDomains),
}

var botDomainUpdateCmd = &cobra.Command{
	Use:   "update <domainId>",
	Short: "도메인 수정 (전체)",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDBodyRunE((*api.BotService).UpdateDomain, printResponse),
}

var botDomainPatchCmd = &cobra.Command{
	Use:   "patch <domainId>",
	Short: "도메인 수정 (부분)",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDBodyRunE((*api.BotService).PatchDomain, printResponse),
}

var botDomainDeleteCmd = &cobra.Command{
	Use:   "delete <domainId>",
	Short: "도메인 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDRunE((*api.BotService).DeleteDomain, printResponse),
}

var botDomainAddMembersCmd = &cobra.Command{
	Use:   "add-members <domainId>",
	Short: "도메인 멤버 추가",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDBodyRunE((*api.BotService).AddDomainMembers, printResponse),
}

var botDomainListMembersCmd = &cobra.Command{
	Use:   "list-members <domainId>",
	Short: "도메인 멤버 목록",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDRunListE([]string{"userId"}, "members", (*api.BotService).ListDomainMembers),
}

var botDomainRemoveMemberCmd = &cobra.Command{
	Use:   "remove-member <domainId> <userId>",
	Short: "도메인 멤버 삭제",
	Args:  cobra.ExactArgs(2),
	RunE:  botScopedTwoIDRunE((*api.BotService).RemoveDomainMember, printResponse),
}

// ─── Persistent Menu (Task 3-5) ───

var botPersistentMenuCmd = &cobra.Command{
	Use:   "persistent-menu",
	Short: "Bot 고정메뉴 관리",
}

var botPersistentMenuSetCmd = &cobra.Command{
	Use:   "set",
	Short: "고정메뉴 설정",
	RunE:  botScopedBodyRunE((*api.BotService).UpsertPersistentMenu, printResponse),
}

var botPersistentMenuGetCmd = &cobra.Command{
	Use:   "get",
	Short: "고정메뉴 조회",
	RunE:  botScopedRunE((*api.BotService).GetPersistentMenu, printBotBody),
}

var botPersistentMenuDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "고정메뉴 삭제",
	RunE:  botScopedRunE((*api.BotService).DeletePersistentMenu, printResponse),
}

// ─── Rich Menu (Task 3-6) ───

var botRichMenuCmd = &cobra.Command{
	Use:   "richmenu",
	Short: "Bot 리치메뉴 관리",
}

var botRichMenuCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "리치메뉴 생성",
	RunE:  botScopedBodyRunE((*api.BotService).CreateRichMenu, printResponse),
}

var botRichMenuListCmd = &cobra.Command{
	Use:   "list",
	Short: "리치메뉴 목록 조회",
	RunE:  botScopedListRunListE([]string{"richMenuId"}, "richmenus", (*api.BotService).ListRichMenus),
}

var botRichMenuGetCmd = &cobra.Command{
	Use:   "get <richmenuId>",
	Short: "리치메뉴 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDRunE((*api.BotService).GetRichMenu, printBotBody),
}

var botRichMenuDeleteCmd = &cobra.Command{
	Use:   "delete <richmenuId>",
	Short: "리치메뉴 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDRunE((*api.BotService).DeleteRichMenu, printResponse),
}

var botRichMenuSetImageCmd = &cobra.Command{
	Use:   "set-image <richmenuId>",
	Short: "리치메뉴 이미지 설정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}
		data, fileName, err := readFileFlagWithLimit(cmd, "file", 3<<20) // 3MB: rich menu image limit
		if err != nil {
			return err
		}
		resp, err := bot.SetRichMenuImage(botID, args[0], "image", fileName, data)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var botRichMenuGetImageCmd = &cobra.Command{
	Use:   "get-image <richmenuId>",
	Short: "리치메뉴 이미지 다운로드 URL 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bot, botID, err := newBotSvc(cmd)
		if err != nil {
			return err
		}
		downloadURL, err := bot.GetRichMenuImage(botID, args[0])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

var botRichMenuSetUserCmd = &cobra.Command{
	Use:   "set-user <richmenuId> <userId>",
	Short: "사용자에게 리치메뉴 설정",
	Args:  cobra.ExactArgs(2),
	RunE:  botScopedTwoIDRunE((*api.BotService).SetUserRichMenu, printResponse),
}

var botRichMenuGetUserCmd = &cobra.Command{
	Use:   "get-user <userId>",
	Short: "사용자 리치메뉴 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDRunE((*api.BotService).GetUserRichMenu, printBotBody),
}

var botRichMenuDeleteUserCmd = &cobra.Command{
	Use:   "delete-user <userId>",
	Short: "사용자 리치메뉴 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDRunE((*api.BotService).DeleteUserRichMenu, printResponse),
}

var botRichMenuSetDefaultCmd = &cobra.Command{
	Use:   "set-default <richmenuId>",
	Short: "기본 리치메뉴 설정",
	Args:  cobra.ExactArgs(1),
	RunE:  botScopedIDRunE((*api.BotService).SetDefaultRichMenu, printResponse),
}

var botRichMenuGetDefaultCmd = &cobra.Command{
	Use:   "get-default",
	Short: "기본 리치메뉴 조회",
	RunE:  botScopedRunE((*api.BotService).GetDefaultRichMenu, printBotBody),
}

var botRichMenuDeleteDefaultCmd = &cobra.Command{
	Use:   "delete-default",
	Short: "기본 리치메뉴 삭제",
	RunE:  botScopedRunE((*api.BotService).DeleteDefaultRichMenu, printResponse),
}

// ─── Helpers ───

func newBotSvc(cmd *cobra.Command) (*api.BotService, string, error) {
	client, cfg, _, err := newAPIClient()
	if err != nil {
		return nil, "", err
	}
	botID, err := resolveBotID(cmd, cfg.BotID)
	if err != nil {
		return nil, "", err
	}
	return api.NewBotService(client), botID, nil
}

func newBotSvcWithClient(cmd *cobra.Command) (*api.BotService, *api.Client, string, error) {
	client, cfg, _, err := newAPIClient()
	if err != nil {
		return nil, nil, "", err
	}
	botID, err := resolveBotID(cmd, cfg.BotID)
	if err != nil {
		return nil, nil, "", err
	}
	return api.NewBotService(client), client, botID, nil
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

	// Bot CRUD flags
	botCreateCmd.Flags().String("json", "", "JSON body (- 이면 stdin)")
	botUpdateCmd.Flags().String("json", "", "JSON body (- 이면 stdin)")
	botPatchCmd.Flags().String("json", "", "JSON body (- 이면 stdin)")

	addListFlags(botListCmd)

	// Send flags
	botSendCmd.Flags().String("to", "", "수신자 userId")
	botSendCmd.Flags().String("channel", "", "채널 ID")
	botSendCmd.Flags().String("text", "", "메시지 텍스트 (- 이면 stdin)")
	botSendCmd.Flags().String("json", "", "구조화 메시지 JSON (--text와 배타적)")

	// Attachment flags
	botCreateAttachmentCmd.Flags().String("file", "", "업로드 파일 경로")

	// Channel flags
	botCreateChannelCmd.Flags().String("json", "", "JSON body (- 이면 stdin)")
	addListFlags(botChannelMembersCmd)

	// Domain flags
	botDomainRegisterCmd.Flags().String("json", "", "JSON body (- 이면 stdin)")
	botDomainUpdateCmd.Flags().String("json", "", "JSON body (- 이면 stdin)")
	botDomainPatchCmd.Flags().String("json", "", "JSON body (- 이면 stdin)")
	botDomainAddMembersCmd.Flags().String("json", "", "JSON body (- 이면 stdin)")
	addListFlags(botDomainListCmd, botDomainListMembersCmd)

	botDomainCmd.AddCommand(
		botDomainRegisterCmd, botDomainListCmd,
		botDomainUpdateCmd, botDomainPatchCmd, botDomainDeleteCmd,
		botDomainAddMembersCmd, botDomainListMembersCmd, botDomainRemoveMemberCmd,
	)

	// Persistent menu flags
	botPersistentMenuSetCmd.Flags().String("json", "", "JSON body (- 이면 stdin)")

	botPersistentMenuCmd.AddCommand(
		botPersistentMenuSetCmd, botPersistentMenuGetCmd, botPersistentMenuDeleteCmd,
	)

	// Rich menu flags
	botRichMenuCreateCmd.Flags().String("json", "", "JSON body (- 이면 stdin)")
	botRichMenuSetImageCmd.Flags().String("file", "", "이미지 파일 경로")
	addListFlags(botRichMenuListCmd)

	botRichMenuCmd.AddCommand(
		botRichMenuCreateCmd, botRichMenuListCmd,
		botRichMenuGetCmd, botRichMenuDeleteCmd,
		botRichMenuSetImageCmd, botRichMenuGetImageCmd,
		botRichMenuSetUserCmd, botRichMenuGetUserCmd, botRichMenuDeleteUserCmd,
		botRichMenuSetDefaultCmd, botRichMenuGetDefaultCmd, botRichMenuDeleteDefaultCmd,
	)

	// Register all to botCmd
	botCmd.AddCommand(
		botListCmd, botGetCmd, botCreateCmd, botUpdateCmd, botPatchCmd, botDeleteCmd,
		botRegenerateSecretCmd,
		botSendCmd,
		botCreateAttachmentCmd, botGetAttachmentCmd,
		botGetChannelCmd, botChannelMembersCmd, botCreateChannelCmd, botLeaveChannelCmd,
		botDomainCmd,
		botPersistentMenuCmd,
		botRichMenuCmd,
	)
	rootCmd.AddCommand(botCmd)
}
