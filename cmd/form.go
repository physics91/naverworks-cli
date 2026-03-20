package cmd

import (
	"encoding/json"
	"os"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
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
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewFormService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"responseId"}, "responses")

		if all {
			var allItems []json.RawMessage
			for {
				resp, err := svc.ListResponses(args[0], cursor, count)
				if err != nil {
					return err
				}
				var page struct {
					Responses        []json.RawMessage `json:"responses"`
					ResponseMetaData struct {
						NextCursor string `json:"nextCursor"`
					} `json:"responseMetaData"`
				}
				json.Unmarshal(resp.Body, &page)
				allItems = append(allItems, page.Responses...)
				if page.ResponseMetaData.NextCursor == "" {
					break
				}
				cursor = page.ResponseMetaData.NextCursor
			}
			merged, _ := json.Marshal(map[string]interface{}{"responses": allItems})
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListResponses(args[0], cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var formDownloadAttachmentCmd = &cobra.Command{
	Use:   "download-attachment <formId> <responseId> <attachmentId>",
	Short: "설문 응답 첨부파일 다운로드",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewFormService(client)

		resp, err := svc.DownloadAttachment(args[0], args[1], args[2])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
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
