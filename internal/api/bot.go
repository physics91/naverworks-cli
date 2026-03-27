package api

import (
	"fmt"
	"net/url"
)

type BotService struct {
	client *Client
}

func NewBotService(client *Client) *BotService {
	return &BotService{client: client}
}

func buildTextMessageBody(text string) ([]byte, error) {
	return marshalBody(map[string]interface{}{
		"content": map[string]interface{}{
			"type": "text",
			"text": text,
		},
	})
}

func (s *BotService) SendTextToUser(botID, userID, text string) (*Response, error) {
	data, err := buildTextMessageBody(text)
	if err != nil {
		return nil, err
	}
	return s.client.Post(fmt.Sprintf("/bots/%s/users/%s/messages", url.PathEscape(botID), url.PathEscape(userID)), data)
}

func (s *BotService) SendTextToChannel(botID, channelID, text string) (*Response, error) {
	data, err := buildTextMessageBody(text)
	if err != nil {
		return nil, err
	}
	return s.client.Post(fmt.Sprintf("/bots/%s/channels/%s/messages", url.PathEscape(botID), url.PathEscape(channelID)), data)
}

func (s *BotService) GetChannel(botID, channelID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s/channels/%s", url.PathEscape(botID), url.PathEscape(channelID)))
}

func (s *BotService) ListChannelMembers(botID, channelID, cursor string, count int) (*Response, error) {
	path := fmt.Sprintf("/bots/%s/channels/%s/members", url.PathEscape(botID), url.PathEscape(channelID)) + BuildPaginationQuery(cursor, count)
	return s.client.Get(path)
}
