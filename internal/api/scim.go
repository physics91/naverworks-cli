package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type ScimService struct {
	client *Client
}

func NewScimService(client *Client) *ScimService {
	return &ScimService{client: client}
}

// Users

func (s *ScimService) ListUsers(startIndex, count int, filter string) (*Response, error) {
	params := url.Values{}
	if startIndex > 0 {
		params.Set("startIndex", fmt.Sprintf("%d", startIndex))
	}
	if count > 0 {
		params.Set("count", fmt.Sprintf("%d", count))
	}
	if filter != "" {
		params.Set("filter", filter)
	}
	query := ""
	if len(params) > 0 {
		query = "?" + params.Encode()
	}
	return s.client.Get("/Users" + query)
}

func (s *ScimService) GetUser(id string) (*Response, error) {
	return s.client.Get("/Users/" + url.PathEscape(id))
}

func (s *ScimService) CreateUser(body map[string]interface{}) (*Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("SCIM 사용자 직렬화 실패: %w", err)
	}
	return s.client.Post("/Users", data)
}

func (s *ScimService) UpdateUser(id string, body map[string]interface{}) (*Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("SCIM 사용자 직렬화 실패: %w", err)
	}
	return s.client.Put("/Users/"+url.PathEscape(id), data)
}

func (s *ScimService) PatchUser(id string, body map[string]interface{}) (*Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("SCIM 사용자 직렬화 실패: %w", err)
	}
	return s.client.Patch("/Users/"+url.PathEscape(id), data)
}

func (s *ScimService) DeleteUser(id string) (*Response, error) {
	return s.client.Delete("/Users/" + url.PathEscape(id))
}

// Groups

func (s *ScimService) ListGroups(startIndex, count int, filter string) (*Response, error) {
	params := url.Values{}
	if startIndex > 0 {
		params.Set("startIndex", fmt.Sprintf("%d", startIndex))
	}
	if count > 0 {
		params.Set("count", fmt.Sprintf("%d", count))
	}
	if filter != "" {
		params.Set("filter", filter)
	}
	query := ""
	if len(params) > 0 {
		query = "?" + params.Encode()
	}
	return s.client.Get("/Groups" + query)
}

func (s *ScimService) GetGroup(id string) (*Response, error) {
	return s.client.Get("/Groups/" + url.PathEscape(id))
}

func (s *ScimService) CreateGroup(body map[string]interface{}) (*Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("SCIM 그룹 직렬화 실패: %w", err)
	}
	return s.client.Post("/Groups", data)
}

func (s *ScimService) UpdateGroup(id string, body map[string]interface{}) (*Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("SCIM 그룹 직렬화 실패: %w", err)
	}
	return s.client.Put("/Groups/"+url.PathEscape(id), data)
}

func (s *ScimService) PatchGroup(id string, body map[string]interface{}) (*Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("SCIM 그룹 직렬화 실패: %w", err)
	}
	return s.client.Patch("/Groups/"+url.PathEscape(id), data)
}

func (s *ScimService) DeleteGroup(id string) (*Response, error) {
	return s.client.Delete("/Groups/" + url.PathEscape(id))
}
