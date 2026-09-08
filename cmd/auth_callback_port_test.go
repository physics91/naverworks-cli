package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCallbackPortFromCmd(t *testing.T) {
	newCmd := func() *cobra.Command {
		c := &cobra.Command{Use: "login"}
		addOAuthCallbackPortFlag(c)
		return c
	}

	t.Run("default zero is allowed", func(t *testing.T) {
		got, err := callbackPortFromCmd(newCmd())
		if err != nil {
			t.Fatal(err)
		}
		if got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})

	t.Run("accepts fixed port", func(t *testing.T) {
		c := newCmd()
		if err := c.Flags().Set("callback-port", "8484"); err != nil {
			t.Fatal(err)
		}
		got, err := callbackPortFromCmd(c)
		if err != nil {
			t.Fatal(err)
		}
		if got != 8484 {
			t.Fatalf("got %d, want 8484", got)
		}
	})

	t.Run("rejects negative", func(t *testing.T) {
		c := newCmd()
		if err := c.Flags().Set("callback-port", "-1"); err != nil {
			t.Fatal(err)
		}
		_, err := callbackPortFromCmd(c)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "유효하지 않은 callback-port") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects out of range", func(t *testing.T) {
		c := newCmd()
		if err := c.Flags().Set("callback-port", "70000"); err != nil {
			t.Fatal(err)
		}
		_, err := callbackPortFromCmd(c)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "유효하지 않은 callback-port") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestOAuthCallbackPortFlag_SharedAcrossCommands(t *testing.T) {
	cmds := []*cobra.Command{authSetupCmd, authLoginCmd, authDoctorCmd}
	for _, cmd := range cmds {
		flag := cmd.Flags().Lookup("callback-port")
		if flag == nil {
			t.Fatalf("%s missing --callback-port", cmd.Name())
		}
		if flag.Usage != oauthCallbackPortFlagUsage {
			t.Fatalf("%s callback-port usage = %q, want %q", cmd.Name(), flag.Usage, oauthCallbackPortFlagUsage)
		}
		if !strings.Contains(flag.Usage, "8484-8494") || !strings.Contains(flag.Usage, "양수") {
			t.Fatalf("%s callback-port usage missing port rule: %q", cmd.Name(), flag.Usage)
		}
	}
}
