package api

import (
	"fmt"
	"net/url"
)

type HRService struct {
	client *Client
}

func NewHRService(client *Client) *HRService {
	return &HRService{client: client}
}

func (s *HRService) ListExtensionProperties(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/human-resource/extension-properties" + BuildPaginationQuery(cursor, count))
}

func (s *HRService) GetUserExtensionProperties(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/business-support/human-resource/user/%s/extension-properties", url.PathEscape(userID)))
}

func (s *HRService) ListLeaveOfAbsences(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/human-resource/leave-of-absences" + BuildPaginationQuery(cursor, count))
}

func (s *HRService) ListOnLeaveUsers(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/human-resource/on-leave-users" + BuildPaginationQuery(cursor, count))
}
