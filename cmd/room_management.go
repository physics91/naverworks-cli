package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/physics91/naverworks-cli/internal/fileutil"
	"github.com/physics91/naverworks-cli/internal/web"
	"github.com/spf13/cobra"
)

func init() {
	for _, entry := range []struct {
		command *cobra.Command
		arity   int
	}{
		{noteListPostsCmd, 0}, {noteGetPostCmd, 1}, {noteCreatePostCmd, 0}, {noteUpdatePostCmd, 1}, {noteDeletePostCmd, 1},
		{taskListCmd, 0}, {taskGetCmd, 1}, {taskCreateCmd, 0}, {taskUpdateCmd, 1}, {taskDeleteCmd, 1}, {taskCompleteCmd, 1}, {taskIncompleteCmd, 1}, {taskListCategoriesCmd, 0},
	} {
		enableRoomManagement(entry.command, entry.arity)
	}
	for _, cmd := range []*cobra.Command{taskCreateCmd, taskUpdateCmd} {
		cmd.Flags().String("assignor-id", "", "메시지방 할 일 요청자의 내부 사용자 번호 (--room-id 전용; 생성 기본값: 로그인 사용자)")
		cmd.Flags().StringSlice("assignee-ids", nil, "메시지방 할 일 담당자의 내부 사용자 번호 목록 (--room-id 전용; 생성 기본값: 로그인 사용자)")
		cmd.Flags().String("completion-condition", "", "메시지방 완료 조건: ANY_ONE 또는 MUST_ALL (--room-id 전용; 생성 기본값: ANY_ONE)")
	}
	for _, cmd := range []*cobra.Command{taskCompleteCmd, taskIncompleteCmd} {
		cmd.Flags().Bool("all-assignees", false, "메시지방 할 일 전체 담당자의 상태 변경 (--room-id 전용; 기본값: 내 상태만 변경)")
	}
}

// The API path keeps its existing arguments and RunE. The browser path accepts
// only the resource ID, and never silently treats a groupId as a roomId.
func enableRoomManagement(cmd *cobra.Command, arity int) {
	apiArgs, apiRun := cmd.Args, cmd.RunE
	cmd.Flags().String("room-id", "", "메시지방 채널 UUID (auth login --method browser 로그인 필요)")
	help := cmd.Long
	if help == "" {
		help = cmd.Short
	}
	cmd.Long = help + "\n\n--room-id 지정 시 전용 세션의 headless 브라우저를 사용합니다. 로그인할 때만 브라우저 창이 열립니다.\n노트의 groupId 인수는 생략합니다. 미리보기는 네트워크 없이 요청 템플릿을 표시하며 ${...} 값은 실행 시 조회합니다.\n--generate-input도 미해결 필드를 명시합니다. 웹 내부 요청은 서비스 변경에 영향을 받을 수 있습니다."
	if strings.Contains(cmd.Use, "<groupId>") {
		cmd.Use = strings.Replace(cmd.Use, "<groupId>", "[groupId]", 1)
	}
	cmd.Args = func(cmd *cobra.Command, args []string) error {
		if cmd.Flags().Changed("room-id") {
			id, _ := cmd.Flags().GetString("room-id")
			if err := web.ValidateRoomID(id); err != nil {
				return err
			}
			return cobra.ExactArgs(arity)(cmd, args)
		}
		if apiArgs != nil {
			return apiArgs(cmd, args)
		}
		return nil
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("room-id") {
			for _, flag := range []string{"assignor-id", "assignee-ids", "completion-condition", "all-assignees"} {
				if cmd.Flags().Changed(flag) {
					return fmt.Errorf("--%s는 --room-id와 함께 사용하세요", flag)
				}
			}
			return apiRun(cmd, args)
		}
		return runRoomManagement(cmd, args)
	}
}

func roomTaskInput(cmd *cobra.Command) (web.TaskInput, error) {
	var input web.TaskInput
	data, _ := cmd.Flags().GetString("data")
	if data != "" {
		for _, flag := range []string{"title", "description", "due-date", "assignor-id", "assignee-ids", "completion-condition"} {
			if cmd.Flags().Changed(flag) {
				return input, fmt.Errorf("--data와 --%s를 함께 사용할 수 없습니다", flag)
			}
		}
		decoder := json.NewDecoder(strings.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			return input, fmt.Errorf("메시지방 --data에는 title, content, dueDate, assignorId, assigneeIdList, taskStatusOption만 허용됩니다")
		}
		if decoder.Decode(new(any)) != io.EOF {
			return input, fmt.Errorf("--data에는 JSON 객체 하나만 지정하세요")
		}
	} else {
		for _, entry := range []struct {
			flag   string
			target **string
		}{{"title", &input.Title}, {"description", &input.Content}, {"due-date", &input.DueDate}, {"assignor-id", &input.AssignorID}, {"completion-condition", &input.StatusOption}} {
			if cmd.Flags().Changed(entry.flag) {
				value, _ := cmd.Flags().GetString(entry.flag)
				*entry.target = &value
			}
		}
		if cmd.Flags().Changed("assignee-ids") {
			input.AssigneeIDs, _ = cmd.Flags().GetStringSlice("assignee-ids")
			if len(input.AssigneeIDs) == 0 {
				return input, fmt.Errorf("--assignee-ids는 비어 있을 수 없습니다")
			}
		}
	}
	create := cmd == taskCreateCmd
	if !create && input.Title == nil && input.Content == nil && input.DueDate == nil && input.AssignorID == nil && input.StatusOption == nil && len(input.AssigneeIDs) == 0 {
		return input, fmt.Errorf("수정할 할 일 필드를 지정하세요")
	}
	return input, input.Validate(create)
}

func runRoomManagement(cmd *cobra.Command, args []string) error {
	id, _ := cmd.Flags().GetString("room-id")
	if cmd.Flags().Changed("user-id") {
		return fmt.Errorf("--room-id는 로그인한 브라우저 계정을 사용하므로 --user-id와 함께 사용할 수 없습니다")
	}
	if cmd.Flags().Changed("search-filter-type") {
		return fmt.Errorf("메시지방에서는 --search-filter-type 대신 --status와 --category-id를 사용하세요")
	}
	var input web.TaskInput
	var title, body string
	var err error
	if cmd == taskCreateCmd || cmd == taskUpdateCmd {
		input, err = roomTaskInput(cmd)
		if err != nil {
			return err
		}
	}
	if cmd == noteCreatePostCmd || cmd == noteUpdatePostCmd {
		post, err := requireTitleBodyPost(cmd)
		if err != nil {
			return err
		}
		title = post["title"].(string)
		body = post["body"].(string)
	}
	operation := roomOperation(cmd, args, id, input, title, body)
	if err := web.ValidateOperation(operation); err != nil {
		return err
	}
	_, profile, err := loadActiveConfig()
	if err != nil {
		return err
	}
	if previewRequested() {
		return printRoomPreview(cmd, profile, operation)
	}
	configPath, err := config.DefaultPathOrError()
	if err != nil {
		return err
	}
	dir, err := web.SessionDir(configPath, profile)
	if err != nil {
		return err
	}
	client, err := (web.Browser{SessionDir: dir, Profile: profile}).Open(cmd.Context())
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.ResolveRoom(id); err != nil {
		return err
	}
	if cmd == noteListPostsCmd {
		return runListCmd(cmd, []string{"postId", "title"}, "posts", func(cursor string, count int) (*api.Response, error) {
			body, err := client.ListPosts(cursor, count)
			return &api.Response{Body: body}, err
		})
	}
	if cmd == taskListCmd {
		category, _ := cmd.Flags().GetString("category-id")
		status, _ := cmd.Flags().GetString("status")
		return runListCmd(cmd, []string{"taskId", "title", "taskStatus", "categoryId"}, "tasks", func(cursor string, count int) (*api.Response, error) {
			body, err := client.ListTasks(cursor, count, category, status)
			return &api.Response{Body: body}, err
		})
	}
	if cmd == taskListCategoriesCmd {
		return runListCmd(cmd, []string{"categoryId", "categoryName"}, "categories", func(cursor string, count int) (*api.Response, error) {
			body, err := client.ListTaskCategories(cursor, count)
			return &api.Response{Body: body}, err
		})
	}
	var result []byte
	switch cmd {
	case noteGetPostCmd:
		post, e := client.GetPost(args[0])
		err = e
		if err == nil {
			result, err = json.Marshal(post)
		}
	case noteCreatePostCmd:
		result, err = client.CreatePost(title, body)
	case noteUpdatePostCmd:
		result, err = client.UpdatePost(args[0], title, body)
	case noteDeletePostCmd:
		result, err = client.DeletePost(args[0])
	case taskGetCmd:
		task, e := client.GetTask(args[0])
		err = e
		if err == nil {
			result, err = json.Marshal(task)
		}
	case taskCreateCmd:
		result, err = client.CreateTask(input)
	case taskUpdateCmd:
		result, err = client.UpdateTask(args[0], input)
	case taskDeleteCmd:
		result, err = client.DeleteTask(args[0])
	case taskCompleteCmd, taskIncompleteCmd:
		status := "DONE"
		if cmd == taskIncompleteCmd {
			status = "TODO"
		}
		all, _ := cmd.Flags().GetBool("all-assignees")
		result, err = client.SetTaskStatus(args[0], status, all)
	default:
		return fmt.Errorf("지원되지 않는 메시지방 명령입니다")
	}
	if err != nil {
		return err
	}
	printBody(result)
	return nil
}

func roomOperation(cmd *cobra.Command, args []string, id string, input web.TaskInput, title, body string) web.Operation {
	operation := web.Operation{Domain: cmd.Parent().Name(), Action: cmd.Name(), RoomID: id, TaskInput: input, Title: title, Body: body}
	if len(args) > 0 {
		operation.ResourceID = args[0]
	}
	operation.Cursor, operation.Count, operation.All = listFlags(cmd)
	operation.CategoryID, _ = cmd.Flags().GetString("category-id")
	operation.Status, _ = cmd.Flags().GetString("status")
	operation.AllAssignees, _ = cmd.Flags().GetBool("all-assignees")
	return operation
}

func printRoomPreview(cmd *cobra.Command, profile string, operation web.Operation) error {
	plan, err := web.Plan(operation)
	if err != nil {
		return err
	}
	plan["profile"] = profile
	if strings.TrimSpace(planOutPath) != "" {
		if err := fileutil.WriteSecureJSON(planOutPath, plan); err != nil {
			return err
		}
	}
	var out []byte
	if generateInput {
		out, err = json.Marshal(plan["input"])
	} else {
		out, err = json.Marshal(plan)
	}
	if err != nil {
		return err
	}
	printBody(out)
	return nil
}
