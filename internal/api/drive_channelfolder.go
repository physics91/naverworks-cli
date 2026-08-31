package api

import (
	"fmt"
	"net/url"
)

// ChannelFolderService provides the Drive APIs for channel folders.
type ChannelFolderService struct {
	client *Client
}

func NewChannelFolderService(client *Client) *ChannelFolderService {
	return &ChannelFolderService{client: client}
}

func (s *ChannelFolderService) ListChannelFolders(cursor string, count int) (*Response, error) {
	return s.client.Get("/channel-folders" + BuildPaginationQuery(cursor, count))
}

func (s *ChannelFolderService) GetChannelFolder(channelFolderID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s", url.PathEscape(channelFolderID)))
}

func (s *ChannelFolderService) CreateRootUploadURL(channelFolderID string, body map[string]interface{}, fileSize int64) (*Response, error) {
	body["fileSize"] = fileSize
	return s.client.PostJSON(fmt.Sprintf("/channel-folders/%s/files", url.PathEscape(channelFolderID)), body)
}

func (s *ChannelFolderService) ListFiles(channelFolderID, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s/files", url.PathEscape(channelFolderID)) + BuildPaginationQuery(cursor, count))
}

func (s *ChannelFolderService) CreateFolderInRoot(channelFolderID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/createfolder", url.PathEscape(channelFolderID)), body)
}

func (s *ChannelFolderService) ListFolderChildren(channelFolderID, fileID, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s/files/%s/children", url.PathEscape(channelFolderID), url.PathEscape(fileID)) + BuildPaginationQuery(cursor, count))
}

func (s *ChannelFolderService) GetFile(channelFolderID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s/files/%s", url.PathEscape(channelFolderID), url.PathEscape(fileID)))
}

func (s *ChannelFolderService) DeleteFile(channelFolderID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/channel-folders/%s/files/%s", url.PathEscape(channelFolderID), url.PathEscape(fileID)))
}

func (s *ChannelFolderService) CreateSubFolder(channelFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/createfolder", url.PathEscape(channelFolderID), url.PathEscape(fileID)), body)
}

func (s *ChannelFolderService) CopyFile(channelFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/copy", url.PathEscape(channelFolderID), url.PathEscape(fileID)), body)
}

func (s *ChannelFolderService) RenameFile(channelFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/rename", url.PathEscape(channelFolderID), url.PathEscape(fileID)), body)
}

func (s *ChannelFolderService) MoveFile(channelFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/move", url.PathEscape(channelFolderID), url.PathEscape(fileID)), body)
}

func (s *ChannelFolderService) ProtectFile(channelFolderID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/protect", url.PathEscape(channelFolderID), url.PathEscape(fileID)), nil)
}

func (s *ChannelFolderService) UnprotectFile(channelFolderID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/unprotect", url.PathEscape(channelFolderID), url.PathEscape(fileID)), nil)
}

func (s *ChannelFolderService) LockFile(channelFolderID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/lock", url.PathEscape(channelFolderID), url.PathEscape(fileID)), nil)
}

func (s *ChannelFolderService) UnlockFile(channelFolderID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/unlock", url.PathEscape(channelFolderID), url.PathEscape(fileID)), nil)
}

func (s *ChannelFolderService) CreateUploadURL(channelFolderID, fileID string, body map[string]interface{}, fileSize int64) (*Response, error) {
	body["fileSize"] = fileSize
	return s.client.PostJSON(fmt.Sprintf("/channel-folders/%s/files/%s", url.PathEscape(channelFolderID), url.PathEscape(fileID)), body)
}

func (s *ChannelFolderService) GetDownloadURL(channelFolderID, fileID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/channel-folders/%s/files/%s/download", url.PathEscape(channelFolderID), url.PathEscape(fileID)))
}

func (s *ChannelFolderService) CreatePermission(channelFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/permissions", url.PathEscape(channelFolderID), url.PathEscape(fileID)), body)
}

func (s *ChannelFolderService) ListPermissions(channelFolderID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s/files/%s/permissions", url.PathEscape(channelFolderID), url.PathEscape(fileID)))
}

func (s *ChannelFolderService) GetPermission(channelFolderID, fileID, permissionID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s/files/%s/permissions/%s", url.PathEscape(channelFolderID), url.PathEscape(fileID), url.PathEscape(permissionID)))
}

func (s *ChannelFolderService) PatchPermission(channelFolderID, fileID, permissionID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/channel-folders/%s/files/%s/permissions/%s", url.PathEscape(channelFolderID), url.PathEscape(fileID), url.PathEscape(permissionID)), body)
}

func (s *ChannelFolderService) DeletePermission(channelFolderID, fileID, permissionID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/channel-folders/%s/files/%s/permissions/%s", url.PathEscape(channelFolderID), url.PathEscape(fileID), url.PathEscape(permissionID)))
}

func (s *ChannelFolderService) DeleteAllPermissions(channelFolderID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/channel-folders/%s/files/%s/permissions", url.PathEscape(channelFolderID), url.PathEscape(fileID)))
}

func (s *ChannelFolderService) EnablePermissions(channelFolderID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/permissions/enable", url.PathEscape(channelFolderID), url.PathEscape(fileID)), nil)
}

func (s *ChannelFolderService) DisablePermissions(channelFolderID, fileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/permissions/disable", url.PathEscape(channelFolderID), url.PathEscape(fileID)), nil)
}

func (s *ChannelFolderService) ListRevisions(channelFolderID, fileID, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s/files/%s/revisions", url.PathEscape(channelFolderID), url.PathEscape(fileID)) + BuildPaginationQuery(cursor, count))
}

func (s *ChannelFolderService) GetRevision(channelFolderID, fileID, revisionID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s/files/%s/revisions/%s", url.PathEscape(channelFolderID), url.PathEscape(fileID), url.PathEscape(revisionID)))
}

func (s *ChannelFolderService) RestoreRevision(channelFolderID, fileID, revisionID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/revisions/%s/restore", url.PathEscape(channelFolderID), url.PathEscape(fileID), url.PathEscape(revisionID)), nil)
}

func (s *ChannelFolderService) GetRevisionDownloadURL(channelFolderID, fileID, revisionID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/channel-folders/%s/files/%s/revisions/%s/download", url.PathEscape(channelFolderID), url.PathEscape(fileID), url.PathEscape(revisionID)))
}

func (s *ChannelFolderService) GetLinkSetting(channelFolderID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s/link-setting", url.PathEscape(channelFolderID)))
}

func (s *ChannelFolderService) GetLink(channelFolderID, fileID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s/files/%s/link", url.PathEscape(channelFolderID), url.PathEscape(fileID)))
}

func (s *ChannelFolderService) CreateLink(channelFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/files/%s/link", url.PathEscape(channelFolderID), url.PathEscape(fileID)), body)
}

func (s *ChannelFolderService) PatchLink(channelFolderID, fileID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/channel-folders/%s/files/%s/link", url.PathEscape(channelFolderID), url.PathEscape(fileID)), body)
}

func (s *ChannelFolderService) DeleteLink(channelFolderID, fileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/channel-folders/%s/files/%s/link", url.PathEscape(channelFolderID), url.PathEscape(fileID)))
}

func (s *ChannelFolderService) ListTrashFiles(channelFolderID, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/channel-folders/%s/trash-files", url.PathEscape(channelFolderID)) + BuildPaginationQuery(cursor, count))
}

func (s *ChannelFolderService) RestoreTrashFile(channelFolderID, trashFileID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/channel-folders/%s/trash-files/%s/restore", url.PathEscape(channelFolderID), url.PathEscape(trashFileID)), nil)
}

func (s *ChannelFolderService) DeleteTrashFile(channelFolderID, trashFileID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/channel-folders/%s/trash-files/%s", url.PathEscape(channelFolderID), url.PathEscape(trashFileID)))
}
