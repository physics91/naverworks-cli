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

		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		service, _ := cmd.Flags().GetString("service")
		if startTime == "" || endTime == "" {
			return fmt.Errorf("--start-time과 --end-time은 필수입니다")
		}

		downloadURL, err := svc.DownloadLogs(startTime, endTime, service)
		if err != nil {
			return err
		}
		result, _ := json.Marshal(map[string]string{"download_url": downloadURL})
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(result)
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
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListPolicyGroups(c, count)
			}, "policyGroups")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"policyGroups": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
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

		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		if startTime == "" || endTime == "" {
			return fmt.Errorf("--start-time과 --end-time은 필수입니다")
		}

		downloadURL, err := svc.DownloadMessages(startTime, endTime)
		if err != nil {
			return err
		}
		result, _ := json.Marshal(map[string]string{"download_url": downloadURL})
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(result)
		return nil
	},
}

func init() {
	auditDownloadLogsCmd.Flags().String("start-time", "", "시작 시간 (필수)")
	auditDownloadLogsCmd.Flags().String("end-time", "", "종료 시간 (필수)")
	auditDownloadLogsCmd.Flags().String("service", "", "서비스 필터")

	auditListPolicyGroupsCmd.Flags().String("cursor", "", "페이지네이션 커서")
	auditListPolicyGroupsCmd.Flags().Int("count", 0, "페이지 크기")
	auditListPolicyGroupsCmd.Flags().Bool("all", false, "전체 페이지 자동 순회")

	auditCmd.AddCommand(auditDownloadLogsCmd, auditListPolicyGroupsCmd)
	rootCmd.AddCommand(auditCmd)

	monitoringDownloadMessagesCmd.Flags().String("start-time", "", "시작 시간 (필수)")
	monitoringDownloadMessagesCmd.Flags().String("end-time", "", "종료 시간 (필수)")

	monitoringCmd.AddCommand(monitoringDownloadMessagesCmd)
	rootCmd.AddCommand(monitoringCmd)
}
