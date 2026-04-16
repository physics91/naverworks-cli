package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var approvalCmd = &cobra.Command{
	Use:   "approval",
	Short: "결재 관리",
}

type approvalIDCall func(*api.ApprovalService, string) (*api.Response, error)
type approvalTwoIDCall func(*api.ApprovalService, string, string) (*api.Response, error)
type approvalBodyCall func(*api.ApprovalService, []byte) (*api.Response, error)
type approvalIDBodyCall func(*api.ApprovalService, string, []byte) (*api.Response, error)
type approvalTwoIDBodyCall func(*api.ApprovalService, string, string, []byte) (*api.Response, error)
type approvalListCall func(*api.ApprovalService, string, int) (*api.Response, error)
type approvalIDListCall func(*api.ApprovalService, string, string, int) (*api.Response, error)
type approvalUserBodyCall func(*api.ApprovalService, string, []byte) (*api.Response, error)
type approvalUserListCall func(*api.ApprovalService, string, string, int) (*api.Response, error)

// Keep these wrappers local so cmd/helpers.go does not grow an approval-only helper family.
func printApprovalBody(resp *api.Response) {
	printBody(resp.Body)
}

func approvalIDRunE(call approvalIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
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

func approvalTwoIDRunE(call approvalTwoIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0], args[1])
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func approvalBodyRunE(call approvalBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := call(svc, body)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func approvalIDBodyRunE(call approvalIDBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0], body)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func approvalTwoIDBodyRunE(call approvalTwoIDBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := call(svc, args[0], args[1], body)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func approvalListRunListE(columns []string, itemKey string, call approvalListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(svc, cursor, count)
		})
	}
}

func approvalIDRunListE(columns []string, itemKey string, call approvalIDListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewApprovalService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(svc, args[0], cursor, count)
		})
	}
}

func approvalUserBodyRunE(call approvalUserBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := call(api.NewApprovalService(client), userID, body)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

func approvalUserRunListE(columns []string, itemKey string, call approvalUserListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewApprovalService(client)
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(svc, userID, cursor, count)
		})
	}
}

var approvalListCmd = &cobra.Command{
	Use:   "list",
	Short: "사용자별 결재 문서 목록 조회",
	RunE:  approvalUserRunListE([]string{"approvalDocumentId", "title"}, "documents", (*api.ApprovalService).ListUserDocuments),
}

var approvalListAllCmd = &cobra.Command{
	Use:   "list-all",
	Short: "전체 결재 문서 목록 조회",
	RunE:  approvalListRunListE([]string{"approvalDocumentId", "title"}, "documents", (*api.ApprovalService).ListDocuments),
}

var approvalGetCmd = &cobra.Command{
	Use:   "get <documentId>",
	Short: "결재 문서 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  approvalIDRunE((*api.ApprovalService).GetDocument, printApprovalBody),
}

var approvalListCategoriesCmd = &cobra.Command{
	Use:   "list-categories",
	Short: "결재 카테고리 목록 조회",
	RunE:  approvalListRunListE([]string{"categoryId", "categoryName"}, "categories", (*api.ApprovalService).ListCategories),
}

var approvalGetCategoryCmd = &cobra.Command{
	Use:   "get-category <categoryId>",
	Short: "결재 카테고리 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  approvalIDRunE((*api.ApprovalService).GetCategory, printApprovalBody),
}

var approvalListFormsCmd = &cobra.Command{
	Use:   "list-forms",
	Short: "결재 양식 목록 조회",
	RunE:  approvalListRunListE([]string{"documentFormId", "documentFormName"}, "documentForms", (*api.ApprovalService).ListDocumentForms),
}

var approvalCreateCategoryCmd = &cobra.Command{
	Use:   "create-category",
	Short: "결재 카테고리 생성",
	RunE:  approvalBodyRunE((*api.ApprovalService).CreateCategory, printResponse),
}

var approvalUpdateCategoryCmd = &cobra.Command{
	Use:   "update-category <categoryId>",
	Short: "결재 카테고리 수정",
	Args:  cobra.ExactArgs(1),
	RunE:  approvalIDBodyRunE((*api.ApprovalService).PatchCategory, printResponse),
}

var approvalDeleteCategoryCmd = &cobra.Command{
	Use:   "delete-category <categoryId>",
	Short: "결재 카테고리 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  approvalIDRunE((*api.ApprovalService).DeleteCategory, printApprovalBody),
}

var approvalCreateDocumentCmd = &cobra.Command{
	Use:   "create-document",
	Short: "결재 문서 생성",
	RunE:  approvalUserBodyRunE((*api.ApprovalService).CreateUserDocument, printResponse),
}

var approvalCreateImportedDocumentCmd = &cobra.Command{
	Use:   "create-imported-document",
	Short: "결재 외부 문서 가져오기",
	RunE:  approvalBodyRunE((*api.ApprovalService).CreateImportedDocument, printResponse),
}

var approvalCreateDocumentLinkCmd = &cobra.Command{
	Use:   "create-document-link",
	Short: "결재 문서 링크 생성",
	RunE:  approvalUserBodyRunE((*api.ApprovalService).CreateDocumentLink, printResponse),
}

var approvalGetFormCmd = &cobra.Command{
	Use:   "get-form <documentFormId>",
	Short: "결재 양식 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  approvalIDRunE((*api.ApprovalService).GetDocumentForm, printApprovalBody),
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
		fileName, fileSize, err := statFileForUpload(filePath)
		if err != nil {
			return err
		}
		uploadBody := map[string]interface{}{
			"fileName": fileName,
			"fileSize": fileSize,
		}
		bodyBytes, err := json.Marshal(uploadBody)
		if err != nil {
			return fmt.Errorf("JSON 직렬화 실패: %w", err)
		}
		svc := api.NewApprovalService(client)
		resp, err := svc.CreateUserDocumentAttachment(userID, bodyBytes)
		if err != nil {
			return err
		}
		if err := doUploadFromResponse(client, resp.Body, filePath); err != nil {
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
		fileName, fileSize, err := statFileForUpload(filePath)
		if err != nil {
			return err
		}
		uploadBody := map[string]interface{}{
			"fileName": fileName,
			"fileSize": fileSize,
		}
		bodyBytes, err := json.Marshal(uploadBody)
		if err != nil {
			return fmt.Errorf("JSON 직렬화 실패: %w", err)
		}
		svc := api.NewApprovalService(client)
		resp, err := svc.CreateImportedDocumentAttachment(bodyBytes)
		if err != nil {
			return err
		}
		if err := doUploadFromResponse(client, resp.Body, filePath); err != nil {
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
	RunE:  approvalListRunListE([]string{"key", "name"}, "linkageCodes", (*api.ApprovalService).ListLinkageCodes),
}

var approvalLinkageCodeGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "연동 코드 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  approvalIDRunE((*api.ApprovalService).GetLinkageCode, printApprovalBody),
}

var approvalLinkageCodeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "연동 코드 생성",
	RunE:  approvalBodyRunE((*api.ApprovalService).CreateLinkageCode, printResponse),
}

var approvalLinkageCodeUpdateCmd = &cobra.Command{
	Use:   "update <key>",
	Short: "연동 코드 수정",
	Args:  cobra.ExactArgs(1),
	RunE:  approvalIDBodyRunE((*api.ApprovalService).PatchLinkageCode, printResponse),
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
	RunE:  approvalIDRunListE([]string{"id", "name"}, "linkageCodeItems", (*api.ApprovalService).ListLinkageCodeItems),
}

var approvalLinkageCodeItemGetCmd = &cobra.Command{
	Use:   "get <key> <id>",
	Short: "연동 코드 항목 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE:  approvalTwoIDRunE((*api.ApprovalService).GetLinkageCodeItem, printApprovalBody),
}

var approvalLinkageCodeItemCreateCmd = &cobra.Command{
	Use:   "create <key>",
	Short: "연동 코드 항목 생성",
	Args:  cobra.ExactArgs(1),
	RunE:  approvalIDBodyRunE((*api.ApprovalService).CreateLinkageCodeItem, printResponse),
}

var approvalLinkageCodeItemUpdateCmd = &cobra.Command{
	Use:   "update <key> <id>",
	Short: "연동 코드 항목 수정",
	Args:  cobra.ExactArgs(2),
	RunE:  approvalTwoIDBodyRunE((*api.ApprovalService).PatchLinkageCodeItem, printResponse),
}

var approvalLinkageCodeItemDeleteCmd = &cobra.Command{
	Use:   "delete <key> <id>",
	Short: "연동 코드 항목 삭제",
	Args:  cobra.ExactArgs(2),
	RunE:  approvalTwoIDRunE((*api.ApprovalService).DeleteLinkageCodeItem, printApprovalBody),
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
