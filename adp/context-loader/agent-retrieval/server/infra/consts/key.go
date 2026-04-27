// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package consts provides common constants for the application.
package consts

// Key is the type of value for context key.
type Key string

// String returns the string representation of the key.
func (k Key) String() string {
	return string(k)
}

const (
	// CtxKeyUserID is the key for user id in context.
	CtxKeyUserID   Key = "request_user_id"
	CtxKeyLanguage Key = "stc_request_language"
)

const (
	HeaderOpRequestID Key = "X-Op-Request-ID"
	HeaderOpTraceID   Key = "X-Op-Trace-ID"
	HeaderOpMetrics   Key = "X-Op-Metrics"
	OpExecStartTime   Key = "opStartTime"
)
