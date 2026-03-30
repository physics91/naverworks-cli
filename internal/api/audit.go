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
	return s.client.GetDownloadURL("/audits/logs/download" + encodeQueryFromValues(params))
}

func (s *AuditService) ListPolicyGroups(cursor string, count int) (*Response, error) {
	return s.client.Get("/audits/policy-groups" + BuildPaginationQuery(cursor, count))
}

func (s *AuditService) CreatePolicyGroup(body []byte) (*Response, error) {
	return s.client.Post("/audits/policy-groups", body)
}

func (s *AuditService) GetPolicyGroup(policyGroupID string) (*Response, error) {
	return s.client.Get("/audits/policy-groups/" + url.PathEscape(policyGroupID))
}

func (s *AuditService) UpdatePolicyGroup(policyGroupID string, body []byte) (*Response, error) {
	return s.client.Put("/audits/policy-groups/"+url.PathEscape(policyGroupID), body)
}

func (s *AuditService) DeletePolicyGroup(policyGroupID string) (*Response, error) {
	return s.client.Delete("/audits/policy-groups/" + url.PathEscape(policyGroupID))
}

func (s *AuditService) AddPolicyGroupMembers(policyGroupID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/audits/policy-groups/%s/members", url.PathEscape(policyGroupID)), body)
}

func (s *AuditService) ListPolicyGroupMembers(policyGroupID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/audits/policy-groups/%s/members", url.PathEscape(policyGroupID)) + BuildPaginationQuery(cursor, count))
}

func (s *AuditService) RemovePolicyGroupMember(policyGroupID, userID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/audits/policy-groups/%s/members/%s", url.PathEscape(policyGroupID), url.PathEscape(userID)))
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
