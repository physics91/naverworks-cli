package cmd

import (
	"fmt"
	"strings"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var contactCmd = &cobra.Command{
	Use:   "contact",
	Short: "연락처 관리",
}

const contactSearchHelp = "인증: 구성원 계정 또는 서비스 계정 Access Token\n최소 scope: contact.read"

type contactIDCall func(*api.ContactService, string) (*api.Response, error)
type contactBodyReader func(*cobra.Command) (map[string]interface{}, error)
type contactBodyCall func(*api.ContactService, map[string]interface{}) (*api.Response, error)
type contactIDBodyCall func(*api.ContactService, string, map[string]interface{}) (*api.Response, error)
type contactListCall func(*api.ContactService, string, int) (*api.Response, error)
type contactUserListCall func(*api.ContactService, string, string, int) (*api.Response, error)
type contactUserBodyCall func(*api.ContactService, string, map[string]interface{}) (*api.Response, error)

// Keep these wrappers local so cmd/helpers.go does not grow a contact-only helper family.
func printContactBody(resp *api.Response) {
	printBody(resp.Body)
}

func buildContactCreateBody(cmd *cobra.Command) (map[string]interface{}, error) {
	body, err := parseOptionalJSONData(cmd)
	if err != nil {
		return nil, err
	}
	if body != nil {
		return body, nil
	}

	contactName, _ := cmd.Flags().GetString("name")
	email, _ := cmd.Flags().GetString("email")
	phone, _ := cmd.Flags().GetString("phone")
	if contactName == "" {
		return nil, fmt.Errorf("--name은 필수입니다")
	}

	body = map[string]interface{}{"name": contactName}
	if email != "" {
		body["email"] = email
	}
	if phone != "" {
		body["phone"] = phone
	}
	return body, nil
}

func buildContactUpdateBody(cmd *cobra.Command) (map[string]interface{}, error) {
	body, err := parseOptionalJSONData(cmd)
	if err != nil {
		return nil, err
	}
	if body != nil {
		return body, nil
	}

	body = map[string]interface{}{}
	if contactName, _ := cmd.Flags().GetString("name"); contactName != "" {
		body["name"] = contactName
	}
	if email, _ := cmd.Flags().GetString("email"); email != "" {
		body["email"] = email
	}
	if phone, _ := cmd.Flags().GetString("phone"); phone != "" {
		body["phone"] = phone
	}
	return body, nil
}

func contactIDRunE(call contactIDCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
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

func contactBodyRunE(readBody contactBodyReader, call contactBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		body, err := readBody(cmd)
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

func contactIDBodyRunE(readBody contactBodyReader, call contactIDBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		body, err := readBody(cmd)
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

func contactListRunListE(columns []string, itemKey string, call contactListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(svc, cursor, count)
		})
	}
}

func contactUserRunListE(columns []string, itemKey string, call contactUserListCall) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewContactService(client)
		return runListCmd(cmd, columns, itemKey, func(cursor string, count int) (*api.Response, error) {
			return call(svc, userID, cursor, count)
		})
	}
}

func contactUserBodyRunE(readBody contactBodyReader, call contactUserBodyCall, printer func(*api.Response)) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readBody(cmd)
		if err != nil {
			return err
		}
		resp, err := call(api.NewContactService(client), userID, body)
		if err != nil {
			return err
		}
		printer(resp)
		return nil
	}
}

var contactListCmd = &cobra.Command{
	Use:   "list",
	Short: "연락처 목록 조회",
	RunE:  contactListRunListE([]string{"contactId", "name", "email"}, "contacts", (*api.ContactService).ListContacts),
}

var contactListUserCmd = &cobra.Command{
	Use:   "list-user",
	Short: "사용자별 연락처 목록 조회",
	RunE:  contactUserRunListE([]string{"contactId", "name", "email"}, "contacts", (*api.ContactService).ListUserContacts),
}

var contactSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "사용자가 접근 가능한 연락처 검색",
	Long:  "사용자가 접근 가능한 연락처 검색\n\n" + contactSearchHelp,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.TrimSpace(args[0])
		if query == "" {
			return fmt.Errorf("query는 비어 있을 수 없습니다")
		}
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		queryFilters, _ := cmd.Flags().GetString("query-filters")
		orderBy, _ := cmd.Flags().GetString("order-by")
		options := api.ContactSearchOptions{Query: query, QueryFilters: queryFilters, OrderBy: orderBy}
		svc := api.NewContactService(client)
		columns := []string{
			"contactId",
			"lastName=contactName.lastName",
			"firstName=contactName.firstName",
			"email=emails.0.email",
		}
		return runListCmd(cmd, columns, "contacts", func(cursor string, count int) (*api.Response, error) {
			return svc.SearchContacts(userID, cursor, count, options)
		})
	},
}

var contactGetCmd = &cobra.Command{
	Use:   "get <contactId>",
	Short: "연락처 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDRunE((*api.ContactService).GetContact, printContactBody),
}

var contactCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "연락처 생성",
	RunE:  contactBodyRunE(buildContactCreateBody, (*api.ContactService).CreateContact, printContactBody),
}

var contactUpdateCmd = &cobra.Command{
	Use:   "update <contactId>",
	Short: "연락처 수정",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDBodyRunE(buildContactUpdateBody, (*api.ContactService).UpdateContact, printContactBody),
}

var contactFullUpdateCmd = &cobra.Command{
	Use:   "full-update <contactId>",
	Short: "연락처 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDBodyRunE(readJSONFlag, (*api.ContactService).FullUpdateContact, printContactBody),
}

var contactDeleteCmd = &cobra.Command{
	Use:   "delete <contactId>",
	Short: "연락처 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDRunE((*api.ContactService).DeleteContact, printResponse),
}

var contactForceDeleteCmd = &cobra.Command{
	Use:   "force-delete <contactId>",
	Short: "연락처 강제 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDRunE((*api.ContactService).ForceDeleteContact, printResponse),
}

// ─── Photo ───

var contactUploadPhotoCmd = &cobra.Command{
	Use:   "upload-photo <contactId>",
	Short: "연락처 사진 업로드",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewContactService(client)

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
		resp, err := svc.CreatePhoto(args[0], uploadBody)
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

var contactGetPhotoCmd = &cobra.Command{
	Use:   "get-photo <contactId>",
	Short: "연락처 사진 다운로드 URL 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		downloadURL, err := svc.GetPhoto(args[0])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

var contactDeletePhotoCmd = &cobra.Command{
	Use:   "delete-photo <contactId>",
	Short: "연락처 사진 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDRunE((*api.ContactService).DeletePhoto, printResponse),
}

// ─── Custom Property ───

var contactCustomPropertyCmd = &cobra.Command{
	Use:   "custom-property",
	Short: "연락처 커스텀 속성 관리",
}

var contactCustomPropertyListCmd = &cobra.Command{
	Use:   "list",
	Short: "커스텀 속성 목록 조회",
	RunE:  contactListRunListE(nil, "customProperties", (*api.ContactService).ListCustomProperties),
}

var contactCustomPropertyGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "커스텀 속성 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDRunE((*api.ContactService).GetCustomProperty, printContactBody),
}

var contactCustomPropertyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "커스텀 속성 생성",
	RunE:  contactBodyRunE(readJSONFlag, (*api.ContactService).CreateCustomProperty, printContactBody),
}

var contactCustomPropertyUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "커스텀 속성 수정",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDBodyRunE(readJSONFlag, (*api.ContactService).PatchCustomProperty, printContactBody),
}

var contactCustomPropertyDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "커스텀 속성 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDRunE((*api.ContactService).DeleteCustomProperty, printResponse),
}

var contactListTagsCmd = &cobra.Command{
	Use:   "list-tags",
	Short: "연락처 태그 목록 조회",
	RunE:  contactListRunListE([]string{"tagId", "tagName"}, "contactTags", (*api.ContactService).ListTags),
}

var contactListUserTagsCmd = &cobra.Command{
	Use:   "list-user-tags",
	Short: "사용자별 연락처 태그 목록 조회",
	RunE:  contactUserRunListE([]string{"tagId", "tagName"}, "contactTags", (*api.ContactService).ListUserTags),
}

// ─── Tag Subcommands ───

var contactTagCmd = &cobra.Command{
	Use:   "tag",
	Short: "연락처 태그 관리",
}

var contactTagCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "태그 생성",
	RunE:  contactBodyRunE(readJSONFlag, (*api.ContactService).CreateTag, printContactBody),
}

var contactTagGetCmd = &cobra.Command{
	Use:   "get <tagId>",
	Short: "태그 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDRunE((*api.ContactService).GetTag, printContactBody),
}

var contactTagUpdateCmd = &cobra.Command{
	Use:   "update <tagId>",
	Short: "태그 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDBodyRunE(readJSONFlag, (*api.ContactService).UpdateTag, printContactBody),
}

var contactTagPatchCmd = &cobra.Command{
	Use:   "patch <tagId>",
	Short: "태그 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDBodyRunE(readJSONFlag, (*api.ContactService).PatchTag, printContactBody),
}

var contactTagDeleteCmd = &cobra.Command{
	Use:   "delete <tagId>",
	Short: "태그 삭제",
	Args:  cobra.ExactArgs(1),
	RunE:  contactIDRunE((*api.ContactService).DeleteTag, printResponse),
}

var contactTagCreateUserTagsCmd = &cobra.Command{
	Use:   "create-user-tags",
	Short: "사용자별 연락처 태그 생성",
	RunE:  contactUserBodyRunE(readJSONFlag, (*api.ContactService).CreateUserTags, printContactBody),
}

func init() {
	addListFlags(contactListCmd, contactListUserCmd, contactSearchCmd, contactListTagsCmd, contactListUserTagsCmd, contactCustomPropertyListCmd)
	for _, c := range []*cobra.Command{contactListUserCmd, contactSearchCmd, contactListUserTagsCmd, contactTagCreateUserTagsCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	contactSearchCmd.Flags().String("query-filters", "", "검색 대상 필드 (contactName,emails,telephones,organizations,contactTagName; 쉼표 구분)")
	contactSearchCmd.Flags().String("order-by", "", "정렬 기준과 순서 (예: name asc)")

	contactCreateCmd.Flags().String("name", "", "연락처 이름 (필수)")
	contactCreateCmd.Flags().String("email", "", "이메일")
	contactCreateCmd.Flags().String("phone", "", "전화번호")
	contactCreateCmd.Flags().String("data", "", "전체 JSON 페이로드")

	contactUpdateCmd.Flags().String("name", "", "연락처 이름")
	contactUpdateCmd.Flags().String("email", "", "이메일")
	contactUpdateCmd.Flags().String("phone", "", "전화번호")
	contactUpdateCmd.Flags().String("data", "", "전체 JSON 페이로드")

	contactFullUpdateCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	contactUploadPhotoCmd.Flags().String("file", "", "사진 파일 경로 (필수)")

	// --json flags for custom-property commands
	contactCustomPropertyCreateCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	contactCustomPropertyUpdateCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")

	// --json flags for tag commands
	contactTagCreateCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	contactTagUpdateCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	contactTagPatchCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	contactTagCreateUserTagsCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")

	// Register custom-property subcommands
	contactCustomPropertyCmd.AddCommand(contactCustomPropertyListCmd, contactCustomPropertyGetCmd,
		contactCustomPropertyCreateCmd, contactCustomPropertyUpdateCmd, contactCustomPropertyDeleteCmd)

	// Register tag subcommands
	contactTagCmd.AddCommand(contactTagCreateCmd, contactTagGetCmd, contactTagUpdateCmd,
		contactTagPatchCmd, contactTagDeleteCmd, contactTagCreateUserTagsCmd)

	contactCmd.AddCommand(contactListCmd, contactListUserCmd, contactSearchCmd, contactGetCmd, contactCreateCmd,
		contactUpdateCmd, contactFullUpdateCmd, contactDeleteCmd, contactForceDeleteCmd,
		contactUploadPhotoCmd, contactGetPhotoCmd, contactDeletePhotoCmd,
		contactListTagsCmd, contactListUserTagsCmd,
		contactCustomPropertyCmd, contactTagCmd)
	rootCmd.AddCommand(contactCmd)
}
