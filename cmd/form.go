package cmd

import (
	"encoding/json"
	"fmt"
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
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewFormService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"responseId"}, "responses")

		if all {
			items, err := api.PaginateAll(func(c string) (*api.Response, error) {
				return svc.ListResponses(args[0], c, count)
			}, "responses")
			if err != nil {
				return err
			}
			merged, err := json.Marshal(map[string]interface{}{"responses": json.RawMessage(items)})
			if err != nil {
				return fmt.Errorf("결과 직렬화 실패: %w", err)
			}
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
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(result)
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
