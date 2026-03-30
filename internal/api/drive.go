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

// MyDrive File Operations

func (s *DriveService) CopyFile(userID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/copy", url.PathEscape(userID), url.PathEscape(fileID)), body)
}

func (s *DriveService) RenameFile(userID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/rename", url.PathEscape(userID), url.PathEscape(fileID)), body)
}

func (s *DriveService) MoveFile(userID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/move", url.PathEscape(userID), url.PathEscape(fileID)), body)
}

func (s *DriveService) ProtectFile(userID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/protect", url.PathEscape(userID), url.PathEscape(fileID)), nil)
}

func (s *DriveService) UnprotectFile(userID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/unprotect", url.PathEscape(userID), url.PathEscape(fileID)), nil)
}

func (s *DriveService) LockFile(userID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/lock", url.PathEscape(userID), url.PathEscape(fileID)), nil)
}

func (s *DriveService) UnlockFile(userID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/unlock", url.PathEscape(userID), url.PathEscape(fileID)), nil)
}

// MyDrive Revisions

func (s *DriveService) ListRevisions(userID, fileID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/files/%s/revisions", url.PathEscape(userID), url.PathEscape(fileID)) + BuildPaginationQuery(cursor, count))
}

func (s *DriveService) GetRevision(userID, fileID, revisionID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/files/%s/revisions/%s", url.PathEscape(userID), url.PathEscape(fileID), url.PathEscape(revisionID)))
}

func (s *DriveService) RestoreRevision(userID, fileID, revisionID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/revisions/%s/restore", url.PathEscape(userID), url.PathEscape(fileID), url.PathEscape(revisionID)), nil)
}

func (s *DriveService) GetRevisionDownloadURL(userID, fileID, revisionID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/users/%s/drive/files/%s/revisions/%s/download", url.PathEscape(userID), url.PathEscape(fileID), url.PathEscape(revisionID)))
}

// MyDrive Trash

func (s *DriveService) DeleteTrashFile(userID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/drive/trash-files/%s", url.PathEscape(userID), url.PathEscape(fileID)))
}

// MyDrive Link

func (s *DriveService) GetLinkSetting(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/link-setting", url.PathEscape(userID)))
}

func (s *DriveService) GetLink(userID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/files/%s/link", url.PathEscape(userID), url.PathEscape(fileID)))
}

func (s *DriveService) CreateLink(userID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/link", url.PathEscape(userID), url.PathEscape(fileID)), body)
}

func (s *DriveService) PatchLink(userID, fileID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/users/%s/drive/files/%s/link", url.PathEscape(userID), url.PathEscape(fileID)), body)
}

func (s *DriveService) DeleteLink(userID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/drive/files/%s/link", url.PathEscape(userID), url.PathEscape(fileID)))
}

// MyDrive Share

func (s *DriveService) GetShare(userID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/files/%s/share", url.PathEscape(userID), url.PathEscape(fileID)))
}

func (s *DriveService) CreateShare(userID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/drive/files/%s/share", url.PathEscape(userID), url.PathEscape(fileID)), body)
}

func (s *DriveService) PatchShare(userID, fileID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/users/%s/drive/files/%s/share", url.PathEscape(userID), url.PathEscape(fileID)), body)
}

func (s *DriveService) DeleteShare(userID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/drive/files/%s/share", url.PathEscape(userID), url.PathEscape(fileID)))
}

func (s *DriveService) ListShareSubFolders(userID, fileID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/files/%s/share-sub-folders", url.PathEscape(userID), url.PathEscape(fileID)) + BuildPaginationQuery(cursor, count))
}

// SharedFolders

func (s *DriveService) ListSharedFolders(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *DriveService) ListSharedFolderFiles(userID, sharedFolderID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/drive/sharedfolders/%s/files", url.PathEscape(userID), url.PathEscape(sharedFolderID)) + BuildPaginationQuery(cursor, count))
}
