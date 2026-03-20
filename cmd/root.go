package cmd

import (
	"github.com/spf13/cobra"
)

var outputFormat string

var rootCmd = &cobra.Command{
	Use:           "nw-cli",
	Short:         "네이버웍스 CLI",
	Long:          "네이버웍스 REST API v1.0 명령줄 도구",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "json", "출력 형식 (json|table)")
}

func Execute() error {
	return rootCmd.Execute()
}
