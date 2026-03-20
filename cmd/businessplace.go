package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var businessPlaceCmd = &cobra.Command{
	Use:   "business-place",
	Short: "사업장 관리",
}

var bpListCmd = &cobra.Command{
	Use:   "list",
	Short: "사업장 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewBusinessPlaceService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"businessPlaceId", "businessPlaceName"}, "businessPlaces")

		if all {
			var allItems []json.RawMessage
			for {
				resp, err := svc.ListBusinessPlaces(cursor, count)
				if err != nil {
					return err
				}
				var page struct {
					BusinessPlaces   []json.RawMessage `json:"businessPlaces"`
					ResponseMetaData struct {
						NextCursor string `json:"nextCursor"`
					} `json:"responseMetaData"`
				}
				json.Unmarshal(resp.Body, &page)
				allItems = append(allItems, page.BusinessPlaces...)
				if page.ResponseMetaData.NextCursor == "" {
					break
				}
				cursor = page.ResponseMetaData.NextCursor
			}
			merged, _ := json.Marshal(map[string]interface{}{"businessPlaces": allItems})
			formatter.PrintRaw(merged)
			return nil
		}

		resp, err := svc.ListBusinessPlaces(cursor, count)
		if err != nil {
			return err
		}
		formatter.PrintRaw(resp.Body)
		return nil
	},
}

var bpGetCmd = &cobra.Command{
	Use:   "get <businessPlaceId>",
	Short: "사업장 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewBusinessPlaceService(client)

		resp, err := svc.GetBusinessPlace(args[0])
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var bpCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "사업장 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewBusinessPlaceService(client)

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name은 필수입니다")
		}

		body := map[string]interface{}{"businessPlaceName": name}

		resp, err := svc.CreateBusinessPlace(body)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var bpUpdateCmd = &cobra.Command{
	Use:   "update <businessPlaceId>",
	Short: "사업장 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewBusinessPlaceService(client)

		body := map[string]interface{}{}
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			body["businessPlaceName"] = name
		}

		resp, err := svc.UpdateBusinessPlace(args[0], body)
		if err != nil {
			return err
		}
		output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		return nil
	},
}

var bpDeleteCmd = &cobra.Command{
	Use:   "delete <businessPlaceId>",
	Short: "사업장 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token)
		svc := api.NewBusinessPlaceService(client)

		resp, err := svc.DeleteBusinessPlace(args[0])
		if err != nil {
			return err
		}
		if len(resp.Body) == 0 || strings.TrimSpace(string(resp.Body)) == "" {
			fmt.Println("{}")
		} else {
			output.NewFormatter(outputFormat, os.Stdout).PrintRaw(resp.Body)
		}
		return nil
	},
}

func init() {
	bpListCmd.Flags().String("cursor", "", "페이지네이션 커서")
	bpListCmd.Flags().Int("count", 0, "페이지 크기")
	bpListCmd.Flags().Bool("all", false, "전체 페이지 자동 순회")

	bpCreateCmd.Flags().String("name", "", "사업장 이름 (필수)")
	bpUpdateCmd.Flags().String("name", "", "사업장 이름")

	businessPlaceCmd.AddCommand(bpListCmd, bpGetCmd, bpCreateCmd, bpUpdateCmd, bpDeleteCmd)
	rootCmd.AddCommand(businessPlaceCmd)
}
