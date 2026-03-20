package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type SharedDriveService struct {
	client *Client
}

func NewSharedDriveService(client *Client) *SharedDriveService {
	return &SharedDriveService{client: client}
}

func (s *SharedDriveService) ListDrives(cursor string, count int) (*Response, error) {
	return s.client.Get("/sharedrives" + BuildPaginationQuery(cursor, count))
}

func (s *SharedDriveService) GetDrive(driveID string) (*Response, error) {
	return s.client.Get("/sharedrives/" + url.PathEscape(driveID))
}

func (s *SharedDriveService) ListFiles(driveID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/files", url.PathEscape(driveID)) + BuildPaginationQuery(cursor, count))
}

func (s *SharedDriveService) ListFolderChildren(driveID, folderID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/files/%s/children", url.PathEscape(driveID), url.PathEscape(folderID)) + BuildPaginationQuery(cursor, count))
}

func (s *SharedDriveService) GetFile(driveID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/files/%s", url.PathEscape(driveID), url.PathEscape(fileID)))
}

func (s *SharedDriveService) GetDownloadURL(driveID, fileID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/sharedrives/%s/files/%s/download", url.PathEscape(driveID), url.PathEscape(fileID)))
}

func (s *SharedDriveService) CreateUploadURL(driveID string, body map[string]interface{}, fileSize int64) (*Response, error) {
	body["fileSize"] = fileSize
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("업로드 요청 직렬화 실패: %w", err)
	}
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files", url.PathEscape(driveID)), data)
}

func (s *SharedDriveService) UploadFile(uploadURL, filePath string) error {
	return s.client.UploadFile(uploadURL, filePath)
}
