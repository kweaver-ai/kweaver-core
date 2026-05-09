// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package rate

import (
	"errors"
	"net/http"
	"strings"
)

// Rate limit error types
var (
	// ErrGlobalLimitExceeded indicates global concurrency limit was exceeded
	ErrGlobalLimitExceeded = errors.New("global concurrency limit exceeded")

	// ErrCatalogLimitExceeded indicates catalog-level concurrency limit was exceeded
	ErrCatalogLimitExceeded = errors.New("catalog concurrency limit exceeded")

	// ErrQueueTimeout indicates timeout waiting for a permit
	ErrQueueTimeout = errors.New("queue timeout waiting for slot")
)

// RateLimitError represents a rate limiting error.
type RateLimitError struct {
	Err        error
	Message    string
	HTTPStatus int
	LimitType  string // "global" or "catalog"
	RetryAfter float64 // Suggested retry time in seconds
	Limit      int     // The limit that was exceeded
	Current    int     // Current usage
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(err error, message string) *RateLimitError {
	rateErr := &RateLimitError{
		Err:        err,
		Message:    message,
		RetryAfter: 5.0, // Default 5 seconds
	}

	if err == ErrGlobalLimitExceeded {
		rateErr.HTTPStatus = http.StatusTooManyRequests
		rateErr.LimitType = "global"
		rateErr.RetryAfter = 5.0
	} else if err == ErrCatalogLimitExceeded {
		rateErr.HTTPStatus = http.StatusTooManyRequests
		rateErr.LimitType = "catalog"
		rateErr.RetryAfter = 10.0
	} else if err == ErrQueueTimeout {
		rateErr.HTTPStatus = http.StatusServiceUnavailable
		rateErr.LimitType = "queue"
		rateErr.RetryAfter = 30.0
	}

	return rateErr
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	return e.Message
}

// Unwrap returns the underlying error.
func (e *RateLimitError) Unwrap() error {
	return e.Err
}

// WithDetails adds additional details to the error.
func (e *RateLimitError) WithDetails(limit, current int) *RateLimitError {
	e.Limit = limit
	e.Current = current
	return e
}

// Is implements error matching.
func (e *RateLimitError) Is(target error) bool {
	return e.Err == target
}

// HasError checks if the error contains the given substring.
func HasError(err error, substr string) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), strings.ToLower(substr))
}
