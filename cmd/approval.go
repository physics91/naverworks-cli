package cmd

import (
	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var approvalCmd = &cobra.Command{
	Use:   "approval",
	Short: "결재 관리",
}

var approvalListCmd = &cobra.Command{
	Use:   "list",
	Short: "사용자별 결재 문서 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewApprovalService(client)
		return runListCmd(cmd, []string{"approvalDocumentId", "title"}, "documents", func(c string, n int) (*api.Response, error) {
			return svc.ListUserDocuments(userID, c, n)
		})
	},
}

var approvalListAllCmd = &cobra.Command{
	Use:   "list-all",
	Short: "전체 결재 문서 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewApprovalService(client)
		return runListCmd(cmd, []string{"approvalDocumentId", "title"}, "documents", svc.ListDocuments)
	},
}

var approvalGetCmd = &cobra.Command{
	Use:   "get <documentId>",
	Short: "결재 문서 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewApprovalService(client)

		resp, err := svc.GetDocument(args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var approvalListCategoriesCmd = &cobra.Command{
	Use:   "list-categories",
	Short: "결재 카테고리 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewApprovalService(client)
		return runListCmd(cmd, []string{"categoryId", "categoryName"}, "categories", svc.ListCategories)
	},
}

var approvalGetCategoryCmd = &cobra.Command{
	Use:   "get-category <categoryId>",
	Short: "결재 카테고리 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewApprovalService(client)

		resp, err := svc.GetCategory(args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var approvalListFormsCmd = &cobra.Command{
	Use:   "list-forms",
	Short: "결재 양식 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewApprovalService(client)
		return runListCmd(cmd, []string{"documentFormId", "documentFormName"}, "documentForms", svc.ListDocumentForms)
	},
}

func init() {
	addListFlags(approvalListCmd, approvalListAllCmd, approvalListCategoriesCmd, approvalListFormsCmd)
	approvalListCmd.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")

	approvalCmd.AddCommand(approvalListCmd, approvalListAllCmd, approvalGetCmd,
		approvalListCategoriesCmd, approvalGetCategoryCmd, approvalListFormsCmd)
	rootCmd.AddCommand(approvalCmd)
}
