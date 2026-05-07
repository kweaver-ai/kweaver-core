// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package interfaces defines entities, DTOs, and service interfaces.
package interfaces

const (
	// DiscoverTask status constants.
	DiscoverTaskStatusPending   string = "pending"
	DiscoverTaskStatusRunning   string = "running"
	DiscoverTaskStatusCompleted string = "completed"
	DiscoverTaskStatusFailed    string = "failed"

	// DiscoverTask trigger type constants.
	DiscoverTaskTriggerManual    string = "manual"    // 手动/立即执行
	DiscoverTaskTriggerScheduled string = "scheduled" // 定时驱动

	// DiscoverTaskType is the task type for discover tasks.
	DiscoverTaskType = "discover:execute"

	// KafkaTopic is the topic for discover task messages.
	DiscoverTaskTopic = "adp-vega-discover-task"
)

// DiscoverTask represents a discover task entity.
type DiscoverTask struct {
	ID          string   `json:"id"`
	CatalogID   string   `json:"catalog_id"`
	ScheduleID  string   `json:"schedule_id"`
	Strategies  []string `json:"strategies"`   // Strategies for discover: can be one or more of ["insert", "delete", "update"], or empty for all
	TriggerType string   `json:"trigger_type"` // manual/scheduled

	Status     string          `json:"status"`   // pending/running/completed/failed
	Progress   int             `json:"progress"` // 0-100
	Message    string          `json:"message"`
	StartTime  int64           `json:"start_time,omitempty"`  // 开始执行时间
	FinishTime int64           `json:"finish_time,omitempty"` // 完成时间
	Result     *DiscoverResult `json:"result,omitempty"`

	Creator    AccountInfo `json:"creator"`
	CreateTime int64       `json:"create_time"`
}

// DiscoverTaskQueryParams holds discover task list query parameters.
type DiscoverTaskQueryParams struct {
	PaginationQueryParams
	CatalogID   string `form:"catalog_id" json:"catalog_id"`
	ScheduleID  string `form:"schedule_id" json:"schedule_id"`
	Status      string `form:"status" json:"status"`
	TriggerType string `form:"trigger_type" json:"trigger_type"`
}

// DiscoverTaskMessage represents the Kafka message for discover task.
type DiscoverTaskMessage struct {
	TaskID string `json:"task_id"`
}
