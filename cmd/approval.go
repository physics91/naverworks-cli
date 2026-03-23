package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
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

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"approvalDocumentId", "title"}, "documents")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListUserDocuments(userID, c, count)
			}, "documents")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"documents": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListUserDocuments(userID, cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
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

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"approvalDocumentId", "title"}, "documents")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListDocuments(c, count)
			}, "documents")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"documents": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListDocuments(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
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
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
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

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"categoryId", "categoryName"}, "categories")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListCategories(c, count)
			}, "categories")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"categories": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListCategories(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
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
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
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

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"documentFormId", "documentFormName"}, "documentForms")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListDocumentForms(c, count)
			}, "documentForms")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"documentForms": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListDocumentForms(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	for _, c := range []*cobra.Command{approvalListCmd, approvalListAllCmd, approvalListCategoriesCmd, approvalListFormsCmd} {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
		c.Flags().Bool("all", false, "전체 페이지 자동 순회")
	}
	approvalListCmd.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")

	approvalCmd.AddCommand(approvalListCmd, approvalListAllCmd, approvalGetCmd,
		approvalListCategoriesCmd, approvalGetCategoryCmd, approvalListFormsCmd)
	rootCmd.AddCommand(approvalCmd)
}
