package rds

import (
	"fmt"
)

const (
	CONF_TABLENAME              = "t_automation_conf"
	AI_MODEL_TABLENAME          = "t_model"
	ALARM_RULE_TABLENAME        = "t_alarm_rule"
	ALARM_USER_TABLENAME        = "t_alarm_user"
	CONTENT_ADMIN_TABLENAME     = "t_content_admin"
	AGENT_TABLENAME             = "t_automation_agent"
	DAG_INSTANCE_EVENT_TABLE    = "t_dag_instance_event"
	DAG_INSTANCE_EXT_DATA_TABLE = "t_automation_dag_instance_ext_data"
	EXECUTOR_TABLENAME          = "t_automation_executor"
	EXECUTOR_ACCESSOR_TABLENAME = "t_automation_executor_accessor"
	EXECUTOR_ACTION_TABLENAME   = "t_automation_executor_action"
	FLOW_STORAGE_TABLENAME      = "t_flow_storage"
	FLOW_FILE_TABLENAME         = "t_flow_file"
	FLOW_FILE_DOWNLOAD_JOB_TABLENAME = "t_flow_file_download_job"
	FLOW_TASK_RESUME_TABLENAME  = "t_flow_task_resume"
)

const (
	TaskCacheTableFormat = `t_task_cache_%s`
)

type ConfModel struct {
	Key   *string `gorm:"column:f_key;type:char(32);primary_key:not null" json:"key"`
	Value *string `gorm:"column:f_value;type:char(255)" json:"value"`
}

type AiModel struct {
	ID          uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	CreatedAt   int64  `gorm:"column:f_created_at;type:bigint" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:f_updated_at;type:bigint" json:"updated_at"`
	TrainStatus string `gorm:"column:f_train_status;type:varchar(16)" json:"train_status"`
	Status      int    `gorm:"column:f_status;type:tinyint" json:"status"`
	Rule        string `gorm:"column:f_rule;type:text" json:"rule"`
	Name        string `gorm:"column:f_name;type:varchar(255)" json:"name"`
	Description string `gorm:"column:f_description;type:varchar(300)" json:"description"`
	UserID      string `gorm:"column:f_userid;type:varchar(40)" json:"userID"`
	Type        int    `gorm:"column:f_type;type:tinyint" json:"type"`
}

type TrainFileOSSInfo struct {
	ID        uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	TrainID   uint64 `gorm:"column:f_train_id;primary_key:not null" json:"trainID"`
	OSSID     string `gorm:"column:f_oss_id;type:varchar(36)" json:"ossID"`
	Key       string `gorm:"column:f_key;type:varchar(36)" json:"key"`
	CreatedAt int64  `gorm:"column:f_created_at;type:bigint" json:"created_at"`
}

type ContentAdmin struct {
	ID       uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	UserID   string `gorm:"column:f_user_id;type:varchar(40)" json:"userID"`
	UserName string `gorm:"column:f_user_name;type:varchar(128)" json:"userName"`
}

type AlarmRule struct {
	ID        uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	RuleID    uint64 `gorm:"column:f_rule_id;type:bigint" json:"ruleID"`
	DagID     uint64 `gorm:"column:f_dag_id;type:bigint" json:"dagID"`
	Frequency int    `gorm:"column:f_frequency;type:unsigned smallint" json:"frequency"`
	Threshold int    `gorm:"column:f_threshold;type:unsigned mediumint" json:"threshold"`
	CreatedAt int64  `gorm:"column:f_created_at;type:bigint" json:"created_at"`
}

type AlarmUser struct {
	ID       uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	RuleID   uint64 `gorm:"column:f_rule_id;type:bigint" json:"ruleID"`
	UserID   string `gorm:"column:f_user_id;type:varchar(36)" json:"userID"`
	UserName string `gorm:"column:f_user_name;type:varchar(128)" json:"userName"`
	UserType string `gorm:"column:f_user_type;type:varchar(10)" json:"userType"`
}

type AgentModel struct {
	ID      uint64 `gorm:"column:f_id;type:bigint unsigned;primary_key:not null" json:"-"`
	Name    string `gorm:"column:f_name;type:varchar(128);not null;default:''" json:"name"`
	AgentID string `gorm:"column:f_agent_id;type:varchar(64);not null;default:''" json:"agent_id"`
	Version string `gorm:"column:f_version;type:varchar(32);not null;default:''" json:"version"`
}

type DagInstanceEventType uint8

const (
	DagInstanceEventTypeVariable     DagInstanceEventType = 1
	DagInstanceEventTypeTaskStatus   DagInstanceEventType = 2
	DagInstanceEventTypeInstructions DagInstanceEventType = 3
	DagInstanceEventTypeVM           DagInstanceEventType = 4
	DagInstanceEventTypeTrace        DagInstanceEventType = 5
)

type DagInstanceEventVisibility uint8

const (
	DagInstanceEventVisibilityPrivate = 0
	DagInstanceEventVisibilityPublic  = 1
)

type DagInstanceEvent struct {
	ID         uint64                     `gorm:"column:f_id" json:"id,omitempty"`
	Type       DagInstanceEventType       `gorm:"column:f_type" json:"type,omitempty"`
	InstanceID string                     `gorm:"column:f_instance_id" json:"instance_id,omitempty"`
	Operator   string                     `gorm:"column:f_operator" json:"operator,omitempty"`
	TaskID     string                     `gorm:"column:f_task_id" json:"task_id,omitempty"`
	Status     string                     `gorm:"column:f_status" json:"status,omitempty"`
	Name       string                     `gorm:"column:f_name" json:"name,omitempty"`
	Data       string                     `gorm:"column:f_data" json:"data,omitempty"`
	Size       int                        `gorm:"column:f_size" json:"size,omitempty"`
	Inline     bool                       `gorm:"column:f_inline" json:"inline,omitempty"`
	Visibility DagInstanceEventVisibility `gorm:"column:f_visibility" json:"visibility,omitempty"`
	Timestamp  int64                      `gorm:"column:f_timestamp" json:"timestamp,omitempty"`
}

type DagInstanceEventField string

const (
	DagInstanceEventFieldID         DagInstanceEventField = "f_id"
	DagInstanceEventFieldType       DagInstanceEventField = "f_type"
	DagInstanceEventFieldInstanceID DagInstanceEventField = "f_instance_id"
	DagInstanceEventFieldOperator   DagInstanceEventField = "f_operator"
	DagInstanceEventFieldTaskID     DagInstanceEventField = "f_task_id"
	DagInstanceEventFieldStatus     DagInstanceEventField = "f_status"
	DagInstanceEventFieldName       DagInstanceEventField = "f_name"
	DagInstanceEventFieldData       DagInstanceEventField = "f_data"
	DagInstanceEventFieldSize       DagInstanceEventField = "f_size"
	DagInstanceEventFieldInline     DagInstanceEventField = "f_inline"
	DagInstanceEventFieldTimestamp  DagInstanceEventField = "f_timestamp"
	DagInstanceEventFieldVisibility DagInstanceEventField = "f_visibility"
)

var (
	DagInstanceEventFieldAll = []DagInstanceEventField{
		DagInstanceEventFieldID,
		DagInstanceEventFieldType,
		DagInstanceEventFieldInstanceID,
		DagInstanceEventFieldOperator,
		DagInstanceEventFieldTaskID,
		DagInstanceEventFieldStatus,
		DagInstanceEventFieldName,
		DagInstanceEventFieldData,
		DagInstanceEventFieldSize,
		DagInstanceEventFieldInline,
		DagInstanceEventFieldTimestamp,
		DagInstanceEventFieldVisibility,
	}
	DagInstanceEventFieldPublic = []DagInstanceEventField{
		DagInstanceEventFieldType,
		DagInstanceEventFieldOperator,
		DagInstanceEventFieldTaskID,
		DagInstanceEventFieldStatus,
		DagInstanceEventFieldName,
		DagInstanceEventFieldData,
		DagInstanceEventFieldSize,
		DagInstanceEventFieldInline,
		DagInstanceEventFieldTimestamp,
	}
)

type DagInstanceEventListOptions struct {
	DagInstanceID string
	Offset        int
	Limit         int
	Visibilities  []DagInstanceEventVisibility
	Types         []DagInstanceEventType
	Fields        []DagInstanceEventField
	Names         []string
	Inline        *bool
	LatestOnly    bool
}

type DagInstanceExtData struct {
	ID        string `gorm:"column:f_id;primary_key:not null" json:"id" bson:"_id"`
	CreatedAt int64  `gorm:"column:f_created_at;type:bigint" json:"createdAt" bson:"createdAt"`
	UpdatedAt int64  `gorm:"column:f_updated_at;type:bigint" json:"updatedAt" bson:"updatedAt"`
	DagID     string `gorm:"column:f_dag_id;type:varchar(64)" json:"dagId" bson:"dagId"`
	DagInsID  string `gorm:"column:f_dag_ins_id;type:varchar(64)" json:"dagInsId" bson:"dagInsId"`
	Field     string `gorm:"column:f_field;type:varchar(64)" json:"field" bson:"field"`
	OssID     string `gorm:"column:f_oss_id;type:varchar(64)" json:"ossId" bson:"ossId"`
	OssKey    string `gorm:"column:f_oss_key;type:varchar(255)" json:"ossKey" bson:"ossKey"`
	Size      int64  `gorm:"column:f_size;type:bigint" json:"size" bson:"size"`
	Removed   bool   `gorm:"column:f_removed;type:tinyint(1)" json:"removed" bson:"removed"`
}

type ExtDataQueryOptions struct {
	IDs         []string
	DagID       string
	DagInsID    string
	Removed     bool
	Limit       int
	MinID       string
	SelectField []string
}

type ExecutorModel struct {
	ID          *uint64                  `gorm:"column:f_id;primary_key:not null" json:"id"`
	Name        *string                  `gorm:"column:f_name;type:varchar(64)" json:"name"`
	Description *string                  `gorm:"column:f_description;type:varchar(256)" json:"description"`
	CreatorID   *string                  `gorm:"column:f_creator_id;type:varchar(40)" json:"creator_id"`
	Status      *int                     `gorm:"column:f_status;type:tinyint" json:"status"`
	CreatedAt   *int64                   `gorm:"column:f_created_at;type:bigint" json:"created_at"`
	UpdatedAt   *int64                   `gorm:"column:f_updated_at;type:bigint" json:"updated_at"`
	Accessors   []*ExecutorAccessorModel `gorm:"-" json:"accessors"`
	Actions     []*ExecutorActionModel   `gorm:"-" json:"actions"`
}

type ExecutorAccessorModel struct {
	ID           *uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	ExecutorID   *uint64 `gorm:"column:f_executor_id;primary_key:not null" json:"executor_id"`
	AccessorID   *string `gorm:"column:f_accessor_id;type:varchar(40)" json:"accessor_id"`
	AccessorType *string `gorm:"column:f_accessor_type;type:varchar(20)" json:"accessor_type"`
}

type ExecutorActionModel struct {
	ID          *uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	ExecutorID  *uint64 `gorm:"column:f_executor_id;primary_key:not null" json:"executor_id"`
	Operator    *string `gorm:"column:f_operator;type:varchar(64)" json:"operator"`
	Name        *string `gorm:"column:f_name;type:varchar(64)" json:"name"`
	Description *string `gorm:"column:f_description;type:varchar(64)" json:"description"`
	Group       *string `gorm:"column:f_group;type:varchar(64)" json:"group"`
	Type        *string `gorm:"column:f_type;type:varchar(16)" json:"type"`
	Inputs      *string `gorm:"column:f_inputs;type:text" json:"inputs"`
	Outputs     *string `gorm:"column:f_outputs;type:text" json:"outputs"`
	Config      *string `gorm:"column:f_config;type:text" json:"config"`
	CreatedAt   *int64  `gorm:"column:f_created_at;type:bigint" json:"created_at"`
	UpdatedAt   *int64  `gorm:"column:f_updated_at;type:bigint" json:"updated_at"`
}

type ExecutorWithActionModel struct {
	ID          *uint64 `gorm:"column:f_id;primary_key:not null" json:"id"`
	Name        *string `gorm:"column:f_name;type:varchar(64)" json:"name"`
	Description *string `gorm:"column:f_description;type:varchar(256)" json:"description"`
	CreatorID   *string `gorm:"column:f_creator_id;type:varchar(40)" json:"creator_id"`
	Status      *int    `gorm:"column:f_status;type:tinyint" json:"status"`
	CreatedAt   *int64  `gorm:"column:f_created_at;type:bigint" json:"created_at"`
	UpdatedAt   *int64  `gorm:"column:f_updated_at;type:bigint" json:"updated_at"`

	ActionID          *uint64 `gorm:"column:f_action_id;type:bigint" json:"action_id"`
	ActionOperator    *string `gorm:"column:f_action_operator;type:varchar(64)" json:"action_operator"`
	ActionName        *string `gorm:"column:f_action_name;type:varchar(64)" json:"action_name"`
	ActionDescription *string `gorm:"column:f_action_description;type:varchar(256)" json:"action_description"`
	ActionGroup       *string `gorm:"column:f_action_group;type:varchar(64)" json:"action_group"`
	ActionType        *string `gorm:"column:f_action_type;type:varchar(64)" json:"action_type"`
	ActionInputs      *string `gorm:"column:f_action_inputs;type:varchar(256)" json:"action_inputs"`
	ActionOutputs     *string `gorm:"column:f_action_outputs;type:varchar(256)" json:"action_outputs"`
	ActionConfig      *string `gorm:"column:f_action_config;type:varchar(256)" json:"action_config"`
	ActionCreatedAt   *int64  `gorm:"column:f_action_created_at;type:bigint" json:"action_created_at"`
	ActionUpdatedAt   *int64  `gorm:"column:f_action_updated_at;type:bigint" json:"action_updated_at"`
}

type TaskStatus int8

const (
	TaskStatusPending TaskStatus = 1
	TaskStatusSuccess TaskStatus = 2
	TaskStatusFailed  TaskStatus = 3
)

type TaskCacheItem struct {
	ID         uint64     `gorm:"column:f_id;primaryKey;type:char(64);not null" json:"id"`
	Hash       string     `gorm:"column:f_hash;type:char(40);not null;default:''" json:"hash"`
	Type       string     `gorm:"column:f_type;type:varchar(32);not null;default:''" json:"type"`
	Status     TaskStatus `gorm:"column:f_status;type:tinyint(4);not null;default:0" json:"status"`
	OssID      string     `gorm:"column:f_oss_id;type:char(36);not null;default:''" json:"ossId"`
	OssKey     string     `gorm:"column:f_oss_key;type:varchar(255);not null;default:''" json:"ossKey"`
	Ext        string     `gorm:"column:f_ext;type:char(20);not null;default:''" json:"ext"`
	Size       int64      `gorm:"column:f_size;type:bigint(20);not null;default:0" json:"size"`
	ErrMsg     string     `gorm:"column:f_err_msg;type:text" json:"errMsg"`
	CreateTime int64      `gorm:"column:f_create_time;type:bigint(20);not null;default:0" json:"createTime"`
	ModifyTime int64      `gorm:"column:f_modify_time;type:bigint(20);not null;default:0" json:"modifyTime"`
	ExpireTime int64      `gorm:"column:f_expire_time;type:bigint(20);not null;default:0" json:"expireTime"`
}

type ListTaskCacheOptions struct {
	TableSuffix string
	Expired     *bool
	Limit       int64
	MinID       uint64
}

type Options struct {
	OrderBy       *string
	Order         *string
	Limit         *int64
	Page          *int64
	SearchOptions []*SearchOption
}

type SearchOption struct {
	Col       string
	Val       interface{}
	Condition string
}

type UpdateParams struct {
	Status      *int64  `column:"f_status"`
	Rule        *string `column:"f_rule"`
	Name        *string `column:"f_name"`
	Description *string `column:"f_description"`
}

type UpdateCondition struct {
	ID     *string `column:"f_id"`
	UserID *string `column:"f_userid"`
}

type QueryCondition UpdateCondition

type ListParams struct {
	UserID *string
	Status *int64
	Name   *string
}

func (opt *Options) BuildQuery(baseQuery string) (sqlStr string, searchSqlVal []interface{}) {
	sqlStr = baseQuery
	if opt == nil {
		return
	}

	if len(opt.SearchOptions) != 0 {
		var searchSqlStr string
		for _, val := range opt.SearchOptions {
			searchSqlStr = fmt.Sprintf("AND %s %s ? ", val.Col, val.Condition)
			searchSqlVal = append(searchSqlVal, val.Val)
		}
		sqlStr = fmt.Sprintf("%s %s", sqlStr, searchSqlStr)
	}

	if opt.Order != nil && opt.OrderBy != nil {
		sqlStr = fmt.Sprintf("%s ORDER BY %s %s", sqlStr, *opt.OrderBy, *opt.Order)
	}

	if opt.Limit != nil && opt.Page != nil {
		offset := (*opt.Limit) * (*opt.Page)
		sqlStr = fmt.Sprintf("%s LIMIT %v, %v", sqlStr, offset, *opt.Limit)
	}

	return
}

// ============================================================
// FlowStorage - Dataflow 存储文件表
// ============================================================

// FlowStorageStatus 存储对象状态
type FlowStorageStatus int8

const (
	FlowStorageStatusNormal       FlowStorageStatus = 1 // 正常
	FlowStorageStatusPendingDelete FlowStorageStatus = 2 // 待删除
	FlowStorageStatusDeleted      FlowStorageStatus = 3 // 已删除
)

// FlowStorage 仅用于描述已经落到 OssGateway 中的物理对象
type FlowStorage struct {
	ID          uint64            `gorm:"column:f_id;primaryKey;type:bigint unsigned;not null" json:"id"`
	OssID       string            `gorm:"column:f_oss_id;type:varchar(64);not null;default:''" json:"oss_id"`
	ObjectKey   string            `gorm:"column:f_object_key;type:varchar(512);not null;default:''" json:"object_key"`
	Name        string            `gorm:"column:f_name;type:varchar(256);not null;default:''" json:"name"`
	ContentType string            `gorm:"column:f_content_type;type:varchar(128);not null;default:''" json:"content_type"`
	Size        uint64            `gorm:"column:f_size;type:bigint unsigned;not null;default:0" json:"size"`
	Etag        string            `gorm:"column:f_etag;type:varchar(128);not null;default:''" json:"etag"`
	Status      FlowStorageStatus `gorm:"column:f_status;type:tinyint;not null;default:1" json:"status"`
	CreatedAt   int64             `gorm:"column:f_created_at;type:bigint;not null;default:0" json:"created_at"`
	UpdatedAt   int64             `gorm:"column:f_updated_at;type:bigint;not null;default:0" json:"updated_at"`
	DeletedAt   int64             `gorm:"column:f_deleted_at;type:bigint;not null;default:0" json:"deleted_at"`
}

// FlowStorageQueryOptions FlowStorage 查询选项
type FlowStorageQueryOptions struct {
	IDs       []uint64
	OssID     string
	ObjectKey string
	Status    *FlowStorageStatus
	Limit     int
	Offset    int
}

// ============================================================
// FlowFile - Dataflow 业务文件表
// ============================================================

// FlowFileStatus 文件对象状态
type FlowFileStatus int8

const (
	FlowFileStatusPending FlowFileStatus = 1 // 待就绪
	FlowFileStatusReady   FlowFileStatus = 2 // 就绪
	FlowFileStatusInvalid FlowFileStatus = 3 // 失效
	FlowFileStatusExpired FlowFileStatus = 4 // 已过期
)

// FlowFile 描述 Dataflow 内部文件对象，并承载 dfs:// 协议
type FlowFile struct {
	ID            uint64         `gorm:"column:f_id;primaryKey;type:bigint unsigned;not null" json:"id"` // 对应 dfs://<id>
	DagID         string         `gorm:"column:f_dag_id;type:varchar(64);not null;default:''" json:"dag_id"`
	DagInstanceID string         `gorm:"column:f_dag_instance_id;type:varchar(64);not null;default:''" json:"dag_instance_id"`
	StorageID     uint64         `gorm:"column:f_storage_id;type:bigint unsigned;not null;default:0" json:"storage_id"` // 存储文件ID，未落OSS时为0
	Status        FlowFileStatus `gorm:"column:f_status;type:tinyint;not null;default:1" json:"status"`
	Name          string         `gorm:"column:f_name;type:varchar(256);not null;default:''" json:"name"`
	ExpiresAt     int64          `gorm:"column:f_expires_at;type:bigint;not null;default:0" json:"expires_at"` // 过期时间 0表示不过期
	CreatedAt     int64          `gorm:"column:f_created_at;type:bigint;not null;default:0" json:"created_at"`
	UpdatedAt     int64          `gorm:"column:f_updated_at;type:bigint;not null;default:0" json:"updated_at"`
}

// FlowFileQueryOptions FlowFile 查询选项
type FlowFileQueryOptions struct {
	ID            *uint64
	IDs           []uint64
	DagID         string
	DagInstanceID string
	StorageID     *uint64
	Status        *FlowFileStatus
	Statuses      []FlowFileStatus
	ExpiresBefore int64 // 过期时间早于此值
	Limit         int
	Offset        int
	OrderBy       string // 排序字段，如 "created_at"
	Order         string // 排序方向，"asc" 或 "desc"
}

// FlowFileUpdateParams FlowFile 更新参数
type FlowFileUpdateParams struct {
	StorageID     *uint64
	Status        *FlowFileStatus
	Name          *string
	ExpiresAt     *int64
	DagInstanceID *string
}

// ============================================================
// FlowFileDownloadJob - Dataflow 文件下载任务表
// ============================================================

// FlowFileDownloadJobStatus 下载任务状态
type FlowFileDownloadJobStatus int8

const (
	FlowFileDownloadJobStatusPending  FlowFileDownloadJobStatus = 1 // 待执行
	FlowFileDownloadJobStatusRunning  FlowFileDownloadJobStatus = 2 // 执行中
	FlowFileDownloadJobStatusSuccess  FlowFileDownloadJobStatus = 3 // 成功
	FlowFileDownloadJobStatusFailed   FlowFileDownloadJobStatus = 4 // 失败
	FlowFileDownloadJobStatusCanceled FlowFileDownloadJobStatus = 5 // 取消
)

// FlowFileDownloadJob 仅用于管理 URL 文件下载任务
type FlowFileDownloadJob struct {
	ID           uint64                    `gorm:"column:f_id;primaryKey;type:bigint unsigned;not null" json:"id"`
	FileID       uint64                    `gorm:"column:f_file_id;type:bigint unsigned;not null" json:"file_id"` // 关联flow_file ID
	Status       FlowFileDownloadJobStatus `gorm:"column:f_status;type:tinyint;not null;default:1" json:"status"`
	RetryCount   int                       `gorm:"column:f_retry_count;type:int;not null;default:0" json:"retry_count"`
	MaxRetry     int                       `gorm:"column:f_max_retry;type:int;not null;default:3" json:"max_retry"`
	NextRetryAt  int64                     `gorm:"column:f_next_retry_at;type:bigint;not null;default:0" json:"next_retry_at"`
	ErrorCode    string                    `gorm:"column:f_error_code;type:varchar(64);not null;default:''" json:"error_code"`
	ErrorMessage string                    `gorm:"column:f_error_message;type:varchar(1024);not null;default:''" json:"error_message"`
	DownloadURL  string                    `gorm:"column:f_download_url;type:varchar(2048);not null;default:''" json:"download_url"`
	StartedAt    int64                     `gorm:"column:f_started_at;type:bigint;not null;default:0" json:"started_at"`
	FinishedAt   int64                     `gorm:"column:f_finished_at;type:bigint;not null;default:0" json:"finished_at"`
	CreatedAt    int64                     `gorm:"column:f_created_at;type:bigint;not null;default:0" json:"created_at"`
	UpdatedAt    int64                     `gorm:"column:f_updated_at;type:bigint;not null;default:0" json:"updated_at"`
}

// FlowFileDownloadJobQueryOptions FlowFileDownloadJob 查询选项
type FlowFileDownloadJobQueryOptions struct {
	ID          *uint64
	FileID      *uint64
	Status      *FlowFileDownloadJobStatus
	Statuses    []FlowFileDownloadJobStatus
	RetryBefore int64 // 下次重试时间早于此值
	Limit       int
	Offset      int
}

// FlowFileDownloadJobUpdateParams FlowFileDownloadJob 更新参数
type FlowFileDownloadJobUpdateParams struct {
	Status       *FlowFileDownloadJobStatus
	RetryCount   *int
	NextRetryAt  *int64
	ErrorCode    *string
	ErrorMessage *string
	StartedAt    *int64
	FinishedAt   *int64
}

// ============================================================
// FlowTaskResume - Dataflow 阻塞任务恢复表
// ============================================================

// FlowTaskResume 提供服务内部可持久化的 task_instance 恢复机制
type FlowTaskResume struct {
	ID              uint64 `gorm:"column:f_id;primaryKey;type:bigint unsigned;not null" json:"id"`
	TaskInstanceID  string `gorm:"column:f_task_instance_id;type:varchar(64);not null;default:''" json:"task_instance_id"` // 被阻塞的任务实例ID
	DagInstanceID   string `gorm:"column:f_dag_instance_id;type:varchar(64);not null;default:''" json:"dag_instance_id"`   // 所属流程实例ID
	ResourceType    string `gorm:"column:f_resource_type;type:varchar(32);not null;default:'file'" json:"resource_type"`  // 资源类型
	ResourceID      uint64 `gorm:"column:f_resource_id;type:bigint unsigned;not null;default:0" json:"resource_id"`       // 资源ID，对文件场景即flow_file ID
	CreatedAt       int64  `gorm:"column:f_created_at;type:bigint;not null;default:0" json:"created_at"`
	UpdatedAt       int64  `gorm:"column:f_updated_at;type:bigint;not null;default:0" json:"updated_at"`
}

// FlowTaskResumeQueryOptions FlowTaskResume 查询选项
type FlowTaskResumeQueryOptions struct {
	ID             *uint64
	TaskInstanceID string
	DagInstanceID  string
	ResourceType   string
	ResourceID     *uint64
	Limit          int
	Offset         int
}
