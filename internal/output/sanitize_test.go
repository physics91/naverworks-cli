package output

import (
	"strings"
	"testing"
)

func TestSanitizeJSON_RedactsUploadURL(t *testing.T) {
	body := []byte(`{"uploadUrl":"https://example.com/upload?sig=secret","offset":7}`)

	sanitized := SanitizeJSON(body)

	if strings.Contains(string(sanitized), "uploadUrl") {
		t.Fatalf("uploadUrl should be redacted: %s", sanitized)
	}
	if strings.Contains(string(sanitized), "secret") {
		t.Fatalf("signature-bearing URL should not remain: %s", sanitized)
	}
	if !strings.Contains(string(sanitized), `"uploaded":true`) {
		t.Fatalf("expected uploaded marker: %s", sanitized)
	}
	if !strings.Contains(string(sanitized), `"offset":7`) {
		t.Fatalf("expected safe fields to remain: %s", sanitized)
	}
}

func TestSanitizeJSON_RedactsNestedUploadURL(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "nested object", body: `{"result":{"uploadUrl":"https://example.com/upload?sig=secret"}}`},
		{name: "array element", body: `[{"uploadUrl":"https://example.com/upload?sig=secret"}]`},
		{name: "deeply nested", body: `{"a":{"b":[{"c":{"upload_url":"https://example.com/upload?sig=secret"}}]}}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sanitized := string(SanitizeJSON([]byte(tc.body)))
			if strings.Contains(sanitized, "secret") {
				t.Fatalf("nested signature-bearing URL should be removed: %s", sanitized)
			}
			if strings.Contains(sanitized, "uploadUrl") || strings.Contains(sanitized, "upload_url") {
				t.Fatalf("nested uploadUrl key should be removed: %s", sanitized)
			}
		})
	}
}

// Re-serialization must not silently reformat numbers that exceed float64
// precision (file sizes, resource IDs).
func TestSanitizeJSON_PreservesNumberPrecision(t *testing.T) {
	body := []byte(`{"uploadUrl":"https://example.com/u?sig=secret","fileId":9007199254740993,"ratio":1.7500}`)

	sanitized := string(SanitizeJSON(body))

	if !strings.Contains(sanitized, "9007199254740993") {
		t.Errorf("large integer should survive verbatim: %s", sanitized)
	}
	if !strings.Contains(sanitized, "1.7500") {
		t.Errorf("decimal should survive verbatim: %s", sanitized)
	}
}

func TestSanitizeJSON_PreservesOtherBodies(t *testing.T) {
	body := []byte(`{"id":"123","name":"test"}`)

	sanitized := SanitizeJSON(body)

	if string(sanitized) != string(body) {
		t.Fatalf("non-upload payload should be unchanged: got %s want %s", sanitized, body)
	}
}
