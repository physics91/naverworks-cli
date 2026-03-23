package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var directoryCmd = &cobra.Command{
	Use:   "directory",
	Short: "디렉토리 관리 (사용자, 그룹)",
}

var dirListUsersCmd = &cobra.Command{
	Use:   "list-users",
	Short: "사용자 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"userId", "userName", "email"}, "users")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return dir.ListUsers(c, count)
			}, "users")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"users": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := dir.ListUsers(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var dirGetUserCmd = &cobra.Command{
	Use:   "get-user <userId>",
	Short: "사용자 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		resp, err := dir.GetUser(args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var dirListGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "그룹 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"groupId", "groupName"}, "groups")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return dir.ListGroups(c, count)
			}, "groups")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"groups": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := dir.ListGroups(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var dirGetGroupCmd = &cobra.Command{
	Use:   "get-group <groupId>",
	Short: "그룹 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		resp, err := dir.GetGroup(args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var dirListOrgUnitsCmd = &cobra.Command{
	Use:   "list-orgunits",
	Short: "조직 단위 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"orgUnitId", "orgUnitName"}, "orgUnits")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return dir.ListOrgUnits(c, count)
			}, "orgUnits")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"orgUnits": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := dir.ListOrgUnits(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var dirGetOrgUnitCmd = &cobra.Command{
	Use:   "get-orgunit <orgUnitId>",
	Short: "조직 단위 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		resp, err := dir.GetOrgUnit(args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var dirListLevelsCmd = &cobra.Command{
	Use:   "list-levels",
	Short: "직급 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"levelId", "levelName"}, "levels")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return dir.ListLevels(c, count)
			}, "levels")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"levels": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := dir.ListLevels(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var dirListPositionsCmd = &cobra.Command{
	Use:   "list-positions",
	Short: "직책 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"positionId", "positionName"}, "positions")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return dir.ListPositions(c, count)
			}, "positions")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"positions": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := dir.ListPositions(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var dirListUserTypesCmd = &cobra.Command{
	Use:   "list-user-types",
	Short: "사용자 유형 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"userTypeId", "userTypeName"}, "userTypes")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return dir.ListUserTypes(c, count)
			}, "userTypes")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"userTypes": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := dir.ListUserTypes(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var dirListEmploymentTypesCmd = &cobra.Command{
	Use:   "list-employment-types",
	Short: "고용 유형 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		dir := api.NewDirectoryService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"employmentTypeId", "employmentTypeName"}, "employmentTypes")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return dir.ListEmploymentTypes(c, count)
			}, "employmentTypes")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"employmentTypes": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := dir.ListEmploymentTypes(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	for _, cmd := range []*cobra.Command{dirListUsersCmd, dirListGroupsCmd, dirListOrgUnitsCmd, dirListLevelsCmd, dirListPositionsCmd, dirListUserTypesCmd, dirListEmploymentTypesCmd} {
		cmd.Flags().String("cursor", "", "페이지네이션 커서")
		cmd.Flags().Int("count", 0, "페이지 크기")
		cmd.Flags().Bool("all", false, "전체 페이지 자동 순회")
	}
	directoryCmd.AddCommand(dirListUsersCmd, dirGetUserCmd, dirListGroupsCmd, dirGetGroupCmd,
		dirListOrgUnitsCmd, dirGetOrgUnitCmd, dirListLevelsCmd, dirListPositionsCmd,
		dirListUserTypesCmd, dirListEmploymentTypesCmd)
	rootCmd.AddCommand(directoryCmd)
}
