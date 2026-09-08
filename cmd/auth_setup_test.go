package cmd

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/config"
)

func TestShouldReselectBot(t *testing.T) {
	tests := []struct {
		name        string
		existingBot string
		answer      string
		want        bool
	}{
		{name: "missing bot id", existingBot: "", answer: "", want: true},
		{name: "decline with enter", existingBot: "bot-1", answer: "", want: false},
		{name: "decline with n", existingBot: "bot-1", answer: "n", want: false},
		{name: "accept with y", existingBot: "bot-1", answer: "y", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldReselectBot(tt.existingBot, tt.answer); got != tt.want {
				t.Fatalf("shouldReselectBot(%q, %q) = %v, want %v", tt.existingBot, tt.answer, got, tt.want)
			}
		})
	}
}

func TestNeedsBotConfigSave(t *testing.T) {
	tests := []struct {
		name   string
		before string
		after  string
		want   bool
	}{
		{name: "unchanged empty", before: "", after: "", want: false},
		{name: "unchanged existing", before: "bot-1", after: "bot-1", want: false},
		{name: "new value", before: "", after: "bot-1", want: true},
		{name: "changed value", before: "bot-1", after: "bot-2", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsBotConfigSave(tt.before, tt.after); got != tt.want {
				t.Fatalf("needsBotConfigSave(%q, %q) = %v, want %v", tt.before, tt.after, got, tt.want)
			}
		})
	}
}

func TestParseSetupBots(t *testing.T) {
	t.Run("uses bot id when name missing", func(t *testing.T) {
		body := []byte(`{"bots":[{"botId":"bot-1","botName":""}]}`)
		bots, err := parseSetupBots(body)
		if err != nil {
			t.Fatal(err)
		}
		if len(bots) != 1 {
			t.Fatalf("len(bots) = %d, want 1", len(bots))
		}
		if bots[0].Label != "bot-1" {
			t.Fatalf("label = %q, want %q", bots[0].Label, "bot-1")
		}
	})

	t.Run("returns bot name when present", func(t *testing.T) {
		body := []byte(`{"bots":[{"botId":"bot-1","botName":"Main Bot"}]}`)
		bots, err := parseSetupBots(body)
		if err != nil {
			t.Fatal(err)
		}
		if bots[0].Label != "Main Bot" {
			t.Fatalf("label = %q, want %q", bots[0].Label, "Main Bot")
		}
	})
}

func TestChooseSetupBot(t *testing.T) {
	t.Run("auto selects single bot", func(t *testing.T) {
		var out bytes.Buffer
		got, changed, err := chooseSetupBot(bufio.NewReader(strings.NewReader("")), &out, []setupBotOption{
			{BotID: "bot-1", Label: "Main Bot"},
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "bot-1" || !changed {
			t.Fatalf("got (%q, %v), want (%q, true)", got, changed, "bot-1")
		}
	})

	t.Run("chooses multiple bots by number", func(t *testing.T) {
		var out bytes.Buffer
		got, changed, err := chooseSetupBot(bufio.NewReader(strings.NewReader("2\n")), &out, []setupBotOption{
			{BotID: "bot-1", Label: "First"},
			{BotID: "bot-2", Label: "Second"},
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "bot-2" || !changed {
			t.Fatalf("got (%q, %v), want (%q, true)", got, changed, "bot-2")
		}
	})

	t.Run("retries invalid number", func(t *testing.T) {
		var out bytes.Buffer
		got, changed, err := chooseSetupBot(bufio.NewReader(strings.NewReader("9\n1\n")), &out, []setupBotOption{
			{BotID: "bot-1", Label: "First"},
			{BotID: "bot-2", Label: "Second"},
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "bot-1" || !changed {
			t.Fatalf("got (%q, %v), want (%q, true)", got, changed, "bot-1")
		}
		if !strings.Contains(out.String(), "다시 입력") {
			t.Fatalf("output = %q, want retry hint", out.String())
		}
	})

	t.Run("manual input path", func(t *testing.T) {
		var out bytes.Buffer
		got, changed, err := chooseSetupBot(bufio.NewReader(strings.NewReader("m\nmanual-bot\n")), &out, []setupBotOption{
			{BotID: "bot-1", Label: "First"},
			{BotID: "bot-2", Label: "Second"},
		}, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "manual-bot" || !changed {
			t.Fatalf("got (%q, %v), want (%q, true)", got, changed, "manual-bot")
		}
	})

	t.Run("empty input skips", func(t *testing.T) {
		var out bytes.Buffer
		got, changed, err := chooseSetupBot(bufio.NewReader(strings.NewReader("\n")), &out, []setupBotOption{
			{BotID: "bot-1", Label: "First"},
			{BotID: "bot-2", Label: "Second"},
		}, "keep-bot")
		if err != nil {
			t.Fatal(err)
		}
		if got != "keep-bot" || changed {
			t.Fatalf("got (%q, %v), want (%q, false)", got, changed, "keep-bot")
		}
	})
}

func TestFetchSetupBots(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bots, err := fetchSetupBots(func() (*api.Response, error) {
			return &api.Response{Body: []byte(`{"bots":[{"botId":"bot-1","botName":"Main Bot"}]}`)}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(bots) != 1 || bots[0].BotID != "bot-1" {
			t.Fatalf("bots = %#v, want one bot-1 entry", bots)
		}
	})

	t.Run("parse error", func(t *testing.T) {
		_, err := fetchSetupBots(func() (*api.Response, error) {
			return &api.Response{Body: []byte(`not-json`)}, nil
		})
		if err == nil {
			t.Fatal("expected parse error")
		}
	})
}

func TestRunPostLoginBotSelection(t *testing.T) {
	t.Run("keeps existing bot when user declines reselection", func(t *testing.T) {
		cfg := &config.Config{BotID: "keep-bot"}
		var out bytes.Buffer
		fetchCalled := false
		saved := false

		err := runPostLoginBotSelection(
			bufio.NewReader(strings.NewReader("n\n")),
			&out,
			cfg,
			func() ([]setupBotOption, error) {
				fetchCalled = true
				return nil, nil
			},
			func() error {
				saved = true
				return nil
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if fetchCalled {
			t.Fatal("fetch should not be called when reselection is declined")
		}
		if saved {
			t.Fatal("config save should not run when bot id is unchanged")
		}
		if cfg.BotID != "keep-bot" {
			t.Fatalf("cfg.BotID = %q, want %q", cfg.BotID, "keep-bot")
		}
	})

	t.Run("auto selects and saves", func(t *testing.T) {
		cfg := &config.Config{}
		var out bytes.Buffer
		saved := false

		err := runPostLoginBotSelection(
			bufio.NewReader(strings.NewReader("")),
			&out,
			cfg,
			func() ([]setupBotOption, error) {
				return []setupBotOption{{BotID: "bot-1", Label: "Main Bot"}}, nil
			},
			func() error {
				saved = true
				return nil
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if !saved {
			t.Fatal("config save should run when bot id changes")
		}
		if cfg.BotID != "bot-1" {
			t.Fatalf("cfg.BotID = %q, want %q", cfg.BotID, "bot-1")
		}
	})

	t.Run("lookup failure falls back to manual input", func(t *testing.T) {
		cfg := &config.Config{}
		var out bytes.Buffer
		saved := false

		err := runPostLoginBotSelection(
			bufio.NewReader(strings.NewReader("manual-bot\n")),
			&out,
			cfg,
			func() ([]setupBotOption, error) {
				return nil, assertiveError("boom")
			},
			func() error {
				saved = true
				return nil
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if !saved {
			t.Fatal("config save should run after manual fallback input")
		}
		if cfg.BotID != "manual-bot" {
			t.Fatalf("cfg.BotID = %q, want %q", cfg.BotID, "manual-bot")
		}
		if !strings.Contains(out.String(), "bot scope") {
			t.Fatalf("output = %q, want bot scope hint", out.String())
		}
	})

	t.Run("empty fallback skip keeps existing bot", func(t *testing.T) {
		cfg := &config.Config{BotID: "keep-bot"}
		var out bytes.Buffer
		saved := false

		err := runPostLoginBotSelection(
			bufio.NewReader(strings.NewReader("y\n\n")),
			&out,
			cfg,
			func() ([]setupBotOption, error) {
				return nil, nil
			},
			func() error {
				saved = true
				return nil
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		if saved {
			t.Fatal("config save should not run when fallback skip keeps the current bot")
		}
		if cfg.BotID != "keep-bot" {
			t.Fatalf("cfg.BotID = %q, want %q", cfg.BotID, "keep-bot")
		}
	})
}

func TestLoadProfileConfigForSetup_ReturnsMalformedConfigError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{bad-json`), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := loadProfileConfigForSetup(path)
	if err == nil {
		t.Fatal("expected malformed config error")
	}
	if !strings.Contains(err.Error(), "config 파싱 실패") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOAuthCallbackPortRange(t *testing.T) {
	start, end := oauthCallbackPortRange(0)
	if start != 8484 || end != 8494 {
		t.Fatalf("default range = %d-%d, want 8484-8494", start, end)
	}

	start, end = oauthCallbackPortRange(8484)
	if start != 8484 || end != 8484 {
		t.Fatalf("fixed port range = %d-%d, want 8484-8484", start, end)
	}
}

func TestAuthSetupLong_DescribesOAuthPromptSet(t *testing.T) {
	long := authSetupCmd.Long
	if !strings.Contains(long, "Client ID") || !strings.Contains(long, "Client Secret") {
		t.Fatalf("setup Long missing OAuth credentials: %q", long)
	}
	if !strings.Contains(long, "Scope·Calendar User ID·로그인 전 Bot ID는 묻지 않습니다") {
		t.Fatalf("setup Long missing OAuth omitted prompts: %q", long)
	}
	if !strings.Contains(long, "서비스 계정 ID") || !strings.Contains(long, "개인키") {
		t.Fatalf("setup Long missing JWT required prompts: %q", long)
	}
}

func TestNewSetupPromptSet_OAuthOmitsPreLoginOptionals(t *testing.T) {
	oauth := newSetupPromptSet("oauth")
	if !oauth.ClientID || !oauth.ClientSecret || !oauth.LoginNow {
		t.Fatalf("oauth prompts = %+v, want client id/secret and login", oauth)
	}
	if oauth.ServiceAccountID || oauth.PrivateKeyPath || oauth.BotID || oauth.Scope || oauth.CalendarUserID {
		t.Fatalf("oauth prompts = %+v, want no JWT or pre-login bot/scope/calendar", oauth)
	}

	jwt := newSetupPromptSet("jwt")
	if !jwt.ClientID || !jwt.ClientSecret || !jwt.LoginNow {
		t.Fatalf("jwt prompts = %+v, want client id/secret and login", jwt)
	}
	if !jwt.ServiceAccountID || !jwt.PrivateKeyPath || !jwt.BotID || !jwt.Scope || !jwt.CalendarUserID {
		t.Fatalf("jwt prompts = %+v, want service account, key, bot, scope, calendar", jwt)
	}
}

func TestCollectJWTSetupFields_OAuthDoesNotReadOptionalPrompts(t *testing.T) {
	cfg := &config.Config{BotID: "keep-bot", Scope: "keep-scope", DefaultCalendarUserID: "keep-cal"}
	reader := bufio.NewReader(strings.NewReader("should-not-consume\n"))

	if err := collectJWTSetupFields(reader, cfg, newSetupPromptSet("oauth")); err != nil {
		t.Fatal(err)
	}
	if cfg.BotID != "keep-bot" || cfg.Scope != "keep-scope" || cfg.DefaultCalendarUserID != "keep-cal" {
		t.Fatalf("oauth collect mutated optional fields: %+v", cfg)
	}
	rest, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if rest != "should-not-consume\n" {
		t.Fatalf("oauth collect consumed stdin: %q", rest)
	}
}

func TestCollectJWTSetupFields_JWTReadsServiceAccountAndOptionals(t *testing.T) {
	cfg := &config.Config{}
	reader := bufio.NewReader(strings.NewReader("sa@example.com\n/tmp/key.pem\nbot-1\nbot directory\nme\n"))

	if err := collectJWTSetupFields(reader, cfg, newSetupPromptSet("jwt")); err != nil {
		t.Fatal(err)
	}
	if cfg.ServiceAccountID != "sa@example.com" {
		t.Fatalf("service_account_id = %q", cfg.ServiceAccountID)
	}
	if cfg.PrivateKeyPath != "/tmp/key.pem" {
		t.Fatalf("private_key_path = %q", cfg.PrivateKeyPath)
	}
	if cfg.BotID != "bot-1" {
		t.Fatalf("bot_id = %q", cfg.BotID)
	}
	if cfg.Scope != "bot directory" {
		t.Fatalf("scope = %q", cfg.Scope)
	}
	if cfg.DefaultCalendarUserID != "me" {
		t.Fatalf("calendar user = %q", cfg.DefaultCalendarUserID)
	}
}

func TestApplySetupAuthMethod_ClearsJWTFieldsForOAuth(t *testing.T) {
	cfg := &config.Config{
		ServiceAccountID: "svc@example.com",
		PrivateKeyPath:   "/tmp/private.pem",
	}

	applySetupAuthMethod(cfg, "oauth")

	if cfg.ServiceAccountID != "" {
		t.Fatalf("service_account_id = %q, want empty", cfg.ServiceAccountID)
	}
	if cfg.PrivateKeyPath != "" {
		t.Fatalf("private_key_path = %q, want empty", cfg.PrivateKeyPath)
	}
}

type assertiveError string

func (e assertiveError) Error() string {
	return string(e)
}
