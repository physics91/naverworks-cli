package web

import (
	"fmt"
	"net/url"
	"strconv"
	"unicode/utf8"
)

type Operation struct {
	Domain, Action, RoomID, ResourceID      string
	Title, Body, Cursor, CategoryID, Status string
	Count                                   int
	All                                     bool
	TaskInput                               TaskInput
	AllAssignees                            bool
}

// Plan describes unresolved request values explicitly. Offline preview cannot
// allocate a note ID, resolve a channel, or read fields preserved by an update.
func Plan(op Operation) (map[string]any, error) {
	if err := ValidateOperation(op); err != nil {
		return nil, err
	}
	method, path, origin := "GET", "", talkOrigin
	var body any = map[string]any{}
	query := url.Values{}
	resolution := []string{"channelNo", "tenantId", "collaboSpaceId"}
	if op.Domain == "note" {
		query = url.Values{"apiVer": {"13"}, "channelNo": {"${channelNo}"}, "tenantId": {"${tenantId}"}, "collaboSpaceId": {"${collaboSpaceId}"}, "rl": {"${rl}"}}
		resolution = append(resolution, "rl")
		switch op.Action {
		case "list-posts":
			path = noteBase + "post/selectList"
			count, _ := pageSize(op.Count)
			cursor := op.Cursor
			if cursor == "" {
				cursor = "1"
			}
			query.Set("startIndex", cursor)
			query.Set("selectCount", strconv.Itoa(count))
			query.Set("selectType", "all")
			query.Set("sortBy", "newest")
			query.Set("bodyMaxLength", "200")
		case "get-post":
			path = noteBase + "post/select"
			query.Set("postNo", op.ResourceID)
		case "create-post", "update-post":
			method = "POST"
			query = url.Values{}
			action := "add"
			id := "${nextPostNo}"
			modified := "${lastModifyUtcYmdt}"
			resolution = append(resolution, "lastModifyUtcYmdt")
			if op.Action == "update-post" {
				action = "modify"
				id = op.ResourceID
				resolution = append(resolution, "existingPostFields")
			} else {
				resolution = append(resolution, "nextPostNo")
			}
			path = noteBase + "post/" + action
			values := noteWriteValues(id, op.Title, op.Body, modified)
			values.Set("channelNo", "${channelNo}")
			values.Set("tenantId", "${tenantId}")
			values.Set("collaboSpaceId", "${collaboSpaceId}")
			values.Set("rl", "${rl}")
			values.Set("apiVer", "13")
			if op.Action == "update-post" {
				values.Set("boardNo", "${existing.boardNo}")
				values.Set("isNoti", "${existing.isNoti}")
				values.Set("isCoedit", "${existing.isCoedit}")
				values.Set("attachedJson", "${existing.attachments}")
			}
			payload := map[string]any{}
			for k, v := range values {
				payload[k] = v[0]
			}
			body = payload
		case "delete-post":
			method = "POST"
			query = url.Values{}
			path = noteBase + "post/remove"
			body = map[string]any{"postNo": op.ResourceID, "apiVer": "13", "channelNo": "${channelNo}", "tenantId": "${tenantId}", "collaboSpaceId": "${collaboSpaceId}", "rl": "${rl}"}
		}
	} else {
		origin = taskOrigin
		resolution = append(resolution, "projectId", "currentUserId")
		base := "/rd/${tenantId}/v1/project/${projectId}"
		switch op.Action {
		case "list":
			categoryID := op.CategoryID
			cursor, _ := parseTaskCursor(op.Cursor, op.RoomID)
			if cursor.CategoryID != "" {
				categoryID = cursor.CategoryID
			}
			path = base + "/category/${categoryId}/tasks"
			if categoryID != "" {
				path = base + "/category/" + url.PathEscape(categoryID) + "/tasks"
			}
			count, _ := pageSize(op.Count)
			query = url.Values{"viewType": {"ALL"}, "pageSize": {strconv.Itoa(count)}, "includeDone": {"true"}}
			if cursor.After != "" {
				query.Set("cursor", cursor.After)
			}
			resolution = append(resolution, "categories")
		case "list-categories":
			path = base + "/project-categories"
			query.Set("viewType", "ALL")
		case "get":
			path = base + "/task/" + url.PathEscape(op.ResourceID)
		case "create":
			method = "POST"
			path = "/rd/${tenantId}/v5/project/${projectId}/task"
			body = taskCreateBody("${currentUserId}", op.TaskInput)
		case "update":
			method = "PUT"
			path = "/rd/${tenantId}/v5/project/${projectId}/task/" + url.PathEscape(op.ResourceID)
			payload := map[string]any{"assignorId": "${existing.assignorId}", "assigneeIdList": []string{"${existing.assigneeIds}"}, "title": "${existing.title}", "content": "${existing.content}", "dueDate": "${existing.dueDate}", "taskStatusOption": "${existing.taskStatusOption}", "categoryId": "${existing.categoryId}"}
			applyTaskInput(payload, op.TaskInput)
			body = payload
			resolution = append(resolution, "existingTaskFields")
		case "delete":
			method = "DELETE"
			path = base + "/task/" + url.PathEscape(op.ResourceID)
		case "complete", "incomplete":
			method = "POST"
			path = "/rd/${tenantId}/v2/project/${projectId}/task/" + url.PathEscape(op.ResourceID) + "/taskStatus"
			status := "DONE"
			if op.Action == "incomplete" {
				status = "TODO"
			}
			body = map[string]any{"taskStatus": status, "forceChangeTotalStatus": op.AllAssignees, "language": "ko_KR"}
		}
	}
	headers := (&Client{}).requestHeaders(origin)
	if origin == taskOrigin {
		headers["task-web-client-user-id"] = "${currentUserId}"
	}
	plan := map[string]any{"dry_run": true, "transport": "browser", "headless": true, "room_id": op.RoomID, "operation": op.Domain + " " + op.Action, "method": method, "origin": origin, "path_template": path, "query_template": query, "headers_template": headers, "body": body, "resolution_required": resolution, "resolved": false, "input": map[string]any{"resolution_required": true, "body": body}, "cursor": op.Cursor, "count": op.Count, "category_id": op.CategoryID, "status": op.Status}
	if op.Action == "list" || op.Action == "list-posts" || op.Action == "list-categories" {
		count, _ := pageSize(op.Count)
		pagination := map[string]any{"all": op.All, "start_cursor": op.Cursor, "count": count, "strategy": "follow_next_cursor_until_empty"}
		if op.Action == "list-categories" {
			pagination["page_source"] = "local_slice_of_server_categories"
		}
		plan["pagination"] = pagination
	}
	return plan, nil
}

func ValidateOperation(op Operation) error {
	if err := ValidateRoomID(op.RoomID); err != nil {
		return err
	}
	if op.Domain == "note" {
		switch op.Action {
		case "create-post", "update-post":
			if err := validatePostInput(op.Title, op.Body); err != nil {
				return err
			}
			if utf8.RuneCountInString(op.Title) > 200 {
				return fmt.Errorf("노트 제목은 200자 이하여야 합니다")
			}
		case "list-posts":
			if _, err := pageSize(op.Count); err != nil {
				return err
			}
			if op.Cursor != "" {
				n, err := strconv.Atoi(op.Cursor)
				if err != nil || n < 1 {
					return fmt.Errorf("노트 --cursor는 1 이상의 시작 위치여야 합니다")
				}
			}
		case "get-post", "delete-post":
		default:
			return fmt.Errorf("지원되지 않는 메시지방 노트 명령입니다")
		}
		if op.Action == "get-post" || op.Action == "update-post" || op.Action == "delete-post" {
			if !positiveNumber(op.ResourceID) {
				return fmt.Errorf("노트 postId는 양의 정수 문자열이어야 합니다")
			}
		}
		return nil
	}
	if op.Domain != "task" {
		return fmt.Errorf("지원되지 않는 메시지방 명령입니다")
	}
	switch op.Action {
	case "create", "update":
		if err := op.TaskInput.Validate(op.Action == "create"); err != nil {
			return err
		}
	case "get", "delete", "complete", "incomplete":
	case "list":
		if _, err := pageSize(op.Count); err != nil {
			return err
		}
		position, err := parseTaskCursor(op.Cursor, op.RoomID)
		if err != nil {
			return err
		}
		if op.CategoryID != "" && position.CategoryID != "" && op.CategoryID != position.CategoryID {
			return fmt.Errorf("할 일 --cursor와 --category-id가 일치하지 않습니다")
		}
		if op.Status != "" && op.Status != "DONE" && op.Status != "TODO" {
			return fmt.Errorf("웹 할 일 --status는 TODO 또는 DONE이어야 합니다")
		}
	case "list-categories":
		if _, err := pageSize(op.Count); err != nil {
			return err
		}
		if _, err := parseCategoryCursor(op.Cursor, op.RoomID); err != nil {
			return err
		}
	default:
		return fmt.Errorf("지원되지 않는 메시지방 할 일 명령입니다")
	}
	if op.Action != "create" && op.Action != "list" && op.Action != "list-categories" {
		if !validTaskID(op.ResourceID) {
			return fmt.Errorf("웹 taskId는 t-UUID 형식이어야 합니다")
		}
	}
	return nil
}
