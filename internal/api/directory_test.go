package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestDirectoryService_ListUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"users":[]}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	dir := NewDirectoryService(client)

	resp, err := dir.ListUsers("", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDirectoryService_GetUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/user1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"userId":"user1"}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	dir := NewDirectoryService(client)

	resp, err := dir.GetUser("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDirectoryService_ListGroups(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/groups" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"groups":[]}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	dir := NewDirectoryService(client)

	resp, err := dir.ListGroups("", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestDirectoryService_SearchAndUserMembershipRequests(t *testing.T) {
	const userID = "externalKey:team/one"

	tests := []struct {
		name         string
		escapedPath  string
		rawQuery     string
		responseBody string
		call         func(*DirectoryService) (*Response, error)
	}{
		{
			name:         "search users",
			escapedPath:  "/users/search",
			rawQuery:     "count=20&cursor=next%2Bcursor&domainId=10000001&orderBy=userName+desc&query=kim+%26+lee",
			responseBody: `{"users":[]}`,
			call: func(service *DirectoryService) (*Response, error) {
				return service.SearchUsers("next+cursor", 20, DirectorySearchOptions{Query: "kim & lee", OrderBy: "userName desc", DomainID: 10000001})
			},
		},
		{
			name:         "search groups",
			escapedPath:  "/groups/search",
			rawQuery:     "count=20&cursor=next%2Bcursor&domainId=10000001&orderBy=groupName+desc&query=kim+%26+lee",
			responseBody: `{"groups":[]}`,
			call: func(service *DirectoryService) (*Response, error) {
				return service.SearchGroups("next+cursor", 20, DirectorySearchOptions{Query: "kim & lee", OrderBy: "groupName desc", DomainID: 10000001})
			},
		},
		{
			name:         "search org units",
			escapedPath:  "/orgunits/search",
			rawQuery:     "count=20&cursor=next%2Bcursor&domainId=10000001&orderBy=orgUnitName+desc&query=kim+%26+lee",
			responseBody: `{"orgUnits":[]}`,
			call: func(service *DirectoryService) (*Response, error) {
				return service.SearchOrgUnits("next+cursor", 20, DirectorySearchOptions{Query: "kim & lee", OrderBy: "orgUnitName desc", DomainID: 10000001})
			},
		},
		{
			name:         "list user groups",
			escapedPath:  "/users/" + url.PathEscape(userID) + "/groups",
			rawQuery:     "count=20&cursor=next%2Bcursor&membershipType=DIRECT",
			responseBody: `{"groups":[]}`,
			call: func(service *DirectoryService) (*Response, error) {
				return service.ListUserGroups(userID, "DIRECT", "next+cursor", 20)
			},
		},
		{
			name:         "list user org units",
			escapedPath:  "/users/" + url.PathEscape(userID) + "/orgunits",
			rawQuery:     "count=20&cursor=next%2Bcursor&membershipType=ALL",
			responseBody: `{"orgUnits":[]}`,
			call: func(service *DirectoryService) (*Response, error) {
				return service.ListUserOrgUnits(userID, "ALL", "next+cursor", 20)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("method = %q, want GET", r.Method)
				}
				if r.URL.EscapedPath() != test.escapedPath {
					t.Errorf("escaped path = %q, want %q", r.URL.EscapedPath(), test.escapedPath)
				}
				if r.URL.RawQuery != test.rawQuery {
					t.Errorf("raw query = %q, want %q", r.URL.RawQuery, test.rawQuery)
				}
				_, _ = w.Write([]byte(test.responseBody))
			}))
			defer server.Close()

			token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(time.Hour)}
			response, err := test.call(NewDirectoryService(NewClient(server.URL, token, nil)))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if response.StatusCode != http.StatusOK {
				t.Errorf("status = %d, want %d", response.StatusCode, http.StatusOK)
			}
		})
	}
}
