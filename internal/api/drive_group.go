package api

import (
	"fmt"
	"net/url"
)

type GroupFolderService struct {
	client *Client
}

func NewGroupFolderService(client *Client) *GroupFolderService {
	return &GroupFolderService{client: client}
}

func (s *GroupFolderService) GetFolder(groupID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder", url.PathEscape(groupID)))
}

func (s *GroupFolderService) ListFiles(groupID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder/files", url.PathEscape(groupID)) + BuildPaginationQuery(cursor, count))
}

func (s *GroupFolderService) ListFolderChildren(groupID, folderID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder/files/%s/children", url.PathEscape(groupID), url.PathEscape(folderID)) + BuildPaginationQuery(cursor, count))
}

func (s *GroupFolderService) GetFile(groupID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder/files/%s", url.PathEscape(groupID), url.PathEscape(fileID)))
}
