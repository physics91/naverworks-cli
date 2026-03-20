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
	return s.client.Get(fmt.Sprintf("/users/%s/calendar-personals", userID) + BuildPaginationQuery(cursor, count))
}

func (s *CalendarService) GetDefaultCalendar(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/calendar", userID))
}

func (s *CalendarService) ListEvents(userID, calendarID, from, until string) (*Response, error) {
	params := url.Values{
		"fromDateTime":  {from},
		"untilDateTime": {until},
	}
	return s.client.Get(fmt.Sprintf("/users/%s/calendars/%s/events?%s", userID, calendarID, params.Encode()))
}

func (s *CalendarService) GetEvent(userID, calendarID, eventID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/calendars/%s/events/%s", userID, calendarID, eventID))
}

func (s *CalendarService) CreateEvent(userID, calendarID string, event map[string]interface{}) (*Response, error) {
	body := map[string]interface{}{
		"eventComponents": []interface{}{event},
	}
	data, _ := json.Marshal(body)
	return s.client.Post(fmt.Sprintf("/users/%s/calendars/%s/events", userID, calendarID), data)
}
