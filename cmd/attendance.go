package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var attendanceCmd = &cobra.Command{
	Use:   "attendance",
	Short: "근태 관리",
}

var attendanceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "근태 상태 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewAttendanceService(client)

		resp, err := svc.GetStatus(userID)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var attendanceClockInCmd = &cobra.Command{
	Use:   "clock-in",
	Short: "출근 기록",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewAttendanceService(client)

		date, _ := cmd.Flags().GetString("date")
		timeVal, _ := cmd.Flags().GetString("time")
		if date == "" || timeVal == "" {
			return fmt.Errorf("--date와 --time은 필수입니다")
		}

		resp, err := svc.ClockIn(userID, date, timeVal)
		if err != nil {
			return err
		}
		if len(resp.Body) == 0 || strings.TrimSpace(string(resp.Body)) == "" {
			fmt.Println("{}")
		} else {
			output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		}
		return nil
	},
}

var attendanceClockOutCmd = &cobra.Command{
	Use:   "clock-out",
	Short: "퇴근 기록",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewAttendanceService(client)

		date, _ := cmd.Flags().GetString("date")
		timeVal, _ := cmd.Flags().GetString("time")
		if date == "" || timeVal == "" {
			return fmt.Errorf("--date와 --time은 필수입니다")
		}

		resp, err := svc.ClockOut(userID, date, timeVal)
		if err != nil {
			return err
		}
		if len(resp.Body) == 0 || strings.TrimSpace(string(resp.Body)) == "" {
			fmt.Println("{}")
		} else {
			output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		}
		return nil
	},
}

var attendanceListAbsencesCmd = &cobra.Command{
	Use:   "list-absences",
	Short: "부재 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewAttendanceService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"absenceId", "userId"}, "absences")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListAbsences(c, count)
			}, "absences")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"absences": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListAbsences(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var attendanceListAnnualLeavesCmd = &cobra.Command{
	Use:   "list-annual-leaves",
	Short: "연차 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewAttendanceService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"userId", "totalDays"}, "annualLeaves")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListAnnualLeaves(c, count)
			}, "annualLeaves")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"annualLeaves": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListAnnualLeaves(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	for _, c := range []*cobra.Command{attendanceStatusCmd, attendanceClockInCmd, attendanceClockOutCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	for _, c := range []*cobra.Command{attendanceListAbsencesCmd, attendanceListAnnualLeavesCmd} {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
		c.Flags().Bool("all", false, "전체 페이지 자동 순회")
	}

	attendanceClockInCmd.Flags().String("date", "", "기준 날짜 YYYY-MM-DD (필수)")
	attendanceClockInCmd.Flags().String("time", "", "출근 시간 HH:mm (필수)")

	attendanceClockOutCmd.Flags().String("date", "", "기준 날짜 YYYY-MM-DD (필수)")
	attendanceClockOutCmd.Flags().String("time", "", "퇴근 시간 HH:mm (필수)")

	attendanceCmd.AddCommand(attendanceStatusCmd, attendanceClockInCmd, attendanceClockOutCmd,
		attendanceListAbsencesCmd, attendanceListAnnualLeavesCmd)
	rootCmd.AddCommand(attendanceCmd)
}
