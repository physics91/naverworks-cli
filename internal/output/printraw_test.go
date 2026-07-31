package output

import (
	"bytes"
	"strings"
	"testing"
)

// Every JSON output path funnels through PrintRaw, so sanitization must happen
// there. Previously it was applied only at two call sites, leaving the list and
// pagination paths able to print presigned URLs.
func TestPrintRaw_SanitizesAllFormats(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		columns   []string
		dataKey   string
		body      string
		mustNotBe string
	}{
		{
			name:      "json format, top level",
			format:    "json",
			body:      `{"uploadUrl":"https://s.example/p?sig=LEAKED"}`,
			mustNotBe: "LEAKED",
		},
		{
			name:      "json format, nested",
			format:    "json",
			body:      `{"data":{"uploadUrl":"https://s.example/p?sig=LEAKED"}}`,
			mustNotBe: "LEAKED",
		},
		{
			name:      "json format, list payload",
			format:    "json",
			body:      `{"files":[{"id":"1","uploadUrl":"https://s.example/p?sig=LEAKED"}]}`,
			mustNotBe: "LEAKED",
		},
		{
			name:      "table format falls back to raw on unknown key",
			format:    "table",
			columns:   []string{"id"},
			dataKey:   "files",
			body:      `{"other":{"uploadUrl":"https://s.example/p?sig=LEAKED"}}`,
			mustNotBe: "LEAKED",
		},
		{
			name:      "table format with rows",
			format:    "table",
			columns:   []string{"id", "uploadUrl"},
			dataKey:   "files",
			body:      `{"files":[{"id":"1","uploadUrl":"https://s.example/p?sig=LEAKED"}]}`,
			mustNotBe: "LEAKED",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := NewFormatter(tc.format, &buf)
			if len(tc.columns) > 0 {
				f = f.WithTable(tc.columns, tc.dataKey)
			}
			f.PrintRaw([]byte(tc.body))

			if strings.Contains(buf.String(), tc.mustNotBe) {
				t.Fatalf("presigned URL reached stdout: %s", buf.String())
			}
		})
	}
}

func TestPrintRaw_LeavesOrdinaryPayloadsIntact(t *testing.T) {
	var buf bytes.Buffer
	NewFormatter("json", &buf).PrintRaw([]byte(`{"id":"123","count":42}`))

	out := buf.String()
	if !strings.Contains(out, `"id": "123"`) || !strings.Contains(out, `"count": 42`) {
		t.Fatalf("ordinary payload should print unchanged: %s", out)
	}
}

// Trailing data means the payload was not fully inspected, so the unchecked
// remainder must not be echoed.
// PrintRaw prints an unparsable body verbatim, so a malformed response that
// mentions a presigned field must be withheld instead of echoed.
func TestSanitizeJSON_WithholdsUnparsableBodyMentioningUploadURL(t *testing.T) {
	body := []byte(`{"uploadUrl":"https://s.example/p?sig=LEAKED"`) // truncated JSON

	out := string(SanitizeJSON(body))

	if strings.Contains(out, "LEAKED") {
		t.Fatalf("unparsable body leaked a presigned URL: %s", out)
	}
	if !strings.Contains(out, "출력을 생략") {
		t.Fatalf("expected a withheld-output notice, got: %s", out)
	}
}

func TestSanitizeJSON_PassesThroughUnparsableBodyWithoutSecrets(t *testing.T) {
	body := []byte(`not json at all`)

	if got := string(SanitizeJSON(body)); got != "not json at all" {
		t.Fatalf("benign unparsable body should pass through, got %q", got)
	}
}

func TestSanitizeJSON_DropsUncheckedTrailingData(t *testing.T) {
	body := []byte(`{"ok":true}` + "\n" + `{"uploadUrl":"https://s.example/p?sig=LEAKED"}`)

	out := string(SanitizeJSON(body))

	if strings.Contains(out, "LEAKED") {
		t.Fatalf("trailing payload should not be emitted: %s", out)
	}
	if !strings.Contains(out, `"ok":true`) {
		t.Fatalf("first value should survive: %s", out)
	}
}
