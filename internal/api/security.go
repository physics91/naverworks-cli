package api

type SecurityService struct {
	client *Client
}

func NewSecurityService(client *Client) *SecurityService {
	return &SecurityService{client: client}
}

func (s *SecurityService) GetExternalBrowser() (*Response, error) {
	return s.client.Get("/security/external-browser")
}

func (s *SecurityService) EnableExternalBrowser() (*Response, error) {
	return s.client.PostJSON("/security/external-browser/enable", nil)
}

func (s *SecurityService) DisableExternalBrowser() (*Response, error) {
	return s.client.PostJSON("/security/external-browser/disable", nil)
}
