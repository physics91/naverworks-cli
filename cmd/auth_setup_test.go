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
