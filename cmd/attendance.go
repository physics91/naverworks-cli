package cmd

import (
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
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
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewAttendanceService(client)

		resp, err := svc.GetStatus(userID)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var attendanceClockInCmd = &cobra.Command{
	Use:   "clock-in",
	Short: "출근 기록",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClockCmd(cmd, func(svc *api.AttendanceService, userID, date, timeVal string) (*api.Response, error) {
			return svc.ClockIn(userID, date, timeVal)
		})
	},
}

var attendanceClockOutCmd = &cobra.Command{
	Use:   "clock-out",
	Short: "퇴근 기록",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClockCmd(cmd, func(svc *api.AttendanceService, userID, date, timeVal string) (*api.Response, error) {
			return svc.ClockOut(userID, date, timeVal)
		})
	},
}

func runClockCmd(cmd *cobra.Command, fn func(*api.AttendanceService, string, string, string) (*api.Response, error)) error {
	cfg, token, name, err := loadConfigAndToken()
	if err != nil {
		return err
	}
	userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
	if err != nil {
		return err
	}
	client := buildAPIClient(cfg, token, name)
	svc := api.NewAttendanceService(client)

	date, _ := cmd.Flags().GetString("date")
	timeVal, _ := cmd.Flags().GetString("time")
	if date == "" || timeVal == "" {
		return fmt.Errorf("--date와 --time은 필수입니다")
	}

	resp, err := fn(svc, userID, date, timeVal)
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

var attendanceListAbsencesCmd = &cobra.Command{
	Use:   "list-absences",
	Short: "부재 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewAttendanceService(client)
		return runListCmd(cmd, []string{"absenceId", "userId"}, "absences", svc.ListAbsences)
	},
}

var attendanceListAnnualLeavesCmd = &cobra.Command{
	Use:   "list-annual-leaves",
	Short: "연차 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewAttendanceService(client)
		return runListCmd(cmd, []string{"userId", "totalDays"}, "annualLeaves", svc.ListAnnualLeaves)
	},
}

func init() {
	for _, c := range []*cobra.Command{attendanceStatusCmd, attendanceClockInCmd, attendanceClockOutCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	addListFlags(attendanceListAbsencesCmd, attendanceListAnnualLeavesCmd)

	attendanceClockInCmd.Flags().String("date", "", "기준 날짜 YYYY-MM-DD (필수)")
	attendanceClockInCmd.Flags().String("time", "", "출근 시간 HH:mm (필수)")

	attendanceClockOutCmd.Flags().String("date", "", "기준 날짜 YYYY-MM-DD (필수)")
	attendanceClockOutCmd.Flags().String("time", "", "퇴근 시간 HH:mm (필수)")

	attendanceCmd.AddCommand(attendanceStatusCmd, attendanceClockInCmd, attendanceClockOutCmd,
		attendanceListAbsencesCmd, attendanceListAnnualLeavesCmd)
	rootCmd.AddCommand(attendanceCmd)
}
