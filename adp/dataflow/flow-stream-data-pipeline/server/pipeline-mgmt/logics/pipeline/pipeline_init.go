package pipeline

import (
	"context"
	"sync"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

var (
	// 审计日志管道
	DIP_AUDIT_LOG = &interfaces.Pipeline{
		PipelineID:   "isf-audit-log",
		PipelineName: "isf-audit-log",
		Builtin:      true,
		OutputType:   "index_base",
		IndexBase:    "isf_audit_log",
		DeploymentConfig: &interfaces.DeploymentConfig{
			CpuLimit:    1,
			MemoryLimit: 512,
		},
	}

	// // log 可观测性管道
	// DIP_O11Y_LOG = &interfaces.Pipeline{
	// 	PipelineID:   "dip-o11y-log",
	// 	PipelineName: "dip-o11y-log",
	// 	Builtin:      true,
	// 	OutputType:   "index_base",
	// 	IndexBase:    "dip_o11y_log",
	// 	DeploymentConfig: &interfaces.DeploymentConfig{
	// 		CpuLimit:    1,
	// 		MemoryLimit: 2048,
	// 	},
	// }

	// // metric 可观测性管道
	// DIP_O11Y_METRIC = &interfaces.Pipeline{
	// 	PipelineID:   "dip-o11y-metric",
	// 	PipelineName: "dip-o11y-metric",
	// 	Builtin:      true,
	// 	OutputType:   "index_base",
	// 	IndexBase:    "dip_o11y_metric",
	// 	DeploymentConfig: &interfaces.DeploymentConfig{
	// 		CpuLimit:    1,
	// 		MemoryLimit: 2048,
	// 	},
	// }

	// // trace 可观测性管道
	// DIP_O11Y_TRACE = &interfaces.Pipeline{
	// 	PipelineID:   "dip-o11y-trace",
	// 	PipelineName: "dip-o11y-trace",
	// 	Builtin:      true,
	// 	OutputType:   "index_base",
	// 	IndexBase:    "dip_o11y_trace",
	// 	DeploymentConfig: &interfaces.DeploymentConfig{
	// 		CpuLimit:    1,
	// 		MemoryLimit: 2048,
	// 	},
	// }

	// 模型持久化任务管道，使用数据里配置的索引库
	DIP_MODEL_PERSISTENCE = &interfaces.Pipeline{
		PipelineID:         "mdl-model-persistence",
		PipelineName:       "mdl-model-persistence",
		Builtin:            true,
		OutputType:         "index_base",
		UseIndexBaseInData: true,
		IndexBase:          "",
		DeploymentConfig: &interfaces.DeploymentConfig{
			CpuLimit:    1,
			MemoryLimit: 2048,
		},
	}
)

var (
	piOnce sync.Once
	pi     *PipelineInit
)

type PipelineInit struct {
	appSetting *common.AppSetting
	pmService  interfaces.PipelineMgmtService
}

func NewPipelineInit(appSetting *common.AppSetting) *PipelineInit {
	piOnce.Do(func() {
		pi = &PipelineInit{
			appSetting: appSetting,
			pmService:  NewPipelineMgmtService(appSetting),
		}
	})
	return pi
}

func (pi *PipelineInit) Init() {
	// userId 存入 context 中
	// ctx := context.WithValue(context.Background(), interfaces.ACCOUNT_INFO_KEY, interfaces.AccountInfo{
	// 	ID:   interfaces.ADMIN_ID,
	// 	Type: interfaces.ACCESSOR_TYPE_USER,
	// })

	// 检查 MQType
	mqType := common.GetMQType()
	if mqType != common.MQType_Kafka {
		logger.Errorf("MQ Type is not '%s', but is '%s', skip create internal pipelines", common.MQType_Kafka, mqType)
		return
	}

	// pi.CreatePipeline(ctx, DIP_AUDIT_LOG)
	// pi.CreatePipeline(ctx, DIP_O11Y_LOG)
	// pi.CreatePipeline(ctx, DIP_O11Y_METRIC)
	// pi.CreatePipeline(ctx, DIP_O11Y_TRACE)
	// pi.CreatePipeline(ctx, DIP_MODEL_PERSISTENCE)
}

func (pi *PipelineInit) CreatePipeline(ctx context.Context, pipeline *interfaces.Pipeline) {
	_, exist, err := pi.pmService.CheckPipelineExistByName(ctx, pipeline.PipelineName)
	if err != nil {
		logger.Errorf("failed to check pipeline exits by name '%s': %v", pipeline.PipelineName, err)
	}

	if !exist {
		_, err = pi.pmService.CreatePipeline(ctx, pipeline)
		if err != nil {
			logger.Errorf("failed to create pipeline '%s': %v", pipeline.PipelineName, err)
		}
	}
}
