package web

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var roomIDPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

type Room struct {
	ID             string `json:"roomId"`
	ChannelNo      string `json:"channelNo"`
	TenantID       string `json:"tenantId"`
	CollaboSpaceID string `json:"collaboSpaceId"`
	Resource       string `json:"-"`
}

func ValidateRoomID(id string) error {
	if !roomIDPattern.MatchString(id) {
		return fmt.Errorf("--room-id에는 채널 UUID를 지정하세요 (웹 내부 방 번호와 다릅니다)")
	}
	return nil
}

// ResolveRoom matches the public channel UUID against the user's visible room
// list. A missing room must never fall back to the last room opened in the UI.
func (c *Client) ResolveRoom(id string) error {
	if err := ValidateRoomID(id); err != nil {
		return err
	}
	var cursor int64
	for page := 0; page < 100; page++ {
		body, err := c.postJSON(strings.TrimSuffix(EntryURL, "/"), "/p/oneapp/client/chat/getVisibleUserChannelList", map[string]any{"syncUpdateTime": cursor, "pagingCount": 100})
		if err != nil {
			return err
		}
		var result struct {
			Code               int
			NextSyncUpdateTime int64
			Result             []struct {
				ChannelNo     json.Number
				ChannelExtras string
			}
		}
		if err := responseJSON(body, &result); err != nil {
			return err
		}
		if result.Code != 200 {
			return fmt.Errorf("웹 메시지방 목록 조회가 거부됐습니다 (code %d)", result.Code)
		}
		for _, channel := range result.Result {
			var extras struct{ ChannelID string }
			if err := json.Unmarshal([]byte(channel.ChannelExtras), &extras); err != nil {
				return fmt.Errorf("웹 메시지방 식별 정보가 변경됐습니다")
			}
			if strings.EqualFold(extras.ChannelID, id) {
				return c.resolveExtras(id, channel.ChannelNo)
			}
		}
		if len(result.Result) < 100 {
			return fmt.Errorf("지정한 채널을 로그인 계정의 메시지방 목록에서 찾을 수 없습니다")
		}
		if result.NextSyncUpdateTime <= cursor {
			return fmt.Errorf("웹 메시지방 목록 커서가 진행되지 않습니다")
		}
		cursor = result.NextSyncUpdateTime
	}
	return fmt.Errorf("웹 메시지방 목록 페이지 제한을 초과했습니다")
}

func (c *Client) resolveExtras(id string, channelNo json.Number) error {
	no, err := strconv.ParseInt(channelNo.String(), 10, 64)
	if err != nil || no <= 0 {
		return fmt.Errorf("웹 내부 방 번호가 유효하지 않습니다")
	}
	data, _ := json.Marshal(map[string]any{"channelNoList": []int64{no}})
	body, err := c.request(strings.TrimSuffix(EntryURL, "/"), "POST", "/p/oneapp/client/chat/getChannelExtrasList", "application/x-www-form-urlencoded", url.Values{"payload": {string(data)}}.Encode())
	if err != nil {
		return err
	}
	var result struct {
		Code              int
		ChannelExtrasList []struct {
			ChannelNo     json.Number
			ChannelExtras string
		}
	}
	if err := responseJSON(body, &result); err != nil {
		return err
	}
	if result.Code != 200 {
		return fmt.Errorf("웹 메시지방 정보 조회가 거부됐습니다 (code %d)", result.Code)
	}
	for _, channel := range result.ChannelExtrasList {
		if channel.ChannelNo.String() != channelNo.String() {
			continue
		}
		var extras struct{ OwnerTenantID, CollaboSpaceID string }
		if err := json.Unmarshal([]byte(channel.ChannelExtras), &extras); err != nil {
			return fmt.Errorf("웹 메시지방 정보 형식이 변경됐습니다")
		}
		if !positiveNumber(extras.OwnerTenantID) || !positiveNumber(extras.CollaboSpaceID) {
			return fmt.Errorf("이 방에는 노트·할 일 공간이 활성화되어 있지 않습니다. 웹에서 먼저 활성화하세요")
		}
		c.room = Room{ID: id, ChannelNo: channelNo.String(), TenantID: extras.OwnerTenantID, CollaboSpaceID: extras.CollaboSpaceID}
		return nil
	}
	return fmt.Errorf("지정한 웹 메시지방 정보를 찾을 수 없습니다")
}

func positiveNumber(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	n, err := strconv.ParseInt(s, 10, 64)
	return err == nil && n > 0
}
