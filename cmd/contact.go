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

func newContactService() (*api.ContactService, error) {
	client, _, _, err := newAPIClient()
	if err != nil {
		return nil, err
	}
	return api.NewContactService(client), nil
}

var contactListCmd = &cobra.Command{
	Use:   "list",
	Short: "연락처 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newContactService()
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
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewContactService(client).GetContact(args[0])
		})
	},
}

var contactCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "연락처 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newContactService()
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
			if contactName == "" {
				return fmt.Errorf("--name은 필수입니다")
			}
			body = map[string]interface{}{"name": contactName}
			if email != "" {
				body["email"] = email
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
		svc, err := newContactService()
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
		}

		resp, err := svc.UpdateContact(args[0], body)
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
		svc, err := newContactService()
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

var contactListTagsCmd = &cobra.Command{
	Use:   "list-tags",
	Short: "연락처 태그 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newContactService()
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

func init() {
	addListFlags(contactListCmd, contactListUserCmd, contactListTagsCmd, contactListUserTagsCmd)
	for _, c := range []*cobra.Command{contactListUserCmd, contactListUserTagsCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	contactCreateCmd.Flags().String("name", "", "연락처 이름 (필수)")
	contactCreateCmd.Flags().String("email", "", "이메일")
	contactCreateCmd.Flags().String("data", "", "전체 JSON 페이로드")

	contactUpdateCmd.Flags().String("name", "", "연락처 이름")
	contactUpdateCmd.Flags().String("email", "", "이메일")
	contactUpdateCmd.Flags().String("data", "", "전체 JSON 페이로드")

	contactCmd.AddCommand(contactListCmd, contactListUserCmd, contactGetCmd, contactCreateCmd,
		contactUpdateCmd, contactDeleteCmd, contactListTagsCmd, contactListUserTagsCmd)
	rootCmd.AddCommand(contactCmd)
}
