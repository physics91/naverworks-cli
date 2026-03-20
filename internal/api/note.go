package api

import (
	"encoding/json"
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
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("노트 게시글 직렬화 실패: %w", err)
	}
	return s.client.Post(fmt.Sprintf("/groups/%s/note/posts", url.PathEscape(groupID)), data)
}

func (s *NoteService) UpdatePost(groupID, postID string, body map[string]interface{}) (*Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("노트 게시글 직렬화 실패: %w", err)
	}
	return s.client.Put(fmt.Sprintf("/groups/%s/note/posts/%s", url.PathEscape(groupID), url.PathEscape(postID)), data)
}

func (s *NoteService) DeletePost(groupID, postID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/groups/%s/note/posts/%s", url.PathEscape(groupID), url.PathEscape(postID)))
}
