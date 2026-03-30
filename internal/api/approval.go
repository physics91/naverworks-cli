package api

import (
	"fmt"
	"net/url"
)

type ApprovalService struct {
	client *Client
}

func NewApprovalService(client *Client) *ApprovalService {
	return &ApprovalService{client: client}
}

func (s *ApprovalService) ListUserDocuments(userID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/business-support/approval/users/%s/documents", url.PathEscape(userID)) + BuildPaginationQuery(cursor, count))
}

func (s *ApprovalService) ListDocuments(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/approval/documents" + BuildPaginationQuery(cursor, count))
}

func (s *ApprovalService) GetDocument(documentID string) (*Response, error) {
	return s.client.Get("/business-support/approval/documents/" + url.PathEscape(documentID))
}

func (s *ApprovalService) ListCategories(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/approval/categories" + BuildPaginationQuery(cursor, count))
}

func (s *ApprovalService) GetCategory(categoryID string) (*Response, error) {
	return s.client.Get("/business-support/approval/categories/" + url.PathEscape(categoryID))
}

func (s *ApprovalService) ListDocumentForms(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/approval/document-forms" + BuildPaginationQuery(cursor, count))
}

func (s *ApprovalService) CreateCategory(body []byte) (*Response, error) {
	return s.client.Post("/business-support/approval/categories", body)
}

func (s *ApprovalService) PatchCategory(categoryID string, body []byte) (*Response, error) {
	return s.client.Patch("/business-support/approval/categories/"+url.PathEscape(categoryID), body)
}

func (s *ApprovalService) DeleteCategory(categoryID string) (*Response, error) {
	return s.client.Delete("/business-support/approval/categories/" + url.PathEscape(categoryID))
}

func (s *ApprovalService) CreateUserDocument(userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/business-support/approval/users/%s/documents", url.PathEscape(userID)), body)
}

func (s *ApprovalService) CreateImportedDocument(body []byte) (*Response, error) {
	return s.client.Post("/business-support/approval/imported-documents", body)
}

func (s *ApprovalService) CreateDocumentLink(userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/business-support/approval/users/%s/documents/create-document-link", url.PathEscape(userID)), body)
}

func (s *ApprovalService) GetDocumentForm(documentFormID string) (*Response, error) {
	return s.client.Get("/business-support/approval/document-forms/" + url.PathEscape(documentFormID))
}

func (s *ApprovalService) CreateUserDocumentAttachment(userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/business-support/approval/users/%s/documents/attachments", url.PathEscape(userID)), body)
}

func (s *ApprovalService) CreateImportedDocumentAttachment(body []byte) (*Response, error) {
	return s.client.Post("/business-support/approval/imported-documents/attachments", body)
}

func (s *ApprovalService) CreateLinkageCode(body []byte) (*Response, error) {
	return s.client.Post("/business-support/approval/linkage-codes", body)
}

func (s *ApprovalService) ListLinkageCodes(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/approval/linkage-codes" + BuildPaginationQuery(cursor, count))
}

func (s *ApprovalService) GetLinkageCode(key string) (*Response, error) {
	return s.client.Get("/business-support/approval/linkage-codes/" + url.PathEscape(key))
}

func (s *ApprovalService) PatchLinkageCode(key string, body []byte) (*Response, error) {
	return s.client.Patch("/business-support/approval/linkage-codes/"+url.PathEscape(key), body)
}

func (s *ApprovalService) CreateLinkageCodeItem(key string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/business-support/approval/linkage-codes/%s/linkage-code-items", url.PathEscape(key)), body)
}

func (s *ApprovalService) ListLinkageCodeItems(key string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/business-support/approval/linkage-codes/%s/linkage-code-items", url.PathEscape(key)) + BuildPaginationQuery(cursor, count))
}

func (s *ApprovalService) GetLinkageCodeItem(key, id string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/business-support/approval/linkage-codes/%s/linkage-code-items/%s", url.PathEscape(key), url.PathEscape(id)))
}

func (s *ApprovalService) PatchLinkageCodeItem(key, id string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/business-support/approval/linkage-codes/%s/linkage-code-items/%s", url.PathEscape(key), url.PathEscape(id)), body)
}

func (s *ApprovalService) DeleteLinkageCodeItem(key, id string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/business-support/approval/linkage-codes/%s/linkage-code-items/%s", url.PathEscape(key), url.PathEscape(id)))
}
