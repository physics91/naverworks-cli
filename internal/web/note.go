package web

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const talkOrigin = "https://talk.worksmobile.com"
const noteBase = "/p/note/open-api/teamNote/"

type Post struct {
	ID        string      `json:"postId"`
	Title     string      `json:"title"`
	Body      string      `json:"body,omitempty"`
	PlainText string      `json:"plainTextBody,omitempty"`
	BoardNo   json.Number `json:"boardNo,omitempty"`
	Modified  json.Number `json:"modifyUtcYmdt,omitempty"`
	ReadOnly  bool        `json:"readOnly"`
}

type notePost struct {
	PostNoString               string
	PostNo                     json.Number
	Title, Body, PlainTextBody string
	BoardNo, ModifyUtcYmdt     json.Number
	RegistUtcYmdt              json.Number
	ReadOnly                   bool
	NotiYn, CoeditYn           string
	FileList                   []noteFile
}

type noteFile struct {
	AttachType string      `json:"attachType"`
	FileName   string      `json:"fileName"`
	FileSize   json.Number `json:"fileSize"`
	FileURL    string      `json:"fileUrl"`
}

func (p notePost) normalized() (Post, error) {
	id := p.PostNoString
	if id == "" {
		id = p.PostNo.String()
	}
	if !positiveNumber(id) {
		return Post{}, fmt.Errorf("노트 게시글 ID가 유효하지 않습니다")
	}
	return Post{ID: id, Title: p.Title, Body: p.Body, PlainText: p.PlainTextBody, BoardNo: p.BoardNo, Modified: p.ModifyUtcYmdt, ReadOnly: p.ReadOnly}, nil
}

func (c *Client) noteQuery() (url.Values, error) {
	q := url.Values{"apiVer": {"13"}, "channelNo": {c.room.ChannelNo}, "tenantId": {c.room.TenantID}, "collaboSpaceId": {c.room.CollaboSpaceID}}
	if c.room.Resource == "" {
		body, err := c.get(talkOrigin, noteBase+"channel/location", q)
		if err != nil {
			return nil, err
		}
		var location struct {
			RL json.Number `json:"rl"`
		}
		if err := decodeNote(body, &location); err != nil {
			return nil, err
		}
		if !positiveNumber(location.RL.String()) {
			return nil, fmt.Errorf("노트 저장 위치가 유효하지 않습니다")
		}
		c.room.Resource = location.RL.String()
	}
	q.Set("rl", c.room.Resource)
	return q, nil
}

func decodeNote(body []byte, out any) error {
	var envelope struct {
		ReturnCode    string
		ReturnMessage string
		Data          json.RawMessage
	}
	if err := responseJSON(body, &envelope); err != nil {
		return err
	}
	if envelope.ReturnCode != "0" || envelope.ReturnMessage != "SUCCESS" {
		code, err := strconv.Atoi(envelope.ReturnCode)
		if err != nil {
			return fmt.Errorf("노트 응답 상태 형식이 변경됐습니다")
		}
		return fmt.Errorf("노트 요청이 거부됐습니다 (returnCode %d)", code)
	}
	if out == nil {
		return nil
	}
	return responseJSON(envelope.Data, out)
}

func pageSize(count int) (int, error) {
	if count == 0 {
		return 20, nil
	}
	if count < 1 || count > 100 {
		return 0, fmt.Errorf("웹 조회 --count는 1~100 범위여야 합니다")
	}
	return count, nil
}

func (c *Client) ListPosts(cursor string, count int) ([]byte, error) {
	count, err := pageSize(count)
	if err != nil {
		return nil, err
	}
	index := 1
	if cursor != "" {
		index, err = strconv.Atoi(cursor)
		if err != nil || index < 1 {
			return nil, fmt.Errorf("노트 --cursor는 1 이상의 시작 위치여야 합니다")
		}
	}
	q, err := c.noteQuery()
	if err != nil {
		return nil, err
	}
	q.Set("startIndex", strconv.Itoa(index))
	q.Set("selectCount", strconv.Itoa(count))
	q.Set("selectType", "all")
	q.Set("sortBy", "newest")
	q.Set("bodyMaxLength", "200")
	body, err := c.get(talkOrigin, noteBase+"post/selectList", q)
	if err != nil {
		return nil, err
	}
	var result struct{ PostList []notePost }
	if err := decodeNote(body, &result); err != nil {
		return nil, err
	}
	posts := make([]Post, 0, len(result.PostList))
	for _, p := range result.PostList {
		post, err := p.normalized()
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	next := ""
	if len(posts) == count {
		next = strconv.Itoa(index + count)
	}
	return json.Marshal(map[string]any{"posts": posts, "responseMetaData": map[string]string{"nextCursor": next}})
}

func (c *Client) GetPost(id string) (Post, error) {
	post, err := c.fetchPost(id)
	if err != nil {
		return Post{}, err
	}
	return post.normalized()
}

func (c *Client) fetchPost(id string) (notePost, error) {
	if !positiveNumber(id) {
		return notePost{}, fmt.Errorf("노트 postId는 양의 정수 문자열이어야 합니다")
	}
	q, err := c.noteQuery()
	if err != nil {
		return notePost{}, err
	}
	q.Set("postNo", id)
	body, err := c.get(talkOrigin, noteBase+"post/select", q)
	if err != nil {
		return notePost{}, err
	}
	var result struct{ Post notePost }
	if err := decodeNote(body, &result); err != nil {
		return notePost{}, err
	}
	post, err := result.Post.normalized()
	if err != nil {
		return notePost{}, err
	}
	if post.ID != id {
		return notePost{}, fmt.Errorf("조회한 노트 ID가 요청과 일치하지 않습니다")
	}
	return result.Post, nil
}

func (c *Client) postNote(action string, values url.Values) ([]byte, error) {
	q, err := c.noteQuery()
	if err != nil {
		return nil, err
	}
	for k, v := range q {
		values[k] = v
	}
	body, err := c.request(talkOrigin, "POST", noteBase+"post/"+action, "application/x-www-form-urlencoded", values.Encode())
	if err != nil {
		return nil, err
	}
	var data json.RawMessage
	if err := decodeNote(body, &data); err != nil {
		return nil, err
	}
	if len(data) == 0 || string(data) == "null" {
		return []byte(`{}`), nil
	}
	return data, nil
}

func validatePostInput(title, body string) error {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(body) == "" {
		return fmt.Errorf("노트 --title과 --body는 비어 있을 수 없습니다")
	}
	return nil
}
