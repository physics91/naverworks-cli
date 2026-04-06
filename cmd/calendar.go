package cmd

import (
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

var calListCalendarsCmd = &cobra.Command{
	Use:   "list-calendars",
	Short: "캘린더 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cal := api.NewCalendarService(client)

		useDefault, _ := cmd.Flags().GetBool("default")
		if useDefault {
			resp, err := cal.GetDefaultCalendar(userID)
			if err != nil {
				return err
			}
			printBody(resp.Body)
			return nil
		}

		return runListCmd(cmd, []string{"calendarId", "calendarName"}, "calendarPersonals", func(c string, n int) (*api.Response, error) {
			return cal.ListCalendars(userID, c, n)
		})
	},
}

var calListEventsCmd = &cobra.Command{
	Use:   "list-events",
	Short: "일정 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
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
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}

		calendarID, _ := cmd.Flags().GetString("calendar-id")
		eventID, _ := cmd.Flags().GetString("event-id")
		if calendarID == "" || eventID == "" {
			return fmt.Errorf("--calendar-id와 --event-id는 필수입니다")
		}

		cal := api.NewCalendarService(client)

		resp, err := cal.GetEvent(userID, calendarID, eventID)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var calCreateEventCmd = &cobra.Command{
	Use:   "create-event",
	Short: "일정 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
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
		timezone, _ := cmd.Flags().GetString("timezone")

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
			startPayload := map[string]string{
				"dateTime": startTime.Format(time.RFC3339),
			}
			endPayload := map[string]string{
				"dateTime": endTime.Format(time.RFC3339),
			}
			if timezone != "" {
				loc, err := time.LoadLocation(timezone)
				if err != nil {
					return fmt.Errorf("--timezone 형식 오류: %w", err)
				}
				startPayload["dateTime"] = startTime.In(loc).Format("2006-01-02T15:04:05")
				endPayload["dateTime"] = endTime.In(loc).Format("2006-01-02T15:04:05")
				startPayload["timeZone"] = timezone
				endPayload["timeZone"] = timezone
			}
			event["start"] = startPayload
			event["end"] = endPayload
		}
		if description != "" {
			event["description"] = description
		}
		if location != "" {
			event["location"] = location
		}

		cal := api.NewCalendarService(client)

		resp, err := cal.CreateEvent(userID, calendarID, event)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

// --- Calendar CRUD commands ---

var calCreateCalendarCmd = &cobra.Command{
	Use:   "create-calendar",
	Short: "캘린더 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewCalendarService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateCalendar(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var calGetCalendarCmd = &cobra.Command{
	Use:   "get-calendar <calendarId>",
	Short: "캘린더 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewCalendarService(client).GetCalendar(args[0])
		})
	},
}

var calUpdateCalendarCmd = &cobra.Command{
	Use:   "update-calendar <calendarId>",
	Short: "캘린더 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewCalendarService)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchCalendar(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var calDeleteCalendarCmd = &cobra.Command{
	Use:   "delete-calendar <calendarId>",
	Short: "캘린더 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewCalendarService)
		if err != nil {
			return err
		}
		resp, err := svc.DeleteCalendar(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// --- Calendar Personal commands ---

var calGetPersonalCmd = &cobra.Command{
	Use:   "get-personal <calendarId>",
	Short: "개인 캘린더 설정 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cal := api.NewCalendarService(client)
		resp, err := cal.GetCalendarPersonal(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var calUpdatePersonalCmd = &cobra.Command{
	Use:   "update-personal <calendarId>",
	Short: "개인 캘린더 설정 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		cal := api.NewCalendarService(client)
		resp, err := cal.PatchCalendarPersonal(userID, args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

// --- User Calendar Membership ---

var calRemoveUserCmd = &cobra.Command{
	Use:   "remove-user <calendarId>",
	Short: "사용자를 캘린더에서 제거",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cal := api.NewCalendarService(client)
		resp, err := cal.RemoveUserFromCalendar(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// --- Event Update/Delete (specific calendar) ---

var calUpdateEventCmd = &cobra.Command{
	Use:   "update-event <calendarId> <eventId>",
	Short: "일정 수정",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		cal := api.NewCalendarService(client)
		resp, err := cal.UpdateEvent(userID, args[0], args[1], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var calDeleteEventCmd = &cobra.Command{
	Use:   "delete-event <calendarId> <eventId>",
	Short: "일정 삭제",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cal := api.NewCalendarService(client)
		resp, err := cal.DeleteEvent(userID, args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

// --- Default Calendar subcommand group ---

var calDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "기본 캘린더 일정 관리",
}

var calDefaultListEventsCmd = &cobra.Command{
	Use:   "list-events",
	Short: "기본 캘린더 일정 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}

		from, _ := cmd.Flags().GetString("from")
		until, _ := cmd.Flags().GetString("until")
		if from == "" || until == "" {
			return fmt.Errorf("--from, --until은 필수입니다")
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

		cal := api.NewCalendarService(client)
		resp, err := cal.ListDefaultEvents(userID, from, until)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).
			WithTable([]string{"eventId", "summary", "start", "end"}, "events").
			PrintRaw(resp.Body)
		return nil
	},
}

var calDefaultGetEventCmd = &cobra.Command{
	Use:   "get-event <eventId>",
	Short: "기본 캘린더 일정 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cal := api.NewCalendarService(client)
		resp, err := cal.GetDefaultEvent(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var calDefaultCreateEventCmd = &cobra.Command{
	Use:   "create-event",
	Short: "기본 캘린더 일정 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		cal := api.NewCalendarService(client)
		resp, err := cal.CreateDefaultEvent(userID, body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var calDefaultUpdateEventCmd = &cobra.Command{
	Use:   "update-event <eventId>",
	Short: "기본 캘린더 일정 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		cal := api.NewCalendarService(client)
		resp, err := cal.UpdateDefaultEvent(userID, args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var calDefaultDeleteEventCmd = &cobra.Command{
	Use:   "delete-event <eventId>",
	Short: "기본 캘린더 일정 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		cal := api.NewCalendarService(client)
		resp, err := cal.DeleteDefaultEvent(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

func init() {
	// Existing commands: user-id flag
	for _, cmd := range []*cobra.Command{calListCalendarsCmd, calListEventsCmd, calGetEventCmd, calCreateEventCmd} {
		cmd.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	calListCalendarsCmd.Flags().Bool("default", false, "기본 캘린더만 조회")
	addListFlags(calListCalendarsCmd)

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
	calCreateEventCmd.Flags().String("timezone", "Asia/Seoul", "타임존 (IANA)")

	// New Calendar CRUD commands
	calCreateCalendarCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	calUpdateCalendarCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")

	// Calendar Personal commands: user-id flag
	for _, c := range []*cobra.Command{calGetPersonalCmd, calUpdatePersonalCmd, calRemoveUserCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	calUpdatePersonalCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")

	// Event update/delete commands: user-id flag
	for _, c := range []*cobra.Command{calUpdateEventCmd, calDeleteEventCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	calUpdateEventCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")

	// Default calendar subcommands: user-id flag
	for _, c := range []*cobra.Command{calDefaultListEventsCmd, calDefaultGetEventCmd, calDefaultCreateEventCmd, calDefaultUpdateEventCmd, calDefaultDeleteEventCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	calDefaultListEventsCmd.Flags().String("from", "", "시작 시간 RFC3339 (필수)")
	calDefaultListEventsCmd.Flags().String("until", "", "종료 시간 RFC3339 (필수)")
	calDefaultCreateEventCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")
	calDefaultUpdateEventCmd.Flags().String("json", "", "JSON 페이로드 (필수, -: stdin)")

	// Register default subcommands
	calDefaultCmd.AddCommand(calDefaultListEventsCmd, calDefaultGetEventCmd, calDefaultCreateEventCmd, calDefaultUpdateEventCmd, calDefaultDeleteEventCmd)

	// Register all to calendarCmd
	calendarCmd.AddCommand(calListCalendarsCmd, calListEventsCmd, calGetEventCmd, calCreateEventCmd,
		calCreateCalendarCmd, calGetCalendarCmd, calUpdateCalendarCmd, calDeleteCalendarCmd,
		calGetPersonalCmd, calUpdatePersonalCmd, calRemoveUserCmd,
		calUpdateEventCmd, calDeleteEventCmd,
		calDefaultCmd)
	rootCmd.AddCommand(calendarCmd)
}
