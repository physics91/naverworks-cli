package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "캘린더 관리",
}

// resolveCalendarUserID는 공통 resolveUserID를 calendar 기본값으로 호출
func resolveCalendarUserID(cmd *cobra.Command, authMethod string, defaultUID string) (string, error) {
	return resolveUserID(cmd, defaultUID, authMethod)
}

var calListCalendarsCmd = &cobra.Command{
	Use:   "list-calendars",
	Short: "캘린더 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveCalendarUserID(cmd, token.AuthMethod, cfg.DefaultCalendarUserID)
		if err != nil {
			return err
		}

		client := buildAPIClient(cfg, token)
		cal := api.NewCalendarService(client)

		useDefault, _ := cmd.Flags().GetBool("default")
		if useDefault {
			resp, err := cal.GetDefaultCalendar(userID)
			if err != nil {
				return err
			}
			output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
			return nil
		}

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"calendarId", "calendarName"}, "calendarPersonals")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return cal.ListCalendars(userID, c, count)
			}, "calendarPersonals")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"calendarPersonals": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := cal.ListCalendars(userID, cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var calListEventsCmd = &cobra.Command{
	Use:   "list-events",
	Short: "일정 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveCalendarUserID(cmd, token.AuthMethod, cfg.DefaultCalendarUserID)
		if err != nil {
			return err
		}

		calendarID, _ := cmd.Flags().GetString("calendar-id")
		from, _ := cmd.Flags().GetString("from")
		until, _ := cmd.Flags().GetString("until")

		if calendarID == "" || from == "" || until == "" {
			return fmt.Errorf("--calendar-id, --from, --until은 필수입니다")
		}

		fromTime, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return fmt.Errorf("--from 형식 오류 (RFC3339): %w", err)
		}
		untilTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			return fmt.Errorf("--until 형식 오류 (RFC3339): %w", err)
		}
		if untilTime.Before(fromTime) {
			return fmt.Errorf("--from이 --until보다 이후입니다")
		}
		if untilTime.Sub(fromTime) > 31*24*time.Hour {
			return fmt.Errorf("--from과 --until 간격은 최대 31일입니다")
		}

		client := buildAPIClient(cfg, token)
		cal := api.NewCalendarService(client)

		resp, err := cal.ListEvents(userID, calendarID, from, until)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).
			WithTable([]string{"eventId", "summary", "start", "end"}, "events").
			PrintRaw(resp.Body)
		return nil
	},
}

var calGetEventCmd = &cobra.Command{
	Use:   "get-event",
	Short: "일정 상세 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveCalendarUserID(cmd, token.AuthMethod, cfg.DefaultCalendarUserID)
		if err != nil {
			return err
		}

		calendarID, _ := cmd.Flags().GetString("calendar-id")
		eventID, _ := cmd.Flags().GetString("event-id")
		if calendarID == "" || eventID == "" {
			return fmt.Errorf("--calendar-id와 --event-id는 필수입니다")
		}

		client := buildAPIClient(cfg, token)
		cal := api.NewCalendarService(client)

		resp, err := cal.GetEvent(userID, calendarID, eventID)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var calCreateEventCmd = &cobra.Command{
	Use:   "create-event",
	Short: "일정 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveCalendarUserID(cmd, token.AuthMethod, cfg.DefaultCalendarUserID)
		if err != nil {
			return err
		}

		calendarID, _ := cmd.Flags().GetString("calendar-id")
		title, _ := cmd.Flags().GetString("title")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		description, _ := cmd.Flags().GetString("description")
		location, _ := cmd.Flags().GetString("location")
		isAllDay, _ := cmd.Flags().GetBool("is-all-day")

		if calendarID == "" || title == "" || start == "" || end == "" {
			return fmt.Errorf("--calendar-id, --title, --start, --end는 필수입니다")
		}

		startTime, err := time.Parse(time.RFC3339, start)
		if err != nil {
			return fmt.Errorf("--start 형식 오류 (RFC3339): %w", err)
		}
		endTime, err := time.Parse(time.RFC3339, end)
		if err != nil {
			return fmt.Errorf("--end 형식 오류 (RFC3339): %w", err)
		}

		event := map[string]interface{}{
			"summary": title,
		}
		if isAllDay {
			event["start"] = map[string]string{"date": startTime.Format("2006-01-02")}
			event["end"] = map[string]string{"date": endTime.Format("2006-01-02")}
			event["isAllDay"] = true
		} else {
			event["start"] = map[string]string{"dateTime": start}
			event["end"] = map[string]string{"dateTime": end}
		}
		if description != "" {
			event["description"] = description
		}
		if location != "" {
			event["location"] = location
		}

		client := buildAPIClient(cfg, token)
		cal := api.NewCalendarService(client)

		resp, err := cal.CreateEvent(userID, calendarID, event)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

func init() {
	for _, cmd := range []*cobra.Command{calListCalendarsCmd, calListEventsCmd, calGetEventCmd, calCreateEventCmd} {
		cmd.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	calListCalendarsCmd.Flags().Bool("default", false, "기본 캘린더만 조회")
	calListCalendarsCmd.Flags().String("cursor", "", "페이지네이션 커서")
	calListCalendarsCmd.Flags().Int("count", 0, "페이지 크기")
	calListCalendarsCmd.Flags().Bool("all", false, "전체 페이지 자동 순회")

	calListEventsCmd.Flags().String("calendar-id", "", "캘린더 ID (필수)")
	calListEventsCmd.Flags().String("from", "", "시작 시간 RFC3339 (필수)")
	calListEventsCmd.Flags().String("until", "", "종료 시간 RFC3339 (필수)")

	calGetEventCmd.Flags().String("calendar-id", "", "캘린더 ID (필수)")
	calGetEventCmd.Flags().String("event-id", "", "이벤트 ID (필수)")

	calCreateEventCmd.Flags().String("calendar-id", "", "캘린더 ID (필수)")
	calCreateEventCmd.Flags().String("title", "", "일정 제목 (필수)")
	calCreateEventCmd.Flags().String("start", "", "시작 시간 RFC3339 (필수)")
	calCreateEventCmd.Flags().String("end", "", "종료 시간 RFC3339 (필수)")
	calCreateEventCmd.Flags().String("description", "", "설명")
	calCreateEventCmd.Flags().String("location", "", "장소")
	calCreateEventCmd.Flags().Bool("is-all-day", false, "종일 일정")

	calendarCmd.AddCommand(calListCalendarsCmd, calListEventsCmd, calGetEventCmd, calCreateEventCmd)
	rootCmd.AddCommand(calendarCmd)
}
