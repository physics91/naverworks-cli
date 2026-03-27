package cmd

import (
	"github.com/physics91/naverworks-cli/internal/api"
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
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		dir := api.NewDirectoryService(client)
		return runListCmd(cmd, []string{"userId", "userName", "email"}, "users", dir.ListUsers)
	},
}

var dirGetUserCmd = &cobra.Command{
	Use:   "get-user <userId>",
	Short: "사용자 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewDirectoryService(client).GetUser(args[0])
		})
	},
}

var dirListGroupsCmd = &cobra.Command{
	Use:   "list-groups",
	Short: "그룹 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		dir := api.NewDirectoryService(client)
		return runListCmd(cmd, []string{"groupId", "groupName"}, "groups", dir.ListGroups)
	},
}

var dirGetGroupCmd = &cobra.Command{
	Use:   "get-group <groupId>",
	Short: "그룹 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewDirectoryService(client).GetGroup(args[0])
		})
	},
}

var dirListOrgUnitsCmd = &cobra.Command{
	Use:   "list-orgunits",
	Short: "조직 단위 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		dir := api.NewDirectoryService(client)
		return runListCmd(cmd, []string{"orgUnitId", "orgUnitName"}, "orgUnits", dir.ListOrgUnits)
	},
}

var dirGetOrgUnitCmd = &cobra.Command{
	Use:   "get-orgunit <orgUnitId>",
	Short: "조직 단위 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewDirectoryService(client).GetOrgUnit(args[0])
		})
	},
}

var dirListLevelsCmd = &cobra.Command{
	Use:   "list-levels",
	Short: "직급 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		dir := api.NewDirectoryService(client)
		return runListCmd(cmd, []string{"levelId", "levelName"}, "levels", dir.ListLevels)
	},
}

var dirListPositionsCmd = &cobra.Command{
	Use:   "list-positions",
	Short: "직책 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		dir := api.NewDirectoryService(client)
		return runListCmd(cmd, []string{"positionId", "positionName"}, "positions", dir.ListPositions)
	},
}

var dirListUserTypesCmd = &cobra.Command{
	Use:   "list-user-types",
	Short: "사용자 유형 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		dir := api.NewDirectoryService(client)
		return runListCmd(cmd, []string{"userTypeId", "userTypeName"}, "userTypes", dir.ListUserTypes)
	},
}

var dirListEmploymentTypesCmd = &cobra.Command{
	Use:   "list-employment-types",
	Short: "고용 유형 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		dir := api.NewDirectoryService(client)
		return runListCmd(cmd, []string{"employmentTypeId", "employmentTypeName"}, "employmentTypes", dir.ListEmploymentTypes)
	},
}

func init() {
	addListFlags(dirListUsersCmd, dirListGroupsCmd, dirListOrgUnitsCmd, dirListLevelsCmd, dirListPositionsCmd, dirListUserTypesCmd, dirListEmploymentTypesCmd)
	directoryCmd.AddCommand(dirListUsersCmd, dirGetUserCmd, dirListGroupsCmd, dirGetGroupCmd,
		dirListOrgUnitsCmd, dirGetOrgUnitCmd, dirListLevelsCmd, dirListPositionsCmd,
		dirListUserTypesCmd, dirListEmploymentTypesCmd)
	rootCmd.AddCommand(directoryCmd)
}
