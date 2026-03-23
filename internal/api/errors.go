package api

import (
	"encoding/json"
	"fmt"
)

type APIError struct {
	StatusCode  int    `json:"-"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API 에러 %d: %s - %s", e.StatusCode, e.Code, e.Description)
}

func DecodeAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode}

	if len(body) == 0 {
		apiErr.Code = "UNKNOWN"
		apiErr.Description = fmt.Sprintf("HTTP %d (응답 본문 없음)", statusCode)
		return apiErr
	}

	if err := json.Unmarshal(body, apiErr); err == nil && apiErr.Code != "" {
		return apiErr
	}

	var oauthErr struct {
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &oauthErr); err == nil && oauthErr.Error != "" {
		apiErr.Code = oauthErr.Error
		apiErr.Description = oauthErr.ErrorDescription
		return apiErr
	}

	apiErr.Code = "UNKNOWN"
	apiErr.Description = string(body)
	return apiErr
}
