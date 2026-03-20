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

func (s *AuditService) DownloadLogs(startTime, endTime, service string) (string, error) {
	params := url.Values{}
	if startTime != "" {
		params.Set("startTime", startTime)
	}
	if endTime != "" {
		params.Set("endTime", endTime)
	}
	if service != "" {
		params.Set("service", service)
	}
	query := ""
	if len(params) > 0 {
		query = "?" + params.Encode()
	}
	return s.client.GetDownloadURL("/audits/logs/download" + query)
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

func (s *MonitoringService) DownloadMessages(startTime, endTime string) (string, error) {
	params := url.Values{}
	if startTime != "" {
		params.Set("startTime", startTime)
	}
	if endTime != "" {
		params.Set("endTime", endTime)
	}
	query := ""
	if len(params) > 0 {
		query = "?" + params.Encode()
	}
	return s.client.GetDownloadURL(fmt.Sprintf("/monitoring/message-contents/download%s", query))
}
