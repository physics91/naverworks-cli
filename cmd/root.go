package cmd

import (
	"github.com/spf13/cobra"
)

var (
	outputFormat  string
	profileName   string
	dryRun        bool
	planOutPath   string
	generateInput bool
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
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "API 요청을 실제로 보내지 않고 실행 계획만 출력")
	rootCmd.PersistentFlags().StringVar(&planOutPath, "plan-out", "", "dry-run 실행 계획을 저장할 파일 경로")
	rootCmd.PersistentFlags().BoolVar(&generateInput, "generate-input", false, "API 요청의 최종 JSON 입력만 출력")
}

func Execute() error {
	return rootCmd.Execute()
}
