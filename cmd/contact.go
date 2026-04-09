package cmd

import (
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var contactCmd = &cobra.Command{
	Use:   "contact",
	Short: "연락처 관리",
}

var contactListCmd = &cobra.Command{
	Use:   "list",
	Short: "연락처 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"contactId", "name", "email"}, "contacts", svc.ListContacts)
	},
}

var contactListUserCmd = &cobra.Command{
	Use:   "list-user",
	Short: "사용자별 연락처 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewContactService(client)
		return runListCmd(cmd, []string{"contactId", "name", "email"}, "contacts", func(c string, n int) (*api.Response, error) {
			return svc.ListUserContacts(userID, c, n)
		})
	},
}

var contactGetCmd = &cobra.Command{
	Use:   "get <contactId>",
	Short: "연락처 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewContactService(client).GetContact(args[0])
		})
	},
}

var contactCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "연락처 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}

		body, err := parseOptionalJSONData(cmd)
		if err != nil {
			return err
		}
		if body == nil {
			contactName, _ := cmd.Flags().GetString("name")
			email, _ := cmd.Flags().GetString("email")
			phone, _ := cmd.Flags().GetString("phone")
			if contactName == "" {
				return fmt.Errorf("--name은 필수입니다")
			}
			body = map[string]interface{}{"name": contactName}
			if email != "" {
				body["email"] = email
			}
			if phone != "" {
				body["phone"] = phone
			}
		}

		resp, err := svc.CreateContact(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var contactUpdateCmd = &cobra.Command{
	Use:   "update <contactId>",
	Short: "연락처 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}

		body, err := parseOptionalJSONData(cmd)
		if err != nil {
			return err
		}
		if body == nil {
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
		}

		resp, err := svc.UpdateContact(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var contactFullUpdateCmd = &cobra.Command{
	Use:   "full-update <contactId>",
	Short: "연락처 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.FullUpdateContact(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var contactDeleteCmd = &cobra.Command{
	Use:   "delete <contactId>",
	Short: "연락처 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteContact(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var contactForceDeleteCmd = &cobra.Command{
	Use:   "force-delete <contactId>",
	Short: "연락처 강제 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		resp, err := svc.ForceDeleteContact(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
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

		if err := doUploadFromResponse(client, resp.Body, filePath); err != nil {
			return err
		}
		printBody(resp.Body)
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
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		resp, err := svc.DeletePhoto(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Custom Property ───

var contactCustomPropertyCmd = &cobra.Command{
	Use:   "custom-property",
	Short: "연락처 커스텀 속성 관리",
}

var contactCustomPropertyListCmd = &cobra.Command{
	Use:   "list",
	Short: "커스텀 속성 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, nil, "customProperties", svc.ListCustomProperties)
	},
}

var contactCustomPropertyGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "커스텀 속성 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewContactService(client).GetCustomProperty(args[0])
		})
	},
}

var contactCustomPropertyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "커스텀 속성 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateCustomProperty(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var contactCustomPropertyUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "커스텀 속성 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchCustomProperty(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var contactCustomPropertyDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "커스텀 속성 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteCustomProperty(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var contactListTagsCmd = &cobra.Command{
	Use:   "list-tags",
	Short: "연락처 태그 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"tagId", "tagName"}, "contactTags", svc.ListTags)
	},
}

var contactListUserTagsCmd = &cobra.Command{
	Use:   "list-user-tags",
	Short: "사용자별 연락처 태그 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewContactService(client)
		return runListCmd(cmd, []string{"tagId", "tagName"}, "contactTags", func(c string, n int) (*api.Response, error) {
			return svc.ListUserTags(userID, c, n)
		})
	},
}

// ─── Tag Subcommands ───

var contactTagCmd = &cobra.Command{
	Use:   "tag",
	Short: "연락처 태그 관리",
}

var contactTagCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "태그 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateTag(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var contactTagGetCmd = &cobra.Command{
	Use:   "get <tagId>",
	Short: "태그 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewContactService(client).GetTag(args[0])
		})
	},
}

var contactTagUpdateCmd = &cobra.Command{
	Use:   "update <tagId>",
	Short: "태그 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpdateTag(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var contactTagPatchCmd = &cobra.Command{
	Use:   "patch <tagId>",
	Short: "태그 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchTag(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var contactTagDeleteCmd = &cobra.Command{
	Use:   "delete <tagId>",
	Short: "태그 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewContactService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteTag(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var contactTagCreateUserTagsCmd = &cobra.Command{
	Use:   "create-user-tags",
	Short: "사용자별 연락처 태그 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewContactService(client)
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateUserTags(userID, body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

func init() {
	addListFlags(contactListCmd, contactListUserCmd, contactListTagsCmd, contactListUserTagsCmd, contactCustomPropertyListCmd)
	for _, c := range []*cobra.Command{contactListUserCmd, contactListUserTagsCmd, contactTagCreateUserTagsCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

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

	contactCmd.AddCommand(contactListCmd, contactListUserCmd, contactGetCmd, contactCreateCmd,
		contactUpdateCmd, contactFullUpdateCmd, contactDeleteCmd, contactForceDeleteCmd,
		contactUploadPhotoCmd, contactGetPhotoCmd, contactDeletePhotoCmd,
		contactListTagsCmd, contactListUserTagsCmd,
		contactCustomPropertyCmd, contactTagCmd)
	rootCmd.AddCommand(contactCmd)
}
