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

// ─── Board CRUD ───

func (s *BoardService) CreateBoard(body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON("/boards", body)
}

func (s *BoardService) UpdateBoard(boardID string, body map[string]interface{}) (*Response, error) {
	return s.client.PutJSON("/boards/"+url.PathEscape(boardID), body)
}

func (s *BoardService) DeleteBoard(boardID string) (*Response, error) {
	return s.client.Delete("/boards/" + url.PathEscape(boardID))
}

// ─── Post Readers ───

func (s *BoardService) ListPostReaders(boardID, postID string, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/boards/%s/posts/%s/readers", url.PathEscape(boardID), url.PathEscape(postID)) + BuildPaginationQuery(cursor, count))
}

// ─── Aggregation Queries ───

func (s *BoardService) ListRecentPosts(cursor string, count int) (*Response, error) {
	return s.client.Get("/boards/recent/posts" + BuildPaginationQuery(cursor, count))
}

func (s *BoardService) ListMyPosts(cursor string, count int) (*Response, error) {
	return s.client.Get("/boards/my/posts" + BuildPaginationQuery(cursor, count))
}

func (s *BoardService) ListMustPosts(cursor string, count int) (*Response, error) {
	return s.client.Get("/boards/must/posts" + BuildPaginationQuery(cursor, count))
}

// ─── Post Attachments ───

func (s *BoardService) CreatePostAttachment(boardID, postID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/boards/%s/posts/%s/attachments", url.PathEscape(boardID), url.PathEscape(postID)), body)
}

func (s *BoardService) ListPostAttachments(boardID, postID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/boards/%s/posts/%s/attachments", url.PathEscape(boardID), url.PathEscape(postID)))
}

func (s *BoardService) GetPostAttachment(boardID, postID, attachmentID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/boards/%s/posts/%s/attachments/%s", url.PathEscape(boardID), url.PathEscape(postID), url.PathEscape(attachmentID)))
}

func (s *BoardService) DeletePostAttachment(boardID, postID, attachmentID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/boards/%s/posts/%s/attachments/%s", url.PathEscape(boardID), url.PathEscape(postID), url.PathEscape(attachmentID)))
}

// ─── Comments CRUD ───

func (s *BoardService) CreateComment(boardID, postID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/boards/%s/posts/%s/comments", url.PathEscape(boardID), url.PathEscape(postID)), body)
}

func (s *BoardService) GetComment(boardID, postID, commentID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/boards/%s/posts/%s/comments/%s", url.PathEscape(boardID), url.PathEscape(postID), url.PathEscape(commentID)))
}

func (s *BoardService) UpdateComment(boardID, postID, commentID string, body map[string]interface{}) (*Response, error) {
	return s.client.PutJSON(fmt.Sprintf("/boards/%s/posts/%s/comments/%s", url.PathEscape(boardID), url.PathEscape(postID), url.PathEscape(commentID)), body)
}

func (s *BoardService) DeleteComment(boardID, postID, commentID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/boards/%s/posts/%s/comments/%s", url.PathEscape(boardID), url.PathEscape(postID), url.PathEscape(commentID)))
}

// ─── Comment Attachments ───

func (s *BoardService) CreateCommentAttachment(boardID, postID, commentID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/boards/%s/posts/%s/comments/%s/attachments", url.PathEscape(boardID), url.PathEscape(postID), url.PathEscape(commentID)), body)
}

func (s *BoardService) ListCommentAttachments(boardID, postID, commentID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/boards/%s/posts/%s/comments/%s/attachments", url.PathEscape(boardID), url.PathEscape(postID), url.PathEscape(commentID)))
}

func (s *BoardService) GetCommentAttachment(boardID, postID, commentID, attachmentID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/boards/%s/posts/%s/comments/%s/attachments/%s", url.PathEscape(boardID), url.PathEscape(postID), url.PathEscape(commentID), url.PathEscape(attachmentID)))
}

func (s *BoardService) DeleteCommentAttachment(boardID, postID, commentID, attachmentID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/boards/%s/posts/%s/comments/%s/attachments/%s", url.PathEscape(boardID), url.PathEscape(postID), url.PathEscape(commentID), url.PathEscape(attachmentID)))
}
