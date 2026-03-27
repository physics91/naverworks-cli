package api

import (
	"fmt"
	"net/url"
)

type BoardService struct {
	client *Client
}

func NewBoardService(client *Client) *BoardService {
	return &BoardService{client: client}
}

func (s *BoardService) ListBoards(cursor string, count int) (*Response, error) {
	return s.client.Get("/boards" + BuildPaginationQuery(cursor, count))
}

func (s *BoardService) GetBoard(boardID string) (*Response, error) {
	return s.client.Get("/boards/" + url.PathEscape(boardID))
}

func (s *BoardService) ListPosts(boardID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/boards/%s/posts", url.PathEscape(boardID)) + BuildPaginationQuery(cursor, count))
}

func (s *BoardService) GetPost(boardID, postID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/boards/%s/posts/%s", url.PathEscape(boardID), url.PathEscape(postID)))
}

func (s *BoardService) CreatePost(boardID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/boards/%s/posts", url.PathEscape(boardID)), body)
}

func (s *BoardService) UpdatePost(boardID, postID string, body map[string]interface{}) (*Response, error) {
	return s.client.PutJSON(fmt.Sprintf("/boards/%s/posts/%s", url.PathEscape(boardID), url.PathEscape(postID)), body)
}

func (s *BoardService) DeletePost(boardID, postID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/boards/%s/posts/%s", url.PathEscape(boardID), url.PathEscape(postID)))
}

func (s *BoardService) ListComments(boardID, postID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/boards/%s/posts/%s/comments", url.PathEscape(boardID), url.PathEscape(postID)) + BuildPaginationQuery(cursor, count))
}
