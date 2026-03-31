package interfaces

import (
	"context"
	"fmt"
)

type Pipeline struct {
	PipelineID         string            `json:"id"`
	PipelineName       string            `json:"name"`
	Tags               []string          `json:"tags"`
	Comment            string            `json:"comment"`
	Builtin            bool              `json:"builtin"`
	OutputType         string            `json:"output_type"`
	IndexBase          string            `json:"index_base"`
	UseIndexBaseInData bool              `json:"use_index_base_in_data"`
	DeploymentConfig   *DeploymentConfig `json:"deployment_config"`
	CreateTime         int64             `json:"create_time"`
	UpdateTime         int64             `json:"update_time"`
	Creator            AccountInfo       `json:"creator"`
	Updater            AccountInfo       `json:"updater"`

	PipelineStatus        string `json:"status"`
	PipelineStatusDetails string `json:"status_details"`

	InputTopic  string `json:"input_topic,omitempty"`
	OutputTopic string `json:"output_topic,omitempty"`
	ErrorTopic  string `json:"error_topic,omitempty"`

	// 操作权限
	Operations []string `json:"operations"`
}

type DeploymentConfig struct {
	CpuLimit    int `json:"cpu_limit"`
	MemoryLimit int `json:"memory_limit"`
}

// 分页查询参数
type PaginationQueryParameters struct {
	Offset    int    `json:"offset"`
	Limit     int    `json:"limit"`
	Sort      string `json:"sort"`
	Direction string `json:"direction"`
}

// 列表查询参数
type ListPipelinesQuery struct {
	NamePattern    string
	PipelineStatus []string
	Tag            string
	Builtin        []bool
	PaginationQueryParameters
}

func (pipeline *Pipeline) String() string {
	return fmt.Sprintf("{pipeline_id = %s, pipeline_name = %s, tags = %v, comment = %s, builtin = %v, output_type = %s, "+
		"index_base = %v, use_index_base_in_data = %v, deployment_config = %+v, create_time = %d, update_time = %d, status = %s, status_details = %s, "+
		"input_topic = %s, output_topic = %s, error_topic = %s}", pipeline.PipelineID, pipeline.PipelineName, pipeline.Tags,
		pipeline.Comment, pipeline.Builtin, pipeline.OutputType, pipeline.IndexBase, pipeline.UseIndexBaseInData, pipeline.DeploymentConfig,
		pipeline.CreateTime, pipeline.UpdateTime, pipeline.PipelineStatus, pipeline.PipelineStatusDetails,
		pipeline.InputTopic, pipeline.OutputTopic, pipeline.ErrorTopic)
}

//go:generate mockgen -source ./pipeline_access.go -destination ./mock/mock_pipeline_access.go
type PipelineMgmtAccess interface {
	CreatePipeline(ctx context.Context, pipelineInfo *Pipeline) error
	DeletePipeline(ctx context.Context, pipelineID string) error
	UpdatePipeline(ctx context.Context, pipelineInfo *Pipeline) error
	GetPipeline(ctx context.Context, pipelineID string) (pipelineInfo *Pipeline, isExist bool, err error)
	ListPipelines(ctx context.Context, pipelineQuery *ListPipelinesQuery) (pipelineInfoList []*Pipeline, err error)
	GetPipelinesTotal(ctx context.Context, pipelineQuery *ListPipelinesQuery) (totals int, err error)
	UpdatePipelineStatus(ctx context.Context, pipelineInfo *Pipeline, isInnerRequest bool) error

	CheckPipelineExistByName(ctx context.Context, name string) (string, bool, error)
	CheckPipelineExistByID(ctx context.Context, ID string) (string, bool, error)
}
