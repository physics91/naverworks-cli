package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestCalendarService_ListCalendars(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/user1/calendar-personals" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"calendarPersonals":[]}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	cal := NewCalendarService(client)

	resp, err := cal.ListCalendars("user1", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCalendarService_GetDefaultCalendar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/user1/calendar" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"calendarId":"default"}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	cal := NewCalendarService(client)

	resp, err := cal.GetDefaultCalendar("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCalendarService_ListEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/user1/calendars/cal1/events" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		from := r.URL.Query().Get("fromDateTime")
		until := r.URL.Query().Get("untilDateTime")
		if from == "" || until == "" {
			t.Error("expected fromDateTime and untilDateTime")
		}
		w.Write([]byte(`{"events":[]}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	cal := NewCalendarService(client)

	resp, err := cal.ListEvents("user1", "cal1", "2026-03-01T00:00:00Z", "2026-03-31T23:59:59Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestCalendarService_CreateEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"eventId":"ev1"}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	cal := NewCalendarService(client)

	event := map[string]interface{}{
		"summary": "Test Event",
		"start":   map[string]string{"dateTime": "2026-03-20T10:00:00Z"},
		"end":     map[string]string{"dateTime": "2026-03-20T11:00:00Z"},
	}
	resp, err := cal.CreateEvent("user1", "cal1", event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
