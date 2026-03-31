package rds

import (
	"context"
)

type ConfDao interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	ListConfigs(ctx context.Context, opt *Options) (configs []ConfModel, err error)
	BatchUpdateConfig(ctx context.Context, configs []*ConfModel) (err error)
}

type AiModelDao interface {
	GetModelInfoByID(ctx context.Context, conditions *QueryCondition) (AiModel, error)
	ListModelInfo(ctx context.Context, params *ListParams, offset, limit int64) ([]AiModel, error)
	DeleteModelInfoByID(ctx context.Context, conditions *QueryCondition) error
	UpdateModelInfo(ctx context.Context, conditions *UpdateCondition, data *UpdateParams) error
	CreateTagsRule(ctx context.Context, data *AiModel) error
	UpdateTrainLog(ctx context.Context, data *AiModel) error
	GetInferSchema(ctx context.Context, trainID string) (string, error)
	CreateTrainFile(ctx context.Context, data *AiModel, trainFile *TrainFileOSSInfo) error
	GetTrainFileInfo(ctx context.Context, trainID string) (TrainFileOSSInfo, error)
	CheckDupName(ctx context.Context, name string) (bool, error)
	VerifyTaskUnique(ctx context.Context) (bool, error)
	GetModelTypeByID(ctx context.Context, id string) (int, error)
}

type AlarmRuleDao interface {
	ModifyAlarmRule(ctx context.Context, ruleID string, rules []*AlarmRule, users []*AlarmUser) error
	GetAlarmRule(ctx context.Context, ruleID string) (*AlarmRule, error)
	ListAlarmRule(ctx context.Context, opt *Options) ([]*AlarmRule, error)
	ListAlarmUser(ctx context.Context, opt *Options) ([]*AlarmUser, error)
	GroupAlarmRule(ctx context.Context) ([]*AlarmRule, error)
	ListDagIDs(ctx context.Context, ruleID string) ([]string, error)
}

type ContentAmdinDao interface {
	CreateAdmin(ctx context.Context, datas []*ContentAdmin) error
	CheckAdminExistByUSerID(ctx context.Context, userID string) (bool, error)
	ListAdmins(ctx context.Context) ([]ContentAdmin, error)
	ListAdminsByUserID(ctx context.Context, userIDs []string) ([]ContentAdmin, error)
	DeleteAdminByID(ctx context.Context, ID string) error
	UpdateAdminByUserID(ctx context.Context, userID, userName string) error
}

type AgentDao interface {
	GetAgents(ctx context.Context) (agents []*AgentModel, err error)
	GetAgentByName(ctx context.Context, name string) (agent *AgentModel, err error)
	DeleteAgentByName(ctx context.Context, name string) (err error)
	CreateAgent(ctx context.Context, agent *AgentModel) (err error)
	UpdateAgent(ctx context.Context, agent *AgentModel) (err error)
}

type DagInstanceEventRepository interface {
	InsertMany(ctx context.Context, events []*DagInstanceEvent) error
	List(ctx context.Context, opts *DagInstanceEventListOptions) ([]*DagInstanceEvent, error)
	ListCount(ctx context.Context, opts *DagInstanceEventListOptions) (int, error)
	DeleteByInstanceIDs(ctx context.Context, instanceIDs []string) error
}

type DagInstanceExtDataDao interface {
	InsertMany(ctx context.Context, items []*DagInstanceExtData) error
	List(ctx context.Context, opts *ExtDataQueryOptions) ([]*DagInstanceExtData, error)
	Remove(ctx context.Context, opts *ExtDataQueryOptions) error
	Delete(ctx context.Context, opts *ExtDataQueryOptions) error
}

type ExecutorDao interface {
	CreateExecutor(ctx context.Context, executor *ExecutorModel) error
	UpdateExecutor(ctx context.Context, executor *ExecutorModel) error
	GetExecutors(ctx context.Context, creatorID string) ([]*ExecutorModel, error)
	GetExecutor(ctx context.Context, id uint64) (*ExecutorModel, error)
	GetExecutorAccessors(ctx context.Context, executorID uint64) ([]*ExecutorAccessorModel, error)
	GetExecutorActions(ctx context.Context, executorID uint64) ([]*ExecutorActionModel, error)
	HasAccessor(ctx context.Context, executorID uint64, accessorIDs []string) (bool, error)
	DeleteExecutor(ctx context.Context, executorID uint64) error
	CreateExecutorAction(ctx context.Context, action *ExecutorActionModel) error
	UpdateExecutorAction(ctx context.Context, action *ExecutorActionModel) error
	DeleteExecutorAction(ctx context.Context, action *ExecutorActionModel) error
	GetAccessableExecutors(ctx context.Context, userID string, accessorIDs []string) ([]*ExecutorModel, error)
	GetAccessableAction(ctx context.Context, actionID uint64, executorID uint64, userID string, accessorIDs []string) (*ExecutorActionModel, error)
	CheckExecutor(ctx context.Context, executor *ExecutorModel) (bool, error)
	CheckExecutorAction(ctx context.Context, action *ExecutorActionModel) (bool, error)
	GetExecutorByName(ctx context.Context, userID string, name string) (executor *ExecutorModel, err error)
}

type TaskCache interface {
	Insert(ctx context.Context, task *TaskCacheItem) error
	GetByHash(ctx context.Context, hash string) (*TaskCacheItem, error)
	Update(ctx context.Context, task *TaskCacheItem) error
	DeleteByHash(ctx context.Context, hash string) error
	ListTaskCache(ctx context.Context, opts ListTaskCacheOptions) ([]*TaskCacheItem, error)
	BatchDeleteByHash(ctx context.Context, hashes []any) error
}

// ============================================================
// FlowStorageDao - Dataflow 存储文件 DAO
// ============================================================

type FlowStorageDao interface {
	Insert(ctx context.Context, storage *FlowStorage) error
	GetByID(ctx context.Context, id uint64) (*FlowStorage, error)
	GetByOssIDAndObjectKey(ctx context.Context, ossID, objectKey string) (*FlowStorage, error)
	List(ctx context.Context, opts *FlowStorageQueryOptions) ([]*FlowStorage, error)
	Update(ctx context.Context, storage *FlowStorage) error
	UpdateStatus(ctx context.Context, id uint64, status FlowStorageStatus) error
	Delete(ctx context.Context, id uint64) error
}

// ============================================================
// FlowFileDao - Dataflow 业务文件 DAO
// ============================================================

type FlowFileDao interface {
	Insert(ctx context.Context, file *FlowFile) error
	GetByID(ctx context.Context, id uint64) (*FlowFile, error)
	List(ctx context.Context, opts *FlowFileQueryOptions) ([]*FlowFile, error)
	Update(ctx context.Context, id uint64, params *FlowFileUpdateParams) error
	UpdateStatus(ctx context.Context, id uint64, status FlowFileStatus) error
	Delete(ctx context.Context, id uint64) error
	CountByStorageID(ctx context.Context, storageID uint64) (int64, error)
}

// ============================================================
// FlowFileDownloadJobDao - Dataflow 文件下载任务 DAO
// ============================================================

type FlowFileDownloadJobDao interface {
	Insert(ctx context.Context, job *FlowFileDownloadJob) error
	GetByID(ctx context.Context, id uint64) (*FlowFileDownloadJob, error)
	GetByFileID(ctx context.Context, fileID uint64) (*FlowFileDownloadJob, error)
	List(ctx context.Context, opts *FlowFileDownloadJobQueryOptions) ([]*FlowFileDownloadJob, error)
	Update(ctx context.Context, id uint64, params *FlowFileDownloadJobUpdateParams) error
	// ClaimJob 乐观锁抢占任务，返回是否成功抢占
	ClaimJob(ctx context.Context, id uint64, startedAt int64) (bool, error)
	Delete(ctx context.Context, id uint64) error
	DeleteByFileID(ctx context.Context, fileID uint64) error
}

// ============================================================
// FlowTaskResumeDao - Dataflow 阻塞任务恢复 DAO
// ============================================================

type FlowTaskResumeDao interface {
	Insert(ctx context.Context, resume *FlowTaskResume) error
	GetByID(ctx context.Context, id uint64) (*FlowTaskResume, error)
	GetByTaskInstanceID(ctx context.Context, taskInstanceID string) (*FlowTaskResume, error)
	List(ctx context.Context, opts *FlowTaskResumeQueryOptions) ([]*FlowTaskResume, error)
	Delete(ctx context.Context, id uint64) error
	DeleteByTaskInstanceID(ctx context.Context, taskInstanceID string) error
	DeleteByResource(ctx context.Context, resourceType string, resourceID uint64) error
}
