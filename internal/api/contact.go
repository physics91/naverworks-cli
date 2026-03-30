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

func (s *ContactService) FullUpdateContact(contactID string, body map[string]interface{}) (*Response, error) {
	return s.client.PutJSON("/contacts/"+url.PathEscape(contactID), body)
}

func (s *ContactService) ForceDeleteContact(contactID string) (*Response, error) {
	return s.client.Delete("/contacts/" + url.PathEscape(contactID) + "/forcedelete")
}

// CreatePhoto requests a presigned upload URL for a contact photo.
// body should contain fileName and fileSize.
func (s *ContactService) CreatePhoto(contactID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/contacts/%s/photo", url.PathEscape(contactID)), body)
}

func (s *ContactService) GetPhoto(contactID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/contacts/%s/photo", url.PathEscape(contactID)))
}

func (s *ContactService) DeletePhoto(contactID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/contacts/%s/photo", url.PathEscape(contactID)))
}

// ─── Custom Properties ───

func (s *ContactService) CreateCustomProperty(body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON("/contacts/custom-properties", body)
}

func (s *ContactService) ListCustomProperties(cursor string, count int) (*Response, error) {
	return s.client.Get("/contacts/custom-properties" + BuildPaginationQuery(cursor, count))
}

func (s *ContactService) GetCustomProperty(id string) (*Response, error) {
	return s.client.Get("/contacts/custom-properties/" + url.PathEscape(id))
}

func (s *ContactService) PatchCustomProperty(id string, body map[string]interface{}) (*Response, error) {
	return s.client.PatchJSON("/contacts/custom-properties/"+url.PathEscape(id), body)
}

func (s *ContactService) DeleteCustomProperty(id string) (*Response, error) {
	return s.client.Delete("/contacts/custom-properties/" + url.PathEscape(id))
}

// ─── Tags ───

func (s *ContactService) CreateTag(body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON("/contact-tags", body)
}

func (s *ContactService) GetTag(tagID string) (*Response, error) {
	return s.client.Get("/contact-tags/" + url.PathEscape(tagID))
}

func (s *ContactService) UpdateTag(tagID string, body map[string]interface{}) (*Response, error) {
	return s.client.PutJSON("/contact-tags/"+url.PathEscape(tagID), body)
}

func (s *ContactService) PatchTag(tagID string, body map[string]interface{}) (*Response, error) {
	return s.client.PatchJSON("/contact-tags/"+url.PathEscape(tagID), body)
}

func (s *ContactService) DeleteTag(tagID string) (*Response, error) {
	return s.client.Delete("/contact-tags/" + url.PathEscape(tagID))
}

func (s *ContactService) CreateUserTags(userID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/users/%s/contact-tags", url.PathEscape(userID)), body)
}
