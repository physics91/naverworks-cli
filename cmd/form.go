package cmd

import (
	"encoding/json"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var formCmd = &cobra.Command{
	Use:   "form",
	Short: "설문 관리",
}

var formListResponsesCmd = &cobra.Command{
	Use:   "list-responses <formId>",
	Short: "설문 응답 목록 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewFormService(client)
		return runListCmd(cmd, []string{"responseId"}, "responses", func(c string, n int) (*api.Response, error) {
			return svc.ListResponses(args[0], c, n)
		})
	},
}

var formDownloadAttachmentCmd = &cobra.Command{
	Use:   "download-attachment <formId> <responseId> <attachmentId>",
	Short: "설문 응답 첨부파일 다운로드",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewFormService(client)

		downloadURL, err := svc.DownloadAttachment(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		result, _ := json.Marshal(map[string]string{"download_url": downloadURL})
		printBody(result)
		return nil
	},
}

func init() {
	formListResponsesCmd.Flags().String("cursor", "", "페이지네이션 커서")
	formListResponsesCmd.Flags().Int("count", 0, "페이지 크기")
	formListResponsesCmd.Flags().Bool("all", false, "전체 페이지 자동 순회")

	formCmd.AddCommand(formListResponsesCmd, formDownloadAttachmentCmd)
	rootCmd.AddCommand(formCmd)
}
