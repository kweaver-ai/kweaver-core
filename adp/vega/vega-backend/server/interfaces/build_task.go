// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	BuildTaskStatusPending   string = "pending"
	BuildTaskStatusRunning   string = "running"
	BuildTaskStatusCompleted string = "completed"
	BuildTaskStatusFailed    string = "failed"
	BuildTaskType            string = "build:execute"

	BuildTaskModeFull        string = "full"
	BuildTaskModeIncremental string = "incremental"
	BuildTaskModeRealtime    string = "realtime"
)

// BuildTask represents a build task entity.
type BuildTask struct {
	ID              string      `json:"id"`
	ResourceID      string      `json:"resource_id"`
	Status          string      `json:"status"`
	Mode            string      `json:"mode"`             // 任务模式：full/incremental/realtime
	TotalCount      int64       `json:"total_count"`      // 总数
	SyncedCount     int64       `json:"synced_count"`     // 已同步数
	VectorizedCount int64       `json:"vectorized_count"` // 已做向量数
	SyncedMark      string      `json:"synced_mark"`      // 同步标记
	ErrorMsg        string      `json:"error_msg,omitempty"`
	Creator         AccountInfo `json:"creator"`
	CreateTime      int64       `json:"create_time"`
	Updater         AccountInfo `json:"updater"`
	UpdateTime      int64       `json:"update_time"`
}

// BuildTaskRequest represents create build task request.
type BuildTaskRequest struct {
	Mode       string `json:"mode" binding:"required,oneof=full incremental realtime"` // 任务模式：full/incremental/realtime
}
