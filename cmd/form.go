package cmd

import (
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
		svc, err := newSvc(api.NewFormService)
		if err != nil {
			return err
		}
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
		svc, err := newSvc(api.NewFormService)
		if err != nil {
			return err
		}
		downloadURL, err := svc.DownloadAttachment(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		printDownloadURL(downloadURL)
		return nil
	},
}

func init() {
	addListFlags(formListResponsesCmd)

	formCmd.AddCommand(formListResponsesCmd, formDownloadAttachmentCmd)
	rootCmd.AddCommand(formCmd)
}
