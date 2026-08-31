package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestSearchServiceRequests(t *testing.T) {
	const (
		groupID = "externalKey:group/one"
		userID  = "externalKey:user/one"
	)

	tests := []struct {
		name        string
		escapedPath string
		rawQuery    string
		call        func(*Client) (*Response, error)
	}{
		{
			name:        "board posts",
			escapedPath: "/boards/posts/search",
			rawQuery:    "boardIds=1%2C2&count=40&cursor=next%2Bcursor&endTime=2026-04-28&hasAttachment=true&query=weekly+%26+notes&startTime=2026-01-01&writerId=user%2Fone",
			call: func(client *Client) (*Response, error) {
				return NewBoardService(client).SearchPosts("next+cursor", 40, BoardPostSearchOptions{
					Query:         "weekly & notes",
					BoardIDs:      "1,2",
					HasAttachment: true,
					WriterID:      "user/one",
					StartTime:     "2026-01-01",
					EndTime:       "2026-04-28",
				})
			},
		},
		{
			name:        "note posts",
			escapedPath: "/groups/" + url.PathEscape(groupID) + "/note/posts/search",
			rawQuery:    "count=40&cursor=next%2Bcursor&query=weekly+%26+notes",
			call: func(client *Client) (*Response, error) {
				return NewNoteService(client).SearchPosts(groupID, "weekly & notes", "next+cursor", 40)
			},
		},
		{
			name:        "calendar events",
			escapedPath: "/users/" + url.PathEscape(userID) + "/calendars/events/search",
			rawQuery:    "count=100&cursor=next%2Bcursor&endTime=2026-06-03T11%3A00%3A00%2B09%3A00&query=weekly+%26+meeting&queryFilters=summary%2Cattendee&startTime=2026-06-03T10%3A00%3A00%2B09%3A00",
			call: func(client *Client) (*Response, error) {
				return NewCalendarService(client).SearchEvents(userID, "next+cursor", 100, CalendarEventSearchOptions{
					Query:        "weekly & meeting",
					QueryFilters: "summary,attendee",
					StartTime:    "2026-06-03T10:00:00+09:00",
					EndTime:      "2026-06-03T11:00:00+09:00",
				})
			},
		},
		{
			name:        "drive files",
			escapedPath: "/users/" + url.PathEscape(userID) + "/drive/search",
			rawQuery:    "count=200&cursor=next%2Bcursor&driveTypeFilters=MY_DRIVE%2CCHANNEL_FOLDER&endTime=2026-06-03T11%3A00%3A00%2B09%3A00&fileTypes=DOC%2CFOLDER&orderBy=modifiedTime+desc&parentFileId=folder%2Fone&query=weekly+%26+notes&queryFilters=fileName%2Ccontent&startTime=2026-06-03T10%3A00%3A00%2B09%3A00",
			call: func(client *Client) (*Response, error) {
				return NewDriveService(client).SearchFiles(userID, "next+cursor", 200, DriveSearchOptions{
					Query:            "weekly & notes",
					QueryFilters:     "fileName,content",
					ParentFileID:     "folder/one",
					FileTypes:        "DOC,FOLDER",
					StartTime:        "2026-06-03T10:00:00+09:00",
					EndTime:          "2026-06-03T11:00:00+09:00",
					DriveTypeFilters: "MY_DRIVE,CHANNEL_FOLDER",
					OrderBy:          "modifiedTime desc",
				})
			},
		},
		{
			name:        "tasks",
			escapedPath: "/users/" + url.PathEscape(userID) + "/tasks/search",
			rawQuery:    "assigneeId=assignee%2Fone&assignorId=assignor%2Fone&count=100&cursor=next%2Bcursor&endTime=2026-06-03T11%3A00%3A00%2B09%3A00&hasAttachment=true&hasDueDate=false&orderBy=createdTime+desc&query=weekly+%26+meeting&startTime=2026-06-03T10%3A00%3A00%2B09%3A00&status=TODO",
			call: func(client *Client) (*Response, error) {
				return NewTaskService(client).SearchTasks(userID, "next+cursor", 100, TaskSearchOptions{
					Query:         "weekly & meeting",
					AssignorID:    "assignor/one",
					AssigneeID:    "assignee/one",
					StartTime:     "2026-06-03T10:00:00+09:00",
					EndTime:       "2026-06-03T11:00:00+09:00",
					Status:        "TODO",
					HasDueDate:    "false",
					HasAttachment: "true",
					OrderBy:       "createdTime desc",
				})
			},
		},
		{
			name:        "contacts",
			escapedPath: "/users/" + url.PathEscape(userID) + "/contacts/search",
			rawQuery:    "count=100&cursor=next%2Bcursor&orderBy=name+desc&query=kim+%26+lee&queryFilters=contactName%2Cemails",
			call: func(client *Client) (*Response, error) {
				return NewContactService(client).SearchContacts(userID, "next+cursor", 100, ContactSearchOptions{
					Query:        "kim & lee",
					QueryFilters: "contactName,emails",
					OrderBy:      "name desc",
				})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodGet {
					t.Errorf("method = %q, want GET", request.Method)
				}
				if request.URL.EscapedPath() != test.escapedPath {
					t.Errorf("escaped path = %q, want %q", request.URL.EscapedPath(), test.escapedPath)
				}
				if request.URL.RawQuery != test.rawQuery {
					t.Errorf("raw query = %q, want %q", request.URL.RawQuery, test.rawQuery)
				}
				_, _ = w.Write([]byte(`{"items":[]}`))
			}))
			defer server.Close()

			token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(time.Hour)}
			response, err := test.call(NewClient(server.URL, token, nil))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if response.StatusCode != http.StatusOK {
				t.Errorf("status = %d, want %d", response.StatusCode, http.StatusOK)
			}
		})
	}
}

func TestDriveServiceSearchFilesOmitsQueryFiltersWhenUnset(t *testing.T) {
	var rawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		rawQuery = request.URL.RawQuery
		_, _ = w.Write([]byte(`{"files":[]}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(time.Hour)}
	_, err := NewDriveService(NewClient(server.URL, token, nil)).SearchFiles("user-1", "", 0, DriveSearchOptions{Query: "weekly"})
	if err != nil {
		t.Fatalf("SearchFiles failed: %v", err)
	}
	if rawQuery != "query=weekly" {
		t.Fatalf("raw query = %q, want query only", rawQuery)
	}
}
