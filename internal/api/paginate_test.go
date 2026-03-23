package api

import (
	"encoding/json"
	"fmt"
	"testing"
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
