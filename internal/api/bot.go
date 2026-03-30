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

func (s *BotService) sendText(path, text string) (*Response, error) {
	data, err := buildTextMessageBody(text)
	if err != nil {
		return nil, err
	}
	return s.client.Post(path, data)
}

// ─── Bot CRUD (Task 3-1) ───

func (s *BotService) CreateBot(body []byte) (*Response, error) {
	return s.client.Post("/bots", body)
}

func (s *BotService) ListBots(cursor string, count int) (*Response, error) {
	return s.client.Get("/bots" + BuildPaginationQuery(cursor, count))
}

func (s *BotService) GetBot(botID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s", url.PathEscape(botID)))
}

func (s *BotService) UpdateBot(botID string, body []byte) (*Response, error) {
	return s.client.Put(fmt.Sprintf("/bots/%s", url.PathEscape(botID)), body)
}

func (s *BotService) PatchBot(botID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/bots/%s", url.PathEscape(botID)), body)
}

func (s *BotService) DeleteBot(botID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/bots/%s", url.PathEscape(botID)))
}

func (s *BotService) RegenerateSecret(botID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/bots/%s/secret", url.PathEscape(botID)), nil)
}

// ─── Messages (Task 3-2) ───

func (s *BotService) SendTextToUser(botID, userID, text string) (*Response, error) {
	return s.sendText(fmt.Sprintf("/bots/%s/users/%s/messages", url.PathEscape(botID), url.PathEscape(userID)), text)
}

func (s *BotService) SendTextToChannel(botID, channelID, text string) (*Response, error) {
	return s.sendText(fmt.Sprintf("/bots/%s/channels/%s/messages", url.PathEscape(botID), url.PathEscape(channelID)), text)
}

func (s *BotService) SendMessageToUser(botID, userID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/bots/%s/users/%s/messages", url.PathEscape(botID), url.PathEscape(userID)), body)
}

func (s *BotService) SendMessageToChannel(botID, channelID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/bots/%s/channels/%s/messages", url.PathEscape(botID), url.PathEscape(channelID)), body)
}

// ─── Attachments (Task 3-2) ───

func (s *BotService) CreateAttachment(botID string, body map[string]interface{}) (*Response, error) {
	return s.client.PostJSON(fmt.Sprintf("/bots/%s/attachments", url.PathEscape(botID)), body)
}

func (s *BotService) GetAttachmentDownloadUrl(botID, fileID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/bots/%s/attachments/%s", url.PathEscape(botID), url.PathEscape(fileID)))
}

// ─── Channels (Task 3-3) ───

func (s *BotService) GetChannel(botID, channelID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s/channels/%s", url.PathEscape(botID), url.PathEscape(channelID)))
}

func (s *BotService) ListChannelMembers(botID, channelID, cursor string, count int) (*Response, error) {
	path := fmt.Sprintf("/bots/%s/channels/%s/members", url.PathEscape(botID), url.PathEscape(channelID)) + BuildPaginationQuery(cursor, count)
	return s.client.Get(path)
}

func (s *BotService) CreateChannel(botID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/bots/%s/channels", url.PathEscape(botID)), body)
}

func (s *BotService) LeaveChannel(botID, channelID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/bots/%s/channels/%s", url.PathEscape(botID), url.PathEscape(channelID)))
}

// ─── Domains (Task 3-4) ───

func (s *BotService) RegisterDomain(botID, domainID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/bots/%s/domains/%s", url.PathEscape(botID), url.PathEscape(domainID)), body)
}

func (s *BotService) ListDomains(botID, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s/domains", url.PathEscape(botID)) + BuildPaginationQuery(cursor, count))
}

func (s *BotService) UpdateDomain(botID, domainID string, body []byte) (*Response, error) {
	return s.client.Put(fmt.Sprintf("/bots/%s/domains/%s", url.PathEscape(botID), url.PathEscape(domainID)), body)
}

func (s *BotService) PatchDomain(botID, domainID string, body []byte) (*Response, error) {
	return s.client.Patch(fmt.Sprintf("/bots/%s/domains/%s", url.PathEscape(botID), url.PathEscape(domainID)), body)
}

func (s *BotService) DeleteDomain(botID, domainID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/bots/%s/domains/%s", url.PathEscape(botID), url.PathEscape(domainID)))
}

func (s *BotService) AddDomainMembers(botID, domainID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/bots/%s/domains/%s/members", url.PathEscape(botID), url.PathEscape(domainID)), body)
}

func (s *BotService) ListDomainMembers(botID, domainID, cursor string, count int) (*Response, error) {
	path := fmt.Sprintf("/bots/%s/domains/%s/members", url.PathEscape(botID), url.PathEscape(domainID)) + BuildPaginationQuery(cursor, count)
	return s.client.Get(path)
}

func (s *BotService) RemoveDomainMember(botID, domainID, userID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/bots/%s/domains/%s/members/%s", url.PathEscape(botID), url.PathEscape(domainID), url.PathEscape(userID)))
}

// ─── Persistent Menu (Task 3-5) ───

func (s *BotService) UpsertPersistentMenu(botID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/bots/%s/persistentmenu", url.PathEscape(botID)), body)
}

func (s *BotService) GetPersistentMenu(botID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s/persistentmenu", url.PathEscape(botID)))
}

func (s *BotService) DeletePersistentMenu(botID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/bots/%s/persistentmenu", url.PathEscape(botID)))
}

// ─── Rich Menus (Task 3-6) ───

func (s *BotService) CreateRichMenu(botID string, body []byte) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/bots/%s/richmenus", url.PathEscape(botID)), body)
}

func (s *BotService) ListRichMenus(botID, cursor string, count int) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s/richmenus", url.PathEscape(botID)) + BuildPaginationQuery(cursor, count))
}

func (s *BotService) GetRichMenu(botID, richmenuID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s/richmenus/%s", url.PathEscape(botID), url.PathEscape(richmenuID)))
}

func (s *BotService) DeleteRichMenu(botID, richmenuID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/bots/%s/richmenus/%s", url.PathEscape(botID), url.PathEscape(richmenuID)))
}

func (s *BotService) SetRichMenuImage(botID, richmenuID string, fieldName, fileName string, data []byte) (*Response, error) {
	return s.client.UploadMultipart(
		fmt.Sprintf("/bots/%s/richmenus/%s/image", url.PathEscape(botID), url.PathEscape(richmenuID)),
		fieldName, fileName, data,
	)
}

func (s *BotService) GetRichMenuImage(botID, richmenuID string) (string, error) {
	return s.client.GetDownloadURL(fmt.Sprintf("/bots/%s/richmenus/%s/image", url.PathEscape(botID), url.PathEscape(richmenuID)))
}

func (s *BotService) SetUserRichMenu(botID, richmenuID, userID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/bots/%s/richmenus/%s/users/%s", url.PathEscape(botID), url.PathEscape(richmenuID), url.PathEscape(userID)), nil)
}

func (s *BotService) GetUserRichMenu(botID, userID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s/richmenus/users/%s", url.PathEscape(botID), url.PathEscape(userID)))
}

func (s *BotService) DeleteUserRichMenu(botID, userID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/bots/%s/richmenus/users/%s", url.PathEscape(botID), url.PathEscape(userID)))
}

func (s *BotService) SetDefaultRichMenu(botID, richmenuID string) (*Response, error) {
	return s.client.Post(fmt.Sprintf("/bots/%s/richmenus/%s/set-default", url.PathEscape(botID), url.PathEscape(richmenuID)), nil)
}

func (s *BotService) GetDefaultRichMenu(botID string) (*Response, error) {
	return s.client.Get(fmt.Sprintf("/bots/%s/richmenus/default", url.PathEscape(botID)))
}

func (s *BotService) DeleteDefaultRichMenu(botID string) (*Response, error) {
	return s.client.Delete(fmt.Sprintf("/bots/%s/richmenus/default", url.PathEscape(botID)))
}
