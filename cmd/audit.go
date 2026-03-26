package cmd

import (
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
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
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
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
		printDownloadURL(downloadURL)
		return nil
	},
}

var auditListPolicyGroupsCmd = &cobra.Command{
	Use:   "list-policy-groups",
	Short: "감사 정책 그룹 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
		svc := api.NewAuditService(client)
		return runListCmd(cmd, []string{"policyGroupId", "policyGroupName"}, "policyGroups", svc.ListPolicyGroups)
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
		client, _, _, err := newAPIClient()
		if err != nil {
			return err
		}
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
		printDownloadURL(downloadURL)
		return nil
	},
}

func init() {
	auditDownloadLogsCmd.Flags().String("start-time", "", "시작 시간 (필수)")
	auditDownloadLogsCmd.Flags().String("end-time", "", "종료 시간 (필수)")
	auditDownloadLogsCmd.Flags().String("service", "", "서비스 필터")

	addListFlags(auditListPolicyGroupsCmd)

	auditCmd.AddCommand(auditDownloadLogsCmd, auditListPolicyGroupsCmd)
	rootCmd.AddCommand(auditCmd)

	monitoringDownloadMessagesCmd.Flags().String("start-time", "", "시작 시간 (필수)")
	monitoringDownloadMessagesCmd.Flags().String("end-time", "", "종료 시간 (필수)")

	monitoringCmd.AddCommand(monitoringDownloadMessagesCmd)
	rootCmd.AddCommand(monitoringCmd)
}
