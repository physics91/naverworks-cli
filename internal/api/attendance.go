package api

import (
	"fmt"
	"net/url"
)

type AttendanceService struct {
	client *Client
}

func NewAttendanceService(client *Client) *AttendanceService {
	return &AttendanceService{client: client}
}

func (s *AttendanceService) GetStatus(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/business-support/attendance/users/%s/status", url.PathEscape(userID)))
}

func (s *AttendanceService) ClockIn(userID, baseDate, clockInTime string) (*Response, error) {
	return s.client.PostJSON(
		fmt.Sprintf("/business-support/attendance/users/%s/clock-in", url.PathEscape(userID)),
		map[string]interface{}{"baseDate": baseDate, "clockInTime": clockInTime},
	)
}

func (s *AttendanceService) ClockOut(userID, baseDate, clockOutTime string) (*Response, error) {
	return s.client.PostJSON(
		fmt.Sprintf("/business-support/attendance/users/%s/clock-out", url.PathEscape(userID)),
		map[string]interface{}{"baseDate": baseDate, "clockOutTime": clockOutTime},
	)
}

func (s *AttendanceService) ListAbsences(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/attendance/absences" + BuildPaginationQuery(cursor, count))
}

func (s *AttendanceService) ListAnnualLeaves(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/attendance/annual-leaves" + BuildPaginationQuery(cursor, count))
}

// ─── Timecard ───

func (s *AttendanceService) CreateTimecard(body []byte) (*Response, error) {
	return s.client.Post("/business-support/attendance/timecards", body)
}

func (s *AttendanceService) ListTimecards(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/attendance/timecards" + BuildPaginationQuery(cursor, count))
}

func (s *AttendanceService) GetTimecard(timecardID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/business-support/attendance/timecards/%s", url.PathEscape(timecardID)))
}

func (s *AttendanceService) PatchTimecard(timecardID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/business-support/attendance/timecards/%s", url.PathEscape(timecardID)), body)
}

// ─── Annual Leave ───

func (s *AttendanceService) AdjustAnnualLeave(body []byte) (*Response, error) {
	return s.client.Post("/business-support/attendance/annual-leaves/adjust", body)
}

// ─── Absence Schedule ───

func (s *AttendanceService) ListAbsenceSchedules(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/attendance/absence-schedule" + BuildPaginationQuery(cursor, count))
}

// ─── Absence (CRUD) ───

func (s *AttendanceService) CreateAbsence(body []byte) (*Response, error) {
	return s.client.Post("/business-support/attendance/absences", body)
}

func (s *AttendanceService) GetAbsence(absenceID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/business-support/attendance/absences/%s", url.PathEscape(absenceID)))
}

func (s *AttendanceService) PatchAbsence(absenceID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/business-support/attendance/absences/%s", url.PathEscape(absenceID)), body)
}

func (s *AttendanceService) DeleteAbsence(absenceID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/business-support/attendance/absences/%s", url.PathEscape(absenceID)))
}
