package output

import (
	"bytes"
	"encoding/json"
	"strings"
)

// sensitiveKeys are response fields that carry a presigned signature. Printing
// them would leak a credential that grants direct write access to storage.
var sensitiveKeys = map[string]bool{
	"uploadUrl":  true,
	"upload_url": true,
}

// SanitizeJSON removes presigned upload URLs from a JSON payload, wherever they
// appear in the tree. It is applied at the single output gate (Formatter.PrintRaw)
// rather than at individual call sites, so no command can bypass it by printing
// a response directly.
//
// The input is returned unchanged when there is nothing to remove, so ordinary
// responses are never reformatted.
func SanitizeJSON(body []byte) []byte {
	if len(body) == 0 || strings.TrimSpace(string(body)) == "" {
		return body
	}

	// UseNumber keeps large integers and precise decimals byte-identical if the
	// payload has to be re-serialized.
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	var payload interface{}
	if err := decoder.Decode(&payload); err != nil {
		return redactUnparsableBody(body)
	}

	cleaned, removed := stripSensitiveKeys(payload)

	// A well-formed response holds exactly one JSON value. Trailing data means
	// the payload was not fully inspected, so emit only the value that was
	// actually sanitized instead of echoing the unchecked remainder.
	hasTrailingData := decoder.More()

	if !removed && !hasTrailingData {
		return body
	}

	if obj, ok := cleaned.(map[string]interface{}); ok && removed {
		if _, exists := obj["uploaded"]; !exists {
			obj["uploaded"] = true
		}
	}

	sanitized, err := json.Marshal(cleaned)
	if err != nil {
		return body
	}
	return sanitized
}

// redactUnparsableBody handles a payload that is not valid JSON. PrintRaw falls
// back to printing such a body verbatim, so it cannot be walked as a tree and
// cleaned. If it mentions a presigned field at all, the whole body is withheld
// rather than echoed.
func redactUnparsableBody(body []byte) []byte {
	lowered := strings.ToLower(string(body))
	for key := range sensitiveKeys {
		if strings.Contains(lowered, strings.ToLower(key)) {
			return []byte(`{"error":"응답을 파싱할 수 없어 민감 정보 포함 가능성으로 출력을 생략했습니다"}`)
		}
	}
	return body
}

// stripSensitiveKeys walks the decoded JSON tree and drops presigned upload
// URLs at any depth, including inside nested objects and arrays.
func stripSensitiveKeys(node interface{}) (interface{}, bool) {
	switch typed := node.(type) {
	case map[string]interface{}:
		removed := false
		for key, value := range typed {
			if sensitiveKeys[key] {
				delete(typed, key)
				removed = true
				continue
			}
			cleaned, childRemoved := stripSensitiveKeys(value)
			typed[key] = cleaned
			removed = removed || childRemoved
		}
		return typed, removed
	case []interface{}:
		removed := false
		for i, value := range typed {
			cleaned, childRemoved := stripSensitiveKeys(value)
			typed[i] = cleaned
			removed = removed || childRemoved
		}
		return typed, removed
	default:
		return node, false
	}
}
