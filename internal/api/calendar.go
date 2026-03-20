package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type CalendarService struct {
	client *Client
}

func NewCalendarService(client *Client) *CalendarService {
	return &CalendarService{client: client}
}

func (s *CalendarService) ListCalendars(userID, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/calendar-personals", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *CalendarService) GetDefaultCalendar(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/calendar", url.PathEscape(userID)))
}

func (s *CalendarService) ListEvents(userID, calendarID, from, until string) (*Response, error) {
	params := url.Values{
		"fromDateTime":  {from},
		"untilDateTime": {until},
	}
	return s.client.Get(fmt.Sprintf("/users/%s/calendars/%s/events?%s", url.PathEscape(userID), url.PathEscape(calendarID), params.Encode()))
}

func (s *CalendarService) GetEvent(userID, calendarID, eventID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/calendars/%s/events/%s", url.PathEscape(userID), url.PathEscape(calendarID), url.PathEscape(eventID)))
}

func (s *CalendarService) CreateEvent(userID, calendarID string, event map[string]interface{}) (*Response, error) {
	body := map[string]interface{}{
		"eventComponents": []interface{}{event},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("이벤트 직렬화 실패: %w", err)
	}
	return s.client.Post(fmt.Sprintf("/users/%s/calendars/%s/events", url.PathEscape(userID), url.PathEscape(calendarID)), data)
}
