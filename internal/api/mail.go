package api

import (
	"encoding/json"
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
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("메일 직렬화 실패: %w", err)
	}
	return s.client.Post(fmt.Sprintf("/users/%s/mail", url.PathEscape(userID)), data)
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
