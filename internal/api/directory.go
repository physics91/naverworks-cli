package api

type DirectoryService struct {
	client *Client
}

func NewDirectoryService(client *Client) *DirectoryService {
	return &DirectoryService{client: client}
}

func (s *DirectoryService) ListUsers(cursor string, count int) (*Response, error) {
	return s.client.Get("/users" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetUser(userID string) (*Response, error) {
	return s.client.Get("/users/" + userID)
}

func (s *DirectoryService) ListGroups(cursor string, count int) (*Response, error) {
	return s.client.Get("/groups" + BuildPaginationQuery(cursor, count))
}

func (s *DirectoryService) GetGroup(groupID string) (*Response, error) {
	return s.client.Get("/groups/" + groupID)
}
