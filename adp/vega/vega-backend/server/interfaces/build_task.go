// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	BuildTaskStatusInit      string = "init"
	BuildTaskStatusRunning   string = "running"
	BuildTaskStatusCompleted string = "completed"
	BuildTaskStatusStopping  string = "stopping"
	BuildTaskStatusStopped   string = "stopped"
	BuildTaskStatusFailed    string = "failed"

	BuildTaskTypeBatch     string = "batch:execute"
	BuildTaskTypeStreaming string = "streaming:execute"
	BuildTaskTypeEmbedding string = "embedding:execute"

	BuildTaskModeStreaming string = "streaming" // 流式
	BuildTaskModeBatch     string = "batch"     // 批量

	BuildTaskExecuteTypeIncremental string = "incremental" // 增量
	BuildTaskExecuteTypeFull        string = "full"        // 全量

	EmptyDocumentID string = "empty_document"

	BUILD_TASK_MAX_RETRY_COUNT = 50 // 最大重试次数
	BUILD_TASK_RETRY_INTERVAL  = 5  // 重试间隔，单位秒

	BUILD_PREFIX = "vega-build"
)

// BuildTask represents a build task entity.
type BuildTask struct {
	ID              string      `json:"id"`
	ResourceID      string      `json:"resource_id"`
	Status          string      `json:"status"`
	Mode            string      `json:"mode"`             // 任务模式：streaming/batch
	TotalCount      int64       `json:"total_count"`      // 总数
	SyncedCount     int64       `json:"synced_count"`     // 已同步数
	VectorizedCount int64       `json:"vectorized_count"` // 已做向量数
	SyncedMark      string      `json:"synced_mark"`      // 同步标记
	ErrorMsg        string      `json:"error_msg,omitempty"`
	Creator         AccountInfo `json:"creator"`
	CreateTime      int64       `json:"create_time"`
	Updater         AccountInfo `json:"updater"`
	UpdateTime      int64       `json:"update_time"`
	EmbeddingFields string      `json:"embedding_fields,omitempty"` // 需向量化嵌入字段
	BuildKeyFields  string      `json:"build_key_fields"`           // 构建中依赖的特殊键字段，如批量构建依赖的有时序性的字段，流式构建依赖的唯一标识某行的字段
	EmbeddingModel  string      `json:"embedding_model,omitempty"`  // 嵌入模型
	ModelDimensions int         `json:"model_dimensions,omitempty"` // 模型维度
}

// BuildTaskRequest represents create build task request.
type BuildTaskRequest struct {
	Mode            string `json:"mode" binding:"required,oneof=streaming batch"` // 任务模式：streaming/batch
	EmbeddingFields string `json:"embedding_fields,omitempty"`                    // 需向量化嵌入字段
	BuildKeyFields  string `json:"build_key_fields"`                              // 构建中依赖的特殊键字段，如批量构建依赖的有时序性的字段，流式构建依赖的唯一标识某行的字段
	EmbeddingModel  string `json:"embedding_model,omitempty"`                     // 嵌入模型
	ModelDimensions int    `json:"model_dimensions,omitempty"`                    // 模型维度
}

// UpdateBuildTaskStatusRequest represents update build task status request.
type UpdateBuildTaskStatusRequest struct {
	Status      string `json:"status" binding:"required,oneof=running stopped"` // 修改任务状态，只允许 running 和 stopped
	ExecuteType string `json:"execute_type,omitempty"`                          // 执行类型,for batch mode, default is "incremental"
}

type KeyValue struct {
	Key   string
	Value any
}
