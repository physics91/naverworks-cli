package api

import (
	"fmt"
	"net/url"
)

type TaskService struct {
	client *Client
}

func NewTaskService(client *Client) *TaskService {
	return &TaskService{client: client}
}

func (s *TaskService) ListTasks(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/tasks", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *TaskService) GetTask(taskID string) (*Response, error) {
	return s.client.Get("/tasks/" + url.PathEscape(taskID))
}

func (s *TaskService) CreateTask(userID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/users/%s/tasks", url.PathEscape(userID)), body)
}

func (s *TaskService) UpdateTask(taskID string, body map[string]interface{}) (*Response, error) {
	return s.client.PatchJSON("/tasks/"+url.PathEscape(taskID), body)
}

func (s *TaskService) DeleteTask(taskID string) (*Response, error) {
	return s.client.Delete("/tasks/" + url.PathEscape(taskID))
}

func (s *TaskService) ListCategories(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/task-categories", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *TaskService) CreateCategory(userID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/users/%s/task-categories", url.PathEscape(userID)), body)
}

func (s *TaskService) GetCategory(userID string, categoryID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/task-categories/%s", url.PathEscape(userID), url.PathEscape(categoryID)))
}

func (s *TaskService) PatchCategory(userID string, categoryID string, body map[string]interface{}) (*Response, error) {
	return s.client.PatchJSON(fmt.Sprintf("/users/%s/task-categories/%s", url.PathEscape(userID), url.PathEscape(categoryID)), body)
}

func (s *TaskService) DeleteCategory(userID string, categoryID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/task-categories/%s", url.PathEscape(userID), url.PathEscape(categoryID)))
}

func (s *TaskService) MoveTask(userID string, taskID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/users/%s/tasks/%s/move", url.PathEscape(userID), url.PathEscape(taskID)), body)
}

func (s *TaskService) CompleteTask(taskID string) (*Response, error) {
	return s.client.PostJSON("/tasks/"+url.PathEscape(taskID)+"/complete", nil)
}

func (s *TaskService) IncompleteTask(taskID string) (*Response, error) {
	return s.client.PostJSON("/tasks/"+url.PathEscape(taskID)+"/incomplete", nil)
}

func (s *TaskService) CompleteAssigneeTask(taskID string, userID string) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/tasks/%s/assignees/%s/complete", url.PathEscape(taskID), url.PathEscape(userID)), nil)
}

func (s *TaskService) IncompleteAssigneeTask(taskID string, userID string) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/tasks/%s/assignees/%s/incomplete", url.PathEscape(taskID), url.PathEscape(userID)), nil)
}
