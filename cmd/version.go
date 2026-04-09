package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "버전 정보 출력",
	Run: func(cmd *cobra.Command, args []string) {
		info := map[string]string{
			"version":    version,
			"commit":     commit,
			"build_date": buildDate,
		}
		data, err := json.Marshal(info)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), `{"error":{"code":"MARSHAL_ERROR","description":"%s"}}`+"\n", err.Error())
			return
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
