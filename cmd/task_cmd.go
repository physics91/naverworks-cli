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
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
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
		return getAndPrint(func(client *api.Client) (*api.Response, error) {
			return api.NewTaskService(client).GetTask(args[0])
		})
	},
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "태스크 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
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
		svc, err := newSvc(api.NewTaskService)
		if err != nil {
			return err
		}

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
		svc, err := newSvc(api.NewTaskService)
		if err != nil {
			return err
		}

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
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewTaskService(client)
		return runListCmd(cmd, []string{"categoryId", "categoryName"}, "taskCategories", func(c string, n int) (*api.Response, error) {
			return svc.ListCategories(userID, c, n)
		})
	},
}

var taskCreateCategoryCmd = &cobra.Command{
	Use:   "create-category",
	Short: "태스크 카테고리 생성",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewTaskService(client)
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.CreateCategory(userID, body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var taskGetCategoryCmd = &cobra.Command{
	Use:   "get-category <categoryId>",
	Short: "태스크 카테고리 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewTaskService(client)
		resp, err := svc.GetCategory(userID, args[0])
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var taskUpdateCategoryCmd = &cobra.Command{
	Use:   "update-category <categoryId>",
	Short: "태스크 카테고리 수정",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewTaskService(client)
		body, err := readJSONFlag(cmd)
		if err != nil {
			return err
		}
		resp, err := svc.PatchCategory(userID, args[0], body)
		if err != nil {
			return err
		}
		printBody(resp.Body)
		return nil
	},
}

var taskDeleteCategoryCmd = &cobra.Command{
	Use:   "delete-category <categoryId>",
	Short: "태스크 카테고리 삭제",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewTaskService(client)
		resp, err := svc.DeleteCategory(userID, args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var taskMoveCmd = &cobra.Command{
	Use:   "move <taskId>",
	Short: "태스크 이동",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, userID, err := newAPIClientWithUser(cmd)
		if err != nil {
			return err
		}
		svc := api.NewTaskService(client)
		categoryID, _ := cmd.Flags().GetString("category")
		if categoryID == "" {
			return fmt.Errorf("--category는 필수입니다")
		}
		body := buildTaskMoveBody(categoryID)
		resp, err := svc.MoveTask(userID, args[0], body)
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var taskCompleteCmd = &cobra.Command{
	Use:   "complete <taskId>",
	Short: "태스크 완료 처리",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewTaskService)
		if err != nil {
			return err
		}
		resp, err := svc.CompleteTask(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var taskIncompleteCmd = &cobra.Command{
	Use:   "incomplete <taskId>",
	Short: "태스크 미완료 처리",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewTaskService)
		if err != nil {
			return err
		}
		resp, err := svc.IncompleteTask(args[0])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var taskCompleteAssigneeCmd = &cobra.Command{
	Use:   "complete-assignee <taskId> <userId>",
	Short: "태스크 담당자 완료 처리",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewTaskService)
		if err != nil {
			return err
		}
		resp, err := svc.CompleteAssigneeTask(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

var taskIncompleteAssigneeCmd = &cobra.Command{
	Use:   "incomplete-assignee <taskId> <userId>",
	Short: "태스크 담당자 미완료 처리",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		svc, err := newSvc(api.NewTaskService)
		if err != nil {
			return err
		}
		resp, err := svc.IncompleteAssigneeTask(args[0], args[1])
		if err != nil {
			return err
		}
		printResponse(resp)
		return nil
	},
}

func init() {
	addListFlags(taskListCmd, taskListCategoriesCmd)
	for _, c := range []*cobra.Command{taskListCmd, taskCreateCmd, taskListCategoriesCmd,
		taskCreateCategoryCmd, taskGetCategoryCmd, taskUpdateCategoryCmd, taskDeleteCategoryCmd, taskMoveCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}

	taskCreateCmd.Flags().String("title", "", "태스크 제목 (필수)")
	taskCreateCmd.Flags().String("data", "", "전체 JSON 페이로드")
	taskUpdateCmd.Flags().String("title", "", "태스크 제목")
	taskUpdateCmd.Flags().String("data", "", "전체 JSON 페이로드")

	taskCreateCategoryCmd.Flags().String("json", "", "JSON 페이로드")
	taskUpdateCategoryCmd.Flags().String("json", "", "JSON 페이로드")

	taskMoveCmd.Flags().String("category", "", "이동할 카테고리 ID (필수)")

	taskCmd.AddCommand(taskListCmd, taskGetCmd, taskCreateCmd, taskUpdateCmd, taskDeleteCmd, taskListCategoriesCmd,
		taskCreateCategoryCmd, taskGetCategoryCmd, taskUpdateCategoryCmd, taskDeleteCategoryCmd,
		taskMoveCmd, taskCompleteCmd, taskIncompleteCmd, taskCompleteAssigneeCmd, taskIncompleteAssigneeCmd)
	rootCmd.AddCommand(taskCmd)
}

func buildTaskMoveBody(categoryID string) map[string]interface{} {
	return map[string]interface{}{"toCategoryId": categoryID}
}
