package api

import (
	"encoding/json"
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
	if len(params) > 0 {
		return "?" + params.Encode()
	}
	return ""
}

func encodeQueryFromValues(params url.Values) string {
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
	_ = json.Unmarshal(body, &resp)
	return resp.ResponseMetaData.NextCursor
}
