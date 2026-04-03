package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
)

func TestBotService_SendTextToUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bots/123/users/user1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(201)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	bot := NewBotService(client)

	resp, err := bot.SendTextToUser("123", "user1", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestBotService_SendTextToChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bots/123/channels/ch1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(201)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	bot := NewBotService(client)

	resp, err := bot.SendTextToChannel("123", "ch1", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestBotService_GetChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bots/123/channels/ch1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"channelId":"ch1"}`))
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
	client := NewClient(server.URL, token, nil)
	bot := NewBotService(client)

	resp, err := bot.GetChannel("123", "ch1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestBotService_CRUDMethods(t *testing.T) {
	tests := []struct {
		name       string
		call       func(svc *BotService)
		wantMethod string
		wantPath   string
	}{
		{
			name:       "CreateBot",
			call:       func(s *BotService) { s.CreateBot([]byte(`{"botName":"test"}`)) },
			wantMethod: "POST",
			wantPath:   "/bots",
		},
		{
			name:       "ListBots",
			call:       func(s *BotService) { s.ListBots("", 10) },
			wantMethod: "GET",
			wantPath:   "/bots",
		},
		{
			name:       "GetBot",
			call:       func(s *BotService) { s.GetBot("bot1") },
			wantMethod: "GET",
			wantPath:   "/bots/bot1",
		},
		{
			name:       "UpdateBot",
			call:       func(s *BotService) { s.UpdateBot("bot1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/bots/bot1",
		},
		{
			name:       "PatchBot",
			call:       func(s *BotService) { s.PatchBot("bot1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/bots/bot1",
		},
		{
			name:       "DeleteBot",
			call:       func(s *BotService) { s.DeleteBot("bot1") },
			wantMethod: "DELETE",
			wantPath:   "/bots/bot1",
		},
		{
			name:       "RegenerateSecret",
			call:       func(s *BotService) { s.RegenerateSecret("bot1") },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/secret",
		},
		{
			name:       "SendMessageToUser",
			call:       func(s *BotService) { s.SendMessageToUser("bot1", "u1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/users/u1/messages",
		},
		{
			name:       "SendMessageToChannel",
			call:       func(s *BotService) { s.SendMessageToChannel("bot1", "ch1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/channels/ch1/messages",
		},
		{
			name:       "CreateAttachment",
			call:       func(s *BotService) { s.CreateAttachment("bot1", map[string]interface{}{"fileName": "a.txt"}) },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/attachments",
		},
		{
			name:       "ListChannelMembers",
			call:       func(s *BotService) { s.ListChannelMembers("bot1", "ch1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/bots/bot1/channels/ch1/members",
		},
		{
			name:       "CreateChannel",
			call:       func(s *BotService) { s.CreateChannel("bot1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/channels",
		},
		{
			name:       "LeaveChannel",
			call:       func(s *BotService) { s.LeaveChannel("bot1", "ch1") },
			wantMethod: "DELETE",
			wantPath:   "/bots/bot1/channels/ch1",
		},
		{
			name:       "RegisterDomain",
			call:       func(s *BotService) { s.RegisterDomain("bot1", "d1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/domains/d1",
		},
		{
			name:       "ListDomains",
			call:       func(s *BotService) { s.ListDomains("bot1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/bots/bot1/domains",
		},
		{
			name:       "UpdateDomain",
			call:       func(s *BotService) { s.UpdateDomain("bot1", "d1", []byte(`{}`)) },
			wantMethod: "PUT",
			wantPath:   "/bots/bot1/domains/d1",
		},
		{
			name:       "PatchDomain",
			call:       func(s *BotService) { s.PatchDomain("bot1", "d1", []byte(`{}`)) },
			wantMethod: "PATCH",
			wantPath:   "/bots/bot1/domains/d1",
		},
		{
			name:       "DeleteDomain",
			call:       func(s *BotService) { s.DeleteDomain("bot1", "d1") },
			wantMethod: "DELETE",
			wantPath:   "/bots/bot1/domains/d1",
		},
		{
			name:       "AddDomainMembers",
			call:       func(s *BotService) { s.AddDomainMembers("bot1", "d1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/domains/d1/members",
		},
		{
			name:       "ListDomainMembers",
			call:       func(s *BotService) { s.ListDomainMembers("bot1", "d1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/bots/bot1/domains/d1/members",
		},
		{
			name:       "RemoveDomainMember",
			call:       func(s *BotService) { s.RemoveDomainMember("bot1", "d1", "u1") },
			wantMethod: "DELETE",
			wantPath:   "/bots/bot1/domains/d1/members/u1",
		},
		{
			name:       "UpsertPersistentMenu",
			call:       func(s *BotService) { s.UpsertPersistentMenu("bot1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/persistentmenu",
		},
		{
			name:       "GetPersistentMenu",
			call:       func(s *BotService) { s.GetPersistentMenu("bot1") },
			wantMethod: "GET",
			wantPath:   "/bots/bot1/persistentmenu",
		},
		{
			name:       "DeletePersistentMenu",
			call:       func(s *BotService) { s.DeletePersistentMenu("bot1") },
			wantMethod: "DELETE",
			wantPath:   "/bots/bot1/persistentmenu",
		},
		{
			name:       "CreateRichMenu",
			call:       func(s *BotService) { s.CreateRichMenu("bot1", []byte(`{}`)) },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/richmenus",
		},
		{
			name:       "ListRichMenus",
			call:       func(s *BotService) { s.ListRichMenus("bot1", "", 10) },
			wantMethod: "GET",
			wantPath:   "/bots/bot1/richmenus",
		},
		{
			name:       "GetRichMenu",
			call:       func(s *BotService) { s.GetRichMenu("bot1", "rm1") },
			wantMethod: "GET",
			wantPath:   "/bots/bot1/richmenus/rm1",
		},
		{
			name:       "DeleteRichMenu",
			call:       func(s *BotService) { s.DeleteRichMenu("bot1", "rm1") },
			wantMethod: "DELETE",
			wantPath:   "/bots/bot1/richmenus/rm1",
		},
		{
			name:       "SetUserRichMenu",
			call:       func(s *BotService) { s.SetUserRichMenu("bot1", "rm1", "u1") },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/richmenus/rm1/users/u1",
		},
		{
			name:       "GetUserRichMenu",
			call:       func(s *BotService) { s.GetUserRichMenu("bot1", "u1") },
			wantMethod: "GET",
			wantPath:   "/bots/bot1/richmenus/users/u1",
		},
		{
			name:       "DeleteUserRichMenu",
			call:       func(s *BotService) { s.DeleteUserRichMenu("bot1", "u1") },
			wantMethod: "DELETE",
			wantPath:   "/bots/bot1/richmenus/users/u1",
		},
		{
			name:       "SetDefaultRichMenu",
			call:       func(s *BotService) { s.SetDefaultRichMenu("bot1", "rm1") },
			wantMethod: "POST",
			wantPath:   "/bots/bot1/richmenus/rm1/set-default",
		},
		{
			name:       "GetDefaultRichMenu",
			call:       func(s *BotService) { s.GetDefaultRichMenu("bot1") },
			wantMethod: "GET",
			wantPath:   "/bots/bot1/richmenus/default",
		},
		{
			name:       "DeleteDefaultRichMenu",
			call:       func(s *BotService) { s.DeleteDefaultRichMenu("bot1") },
			wantMethod: "DELETE",
			wantPath:   "/bots/bot1/richmenus/default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				w.WriteHeader(200)
				w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			token := &auth.Token{AccessToken: "t", ExpiresAt: time.Now().Add(1 * time.Hour)}
			client := NewClient(srv.URL, token, nil)
			svc := NewBotService(client)
			tt.call(svc)

			if gotMethod != tt.wantMethod {
				t.Errorf("method: got %s, want %s", gotMethod, tt.wantMethod)
			}
			if gotPath != tt.wantPath {
				t.Errorf("path: got %s, want %s", gotPath, tt.wantPath)
			}
		})
	}
}
