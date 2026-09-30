package web

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"unicode/utf8"
)

func (c *Client) preparePost(id string) (string, string, error) {
	q, err := c.noteQuery()
	if err != nil {
		return "", "", err
	}
	q.Set("postNo", id)
	body, err := c.get(talkOrigin, noteBase+"post/preparePostWrite", q)
	if err != nil {
		return "", "", err
	}
	var result struct{ NextPostNo, LastModifyUtcYmdt json.Number }
	if err := decodeNote(body, &result); err != nil {
		return "", "", err
	}
	if !positiveNumber(result.NextPostNo.String()) {
		return "", "", fmt.Errorf("노트 작성 준비 결과가 유효하지 않습니다")
	}
	return result.NextPostNo.String(), result.LastModifyUtcYmdt.String(), nil
}

func noteWriteValues(id, title, body, modified string) url.Values {
	v := url.Values{"postNo": {id}, "title": {title}, "body": {body}, "boardNo": {"0"}, "isCoedit": {"true"}, "isNoti": {"false"}, "notify": {"false"}}
	if modified != "" {
		v.Set("lastModifyUtcYmdt", modified)
	}
	return v
}

func (c *Client) CreatePost(title, body string) ([]byte, error) {
	if err := validatePostInput(title, body); err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(title) > 200 {
		return nil, fmt.Errorf("노트 제목은 200자 이하여야 합니다")
	}
	id, modified, err := c.preparePost("0")
	if err != nil {
		return nil, err
	}
	if _, err := c.postNote("add", noteWriteValues(id, title, body, modified)); err != nil {
		return nil, err
	}
	post, err := c.GetPost(id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(post)
}

func (c *Client) UpdatePost(id, title, body string) ([]byte, error) {
	if err := validatePostInput(title, body); err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(title) > 200 {
		return nil, fmt.Errorf("노트 제목은 200자 이하여야 합니다")
	}
	post, err := c.fetchPost(id)
	if err != nil {
		return nil, err
	}
	if post.ReadOnly {
		return nil, fmt.Errorf("이 노트를 수정할 권한이 없습니다")
	}
	preparedID, modified, err := c.preparePost(id)
	if err != nil {
		return nil, err
	}
	expectedModified := post.ModifyUtcYmdt.String()
	if expectedModified == "" || expectedModified == "0" {
		expectedModified = post.RegistUtcYmdt.String()
	}
	if preparedID != id || modified != expectedModified {
		return nil, fmt.Errorf("노트가 변경됐습니다. 다시 조회한 뒤 수정하세요")
	}
	v := noteWriteValues(id, title, body, modified)
	if post.BoardNo != "" {
		v.Set("boardNo", post.BoardNo.String())
	}
	v.Set("isNoti", strconv.FormatBool(post.NotiYn == "Y"))
	v.Set("isCoedit", strconv.FormatBool(post.CoeditYn == "Y"))
	if len(post.FileList) > 0 {
		attached, err := json.Marshal(map[string]any{"fileList": post.FileList})
		if err != nil {
			return nil, err
		}
		v.Set("attachedJson", string(attached))
	}
	if _, err := c.postNote("modify", v); err != nil {
		return nil, err
	}
	updated, err := c.GetPost(id)
	if err != nil {
		return nil, err
	}
	return json.Marshal(updated)
}

func (c *Client) DeletePost(id string) ([]byte, error) {
	if _, err := c.GetPost(id); err != nil {
		return nil, err
	}
	if _, err := c.postNote("remove", url.Values{"postNo": {id}}); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{"postId": id, "deleted": true})
}
