package api

import (
	"fmt"
	"net/url"
)

type DriveService struct {
	client *Client
}

func NewDriveService(client *Client) *DriveService {
	return &DriveService{client: client}
}

// MyDrive

func (s *DriveService) GetDriveInfo(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive", url.PathEscape(userID)))
}

func (s *DriveService) ListFiles(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/files", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *DriveService) ListFolderChildren(userID, folderID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/files/%s/children", url.PathEscape(userID), url.PathEscape(folderID)) + BuildPaginationQuery(cursor, count))
}

func (s *DriveService) GetFile(userID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/files/%s", url.PathEscape(userID), url.PathEscape(fileID)))
}

func (s *DriveService) GetDownloadURL(userID, fileID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/users/%s/drive/files/%s/download", url.PathEscape(userID), url.PathEscape(fileID)))
}

func (s *DriveService) CreateUploadURL(userID string, body map[string]interface{}, fileSize int64) (*Response, error) {
	body["fileSize"] = fileSize
	return s.client.PostJSON(fmt.Sprintf("/users/%s/drive/files", url.PathEscape(userID)), body)
}

func (s *DriveService) CreateUploadURLInFolder(userID, folderID string, body map[string]interface{}, fileSize int64) (*Response, error) {
	body["fileSize"] = fileSize
	return s.client.PostJSON(fmt.Sprintf("/users/%s/drive/files/%s", url.PathEscape(userID), url.PathEscape(folderID)), body)
}

func (s *DriveService) CreateFolder(userID, name string) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/users/%s/drive/files/createfolder", url.PathEscape(userID)), map[string]string{"fileName": name})
}

func (s *DriveService) CreateFolderInParent(userID, parentID, name string) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/users/%s/drive/files/%s/createfolder", url.PathEscape(userID), url.PathEscape(parentID)), map[string]string{"fileName": name})
}

func (s *DriveService) DeleteFile(userID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/drive/files/%s", url.PathEscape(userID), url.PathEscape(fileID)))
}

func (s *DriveService) ListTrashFiles(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/trash-files", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *DriveService) RestoreTrashFile(userID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/trash-files/%s/restore", url.PathEscape(userID), url.PathEscape(fileID)), nil)
}

// SharedFolders

func (s *DriveService) ListSharedFolders(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *DriveService) ListSharedFolderFiles(userID, sharedFolderID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files", url.PathEscape(userID), url.PathEscape(sharedFolderID)) + BuildPaginationQuery(cursor, count))
}
