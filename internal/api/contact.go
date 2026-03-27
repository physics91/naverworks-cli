package api

import (
	"fmt"
	"net/url"
)

type ContactService struct {
	client *Client
}

func NewContactService(client *Client) *ContactService {
	return &ContactService{client: client}
}

func (s *ContactService) ListContacts(cursor string, count int) (*Response, error) {
	return s.client.Get("/contacts" + BuildPaginationQuery(cursor, count))
}

func (s *ContactService) ListUserContacts(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/contacts", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *ContactService) GetContact(contactID string) (*Response, error) {
	return s.client.Get("/contacts/" + url.PathEscape(contactID))
}

func (s *ContactService) CreateContact(body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON("/contacts", body)
}

func (s *ContactService) UpdateContact(contactID string, body map[string]interface{}) (*Response, error) {
	return s.client.PatchJSON("/contacts/"+url.PathEscape(contactID), body)
}

func (s *ContactService) DeleteContact(contactID string) (*Response, error) {
	return s.client.Delete("/contacts/" + url.PathEscape(contactID))
}

func (s *ContactService) ListTags(cursor string, count int) (*Response, error) {
	return s.client.Get("/contact-tags" + BuildPaginationQuery(cursor, count))
}

func (s *ContactService) ListUserTags(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/users/%s/contact-tags", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}
