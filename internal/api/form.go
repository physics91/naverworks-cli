package api

import (
	"fmt"
	"net/url"
)

type FormService struct {
	client *Client
}

func NewFormService(client *Client) *FormService {
	return &FormService{client: client}
}

func (s *FormService) ListResponses(formID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/forms/%s/responses", url.PathEscape(formID)) + BuildPaginationQuery(cursor, count))
}

func (s *FormService) DownloadAttachment(formID, responseID, attachmentID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/forms/%s/responses/%s/attachments/%s",
		url.PathEscape(formID), url.PathEscape(responseID), url.PathEscape(attachmentID)))
}
