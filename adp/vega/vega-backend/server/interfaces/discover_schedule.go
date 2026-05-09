// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

// DiscoverSchedule represents a scheduled discover task configuration.
type DiscoverSchedule struct {
	ID         string   `json:"id"`
	CatalogID  string   `json:"catalog_id"`
	CronExpr   string   `json:"cron_expr"`
	StartTime  int64    `json:"start_time"` // Unix timestamp in milliseconds
	EndTime    int64    `json:"end_time"`   // Unix timestamp in milliseconds, 0 means no end time
	Enabled    bool     `json:"enabled"`
	Strategies []string `json:"strategies"` // Strategies for discover: can be one or more of ["insert", "delete", "update"], or empty for all
	LastRun    int64    `json:"last_run"`   // Unix timestamp in milliseconds of last execution
	NextRun    int64    `json:"next_run"`   // Unix timestamp in milliseconds of next scheduled execution

	Creator    AccountInfo `json:"creator"`
	CreateTime int64       `json:"create_time"`
	Updater    AccountInfo `json:"updater"`
	UpdateTime int64       `json:"update_time"`
}

// DiscoverScheduleQueryParams holds query parameters for scheduled discover tasks.
type DiscoverScheduleQueryParams struct {
	PaginationQueryParams
	CatalogID string `json:"catalog_id"`
	Enabled   *bool  `json:"enabled"`
}

// DiscoverScheduleRequest represents a scheduled discover request.
// Note: This is a simplified version for API requests.
// The full DiscoverSchedule structure is defined in discover_schedule.go
type DiscoverScheduleRequest struct {
	CatalogID string `json:"catalog_id"`
	// Cron expression for scheduling
	CronExpr string `json:"cron_expr"`
	// Optional: start time for the schedule (Unix timestamp in milliseconds)
	StartTime int64 `json:"start_time,omitempty"`
	// Optional: end time for the schedule (Unix timestamp in milliseconds)
	EndTime int64 `json:"end_time,omitempty"`
	// Optional: strategies for discover: can be one or more of ["insert", "delete", "update"], or empty for all
	Strategies []string `json:"strategies,omitempty"`
	// Optional: whether to enable the schedule (default: false)
	Enabled bool `json:"enabled"`
}
