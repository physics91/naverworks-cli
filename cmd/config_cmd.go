package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "설정 관리",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> [value]",
	Short: "설정값 저장",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}

		key := args[0]
		if !config.IsValidKey(key) {
			return fmt.Errorf("유효하지 않은 설정 키: %s", key)
		}

		var value string
		useStdin, _ := cmd.Flags().GetBool("stdin")
		if useStdin {
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				value = scanner.Text()
			}
		} else if len(args) == 2 {
			value = args[1]
		} else {
			return fmt.Errorf("값을 지정하세요: nw-cli config set %s <value>", key)
		}

		if err := cfg.Set(key, value); err != nil {
			return err
		}
		return cfg.Save(path)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "설정값 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !config.IsValidKey(args[0]) {
			return fmt.Errorf("유효하지 않은 설정 키: %s", args[0])
		}
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		cfg.ApplyEnvOverrides()
		fmt.Println(cfg.GetMasked(args[0]))
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "전체 설정 목록",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		cfg.ApplyEnvOverrides()

		masked := map[string]string{
			"client_id":                cfg.GetMasked("client_id"),
			"client_secret":            cfg.GetMasked("client_secret"),
			"service_account_id":       cfg.GetMasked("service_account_id"),
			"private_key_path":         cfg.GetMasked("private_key_path"),
			"domain_id":               cfg.GetMasked("domain_id"),
			"bot_id":                  cfg.GetMasked("bot_id"),
			"scope":                   cfg.GetMasked("scope"),
			"default_calendar_user_id": cfg.GetMasked("default_calendar_user_id"),
		}
		data, _ := json.MarshalIndent(masked, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

func init() {
	configSetCmd.Flags().Bool("stdin", false, "stdin에서 값 읽기")
	configCmd.AddCommand(configSetCmd, configGetCmd, configListCmd)
	rootCmd.AddCommand(configCmd)
}
