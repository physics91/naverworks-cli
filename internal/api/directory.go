package api

import "net/url"

type DirectoryService struct {
	client *Client
}

func NewDirectoryService(client *Client) *DirectoryService {
	return &DirectoryService{client: client}
}

func (s *DirectoryService) ListUsers(cursor string, count int) (*Response, error) {
	return s.client.Get("/users" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetUser(userID string) (*Response, error) {
	return s.client.Get("/users/" + url.PathEscape(userID))
}

func (s *DirectoryService) ListGroups(cursor string, count int) (*Response, error) {
	return s.client.Get("/groups" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetGroup(groupID string) (*Response, error) {
	return s.client.Get("/groups/" + url.PathEscape(groupID))
}

func (s *DirectoryService) ListOrgUnits(cursor string, count int) (*Response, error) {
	return s.client.Get("/orgunits" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetOrgUnit(orgUnitID string) (*Response, error) {
	return s.client.Get("/orgunits/" + url.PathEscape(orgUnitID))
}

func (s *DirectoryService) ListLevels(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/levels" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) ListPositions(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/positions" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) ListUserTypes(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/user-types" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) ListEmploymentTypes(cursor string, count int) (*Response, error) {
	return s.client.Get("/directory/employment-types" + BuildPaginationQuery(cursor, count))
}
