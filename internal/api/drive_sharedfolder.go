package api

import (
	"fmt"
	"net/url"
)

type SharedFolderService struct {
	client *Client
}

func NewSharedFolderService(client *Client) *SharedFolderService {
	return &SharedFolderService{client: client}
}

// Task 5-12: SharedFolder 관리 + 파일

func (s *SharedFolderService) GetFolder(userID, sharedFolderID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s", url.PathEscape(userID), url.PathEscape(sharedFolderID)))
}

func (s *SharedFolderService) LeaveFolder(userID, sharedFolderID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/drive/sharedfolders/%s", url.PathEscape(userID), url.PathEscape(sharedFolderID)))
}

func (s *SharedFolderService) ListMembers(userID, sharedFolderID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/members", url.PathEscape(userID), url.PathEscape(sharedFolderID)))
}

func (s *SharedFolderService) ListFiles(userID, sharedFolderID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files", url.PathEscape(userID), url.PathEscape(sharedFolderID)) + BuildPaginationQuery(cursor, count))
}

func (s *SharedFolderService) CreateUploadURLInRoot(userID, sharedFolderID string, body map[string]interface{}, fileSize int64) (*Response, error) {
	body["fileSize"] = fileSize
	return s.client.PostJSON(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files", url.PathEscape(userID), url.PathEscape(sharedFolderID)), body)
}

func (s *SharedFolderService) CreateFolderInRoot(userID, sharedFolderID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/createfolder", url.PathEscape(userID), url.PathEscape(sharedFolderID)), body)
}

func (s *SharedFolderService) CreateSubFolder(userID, sharedFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/createfolder", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), body)
}

func (s *SharedFolderService) GetFile(userID, sharedFolderID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)))
}

func (s *SharedFolderService) ListFolderChildren(userID, sharedFolderID, folderID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/children", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(folderID)) + BuildPaginationQuery(cursor, count))
}

func (s *SharedFolderService) DeleteFile(userID, sharedFolderID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)))
}

func (s *SharedFolderService) CreateUploadURL(userID, sharedFolderID, fileID string, body map[string]interface{}, fileSize int64) (*Response, error) {
	body["fileSize"] = fileSize
	return s.client.PostJSON(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), body)
}

func (s *SharedFolderService) GetDownloadURL(userID, sharedFolderID, fileID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/download", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)))
}

// Task 5-13: SharedFolder 파일조작 + 리비전

func (s *SharedFolderService) CopyFile(userID, sharedFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/copy", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), body)
}

func (s *SharedFolderService) RenameFile(userID, sharedFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/rename", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), body)
}

func (s *SharedFolderService) MoveFile(userID, sharedFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/move", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), body)
}

func (s *SharedFolderService) ProtectFile(userID, sharedFolderID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/protect", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), nil)
}

func (s *SharedFolderService) UnprotectFile(userID, sharedFolderID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/unprotect", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), nil)
}

func (s *SharedFolderService) LockFile(userID, sharedFolderID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/lock", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), nil)
}

func (s *SharedFolderService) UnlockFile(userID, sharedFolderID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/unlock", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), nil)
}

func (s *SharedFolderService) ListRevisions(userID, sharedFolderID, fileID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/revisions", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)) + BuildPaginationQuery(cursor, count))
}

func (s *SharedFolderService) GetRevision(userID, sharedFolderID, fileID, revisionID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/revisions/%s", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID), url.PathEscape(revisionID)))
}

func (s *SharedFolderService) RestoreRevision(userID, sharedFolderID, fileID, revisionID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/revisions/%s/restore", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID), url.PathEscape(revisionID)), nil)
}

func (s *SharedFolderService) GetRevisionDownloadURL(userID, sharedFolderID, fileID, revisionID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/revisions/%s/download", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID), url.PathEscape(revisionID)))
}

// Task 5-14: SharedFolder 링크

func (s *SharedFolderService) GetLinkSetting(userID, sharedFolderID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/link-setting", url.PathEscape(userID), url.PathEscape(sharedFolderID)))
}

func (s *SharedFolderService) GetLink(userID, sharedFolderID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/link", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)))
}

func (s *SharedFolderService) CreateLink(userID, sharedFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/link", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), body)
}

func (s *SharedFolderService) PatchLink(userID, sharedFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/link", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)), body)
}

func (s *SharedFolderService) DeleteLink(userID, sharedFolderID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files/%s/link", url.PathEscape(userID), url.PathEscape(sharedFolderID), url.PathEscape(fileID)))
}
