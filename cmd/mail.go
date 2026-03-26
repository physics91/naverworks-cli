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
		client, cfg, token, err := newAPIClient()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
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
		client, cfg, token, err := newAPIClient()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		svc := api.NewMailService(client)

		resp, err := svc.GetMail(userID, args[0])
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
		client, cfg, token, err := newAPIClient()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		svc := api.NewMailService(client)

		resp, err := svc.DeleteMail(userID, args[0])
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
		client, cfg, token, err := newAPIClient()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
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
		client, cfg, token, err := newAPIClient()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		svc := api.NewMailService(client)

		resp, err := svc.GetFolder(userID, args[0])
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
		client, cfg, token, err := newAPIClient()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		svc := api.NewMailService(client)
		return runListCmd(cmd, []string{"mailId", "subject"}, "mails", func(c string, n int) (*api.Response, error) {
			return svc.ListMails(userID, args[0], c, n)
		})
	},
}

func init() {
	for _, c := range []*cobra.Command{mailSendCmd, mailGetCmd, mailDeleteCmd, mailListFoldersCmd, mailGetFolderCmd, mailListCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	addListFlags(mailListFoldersCmd, mailListCmd)

	mailSendCmd.Flags().String("to", "", "수신자 (필수)")
	mailSendCmd.Flags().String("subject", "", "제목 (필수)")
	mailSendCmd.Flags().String("body", "", "본문 (필수)")

	mailCmd.AddCommand(mailSendCmd, mailGetCmd, mailDeleteCmd, mailListFoldersCmd, mailGetFolderCmd, mailListCmd)
	rootCmd.AddCommand(mailCmd)
}
