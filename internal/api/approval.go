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
