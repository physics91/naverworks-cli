package api

import (
	"encoding/json"
	"fmt"
)

type BotService struct {
	client *Client
}

func NewBotService(client *Client) *BotService {
	return &BotService{client: client}
}

func (s *BotService) SendTextToUser(botID, userID, text string) (*Response, error) {
	body := map[string]interface{}{
		"content": map[string]interface{}{
			"type": "text",
			"text": text,
		},
	}
	data, _ := json.Marshal(body)
	return s.client.Post(fmt.Sprintf("/bots/%s/users/%s/messages", botID, userID), data)
}

func (s *BotService) SendTextToChannel(botID, channelID, text string) (*Response, error) {
	body := map[string]interface{}{
		"content": map[string]interface{}{
			"type": "text",
			"text": text,
		},
	}
	data, _ := json.Marshal(body)
	return s.client.Post(fmt.Sprintf("/bots/%s/channels/%s/messages", botID, channelID), data)
}

func (s *BotService) GetChannel(botID, channelID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s/channels/%s", botID, channelID))
}

func (s *BotService) ListChannelMembers(botID, channelID, cursor string, count int) (*Response, error) {
	path := fmt.Sprintf("/bots/%s/channels/%s/members", botID, channelID) + BuildPaginationQuery(cursor, count)
	return s.client.Get(path)
}
