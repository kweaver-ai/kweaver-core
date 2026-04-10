// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	BuildTaskStatusPending   string = "pending"
	BuildTaskStatusRunning   string = "running"
	BuildTaskStatusCompleted string = "completed"
	BuildTaskStatusStopping  string = "stopping"
	BuildTaskStatusStopped   string = "stopped"

	BuildTaskTypeBatch     string = "batch:execute"
	BuildTaskTypeStreaming string = "streaming:execute"
	BuildTaskTypeEmbedding string = "embedding:execute"

	BuildTaskModeStreaming string = "streaming" // 流式
	BuildTaskModeBatch     string = "batch"     // 批量

	BuildTaskExecuteTypeIncremental string = "incremental" // 增量
	BuildTaskExecuteTypeFull        string = "full"        // 全量

	EmptyDocumentID string = "empty_document"

	DATASET_BUILD_MAX_RETRY_COUNT = 50 // 最大重试次数
	DATASET_BUILD_RETRY_INTERVAL  = 5  // 重试间隔，单位秒
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
	Mode string `json:"mode" binding:"required,oneof=streaming batch"` // 任务模式：streaming/batch
}

// UpdateBuildTaskStatusRequest represents update build task status request.
type UpdateBuildTaskStatusRequest struct {
	Status      string `json:"status" binding:"required,oneof=pending running completed stopping stopped"` // 任务状态
	ExecuteType string `json:"execute_type,omitempty"`                                                     // 执行类型,for batch mode, default is "incremental"
}
