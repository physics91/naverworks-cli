package cmd

import (
	"fmt"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "태스크 관리",
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "태스크 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewTaskService(client)
		return runListCmd(cmd, []string{"taskId", "title"}, "tasks", func(c string, n int) (*api.Response, error) {
			return svc.ListTasks(userID, c, n)
		})
	},
}

var taskGetCmd = &cobra.Command{
	Use:   "get <taskId>",
	Short: "태스크 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewTaskService(client)

		resp, err := svc.GetTask(args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "태스크 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewTaskService(client)

		body, err := parseOptionalJSONData(cmd)
		if err != nil {
			return err
		}
		if body == nil {
			title, _ := cmd.Flags().GetString("title")
			if title == "" {
				return fmt.Errorf("--title은 필수입니다")
			}
			body = map[string]interface{}{"title": title}
		}

		resp, err := svc.CreateTask(userID, body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update <taskId>",
	Short: "태스크 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewTaskService(client)

		body, err := parseOptionalJSONData(cmd)
		if err != nil {
			return err
		}
		if body == nil {
			body = map[string]interface{}{}
			if title, _ := cmd.Flags().GetString("title"); title != "" {
				body["title"] = title
			}
		}

		resp, err := svc.UpdateTask(args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete <taskId>",
	Short: "태스크 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewTaskService(client)

		resp, err := svc.DeleteTask(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var taskListCategoriesCmd = &cobra.Command{
	Use:   "list-categories",
	Short: "태스크 카테고리 목록 조회",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, token, name, err := loadConfigAndToken()
		if err != nil {
			return err
		}
		userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
		if err != nil {
			return err
		}
		client := buildAPIClient(cfg, token, name)
		svc := api.NewTaskService(client)
		return runListCmd(cmd, []string{"categoryId", "categoryName"}, "taskCategories", func(c string, n int) (*api.Response, error) {
			return svc.ListCategories(userID, c, n)
		})
	},
}

func init() {
	for _, c := range []*cobra.Command{taskListCmd, taskListCategoriesCmd} {
		c.Flags().String("cursor", "", "페이지네이션 커서")
		c.Flags().Int("count", 0, "페이지 크기")
		c.Flags().Bool("all", false, "전체 페이지 자동 순회")
	}
	for _, c := range []*cobra.Command{taskListCmd, taskCreateCmd, taskListCategoriesCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	taskCreateCmd.Flags().String("title", "", "태스크 제목 (필수)")
	taskCreateCmd.Flags().String("data", "", "전체 JSON 페이로드")
	taskUpdateCmd.Flags().String("title", "", "태스크 제목")
	taskUpdateCmd.Flags().String("data", "", "전체 JSON 페이로드")

	taskCmd.AddCommand(taskListCmd, taskGetCmd, taskCreateCmd, taskUpdateCmd, taskDeleteCmd, taskListCategoriesCmd)
	rootCmd.AddCommand(taskCmd)
}
