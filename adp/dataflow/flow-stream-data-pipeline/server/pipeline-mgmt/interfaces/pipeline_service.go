package interfaces

import (
	"context"
	"time"
)

const (
	ServiceName       = "sdp"
	ManagerDeployName = "dataflow"
	ServiceAccount    = "dataflow-service-account"
)

const (
	// 对象id的校验
	RegexPattern_ID = "^[a-z0-9][a-z0-9-]{0,39}$"
)

const (
	COMMENT_MAX_LENGTH     = 255
	OBJECT_NAME_MAX_LENGTH = 40
	DEFAULT_OFFEST         = "0"
	DEFAULT_LIMIT          = "10"
	RESOURCES_PAGE_LIMIT   = "50"
	DESC_DIRECTION         = "desc"
	ASC_DIRECTION          = "asc"
	MIN_OFFSET             = 0
	MIN_LIMIT              = 1
	MAX_LIMIT              = 1000
	NO_LIMIT               = -1
	DEFAULT_SORT           = "update_time"

	// 对象类型 .toml中的messageId
	OBJECTTYPE = "ID_AUDIT_STREAM_DATA_PIPELINE"

	AUTO_OFFSET_RESET_LATEST   = "latest"
	AUTO_OFFSET_RESET_EARLIEST = "earliest"

	CPU_LIMIT_MIN    = 1
	MEMORY_LIMIT_MIN = 100 // MiB

	TAGS_MAX_NUMBER            = 5
	TAG_NAME_INVALID_CHARACTER = "/:?\\\"<>|：？‘’“”！《》,#[]{}%&*$^!=.'"

	MaxSubCondition = 10
)

const (
	PipelineStatus_Error   = "error"
	PipelineStatus_Running = "running"
	PipelineStatus_Close   = "close"
	PipelineStatus_Closing = "closing"

	EnvPipelineNamespace = "SDP_PIPELINE_NAMESPACE"
	EnvPipelineImage     = "SDP_PIPELINE_IMAGE"
	PipelineDeployLabels = "pipeline-worker"
)

const (
	INDEX_BASE = "index_base"
)

var (
	TABLE_SORT = map[string]string{
		"update_time": "f_update_time",
		"name":        "f_pipeline_name",
	}

	Pipeline_Status_MAP = map[string]string{
		PipelineStatus_Running: "running",
		PipelineStatus_Close:   "close",
		PipelineStatus_Error:   "error",
	}

	OutputTypeMap = map[string]struct{}{
		INDEX_BASE: {},
	}
)

type PipelineStatusParamter struct {
	Status  string `json:"status"`
	Details string `json:"status_details"`
}

//go:generate mockgen -source ../interfaces/pipeline_service.go -destination ../interfaces/mock/mock_pipeline_service.go
type PipelineMgmtService interface {
	CreatePipeline(ctx context.Context, pipelineInfo *Pipeline) (pipelineID string, err error)
	DeletePipeline(ctx context.Context, pipelineID string) error
	UpdatePipeline(ctx context.Context, pipelineInfo *Pipeline) error
	GetPipeline(ctx context.Context, pipelineID string, isListen bool) (pipelineInfo *Pipeline, isExist bool, err error)
	ListPipelines(ctx context.Context, pipelineQuery *ListPipelinesQuery) (pipelineInfoList []*Pipeline, totals int, err error)
	GetPipelineTotals(ctx context.Context, pipelineQuery *ListPipelinesQuery) (totals int, err error)
	UpdatePipelineStatus(ctx context.Context, pipelineID string, pipelineStatusInfo *PipelineStatusParamter, isInnerRequest bool) error
	WatchPipelineDeploys(ctx context.Context, interval time.Duration)

	CheckPipelineExistByName(ctx context.Context, name string) (string, bool, error)
	CheckPipelineExistByID(ctx context.Context, ID string) (string, bool, error)

	ListPipelineResources(ctx context.Context, param *ListPipelinesQuery) ([]*Resource, int, error)
}
