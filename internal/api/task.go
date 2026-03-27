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
