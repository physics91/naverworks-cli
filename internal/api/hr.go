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

func (s *HRService) CreateExtensionProperty(body []byte) (*Response, error) {
	return s.client.Post("/business-support/human-resource/extension-properties", body)
}

func (s *HRService) GetExtensionProperty(id string) (*Response, error) {
	return s.client.Get("/business-support/human-resource/extension-properties/" + url.PathEscape(id))
}

func (s *HRService) PatchExtensionProperty(id string, body []byte) (*Response, error) {
	return s.client.Patch("/business-support/human-resource/extension-properties/"+url.PathEscape(id), body)
}

func (s *HRService) DeleteExtensionProperty(id string) (*Response, error) {
	return s.client.Delete("/business-support/human-resource/extension-properties/" + url.PathEscape(id))
}

func (s *HRService) GetUserExtensionProperty(userID, id string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/business-support/human-resource/user/%s/extension-properties/%s", url.PathEscape(userID), url.PathEscape(id)))
}

func (s *HRService) PatchUserExtensionProperty(userID, id string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/business-support/human-resource/user/%s/extension-properties/%s", url.PathEscape(userID), url.PathEscape(id)), body)
}

func (s *HRService) CreateLeaveOfAbsence(body []byte) (*Response, error) {
	return s.client.Post("/business-support/human-resource/leave-of-absences", body)
}

func (s *HRService) GetLeaveOfAbsence(id string) (*Response, error) {
	return s.client.Get("/business-support/human-resource/leave-of-absences/" + url.PathEscape(id))
}

func (s *HRService) PatchLeaveOfAbsence(id string, body []byte) (*Response, error) {
	return s.client.Patch("/business-support/human-resource/leave-of-absences/"+url.PathEscape(id), body)
}

func (s *HRService) DeleteLeaveOfAbsence(id string) (*Response, error) {
	return s.client.Delete("/business-support/human-resource/leave-of-absences/" + url.PathEscape(id))
}
