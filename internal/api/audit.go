package api

import (
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
	return s.client.GetDownloadURL("/audits/logs/download" + encodeQueryFromValues(params))
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
	return s.client.GetDownloadURL("/monitoring/message-contents/download" + encodeQueryFromValues(params))
}
