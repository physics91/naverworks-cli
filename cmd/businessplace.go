package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBusinessPlaceService(client)

		cursor, _ := cmd.Flags().GetString("cursor")
		count, _ := cmd.Flags().GetInt("count")
		all, _ := cmd.Flags().GetBool("all")

		formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable([]string{"businessPlaceId", "businessPlaceName"}, "businessPlaces")

		if all {
			return paginateAndPrint(func(c string) (*api.Response, error) {
				return svc.ListBusinessPlaces(c, count)
			}, "businessPlaces", formatter)
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
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
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
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBusinessPlaceService(client)

		var body map[string]interface{}
		data, _ := cmd.Flags().GetString("data")
		if data != "" {
			if err := json.Unmarshal([]byte(data), &body); err != nil {
				return fmt.Errorf("--data JSON 파싱 실패: %w", err)
			}
		} else {
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name은 필수입니다")
			}
			body = map[string]interface{}{"businessPlaceName": name}
		}

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
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBusinessPlaceService(client)

		var body map[string]interface{}
		data, _ := cmd.Flags().GetString("data")
		if data != "" {
			if err := json.Unmarshal([]byte(data), &body); err != nil {
				return fmt.Errorf("--data JSON 파싱 실패: %w", err)
			}
		} else {
			body = map[string]interface{}{}
			if name, _ := cmd.Flags().GetString("name"); name != "" {
				body["businessPlaceName"] = name
			}
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
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewBusinessPlaceService(client)

		resp, err := svc.DeleteBusinessPlace(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

func init() {
	bpListCmd.Flags().String("cursor", "", "페이지네이션 커서")
	bpListCmd.Flags().Int("count", 0, "페이지 크기")
	bpListCmd.Flags().Bool("all", false, "전체 페이지 자동 순회")

	bpCreateCmd.Flags().String("name", "", "사업장 이름 (필수)")
	bpCreateCmd.Flags().String("data", "", "전체 JSON 페이로드")
	bpUpdateCmd.Flags().String("name", "", "사업장 이름")
	bpUpdateCmd.Flags().String("data", "", "전체 JSON 페이로드")

	businessPlaceCmd.AddCommand(bpListCmd, bpGetCmd, bpCreateCmd, bpUpdateCmd, bpDeleteCmd)
	rootCmd.AddCommand(businessPlaceCmd)
}
