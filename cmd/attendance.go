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
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		resp, err := api.NewAttendanceService(client).GetStatus(userID)
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
	client, userID, err := newAPIClientWithUser(cmd)
	if err != nil {
		return err
	}
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
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"absenceId", "userId"}, "absences", svc.ListAbsences)
	},
}

var attendanceListAnnualLeavesCmd = &cobra.Command{
	Use:   "list-annual-leaves",
	Short: "연차 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"userId", "totalDays"}, "annualLeaves", svc.ListAnnualLeaves)
	},
}

// ─── Timecard Commands ───

var attendanceCreateTimecardCmd = &cobra.Command{
	Use:   "create-timecard",
	Short: "타임카드 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateTimecard(body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var attendanceListTimecardsCmd = &cobra.Command{
	Use:   "list-timecards",
	Short: "타임카드 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"timecardId", "userId"}, "timecards", svc.ListTimecards)
	},
}

var attendanceGetTimecardCmd = &cobra.Command{
	Use:   "get-timecard <timecardId>",
	Short: "타임카드 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		resp, err := svc.GetTimecard(args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var attendanceUpdateTimecardCmd = &cobra.Command{
	Use:   "update-timecard <timecardId>",
	Short: "타임카드 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchTimecard(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Annual Leave Command ───

var attendanceAdjustAnnualLeaveCmd = &cobra.Command{
	Use:   "adjust-annual-leave",
	Short: "연차 조정",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.AdjustAnnualLeave(body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// ─── Absence Schedule Command ───

var attendanceListAbsenceSchedulesCmd = &cobra.Command{
	Use:   "list-absence-schedules",
	Short: "부재 스케줄 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		return runListCmd(cmd, []string{"absenceId", "userId"}, "absenceSchedules", svc.ListAbsenceSchedules)
	},
}

// ─── Absence CRUD Commands ───

var attendanceCreateAbsenceCmd = &cobra.Command{
	Use:   "create-absence",
	Short: "부재 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateAbsence(body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var attendanceGetAbsenceCmd = &cobra.Command{
	Use:   "get-absence <absenceId>",
	Short: "부재 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		resp, err := svc.GetAbsence(args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var attendanceUpdateAbsenceCmd = &cobra.Command{
	Use:   "update-absence <absenceId>",
	Short: "부재 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		body, err := readJSONFlagRaw(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchAbsence(args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var attendanceDeleteAbsenceCmd = &cobra.Command{
	Use:   "delete-absence <absenceId>",
	Short: "부재 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewAttendanceService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteAbsence(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

func init() {
	for _, c := range []*cobra.Command{attendanceStatusCmd, attendanceClockInCmd, attendanceClockOutCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	addListFlags(attendanceListAbsencesCmd, attendanceListAnnualLeavesCmd,
		attendanceListTimecardsCmd, attendanceListAbsenceSchedulesCmd)

	for _, c := range []*cobra.Command{attendanceClockInCmd, attendanceClockOutCmd} {
		c.Flags().String("date", "", "기준 날짜 YYYY-MM-DD (필수)")
	}
	attendanceClockInCmd.Flags().String("time", "", "출근 시간 HH:mm (필수)")
	attendanceClockOutCmd.Flags().String("time", "", "퇴근 시간 HH:mm (필수)")

	for _, c := range []*cobra.Command{
		attendanceCreateTimecardCmd, attendanceUpdateTimecardCmd,
		attendanceAdjustAnnualLeaveCmd,
		attendanceCreateAbsenceCmd, attendanceUpdateAbsenceCmd,
	} {
		c.Flags().String("json", "", "JSON 본문 ('-'이면 stdin)")
	}

	attendanceCmd.AddCommand(attendanceStatusCmd, attendanceClockInCmd, attendanceClockOutCmd,
		attendanceListAbsencesCmd, attendanceListAnnualLeavesCmd,
		attendanceCreateTimecardCmd, attendanceListTimecardsCmd, attendanceGetTimecardCmd, attendanceUpdateTimecardCmd,
		attendanceAdjustAnnualLeaveCmd,
		attendanceListAbsenceSchedulesCmd,
		attendanceCreateAbsenceCmd, attendanceGetAbsenceCmd, attendanceUpdateAbsenceCmd, attendanceDeleteAbsenceCmd)
	rootCmd.AddCommand(attendanceCmd)
}
