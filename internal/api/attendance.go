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
	data, err := marshalBody(map[string]interface{}{
		"baseDate":    baseDate,
		"clockInTime": clockInTime,
	})
	if err != nil {
		return nil, err
	}
	return s.client.Post(fmt.Sprintf("/business-support/attendance/users/%s/clock-in", url.PathEscape(userID)), data)
}

func (s *AttendanceService) ClockOut(userID, baseDate, clockOutTime string) (*Response, error) {
	data, err := marshalBody(map[string]interface{}{
		"baseDate":     baseDate,
		"clockOutTime": clockOutTime,
	})
	if err != nil {
		return nil, err
	}
	return s.client.Post(fmt.Sprintf("/business-support/attendance/users/%s/clock-out", url.PathEscape(userID)), data)
}

func (s *AttendanceService) ListAbsences(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/attendance/absences" + BuildPaginationQuery(cursor, count))
}

func (s *AttendanceService) ListAnnualLeaves(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/attendance/annual-leaves" + BuildPaginationQuery(cursor, count))
}
