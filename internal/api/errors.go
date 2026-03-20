package api

import "fmt"

type APIError struct {
	StatusCode  int    `json:"-"`
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API 에러 %d: %s - %s", e.StatusCode, e.Code, e.Description)
}
