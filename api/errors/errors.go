package errors

import (
	"fmt"
)

type ErrorList struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

type APIError struct {
	Code int
	Err  error
}

func (apiErr APIError) Error() string {
	return fmt.Sprintf("%v", apiErr.Err)
}

func NewAPIError(code int, message string) APIError {
	return APIError{
		Code: code,
		Err:  fmt.Errorf("%s", message),
	}
}
