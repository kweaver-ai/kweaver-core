package errors

import (
	"encoding/json"
	"errors"
	"fmt"
)

// HTTPError http err struct
type HTTPError struct {
	Status int                    `json:"status"`
	Body   map[string]interface{} `json:"body"`
	Err    error                  `json:"err"`
}

// NewHTTPError  new http error instance
func NewHTTPError(info string, status int, body map[string]interface{}) *HTTPError {
	return &HTTPError{
		Status: status,
		Body:   body,
		Err:    errors.New(info),
	}
}

func (h *HTTPError) Error() string {
	return ""
}

// ExHTTPError http错误
type ExHTTPError struct {
	Body   string
	Status int
	Err    error
}

func (e ExHTTPError) Error() string {
	errorinfo := fmt.Sprintf("body : %s , status : %v", e.Body, e.Status)
	return errorinfo
}

// ExHTTPErrorParser parse http error
func ExHTTPErrorParser(err error) (HTTPError, error) {
	httpError, ok := err.(ExHTTPError)
	var httpErrorBody HTTPError

	if !ok {
		return httpErrorBody, err
	}

	httpErrorBody.Status = httpError.Status
	parseErr := json.Unmarshal([]byte(httpError.Body), &httpErrorBody.Body)

	if parseErr != nil {
		return httpErrorBody, err
	}

	return httpErrorBody, nil
}
