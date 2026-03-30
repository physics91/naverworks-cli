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

// Task 5-4: GroupFolder 관리 + 파일 보완

func (s *GroupFolderService) CreateFolder(groupID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder", url.PathEscape(groupID)), nil)
}

func (s *GroupFolderService) DeleteFolder(groupID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/folder", url.PathEscape(groupID)))
}

func (s *GroupFolderService) CreateFolderInRoot(groupID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/createfolder", url.PathEscape(groupID)), body)
}

func (s *GroupFolderService) CreateSubFolder(groupID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/createfolder", url.PathEscape(groupID), url.PathEscape(fileID)), body)
}

func (s *GroupFolderService) DeleteFile(groupID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/folder/files/%s", url.PathEscape(groupID), url.PathEscape(fileID)))
}

func (s *GroupFolderService) CreateUploadURL(groupID, fileID string, body map[string]interface{}, fileSize int64) (*Response, error) {
	body["fileSize"] = fileSize
	return s.client.PostJSON(fmt.Sprintf("/groups/%s/folder/files/%s", url.PathEscape(groupID), url.PathEscape(fileID)), body)
}

func (s *GroupFolderService) GetDownloadURL(groupID, fileID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/groups/%s/folder/files/%s/download", url.PathEscape(groupID), url.PathEscape(fileID)))
}

// Task 5-5: GroupFolder 파일조작

func (s *GroupFolderService) CopyFile(groupID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/copy", url.PathEscape(groupID), url.PathEscape(fileID)), body)
}

func (s *GroupFolderService) RenameFile(groupID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/rename", url.PathEscape(groupID), url.PathEscape(fileID)), body)
}

func (s *GroupFolderService) MoveFile(groupID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/move", url.PathEscape(groupID), url.PathEscape(fileID)), body)
}

func (s *GroupFolderService) ProtectFile(groupID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/protect", url.PathEscape(groupID), url.PathEscape(fileID)), nil)
}

func (s *GroupFolderService) UnprotectFile(groupID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/unprotect", url.PathEscape(groupID), url.PathEscape(fileID)), nil)
}

func (s *GroupFolderService) LockFile(groupID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/lock", url.PathEscape(groupID), url.PathEscape(fileID)), nil)
}

func (s *GroupFolderService) UnlockFile(groupID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/unlock", url.PathEscape(groupID), url.PathEscape(fileID)), nil)
}

// Task 5-6: GroupFolder 리비전 + 휴지통

func (s *GroupFolderService) ListRevisions(groupID, fileID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder/files/%s/revisions", url.PathEscape(groupID), url.PathEscape(fileID)) + BuildPaginationQuery(cursor, count))
}

func (s *GroupFolderService) GetRevision(groupID, fileID, revisionID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder/files/%s/revisions/%s", url.PathEscape(groupID), url.PathEscape(fileID), url.PathEscape(revisionID)))
}

func (s *GroupFolderService) RestoreRevision(groupID, fileID, revisionID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/revisions/%s/restore", url.PathEscape(groupID), url.PathEscape(fileID), url.PathEscape(revisionID)), nil)
}

func (s *GroupFolderService) GetRevisionDownloadURL(groupID, fileID, revisionID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/groups/%s/folder/files/%s/revisions/%s/download", url.PathEscape(groupID), url.PathEscape(fileID), url.PathEscape(revisionID)))
}

func (s *GroupFolderService) ListTrashFiles(groupID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder/trash-files", url.PathEscape(groupID)) + BuildPaginationQuery(cursor, count))
}

func (s *GroupFolderService) RestoreTrashFile(groupID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/trash-files/%s/restore", url.PathEscape(groupID), url.PathEscape(fileID)), nil)
}

func (s *GroupFolderService) DeleteTrashFile(groupID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/folder/trash-files/%s", url.PathEscape(groupID), url.PathEscape(fileID)))
}

// Task 5-7: GroupFolder 링크 + 권한

func (s *GroupFolderService) GetLinkSetting(groupID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder/link-setting", url.PathEscape(groupID)))
}

func (s *GroupFolderService) GetLink(groupID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder/files/%s/link", url.PathEscape(groupID), url.PathEscape(fileID)))
}

func (s *GroupFolderService) CreateLink(groupID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/link", url.PathEscape(groupID), url.PathEscape(fileID)), body)
}

func (s *GroupFolderService) PatchLink(groupID, fileID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/groups/%s/folder/files/%s/link", url.PathEscape(groupID), url.PathEscape(fileID)), body)
}

func (s *GroupFolderService) DeleteLink(groupID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/folder/files/%s/link", url.PathEscape(groupID), url.PathEscape(fileID)))
}

func (s *GroupFolderService) ListPermissions(groupID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder/files/%s/permissions", url.PathEscape(groupID), url.PathEscape(fileID)))
}

func (s *GroupFolderService) CreatePermission(groupID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/folder/files/%s/permissions", url.PathEscape(groupID), url.PathEscape(fileID)), body)
}

func (s *GroupFolderService) GetPermission(groupID, fileID, permissionID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/folder/files/%s/permissions/%s", url.PathEscape(groupID), url.PathEscape(fileID), url.PathEscape(permissionID)))
}

func (s *GroupFolderService) PatchPermission(groupID, fileID, permissionID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/groups/%s/folder/files/%s/permissions/%s", url.PathEscape(groupID), url.PathEscape(fileID), url.PathEscape(permissionID)), body)
}

func (s *GroupFolderService) DeletePermission(groupID, fileID, permissionID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/folder/files/%s/permissions/%s", url.PathEscape(groupID), url.PathEscape(fileID), url.PathEscape(permissionID)))
}

func (s *GroupFolderService) DeleteAllPermissions(groupID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/folder/files/%s/permissions", url.PathEscape(groupID), url.PathEscape(fileID)))
}
