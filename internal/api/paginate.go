package api

import (
	"encoding/json"
	"fmt"
)

type FetchFunc func(cursor string) (*Response, error)

func PaginateAll(fetch FetchFunc, itemsKey string) (json.RawMessage, error) {
	var allItems []json.RawMessage
	cursor := ""

	for {
		resp, err := fetch(cursor)
		if err != nil {
			return nil, err
		}

		var raw map[string]json.RawMessage
		if err := json.Unmarshal(resp.Body, &raw); err != nil {
			return nil, fmt.Errorf("페이지 응답 파싱 실패: %w", err)
		}

		if items, ok := raw[itemsKey]; ok {
			var pageItems []json.RawMessage
			if err := json.Unmarshal(items, &pageItems); err != nil {
				return nil, fmt.Errorf("%s 파싱 실패: %w", itemsKey, err)
			}
			allItems = append(allItems, pageItems...)
		}

		cursor = ExtractNextCursor(resp.Body)
		if cursor == "" {
			break
		}
	}

	return json.Marshal(allItems)
}
