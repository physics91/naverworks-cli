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
		svc, err := newSvc(api.NewHRService)
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
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewHRService(client).GetUserExtensionProperties(args[0])
		})
	},
}

var hrListLeaveTypesCmd = &cobra.Command{
	Use:   "list-leave-types",
	Short: "휴직 유형 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewHRService)
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
		svc, err := newSvc(api.NewHRService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"userId", "userName"}, "onLeaveUsers", svc.ListOnLeaveUsers)
	},
}

var hrCreateExtensionPropertyCmd = &cobra.Command{
	Use:   "create-extension-property",
	Short: "확장 속성 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewHRService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateExtensionProperty(body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var hrGetExtensionPropertyCmd = &cobra.Command{
	Use:   "get-extension-property <id>",
	Short: "확장 속성 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewHRService(client).GetExtensionProperty(args[0])
		})
	},
}

var hrUpdateExtensionPropertyCmd = &cobra.Command{
	Use:   "update-extension-property <id>",
	Short: "확장 속성 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewHRService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchExtensionProperty(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var hrDeleteExtensionPropertyCmd = &cobra.Command{
	Use:   "delete-extension-property <id>",
	Short: "확장 속성 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewHRService(client).DeleteExtensionProperty(args[0])
		})
	},
}

var hrGetUserPropertyCmd = &cobra.Command{
	Use:   "get-user-property <userId> <id>",
	Short: "사용자 확장 속성 상세 조회",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewHRService(client).GetUserExtensionProperty(args[0], args[1])
		})
	},
}

var hrUpdateUserPropertyCmd = &cobra.Command{
	Use:   "update-user-property <userId> <id>",
	Short: "사용자 확장 속성 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewHRService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchUserExtensionProperty(args[0], args[1], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var hrCreateLeaveOfAbsenceCmd = &cobra.Command{
	Use:   "create-leave-of-absence",
	Short: "휴직 유형 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewHRService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateLeaveOfAbsence(body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var hrGetLeaveOfAbsenceCmd = &cobra.Command{
	Use:   "get-leave-of-absence <id>",
	Short: "휴직 유형 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewHRService(client).GetLeaveOfAbsence(args[0])
		})
	},
}

var hrUpdateLeaveOfAbsenceCmd = &cobra.Command{
	Use:   "update-leave-of-absence <id>",
	Short: "휴직 유형 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewHRService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchLeaveOfAbsence(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var hrDeleteLeaveOfAbsenceCmd = &cobra.Command{
	Use:   "delete-leave-of-absence <id>",
	Short: "휴직 유형 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewHRService(client).DeleteLeaveOfAbsence(args[0])
		})
	},
}

func init() {
	addListFlags(hrListExtensionPropertiesCmd, hrListLeaveTypesCmd, hrListOnLeaveCmd)

	for _, c := range []*cobra.Command{
		hrCreateExtensionPropertyCmd, hrUpdateExtensionPropertyCmd,
		hrUpdateUserPropertyCmd,
		hrCreateLeaveOfAbsenceCmd, hrUpdateLeaveOfAbsenceCmd,
	} {
		c.Flags().String("json", "", "JSON 데이터 (- 이면 stdin)")
	}

	hrCmd.AddCommand(
		hrListExtensionPropertiesCmd, hrGetUserPropertiesCmd, hrListLeaveTypesCmd, hrListOnLeaveCmd,
		hrCreateExtensionPropertyCmd, hrGetExtensionPropertyCmd, hrUpdateExtensionPropertyCmd, hrDeleteExtensionPropertyCmd,
		hrGetUserPropertyCmd, hrUpdateUserPropertyCmd,
		hrCreateLeaveOfAbsenceCmd, hrGetLeaveOfAbsenceCmd, hrUpdateLeaveOfAbsenceCmd, hrDeleteLeaveOfAbsenceCmd,
	)
	rootCmd.AddCommand(hrCmd)
}
