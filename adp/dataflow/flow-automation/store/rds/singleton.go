package rds

import (
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
)

func InitSingleton() {
	rds.SetConfDao(NewConf())
	rds.SetAiModelDao(NewAiModel())
	rds.SetAlarmRuleDao(NewAlarmRule())
	rds.SetContentAdminDao(NewContentAmdin())
	rds.SetAgentDao(NewAgent())
	rds.SetDagInstanceEventRepository(NewDagInstanceEventRepository())
	rds.SetDagInstanceExtDataDao(NewDagInstanceExtDataDao())
	rds.SetExecutorDao(NewExecutor())
	rds.SetTaskCache(NewTaskCache())
	rds.SetFlowStorageDao(NewFlowStorageDao())
	rds.SetFlowFileDao(NewFlowFileDao())
	rds.SetFlowFileDownloadJobDao(NewFlowFileDownloadJobDao())
	rds.SetFlowTaskResumeDao(NewFlowTaskResumeDao())
}
