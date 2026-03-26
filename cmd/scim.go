package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var scimCmd = &cobra.Command{
	Use:   "scim",
	Short: "SCIM 사용자/그룹 관리",
}

func loadScimService() (*api.ScimService, error) {
	cfg, _, err := loadActiveConfig()
	if err != nil {
		return nil, err
	}
	client, err := buildScimClient(cfg)
	if err != nil {
		return nil, err
	}
	return api.NewScimService(client), nil
}

func parseRequiredJSONData(cmd *cobra.Command) (map[string]interface{}, error) {
	data, _ := cmd.Flags().GetString("data")
	if data == "" {
		return nil, fmt.Errorf("--data는 필수입니다 (JSON 문자열)")
	}
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(data), &body); err != nil {
		return nil, fmt.Errorf("JSON 파싱 실패: %w", err)
	}
	return body, nil
}

// --- Users ---

var scimListUsersCmd = &cobra.Command{
	Use:   "list-users",
	Short: "SCIM 사용자 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		startIndex, _ := cmd.Flags().GetInt("start-index")
		count, _ := cmd.Flags().GetInt("count")
		filter, _ := cmd.Flags().GetString("filter")

		resp, err := svc.ListUsers(startIndex, count, filter)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var scimGetUserCmd = &cobra.Command{
	Use:   "get-user <id>",
	Short: "SCIM 사용자 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		resp, err := svc.GetUser(args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var scimCreateUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "SCIM 사용자 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		body, err := parseRequiredJSONData(cmd)
		if err != nil {
			return err
		}

		resp, err := svc.CreateUser(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var scimUpdateUserCmd = &cobra.Command{
	Use:   "update-user <id>",
	Short: "SCIM 사용자 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		body, err := parseRequiredJSONData(cmd)
		if err != nil {
			return err
		}

		resp, err := svc.UpdateUser(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var scimPatchUserCmd = &cobra.Command{
	Use:   "patch-user <id>",
	Short: "SCIM 사용자 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		body, err := parseRequiredJSONData(cmd)
		if err != nil {
			return err
		}

		resp, err := svc.PatchUser(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var scimDeleteUserCmd = &cobra.Command{
	Use:   "delete-user <id>",
	Short: "SCIM 사용자 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		resp, err := svc.DeleteUser(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// --- Groups ---

var scimListGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "SCIM 그룹 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		startIndex, _ := cmd.Flags().GetInt("start-index")
		count, _ := cmd.Flags().GetInt("count")
		filter, _ := cmd.Flags().GetString("filter")

		resp, err := svc.ListGroups(startIndex, count, filter)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var scimGetGroupCmd = &cobra.Command{
	Use:   "get-group <id>",
	Short: "SCIM 그룹 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		resp, err := svc.GetGroup(args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var scimCreateGroupCmd = &cobra.Command{
	Use:   "create-group",
	Short: "SCIM 그룹 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		body, err := parseRequiredJSONData(cmd)
		if err != nil {
			return err
		}

		resp, err := svc.CreateGroup(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var scimUpdateGroupCmd = &cobra.Command{
	Use:   "update-group <id>",
	Short: "SCIM 그룹 전체 수정 (PUT)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		body, err := parseRequiredJSONData(cmd)
		if err != nil {
			return err
		}

		resp, err := svc.UpdateGroup(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var scimPatchGroupCmd = &cobra.Command{
	Use:   "patch-group <id>",
	Short: "SCIM 그룹 부분 수정 (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		body, err := parseRequiredJSONData(cmd)
		if err != nil {
			return err
		}

		resp, err := svc.PatchGroup(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var scimDeleteGroupCmd = &cobra.Command{
	Use:   "delete-group <id>",
	Short: "SCIM 그룹 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := loadScimService()
		if err != nil {
			return err
		}

		resp, err := svc.DeleteGroup(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

func init() {
	for _, c := range []*cobra.Command{scimListUsersCmd, scimListGroupsCmd} {
		c.Flags().Int("start-index", 0, "시작 인덱스")
		c.Flags().Int("count", 0, "페이지 크기")
		c.Flags().String("filter", "", "SCIM 필터 표현식")
	}

	for _, c := range []*cobra.Command{scimCreateUserCmd, scimUpdateUserCmd, scimPatchUserCmd, scimCreateGroupCmd, scimUpdateGroupCmd, scimPatchGroupCmd} {
		c.Flags().String("data", "", "요청 본문 (JSON 문자열)")
	}

	scimCmd.AddCommand(
		scimListUsersCmd, scimGetUserCmd, scimCreateUserCmd,
		scimUpdateUserCmd, scimPatchUserCmd, scimDeleteUserCmd,
		scimListGroupsCmd, scimGetGroupCmd, scimCreateGroupCmd,
		scimUpdateGroupCmd, scimPatchGroupCmd, scimDeleteGroupCmd,
	)
	rootCmd.AddCommand(scimCmd)
}
