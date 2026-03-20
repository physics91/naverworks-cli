package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func BuildPaginationQuery(cursor string, count int) string {
	params := url.Values{}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	if count > 0 {
		params.Set("count", fmt.Sprintf("%d", count))
	}
	if len(params) > 0 {
		return "?" + params.Encode()
	}
	return ""
}

func ExtractNextCursor(body []byte) string {
	var resp struct {
		ResponseMetaData struct {
			NextCursor string `json:"nextCursor"`
		} `json:"responseMetaData"`
	}
	json.Unmarshal(body, &resp)
	return resp.ResponseMetaData.NextCursor
}
