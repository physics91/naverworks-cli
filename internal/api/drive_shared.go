package api

import (
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
	return s.client.PostJSON(fmt.Sprintf("/sharedrives/%s/files", url.PathEscape(driveID)), body)
}

// Task 5-8: SharedDrive 관리 + 파일 보완

func (s *SharedDriveService) CreateDrive(body []byte) (*Response, error) {
	return s.client.Post("/sharedrives", body)
}

func (s *SharedDriveService) PatchDrive(driveID string, body []byte) (*Response, error) {
	return s.client.Patch("/sharedrives/"+url.PathEscape(driveID), body)
}

func (s *SharedDriveService) DeleteDrive(driveID string) (*Response, error) {
	return s.client.Delete("/sharedrives/" + url.PathEscape(driveID))
}

func (s *SharedDriveService) CreateFolderInRoot(driveID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/createfolder", url.PathEscape(driveID)), body)
}

func (s *SharedDriveService) CreateSubFolder(driveID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/createfolder", url.PathEscape(driveID), url.PathEscape(fileID)), body)
}

func (s *SharedDriveService) DeleteFile(driveID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/sharedrives/%s/files/%s", url.PathEscape(driveID), url.PathEscape(fileID)))
}

func (s *SharedDriveService) CreateUploadURLInFolder(driveID, fileID string, body map[string]interface{}, fileSize int64) (*Response, error) {
	body["fileSize"] = fileSize
	return s.client.PostJSON(fmt.Sprintf("/sharedrives/%s/files/%s", url.PathEscape(driveID), url.PathEscape(fileID)), body)
}

// Task 5-9: SharedDrive 파일조작

func (s *SharedDriveService) CopyFile(driveID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/copy", url.PathEscape(driveID), url.PathEscape(fileID)), body)
}

func (s *SharedDriveService) RenameFile(driveID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/rename", url.PathEscape(driveID), url.PathEscape(fileID)), body)
}

func (s *SharedDriveService) MoveFile(driveID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/move", url.PathEscape(driveID), url.PathEscape(fileID)), body)
}

func (s *SharedDriveService) ProtectFile(driveID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/protect", url.PathEscape(driveID), url.PathEscape(fileID)), nil)
}

func (s *SharedDriveService) UnprotectFile(driveID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/unprotect", url.PathEscape(driveID), url.PathEscape(fileID)), nil)
}

func (s *SharedDriveService) LockFile(driveID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/lock", url.PathEscape(driveID), url.PathEscape(fileID)), nil)
}

func (s *SharedDriveService) UnlockFile(driveID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/unlock", url.PathEscape(driveID), url.PathEscape(fileID)), nil)
}

// Task 5-10: SharedDrive 리비전 + 휴지통

func (s *SharedDriveService) ListRevisions(driveID, fileID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/files/%s/revisions", url.PathEscape(driveID), url.PathEscape(fileID)) + BuildPaginationQuery(cursor, count))
}

func (s *SharedDriveService) GetRevision(driveID, fileID, revisionID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/files/%s/revisions/%s", url.PathEscape(driveID), url.PathEscape(fileID), url.PathEscape(revisionID)))
}

func (s *SharedDriveService) RestoreRevision(driveID, fileID, revisionID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/revisions/%s/restore", url.PathEscape(driveID), url.PathEscape(fileID), url.PathEscape(revisionID)), nil)
}

func (s *SharedDriveService) GetRevisionDownloadURL(driveID, fileID, revisionID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/sharedrives/%s/files/%s/revisions/%s/download", url.PathEscape(driveID), url.PathEscape(fileID), url.PathEscape(revisionID)))
}

func (s *SharedDriveService) ListTrashFiles(driveID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/trash-files", url.PathEscape(driveID)) + BuildPaginationQuery(cursor, count))
}

func (s *SharedDriveService) RestoreTrashFile(driveID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/trash-files/%s/restore", url.PathEscape(driveID), url.PathEscape(fileID)), nil)
}

func (s *SharedDriveService) DeleteTrashFile(driveID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/sharedrives/%s/trash-files/%s", url.PathEscape(driveID), url.PathEscape(fileID)))
}

// Task 5-11: SharedDrive 링크 + 권한

func (s *SharedDriveService) GetLinkSetting(driveID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/link-setting", url.PathEscape(driveID)))
}

func (s *SharedDriveService) GetLink(driveID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/files/%s/link", url.PathEscape(driveID), url.PathEscape(fileID)))
}

func (s *SharedDriveService) CreateLink(driveID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/link", url.PathEscape(driveID), url.PathEscape(fileID)), body)
}

func (s *SharedDriveService) PatchLink(driveID, fileID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/sharedrives/%s/files/%s/link", url.PathEscape(driveID), url.PathEscape(fileID)), body)
}

func (s *SharedDriveService) DeleteLink(driveID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/sharedrives/%s/files/%s/link", url.PathEscape(driveID), url.PathEscape(fileID)))
}

// Drive-level permissions

func (s *SharedDriveService) ListPermissions(driveID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/permissions", url.PathEscape(driveID)))
}

func (s *SharedDriveService) CreatePermission(driveID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/permissions", url.PathEscape(driveID)), body)
}

func (s *SharedDriveService) GetPermission(driveID, permissionID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/permissions/%s", url.PathEscape(driveID), url.PathEscape(permissionID)))
}

func (s *SharedDriveService) PatchPermission(driveID, permissionID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/sharedrives/%s/permissions/%s", url.PathEscape(driveID), url.PathEscape(permissionID)), body)
}

func (s *SharedDriveService) DeletePermission(driveID, permissionID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/sharedrives/%s/permissions/%s", url.PathEscape(driveID), url.PathEscape(permissionID)))
}

func (s *SharedDriveService) DeleteAllPermissions(driveID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/sharedrives/%s/permissions", url.PathEscape(driveID)))
}

func (s *SharedDriveService) EnablePermissions(driveID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/permissions/enable", url.PathEscape(driveID)), nil)
}

func (s *SharedDriveService) DisablePermissions(driveID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/permissions/disable", url.PathEscape(driveID)), nil)
}

// File-level permissions

func (s *SharedDriveService) ListFilePermissions(driveID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/files/%s/permissions", url.PathEscape(driveID), url.PathEscape(fileID)))
}

func (s *SharedDriveService) CreateFilePermission(driveID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/permissions", url.PathEscape(driveID), url.PathEscape(fileID)), body)
}

func (s *SharedDriveService) GetFilePermission(driveID, fileID, permissionID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/sharedrives/%s/files/%s/permissions/%s", url.PathEscape(driveID), url.PathEscape(fileID), url.PathEscape(permissionID)))
}

func (s *SharedDriveService) PatchFilePermission(driveID, fileID, permissionID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/sharedrives/%s/files/%s/permissions/%s", url.PathEscape(driveID), url.PathEscape(fileID), url.PathEscape(permissionID)), body)
}

func (s *SharedDriveService) DeleteFilePermission(driveID, fileID, permissionID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/sharedrives/%s/files/%s/permissions/%s", url.PathEscape(driveID), url.PathEscape(fileID), url.PathEscape(permissionID)))
}

func (s *SharedDriveService) DeleteAllFilePermissions(driveID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/sharedrives/%s/files/%s/permissions", url.PathEscape(driveID), url.PathEscape(fileID)))
}

func (s *SharedDriveService) EnableFilePermissions(driveID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/permissions/enable", url.PathEscape(driveID), url.PathEscape(fileID)), nil)
}

func (s *SharedDriveService) DisableFilePermissions(driveID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/sharedrives/%s/files/%s/permissions/disable", url.PathEscape(driveID), url.PathEscape(fileID)), nil)
}
