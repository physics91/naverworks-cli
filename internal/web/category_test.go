package web

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCategoryPagingBindsRoomAndResource(t *testing.T) {
	categories := []Category{{ID: "one"}, {ID: "two"}, {ID: "three"}}
	var first struct {
		Categories       []Category
		ResponseMetaData struct{ NextCursor string }
	}
	data, err := categoryPage(categories, "", 1, testRoomID)
	if err != nil || json.Unmarshal(data, &first) != nil || len(first.Categories) != 1 || first.Categories[0].ID != "one" {
		t.Fatalf("first category page: %s, %v", data, err)
	}
	cursor := first.ResponseMetaData.NextCursor
	data, err = categoryPage(categories, cursor, 2, strings.ToUpper(testRoomID))
	if err != nil || json.Unmarshal(data, &first) != nil || len(first.Categories) != 2 || first.Categories[0].ID != "two" || first.ResponseMetaData.NextCursor != "" {
		t.Fatalf("remaining category page: %s, %v", data, err)
	}
	if _, err := parseCategoryCursor(cursor, "00000000-0000-0000-0000-000000000000"); err == nil {
		t.Fatal("category cursor crossed rooms")
	}
	if _, err := parseTaskCursor(cursor, testRoomID); err == nil {
		t.Fatal("category cursor was accepted as a task cursor")
	}
	if _, err := parseCategoryCursor(strings.TrimPrefix(cursor, "categories:"), testRoomID); err == nil {
		t.Fatal("task cursor was accepted as a category cursor")
	}
	if _, err := categoryPage(categories[:1], cursor, 1, testRoomID); err == nil {
		t.Fatal("removed cursor category silently reset the page")
	}
	data, err = categoryPage(nil, "", 1, testRoomID)
	if err != nil || !strings.Contains(string(data), `"categories":[]`) {
		t.Fatalf("empty categories: %s, %v", data, err)
	}
}
