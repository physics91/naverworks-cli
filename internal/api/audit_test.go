package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestMonitoringServiceDownloadMessagesChannelID(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		wantQuery string
	}{
		{
			name:      "specified",
			channelID: "channel/one",
			wantQuery: "channelId=channel%2Fone&endTime=2026-06-03T11%3A00%3A00%2B09%3A00&startTime=2026-06-03T10%3A00%3A00%2B09%3A00",
		},
		{
			name:      "omitted",
			wantQuery: "endTime=2026-06-03T11%3A00%3A00%2B09%3A00&startTime=2026-06-03T10%3A00%3A00%2B09%3A00",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var rawQuery string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				rawQuery = request.URL.RawQuery
				w.Header().Set("Location", "https://download.example.com/messages.csv")
				w.WriteHeader(http.StatusFound)
			}))
			defer server.Close()

			token := &auth.Token{AccessToken: "test-token", ExpiresAt: time.Now().Add(time.Hour)}
			got, err := NewMonitoringService(NewClient(server.URL, token, nil)).DownloadMessages(
				"2026-06-03T10:00:00+09:00",
				"2026-06-03T11:00:00+09:00",
				test.channelID,
			)
			if err != nil {
				t.Fatalf("DownloadMessages failed: %v", err)
			}
			if got != "https://download.example.com/messages.csv" {
				t.Fatalf("download URL = %q", got)
			}
			if rawQuery != test.wantQuery {
				t.Fatalf("raw query = %q, want %q", rawQuery, test.wantQuery)
			}
		})
	}
}
