package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "태스크 관리",
}

const taskSearchHelp = "인증: 구성원 계정 Access Token 전용 (서비스 계정 사용 불가)\n최소 scope: task.read"

func newTaskSearchClientWithUser(cmd *cobra.Command) (*api.Client, string, error) {
	client, cfg, token, err := newAPIClient()
	if err != nil {
		return nil, "", err
	}
	if token.AuthMethod == auth.AuthMethodJWT {
		return nil, "", fmt.Errorf("Task 검색은 구성원 계정 Access Token만 지원합니다")
	}
	userID, err := resolveUserID(cmd, cfg.DefaultCalendarUserID, token.AuthMethod)
	if err != nil {
		return nil, "", err
	}
	return client, userID, nil
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
		categoryID, _ := cmd.Flags().GetString("category-id")
		status, _ := cmd.Flags().GetString("status")
		searchFilterType, _ := cmd.Flags().GetString("search-filter-type")
		opts := api.TaskListOptions{
			CategoryID:       categoryID,
			Status:           status,
			SearchFilterType: searchFilterType,
		}
		return runListCmd(cmd, []string{"taskId", "title"}, "tasks", func(c string, n int) (*api.Response, error) {
			return svc.ListTasks(userID, c, n, opts)
		})
	},
}

var taskSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "사용자 할 일 검색",
	Long:  "사용자 할 일 검색\n\n" + taskSearchHelp,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) == 1 {
			query = strings.TrimSpace(args[0])
			if query == "" {
				return fmt.Errorf("query는 비어 있을 수 없습니다")
			}
		}
		assignorID, _ := cmd.Flags().GetString("assignor-id")
		assigneeID, _ := cmd.Flags().GetString("assignee-id")
		startTime, _ := cmd.Flags().GetString("start-time")
		endTime, _ := cmd.Flags().GetString("end-time")
		if query == "" && strings.TrimSpace(assignorID) == "" && strings.TrimSpace(assigneeID) == "" && strings.TrimSpace(startTime) == "" && strings.TrimSpace(endTime) == "" {
			return fmt.Errorf("query 또는 --assignor-id, --assignee-id, --start-time, --end-time 중 하나가 필요합니다")
		}

		client, userID, err := newTaskSearchClientWithUser(cmd)
		if err != nil {
			return err
		}
		status, _ := cmd.Flags().GetString("status")
		orderBy, _ := cmd.Flags().GetString("order-by")
		hasDueDate, _ := cmd.Flags().GetBool("has-due-date")
		hasAttachment, _ := cmd.Flags().GetBool("has-attachment")
		opts := api.TaskSearchOptions{
			Query:      query,
			AssignorID: assignorID,
			AssigneeID: assigneeID,
			StartTime:  startTime,
			EndTime:    endTime,
			Status:     status,
			OrderBy:    orderBy,
		}
		if cmd.Flags().Changed("has-due-date") {
			opts.HasDueDate = strconv.FormatBool(hasDueDate)
		}
		if cmd.Flags().Changed("has-attachment") {
			opts.HasAttachment = strconv.FormatBool(hasAttachment)
		}

		svc := api.NewTaskService(client)
		return runListCmd(cmd, []string{"taskId", "title", "status", "dueDate"}, "tasks", func(cursor string, count int) (*api.Response, error) {
			return svc.SearchTasks(userID, cursor, count, opts)
		})
	},
}

var taskGetCmd = &cobra.Command{
	Use:   "get <taskId>",
	Short: "태스크 상세 조회",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndPrint(func(client *api.Client) (*api.Response, error) {
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
			if desc, _ := cmd.Flags().GetString("description"); desc != "" {
				body["description"] = desc
			}
			if dueDate, _ := cmd.Flags().GetString("due-date"); dueDate != "" {
				body["dueDate"] = dueDate
			}
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
			if desc, _ := cmd.Flags().GetString("description"); desc != "" {
				body["description"] = desc
			}
			if dueDate, _ := cmd.Flags().GetString("due-date"); dueDate != "" {
				body["dueDate"] = dueDate
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
	addListFlags(taskListCmd, taskSearchCmd, taskListCategoriesCmd)
	for _, c := range []*cobra.Command{taskListCmd, taskCreateCmd, taskListCategoriesCmd,
		taskCreateCategoryCmd, taskGetCategoryCmd, taskUpdateCategoryCmd, taskDeleteCategoryCmd, taskMoveCmd} {
		c.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	}
	taskSearchCmd.Flags().String("user-id", "", "사용자 ID (OAuth: me 허용)")
	taskSearchCmd.Flags().String("assignor-id", "", "요청자 사용자 ID")
	taskSearchCmd.Flags().String("assignee-id", "", "담당자 사용자 ID")
	taskSearchCmd.Flags().String("start-time", "", "검색 시작 일시 (ISO-8601)")
	taskSearchCmd.Flags().String("end-time", "", "검색 종료 일시 (ISO-8601)")
	taskSearchCmd.Flags().String("status", "", "할 일 상태 (DONE|TODO)")
	taskSearchCmd.Flags().Bool("has-due-date", false, "마감일 존재 여부")
	taskSearchCmd.Flags().Bool("has-attachment", false, "첨부 파일 존재 여부")
	taskSearchCmd.Flags().String("order-by", "", "정렬 기준과 순서 (createdTime|dueDate, asc|desc)")

	taskListCmd.Flags().String("category-id", "", "카테고리 ID (categoryId)")
	taskListCmd.Flags().String("status", "", "상태 필터 (status)")
	taskListCmd.Flags().String("search-filter-type", "", "검색 필터 (searchFilterType: ALL|ASSIGNEE|ASSIGNOR)")

	taskCreateCmd.Flags().String("title", "", "태스크 제목 (필수)")
	taskCreateCmd.Flags().String("description", "", "태스크 설명")
	taskCreateCmd.Flags().String("due-date", "", "마감일 (YYYY-MM-DD)")
	taskCreateCmd.Flags().String("data", "", "전체 JSON 페이로드")
	taskUpdateCmd.Flags().String("title", "", "태스크 제목")
	taskUpdateCmd.Flags().String("description", "", "태스크 설명")
	taskUpdateCmd.Flags().String("due-date", "", "마감일 (YYYY-MM-DD)")
	taskUpdateCmd.Flags().String("data", "", "전체 JSON 페이로드")

	taskCreateCategoryCmd.Flags().String("json", "", "JSON 페이로드")
	taskUpdateCategoryCmd.Flags().String("json", "", "JSON 페이로드")

	taskMoveCmd.Flags().String("category", "", "이동할 카테고리 ID (필수)")

	taskCmd.AddCommand(taskListCmd, taskSearchCmd, taskGetCmd, taskCreateCmd, taskUpdateCmd, taskDeleteCmd, taskListCategoriesCmd,
		taskCreateCategoryCmd, taskGetCategoryCmd, taskUpdateCategoryCmd, taskDeleteCategoryCmd,
		taskMoveCmd, taskCompleteCmd, taskIncompleteCmd, taskCompleteAssigneeCmd, taskIncompleteAssigneeCmd)
	rootCmd.AddCommand(taskCmd)
}

func buildTaskMoveBody(categoryID string) map[string]interface{} {
	return map[string]interface{}{"toCategoryId": categoryID}
}
