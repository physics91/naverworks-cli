package cmd

import (
	"encoding/json"
	"os"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
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
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewHRService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"propertyId", "propertyName"}, "extensionProperties")

		if all {
			var allItems []json.RawMessage
			for {
				resp, err := svc.ListExtensionProperties(cursor, count)
				if err != nil {
					return err
				}
				var page struct {
					ExtensionProperties []json.RawMessage `json:"extensionProperties"`
					ResponseMetaData    struct {
						NextCursor string `json:"nextCursor"`
					} `json:"responseMetaData"`
				}
				json.Unmarshal(resp.Body, &page)
				allItems = append(allItems, page.ExtensionProperties...)
				if page.ResponseMetaData.NextCursor == "" {
					break
				}
				cursor = page.ResponseMetaData.NextCursor
			}
			merged, _ := json.Marshal(map[string]interface{}{"extensionProperties": allItems})
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListExtensionProperties(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var hrGetUserPropertiesCmd = &cobra.Command{
	Use:   "get-user-properties <userId>",
	Short: "사용자 확장 속성 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewHRService(client)

		resp, err := svc.GetUserExtensionProperties(args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var hrListLeaveTypesCmd = &cobra.Command{
	Use:   "list-leave-types",
	Short: "휴직 유형 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewHRService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"leaveTypeId", "leaveTypeName"}, "leaveOfAbsences")

		if all {
			var allItems []json.RawMessage
			for {
				resp, err := svc.ListLeaveOfAbsences(cursor, count)
				if err != nil {
					return err
				}
				var page struct {
					LeaveOfAbsences  []json.RawMessage `json:"leaveOfAbsences"`
					ResponseMetaData struct {
						NextCursor string `json:"nextCursor"`
					} `json:"responseMetaData"`
				}
				json.Unmarshal(resp.Body, &page)
				allItems = append(allItems, page.LeaveOfAbsences...)
				if page.ResponseMetaData.NextCursor == "" {
					break
				}
				cursor = page.ResponseMetaData.NextCursor
			}
			merged, _ := json.Marshal(map[string]interface{}{"leaveOfAbsences": allItems})
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListLeaveOfAbsences(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var hrListOnLeaveCmd = &cobra.Command{
	Use:   "list-on-leave",
	Short: "휴직 중 사용자 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewHRService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"userId", "userName"}, "onLeaveUsers")

		if all {
			var allItems []json.RawMessage
			for {
				resp, err := svc.ListOnLeaveUsers(cursor, count)
				if err != nil {
					return err
				}
				var page struct {
					OnLeaveUsers     []json.RawMessage `json:"onLeaveUsers"`
					ResponseMetaData struct {
						NextCursor string `json:"nextCursor"`
					} `json:"responseMetaData"`
				}
				json.Unmarshal(resp.Body, &page)
				allItems = append(allItems, page.OnLeaveUsers...)
				if page.ResponseMetaData.NextCursor == "" {
					break
				}
				cursor = page.ResponseMetaData.NextCursor
			}
			merged, _ := json.Marshal(map[string]interface{}{"onLeaveUsers": allItems})
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListOnLeaveUsers(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	for _, c := range []*cobra.Command{hrListExtensionPropertiesCmd, hrListLeaveTypesCmd, hrListOnLeaveCmd} {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
		c.Flags().Bool("all", false, "전체 페이지 자동 순회")
	}

	hrCmd.AddCommand(hrListExtensionPropertiesCmd, hrGetUserPropertiesCmd, hrListLeaveTypesCmd, hrListOnLeaveCmd)
	rootCmd.AddCommand(hrCmd)
}
