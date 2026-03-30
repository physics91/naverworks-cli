package api

import (
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
	return s.client.PostJSON(
		fmt.Sprintf("/users/%s/calendars/%s/events", url.PathEscape(userID), url.PathEscape(calendarID)),
		map[string]interface{}{"eventComponents": []interface{}{event}},
	)
}

// --- Calendar CRUD ---

func (s *CalendarService) CreateCalendar(body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON("/calendars", body)
}

func (s *CalendarService) GetCalendar(calendarID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/calendars/%s", url.PathEscape(calendarID)))
}

func (s *CalendarService) PatchCalendar(calendarID string, body map[string]interface{}) (*Response, error) {
	return s.client.PatchJSON(fmt.Sprintf("/calendars/%s", url.PathEscape(calendarID)), body)
}

func (s *CalendarService) DeleteCalendar(calendarID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/calendars/%s", url.PathEscape(calendarID)))
}

// --- Calendar Personal ---

func (s *CalendarService) GetCalendarPersonal(userID, calendarID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/calendar-personals/%s", url.PathEscape(userID), url.PathEscape(calendarID)))
}

func (s *CalendarService) PatchCalendarPersonal(userID, calendarID string, body map[string]interface{}) (*Response, error) {
	return s.client.PatchJSON(fmt.Sprintf("/users/%s/calendar-personals/%s", url.PathEscape(userID), url.PathEscape(calendarID)), body)
}

// --- User Calendar Membership ---

func (s *CalendarService) RemoveUserFromCalendar(userID, calendarID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/calendars/%s", url.PathEscape(userID), url.PathEscape(calendarID)))
}

// --- Event Update/Delete (specific calendar) ---

func (s *CalendarService) UpdateEvent(userID, calendarID, eventID string, body map[string]interface{}) (*Response, error) {
	return s.client.PutJSON(fmt.Sprintf("/users/%s/calendars/%s/events/%s", url.PathEscape(userID), url.PathEscape(calendarID), url.PathEscape(eventID)), body)
}

func (s *CalendarService) DeleteEvent(userID, calendarID, eventID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/calendars/%s/events/%s", url.PathEscape(userID), url.PathEscape(calendarID), url.PathEscape(eventID)))
}

// --- Default Calendar Events ---

func (s *CalendarService) CreateDefaultEvent(userID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/users/%s/calendar/events", url.PathEscape(userID)), body)
}

func (s *CalendarService) ListDefaultEvents(userID, from, until string) (*Response, error) {
	params := url.Values{
		"fromDateTime":  {from},
		"untilDateTime": {until},
	}
	return s.client.Get(fmt.Sprintf("/users/%s/calendar/events?%s", url.PathEscape(userID), params.Encode()))
}

func (s *CalendarService) GetDefaultEvent(userID, eventID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/calendar/events/%s", url.PathEscape(userID), url.PathEscape(eventID)))
}

func (s *CalendarService) UpdateDefaultEvent(userID, eventID string, body map[string]interface{}) (*Response, error) {
	return s.client.PutJSON(fmt.Sprintf("/users/%s/calendar/events/%s", url.PathEscape(userID), url.PathEscape(eventID)), body)
}

func (s *CalendarService) DeleteDefaultEvent(userID, eventID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/calendar/events/%s", url.PathEscape(userID), url.PathEscape(eventID)))
}
