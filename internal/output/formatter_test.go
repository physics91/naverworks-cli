package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("json", &buf)
	data := map[string]string{"name": "test", "id": "123"}
	f.Print(data)

	got := buf.String()
	if !strings.Contains(got, `"name"`) {
		t.Errorf("expected JSON with name field, got %s", got)
	}
}

func TestPrintRaw_JSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("json", &buf)
	f.PrintRaw([]byte(`{"users":[{"id":"1"}]}`))

	got := buf.String()
	if !strings.Contains(got, "users") {
		t.Errorf("expected pretty JSON, got %s", got)
	}
}

func TestTable_WithDataKey(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("table", &buf).WithTable([]string{"name", "id"}, "items")
	f.PrintRaw([]byte(`{"items":[{"name":"Alice","id":"1"},{"name":"Bob","id":"2"}]}`))

	got := buf.String()
	if !strings.Contains(got, "Alice") {
		t.Errorf("expected Alice in table, got %s", got)
	}
	if !strings.Contains(got, "Bob") {
		t.Errorf("expected Bob in table, got %s", got)
	}
}

func TestTable_NoDataKey_FallbackJSON(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("table", &buf).WithTable([]string{"name"}, "items")
	f.PrintRaw([]byte(`{"other":"data"}`))

	got := buf.String()
	if !strings.Contains(got, "other") {
		t.Errorf("expected fallback JSON output, got %s", got)
	}
}

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter("table", &buf)
	f.PrintTable([]string{"ID", "Name"}, [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	})

	got := buf.String()
	if !strings.Contains(got, "Alice") {
		t.Errorf("expected Alice in table, got %s", got)
	}
}
