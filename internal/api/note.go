package api

import (
	"fmt"
	"net/url"
)

type NoteService struct {
	client *Client
}

func NewNoteService(client *Client) *NoteService {
	return &NoteService{client: client}
}

func (s *NoteService) CreateNote(groupID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/groups/%s/note", url.PathEscape(groupID)), nil)
}

func (s *NoteService) DeleteNote(groupID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/note", url.PathEscape(groupID)))
}

func (s *NoteService) ListPosts(groupID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/note/posts", url.PathEscape(groupID)) + BuildPaginationQuery(cursor, count))
}

func (s *NoteService) GetPost(groupID, postID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/groups/%s/note/posts/%s", url.PathEscape(groupID), url.PathEscape(postID)))
}

func (s *NoteService) CreatePost(groupID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/groups/%s/note/posts", url.PathEscape(groupID)), body)
}

func (s *NoteService) UpdatePost(groupID, postID string, body map[string]interface{}) (*Response, error) {
	return s.client.PutJSON(fmt.Sprintf("/groups/%s/note/posts/%s", url.PathEscape(groupID), url.PathEscape(postID)), body)
}

func (s *NoteService) DeletePost(groupID, postID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/note/posts/%s", url.PathEscape(groupID), url.PathEscape(postID)))
}
