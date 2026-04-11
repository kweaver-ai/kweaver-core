// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

// ScheduledDiscoverRequest represents a scheduled discover request.
// Note: This is a simplified version for API requests.
// The full ScheduledDiscoverTask structure is defined in scheduled_discover_task.go
type ScheduledDiscoverRequest struct {
	CatalogID string `json:"catalog_id"`
	// Cron expression for scheduling
	CronExpr string `json:"cron_expr"`
	// Optional: start time for the schedule (Unix timestamp in milliseconds)
	StartTime int64 `json:"start_time,omitempty"`
	// Optional: end time for the schedule (Unix timestamp in milliseconds)
	EndTime int64 `json:"end_time,omitempty"`
	// Optional: strategies for discover: can be one or more of ["insert", "delete", "update"], or empty for all
	Strategies []string `json:"strategies,omitempty"`
}
