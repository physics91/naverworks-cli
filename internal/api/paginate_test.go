package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestPaginateAll_SinglePage(t *testing.T) {
	callCount := 0
	fetcher := func(cursor string) (*Response, error) {
		callCount++
		body := []byte(`{"users":[{"id":"1"}],"responseMetaData":{"nextCursor":""}}`)
		return &Response{StatusCode: 200, Body: body}, nil
	}

	result, err := PaginateAll(fetcher, "users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(result, &items); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestPaginateAll_MultiplePages(t *testing.T) {
	callCount := 0
	fetcher := func(cursor string) (*Response, error) {
		callCount++
		if callCount == 1 {
			body := []byte(`{"items":[{"id":"1"}],"responseMetaData":{"nextCursor":"page2"}}`)
			return &Response{StatusCode: 200, Body: body}, nil
		}
		body := []byte(`{"items":[{"id":"2"}],"responseMetaData":{"nextCursor":""}}`)
		return &Response{StatusCode: 200, Body: body}, nil
	}

	result, err := PaginateAll(fetcher, "items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(result, &items); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestPaginateAll_FetchError(t *testing.T) {
	fetcher := func(cursor string) (*Response, error) {
		return nil, fmt.Errorf("network error")
	}

	_, err := PaginateAll(fetcher, "items")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPaginateAll_InvalidJSON(t *testing.T) {
	fetcher := func(cursor string) (*Response, error) {
		return &Response{StatusCode: 200, Body: []byte(`not json`)}, nil
	}

	_, err := PaginateAll(fetcher, "items")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// TestPaginateAll_WithMockServer tests PaginateAll through a real httptest
// server that returns paginated responses, verifying the full flow.
func TestPaginateAll_WithMockServer(t *testing.T) {
	var requestCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := atomic.AddInt32(&requestCount, 1)
		cursor := r.URL.Query().Get("cursor")

		var body string
		switch {
		case page == 1 && cursor == "":
			body = `{"users":[{"id":"1"},{"id":"2"}],"responseMetaData":{"nextCursor":"page2"}}`
		case page == 2 && cursor == "page2":
			body = `{"users":[{"id":"3"}],"responseMetaData":{"nextCursor":"page3"}}`
		case page == 3 && cursor == "page3":
			body = `{"users":[{"id":"4"},{"id":"5"}],"responseMetaData":{"nextCursor":""}}`
		default:
			t.Errorf("unexpected request: page=%d cursor=%q", page, cursor)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(body))
	}))
	defer srv.Close()

	token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(srv.URL, token, nil)

	fetcher := func(cursor string) (*Response, error) {
		query := ""
		if cursor != "" {
			query = "?cursor=" + cursor
		}
		return client.Get("/users" + query)
	}

	result, err := PaginateAll(fetcher, "users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(result, &items); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(items) != 5 {
		t.Errorf("expected 5 items across 3 pages, got %d", len(items))
	}
	if atomic.LoadInt32(&requestCount) != 3 {
		t.Errorf("expected 3 requests, got %d", requestCount)
	}
}
