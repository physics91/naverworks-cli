package api

import (
	"fmt"
	"net/url"
)

type MailService struct {
	client *Client
}

func NewMailService(client *Client) *MailService {
	return &MailService{client: client}
}

func (s *MailService) SendMail(userID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/users/%s/mail", url.PathEscape(userID)), body)
}

func (s *MailService) GetMail(userID, mailID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/mail/%s", url.PathEscape(userID), url.PathEscape(mailID)))
}

func (s *MailService) DeleteMail(userID, mailID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/mail/%s", url.PathEscape(userID), url.PathEscape(mailID)))
}

func (s *MailService) ListFolders(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/mail/mailfolders", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *MailService) GetFolder(userID, folderID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/mail/mailfolders/%s", url.PathEscape(userID), url.PathEscape(folderID)))
}

func (s *MailService) ListMails(userID, folderID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/mail/mailfolders/%s/children", url.PathEscape(userID), url.PathEscape(folderID)) + BuildPaginationQuery(cursor, count))
}

func (s *MailService) PatchMail(userID, mailID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/users/%s/mail/%s", url.PathEscape(userID), url.PathEscape(mailID)), body)
}

func (s *MailService) GetUnreadCount(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/mail/unread-count", url.PathEscape(userID)))
}

func (s *MailService) GetAttachment(userID, mailID, attachmentID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/mail/%s/attachments/%s", url.PathEscape(userID), url.PathEscape(mailID), url.PathEscape(attachmentID)))
}

func (s *MailService) ListFavoriteContactsFolders(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/mail/mailfolders/favorite-contacts", url.PathEscape(userID)))
}

func (s *MailService) CreateMailFolder(userID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/users/%s/mail/mailfolders", url.PathEscape(userID)), body)
}

func (s *MailService) UpdateMailFolder(userID, folderID string, body []byte) (*Response, error) {
	return s.client.Put(fmt.Sprintf("/users/%s/mail/mailfolders/%s", url.PathEscape(userID), url.PathEscape(folderID)), body)
}

func (s *MailService) DeleteMailFolder(userID, folderID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/mail/mailfolders/%s", url.PathEscape(userID), url.PathEscape(folderID)))
}

func (s *MailService) CreateFilter(userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/mail/filters", url.PathEscape(userID)), body)
}

func (s *MailService) ListFilters(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/mail/filters", url.PathEscape(userID)))
}

func (s *MailService) GetFilter(userID, filterID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/mail/filters/%s", url.PathEscape(userID), url.PathEscape(filterID)))
}

func (s *MailService) DeleteFilter(userID, filterID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/mail/filters/%s", url.PathEscape(userID), url.PathEscape(filterID)))
}

func (s *MailService) CreateImapMigration(userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/mail/migration/imap", url.PathEscape(userID)), body)
}

func (s *MailService) GetImapMigration(userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/mail/migration/imap", url.PathEscape(userID)))
}

func (s *MailService) DeleteImapMigration(userID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/mail/migration/imap", url.PathEscape(userID)))
}

func (s *MailService) CreatePop3Migration(userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/mail/migration/pop3", url.PathEscape(userID)), body)
}

func (s *MailService) CreateForwarding(userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/users/%s/mail/settings/forwarding", url.PathEscape(userID)), body)
}

func (s *MailService) DeleteForwarding(userID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/users/%s/mail/settings/forwarding", url.PathEscape(userID)))
}
