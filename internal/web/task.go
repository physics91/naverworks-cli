package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/chromedp/chromedp"
)

func validTaskID(id string) bool {
	return strings.HasPrefix(id, "t-") && roomIDPattern.MatchString(strings.TrimPrefix(id, "t-"))
}

type Assignee struct {
	ID     string `json:"memberId"`
	Status string `json:"taskStatus"`
}

type Task struct {
	ID           string     `json:"taskId"`
	ProjectID    string     `json:"projectId"`
	CategoryID   string     `json:"categoryId"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Status       string     `json:"taskStatus"`
	StatusOption string     `json:"taskStatusOption"`
	AssignorID   string     `json:"assignorId"`
	Assignees    []Assignee `json:"assigneeList"`
	DueDate      *string    `json:"dueDate"`
	Created      string     `json:"created,omitempty"`
	Modified     string     `json:"modified,omitempty"`
}

type Category struct {
	ID      string `json:"categoryId"`
	Name    string `json:"categoryName"`
	Default bool   `json:"defaultProjectCategory"`
}

func (c *Client) prepareTasks() error {
	if c.project != "" {
		return nil
	}
	q := url.Values{"viewType": {"COLLABOSPACE"}, "serviceCode": {"WEB_TALK"}, "channelNo": {c.room.ChannelNo}, "viewSubId": {c.room.TenantID + "_" + c.room.CollaboSpaceID}}
	if err := chromedp.Run(c.ctx, chromedp.Navigate(taskOrigin+"/main/tasks/webtalk/list?"+q.Encode())); err != nil {
		return fmt.Errorf("할 일 웹 서비스에 연결하지 못했습니다")
	}
	body, err := c.get(taskOrigin, "/v2/userInfo", nil)
	if err != nil {
		return err
	}
	var user struct{ UserID string }
	if err := responseJSON(body, &user); err != nil {
		return err
	}
	if !positiveNumber(user.UserID) {
		return fmt.Errorf("할 일 로그인 사용자 정보가 유효하지 않습니다")
	}
	c.userID = user.UserID
	body, err = c.get(taskOrigin, "/rd/"+url.PathEscape(c.room.TenantID)+"/v1/collaboSpace/"+url.PathEscape(c.room.CollaboSpaceID)+"/projectId", nil)
	if err != nil {
		return err
	}
	project := strings.TrimSpace(string(body))
	if len(project) > 100 || !strings.HasPrefix(project, "p"+c.room.TenantID+"-") || !roomIDPattern.MatchString(strings.TrimPrefix(project, "p"+c.room.TenantID+"-")) {
		return fmt.Errorf("메시지방 할 일 프로젝트 정보가 유효하지 않습니다")
	}
	c.project = project
	return nil
}

func (c *Client) taskPath(version, suffix string) string {
	return "/rd/" + url.PathEscape(c.room.TenantID) + "/" + version + "/project/" + url.PathEscape(c.project) + suffix
}

func (c *Client) taskCategories() ([]Category, error) {
	if err := c.prepareTasks(); err != nil {
		return nil, err
	}
	body, err := c.get(taskOrigin, c.taskPath("v1", "/project-categories"), url.Values{"viewType": {"ALL"}})
	if err != nil {
		return nil, err
	}
	var result []struct{ ProjectCategory Category }
	if err := responseJSON(body, &result); err != nil {
		return nil, err
	}
	categories := make([]Category, 0, len(result))
	for _, e := range result {
		if e.ProjectCategory.ID == "" {
			return nil, fmt.Errorf("할 일 카테고리 정보가 변경됐습니다")
		}
		categories = append(categories, e.ProjectCategory)
	}
	return categories, nil
}

func (c *Client) ListTaskCategories(cursor string, count int) ([]byte, error) {
	if _, err := parseCategoryCursor(cursor, c.room.ID); err != nil {
		return nil, err
	}
	if _, err := pageSize(count); err != nil {
		return nil, err
	}
	categories, err := c.taskCategories()
	if err != nil {
		return nil, err
	}
	return categoryPage(categories, cursor, count, c.room.ID)
}

func parseCategoryCursor(cursor, roomID string) (taskCursor, error) {
	if cursor == "" {
		return taskCursor{}, nil
	}
	if !strings.HasPrefix(cursor, "categories:") {
		return taskCursor{}, fmt.Errorf("카테고리 --cursor가 유효하지 않습니다")
	}
	position, err := parseTaskCursor(strings.TrimPrefix(cursor, "categories:"), roomID)
	if err != nil || position.CategoryID == "" || position.After != "" || position.Read != 0 {
		return taskCursor{}, fmt.Errorf("카테고리 --cursor가 유효하지 않거나 다른 방의 커서입니다")
	}
	return position, nil
}

func categoryPage(categories []Category, cursor string, count int, roomID string) ([]byte, error) {
	position, err := parseCategoryCursor(cursor, roomID)
	if err != nil {
		return nil, err
	}
	count, err = pageSize(count)
	if err != nil {
		return nil, err
	}
	start := 0
	if position.CategoryID != "" {
		start = -1
		for i, category := range categories {
			if category.ID == position.CategoryID {
				start = i
				break
			}
		}
		if start < 0 {
			return nil, fmt.Errorf("카테고리 커서의 항목을 찾을 수 없습니다")
		}
	}
	end := min(start+count, len(categories))
	next := ""
	if end < len(categories) {
		data, _ := json.Marshal(taskCursor{RoomID: roomID, CategoryID: categories[end].ID})
		next = "categories:" + base64.RawURLEncoding.EncodeToString(data)
	}
	page := append([]Category{}, categories[start:end]...)
	return json.Marshal(map[string]any{"categories": page, "responseMetaData": map[string]string{"nextCursor": next}})
}

type taskCursor struct {
	RoomID     string `json:"room"`
	CategoryID string `json:"category"`
	After      string `json:"after,omitempty"`
	Read       int    `json:"read,omitempty"`
}

func parseTaskCursor(cursor, roomID string) (taskCursor, error) {
	var value taskCursor
	if cursor == "" {
		return value, nil
	}
	if len(cursor) > 4096 {
		return value, fmt.Errorf("할 일 --cursor가 너무 깁니다")
	}
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil || json.Unmarshal(data, &value) != nil || !strings.EqualFold(value.RoomID, roomID) || value.CategoryID == "" || value.Read < 0 {
		return value, fmt.Errorf("할 일 --cursor가 유효하지 않거나 다른 방의 커서입니다")
	}
	return value, nil
}

func (c *Client) ListTasks(cursor string, count int, categoryID, status string) ([]byte, error) {
	count, err := pageSize(count)
	if err != nil {
		return nil, err
	}
	if status != "" && status != "TODO" && status != "DONE" {
		return nil, fmt.Errorf("웹 할 일 --status는 TODO 또는 DONE이어야 합니다")
	}
	position, err := parseTaskCursor(cursor, c.room.ID)
	if err != nil {
		return nil, err
	}
	categories, err := c.taskCategories()
	if err != nil {
		return nil, err
	}
	if categoryID != "" {
		var selected []Category
		for _, cat := range categories {
			if cat.ID == categoryID {
				selected = append(selected, cat)
			}
		}
		categories = selected
		if len(categories) == 0 {
			return nil, fmt.Errorf("지정한 카테고리는 이 방에 속하지 않습니다")
		}
	}
	index := 0
	if position.CategoryID != "" {
		index = -1
		for i, cat := range categories {
			if cat.ID == position.CategoryID {
				index = i
			}
		}
		if index < 0 {
			return nil, fmt.Errorf("할 일 커서의 카테고리를 찾을 수 없습니다")
		}
	}
	tasks := make([]Task, 0)
	next := ""
	if len(categories) > 0 {
		cat := categories[index]
		q := url.Values{"viewType": {"ALL"}, "pageSize": {strconv.Itoa(count)}, "includeDone": {"true"}}
		if position.After != "" {
			q.Set("cursor", position.After)
		}
		body, err := c.get(taskOrigin, c.taskPath("v1", "/category/"+url.PathEscape(cat.ID)+"/tasks"), q)
		if err != nil {
			return nil, err
		}
		var result struct {
			CategoryTaskList struct {
				CategoryID   string
				AllTaskCount int
				Tasks        []Task
			}
			ResponseMetaData struct{ NextCursor string }
		}
		if err := responseJSON(body, &result); err != nil {
			return nil, err
		}
		if result.CategoryTaskList.CategoryID != cat.ID {
			return nil, fmt.Errorf("할 일 조회 카테고리가 요청과 일치하지 않습니다")
		}
		for _, t := range result.CategoryTaskList.Tasks {
			if t.ProjectID != c.project {
				return nil, fmt.Errorf("다른 방의 할 일이 반환됐습니다")
			}
			if status == "" || t.Status == status {
				tasks = append(tasks, t)
			}
		}
		read := position.Read + len(result.CategoryTaskList.Tasks)
		var nextPosition taskCursor
		if read < result.CategoryTaskList.AllTaskCount && len(result.CategoryTaskList.Tasks) > 0 {
			if result.ResponseMetaData.NextCursor == "" || result.ResponseMetaData.NextCursor == position.After {
				return nil, fmt.Errorf("할 일 목록 커서가 진행되지 않습니다")
			}
			nextPosition = taskCursor{RoomID: c.room.ID, CategoryID: cat.ID, After: result.ResponseMetaData.NextCursor, Read: read}
		} else if index+1 < len(categories) {
			nextPosition = taskCursor{RoomID: c.room.ID, CategoryID: categories[index+1].ID}
		}
		if nextPosition.CategoryID != "" {
			encoded, _ := json.Marshal(nextPosition)
			next = base64.RawURLEncoding.EncodeToString(encoded)
		}
	}
	return json.Marshal(map[string]any{"tasks": tasks, "responseMetaData": map[string]string{"nextCursor": next}})
}

func (c *Client) GetTask(id string) (Task, error) {
	if !validTaskID(id) {
		return Task{}, fmt.Errorf("웹 taskId는 t-UUID 형식이어야 합니다")
	}
	if err := c.prepareTasks(); err != nil {
		return Task{}, err
	}
	body, err := c.get(taskOrigin, c.taskPath("v1", "/task/"+url.PathEscape(id)), nil)
	if err != nil {
		return Task{}, err
	}
	var task Task
	if err := responseJSON(body, &task); err != nil {
		return Task{}, err
	}
	if task.ID != id || task.ProjectID != c.project {
		return Task{}, fmt.Errorf("할 일이 지정한 방에 속하지 않습니다")
	}
	return task, nil
}

type TaskInput struct {
	Title        *string  `json:"title,omitempty"`
	Content      *string  `json:"content,omitempty"`
	DueDate      *string  `json:"dueDate,omitempty"`
	AssignorID   *string  `json:"assignorId,omitempty"`
	AssigneeIDs  []string `json:"assigneeIdList,omitempty"`
	StatusOption *string  `json:"taskStatusOption,omitempty"`
}

func (in TaskInput) Validate(create bool) error {
	if create && in.Title == nil || in.Title != nil && strings.TrimSpace(*in.Title) == "" {
		return fmt.Errorf("할 일 --title은 비어 있을 수 없습니다")
	}
	if in.Title != nil && utf8.RuneCountInString(*in.Title) > 200 {
		return fmt.Errorf("할 일 제목은 200자 이하여야 합니다")
	}
	if in.Content != nil && utf8.RuneCountInString(*in.Content) > 5000 {
		return fmt.Errorf("할 일 설명은 5000자 이하여야 합니다")
	}
	if in.AssigneeIDs != nil && len(in.AssigneeIDs) == 0 {
		return fmt.Errorf("할 일 담당자는 비어 있을 수 없습니다")
	}
	if in.DueDate != nil && *in.DueDate != "" {
		if _, err := time.Parse("2006-01-02", *in.DueDate); err != nil {
			return fmt.Errorf("할 일 마감일은 YYYY-MM-DD 형식이어야 합니다")
		}
	}
	if in.AssignorID != nil && !positiveNumber(*in.AssignorID) {
		return fmt.Errorf("웹 요청자 ID는 내부 사용자 번호여야 합니다")
	}
	seen := make(map[string]bool, len(in.AssigneeIDs))
	for _, id := range in.AssigneeIDs {
		if !positiveNumber(id) {
			return fmt.Errorf("웹 담당자 ID는 내부 사용자 번호여야 합니다")
		}
		if seen[id] {
			return fmt.Errorf("할 일 담당자를 중복 지정할 수 없습니다")
		}
		seen[id] = true
	}
	if in.StatusOption != nil && *in.StatusOption != "ANY_ONE" && *in.StatusOption != "MUST_ALL" {
		return fmt.Errorf("완료 조건은 ANY_ONE 또는 MUST_ALL이어야 합니다")
	}
	return nil
}

func (c *Client) CreateTask(input TaskInput) ([]byte, error) {
	if err := input.Validate(true); err != nil {
		return nil, err
	}
	if err := c.prepareTasks(); err != nil {
		return nil, err
	}
	body := taskCreateBody(c.userID, input)
	before, err := c.allTasks()
	if err != nil {
		return nil, err
	}
	if _, err := c.postJSON(taskOrigin, c.taskPath("v5", "/task"), body); err != nil {
		return nil, err
	}
	after, err := c.allTasks()
	if err != nil {
		return nil, err
	}
	id, err := createdTaskID(before, after, body)
	if err != nil {
		return nil, err
	}
	created, err := c.GetTask(id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(created)
}

func (c *Client) allTasks() ([]Task, error) {
	tasks := make([]Task, 0)
	cursor := ""
	for page := 0; page < 1000; page++ {
		data, err := c.ListTasks(cursor, 100, "", "")
		if err != nil {
			return nil, err
		}
		var result struct {
			Tasks            []Task
			ResponseMetaData struct{ NextCursor string }
		}
		if err := responseJSON(data, &result); err != nil {
			return nil, err
		}
		tasks = append(tasks, result.Tasks...)
		if result.ResponseMetaData.NextCursor == "" {
			return tasks, nil
		}
		if result.ResponseMetaData.NextCursor == cursor {
			return nil, fmt.Errorf("할 일 목록 커서가 진행되지 않습니다")
		}
		cursor = result.ResponseMetaData.NextCursor
	}
	return nil, fmt.Errorf("할 일 생성 확인용 목록의 페이지 제한을 초과했습니다")
}

// The web editor reloads the list after a write instead of requiring a complete
// task in the response. Match exactly one newly added task; never retry a write
// or adopt a pre-existing task with the same title.
func createdTaskID(before, after []Task, body map[string]any) (string, error) {
	existing := make(map[string]bool, len(before))
	for _, task := range before {
		existing[task.ID] = true
	}
	wantAssignees := body["assigneeIdList"].([]string)
	wantIDs := make(map[string]bool, len(wantAssignees))
	for _, id := range wantAssignees {
		wantIDs[id] = true
	}
	id := ""
	wantDueDate, _ := body["dueDate"].(string)
	for _, task := range after {
		dueDate := ""
		if task.DueDate != nil {
			dueDate = *task.DueDate
		}
		if existing[task.ID] || task.Title != body["title"] || task.Content != body["content"] || task.AssignorID != body["assignorId"] || task.StatusOption != body["taskStatusOption"] || len(task.Assignees) != len(wantAssignees) || dueDate != wantDueDate {
			continue
		}
		matches := true
		seen := make(map[string]bool, len(task.Assignees))
		for _, assignee := range task.Assignees {
			matches = matches && wantIDs[assignee.ID] && !seen[assignee.ID]
			seen[assignee.ID] = true
		}
		if !matches {
			continue
		}
		if id != "" || !validTaskID(task.ID) {
			return "", fmt.Errorf("할 일 생성 결과를 하나로 확인하지 못했습니다. 웹에서 반영 여부를 확인하세요")
		}
		id = task.ID
	}
	if id == "" {
		return "", fmt.Errorf("할 일 생성 결과를 확인하지 못했습니다. 웹에서 반영 여부를 확인하세요")
	}
	return id, nil
}

func taskCreateBody(userID string, input TaskInput) map[string]any {
	body := map[string]any{"assignorId": userID, "assigneeIdList": []string{userID}, "title": "", "content": "", "dueDate": "", "taskStatusOption": "ANY_ONE", "collaboSpaceTaskInfo": map[string]bool{"collaboSpaceTaskNotify": false}}
	applyTaskInput(body, input)
	return body
}

func applyTaskInput(body map[string]any, input TaskInput) {
	if input.Title != nil {
		body["title"] = strings.TrimSpace(*input.Title)
	}
	if input.Content != nil {
		body["content"] = strings.TrimSpace(*input.Content)
	}
	if input.DueDate != nil {
		if *input.DueDate == "" {
			body["dueDate"] = nil
		} else {
			body["dueDate"] = *input.DueDate
		}
	}
	if input.AssignorID != nil {
		body["assignorId"] = *input.AssignorID
	}
	if len(input.AssigneeIDs) > 0 {
		body["assigneeIdList"] = input.AssigneeIDs
	}
	if input.StatusOption != nil {
		body["taskStatusOption"] = *input.StatusOption
	}
}

func (c *Client) UpdateTask(id string, input TaskInput) ([]byte, error) {
	if err := input.Validate(false); err != nil {
		return nil, err
	}
	task, err := c.GetTask(id)
	if err != nil {
		return nil, err
	}
	body := taskUpdateBody(task, input)
	data, _ := json.Marshal(body)
	if _, err := c.request(taskOrigin, "PUT", c.taskPath("v5", "/task/"+url.PathEscape(id)), "application/json", string(data)); err != nil {
		return nil, err
	}
	updated, err := c.GetTask(id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(updated)
}

func taskUpdateBody(task Task, input TaskInput) map[string]any {
	ids := make([]string, 0, len(task.Assignees))
	for _, a := range task.Assignees {
		ids = append(ids, a.ID)
	}
	body := map[string]any{"assignorId": task.AssignorID, "assigneeIdList": ids, "title": task.Title, "content": task.Content, "dueDate": task.DueDate, "taskStatusOption": task.StatusOption, "categoryId": task.CategoryID}
	applyTaskInput(body, input)
	return body
}

func (c *Client) DeleteTask(id string) ([]byte, error) {
	if _, err := c.GetTask(id); err != nil {
		return nil, err
	}
	if _, err := c.request(taskOrigin, "DELETE", c.taskPath("v1", "/task/"+url.PathEscape(id)), "", ""); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"taskId": id, "deleted": true})
}

func (c *Client) SetTaskStatus(id, status string, all bool) ([]byte, error) {
	if status != "DONE" && status != "TODO" {
		return nil, fmt.Errorf("할 일 상태가 유효하지 않습니다")
	}
	if _, err := c.GetTask(id); err != nil {
		return nil, err
	}
	if _, err := c.postJSON(taskOrigin, c.taskPath("v2", "/task/"+url.PathEscape(id)+"/taskStatus"), map[string]any{"taskStatus": status, "forceChangeTotalStatus": all, "language": "ko_KR"}); err != nil {
		return nil, err
	}
	task, err := c.GetTask(id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(task)
}
