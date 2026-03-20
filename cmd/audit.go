package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "감사 로그 관리",
}

var auditDownloadLogsCmd = &cobra.Command{
	Use:   "download-logs",
	Short: "감사 로그 다운로드",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewAuditService(client)

		from, _ := cmd.Flags().GetString("from")
		until, _ := cmd.Flags().GetString("until")
		if from == "" || until == "" {
			return fmt.Errorf("--from과 --until은 필수입니다")
		}

		resp, err := svc.DownloadLogs(from, until)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var auditListPolicyGroupsCmd = &cobra.Command{
	Use:   "list-policy-groups",
	Short: "감사 정책 그룹 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewAuditService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"policyGroupId", "policyGroupName"}, "policyGroups")

		if all {
			var allItems []json.RawMessage
			for {
				resp, err := svc.ListPolicyGroups(cursor, count)
				if err != nil {
					return err
				}
				var page struct {
					PolicyGroups     []json.RawMessage `json:"policyGroups"`
					ResponseMetaData struct {
						NextCursor string `json:"nextCursor"`
					} `json:"responseMetaData"`
				}
				json.Unmarshal(resp.Body, &page)
				allItems = append(allItems, page.PolicyGroups...)
				if page.ResponseMetaData.NextCursor == "" {
					break
				}
				cursor = page.ResponseMetaData.NextCursor
			}
			merged, _ := json.Marshal(map[string]interface{}{"policyGroups": allItems})
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListPolicyGroups(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var monitoringCmd = &cobra.Command{
	Use:   "monitoring",
	Short: "모니터링 관리",
}

var monitoringDownloadMessagesCmd = &cobra.Command{
	Use:   "download-messages",
	Short: "메시지 콘텐츠 다운로드",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewMonitoringService(client)

		from, _ := cmd.Flags().GetString("from")
		until, _ := cmd.Flags().GetString("until")
		if from == "" || until == "" {
			return fmt.Errorf("--from과 --until은 필수입니다")
		}

		resp, err := svc.DownloadMessages(from, until)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	auditDownloadLogsCmd.Flags().String("from", "", "시작 날짜 (필수)")
	auditDownloadLogsCmd.Flags().String("until", "", "종료 날짜 (필수)")

	auditListPolicyGroupsCmd.Flags().String("cursor", "", "페이지네이션 커서")
	auditListPolicyGroupsCmd.Flags().Int("count", 0, "페이지 크기")
	auditListPolicyGroupsCmd.Flags().Bool("all", false, "전체 페이지 자동 순회")

	auditCmd.AddCommand(auditDownloadLogsCmd, auditListPolicyGroupsCmd)
	rootCmd.AddCommand(auditCmd)

	monitoringDownloadMessagesCmd.Flags().String("from", "", "시작 날짜 (필수)")
	monitoringDownloadMessagesCmd.Flags().String("until", "", "종료 날짜 (필수)")

	monitoringCmd.AddCommand(monitoringDownloadMessagesCmd)
	rootCmd.AddCommand(monitoringCmd)
}
