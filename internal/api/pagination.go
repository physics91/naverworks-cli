package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

func BuildPaginationQuery(cursor string, count int) string {
	params := url.Values{}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	if count > 0 {
		params.Set("count", strconv.Itoa(count))
	}
	return encodeQueryFromValues(params)
}

func encodeQueryFromValues(params url.Values) string {
	if len(params) > 0 {
		return "?" + params.Encode()
	}
	return ""
}

func ExtractNextCursor(body []byte) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return ""
	}
	return extractNextCursorFromParsed(raw)
}

func extractNextCursorFromParsed(raw map[string]json.RawMessage) string {
	metaRaw, ok := raw["responseMetaData"]
	if !ok {
		return ""
	}
	var meta struct {
		NextCursor string `json:"nextCursor"`
	}
	if json.Unmarshal(metaRaw, &meta) != nil {
		return ""
	}
	return meta.NextCursor
}

type FetchFunc func(cursor string) (*Response, error)

func PaginateAll(fetch FetchFunc, itemsKey string) (json.RawMessage, error) {
	allItems := make([]json.RawMessage, 0)
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

		cursor = extractNextCursorFromParsed(raw)
		if cursor == "" {
			break
		}
	}

	return json.Marshal(allItems)
}
