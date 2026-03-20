package api

import (
	"encoding/json"
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

func (s *DriveService) GetDownloadURL(userID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/files/%s/download", url.PathEscape(userID), url.PathEscape(fileID)))
}

func (s *DriveService) CreateUploadURL(userID string, body map[string]interface{}) (*Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("업로드 요청 직렬화 실패: %w", err)
	}
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files", url.PathEscape(userID)), data)
}

func (s *DriveService) CreateUploadURLInFolder(userID, folderID string, body map[string]interface{}) (*Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("업로드 요청 직렬화 실패: %w", err)
	}
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s", url.PathEscape(userID), url.PathEscape(folderID)), data)
}

func (s *DriveService) CreateFolder(userID, name string) (*Response, error) {
	data, _ := json.Marshal(map[string]string{"folderName": name})
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/createfolder", url.PathEscape(userID)), data)
}

func (s *DriveService) CreateFolderInParent(userID, parentID, name string) (*Response, error) {
	data, _ := json.Marshal(map[string]string{"folderName": name})
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/createfolder", url.PathEscape(userID), url.PathEscape(parentID)), data)
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

// UploadFile delegates to client's UploadFile for pre-signed URL upload.
func (s *DriveService) UploadFile(uploadURL, filePath string) error {
	return s.client.UploadFile(uploadURL, filePath)
}
