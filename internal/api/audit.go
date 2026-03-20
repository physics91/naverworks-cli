package api

import (
	"fmt"
	"net/url"
)

type AuditService struct {
	client *Client
}

func NewAuditService(client *Client) *AuditService {
	return &AuditService{client: client}
}

func (s *AuditService) DownloadLogs(from, until string) (*Response, error) {
	params := url.Values{}
	if from != "" {
		params.Set("from", from)
	}
	if until != "" {
		params.Set("until", until)
	}
	query := ""
	if len(params) > 0 {
		query = "?" + params.Encode()
	}
	return s.client.Get("/audits/logs/download" + query)
}

func (s *AuditService) ListPolicyGroups(cursor string, count int) (*Response, error) {
	return s.client.Get("/audits/policy-groups" + BuildPaginationQuery(cursor, count))
}

type MonitoringService struct {
	client *Client
}

func NewMonitoringService(client *Client) *MonitoringService {
	return &MonitoringService{client: client}
}

func (s *MonitoringService) DownloadMessages(from, until string) (*Response, error) {
	params := url.Values{}
	if from != "" {
		params.Set("from", from)
	}
	if until != "" {
		params.Set("until", until)
	}
	query := ""
	if len(params) > 0 {
		query = "?" + params.Encode()
	}
	return s.client.Get(fmt.Sprintf("/monitoring/message-contents/download%s", query))
}
