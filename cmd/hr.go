package cmd

import (
	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var hrCmd = &cobra.Command{
	Use:   "hr",
	Short: "인사 관리",
}

func newHRService() (*api.HRService, error) {
	client, _, _, err := newAPIClient()
	if err != nil {
		return nil, err
	}
	return api.NewHRService(client), nil
}

var hrListExtensionPropertiesCmd = &cobra.Command{
	Use:   "list-extension-properties",
	Short: "확장 속성 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newHRService()
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"propertyId", "propertyName"}, "extensionProperties", svc.ListExtensionProperties)
	},
}

var hrGetUserPropertiesCmd = &cobra.Command{
	Use:   "get-user-properties <userId>",
	Short: "사용자 확장 속성 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewHRService(client).GetUserExtensionProperties(args[0])
		})
	},
}

var hrListLeaveTypesCmd = &cobra.Command{
	Use:   "list-leave-types",
	Short: "휴직 유형 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newHRService()
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"leaveTypeId", "leaveTypeName"}, "leaveOfAbsences", svc.ListLeaveOfAbsences)
	},
}

var hrListOnLeaveCmd = &cobra.Command{
	Use:   "list-on-leave",
	Short: "휴직 중 사용자 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newHRService()
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"userId", "userName"}, "onLeaveUsers", svc.ListOnLeaveUsers)
	},
}

func init() {
	addListFlags(hrListExtensionPropertiesCmd, hrListLeaveTypesCmd, hrListOnLeaveCmd)

	hrCmd.AddCommand(hrListExtensionPropertiesCmd, hrGetUserPropertiesCmd, hrListLeaveTypesCmd, hrListOnLeaveCmd)
	rootCmd.AddCommand(hrCmd)
}
