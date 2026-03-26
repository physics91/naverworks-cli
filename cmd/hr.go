package cmd

import (
	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var hrCmd = &cobra.Command{
	Use:   "hr",
	Short: "인사 관리",
}

var hrListExtensionPropertiesCmd = &cobra.Command{
	Use:   "list-extension-properties",
	Short: "확장 속성 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewHRService(client)
		return runListCmd(cmd, []string{"propertyId", "propertyName"}, "extensionProperties", svc.ListExtensionProperties)
	},
}

var hrGetUserPropertiesCmd = &cobra.Command{
	Use:   "get-user-properties <userId>",
	Short: "사용자 확장 속성 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewHRService(client)

		resp, err := svc.GetUserExtensionProperties(args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var hrListLeaveTypesCmd = &cobra.Command{
	Use:   "list-leave-types",
	Short: "휴직 유형 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewHRService(client)
		return runListCmd(cmd, []string{"leaveTypeId", "leaveTypeName"}, "leaveOfAbsences", svc.ListLeaveOfAbsences)
	},
}

var hrListOnLeaveCmd = &cobra.Command{
	Use:   "list-on-leave",
	Short: "휴직 중 사용자 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewHRService(client)
		return runListCmd(cmd, []string{"userId", "userName"}, "onLeaveUsers", svc.ListOnLeaveUsers)
	},
}

func init() {
	addListFlags(hrListExtensionPropertiesCmd, hrListLeaveTypesCmd, hrListOnLeaveCmd)

	hrCmd.AddCommand(hrListExtensionPropertiesCmd, hrGetUserPropertiesCmd, hrListLeaveTypesCmd, hrListOnLeaveCmd)
	rootCmd.AddCommand(hrCmd)
}
