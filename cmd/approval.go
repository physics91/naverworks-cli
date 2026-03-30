package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
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
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"approvalDocumentId", "title"}, "documents", svc.ListDocuments)
	},
}

var approvalGetCmd = &cobra.Command{
	Use:   "get <documentId>",
	Short: "결재 문서 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewApprovalService(client).GetDocument(args[0])
		})
	},
}

var approvalListCategoriesCmd = &cobra.Command{
	Use:   "list-categories",
	Short: "결재 카테고리 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"categoryId", "categoryName"}, "categories", svc.ListCategories)
	},
}

var approvalGetCategoryCmd = &cobra.Command{
	Use:   "get-category <categoryId>",
	Short: "결재 카테고리 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewApprovalService(client).GetCategory(args[0])
		})
	},
}

var approvalListFormsCmd = &cobra.Command{
	Use:   "list-forms",
	Short: "결재 양식 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"documentFormId", "documentFormName"}, "documentForms", svc.ListDocumentForms)
	},
}

var approvalCreateCategoryCmd = &cobra.Command{
	Use:   "create-category",
	Short: "결재 카테고리 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateCategory(body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var approvalUpdateCategoryCmd = &cobra.Command{
	Use:   "update-category <categoryId>",
	Short: "결재 카테고리 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchCategory(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var approvalDeleteCategoryCmd = &cobra.Command{
	Use:   "delete-category <categoryId>",
	Short: "결재 카테고리 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewApprovalService(client).DeleteCategory(args[0])
		})
	},
}

var approvalCreateDocumentCmd = &cobra.Command{
	Use:   "create-document",
	Short: "결재 문서 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		svc := api.NewApprovalService(client)
		resp, err := svc.CreateUserDocument(userID, body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var approvalCreateImportedDocumentCmd = &cobra.Command{
	Use:   "create-imported-document",
	Short: "결재 외부 문서 가져오기",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateImportedDocument(body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var approvalCreateDocumentLinkCmd = &cobra.Command{
	Use:   "create-document-link",
	Short: "결재 문서 링크 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		svc := api.NewApprovalService(client)
		resp, err := svc.CreateDocumentLink(userID, body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var approvalGetFormCmd = &cobra.Command{
	Use:   "get-form <documentFormId>",
	Short: "결재 양식 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewApprovalService(client).GetDocumentForm(args[0])
		})
	},
}

var approvalUploadAttachmentCmd = &cobra.Command{
	Use:   "upload-attachment",
	Short: "결재 문서 첨부파일 업로드",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file 플래그가 필요합니다")
		}
		stat, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("파일 정보 조회 실패: %w", err)
		}
		uploadBody := map[string]interface{}{
			"fileName": filepath.Base(filePath),
			"fileSize": stat.Size(),
		}
		bodyBytes, _ := json.Marshal(uploadBody)
		svc := api.NewApprovalService(client)
		resp, err := svc.CreateUserDocumentAttachment(userID, bodyBytes)
		if err != nil {
			return err
		}
		var result struct {
			UploadURL string `json:"uploadUrl"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			return fmt.Errorf("업로드 URL 파싱 실패: %w", err)
		}
		if result.UploadURL == "" {
			return fmt.Errorf("업로드 URL을 받지 못했습니다")
		}
		if err := client.UploadFile(result.UploadURL, filePath); err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var approvalUploadImportedAttachmentCmd = &cobra.Command{
	Use:   "upload-imported-attachment",
	Short: "결재 외부 문서 첨부파일 업로드",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			return fmt.Errorf("--file 플래그가 필요합니다")
		}
		stat, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("파일 정보 조회 실패: %w", err)
		}
		uploadBody := map[string]interface{}{
			"fileName": filepath.Base(filePath),
			"fileSize": stat.Size(),
		}
		bodyBytes, _ := json.Marshal(uploadBody)
		svc := api.NewApprovalService(client)
		resp, err := svc.CreateImportedDocumentAttachment(bodyBytes)
		if err != nil {
			return err
		}
		var result struct {
			UploadURL string `json:"uploadUrl"`
		}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			return fmt.Errorf("업로드 URL 파싱 실패: %w", err)
		}
		if result.UploadURL == "" {
			return fmt.Errorf("업로드 URL을 받지 못했습니다")
		}
		if err := client.UploadFile(result.UploadURL, filePath); err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

// ─── Linkage Code subcommand group ───

var approvalLinkageCodeCmd = &cobra.Command{
	Use:   "linkage-code",
	Short: "결재 연동 코드 관리",
}

var approvalLinkageCodeListCmd = &cobra.Command{
	Use:   "list",
	Short: "연동 코드 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"key", "name"}, "linkageCodes", svc.ListLinkageCodes)
	},
}

var approvalLinkageCodeGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "연동 코드 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewApprovalService(client).GetLinkageCode(args[0])
		})
	},
}

var approvalLinkageCodeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "연동 코드 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateLinkageCode(body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var approvalLinkageCodeUpdateCmd = &cobra.Command{
	Use:   "update <key>",
	Short: "연동 코드 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchLinkageCode(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Linkage Code Item subcommand group ───

var approvalLinkageCodeItemCmd = &cobra.Command{
	Use:   "linkage-code-item",
	Short: "결재 연동 코드 항목 관리",
}

var approvalLinkageCodeItemListCmd = &cobra.Command{
	Use:   "list <key>",
	Short: "연동 코드 항목 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"id", "name"}, "linkageCodeItems", func(c string, n int) (*api.Response, error) {
			return svc.ListLinkageCodeItems(args[0], c, n)
		})
	},
}

var approvalLinkageCodeItemGetCmd = &cobra.Command{
	Use:   "get <key> <id>",
	Short: "연동 코드 항목 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewApprovalService(client).GetLinkageCodeItem(args[0], args[1])
		})
	},
}

var approvalLinkageCodeItemCreateCmd = &cobra.Command{
	Use:   "create <key>",
	Short: "연동 코드 항목 생성",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateLinkageCodeItem(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var approvalLinkageCodeItemUpdateCmd = &cobra.Command{
	Use:   "update <key> <id>",
	Short: "연동 코드 항목 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchLinkageCodeItem(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var approvalLinkageCodeItemDeleteCmd = &cobra.Command{
	Use:   "delete <key> <id>",
	Short: "연동 코드 항목 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewApprovalService(client).DeleteLinkageCodeItem(args[0], args[1])
		})
	},
}

func init() {
	addListFlags(approvalListCmd, approvalListAllCmd, approvalListCategoriesCmd, approvalListFormsCmd,
		approvalLinkageCodeListCmd, approvalLinkageCodeItemListCmd)

	approvalListCmd.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	for _, c := range []*cobra.Command{approvalCreateDocumentCmd, approvalCreateDocumentLinkCmd, approvalUploadAttachmentCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	for _, c := range []*cobra.Command{
		approvalCreateCategoryCmd, approvalUpdateCategoryCmd,
		approvalCreateDocumentCmd, approvalCreateImportedDocumentCmd, approvalCreateDocumentLinkCmd,
		approvalLinkageCodeCreateCmd, approvalLinkageCodeUpdateCmd,
		approvalLinkageCodeItemCreateCmd, approvalLinkageCodeItemUpdateCmd,
	} {
		c.Flags().String("json", "", "JSON 데이터 (- 이면 stdin)")
	}

	for _, c := range []*cobra.Command{approvalUploadAttachmentCmd, approvalUploadImportedAttachmentCmd} {
		c.Flags().String("file", "", "업로드할 파일 경로")
	}

	// linkage-code subcommand group
	approvalLinkageCodeCmd.AddCommand(approvalLinkageCodeListCmd, approvalLinkageCodeGetCmd,
		approvalLinkageCodeCreateCmd, approvalLinkageCodeUpdateCmd)

	// linkage-code-item subcommand group
	approvalLinkageCodeItemCmd.AddCommand(approvalLinkageCodeItemListCmd, approvalLinkageCodeItemGetCmd,
		approvalLinkageCodeItemCreateCmd, approvalLinkageCodeItemUpdateCmd, approvalLinkageCodeItemDeleteCmd)

	approvalCmd.AddCommand(approvalListCmd, approvalListAllCmd, approvalGetCmd,
		approvalListCategoriesCmd, approvalGetCategoryCmd, approvalListFormsCmd,
		approvalCreateCategoryCmd, approvalUpdateCategoryCmd, approvalDeleteCategoryCmd,
		approvalCreateDocumentCmd, approvalCreateImportedDocumentCmd, approvalCreateDocumentLinkCmd,
		approvalGetFormCmd, approvalUploadAttachmentCmd, approvalUploadImportedAttachmentCmd,
		approvalLinkageCodeCmd, approvalLinkageCodeItemCmd,
	)
	rootCmd.AddCommand(approvalCmd)
}
