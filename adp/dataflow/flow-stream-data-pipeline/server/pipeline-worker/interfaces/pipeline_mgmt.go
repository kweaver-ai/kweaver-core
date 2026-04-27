package interfaces

import (
	"context"
	"fmt"
)

type contextKey string // 自定义专属的key类型

const (
	CONTENT_TYPE_NAME = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"

	HTTP_HEADER_ACCOUNT_ID   = "x-account-id"
	HTTP_HEADER_ACCOUNT_TYPE = "x-account-type"

	ACCOUNT_INFO_KEY contextKey = "x-account-info" // 避免直接使用string
)

type AccountInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type Pipeline struct {
	PipelineID         string            `json:"id"`
	PipelineName       string            `json:"name"`
	Tags               []string          `json:"-"`
	Comment            string            `json:"-"`
	Builtin            bool              `json:"builtin"`
	OutputType         string            `json:"output_type"`
	IndexBase          string            `json:"index_base"`
	UseIndexBaseInData bool              `json:"use_index_base_in_data"`
	DeploymentConfig   *DeploymentConfig `json:"-"`
	CreateTime         int64             `json:"-"`
	UpdateTime         int64             `json:"-"`

	PipelineStatus        string `json:"status"`
	PipelineStatusDetails string `json:"status_details"`

	InputTopic  string `json:"input_topic"`
	OutputTopic string `json:"output_topic"`
	ErrorTopic  string `json:"error_topic"`
}

func (pipeline *Pipeline) String() string {
	return fmt.Sprintf("{pipeline_id = %s, pipeline_name = %s, builtin = %v, output_type = %s, "+
		"index_base = %v, use_index_base_in_data = %v, pipeline_status = %s, pipeline_status_details = %s, "+
		"input_topic = %s, output_topic = %s, error_topic = %s}", pipeline.PipelineID, pipeline.PipelineName,
		pipeline.Builtin, pipeline.OutputType, pipeline.IndexBase, pipeline.UseIndexBaseInData, pipeline.PipelineStatus, pipeline.PipelineStatusDetails,
		pipeline.InputTopic, pipeline.OutputTopic, pipeline.ErrorTopic)
}

type DeploymentConfig struct {
	CpuLimit    int `json:"cpu_limit"`
	MemoryLimit int `json:"memory_limit"`
}

type PipelineStatusInfo struct {
	Status  string `json:"status"`
	Details string `json:"status_details"`
}

//go:generate mockgen -source ../interfaces/pipeline_mgmt.go -destination ../interfaces/mock/mock_pipeline_mgmt.go
type PipelineMgmtAccess interface {
	GetConfigs(ctx context.Context, pipelineID string, isListen bool) (*Pipeline, bool, error)
	UpdatePipelineStatus(ctx context.Context, pipelineID string, status *PipelineStatusInfo) error
}
