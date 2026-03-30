package cmd

import (
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "메일 관리",
}

var mailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "메일 전송",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewMailService(client)

		to, _ := cmd.Flags().GetString("to")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")

		if to == "" || subject == "" || body == "" {
			return fmt.Errorf("--to, --subject, --body는 필수입니다")
		}

		mail := map[string]interface{}{
			"to":      to,
			"subject": subject,
			"body":    body,
		}

		resp, err := svc.SendMail(userID, mail)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var mailGetCmd = &cobra.Command{
	Use:   "get <mailId>",
	Short: "메일 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).GetMail(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var mailDeleteCmd = &cobra.Command{
	Use:   "delete <mailId>",
	Short: "메일 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).DeleteMail(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var mailListFoldersCmd = &cobra.Command{
	Use:   "list-folders",
	Short: "메일 폴더 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewMailService(client)
		return runListCmd(cmd, []string{"folderId", "folderName"}, "mailFolders", func(c string, n int) (*api.Response, error) {
			return svc.ListFolders(userID, c, n)
		})
	},
}

var mailGetFolderCmd = &cobra.Command{
	Use:   "get-folder <folderId>",
	Short: "메일 폴더 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).GetFolder(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var mailListCmd = &cobra.Command{
	Use:   "list <folderId>",
	Short: "메일 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewMailService(client)
		return runListCmd(cmd, []string{"mailId", "subject"}, "mails", func(c string, n int) (*api.Response, error) {
			return svc.ListMails(userID, args[0], c, n)
		})
	},
}

var mailUpdateCmd = &cobra.Command{
	Use:   "update <mailId>",
	Short: "메일 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).PatchMail(userID, args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var mailUnreadCountCmd = &cobra.Command{
	Use:   "unread-count",
	Short: "읽지 않은 메일 수 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).GetUnreadCount(userID)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var mailGetAttachmentCmd = &cobra.Command{
	Use:   "get-attachment <mailId> <attachmentId>",
	Short: "메일 첨부파일 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).GetAttachment(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var mailListFavoriteFoldersCmd = &cobra.Command{
	Use:   "list-favorite-folders",
	Short: "즐겨찾기 연락처 폴더 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).ListFavoriteContactsFolders(userID)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var mailCreateFolderCmd = &cobra.Command{
	Use:   "create-folder",
	Short: "메일 폴더 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).CreateMailFolder(userID, body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var mailUpdateFolderCmd = &cobra.Command{
	Use:   "update-folder <folderId>",
	Short: "메일 폴더 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).UpdateMailFolder(userID, args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var mailDeleteFolderCmd = &cobra.Command{
	Use:   "delete-folder <folderId>",
	Short: "메일 폴더 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).DeleteMailFolder(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// --- Mail Filter subcommand group ---

var mailFilterCmd = &cobra.Command{
	Use:   "filter",
	Short: "메일 필터 관리",
}

var mailFilterListCmd = &cobra.Command{
	Use:   "list",
	Short: "메일 필터 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).ListFilters(userID)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var mailFilterGetCmd = &cobra.Command{
	Use:   "get <filterId>",
	Short: "메일 필터 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).GetFilter(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var mailFilterCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "메일 필터 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).CreateFilter(userID, body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var mailFilterDeleteCmd = &cobra.Command{
	Use:   "delete <filterId>",
	Short: "메일 필터 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).DeleteFilter(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// --- Mail Migration subcommand group ---

var mailMigrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "메일 마이그레이션 관리",
}

var mailMigrationCreateImapCmd = &cobra.Command{
	Use:   "create-imap",
	Short: "IMAP 마이그레이션 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).CreateImapMigration(userID, body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var mailMigrationGetImapCmd = &cobra.Command{
	Use:   "get-imap",
	Short: "IMAP 마이그레이션 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).GetImapMigration(userID)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var mailMigrationDeleteImapCmd = &cobra.Command{
	Use:   "delete-imap",
	Short: "IMAP 마이그레이션 삭제",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).DeleteImapMigration(userID)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var mailMigrationCreatePop3Cmd = &cobra.Command{
	Use:   "create-pop3",
	Short: "POP3 마이그레이션 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).CreatePop3Migration(userID, body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// --- Mail Forwarding subcommand group ---

var mailForwardingCmd = &cobra.Command{
	Use:   "forwarding",
	Short: "메일 포워딩 관리",
}

var mailForwardingCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "메일 포워딩 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).CreateForwarding(userID, body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var mailForwardingDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "메일 포워딩 삭제",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewMailService(client).DeleteForwarding(userID)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

func init() {
	// user-id flag for all mail commands
	for _, c := range []*cobra.Command{
		mailSendCmd, mailGetCmd, mailDeleteCmd, mailListFoldersCmd, mailGetFolderCmd, mailListCmd,
		mailUpdateCmd, mailUnreadCountCmd, mailGetAttachmentCmd, mailListFavoriteFoldersCmd,
		mailCreateFolderCmd, mailUpdateFolderCmd, mailDeleteFolderCmd,
	} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	// user-id flag for filter subcommands
	for _, c := range []*cobra.Command{mailFilterListCmd, mailFilterGetCmd, mailFilterCreateCmd, mailFilterDeleteCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	// user-id flag for migration subcommands
	for _, c := range []*cobra.Command{mailMigrationCreateImapCmd, mailMigrationGetImapCmd, mailMigrationDeleteImapCmd, mailMigrationCreatePop3Cmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	// user-id flag for forwarding subcommands
	for _, c := range []*cobra.Command{mailForwardingCreateCmd, mailForwardingDeleteCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	addListFlags(mailListFoldersCmd, mailListCmd)

	mailSendCmd.Flags().String("to", "", "수신자 (필수)")
	mailSendCmd.Flags().String("subject", "", "제목 (필수)")
	mailSendCmd.Flags().String("body", "", "본문 (필수)")

	// --json flags
	mailUpdateCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	mailCreateFolderCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	mailUpdateFolderCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	mailFilterCreateCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	mailMigrationCreateImapCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	mailMigrationCreatePop3Cmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	mailForwardingCreateCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")

	// Register filter subcommands
	mailFilterCmd.AddCommand(mailFilterListCmd, mailFilterGetCmd, mailFilterCreateCmd, mailFilterDeleteCmd)

	// Register migration subcommands
	mailMigrationCmd.AddCommand(mailMigrationCreateImapCmd, mailMigrationGetImapCmd, mailMigrationDeleteImapCmd, mailMigrationCreatePop3Cmd)

	// Register forwarding subcommands
	mailForwardingCmd.AddCommand(mailForwardingCreateCmd, mailForwardingDeleteCmd)

	// Register all to mailCmd
	mailCmd.AddCommand(mailSendCmd, mailGetCmd, mailDeleteCmd, mailListFoldersCmd, mailGetFolderCmd, mailListCmd,
		mailUpdateCmd, mailUnreadCountCmd, mailGetAttachmentCmd, mailListFavoriteFoldersCmd,
		mailCreateFolderCmd, mailUpdateFolderCmd, mailDeleteFolderCmd,
		mailFilterCmd, mailMigrationCmd, mailForwardingCmd)
	rootCmd.AddCommand(mailCmd)
}
