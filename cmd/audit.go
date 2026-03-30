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

func requireTimeRange(cmd *cobra.Command) (startTime, endTime string, err error) {
	startTime, _ = cmd.Flags().GetString("start-time")
	endTime, _ = cmd.Flags().GetString("end-time")
	if startTime == "" || endTime == "" {
		return "", "", fmt.Errorf("--start-time과 --end-time은 필수입니다")
	}
	return startTime, endTime, nil
}

var auditDownloadLogsCmd = &cobra.Command{
	Use:   "download-logs",
	Short: "감사 로그 다운로드",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAuditService)
		if err != nil {
			return err
		}
		startTime, endTime, err := requireTimeRange(cmd)
		if err != nil {
			return err
		}
		service, _ := cmd.Flags().GetString("service")
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
		svc, err := newSvc(api.NewAuditService)
		if err != nil {
			return err
		}
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
		svc, err := newSvc(api.NewMonitoringService)
		if err != nil {
			return err
		}
		startTime, endTime, err := requireTimeRange(cmd)
		if err != nil {
			return err
		}
		downloadURL, err := svc.DownloadMessages(startTime, endTime)
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

var auditCreatePolicyGroupCmd = &cobra.Command{
	Use:   "create-policy-group",
	Short: "감사 정책 그룹 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAuditService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreatePolicyGroup(body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var auditGetPolicyGroupCmd = &cobra.Command{
	Use:   "get-policy-group <policyGroupId>",
	Short: "감사 정책 그룹 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewAuditService(client).GetPolicyGroup(args[0])
		})
	},
}

var auditUpdatePolicyGroupCmd = &cobra.Command{
	Use:   "update-policy-group <policyGroupId>",
	Short: "감사 정책 그룹 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAuditService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.UpdatePolicyGroup(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var auditDeletePolicyGroupCmd = &cobra.Command{
	Use:   "delete-policy-group <policyGroupId>",
	Short: "감사 정책 그룹 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewAuditService(client).DeletePolicyGroup(args[0])
		})
	},
}

var auditAddPolicyMembersCmd = &cobra.Command{
	Use:   "add-policy-members <policyGroupId>",
	Short: "감사 정책 그룹 멤버 추가",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAuditService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.AddPolicyGroupMembers(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var auditListPolicyMembersCmd = &cobra.Command{
	Use:   "list-policy-members <policyGroupId>",
	Short: "감사 정책 그룹 멤버 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAuditService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"userId", "userName"}, "members", func(c string, n int) (*api.Response, error) {
			return svc.ListPolicyGroupMembers(args[0], c, n)
		})
	},
}

var auditRemovePolicyMemberCmd = &cobra.Command{
	Use:   "remove-policy-member <policyGroupId> <userId>",
	Short: "감사 정책 그룹 멤버 제거",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewAuditService(client).RemovePolicyGroupMember(args[0], args[1])
		})
	},
}

func init() {
	for _, c := range []*cobra.Command{auditDownloadLogsCmd, monitoringDownloadMessagesCmd} {
		c.Flags().String("start-time", "", "시작 시간 (필수)")
		c.Flags().String("end-time", "", "종료 시간 (필수)")
	}
	auditDownloadLogsCmd.Flags().String("service", "", "서비스 필터")

	addListFlags(auditListPolicyGroupsCmd, auditListPolicyMembersCmd)

	for _, c := range []*cobra.Command{
		auditCreatePolicyGroupCmd, auditUpdatePolicyGroupCmd, auditAddPolicyMembersCmd,
	} {
		c.Flags().String("json", "", "JSON 데이터 (- 이면 stdin)")
	}

	auditCmd.AddCommand(auditDownloadLogsCmd, auditListPolicyGroupsCmd,
		auditCreatePolicyGroupCmd, auditGetPolicyGroupCmd, auditUpdatePolicyGroupCmd, auditDeletePolicyGroupCmd,
		auditAddPolicyMembersCmd, auditListPolicyMembersCmd, auditRemovePolicyMemberCmd,
	)
	rootCmd.AddCommand(auditCmd)

	monitoringCmd.AddCommand(monitoringDownloadMessagesCmd)
	rootCmd.AddCommand(monitoringCmd)
}
