package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
		path, err := config.DefaultPathOrError()
		if err != nil {
			return err
		}
		pc, err := config.LoadProfileConfig(path)
		if err != nil {
			return err
		}

		cfg, _ := resolveOrCreateProfile(pc)

		key := args[0]
		if !config.IsValidKey(key) {
			return fmt.Errorf("유효하지 않은 설정 키: %s", key)
		}

		var value string
		useStdin, _ := cmd.Flags().GetBool("stdin")
		if useStdin {
			data, err := readStdinLimited(os.Stdin, maxStdinSize)
			if err != nil {
				return fmt.Errorf("stdin 읽기 실패: %w", err)
			}
			value = strings.TrimRight(string(data), "\r\n")
		} else if len(args) == 2 {
			value = args[1]
		} else {
			return fmt.Errorf("값을 지정하세요: naverworks config set %s <value>", key)
		}

		if err := cfg.Set(key, value); err != nil {
			return err
		}
		return pc.Save(path)
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
		cfg, _, err := loadActiveConfig()
		if err != nil {
			return err
		}
		fmt.Println(cfg.GetMasked(args[0]))
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "전체 설정 목록",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _, err := loadActiveConfig()
		if err != nil {
			return err
		}

		masked := make(map[string]string, len(config.AllKeys))
		for _, key := range config.AllKeys {
			masked[key] = cfg.GetMasked(key)
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
