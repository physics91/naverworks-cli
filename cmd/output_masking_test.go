//go:build !windows

package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
)

// paginateAndPrint and runListCmd write through formatter.PrintRaw directly
// rather than through printBody, so they used to bypass sanitization entirely.
// This pins the behaviour of the aggregated pagination output.
func TestPaginateAndPrint_MasksPresignedURLInAggregatedPages(t *testing.T) {
	pages := []string{
		`{"files":[{"id":"1","uploadUrl":"https://s.example/p?sig=LEAKED"}],"responseMetaData":{"nextCursor":"c2"}}`,
		`{"files":[{"id":"2"}]}`,
	}
	call := 0

	var buf bytes.Buffer
	formatter := output.NewFormatter("json", &buf)

	err := paginateAndPrint(func(cursor string) (*api.Response, error) {
		if call >= len(pages) {
			return nil, fmt.Errorf("unexpected extra page request")
		}
		body := pages[call]
		call++
		return &api.Response{StatusCode: 200, Body: []byte(body)}, nil
	}, "files", formatter)
	if err != nil {
		t.Fatalf("paginateAndPrint failed: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "LEAKED") {
		t.Fatalf("presigned URL reached stdout via pagination: %s", out)
	}
	if !strings.Contains(out, `"id": "1"`) || !strings.Contains(out, `"id": "2"`) {
		t.Fatalf("expected both pages to be merged: %s", out)
	}
}
