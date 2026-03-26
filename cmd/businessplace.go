package cmd

import (
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
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
		return runListCmd(cmd, []string{"businessPlaceId", "businessPlaceName"}, "businessPlaces", svc.ListBusinessPlaces)
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
		printBody(resp.Body)
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

		body, err := parseOptionalJSONData(cmd)
		if err != nil {
			return err
		}
		if body == nil {
			bpName, _ := cmd.Flags().GetString("name")
			if bpName == "" {
				return fmt.Errorf("--name은 필수입니다")
			}
			body = map[string]interface{}{"businessPlaceName": bpName}
		}

		resp, err := svc.CreateBusinessPlace(body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
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

		body, err := parseOptionalJSONData(cmd)
		if err != nil {
			return err
		}
		if body == nil {
			body = map[string]interface{}{}
			if bpName, _ := cmd.Flags().GetString("name"); bpName != "" {
				body["businessPlaceName"] = bpName
			}
		}

		resp, err := svc.UpdateBusinessPlace(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
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
