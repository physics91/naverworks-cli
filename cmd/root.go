package cmd

import (
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	profileName  string
)

var rootCmd = &cobra.Command{
	Use:           "naverworks",
	Short:         "네이버웍스 CLI",
	Long:          "네이버웍스 REST API v1.0 명령줄 도구",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "json", "출력 형식 (json|table)")
	rootCmd.PersistentFlags().StringVar(&profileName, "profile", "", "설정/토큰 프로필명")
}

func Execute() error {
	return rootCmd.Execute()
}
