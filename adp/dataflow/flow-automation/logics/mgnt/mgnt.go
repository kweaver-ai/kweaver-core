// Package mgnt logics 自动化任务管理
package mgnt

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/errors"
	libErrors "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/errors"
	threadPool "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/pools"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/s3"
	cstore "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/store"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/logics/perm"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/actions"
	pkgconfig "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/config"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/dependency"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/log"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/mod"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/rds"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/render"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/utils/value"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/pkg/vm/state"
	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils"
	normalizeutil "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/utils/normalize"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//go:generate mockgen -package mock_logics -source ../../logics/mgnt/mgnt.go -destination ../../tests/mock_logics/mgnt_mock.go
const taskTotal = 50

// AnyshareEventTriggerList anyshare event trigger
var AnyshareEventTriggerList = []string{
	common.AnyshareFileCopyTrigger,
	common.AnyshareFileUploadTrigger,
	common.AnyshareFileMoveTrigger,
	common.AnyshareFileRemoveTrigger,
	common.AnyshareFolderCreateTrigger,
	common.AnyshareFolderMoveTrigger,
	common.AnyshareFolderCopyTrigger,
	common.AnyshareFolderRemoveTrigger,
}

// templateName 模板key对用模板名称
var templateName = map[string]string{
	"identifyInvoice":      "自动识别发票信息并同步至表格",
	"identifyIdCard":       "自动识别身份证信息并同步至表格",
	"deleteApproval":       "文档操作（删除）申请",
	"renameApproval":       "文档操作（重命名）申请",
	"permission":           "文档权限申请",
	"contractRelay":        "合同类文件流转管理",
	"expansion":            "配额空间扩容申请",
	"docRelay":             "外发公文流转管理",
	"knowledgeFlow":        "产品知识发布流程",
	"directory":            "自动创建项目文件夹目录",
	"recognizeResume":      "特定文件类型的内容管理",
	"recognizeMove":        "识别文件内容后，基于文件内容新建文件夹层级并移动文件",
	"automaticArchiving":   "基于文件创建时间设置自动归档流程",
	"tagging":              "自动添加文档分类以便于内容管理",
	"batchDeletion":        "自动批量清除文件",
	"setcsfLevel":          "基于敏感词出现频率自动设置对应密级",
	"matchTextForTemplate": "基于文件分类自动添加编目属性，便于内容理解和发现",
	"getPageForTemplate":   "自动识别文件页数后，填入编目模板中的相应字段",
	"regularDeleteFiles":   "每周定时清理创建时间超过一年的文件",
	"docSummary":           "使用大模型提取文档关键内容自动生成发布文件",
	"meetingSummary":       "使用大模型总结会议音频文件后提取会议纪要并生成文件",
}

const (
	// DocNotFound 文档未找到
	DocNotFound = 404002006
	// DocNoPerm ��档无权限
	DocNoPerm = 403001002
	// TokenExpired token过期
	TokenExpired = 401001001
)

// CreateDagReq create automation struct
type CreateDagReq struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Steps       []entity.Step     `json:"steps"`
	Shortcuts   []string          `json:"shortcuts"`
	Accessors   []entity.Accessor `json:"accessors"`
	AppInfo     entity.AppInfo    `json:"appinfo"`
	Emails      []string          `json:"emails"`
	Template    string            `json:"template"`
	CreateBy    string            `json:"create_by"`
	Published   bool              `json:"published"`
	Type        string            `json:"type,omitempty"`
	Version     entity.Version    `json:"version"`
	ChangeLog   string            `json:"change_log"`
	BizDomainID string            `json:"-"`
}

// OptionalUpdateDagReq update dag fields which is user want to update
type OptionalUpdateDagReq struct {
	Title       *string            `json:"title"`
	Description *string            `json:"description"`
	Status      *string            `json:"status"`
	Steps       *[]entity.Step     `json:"steps"`
	Shortcuts   *[]string          `json:"shortcuts"`
	Accessors   *[]entity.Accessor `json:"accessors"`
	AppInfo     *entity.AppInfo    `json:"appinfo"`
	Emails      *[]string          `json:"emails"`
	Published   *bool              `json:"published"`
	Version     *entity.Version    `json:"version"`
	ChangeLog   string             `json:"change_log"`
}

// DagInfo dag detail info
type DagInfo struct {
	ID            string                `json:"id"`
	Title         string                `json:"title"`
	Description   string                `json:"description"`
	Status        string                `json:"status"`
	Steps         []entity.Step         `json:"steps"`
	CreatedAt     int64                 `json:"created_at"`
	UpdatedAt     int64                 `json:"updated_at"`
	Shortcuts     *[]string             `json:"shortcuts"`
	Accessors     *[]entity.Accessor    `json:"accessors"`
	Cron          string                `json:"cron"`
	Published     bool                  `json:"published"`
	Type          string                `json:"type,omitempty"`
	TriggerConfig *entity.TriggerConfig `json:"trigger_config,omitempty"`
	ExecMode      string                `json:"exec_mode,omitempty"`
	Category      string                `json:"category,omitempty"`
	OutPuts       []*entity.OutPut      `json:"outputs,omitempty"`
	UserID        string                `json:"userid,omitempty"`
	DeBugID       string                `json:"debug_id,omitempty"`
}

// DagSimpleInfo  dag simple info
type DagSimpleInfo struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Actions     []string       `json:"actions"`
	TriggerStep *entity.Step   `json:"trigger_step,omitempty"`
	CreatedAt   int64          `json:"created_at"`
	UpdatedAt   int64          `json:"updated_at"`
	Status      string         `json:"status"`
	UserID      string         `json:"userid,omitempty"`
	Creator     string         `json:"creator,omitempty"`
	Trigger     entity.Trigger `json:"trigger,omitempty"`
	Type        string         `json:"type,omitempty"`
	VersionID   string         `json:"version_id,omitempty"`
}

// DagInfoOption  dag info
type DagInfoOption struct {
	ID           string                        `json:"id"`
	Name         string                        `json:"name,omitempty"`
	Description  string                        `json:"description,omitempty"`
	Status       string                        `json:"status,omitempty"`
	TmpCreatedAt int64                         `json:"createdAt,omitempty"`
	TmpUpdatedAt int64                         `json:"updatedAt,omitempty"`
	CreatedAt    int64                         `json:"created_at,omitempty"`
	UpdatedAt    int64                         `json:"updated_at,omitempty"`
	Type         string                        `json:"type,omitempty"`
	UserID       string                        `json:"userid,omitempty"`
	Creator      *drivenadapters.UserAttribute `json:"creator,omitempty"`
	Trigger      entity.Trigger                `json:"trigger,omitempty"`
}

type DagInfoOptionReq struct {
	Method string   `json:"method"`
	DagIDs []string `json:"dag_ids"`
}

// DagInsStatusReq dag instance status struct
type DagInsStatusReq struct {
	Status string `json:"status"`
}

// DagInstanceRunInfo dag ins info
type DagInstanceRunInfo struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	StartedAt int64  `json:"started_at"`
	EndedAt   int64  `json:"ended_at"`
	Source    any    `json:"source"`
	Reason    any    `json:"reason"`
}

// Progress dag ins run detail info
type Progress struct {
	Total   int64 `json:"total"`
	Success int64 `json:"success"`
	Failed  int64 `json:"failed"`
}

// DagInstanceRunList  dag ins run list
type DagInstanceRunList struct {
	DagInstanceRunInfo []*DagInstanceRunInfo `json:"resultes"`
	Progress           *Progress             `json:"progress"`
}

// TaskInstanceRunInfo task ins run info
type TaskInstanceRunInfo struct {
	ID             string               `json:"id"`
	Name           string               `json:"name,omitempty"`
	Operator       string               `json:"operator"`
	StartedAt      int64                `json:"started_at"`
	UpdatedAt      int64                `json:"updated_at"`
	Status         string               `json:"status"`
	Inputs         interface{}          `json:"inputs"`
	Outputs        interface{}          `json:"outputs"`
	TaskID         string               `json:"taskId"`
	LastModifiedAt int64                `json:"last_modified_at"`
	MetaData       *entity.TaskMetaData `json:"metadata,omitempty"`
}

// RunInstanceWithFormReq 表单运行流程时参数
type RunInstanceWithFormReq struct {
	Data map[string]interface{} `json:"data"`
}

// AuditorInfo 审核员信息
type AuditorInfo struct {
	ApplyID  string   `json:"apply_id"` // 申请ID，返回发起申请时传入的申请ID
	Auditors []string `json:"auditors"` // 当前匹配到的审核员列表
}

// ListDagsConfig 列举dags参数
type ListDagsConfig struct {
	IsShared bool `json:"is_shared"`
	// Scope 此参数用于子流程调用时列举用户可展示的所有流程信息(自己创建或他人分享)
	Scope string `json:"scope"` // "all"
}

// TriggerConfig 触发器节点配置信息
type TriggerConfig struct {
	ID       string      `json:"id,omitempty"`
	Operator string      `json:"operator,omitempty"`
	Params   interface{} `json:"parameters,omitempty"`
	Result   interface{} `json:"result,omitempty"`
}

type QueryParams struct {
	Page           int64    `json:"page"`
	Limit          int64    `json:"limit"`
	SortBy         string   `json:"sortby"`
	Order          string   `json:"order"`
	TriggerType    string   `json:"trigger_type"`
	TriggerTypes   []string `json:"trigger_types"`
	Type           string   `json:"type"`
	KeyWord        string   `json:"keyword"`
	BizDomainID    string   `json:"-"`
	TriggerExclude []string `json:"-"`
	Accessors      []string `json:"-"`
	UserID         string   `json:"-"`
	Scope          string   `json:"-"`
}

// RunDagParams 运行数据流请求参数
type RunDagParams struct {
	ID                  string
	VersionID           string
	FormData            map[string]any
	ParentDagInstanceID string
	CallBack            string
}

// DocMsg doc msg
type DocMsg = common.DocMsg

// S3ValidationResult S3配置验证结果
type S3ValidationResult struct {
	BucketExists   bool   `json:"bucket_exists"`
	PathAccessible bool   `json:"path_accessible"`
	FileCount      int    `json:"file_count"`
	Message        string `json:"message,omitempty"`
}

// FlowFileInfo 文件信息
type FlowFileInfo struct {
	FileID        string `json:"file_id"`
	DocID         string `json:"docid"`
	DagID         string `json:"dag_id,omitempty"`
	DagInstanceID string `json:"dag_instance_id,omitempty"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Size          int64  `json:"size"`
	ContentType   string `json:"content_type"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at,omitempty"`
}

// FlowFileDownloadInfo 文件下载信息
type FlowFileDownloadInfo struct {
	FileID string `json:"file_id"`
	DocID  string `json:"docid"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Size   int64  `json:"size"`
}

// MgntHandler mgnt interface method
type MgntHandler interface { //nolint
	CreateDag(ctx context.Context, param *CreateDagReq, userInfo *drivenadapters.UserInfo) (string, error)
	UpdateDag(ctx context.Context, dagID string, param *OptionalUpdateDagReq, userInfo *drivenadapters.UserInfo) error
	GetDagByID(ctx context.Context, dagID, versionID, bizDomainID string, userInfo *drivenadapters.UserInfo) (*DagInfo, error)
	DeleteDagByID(ctx context.Context, dagID, bizDomainID string, userInfo *drivenadapters.UserInfo) error
	ListDag(ctx context.Context, param QueryParams, userInfo *drivenadapters.UserInfo, config *ListDagsConfig) ([]*DagSimpleInfo, int64, error)
	ListDagV2(ctx context.Context, param QueryParams, userInfo *drivenadapters.UserInfo) ([]*DagSimpleInfo, int64, error)
	ListDagByFields(ctx context.Context, filter bson.M, opt options.FindOptions) ([]*DagSimpleInfo, int64, error)
	RunInstance(ctx context.Context, params *RunDagParams, userInfo *drivenadapters.UserInfo) error
	RunFormInstance(ctx context.Context, params *RunDagParams, userInfo *drivenadapters.UserInfo) (string, error)
	HandleDocEvent(ctx context.Context, msg *DocMsg, topic string) error
	CancelRunningInstance(ctx context.Context, id string, dagInsReq *DagInsStatusReq, userInfo *drivenadapters.UserInfo) error
	GetSuggestDagName(ctx context.Context, name string, userInfo *drivenadapters.UserInfo) (string, error)
	ListDagInstance(ctx context.Context, dagID string, param map[string]interface{}, userInfo *drivenadapters.UserInfo) (*DagInstanceRunList, int64, error)
	ListTaskInstance(ctx context.Context, dagID, dagInstanceID string, page, limit int64, userInfo *drivenadapters.UserInfo) ([]*TaskInstanceRunInfo, int64, error)
	ListActions(ctx context.Context, userInfo *drivenadapters.UserInfo) ([]map[string]interface{}, error)
	ContinueBlockInstances(ctx context.Context, blockedTaskIDs []string, res map[string]interface{}, status entity.TaskInstanceStatus) error
	HandleAuditorsMacth(ctx context.Context, msg *AuditorInfo) error
	HandleUserInfoEvent(ctx context.Context, msg *common.UserInfoMsg, topic string) error
	HandleKCUserInfoEvent(ctx context.Context, msg *common.UserInfoMsg) error
	HandleTagInfoChangeEvent(ctx context.Context, tag *common.TagInfo, msg []byte, topic string) error
	HandleTagTreeCreateEvent(ctx context.Context, msg *common.TagInfo, topic string) error
	CreateSecurityPolicyFlow(ctx context.Context, param *CreateFlowParams, userInfo *drivenadapters.UserInfo) (string, error)
	UpdateSecurityPolicyFlow(ctx context.Context, dagID string, steps []entity.Step, userInfo *drivenadapters.UserInfo) error
	DeleteSecurityPolicyFlow(ctx context.Context, dagID string, userInfo *drivenadapters.UserInfo) error
	GetSecurityPolicyFlowByID(ctx context.Context, dagID string) (Flow, error)
	StartSecurityPolicyFlowProc(ctx context.Context, params ProcParams) (string, error)
	StopSecurityPolicyFlowProc(ctx context.Context, pid string, userInfo *drivenadapters.UserInfo) error
	RunCronInstance(ctx context.Context, id, webhook string) error
	UpdateTaskResults(ctx context.Context, taskId string, results map[string]interface{}, userInfo *drivenadapters.UserInfo) error
	RunInstanceWithDoc(ctx context.Context, id string, params RunWithDocParams, userInfo *drivenadapters.UserInfo) error
	ListModelBindDags(ctx context.Context, id, userID string) ([]*DagSimpleInfo, error)
	GetDagTriggerConfig(ctx context.Context, taskInsID, typeBy string, userInfo *drivenadapters.UserInfo) (TriggerConfig, error)
	CallAgent(ctx context.Context, name string, inputs map[string]interface{}, options *drivenadapters.CallAgentOptions, token string) (res *drivenadapters.CallAgentRes, ch chan *drivenadapters.CallAgentRes, err error)
	GetAgents(ctx context.Context) (res []*rds.AgentModel, err error)
	CreateDataFlow(ctx context.Context, param *CreateDataFlowReq, userInfo *drivenadapters.UserInfo) (string, error)
	UpdateDataFlow(ctx context.Context, dagID string, param *UpdateDataFlowReq, userInfo *drivenadapters.UserInfo) error
	DeleteDataFlow(ctx context.Context, dagID, bizDomainID string, userInfo *drivenadapters.UserInfo) error
	GetDagInstanceCount(ctx context.Context, dagID string, param map[string]interface{}, userInfo *drivenadapters.UserInfo) (int64, error)
	ListDagInstanceV2(ctx context.Context, dagID string, param map[string]interface{}, userInfo *drivenadapters.UserInfo) ([]*DagInstanceRunInfo, int64, error)
	RetryDagInstance(ctx context.Context, dagInsID string, userDetail *drivenadapters.UserInfo) error
	RunFormInstanceV2(ctx context.Context, id string, formData map[string]interface{}, successCallback string, errorCallback string, userInfo *drivenadapters.UserInfo) (dagIns *entity.DagInstance, vmIns *mod.VMExt, err error)
	RunOperator(ctx context.Context, id string, formData map[string]any, successCallback, errorCallback string, parentDagInsID string, userInfo *drivenadapters.UserInfo) (dagIns *entity.DagInstance, vmIns *mod.VMExt, err error)
	GetDagInstanceResultVM(ctx context.Context, id string, userInfo *drivenadapters.UserInfo) (state.State, any, error)
	BatchGetDag(ctx context.Context, dagIDs []string, fields string, userInfo *drivenadapters.UserInfo) ([]*DagInfoOption, error)
	SingleDeBug(ctx context.Context, params SingleDeBugReq, userInfo *drivenadapters.UserInfo) (string, error)
	SingleDeBugResult(ctx context.Context, id string) (entity.TaskInstanceStatus, any, error)
	FullDebug(ctx context.Context, params FullDeBugReq, userInfo *drivenadapters.UserInfo) (string, string, error)

	CreateComboOperator(ctx context.Context, param *ComboOperatorReq, userInfo *drivenadapters.UserInfo) (string, string, error)
	UpdateComboOperator(ctx context.Context, param *OptionalComboOperatorReq, userInfo *drivenadapters.UserInfo) error
	ListComboOperator(ctx context.Context, query map[string]interface{}, userInfo *drivenadapters.UserInfo) (*ComboOperatorList, error)
	DeleteComboOperator(ctx context.Context, operatorID string) error
	ExportOperator(ctx context.Context, dagIDs []string) (ExportOperator, error)
	ImportOperator(ctx context.Context, params *ImportOperatorReq, userInfo *drivenadapters.UserInfo) error

	// 业务域数据迁移列举接口
	ListHistoryData(ctx context.Context, page, limit int64) (HistoryDataResp, error)
	ListDagInstanceEvents(ctx context.Context, dagID, dagInsID string, offset, limit int, userInfo *drivenadapters.UserInfo) (
		logs []*entity.DagInstanceEvent, dagIns *entity.DagInstance, total int, next int, err error)

	// S3配置验证接口
	ValidateS3Config(ctx context.Context, bucket, path, mode string) (*S3ValidationResult, error)

	// S3文件管理接口
	UploadS3File(ctx context.Context, dagID string, fileHeader *multipart.FileHeader, userInfo *drivenadapters.UserInfo) (*S3DataItem, error)
	ListS3Files(ctx context.Context, dagID string, userInfo *drivenadapters.UserInfo) ([]*S3DataItem, error)
	DeleteS3File(ctx context.Context, dagID, key string, userInfo *drivenadapters.UserInfo) error
	MoveS3Files(ctx context.Context, sources []string, targetDagID string) ([]string, error)
	GetS3FileDownloadURL(ctx context.Context, dagID, key string, userInfo *drivenadapters.UserInfo) (string, error)

	// Dataflow 文件子系统接口
	TriggerDataflowDoc(ctx context.Context, params *TriggerDataflowDocParams, userInfo *drivenadapters.UserInfo) (*TriggerDataflowDocResult, error)
	CompleteDataflowDocUpload(ctx context.Context, params *CompleteDataflowDocUploadParams, userInfo *drivenadapters.UserInfo) (*CompleteDataflowDocUploadResult, error)

	// Dataflow 文件访问接口
	ListFlowFiles(ctx context.Context, dagInstanceID string, userInfo *drivenadapters.UserInfo) ([]*FlowFileInfo, error)
	GetFlowFile(ctx context.Context, fileID string, userInfo *drivenadapters.UserInfo) (*FlowFileInfo, error)
	GetFlowFileDownloadURL(ctx context.Context, fileID string, userInfo *drivenadapters.UserInfo) (*FlowFileDownloadInfo, error)
}

var (
	mOnce sync.Once
	m     MgntHandler
)

type mgnt struct {
	efast             drivenadapters.Efast
	mongo             mod.Store
	exec              mod.Executor
	usermgnt          drivenadapters.UserManagement
	docshare          drivenadapters.DocShare
	tika              drivenadapters.Tika
	appstore          drivenadapters.Appstore
	ecron             drivenadapters.ECron
	ecotag            drivenadapters.EcoTag
	kcmc              drivenadapters.Kcmc
	ad                drivenadapters.AnyData
	config            common.Config
	paramRender       *render.TplRender
	dependency        dependency.Repo
	executeMethods    entity.ExecuteMethods
	personalConfig    drivenadapters.PersonalConfig
	executor          rds.ExecutorDao
	admin             rds.ContentAmdinDao
	extData           rds.DagInstanceExtDataDao
	eventRepository   rds.DagInstanceEventRepository
	taskTimeoutConfig *common.TimeoutConfig
	mq                mod.MQHandler
	logger            drivenadapters.Logger
	agent             rds.AgentDao
	operator          drivenadapters.AgentOperatorIntegration
	permPolicy        perm.PermPolicyHandler
	uniquery          drivenadapters.UniqueryDriven
	rdb               cstore.RDB
	permCheck         perm.PermCheckerService
	memoryCache       cstore.LocalCache
	pool              *threadPool.PoolManager
	bizDomain         drivenadapters.BusinessDomain
	s3Adapter         drivenadapters.S3Adapter  // S3适配器
	ossGateway        drivenadapters.OssGateWay // OssGateway文件存储
	flowFileDao       rds.FlowFileDao           // FlowFile DAO
	flowStorageDao    rds.FlowStorageDao        // FlowStorage DAO
}

// NewMgnt mgnt instance
func NewMgnt() MgntHandler {
	mOnce.Do(func() {
		em := entity.ExecuteMethods{
			Publish: mod.NewMQHandler().Publish,
		}

		mIns := &mgnt{
			efast:             drivenadapters.NewEfast(),
			mongo:             mod.GetStore(),
			exec:              mod.GetExecutor(),
			usermgnt:          drivenadapters.NewUserManagement(),
			tika:              drivenadapters.NewTika(),
			config:            *common.NewConfig(),
			paramRender:       render.NewTplRender(),
			docshare:          drivenadapters.NewDocShare(),
			appstore:          drivenadapters.NewAppStore(),
			ecron:             drivenadapters.NewECron(),
			ecotag:            drivenadapters.NewEcoTag(),
			kcmc:              drivenadapters.NewKcmc(),
			ad:                drivenadapters.NewAnyData(),
			dependency:        dependency.NewDriven(),
			executeMethods:    em,
			personalConfig:    drivenadapters.NewPersonalConfig(),
			executor:          rds.GetExecutorDao(),
			admin:             rds.GetContentAdminDao(),
			extData:           rds.GetDagInstanceExtDataDao(),
			eventRepository:   rds.GetDagInstanceEventRepository(),
			taskTimeoutConfig: common.NewTimeoutConfig(),
			mq:                mod.NewMQHandler(),
			agent:             rds.GetAgentDao(),
			operator:          drivenadapters.NewAgentOperatorIntegration(),
			logger:            drivenadapters.NewLogger(),
			permPolicy:        perm.NewPermPolicy(),
			uniquery:          drivenadapters.NewUniquery(),
			rdb:               cstore.NewRedis(),
			permCheck:         perm.NewPermCheckerService(),
			memoryCache: cstore.NewLocalCache(&cstore.Option{
				Expiration:      5 * time.Minute,
				CleanUpInterval: 10 * time.Minute,
			}),
			bizDomain: drivenadapters.NewBusinessDomain(),
			// dagModel:  dagmodel.NewDagRepository(),
		}

		// Initialize OssGateway for Dataflow file subsystem
		mIns.ossGateway = drivenadapters.NewOssGateWay()

		// Initialize FlowFile and FlowStorage DAOs
		mIns.flowFileDao = rds.GetFlowFileDao()
		mIns.flowStorageDao = rds.GetFlowStorageDao()

		// Initialize S3 adapter if configured
		s3Conn := s3.NewS3().GetDefaultConnection()
		if s3Conn != nil {
			s3cfg := &pkgconfig.S3Config{
				Endpoint:        s3Conn.Endpoint,
				Region:          s3Conn.Region,
				AccessKeyID:     s3Conn.AccessKeyID,
				SecretAccessKey: s3Conn.SecretAccessKey,
				BucketName:      s3Conn.BucketName,
				UseSSL:          !s3Conn.DisableHTTPS,
				SkipVerify:      s3Conn.SkipVerify,
			}
			if s3Adapter, err := drivenadapters.NewS3Adapter(s3cfg); err == nil {
				mIns.s3Adapter = s3Adapter
				log.Info("[mgnt.NewMgnt] S3 adapter initialized successfully")
			} else {
				log.Warnf("[mgnt.NewMgnt] Failed to create S3 adapter: %v", err)
			}
		} else {
			log.Warnf("[mgnt.NewMgnt] S3 connection not found or failed to load")
		}

		perm.RegisterChecker(common.DagTypeDataFlow, &perm.DataFlowDagPermChecker{PermPolicy: perm.NewPermPolicy()})
		perm.RegisterChecker(common.DagTypeComboOperator, &perm.ComBoOperatorPermChecker{PermPolicy: perm.NewPermPolicy()})
		perm.RegisterChecker(common.DagTypeDefault, &perm.DefaultDagPermChecker{Usermgnt: drivenadapters.NewUserManagement(), IsAccessible: mIns.isAccessible})

		poolConfig := threadPool.NewPoolConfig(threadPool.WithName("DeBug-Pool"), threadPool.WithCapacity(mIns.config.Server.DebugExecutorCount), threadPool.WithNonblocking(true))
		mIns.pool, _ = threadPool.NewPoolManager(poolConfig)
		RegisterTriggerHandlers(mIns.efast, mIns.usermgnt, mIns.uniquery)

		m = mIns
	})
	return m
}

// CreateDag create a dag
func (m *mgnt) CreateDag(ctx context.Context, param *CreateDagReq, userInfo *drivenadapters.UserInfo) (string, error) {
	var dagID string
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	logger := traceLog.WithContext(ctx)

	// trim space
	param.Title = strings.TrimSpace(param.Title)
	param.Description = strings.TrimSpace(param.Description)
	param.Type = strings.TrimSpace(param.Type)
	// 仅当没有类型时，校验用户创建工作流总数
	if len(param.Type) == 0 {
		total, err := m.mongo.ListDagCount(ctx, &mod.ListDagInput{UserID: userInfo.UserID, BizDomainID: param.BizDomainID})
		if err != nil {
			logger.Warnf("[logic.CreateDag] GetDagCount err, detail: %s", err.Error())
			return dagID, ierrors.NewIError(ierrors.InternalError, "", nil)
		}
		if total >= taskTotal {
			return dagID, ierrors.NewIError(ierrors.Forbidden, ierrors.NumberOfTasksLimited, map[string]string{"task": "total number of tasks cannot exceed 50"})
		}
	}

	userDetail, err := m.usermgnt.GetUserInfoByType(userInfo.UserID, userInfo.AccountType)
	if err != nil {
		logger.Warnf("[logic.CreateDag] GetUserInfoByType err, detail: %s", err.Error())
		return dagID, ierrors.NewIError(ierrors.ErrorDepencyService, "", err.Error())
	}

	userInfo.UserName = userDetail.UserName
	isAdminRole := utils.IsAdminRole(userDetail.Roles)

	var tasks = make([]entity.Task, 0)
	var stepList = make([]map[string]interface{}, 0)
	steps := make([]entity.Step, len(param.Steps))
	copy(steps, param.Steps)
	m.buildTasks(&steps[0], steps, &tasks, nil, &stepList, nil, nil)

	subDagIDs, err := m.validSubFlowInSteps(ctx, "", stepList)
	if err != nil {
		return dagID, err
	}

	err = m.validSteps(&Validate{
		Ctx:         ctx,
		Steps:       stepList,
		IsAdminRole: isAdminRole,
		UserInfo:    userInfo,
		ErrType:     ErrTypeV1,
		ParseFunc:   common.JSONSchemaValid,
	}).BuildError()
	if err != nil {
		return dagID, err
	}
	trigger := m.getTriggerType(param.Steps[0].Operator)
	// param.Steps[0].
	dag := &entity.Dag{
		UserID: userInfo.UserID,
		Name:   param.Title,
		Vars: entity.DagVars{
			"userid": {DefaultValue: userInfo.UserID},
			"docid":  {DefaultValue: ""},
		},
		Trigger: trigger,
		// Cron:   "* * * * *",
		Tasks:       tasks,
		Steps:       param.Steps,
		Description: param.Description,
		Status:      entity.DagStatusNormal,
		Shortcuts:   param.Shortcuts,
		Accessors:   param.Accessors,
		Emails:      param.Emails,
		Priority:    common.PriorityLowest,
		Template:    param.Template,
		Published:   param.Published,
		Type:        param.Type,
		BizDomainID: param.BizDomainID,
		ModifyBy:    userInfo.UserID,
		SubIDs:      subDagIDs,
	}

	if param.Status == common.StoppedStatus {
		dag.Status = entity.DagStatusStopped
	}

	// check duplicated name
	dagInfo, err := m.mongo.GetDagByFields(ctx, map[string]interface{}{"name": param.Title, "userid": userInfo.UserID})
	if err != nil && !ierrors.IsNotFoundErr(err) {
		logger.Warnf("[logic.CreateDag] GetDagByFields err, deail: %s", err.Error())
		return dagID, ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	if dagInfo != nil {
		return dagID, ierrors.NewIError(ierrors.DuplicatedName, "", map[string]string{"title": param.Title})
	}

	if isAdminRole && param.AppInfo.Enable {
		dag.AppInfo = param.AppInfo
		dag.Priority = common.PriorityHigh
	}

	if !isAdminRole && param.AppInfo.Enable {
		return dagID, ierrors.NewIError(ierrors.Forbidden, "", map[string]string{"title": param.Title})
	}

	dagVersions, err := m.buildDagVersions(&BuildDagVersionParams{
		Dag:        dag,
		CurVersion: &param.Version,
		ChangeLog:  param.ChangeLog,
		UserID:     userInfo.UserID,
		IsCreate:   true,
	})
	if err != nil {
		return "", err
	}

	dag.SetCheckRootNode(func(t []entity.Task) error {
		_, bErr := mod.BuildRootNode(mod.MapTasksToGetter(dag.Tasks))
		if bErr != nil {
			return bErr
		}
		return nil
	})
	err = m.mongo.WithTransaction(ctx, func(nctx context.Context, ms mod.Store) error {
		dagID, err = ms.CreateDag(nctx, dag)
		if err != nil {
			logger.Warnf("[logic.CreateDag] CreateDag err, deail: %s", err.Error())
			return ierrors.NewIError(ierrors.InternalError, "", nil)
		}

		config, _ := json.Marshal(dag)
		dagVersion := dagVersions[0]
		dagVersion.DagID = dagID
		dagVersion.Config = entity.Config(config)
		_, err = ms.CreateDagVersion(nctx, dagVersion)
		if err != nil {
			log.Warnf("[logic.CreateDag] CreateDagVersion err, detail: %s", err.Error())
			return err
		}

		return nil
	})
	if err != nil {
		return "", ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	bizDomainParams := drivenadapters.BizDomainResourceParams{
		BizDomainID:  param.BizDomainID,
		ResourceID:   fmt.Sprintf("%s:%s", dagID, common.DagTypeDefault),
		ResourceType: perm.DataFlowResourceType,
	}
	err = m.bizDomain.BindResourceInternal(ctx, bizDomainParams)
	if err != nil {
		logger.Warnf("[logic.CreateDag] BindResourceInternal err, deail: %s", err.Error())
		dErr := m.mongo.DeleteDag(ctx, dagID)
		if dErr != nil {
			logger.Warnf("[logic.CreateDag] BindResourceInternal failed, DeleteDag err, deail: %s", dErr.Error())
		}
		return "", ierrors.NewIError(ierrors.InternalError, ierrors.ErrorDepencyService, err.Error())
	}

	if dag.Status == entity.DagStatusNormal && trigger == entity.TriggerCron {
		timely := param.Steps[0].Cron
		param, ok := param.Steps[0].Parameters["cron"]
		if ok && param != "" {
			timely = param.(string)
		}
		jobID, _, err := m.ecron.RegisterCronJob(
			ctx,
			fmt.Sprintf("auto_%s", dagID),
			fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/cron/%s", m.config.ContentAutomation.PrivateHost, m.config.ContentAutomation.PrivatePort, dagID),
			timely,
		)
		if err != nil {
			logger.Warnf("[logic.CreateDag] RegisterCronJob err, deail: %s", err.Error())
			if berr := m.mongo.BatchDeleteDagWithTransaction(ctx, []string{dagID}); berr != nil {
				logger.Warnf("[logic.CreateDag] BatchDeleteDagWithTransaction err, deail: %s", err.Error())
			}

			if err := m.extData.Remove(ctx, &rds.ExtDataQueryOptions{
				DagID: dagID,
			}); err != nil {
				logger.Warnf("[logic.CreateDag] Remove extData err, deail: %s", err.Error())
			}
			_ = m.bizDomain.UnBindResourceInternal(ctx, bizDomainParams)

			return dagID, ierrors.NewIError(ierrors.InternalError, "", nil)
		}

		dag.Cron = jobID
		err = m.mongo.UpdateDag(ctx, dag)
		if err != nil {
			logger.Warnf("[logic.CreateDag] UpdateDag with cron job id err, deail: %s", err.Error())

			return dagID, ierrors.NewIError(ierrors.InternalError, "", nil)
		}
	}

	go func() {
		detail, extMsg := common.GetLogBody(common.CreateTask, []interface{}{dag.Name},
			[]interface{}{})
		tmp := templateName[dag.Template]
		if tmp == "" {
			tmp = "null"
		}

		createBy := common.CreateByDirectName
		switch param.CreateBy {
		case common.CreateByTemplate:
			createBy = common.CreateByTemplateName
		case common.CreateByLocal:
			createBy = common.CreateByLocalName
		}

		did, _ := strconv.Atoi(dagID)
		object := map[string]interface{}{
			"type":      trigger,
			"id":        dagID,
			"did":       did,
			"name":      dag.Name,
			"creator":   userInfo.UserID,
			"priority":  dag.Priority,
			"template":  tmp,
			"create_by": createBy,
		}

		userInfo.Type = common.User.ToString()
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		m.logger.Log(drivenadapters.LogTypeASOperationLog, &drivenadapters.BuildARLogParams{
			Operation:   common.ArLogCreateDag,
			Description: detail,
			UserInfo:    userInfo,
			Object:      object,
		}, write)

		logger.Infof("detail: %s, extMsg: %s", detail, extMsg)
		m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: dagID,
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, write)
	}()

	return dagID, nil
}

// UpdateDag update dag if exisit
func (m *mgnt) UpdateDag(ctx context.Context, dagID string, param *OptionalUpdateDagReq, userInfo *drivenadapters.UserInfo) error { //nolint
	var err error
	var stopRunningTask bool
	var appCountInfoMap = make(map[string]string, 0)
	query := map[string]interface{}{"_id": dagID, "userid": userInfo.UserID}

	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// check dag whether exisis
	dag, err := m.mongo.GetDagByFields(ctx, query)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": dagID})
		}
		log.Warnf("[logic.UpdateDag] GetDagByFields err, query: %v, deail: %s", query, err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	trigger := m.getTriggerType(dag.Steps[0].Operator)

	dagBytes, err := json.Marshal(dag)
	if err != nil {
		log.Warnf("[logic.UpdateDataFlow] Marshal err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err.Error())
	}

	userDetail, err := m.usermgnt.GetUserInfoByType(userInfo.UserID, userInfo.AccountType)
	if err != nil {
		log.Warnf("[logic.UpdateDag] GetUserInfoByType err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.ErrorDepencyService, "", err.Error())
	}

	isAdminRole := utils.IsAdminRole(userDetail.Roles)

	if isAdminRole && param.AppInfo != nil {
		dag.AppInfo = *param.AppInfo
		dag.Priority = common.PriorityHigh
	}

	if param.Title != nil {
		title := strings.TrimSpace(*param.Title)
		_query := map[string]interface{}{
			"userid": userInfo.UserID,
			"name":   title,
			"removed": bson.M{
				"$ne": true,
			},
		}
		_dag, err := m.mongo.GetDagByFields(ctx, _query) //nolint
		if err != nil {
			if !ierrors.IsNotFoundErr(err) {
				log.Warnf("[logic.UpdateDag] GetDagByFields err, query: %v, deail: %s", _query, err.Error())
				return ierrors.NewIError(ierrors.InternalError, "", nil)
			}
		}
		if err == nil && _dag.ID != dagID {
			return ierrors.NewIError(ierrors.DuplicatedName, "", map[string]string{"title": title})
		}
		dag.Name = title
	}

	if param.Description != nil {
		dag.Description = strings.TrimSpace(*param.Description)
	}

	if param.Shortcuts != nil {
		dag.Shortcuts = *param.Shortcuts
	}

	if param.Accessors != nil {
		dag.Accessors = *param.Accessors
	}

	if param.Steps != nil {
		resetPermApplyStepAppPwd(appCountInfoMap, dag.Steps, *param.Steps)
		var tasks = make([]entity.Task, 0)
		var stepList = make([]map[string]interface{}, 0)
		dag.Steps = *param.Steps
		steps := make([]entity.Step, len(*param.Steps))
		copy(steps, *param.Steps)
		m.buildTasks(&steps[0], steps, &tasks, nil, &stepList, nil, nil)

		dag.SubIDs, err = m.validSubFlowInSteps(ctx, dagID, stepList)
		if err != nil {
			return err
		}

		// steps paramas validate
		err = m.validSteps(&Validate{
			Ctx:         ctx,
			Steps:       stepList,
			IsAdminRole: isAdminRole,
			UserInfo:    userInfo,
			ErrType:     ErrTypeV1,
			ParseFunc:   common.JSONSchemaValid,
		}).BuildError()
		if err != nil {
			return err
		}
		dag.Tasks = tasks
		trigger = m.getTriggerType(steps[0].Operator)
		dag.Trigger = trigger

		timely := steps[0].Cron
		param, ok := steps[0].Parameters["cron"]
		if ok && param != "" {
			timely = param.(string)
		}

		// 更新流程后，可能涉及到新建/更新和删除定时任务
		if dag.Cron == "" && timely != "" && trigger == entity.TriggerCron {
			// 需新建定时任务
			jobID, exist, err := m.ecron.RegisterCronJob(ctx, fmt.Sprintf("auto_%s", dagID), fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/cron/%s", m.config.ContentAutomation.PrivateHost, m.config.ContentAutomation.PrivatePort, dagID), timely)
			if err != nil {
				if !exist {
					log.Warnf("[logic.UpdateDag] RegisterCronJob err, deail: %s", err.Error())
					return ierrors.NewIError(ierrors.InternalError, "", nil)
				}
			} else {
				dag.Cron = jobID
			}
		} else if dag.Cron != "" && timely == "" && trigger != entity.TriggerCron {
			// 需删除定时任务
			err = m.ecron.DeleteEcronJob(ctx, dag.Cron)
			if err != nil {
				log.Warnf("[logic.UpdateDag] DeleteEcronJob err, detail: %s", err.Error())
				return ierrors.NewIError(ierrors.InternalError, "", nil)
			}
			dag.Cron = ""
		} else if dag.Cron != "" && timely != "" && trigger == entity.TriggerCron {
			// 需要更新定时任务
			err = m.ecron.UpdateCronJob(ctx, fmt.Sprintf("auto_%s", dagID), fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/cron/%s", m.config.ContentAutomation.PrivateHost, m.config.ContentAutomation.PrivatePort, dagID), timely, dag.Cron)
			if err != nil {
				if httpErr, ok := err.(libErrors.ExHTTPError); ok && httpErr.Status == http.StatusNotFound {
					// 需新建定时任务
					jobID, exist, err := m.ecron.RegisterCronJob(ctx, fmt.Sprintf("auto_%s", dagID), fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/cron/%s", m.config.ContentAutomation.PrivateHost, m.config.ContentAutomation.PrivatePort, dagID), timely)
					if err != nil {
						if !exist {
							log.Warnf("[logic.UpdateDag] RegisterCronJob err, deail: %s", err.Error())
							return ierrors.NewIError(ierrors.InternalError, "", nil)
						}
					} else {
						dag.Cron = jobID
					}
				} else {
					log.Warnf("[logic.UpdateDag] UpdateCronJob err, detail: %s", err.Error())
					return ierrors.NewIError(ierrors.InternalError, "", nil)
				}
			}
		}
	}

	if param.Status != nil {
		dag.Status = entity.DagStatusNormal
		if *param.Status == common.StoppedStatus {
			dag.Status = entity.DagStatusStopped
			stopRunningTask = true
		}
	}

	if param.Emails != nil {
		dag.Emails = *param.Emails
	}

	if param.Published != nil {
		dag.Published = *param.Published
	}

	var dagVersions []*entity.DagVersion
	dagVersions, err = m.buildDagVersions(&BuildDagVersionParams{
		OldDagBytes: dagBytes,
		Dag:         dag,
		CurVersion:  param.Version,
		ChangeLog:   param.ChangeLog,
		UserID:      userInfo.UserID,
		IsCreate:    false,
	})
	if err != nil {
		return err
	}

	err = m.mongo.WithTransaction(ctx, func(nctx context.Context, ms mod.Store) error {
		// 更新dag
		if err := ms.UpdateDag(nctx, dag); err != nil {
			log.Warnf("[logic.UpdateDag] UpdateDag err, deail: %s", err.Error())
			return ierrors.NewIError(ierrors.InternalError, "", nil)
		}

		for _, dagVersion := range dagVersions {
			_, err = ms.CreateDagVersion(nctx, dagVersion)
			if err != nil {
				log.Warnf("[logic.UpdateDag] CreateDagVersion err, detail: %s", err.Error())
				return err
			}
		}

		return nil
	})
	if err != nil {
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// goroutine exec avoid block
	// stopped status terminal all running dagIns
	go func(stopRunningTask bool) {
		if stopRunningTask {
			gctx := context.Background()
			var input = &mod.ListDagInstanceInput{
				DagIDs: []string{dagID},
				Status: []entity.DagInstanceStatus{entity.DagInstanceStatusRunning,
					entity.DagInstanceStatusScheduled,
					entity.DagInstanceStatusInit,
					entity.DagInstanceStatusBlocked,
				},
				SelectField: []string{"_id"},
			}
			dagInsList, err := m.mongo.ListDagInstance(gctx, input)
			if err != nil {
				log.Warnf("[logic.UpdateDag] ListDagInstance err, detail: %s", err.Error())
				return
			}
			var dagInsArr = make([]*entity.DagInstance, 0)
			for _, dagIns := range dagInsList {
				_dagIns := *dagIns
				_dagIns.Status = entity.DagInstanceStatusCancled
				dagInsArr = append(dagInsArr, &_dagIns)
			}
			// update dagIns status
			err = m.mongo.BatchUpdateDagIns(gctx, dagInsArr)
			if err != nil {
				log.Warnf("[logic.UpdateDag] BatchUpdateDagIns err, detail: %s", err.Error())
				return
			}
		}
	}(stopRunningTask)
	go func() {
		detail, extMsg := common.GetLogBody(common.UpdateTask, []interface{}{dag.Name},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
			UserInfo: userInfo,
			Msg:      detail,
			ExtMsg:   extMsg,
			OutBizID: dagID,
			LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		}, &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish})

	}()
	return nil
}

// GetDagByID get dag by dagid
func (m *mgnt) GetDagByID(ctx context.Context, dagID, versionID, bizDomainID string, userInfo *drivenadapters.UserInfo) (*DagInfo, error) {
	var err error
	var dag = new(DagInfo)

	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 检查 dag 是否存在
	DBDag, err := m.mongo.GetDag(ctx, dagID)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return nil, ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": dagID})
		}
		log.Warnf("[logic.UpdateDag] GetDag err, deail: %s", err.Error())
		return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	if DBDag.Type != common.DagTypeComboOperator {
		exist, err := m.bizDomain.CheckerResource(ctx, drivenadapters.BizDomainResourceParams{
			BizDomainID:  bizDomainID,
			ResourceID:   fmt.Sprintf("%s:%s", DBDag.ID, utils.IfNot(DBDag.Type == "", common.DagTypeDefault, DBDag.Type)),
			ResourceType: perm.DataFlowResourceType,
		}, userInfo.TokenID)
		if err != nil {
			return nil, err
		}

		if !exist {
			return nil, ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": dagID})
		}
	}

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow:      {perm.ViewOperation},
			common.DagTypeComboOperator: {perm.ViewOperation},
			common.DagTypeDefault:       {perm.OldAdminOperation, perm.OldShareOperation},
		},
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dagID, userInfo, opMap)
	if err != nil {
		return nil, err
	}

	// 获取Dag历史版本记录详情
	if versionID != "" && DBDag.VersionID != versionID {
		historyDag, err := m.mongo.GetHistoryDagByVersionID(ctx, dagID, versionID)
		if err != nil {
			if ierrors.IsNotFoundErr(err) {
				return nil, ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": dagID, "version": versionID})
			}
			log.Warnf("[logic.GetDagByID] GetHistoryDagByVersionID err, deail: %s", err.Error())
			return nil, ierrors.NewIError(ierrors.InternalError, "", nil)
		}

		DBDag, err = historyDag.Config.ParseToDag()
		if err != nil {
			log.Warnf("[logic.GetDagByID] ParseToDag err, deail: %s", err.Error())
			return nil, ierrors.NewIError(ierrors.InternalError, "", err.Error())
		}
	}

	for _, task := range DBDag.Tasks { //nolint
		if task.ActionName == common.AnyshareFileSetPermOpt {
			delete(task.Params, "apppwd")
		}
	}

	for _, step := range DBDag.Steps {
		if step.Operator == common.AnyshareFileSetPermOpt {
			delete(step.Parameters, "apppwd")
		}
	}

	dag.ID = DBDag.ID
	dag.Title = DBDag.Name
	dag.Steps = DBDag.Steps
	dag.Status = fmt.Sprintf("%v", DBDag.Status)
	dag.Description = DBDag.Description
	dag.CreatedAt = DBDag.CreatedAt
	dag.UpdatedAt = DBDag.UpdatedAt
	dag.Shortcuts = &DBDag.Shortcuts
	dag.Cron = DBDag.Cron
	dag.Published = DBDag.Published
	dag.Type = DBDag.Type
	dag.TriggerConfig = DBDag.TriggerConfig
	dag.ExecMode = DBDag.ExecMode
	dag.Category = DBDag.Category
	dag.OutPuts = DBDag.OutPuts
	dag.UserID = DBDag.UserID
	dag.DeBugID = DBDag.DeBugID

	if dag.TriggerConfig != nil {
		dataSource := dag.TriggerConfig.DataSource
		if dataSource != nil && len(dataSource.Parameters.DocIDs) > 0 {
			dag.TriggerConfig.DataSource.Parameters.Docs = make([]map[string]interface{}, 0)
			for _, docID := range dataSource.Parameters.DocIDs {
				query := []string{"doc_lib_type", "name", "path"}
				if !utils.IsGNS(docID) {
					dag.TriggerConfig.DataSource.Parameters.Docs = append(dag.TriggerConfig.DataSource.Parameters.Docs, map[string]interface{}{
						"docid":        docID,
						"doc_lib_type": docID,
						"path":         docID,
						"name":         docID,
					})
					continue
				}
				docInfo, merr := m.efast.GetDocMetaData(context.Background(), docID, query)
				if merr != nil {
					log.Warnf("[logic.GetDagByID] GetDocMetaData err, detail: %s", merr.Error())
					continue
				}
				dag.TriggerConfig.DataSource.Parameters.Docs = append(dag.TriggerConfig.DataSource.Parameters.Docs, map[string]interface{}{
					"docid":        docID,
					"doc_lib_type": docInfo.DocLibType,
					"path":         docInfo.Path,
					"name":         docInfo.Name,
				})
			}
		}

		parameters := dag.TriggerConfig.Parameters
		if parameters != nil && parameters["docids"] != nil {
			var docIDs []interface{}
			if ids, ok := normalizeutil.AsSlice(parameters["docids"]); ok {
				docIDs = ids
			} else {
				log.Warnf("[logic.GetDagByID] docids parameter is not a valid array type")
			}
			// Initialize the docs array before appending
			if docIDs != nil {
				dag.TriggerConfig.Parameters["docs"] = make([]map[string]interface{}, 0)

				for _, docID := range docIDs {
					query := []string{"doc_lib_type", "name", "path"}
					if !utils.IsGNS(docID.(string)) {
						dag.TriggerConfig.Parameters["docs"] = append(dag.TriggerConfig.Parameters["docs"].([]map[string]interface{}), map[string]interface{}{
							"docid":        docID,
							"doc_lib_type": docID,
							"path":         docID,
							"name":         docID,
						})
						continue
					}
					docInfo, merr := m.efast.GetDocMetaData(context.Background(), docID.(string), query)
					if merr != nil {
						log.Warnf("[logic.GetDagByID] GetDocMetaData err, detail: %s", merr.Error())
						continue
					}
					dag.TriggerConfig.Parameters["docs"] = append(dag.TriggerConfig.Parameters["docs"].([]map[string]interface{}), map[string]interface{}{
						"docid":        docID,
						"doc_lib_type": docInfo.DocLibType,
						"path":         docInfo.Path,
						"name":         docInfo.Name,
					})
				}
			}
		}
	}

	if DBDag.UserID == userInfo.UserID {
		var accessorMap = make(map[string]string, 0)
		for _, accessor := range DBDag.Accessors {
			accessorMap[accessor.ID] = accessor.Type
		}

		userMap, cerr := m.usermgnt.GetNameByAccessorIDs(accessorMap)
		if cerr != nil {
			log.Warnf("[logic.GetDagByID] GetNameByAccessorIDs err, detail: %s", cerr.Error())
			return nil, cerr
		}

		for index, accessor := range DBDag.Accessors {
			name, ok := userMap[accessor.ID]
			if ok {
				DBDag.Accessors[index].Name = name
			}
		}
		dag.Accessors = &DBDag.Accessors
		return dag, nil
	}

	return dag, nil
}

// DeleteDagByID 普通用户删除自己的DAG
func (m *mgnt) DeleteDagByID(ctx context.Context, dagID, bizDomainID string, userInfo *drivenadapters.UserInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// 普通用户只能删除自己的DAG
	query := bson.M{
		"_id":    dagID,
		"userid": userInfo.UserID,
	}

	dag, err := m.mongo.GetDagByFields(ctx, query)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": dagID})
		}
		log.Warnf("[logic.DeleteDagByID] GetDagByID err, detail: %s", err.Error())
		return err
	}

	// 执行删除逻辑...
	return m.deleteDag(ctx, dag, userInfo, utils.IfNot(bizDomainID == "", dag.BizDomainID, bizDomainID))
}

// deleteDag 公共删除逻辑
func (m *mgnt) deleteDag(ctx context.Context, dag *entity.Dag, userInfo *drivenadapters.UserInfo, bizDomainID string) error {
	log := traceLog.WithContext(ctx)

	go func(dag *entity.Dag) {
		err := m.mongo.BatchDeleteDagWithTransaction(context.Background(), []string{dag.ID})
		if err != nil {
			log.Warnf("[logic.deleteDag] DeleteDagWithTransaction err, detail: %s", err.Error())
			return
		}

		if dag.Cron != "" {
			err = m.ecron.DeleteEcronJob(context.Background(), dag.Cron)
			if err != nil {
				log.Warnf("[logic.deleteDag] DeleteEcronJob err, detail: %s", err.Error())
				return
			}
		}

		if err := m.extData.Remove(ctx, &rds.ExtDataQueryOptions{
			DagID: dag.ID,
		}); err != nil {
			log.Warnf("[logic.CreateDag] Remove extData err, deail: %s", err.Error())
		}

		resourceID := fmt.Sprintf("%s:%s", dag.ID, utils.IfNot(dag.Type == "", common.DagTypeDefault, dag.Type))
		// 此处由于部分是数据迁移的记录，导致数据库中并未真正记录业务域ID，一些内部接口或事件进行删除操作时，会因为空而报错
		bizDomainID = utils.IfNot(bizDomainID == "", common.BizDomainDefaultID, bizDomainID)
		err = m.bizDomain.UnBindResourceInternal(context.Background(), drivenadapters.BizDomainResourceParams{
			BizDomainID:  bizDomainID,
			ResourceID:   resourceID,
			ResourceType: perm.DataFlowResourceType,
		})
		if err != nil {
			log.Warnf("[logic.deleteDag] UnBindResourceInternal err, BizDomainID: %s, ResourceID: %s, detail: %s", bizDomainID, resourceID, err.Error())
		}

		detail, extMsg := common.GetLogBody(common.DeleteTask, []interface{}{dag.Name}, []interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: dag.ID,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_WARN,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dag.ID,
			Operation: drivenadapters.DeleteOperation,
			ObjID:     dag.ID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_WARN_Str,
		}, write)
	}(dag)

	return nil
}

func (m *mgnt) ListDag(ctx context.Context, param QueryParams, userInfo *drivenadapters.UserInfo, config *ListDagsConfig) ([]*DagSimpleInfo, int64, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	// Get user details to check if admin
	userDetail, err := m.usermgnt.GetUserInfoByType(userInfo.UserID, userInfo.AccountType)
	if err != nil {
		log.Warnf("[logic.ListDag] GetUserInfoByType err, detail: %s", err.Error())
		return nil, 0, ierrors.NewIError(ierrors.ErrorDepencyService, "", err.Error())
	}

	isAdmin := utils.IsAdminRole(userDetail.Roles)

	if config != nil && config.Scope == "all" {
		accessors, merr := m.usermgnt.GetUserAccessorIDs(userInfo.UserID)
		if merr != nil {
			log.Warnf("[logic.ListDag] GetUserAccessorIDs err, detail: %s", merr.Error())
			return nil, 0, merr
		}
		param.TriggerExclude = []string{common.OpAnyShareSelectedFileTrigger, common.OpAnyShareSelectedFolderTrigger}
		param.Accessors = accessors
		param.UserID = userInfo.UserID
		param.Scope = config.Scope
	} else if config != nil && config.IsShared {
		accessors, merr := m.usermgnt.GetUserAccessorIDs(userInfo.UserID)
		if merr != nil {
			log.Warnf("[logic.ListDag] GetUserAccessorIDs err, detail: %s", merr.Error())
			return nil, 0, merr
		}
		param.TriggerExclude = []string{common.OpAnyShareSelectedFileTrigger, common.OpAnyShareSelectedFolderTrigger}
		param.Accessors = accessors
	} else if !isAdmin || param.Type != common.DagTypeDataFlow { // Non-admin users only see their own data
		param.UserID = userInfo.UserID
	}
	// Admin users will see all data when not in shared mode (listDagInput.UserID remains unset)

	var dagArr = make([]*DagSimpleInfo, 0)
	param.KeyWord = regexp.QuoteMeta(param.KeyWord)
	dags, total, err := ListDagWithFilters(ctx, param,
		WithBizDomainFilter(m.bizDomain, param.BizDomainID, "", param.Type, userInfo.TokenID),
		WithSharedDagFilter(param))
	if err != nil {
		return dagArr, total, err
	}

	var accessorIDs = make(map[string]string)

	for _, dag := range dags {
		accessorIDs[dag.UserID] = common.User.ToString()
	}

	accessors, _ := m.usermgnt.GetNameByAccessorIDs(accessorIDs)

	for _, dag := range dags {
		_dag := dag
		dagArr = append(dagArr, &DagSimpleInfo{
			ID:          _dag.ID,
			Title:       _dag.Name,
			Description: _dag.Description,
			CreatedAt:   _dag.CreatedAt,
			UpdatedAt:   _dag.UpdatedAt,
			Actions:     ListOperators(_dag.Steps),
			Status:      fmt.Sprintf("%v", _dag.Status),
			UserID:      _dag.UserID,
			Creator:     accessors[_dag.UserID],
			Trigger:     _dag.Trigger,
			Type:        _dag.Type,
			VersionID:   _dag.VersionID,
		})
	}

	return dagArr, total, nil
}

func (m *mgnt) ListDagByFields(ctx context.Context, filter bson.M, opt options.FindOptions) ([]*DagSimpleInfo, int64, error) {

	var total int64
	var dagArr = make([]*DagSimpleInfo, 0)

	total, err := m.mongo.ListDagCountByFields(ctx, filter)
	if err != nil {
		log.Warnf("[logic.ListDagByFields] ListDagCountByFields err, detail: %s", err.Error())
		return dagArr, total, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	dags, err := m.mongo.ListDagByFields(ctx, filter, opt)

	if err != nil {
		log.Warnf("[logic.ListDagByFields] ListDagByFields err, detail: %s", err.Error())
		return dagArr, total, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	var accessorIDs = make(map[string]string)

	for _, dag := range dags {
		accessorIDs[dag.UserID] = common.User.ToString()
	}

	accessors, _ := m.usermgnt.GetNameByAccessorIDs(accessorIDs)

	for _, dag := range dags {
		_dag := dag

		var triggerStep *entity.Step

		if len(dag.Steps) > 0 {
			triggerStep = &dag.Steps[0]
		}

		dagArr = append(dagArr, &DagSimpleInfo{
			ID:          _dag.ID,
			Title:       _dag.Name,
			CreatedAt:   _dag.CreatedAt,
			UpdatedAt:   _dag.UpdatedAt,
			TriggerStep: triggerStep,
			Actions:     ListOperators(_dag.Steps),
			Status:      fmt.Sprintf("%v", _dag.Status),
			UserID:      _dag.UserID,
			Creator:     accessors[_dag.UserID],
		})
	}

	return dagArr, total, nil
}

// RunCronInstance 运行定时任务
func (m *mgnt) RunCronInstance(ctx context.Context, id, webhook string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() {
		trace.TelemetrySpanEnd(span, err)
	}()
	log := traceLog.WithContext(ctx)
	go func() {
		merr := m.ecron.PostEcronJobEnd(ctx, webhook)
		if merr != nil {
			log.Warnf("[logic.RunCronInstance] PostEcronJobEnd err, deail: %s", merr.Error())
		}
	}()
	dag, err := m.mongo.GetDag(ctx, id)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": id})
		}
		log.Warnf("[logic.RunCronInstance] GetDag err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	if dag.Status == entity.DagStatusStopped {
		return nil
	}
	dag.SetPushMessage(m.executeMethods.Publish)

	ins, err := m.mongo.ListDagInstance(ctx, &mod.ListDagInstanceInput{
		DagIDs:      []string{id},
		Status:      []entity.DagInstanceStatus{entity.DagInstanceStatusInit, entity.DagInstanceStatusScheduled},
		Limit:       10,
		SelectField: []string{"_id"},
	})

	if err != nil {
		log.Warnf("[logic.RunCronInstance] ListDagInstance err, deail: %s", err.Error())
		return err
	}

	if len(ins) > 0 {
		return nil
	}

	userDetail, tokenInfo, err := m.getUserDetail(dag.UserID, &dag.AppInfo)
	if err != nil {
		log.Warnf("[logic.RunCronInstance] getUserDetail err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.UnAuthorization, "", map[string]interface{}{"info": err.Error()})
	}

	userDetail.VisitorType = common.InternalServiceUserType
	if dag.AppInfo.Enable {
		userDetail.VisitorType = common.APP.ToString()
	}

	var datasourceid = ""
	if dag.Steps[0].DataSource != nil {
		datasourceid = dag.Steps[0].DataSource.ID
	}

	runVar := map[string]string{
		"userid":        userDetail.UserID,
		"operator_id":   userDetail.UserID,
		"operator_name": userDetail.UserName,
		"operator_type": userDetail.AccountType,
		"datasourceid":  datasourceid,
	}

	triggerType := m.getTriggerType(dag.Steps[0].Operator)
	runVar["source_type"] = m.getDataSourceType(dag.Steps[0].DataSource)
	dataSource := dag.Steps[0].DataSource
	if dag.Type == common.DagTypeDataFlow {
		runVar["source_type"] = m.getDataSourceType(dag.TriggerConfig.DataSource)
		runVar["datasourceid"] = dag.Steps[0].ID
		dataSource = dag.TriggerConfig.DataSource
		triggerType = m.getTriggerType(dag.TriggerConfig.Operator)
	}

	if triggerType != entity.TriggerCron {
		err := ierrors.NewIError(ierrors.Forbidden, ierrors.ErrorIncorretTrigger, map[string]interface{}{
			"trigger": fmt.Sprintf("%s trigger type is not allowed to run cron", triggerType),
		})
		return err
	}

	if dag.Steps[0].Operator == common.MDLDataViewTrigger {
		err = m.triggerFromMDLDataView(ctx, dag, triggerType, runVar, userDetail.UserID, userDetail.AccountType, tokenInfo.LoginIP)
		if err != nil {
			log.Warnf("[logic.RunCronInstance] triggerFromMDLDataView err %s", err.Error())
			return err
		}

		return nil
	}

	if runVar["source_type"] == "" {
		err = m.runDagInstance(ctx, dag, triggerType, runVar, userDetail)
		if err != nil {
			return err
		}
		return nil
	}
	_, err = m.triggerfromDataSource(ctx, dataSource, tokenInfo.UserID, tokenInfo.Token, tokenInfo.LoginIP, triggerType, runVar, dag)
	if err != nil {
		return err
	}

	go func() {
		detail, extMsg := common.GetLogBody(common.TriggerTaskCron, []interface{}{dag.Name},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: userDetail,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: id,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		randomID, _ := utils.GetUniqueID()
		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  userDetail,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  fmt.Sprintf("%v", randomID),
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     id,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}()
	return nil
}

// RunInstance 手动运行
func (m *mgnt) RunInstance(ctx context.Context, params *RunDagParams, userInfo *drivenadapters.UserInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dag, err := m.mongo.GetDagWithOptionalVersion(ctx, params.ID, params.VersionID)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": params.ID})
		}
		log.Warnf("[logic.RunInstance] GetDag err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	dag.SetPushMessage(m.executeMethods.Publish)

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow: {perm.ManualExecOperation},
			common.DagTypeDefault:  {perm.OldAdminOperation, perm.OldShareOperation},
		},
	}

	if userInfo.AccountType == common.APP.ToString() {
		opMap.OpMap[common.DagTypeDefault] = []string{perm.OldAppTokenOperation}
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dag.ID, userInfo, opMap)
	if err != nil {
		return err
	}

	_, tokenInfo, err := m.getUserDetail(dag.UserID, &dag.AppInfo)
	if err != nil {
		return ierrors.NewIError(ierrors.UnAuthorization, "", map[string]interface{}{"info": err.Error()})
	}

	var datasourceid = ""
	if dag.Steps[0].DataSource != nil {
		datasourceid = dag.Steps[0].DataSource.ID
	}

	runVar := map[string]string{
		"userid":        userInfo.UserID,
		"operator_id":   userInfo.UserID,
		"operator_name": userInfo.UserName,
		"operator_type": common.User.ToString(),
		"datasourceid":  datasourceid,
	}

	triggerType := m.getTriggerType(dag.Steps[0].Operator)

	runVar["source_type"] = m.getDataSourceType(dag.Steps[0].DataSource)
	dataSource := dag.Steps[0].DataSource

	if dag.Type == common.DagTypeDataFlow {
		runVar["source_type"] = m.getDataSourceType(dag.TriggerConfig.DataSource)
		runVar["datasourceid"] = dag.Steps[0].ID
		dataSource = dag.TriggerConfig.DataSource
		triggerType = m.getTriggerType(dag.TriggerConfig.Operator)
	}

	if triggerType != entity.TriggerManually {
		err := ierrors.NewIError(ierrors.Forbidden, ierrors.ErrorIncorretTrigger, map[string]interface{}{
			"trigger": fmt.Sprintf("%s trigger type is not allowed to run manually", triggerType),
		})
		return err
	}

	if params.ParentDagInstanceID != "" {
		parentDagIns, err1 := m.mongo.GetDagInstance(ctx, params.ParentDagInstanceID)
		if err1 != nil {
			log.Warnf("[logic.RunInstance] GetDagInstance err, deail: %s", err1.Error())
			return ierrors.NewIError(ierrors.InternalError, "", nil)
		}

		callChain := append(parentDagIns.CallChain, dag.ID)
		if len(callChain) > MaxCallDepth {
			return libErrors.NewPublicRestError(ctx, libErrors.PErrorBadRequest,
				libErrors.PErrorBadRequest, map[string]interface{}{
					"message":        fmt.Sprintf("Maximum call chain depth exceeded (limit: %d).", MaxCallDepth),
					"max_call_depth": MaxCallDepth,
				},
			)
		}
		runVar["call_chain"] = strings.Join(callChain, ",")
	}

	if params.CallBack != "" {
		runVar["call_back"] = params.CallBack
		runVar["batch_run_id"], err = utils.GetUniqueIDStr()
		if err != nil {
			return err
		}
	}

	if dag.Steps[0].Operator == common.MDLDataViewTrigger {
		err = m.triggerFromMDLDataView(ctx, dag, triggerType, runVar, userInfo.UserID, userInfo.AccountType, tokenInfo.LoginIP)
		if err != nil {
			log.Warnf("[logic.RunInstance] triggerFromMDLDataView err %s", err.Error())
			return err
		}
		return nil
	}

	if runVar["source_type"] == "" {
		err = m.runDagInstance(ctx, dag, triggerType, runVar, userInfo)
		if err != nil {
			return err
		}
		return nil
	}

	_, err = m.triggerfromDataSource(ctx, dataSource, tokenInfo.UserID, tokenInfo.Token, tokenInfo.LoginIP, triggerType, runVar, dag)
	if err != nil {
		return err
	}

	go func() {
		detail, extMsg := common.GetLogBody(common.TriggerTaskManually, []interface{}{dag.Name},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: id,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		randomID, _ := utils.GetUniqueID()
		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  fmt.Sprintf("%v", randomID),
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     params.ID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}()
	return nil
}

func (m *mgnt) RunFormInstance(ctx context.Context, params *RunDagParams, userInfo *drivenadapters.UserInfo) (string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dag, err := m.mongo.GetDagWithOptionalVersion(ctx, params.ID, params.VersionID)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return "", ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": params.ID})
		}
		log.Warnf("[logic.RunFormInstance] GetDag err, deail: %s", err.Error())
		return "", ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	dag.SetPushMessage(m.executeMethods.Publish)

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow: {perm.ManualExecOperation},
		},
	}
	if dag.Published {
		opMap.OpMap[common.DagTypeDefault] = []string{perm.OldPublishOperatiuon}
	} else if userInfo.AccountType == common.APP.ToString() {
		opMap.OpMap[common.DagTypeDefault] = []string{perm.OldAppTokenOperation}
	} else {
		opMap.OpMap[common.DagTypeDefault] = []string{perm.OldShareOperation}
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dag.ID, userInfo, opMap)
	if err != nil {
		return "", err
	}

	if !dag.Published && userInfo == nil {
		return "", ierrors.NewIError(ierrors.Forbidden, "", map[string]string{"dagId": params.ID})
	}

	if userInfo != nil {
		userDetail, err := m.usermgnt.GetUserInfoByType(userInfo.UserID, userInfo.AccountType)
		if err != nil {
			log.Warnf("[logic.RunFormInstance] GetUserInfoByType err, deail: %s", err.Error())
			return "", err
		}

		userInfo.UserName = userDetail.UserName
	}

	// Check if the dag is published
	if dag.Published {
		// Use existing userInfo if available, otherwise create a default anonymous user
		if userInfo == nil {
			userInfo = &drivenadapters.UserInfo{
				UserID:      "anonymous",      // or any default ID
				UserName:    "Anonymous User", // or any default name
				Type:        "anonymous",      // or any default type
				VisitorType: common.AnonymousUserType,
			}
		}
	}

	runVar := map[string]string{
		"userid":        userInfo.UserID,
		"operator_id":   userInfo.UserID,
		"operator_name": userInfo.UserName,
		"operator_type": userInfo.AccountType,
	}

	if params.ParentDagInstanceID != "" {
		parentDagIns, err1 := m.mongo.GetDagInstance(ctx, params.ParentDagInstanceID)
		if err1 != nil {
			log.Warnf("[logic.RunFormInstance] GetDagInstance err, deail: %s", err1.Error())
			return "", ierrors.NewIError(ierrors.InternalError, "", nil)
		}

		callChain := append(parentDagIns.CallChain, dag.ID)
		if len(callChain) > MaxCallDepth {
			return "", libErrors.NewPublicRestError(ctx, libErrors.PErrorBadRequest,
				libErrors.PErrorBadRequest, map[string]interface{}{
					"message":        fmt.Sprintf("Maximum call chain depth exceeded (limit: %d).", MaxCallDepth),
					"max_call_depth": MaxCallDepth,
				},
			)
		}
		runVar["call_chain"] = strings.Join(callChain, ",")
	}

	if params.CallBack != "" {
		runVar["call_back"] = params.CallBack
	}

	for key, val := range params.FormData {
		switch v := val.(type) {
		case string, int, float64:
			runVar[key] = fmt.Sprintf("%v", v)
		case map[string]interface{}, []string, []interface{}, []map[string]interface{}:
			bytes, _ := json.Marshal(v)
			runVar[key] = string(bytes)
		default:
			runVar[key] = fmt.Sprintf("%v", v)
		}
	}

	triggerType := m.getTriggerType(dag.Steps[0].Operator)

	if triggerType != entity.TriggerForm {
		err := ierrors.NewIError(ierrors.Forbidden, ierrors.ErrorIncorretTrigger, map[string]interface{}{
			"trigger": fmt.Sprintf("%s trigger type is not allowed to run form", triggerType),
		})
		return "", err
	}

	dagIns, dagErr := dag.Run(ctx, triggerType, runVar, entity.WithKeyWords(userInfo.UserName))
	if dagErr != nil {
		return "", ierrors.NewIError(ierrors.Forbidden, ierrors.DagStatusNotNormal, map[string]interface{}{"id:": dag.ID, "status": dag.Status})
	}

	dagInsID, err := m.mongo.CreateDagIns(ctx, dagIns)
	if err != nil {
		log.Warnf("[logic.RunFromInstance] CreateDagIns err, deail: %s", err.Error())
		return "", ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	go func() {
		detail, extMsg := common.GetLogBody(common.TriggerTaskManually, []interface{}{dag.Name},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: id,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagIns.ID,
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     params.ID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}()
	return dagInsID, nil
}

func (m *mgnt) getSource(msg *common.DocMsg, triggerType string, log *traceLog.Logger) []string {
	if triggerType == common.AnyshareFileCopyTrigger ||
		triggerType == common.AnyshareFolderCopyTrigger {
		docs := utils.GetParentDocIDs(msg.NewID)
		docInfo, err := m.efast.GetDocMetaData(context.Background(), msg.NewID, []string{"doc_lib_type"})
		if err != nil {
			(*log).Warnf("[getSource] err: %s", err.Error())
			return docs
		}
		docs = append(docs, docInfo.DocLibType)
		return docs
	}
	if triggerType == common.AnyshareFileMoveTrigger ||
		triggerType == common.AnyshareFolderMoveTrigger {
		docs := utils.GetParentDocIDs(msg.DocID)
		newDocs := utils.GetParentDocIDs(msg.NewID)
		docs = append(docs, newDocs...)
		docInfo, err := m.efast.GetDocMetaData(context.Background(), msg.NewID, []string{"doc_lib_type"})
		if err != nil {
			(*log).Warnf("[getSource] err: %s", err.Error())
			return docs
		}
		return append(docs, docInfo.DocLibType)
	}
	docs := utils.GetParentDocIDs(msg.DocID)
	docInfo, err := m.efast.GetDocMetaData(context.Background(), docs[0], []string{"doc_lib_type"})
	if err != nil {
		(*log).Warnf("[getSource] err: %s", err.Error())
		return docs
	}
	docs = append(docs, docInfo.DocLibType)
	return docs
}

// HandleDocEvent 处理事件
func (m *mgnt) HandleDocEvent(ctx context.Context, msg *DocMsg, topic string) error { //nolint
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var triggerType = common.GetTriggerTypeFromTopic(topic)
	if len(triggerType) == 0 {
		return nil
	}
	const maxSource = 100
	sources := m.getSource(msg, triggerType[0], &log)
	// 超过100层级不再触发
	if len(sources) > maxSource {
		sources = sources[len(sources)-maxSource:]
	}

	if msg.Operator.ID == "" && topic != common.TopicFileDelete {
		metadata, merr := m.efast.GetDocMsg(ctx, msg.DocID)
		if merr != nil {
			return merr
		}
		msg.Operator.ID = metadata.CreatorID
		msg.Operator.Name = metadata.Creator
	}

	dags, err := m.mongo.ListDag(ctx, &mod.ListDagInput{
		Trigger: triggerType,
		Sources: sources,
		Type:    "all",
	})

	// 用于审计日志操作人信息
	userInfo, err := m.usermgnt.GetUserInfo(msg.Operator.ID)
	userInfo.VisitorType = common.InternalServiceUserType
	userInfo.LoginIP = os.Getenv("POD_IP")
	if err != nil {
		log.Warnf("[logics.mgnt.HandleDocEvent] GetUserInfo failed: %v", err.Error())
		userInfo.UserID = msg.Operator.ID
		userInfo.UserName = msg.Operator.Name
	}

	for _, dag := range dags {
		var (
			macthed = false
			inherit = false
			docid   = dag.Tasks[0].Params["docid"]
			docids  = dag.Tasks[0].Params["docids"]
		)

		if dag.Type == common.DagTypeDataFlow {
			docids = dag.TriggerConfig.Parameters["docids"]
			inherit = true
		}

		if dag.Status == entity.DagStatusStopped {
			continue
		}

		if docids == nil {
			if len(sources) > 1 && docid != sources[len(sources)-2] && dag.Tasks[0].Params["inherit"] != true {
				continue
			}
		} else {
			ids, ok := normalizeutil.AsSlice(docids)
			if !ok {
				continue
			}

			for _, cid := range ids {
				if (len(sources) > 1 && cid == sources[len(sources)-2]) || dag.Tasks[0].Params["inherit"] == true || inherit {
					macthed = true
					break
				}
			}

			if !macthed {
				continue
			}
		}

		runVar := map[string]string{
			"userid":        dag.UserID,
			"id":            msg.DocID,
			"docid":         msg.DocID,
			"new_id":        msg.NewID,
			"name":          msg.DocName,
			"path":          msg.Path,
			"new_path":      msg.NewPath,
			"size":          fmt.Sprintf("%v", msg.Size),
			"operator_id":   msg.Operator.ID,
			"operator_name": msg.Operator.Name,
			"operator_type": msg.Operator.Type,
			"source_type":   "doc",
		}

		if dag.Type == common.DagTypeDataFlow && msg.NewID != "" {
			runVar["docid"] = msg.NewID
			runVar["id"] = msg.NewID
		}

		dag.SetPushMessage(m.executeMethods.Publish)
		dagIns, err := dag.Run(ctx, entity.TriggerEvent, runVar, entity.WithKeyWords(msg.DocName)) //nolint
		if err != nil {
			return err
		}

		_, err = m.mongo.CreateDagIns(ctx, dagIns)
		if err != nil {
			return err
		}

		detail, extMsg := common.GetDocTriggerLogBody(triggerType[0], dag.Name, msg)
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: &userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: dagIns.ID,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  &userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagIns.ID,
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     dag.ID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}

	if err != nil {
		return err
	}

	return nil
}

// HandleUserInfoEvent 处理用户及部门信息变更事件
func (m *mgnt) HandleUserInfoEvent(ctx context.Context, msg *common.UserInfoMsg, topic string) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var triggerType = common.GetTriggerTypeFromTopic(topic)
	if len(triggerType) == 0 {
		return nil
	}
	runVar := map[string]string{
		"id":            msg.ID,
		"new_name":      msg.NewName,
		"name":          msg.Name,
		"type":          msg.Type,
		"old_dept_path": msg.OldDeptPath,
		"new_dept_path": msg.NewDeptPath,
		"source_type":   "dept",
	}

	if triggerType[0] == common.AnyshareOrgNameModifyTrigger && msg.Type == common.User.ToString() {
		triggerType[0] = common.AnyshareUserChangeTrigger // 区分用户改名和组织改名
		curUserInfo, gerr := m.usermgnt.GetUserInfo(msg.ID)
		if err != nil {
			log.Warnln("[logics.HandleUserInfoEvent] GetUserInfo failed: %s", gerr.Error())
			return gerr
		}
		runVar["name"] = curUserInfo.UserName
		runVar["role"] = fmt.Sprintln(curUserInfo.Roles)
		runVar["csflevel"] = fmt.Sprintln(curUserInfo.CsfLevel)
		runVar["email"] = curUserInfo.Email
		runVar["contact"] = curUserInfo.Telephone
		runVar["status"] = "enabled"
		if !curUserInfo.Enabled {
			runVar["status"] = "disabled"
		}
		runVar["source_type"] = common.User.ToString()
	}
	dags, err := m.mongo.ListDag(ctx, &mod.ListDagInput{
		Trigger: triggerType,
		Type:    "all",
	})

	if err != nil {
		log.Warnf("[logics.HandleUserInfoEvent] ListDag failed: %s", err.Error())
		return err
	}

	keywords := []string{}

	if msg.Name != "" {
		keywords = append(keywords, msg.Name)
	}

	if msg.NewName != "" {
		keywords = append(keywords, msg.NewName)
	}

	if msg.Email != "" {
		keywords = append(keywords, msg.Email)
	}

	for _, dag := range dags {
		if dag.Status == entity.DagStatusStopped {
			continue
		}

		if dag.TriggerConfig != nil &&
			(dag.TriggerConfig.Operator == common.AnyshareUserChangeTrigger ||
				dag.TriggerConfig.Operator == common.AnyshareUserCreateTrigger ||
				dag.TriggerConfig.Operator == common.AnyshareUserDeleteTrigger) {
			runVar["source_type"] = common.User.ToString()
		}

		dag.SetPushMessage(m.executeMethods.Publish)
		runVar["userid"] = dag.UserID
		runVar["operator_id"] = dag.UserID

		if msg.DeptPaths != nil {
			paths, _ := json.Marshal(msg.DeptPaths)
			runVar["dept_paths"] = string(paths)
		}

		dagIns, err := dag.Run(ctx, entity.TriggerEvent, runVar, entity.WithKeyWords(keywords...))
		if err != nil {
			log.Warnf("[logics.HandleUserInfoEvent] Run dag failed: %s", err.Error())
			return err
		}

		_, err = m.mongo.CreateDagIns(ctx, dagIns)
		if err != nil {
			log.Warnf("[logics.HandleUserInfoEvent] CreateDagIns failed: %s", err.Error())
			return err
		}

		detail, extMsg := common.GetUserInfoTriggerLogBody(triggerType[0], dag.Name, msg)
		userInfo, err := m.usermgnt.GetUserInfo(dag.UserID)
		userInfo.VisitorType = common.InternalServiceUserType
		userInfo.LoginIP = os.Getenv("POD_IP")
		if err != nil {
			log.Warnf("[logics.mgnt.HandleUserInfoEvent] GetUserInfo failed: %v", err.Error())
		}
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: &userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: dagIns.ID,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  &userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagIns.ID,
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     dag.ID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}

	return nil
}

// HandleTagInfoChangeEvent 官方标签信息变更事件
func (m *mgnt) HandleTagInfoChangeEvent(ctx context.Context, tag *common.TagInfo, msg []byte, topic string) error { //nolint
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	// log := traceLog.WithContext(ctx)

	var triggerType = common.GetTriggerTypeFromTopic(topic)
	if len(triggerType) == 0 {
		return nil
	}
	dags, err := m.mongo.ListDag(ctx, &mod.ListDagInput{
		Trigger: triggerType,
		Type:    "all",
	})

	for _, dag := range dags {
		if dag.Status == entity.DagStatusStopped {
			continue
		}

		runVar := map[string]string{
			"userid":      dag.UserID,
			"tags":        string(msg),
			"operator_id": dag.UserID,
			"source_type": "tagtree",
			"id":          tag.ID,
		}

		dag.SetPushMessage(m.executeMethods.Publish)
		dagIns, err := dag.Run(ctx, entity.TriggerEvent, runVar, entity.WithKeyWords(tag.Name)) //nolint
		if err != nil {
			return err
		}

		_, err = m.mongo.CreateDagIns(ctx, dagIns)
		if err != nil {
			return err
		}

		detail, extMsg := common.GetTagTreeTriggerLogBody(triggerType[0], dag.Name, string(msg))
		userInfo, err := m.usermgnt.GetUserInfo(dag.UserID)
		userInfo.VisitorType = common.InternalServiceUserType
		userInfo.LoginIP = os.Getenv("POD_IP")
		if err != nil {
			log.Warnf("[logics.mgnt.HandleUserInfoEvent] GetUserInfo failed: %v", err.Error())
		}
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: &userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: dagIns.ID,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  &userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagIns.ID,
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     dag.ID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}

	if err != nil {
		return err
	}

	return nil
}

// HandleTagTreeCreateEvent 官方标签树创建事件
func (m *mgnt) HandleTagTreeCreateEvent(ctx context.Context, msg *common.TagInfo, topic string) error { //nolint
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	// log := traceLog.WithContext(ctx)

	var triggerType = common.GetTriggerTypeFromTopic(topic)
	if len(triggerType) == 0 {
		return nil
	}
	dags, err := m.mongo.ListDag(ctx, &mod.ListDagInput{
		Trigger: triggerType,
		Type:    "all",
	})

	for _, dag := range dags {
		if dag.Status == entity.DagStatusStopped {
			continue
		}

		runVar := map[string]string{
			"userid":      dag.UserID,
			"id":          msg.ID,
			"path":        msg.Path,
			"version":     fmt.Sprintln(msg.Version),
			"name":        msg.Name,
			"parent_id":   msg.ParentID,
			"source_type": "tagtree",
			"operator_id": dag.UserID,
		}

		dag.SetPushMessage(m.executeMethods.Publish)
		dagIns, err := dag.Run(ctx, entity.TriggerEvent, runVar, entity.WithKeyWords(msg.Name)) //nolint
		if err != nil {
			return err
		}

		_, err = m.mongo.CreateDagIns(ctx, dagIns)
		if err != nil {
			return err
		}

		detail, extMsg := common.GetTagTreeTriggerLogBody(triggerType[0], dag.Name, msg)
		userInfo, err := m.usermgnt.GetUserInfo(dag.UserID)
		userInfo.VisitorType = common.InternalServiceUserType
		userInfo.LoginIP = os.Getenv("POD_IP")
		if err != nil {
			log.Warnf("[logics.mgnt.HandleUserInfoEvent] GetUserInfo failed: %v", err.Error())
		}
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: &userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: dagIns.ID,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  &userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagIns.ID,
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     dag.ID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}

	if err != nil {
		return err
	}

	return nil
}

// HandleKCUserInfoEvent 处理KC用户信息变更事件
func (m *mgnt) HandleKCUserInfoEvent(ctx context.Context, msg *common.UserInfoMsg) error { //nolint
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	// log := traceLog.WithContext(ctx)

	var triggerType = []string{common.AnyshareUserChangeTrigger}
	if msg.IsDelete == 1 && msg.Status != 1 {
		triggerType = []string{common.AnyshareUserDeleteTrigger}
	}
	dags, err := m.mongo.ListDag(ctx, &mod.ListDagInput{
		Trigger: triggerType,
		Type:    "all",
	})

	for _, dag := range dags {
		if dag.Status == entity.DagStatusStopped {
			continue
		}

		runVar := map[string]string{
			"userid":            dag.UserID,
			"id":                msg.ID,
			"name":              msg.Name,
			"email":             msg.Email,
			"tags":              msg.Tags,
			"is_expert":         fmt.Sprint(msg.IsExpert),
			"verification_info": msg.VerificationInfo,
			"university":        msg.University,
			"contact":           msg.Contact,
			"position":          msg.Position,
			"work_at":           msg.WorkAt,
			"is_delete":         fmt.Sprint(msg.IsDelete),
			"professional":      msg.Professional,
			"status":            "enabled",
			"source_type":       common.User.ToString(),
		}

		if msg.Status > 0 {
			runVar["status"] = "disabled"
		}

		dag.SetPushMessage(m.executeMethods.Publish)
		dagIns, err := dag.Run(ctx, entity.TriggerEvent, runVar, entity.WithKeyWords(msg.Name, msg.Email)) //nolint
		if err != nil {
			return err
		}

		_, err = m.mongo.CreateDagIns(ctx, dagIns)
		if err != nil {
			return err
		}

		detail, extMsg := common.GetUserInfoTriggerLogBody(triggerType[0], dag.Name, msg)
		userInfo, err := m.usermgnt.GetUserInfo(dag.UserID)
		userInfo.VisitorType = common.InternalServiceUserType
		userInfo.LoginIP = os.Getenv("POD_IP")
		if err != nil {
			log.Warnf("[logics.mgnt.HandleUserInfoEvent] GetUserInfo failed: %v", err.Error())
		}
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: &userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: dagIns.ID,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  &userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagIns.ID,
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     dag.ID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}

	if err != nil {
		return err
	}

	return nil
}

func (m *mgnt) CancelRunningInstance(ctx context.Context, id string, dagInsReq *DagInsStatusReq, userInfo *drivenadapters.UserInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	query := map[string]interface{}{"_id": id}
	dagIns, err := m.mongo.GetDagInstanceByFields(ctx, query)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.DagInsNotFound, "", map[string]string{"dagInsID": id})
		}
		log.Warnf("[logic.CancelRunningInstance] GetDagInstanceByFields err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	dag, err := m.mongo.GetDagByFields(ctx, map[string]interface{}{"_id": dagIns.DagID})
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.DagInsNotFound, "", map[string]string{"dagInsID": id})
		}
		log.Warnf("[logic.CancelRunningInstance] GetDagByFields err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow:      {perm.ModeifyOperation},
			common.DagTypeComboOperator: {perm.ModeifyOperation},
			common.DagTypeDefault:       {perm.OldAdminOperation, perm.OldShareOperation},
		},
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dag.ID, userInfo, opMap)
	if err != nil {
		return err
	}

	// dag ins success、failed、cancled
	if dagIns.Status == entity.DagInstanceStatusSuccess ||
		dagIns.Status == entity.DagInstanceStatusFailed ||
		dagIns.Status == entity.DagInstanceStatusCancled ||
		dagIns.Status == "" {
		return ierrors.NewIError(ierrors.Forbidden, ierrors.DagInsNotRunning, map[string]string{"status": fmt.Sprintf("status: %s, dag Ins is success failed cancled or ' '", dagIns.Status)})
	}

	dagIns.Status = entity.DagInstanceStatus(dagInsReq.Status)
	dagIns.EndedAt = time.Now().Unix()
	taskIns, err := m.mongo.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
		DagInsID: dagIns.ID,
		Status:   []entity.TaskInstanceStatus{entity.TaskInstanceStatusInit},
	})
	if err != nil {
		log.Warnf("[logic.CancelRunningInstance] ListTaskInstance err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	for _, taskIn := range taskIns {
		taskIn.Status = entity.TaskInstanceStatusCanceled
		err = m.mongo.UpdateTaskIns(ctx, taskIn)

		if err != nil {
			log.Warnf("[logic.CancelRunningInstance] UpdateTaskIns err, deail: %s", err.Error())
			continue
		}
	}

	err = m.mongo.UpdateDagIns(ctx, dagIns)
	if err != nil {
		log.Warnf("[logic.CancelRunningInstance] UpdateDagIns err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	go func() {
		dag, err := m.mongo.GetDagWithOptionalVersion(context.Background(), dagIns.DagID, dagIns.VersionID)
		if err != nil {
			log.Errorf("[logic.CancelRunningInstance] get dag[%s] failed: %s", dagIns.DagID, err)
			return
		}

		detail, _ := common.GetLogBody(common.CancelRunningInstance, []interface{}{dag.Name}, []interface{}{})

		object := map[string]interface{}{
			"type":          dag.Trigger,
			"id":            dagIns.ID,
			"dagId":         dagIns.DagID,
			"name":          dag.Name,
			"priority":      dagIns.Priority,
			"status":        entity.DagInstanceStatusCancled,
			"biz_domain_id": utils.IfNot(dag.BizDomainID == "", common.BizDomainDefaultID, dag.BizDomainID),
		}

		if len(dag.Type) != 0 {
			object["dagType"] = dag.Type
		} else {
			object["dagType"] = common.DagTypeDefault
		}

		if dagIns.EndedAt < dagIns.CreatedAt {
			endedAt := time.Now().Unix()
			object["duration"] = endedAt - dagIns.CreatedAt
		} else {
			object["duration"] = dagIns.EndedAt - dagIns.CreatedAt
		}

		m.logger.LogO11y(&drivenadapters.BuildARLogParams{
			Operation:   common.ArLogEndDagIns,
			Description: detail,
			UserInfo:    userInfo,
			Object:      object,
		}, &drivenadapters.O11yLogWriter{Logger: traceLog.NewFlowO11yLogger()})
	}()

	return nil
}

func (m *mgnt) GetSuggestDagName(ctx context.Context, name string, userInfo *drivenadapters.UserInfo) (string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var suggestName string
	var reg = regexp.MustCompile("^[^/:\\*?\"<>\\|]+$")
	res := reg.FindAllString(name, -1)
	if len(res) == 0 {
		return suggestName, ierrors.NewIError(ierrors.InvalidParameter, "", map[string]string{"name": fmt.Sprintf("%s: contain invalid chart / : * ? \" < > |", name)})
	}
	query := &mod.ListDagInput{
		KeyWord: fmt.Sprintf("^%v(\\(\\d+\\))?$", name),
		Type:    "all",
	}
	dags, err := m.mongo.ListDag(ctx, query)
	if err != nil {
		log.Warnf("[logic.GetSuggestDagName] ListDag err, deail: %s", err.Error())
		return suggestName, ierrors.NewIError(ierrors.InternalError, "", map[string]string{"query": fmt.Sprintf("keyword: %s", query.KeyWord)})
	}

	if len(dags) == 0 {
		return name, nil
	}

	var nameList = make([]string, len(dags))
	for _, dag := range dags {
		_dag := dag
		dagName := strings.TrimPrefix(_dag.Name, name)
		if dagName == "" {
			nameList[0] = _dag.Name
		} else {
			dagName = strings.TrimPrefix(dagName, "(")
			dagName = strings.TrimSuffix(dagName, ")")
			index, err := strconv.Atoi(dagName)
			if err != nil {
				return suggestName, ierrors.NewIError(ierrors.InternalError, "", err.Error())
			}
			if index >= len(nameList) {
				continue
			}
			nameList[index] = _dag.Name
		}
	}

	for k, v := range nameList {
		if v == "" && k == 0 {
			suggestName = name
			break
		}
		if v == "" {
			suggestName = fmt.Sprintf("%s(%v)", name, k)
			break
		}
	}
	if suggestName == "" {
		suggestName = fmt.Sprintf("%s(%v)", name, len(nameList))
	}
	return suggestName, nil
}

func (m *mgnt) ListDagInstance(ctx context.Context, dagID string, param map[string]interface{}, userInfo *drivenadapters.UserInfo) (*DagInstanceRunList, int64, error) { //nolint
	var success, failed, total int64
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var dagInsRunList = &DagInstanceRunList{}

	query := map[string]interface{}{"_id": dagID}
	isAdmin := utils.IsAdminRole(userInfo.Roles)
	if !isAdmin {
		query["userid"] = userInfo.UserID
	}

	// Step1: check dag whether exist
	err = m.isDagExist(ctx, query)
	if err != nil {
		return dagInsRunList, total, err
	}

	query = map[string]interface{}{"dagId": dagID}
	if param["start_time"] != nil && param["end_time"] != nil {
		query["createdAt"] = bson.M{"$gte": param["start_time"].(int64), "$lte": param["end_time"].(int64)}
	}

	// Step2: compute dag ins count
	total, err = m.mongo.GetDagInstanceCount(ctx, query)
	if err != nil {
		log.Warnf("[logic.ListDagInstance] GetDagInstanceCount total err, detail: %s", err.Error())
		return dagInsRunList, total, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	query["status"] = entity.DagInstanceStatusSuccess
	success, err = m.mongo.GetDagInstanceCount(ctx, query)
	if err != nil {
		log.Warnf("[logic.ListDagInstance] GetDagInstanceCount success err, detail: %s", err.Error())
		return dagInsRunList, total, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	query["status"] = entity.DagInstanceStatusFailed
	failed, err = m.mongo.GetDagInstanceCount(ctx, query)
	if err != nil {
		log.Warnf("[logic.ListDagInstance] GetDagInstanceCount failed err, detail: %s", err.Error())
		return dagInsRunList, total, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// Ste3: list all dag ins
	input := &mod.ListDagInstanceInput{
		DagIDs: []string{dagID},
		Limit:  param["limit"].(int64),
		Offset: param["page"].(int64),
		Order:  -1,
	}

	if param["type"] != nil {
		_status := param["type"].([]string)
		for _, value := range _status {
			_value := value
			switch _value {
			case common.SuccessStatus:
				input.Status = append(input.Status, entity.DagInstanceStatusSuccess)
			case common.FailedStatus:
				input.Status = append(input.Status, entity.DagInstanceStatusFailed)
			case common.CanceledStatus:
				input.Status = append(input.Status, entity.DagInstanceStatusCancled)
			case common.RunningStatus:
				input.Status = append(input.Status, entity.DagInstanceStatusRunning)
			default:
				input.Status = append(input.Status, entity.DagInstanceStatusInit,
					entity.DagInstanceStatusScheduled,
					entity.DagInstanceStatusBlocked)
			}
		}
	}
	if param["order"].(string) == common.ASC {
		input.Order = 1
	}
	// sort by custome filed
	sortBy := param["sortby"].(string)
	if sortBy == common.Started_At {
		input.SortBy = common.CreatedAt
	} else if sortBy == common.Ended_At {
		input.SortBy = common.EndedAt
	}
	if param["start_time"] != nil && param["end_time"] != nil {
		input.TimeRange = &mod.TimeRangeSearch{
			Begin: param["start_time"].(int64),
			End:   param["end_time"].(int64),
			Field: "createdAt",
		}
	}
	if param["name"] != nil {
		filter := bson.M{"$regex": param["name"], "$options": "i"}
		input.MatchQuery = &mod.MatchQuery{
			Field: "keywords",
			Value: filter,
		}
	}

	input.SelectField = []string{"_id", "createdAt", "endedAt", "status", "source", "reason"}

	dagInsList, err := m.mongo.ListDagInstance(ctx, input)
	if err != nil {
		log.Warnf("[logic.ListDagInstance] ListDagInstance err, detail: %s", err.Error())
		return dagInsRunList, total, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// Step4: compute result
	for _, dag := range dagInsList {
		_dag := dag
		var source any

		if len(_dag.Source) > 0 {
			_ = json.Unmarshal([]byte(_dag.Source), &source)
		}

		var reason any

		if len(_dag.Reason) > 0 {
			err := json.Unmarshal([]byte(_dag.Reason), &reason)
			if err != nil {
				reason = map[string]any{
					"detail": _dag.Reason,
				}
			}
		}

		dagInstanceRunInfo := &DagInstanceRunInfo{
			ID:        _dag.ID,
			StartedAt: _dag.CreatedAt,
			EndedAt:   _dag.EndedAt,
			Source:    source,
			Reason:    reason,
		}

		if _dag.Status == entity.DagInstanceStatusBlocked {
			dagInstanceRunInfo.EndedAt = 0
		}
		switch fmt.Sprintf("%v", _dag.Status) {
		case common.SuccessStatus:
			dagInstanceRunInfo.Status = common.SuccessStatus
		case common.FailedStatus:
			dagInstanceRunInfo.Status = common.FailedStatus
		case common.CanceledStatus:
			dagInstanceRunInfo.Status = common.CanceledStatus
		case common.RunningStatus:
			dagInstanceRunInfo.Status = common.RunningStatus
		default:
			dagInstanceRunInfo.Status = common.ScheduledStatus
		}
		dagInsRunList.DagInstanceRunInfo = append(dagInsRunList.DagInstanceRunInfo, dagInstanceRunInfo)
	}
	if len(dagInsList) == 0 {
		// avoid return null
		dagInsRunList.DagInstanceRunInfo = make([]*DagInstanceRunInfo, 0)
	}

	dagInsRunList.Progress = &Progress{
		Total:   total,
		Success: success,
		Failed:  failed,
	}

	dagInsArrLen := total
	// not use paging
	if len(input.Status) != 0 {
		query["status"] = bson.M{"$in": input.Status}
		dagInsArrLen, err = m.mongo.GetDagInstanceCount(ctx, query)
		if err != nil {
			log.Warnf("[logic.ListDagInstance] GetDagInstanceCount total err, query: %v, detail: %s", query, err.Error())
			return dagInsRunList, total, ierrors.NewIError(ierrors.InternalError, "", nil)
		}
	}

	return dagInsRunList, dagInsArrLen, nil
}

func (m *mgnt) GetDagInstanceCount(ctx context.Context, dagID string, param map[string]interface{}, userInfo *drivenadapters.UserInfo) (int64, error) {
	var err error
	var total int64
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow:      {perm.RunStatisticsOperation},
			common.DagTypeComboOperator: {perm.ViewOperation},
			common.DagTypeDefault:       {perm.OldOnlyAdminOperation},
		},
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dagID, userInfo, opMap)
	if err != nil {
		return total, err
	}

	query := map[string]interface{}{"dagId": dagID}
	if param["start_time"] != nil && param["end_time"] != nil {
		query["createdAt"] = bson.M{"$gte": param["start_time"].(int64), "$lte": param["end_time"].(int64)}
	}

	if param["type"] != nil {
		_status, ok := param["type"].([]string)
		if ok {
			query["status"] = bson.M{"$in": _status}
		}
	}

	total, err = m.mongo.GetDagInstanceCount(ctx, query)
	if err != nil {
		log.Warnf("[logic.GetDagInstanceCount] GetDagInstanceCount success err, detail: %s", err.Error())
		return total, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	return total, nil
}

func (m *mgnt) ListDagInstanceV2(ctx context.Context, dagID string, param map[string]interface{}, userInfo *drivenadapters.UserInfo) ([]*DagInstanceRunInfo, int64, error) {
	var err error
	var total int64
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var dagInsRunList = make([]*DagInstanceRunInfo, 0)

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow:      {perm.RunStatisticsOperation},
			common.DagTypeComboOperator: {perm.ViewOperation},
			common.DagTypeDefault:       {perm.OldOnlyAdminOperation},
		},
	}

	if userInfo.AccountType == common.APP.ToString() {
		opMap.OpMap[common.DagTypeDefault] = []string{perm.OldAppTokenOperation}
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dagID, userInfo, opMap)
	if err != nil {
		return dagInsRunList, total, err
	}

	query := map[string]interface{}{"dagId": dagID}
	if param["start_time"] != nil && param["end_time"] != nil {
		query["createdAt"] = bson.M{"$gte": param["start_time"].(int64), "$lte": param["end_time"].(int64)}
	}

	input := &mod.ListDagInstanceInput{
		DagIDs: []string{dagID},
		Limit:  param["limit"].(int64),
		Offset: param["page"].(int64),
		Order:  -1,
	}

	if param["type"] != nil {
		statusList := []entity.DagInstanceStatus{}
		_status := param["type"].([]string)
		for _, value := range _status {
			_value := value
			switch _value {
			case common.SuccessStatus:
				statusList = append(statusList, entity.DagInstanceStatusSuccess)
			case common.FailedStatus:
				statusList = append(statusList, entity.DagInstanceStatusFailed)
			case common.CanceledStatus:
				statusList = append(statusList, entity.DagInstanceStatusCancled)
			case common.RunningStatus:
				statusList = append(statusList, entity.DagInstanceStatusRunning)
			default:
				statusList = append(statusList, entity.DagInstanceStatusInit,
					entity.DagInstanceStatusScheduled,
					entity.DagInstanceStatusBlocked)
			}
		}
		query["status"] = bson.M{"$in": statusList}
		input.Status = statusList
	}

	if param["name"] != nil {
		filter := bson.M{"$regex": param["name"], "$options": "i"}
		query["keywords"] = filter
		input.MatchQuery = &mod.MatchQuery{
			Field: "keywords",
			Value: filter,
		}
	}

	total, err = m.mongo.GetDagInstanceCount(ctx, query)
	if err != nil {
		log.Errorf("[logic.ListDagInstanceV2] GetDagInstanceCount total err, detail: %s", err.Error())
		return dagInsRunList, total, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	if param["order"].(string) == common.ASC {
		input.Order = 1
	}

	sortBy := param["sortby"].(string)
	if sortBy == common.Started_At {
		input.SortBy = common.CreatedAt
	} else if sortBy == common.Ended_At {
		input.SortBy = common.EndedAt
	}

	if param["start_time"] != nil && param["end_time"] != nil {
		input.TimeRange = &mod.TimeRangeSearch{
			Begin: param["start_time"].(int64),
			End:   param["end_time"].(int64),
			Field: "createdAt",
		}
	}

	input.SelectField = []string{"_id", "createdAt", "endedAt", "status", "source", "reason"}
	dagInsList, err := m.mongo.ListDagInstance(ctx, input)
	if err != nil {
		log.Warnf("[logic.ListDagInstanceV2] ListDagInstance err, detail: %s", err.Error())
		return dagInsRunList, total, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	// Step4: compute result
	for _, dag := range dagInsList {
		_dag := dag
		var source any

		if len(_dag.Source) > 0 {
			_ = json.Unmarshal([]byte(_dag.Source), &source)
		}

		var reason any

		if len(_dag.Reason) > 0 {
			err := json.Unmarshal([]byte(_dag.Reason), &reason)
			if err != nil {
				reason = map[string]any{
					"detail": _dag.Reason,
				}
			}
		}

		dagInstanceRunInfo := &DagInstanceRunInfo{
			ID:        _dag.ID,
			StartedAt: _dag.CreatedAt,
			EndedAt:   _dag.EndedAt,
			Source:    source,
			Reason:    reason,
		}
		if _dag.Status != entity.DagInstanceStatusSuccess && _dag.Status != entity.DagInstanceStatusFailed && _dag.Status != entity.DagInstanceStatusCancled {
			dagInstanceRunInfo.EndedAt = 0
		}
		switch fmt.Sprintf("%v", _dag.Status) {
		case common.SuccessStatus:
			dagInstanceRunInfo.Status = common.SuccessStatus
		case common.FailedStatus:
			dagInstanceRunInfo.Status = common.FailedStatus
		case common.CanceledStatus:
			dagInstanceRunInfo.Status = common.CanceledStatus
		case common.RunningStatus:
			dagInstanceRunInfo.Status = common.RunningStatus
		default:
			dagInstanceRunInfo.Status = common.ScheduledStatus
		}
		dagInsRunList = append(dagInsRunList, dagInstanceRunInfo)
	}

	return dagInsRunList, total, nil
}

// ListTaskInstance list task instance
func (m *mgnt) ListTaskInstance(ctx context.Context, dagID, dagInstanceID string, page, limit int64, userInfo *drivenadapters.UserInfo) ([]*TaskInstanceRunInfo, int64, error) { //nolint
	var taskInsResultList = make([]*TaskInstanceRunInfo, 0)
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow:      {perm.RunStatisticsOperation},
			common.DagTypeComboOperator: {perm.ViewOperation},
			common.DagTypeDefault:       {perm.OldOnlyAdminOperation},
		},
	}

	if userInfo.AccountType == common.APP.ToString() {
		opMap.OpMap[common.DagTypeDefault] = []string{perm.OldAppTokenOperation}
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dagID, userInfo, opMap)
	if err != nil {
		return taskInsResultList, 0, err
	}

	err = m.isDagInstanceExist(ctx, map[string]interface{}{"_id": dagInstanceID, "dagId": dagID})
	if err != nil {
		return taskInsResultList, 0, err
	}

	dagInsInfo, err := m.mongo.GetDagInstanceByFields(ctx, map[string]interface{}{"_id": dagInstanceID, "dagId": dagID})
	if err != nil {
		log.Warnf("[logic.ListTaskInstance] GetDagInstanceByFields err, detail: %s", err.Error())
		return taskInsResultList, 0, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	if err := dagInsInfo.LoadExtData(ctx); err != nil {
		log.Warnf("[logic.ListTaskInstance] dagIns LoadExtData err, detail: %s", err.Error())
		return taskInsResultList, 0, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	var taskInsList []*entity.TaskInstance
	var total int64

	if dagInsInfo.Status == entity.DagInstanceStatusSuccess {
		total, taskInsList, err = m.GenerateTaskResults(ctx, dagID, dagInstanceID, page, limit)
		if err != nil {
			return taskInsResultList, 0, err
		}
	} else {
		query := &mod.ListTaskInstanceInput{
			DagInsID: dagInstanceID,
			SelectField: []string{
				"_id", "name", "actionName", "createdAt", "status", "params", "results", "reason", "updatedAt", "taskId", "lastModifiedAt", "renderedParams", "metadata",
			},
			Limit:  limit,
			Offset: page,
		}

		if dagInsInfo.Mode != entity.DagInstanceModeVM {
			query.Order = 1
			query.SortBy = "lastModifiedAt"
		}

		taskInsList, err = m.mongo.ListTaskInstance(ctx, query)
		if err != nil {
			log.Warnf("[logic.ListTaskInstance] ListTaskInstance err, detail: %s", err.Error())
			return taskInsResultList, 0, ierrors.NewIError(ierrors.InternalError, "", nil)
		}

		total, err = m.mongo.GetTaskInstanceCount(ctx, map[string]any{"dagInsId": dagInstanceID})

		if err != nil {
			log.Warnf("[logic.ListTaskInstance] GetTaskInstanceCount err, detail: %s", err.Error())
			return taskInsResultList, 0, ierrors.NewIError(ierrors.InternalError, "", nil)
		}
	}

	for _, taskIns := range taskInsList {
		_taskIns := taskIns
		taskInstanceRunInfo := &TaskInstanceRunInfo{
			ID:             _taskIns.ID,
			Name:           _taskIns.Name,
			Operator:       _taskIns.ActionName,
			StartedAt:      _taskIns.CreatedAt,
			UpdatedAt:      _taskIns.UpdatedAt,
			TaskID:         _taskIns.TaskID,
			LastModifiedAt: _taskIns.LastModifiedAt,
			MetaData:       _taskIns.MetaData,
		}

		switch _taskIns.Status { //nolint
		case entity.TaskInstanceStatusSuccess:
			taskInstanceRunInfo.Status = fmt.Sprintf("%v", _taskIns.Status)
		case entity.TaskInstanceStatusFailed:
			taskInstanceRunInfo.Status = fmt.Sprintf("%v", _taskIns.Status)
		case entity.TaskInstanceStatusBlocked:
			taskInstanceRunInfo.Status = fmt.Sprintf("%v", _taskIns.Status)
		case entity.TaskInstanceStatusSkipped:
			taskInstanceRunInfo.Status = fmt.Sprintf("%v", _taskIns.Status)
		default:
			taskInstanceRunInfo.Status = common.UndoStatus
		}

		// node status not success and failded skip input and output parse
		if taskInstanceRunInfo.Status == common.UndoStatus {
			taskInsResultList = append(taskInsResultList, taskInstanceRunInfo)
			continue
		}

		_taskIns.RelatedDagInstance = dagInsInfo

		if dagInsInfo.Mode != entity.DagInstanceModeVM {
			if _taskIns.RenderedParams != nil {
				_taskIns.Params = _taskIns.RenderedParams
			} else {
				err := m.renderParamsV2(_taskIns)
				if err != nil {
					log.Warnf("[logic.ListTaskInstance] renderParams err, detail: %s", err.Error())
					taskInstanceRunInfo.Inputs = nil
				}
			}
		}

		taskInstanceRunInfo.Inputs = _taskIns.GetParams()

		if _taskIns.Status == entity.TaskInstanceStatusFailed {
			result := _taskIns.Reason
			switch result := result.(type) {
			case bson.D:
				itemMap := map[string]interface{}{}
				b, _ := bson.Marshal(result)
				bson.Unmarshal(b, &itemMap) //nolint
				taskInstanceRunInfo.Outputs = itemMap
			case string:
				taskInstanceRunInfo.Outputs = result
			default:
				taskInstanceRunInfo.Outputs = result
			}
		} else {
			result := _taskIns.Results
			switch result := result.(type) {
			case bson.D:
				itemMap := map[string]interface{}{}
				b, _ := bson.Marshal(result)
				bson.Unmarshal(b, &itemMap) //nolint
				taskInstanceRunInfo.Outputs = itemMap
			case string:
				taskInstanceRunInfo.Outputs = result
			default:
				taskInstanceRunInfo.Outputs = _taskIns.Results
			}
		}
		taskInsResultList = append(taskInsResultList, taskInstanceRunInfo)
	}
	return taskInsResultList, total, nil
}

func (m *mgnt) ContinueBlockInstances(ctx context.Context, blockedTaskIDs []string, res map[string]interface{}, status entity.TaskInstanceStatus) error {
	log.Infof("[logic.ContinueBlockInstances] status: %s, ids: %v", status, blockedTaskIDs)
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)
	maxRetries := 5

	// 单步调试内存操作数据
	if len(blockedTaskIDs) > 0 && strings.HasPrefix(blockedTaskIDs[0], "DEBUG:") {
		for _, id := range blockedTaskIDs {
			raw, exist := m.memoryCache.GetRaw(id)
			if !exist {
				continue
			}

			taskIns, ok := raw.(*entity.TaskInstance)
			if !ok {
				continue
			}

			taskIns.Results = res
			taskIns.Status = status
			taskIns.RelatedDagInstance.Status = entity.DagInstanceStatusSuccess
			if status == entity.TaskInstanceStatusFailed {
				taskIns.RelatedDagInstance.Status = entity.DagInstanceStatusFailed
			}

			m.memoryCache.Set(id, taskIns, 5*time.Minute)
		}
		return nil
	}

	err = utils.RetryTimes("continueblock", maxRetries, 3*time.Second, func(retryCount int) error {
		ins, err := m.mongo.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
			IDs:    blockedTaskIDs,
			Status: []entity.TaskInstanceStatus{entity.TaskInstanceStatusBlocked},
		})

		if err != nil {
			if retryCount == maxRetries {
				log.Warnf("[logic.ContinueBlockInstances] ListTaskInstance err, detail: %s", err.Error())
			}
			return ierrors.NewIError(ierrors.InternalError, "", nil)
		}

		if len(ins) == 0 {
			return ierrors.NewIError(ierrors.TaskNotFound, "", nil)
		}

		for _, instance := range ins {
			dagIns, err := m.mongo.GetDagInstance(ctx, instance.DagInsID)
			preStatus := dagIns.Status

			if err != nil {
				log.Warnf("[logic.ContinueBlockInstances] GetDagInstance err, detail: %s", err.Error())
				return ierrors.NewIError(ierrors.Forbidden, "", err.Error())
			}

			if dagIns.Status == entity.DagInstanceStatusSuccess || dagIns.Status == entity.DagInstanceStatusFailed ||
				(dagIns.Status == entity.DagInstanceStatusRunning && retryCount != maxRetries) ||
				dagIns.Mode == entity.DagInstanceModeVM && dagIns.Status != entity.DagInstanceStatusBlocked {
				log.Warnf("[logic.ContinueBlockInstances] dag instance status: %s, not blocked, taskInsID: %s", dagIns.Status, instance.ID)
				return ierrors.NewIError(ierrors.Forbidden, "", nil)
			}

			if instance.Results == nil {
				log.Warnf("[logic.ContinueBlockInstances] instance result is nil")
				return ierrors.NewIError(ierrors.InternalError, "", nil)
			}

			instanceData, ok := normalizeutil.AsMap(instance.Results)
			if !ok {
				continue
			}
			groupID, ok := instanceData["group_id"].(string)
			if ok && groupID != "" {
				err := m.usermgnt.DeleteInternalGroup([]string{groupID})
				if err != nil {
					log.Warnf("[logic.ContinueBlockInstances] DeleteInternalGroup, groupID: %s, err: %s", groupID, err.Error())
				}
			}

			if status == entity.TaskInstanceStatusRetrying && dagIns.Mode == entity.DagInstanceModeVM {
				status = entity.TaskInstanceStatusFailed
			}

			instance.Status = status

			var results = make(map[string]interface{}, 0)
			for key, value := range instanceData {
				if key != "group_id" {
					results[key] = value
				}
			}

			for k, v := range res {
				results[k] = v
			}

			instance.Results = results
			if status == entity.TaskInstanceStatusFailed {
				instance.Reason = results
			}
			if err := m.mongo.PatchTaskIns(ctx, instance); err != nil {
				log.Warnf("[logic.ContinueBlockInstances] SetStatus err, detail: %s", err.Error())
				return ierrors.NewIError(ierrors.InternalError, "", nil)
			}
			log.Infof("[logic.ContinueBlockInstances] set status success status: %s, ids: %v", status, blockedTaskIDs)

			err = dagIns.LoadExtData(ctx)

			if err != nil {
				log.Warnf("[logic.ContinueBlockInstances] dagIns LoadExtData err, detail: %s", err.Error())
				return err
			}

			if instance.ActionName == common.WorkflowApproval {
				workflowApprovalTaskIds, ok := dagIns.ShareData.Get(common.WorkflowApprovalTaskIds)
				if ok {
					taskIDs := make([]interface{}, 0)
					if ids, idsOK := normalizeutil.AsSlice(workflowApprovalTaskIds); idsOK {
						taskIDs = append(taskIDs, ids...)
					}
					taskIDs = append(taskIDs, instance.TaskID)

					dagIns.ShareData.Set(common.WorkflowApprovalTaskIds, taskIDs)
				} else {
					dagIns.ShareData.Set(common.WorkflowApprovalTaskIds, []string{instance.TaskID})
				}
			}
			if len(results) > 0 {
				dagIns.ShareData.Set(instance.TaskID, results)
			} else {
				dagIns.ShareData.Set(instance.TaskID, "null")
			}

			if dagIns.EventPersistence == entity.DagInstanceEventPersistenceSql &&
				dagIns.Mode != entity.DagInstanceModeVM {
				var events []*entity.DagInstanceEvent

				taskID := instance.TaskID

				for _, re := range []string{`^\d+_i\d+_s\d+.+_(\d+)$`, `^\d+_i\d+_s(\d+)$`, `^(\d+)_i\d+$`} {
					if matches := regexp.MustCompile(re).FindStringSubmatch(taskID); len(matches) == 2 {
						taskID = matches[1]
						break
					}
				}

				if instance.Status == entity.TaskInstanceStatusFailed {
					events = append(events, &entity.DagInstanceEvent{
						Type:       rds.DagInstanceEventTypeTaskStatus,
						InstanceID: dagIns.ID,
						Operator:   instance.ActionName,
						TaskID:     taskID,
						Status:     string(status),
						Data:       results,
						Timestamp:  time.Now().UnixMicro(),
						Visibility: rds.DagInstanceEventVisibilityPublic,
					})
				} else {
					events = append(events, &entity.DagInstanceEvent{
						Type:       rds.DagInstanceEventTypeTaskStatus,
						InstanceID: dagIns.ID,
						Operator:   instance.ActionName,
						TaskID:     taskID,
						Status:     string(status),
						Timestamp:  time.Now().UnixMicro(),
						Visibility: rds.DagInstanceEventVisibilityPublic,
					})
				}

				// 异步节点执行后置信息,任务的执行时间动态拼接，此逻辑暂时去除
				// key := fmt.Sprintf("__%s_trace_async", instance.TaskID)
				// trace := map[string]any{
				// 	"ended_at": time.Now().UnixMilli(),
				// }
				// events = append(events, &entity.DagInstanceEvent{
				// 	Type:       rds.DagInstanceEventTypeTrace,
				// 	InstanceID: dagIns.ID,
				// 	Name:       key,
				// 	Data:       trace,
				// 	Visibility: rds.DagInstanceEventVisibilityPublic,
				// 	Timestamp:  time.Now().UnixMicro(),
				// })

				err := dagIns.WriteEvents(ctx, events)
				if err != nil {
					return ierrors.NewIError(ierrors.InternalError, "", nil)
				}
			}

			dagIns.Status = entity.DagInstanceStatusInit

			if dagIns.Mode == entity.DagInstanceModeVM {
				dagIns.ResumeStatus = status
				if resumeData, err := json.Marshal([]any{results}); err != nil {
					dagIns.ResumeStatus = entity.TaskInstanceStatusFailed
					dagIns.ResumeData = "[null]"
				} else {
					dagIns.ResumeData = string(resumeData)
				}
			} else {
				if status == entity.TaskInstanceStatusFailed {
					dagIns.Status = entity.DagInstanceStatusFailed
					reason := map[string]any{
						"taskId":     instance.TaskID,
						"name":       instance.Name,
						"actionName": instance.ActionName,
						"detail":     instance.Reason,
					}
					b, _ := json.Marshal(reason)
					dagIns.Reason = string(b)
					mod.HandleSubprocessCallback(ctx, entity.DagInstanceStatusFailed, dagIns, instance)
					// 如果dagIns已经是取消状态，不需要再推一次可观测性日志
					if preStatus != entity.DagInstanceStatusCancled {
						go m.LogDagInsResult(ctx, dagIns)
					}
				}
			}

			if err := dagIns.SaveExtData(context.Background()); err != nil {
				return err
			}

			// 如果取消状态,需还原dagIns的状态
			if preStatus == entity.DagInstanceStatusCancled {
				dagIns.Status = preStatus
			}

			if err := m.mongo.PatchDagIns(ctx, dagIns); err != nil {
				log.Warnf("[logic.ContinueBlockInstances] PatchDagIns err, detail: %s", err.Error())
				return err
			}
		}

		return nil
	})

	return nil
}

// HandleAuditorsMacth 处理匹配审核员结果
func (m *mgnt) HandleAuditorsMacth(ctx context.Context, msg *AuditorInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	applyID := msg.ApplyID
	instances, err := m.mongo.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
		IDs:    []string{applyID},
		Status: []entity.TaskInstanceStatus{entity.TaskInstanceStatusBlocked},
	})

	if err != nil {
		log.Warnf("[logic.HandleAuditorsMacth] ListTaskInstance err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	for index := range instances {
		instance := instances[index]
		// instance.Results 为 nil
		instanceData, ok := normalizeutil.AsMap(instance.Results)
		if !ok {
			continue
		}
		groupID, ok := instanceData["group_id"].(string)
		if !ok || groupID == "" {
			continue
		}
		users, err := m.usermgnt.GetInternalGroupMembers(groupID)
		if err != nil {
			log.Warnf("[HandleAuditorsMacth] GetInternalGroupMembers failed, err: %s", err.Error())
			continue
		}
		var members = append(users, msg.Auditors...)
		members = utils.RemoveRepByMap(members)
		err = m.usermgnt.UpdateInternalGroupMember(groupID, members)
		if err != nil {
			log.Warnf("[HandleAuditorsMacth] UpdateInternalGroupMenmber failed, err: %s", err.Error())
			continue
		}
	}

	return nil
}

// getLastTaskIDs 递归获取步骤列表中的最后一个(或多个)任务ID
// 对于普通步骤,返回该步骤的ID
// 对于并行分支,递归收集所有子分支的最后任务ID
func getLastTaskIDs(steps []entity.Step) []string {
	if len(steps) == 0 {
		return []string{}
	}

	lastStep := steps[len(steps)-1]

	// 如果最后一个步骤是并行节点,递归获取其所有分支的最后任务
	if lastStep.Operator == common.ControlFlowParallel {
		result := []string{}
		for _, branch := range lastStep.Branches {
			branchLastTasks := getLastTaskIDs(branch.Steps)
			result = append(result, branchLastTasks...)
		}
		return result
	}

	// 否则,返回该步骤的ID
	return []string{lastStep.ID}
}

func (m *mgnt) buildTasks(triggerStep *entity.Step, steps []entity.Step, tasks *[]entity.Task, bch *entity.Branch, stepList *[]map[string]interface{}, inheritChecks *entity.PreChecks, branchsID *string) { //nolint
	prechecks := entity.PreChecks{}
	if inheritChecks != nil {
		for k, v := range *inheritChecks {
			prechecks[k] = v
		}
	}
	if bch != nil {
		for index, con := range bch.Conditions {
			check := &entity.Check{
				Conditions: con,
				Act:        entity.ActiveActionSkip,
			}

			prechecks[fmt.Sprintf("%s_%v", *branchsID, index)] = check
			var conMap = []map[string]interface{}{}
			conditionByte, _ := json.Marshal(con)
			json.Unmarshal(conditionByte, &conMap) //nolint
			*stepList = append(*stepList, conMap...)
		}
	}
	dependOn := []string{} // 在循环外声明,以便保留并行分支设置的依赖数组
	for _, step := range steps {
		// 只有在dependOn为空时才从上一个task计算依赖
		// 这样可以保留并行分支设置的多个依赖项
		if len(dependOn) == 0 && len(*tasks) != 0 {
			dependOn = append(dependOn, (*tasks)[len(*tasks)-1].ID)
		}
		var stepMap = map[string]interface{}{}
		stepByte, _ := json.Marshal(step)
		json.Unmarshal(stepByte, &stepMap) //nolint
		delete(stepMap, "branches")
		delete(stepMap, "steps")
		*stepList = append(*stepList, stepMap)

		if step.Operator == actions.ControlFlowBranches {
			task := entity.Task{
				ID:         step.ID,
				ActionName: step.Operator,
				Name:       step.Title,
				DependOn:   dependOn,
				PreChecks:  prechecks,
			}
			*tasks = append(*tasks, task)

			for _, bch := range step.Branches {
				m.buildTasks(triggerStep, bch.Steps, tasks, &bch, stepList, &prechecks, &step.ID)
			}
		} else if step.Operator == common.ControlFlowParallel {
			// 并行分支节点处理
			parallelBranchLastTasks := []string{}

			// 保存进入并行分支前的依赖节点
			// 所有并行分支的第一个节点都应该依赖这个节点
			branchEntryDependOn := dependOn

			for _, branch := range step.Branches {
				// 记录当前tasks的长度，用于确定分支起始位置
				branchStartIdx := len(*tasks)

				// 递归处理每个分支内的 steps
				// 这样可以支持嵌套的并行分支、条件分支、循环等控制流结构
				m.buildTasks(triggerStep, branch.Steps, tasks, &branch, stepList, &prechecks, &step.ID)

				// 修正该分支第一个task的依赖关系
				// 并行分支中不同分支的step不会互相依赖
				// 各个分支第一个节点依赖的都是进入并行分支前的那个节点
				if branchStartIdx < len(*tasks) {
					(*tasks)[branchStartIdx].DependOn = branchEntryDependOn
				}

				// 找到该分支的最后一个(或多个)task ID
				// 使用递归函数处理嵌套并行分支的情况
				branchLastTasks := getLastTaskIDs(branch.Steps)
				parallelBranchLastTasks = append(parallelBranchLastTasks, branchLastTasks...)
			}

			// 更新dependOn，后续step依赖所有分支的最后一个task
			dependOn = parallelBranchLastTasks
		} else if step.Operator == common.Loop {
			task := entity.Task{
				ID:         step.ID,
				ActionName: step.Operator,
				Name:       step.Title,
				DependOn:   dependOn,
				PreChecks:  prechecks,
				Params:     step.Parameters,
				Steps:      step.Steps,
			}
			*tasks = append(*tasks, task)
			// m.buildTasks(triggerStep, step.Steps, tasks, nil, stepList, &prechecks, &step.ID)
		} else {
			pre := prechecks
			isCycle := m.chargeCycle(triggerStep, &step)
			if isCycle {
				con := []entity.TaskCondition{
					{
						ID: "0000000000",
						Op: entity.OperateStringEq,
						Parameter: entity.TaskConditionParameter{
							A: "0",
							B: "1",
						},
					},
				}
				pre = entity.PreChecks{"end": &entity.Check{
					Conditions: con,
					Act:        entity.ActiveActionSkip,
				}}
			}

			// 审核节点或图谱写入节点，显式的配置了超时和失败重试策略
			if step.Operator == common.WorkflowApproval || step.Operator == common.IntelliinfoTranfer {
				step.Settings = nil
			}

			task := entity.Task{
				ID:         step.ID,
				ActionName: step.Operator,
				Name:       step.Title,
				DependOn:   dependOn,
				Params:     step.Parameters,
				PreChecks:  pre,
				Settings:   step.Settings,
			}

			if step.Settings == nil {
				task.TimeoutSecs = m.taskTimeoutConfig.GetTimeout(step.Operator)
			} else {
				task.TimeoutSecs = step.Settings.TimeOut.Delay
			}

			// 看门狗和超时策略使用同一个超时时间会存在资源竞争，因此看门狗在此基础上增加60s的时间窗口
			task.TimeoutSecs += 60

			*tasks = append(*tasks, task)
		}

		// 清空dependOn,让下一个step重新计算依赖
		// 但并行分支除外,因为它已经设置了正确的多依赖数组
		if step.Operator != common.ControlFlowParallel {
			dependOn = []string{}
		}
	}
}

// getTriggerType 任务触发类型
func (m *mgnt) getTriggerType(trigger string) entity.Trigger {
	switch trigger {
	case common.MannualTrigger:
		return entity.TriggerManually
	case common.EventTrigger,
		common.AnyshareFileCopyTrigger,
		common.AnyshareFileUploadTrigger,
		common.AnyshareFileMoveTrigger,
		common.AnyshareFileRemoveTrigger,
		common.AnyshareFolderCreateTrigger,
		common.AnyshareFolderMoveTrigger,
		common.AnyshareFolderCopyTrigger,
		common.AnyshareFolderRemoveTrigger,
		common.AnyshareFileVersionUpdateTrigger,
		common.AnyshareFilePathUpdateTrigger,
		common.AnyshareFileVersionDeleteTrigger,
		common.AnyshareUserDeleteTrigger,
		common.AnyshareUserChangeTrigger,
		common.AnyshareUserFreezeTrigger,
		common.AnyshareUserCreateTrigger,
		common.AnyshareUserUpdateDeptTrigger,
		common.AnyshareOrgNameModifyTrigger,
		common.AnyshareDeptDeleteTrigger,
		common.AnyshareDeptCreateTrigger,
		common.AnyshareUserMovedTrigger,
		common.AnyshareDeptMovedTrigger,
		common.AnyshareUserAddDeptTrigger,
		common.AnyshareUserRemoveDeptTrigger,
		common.AnyshareTagTreeCreateTrigger,
		common.AnyshareTagTreeAddedTrigger,
		common.AnyshareTagTreeEditedTrigger,
		common.AnyshareTagTreeDeletedTrigger:
		return entity.TriggerEvent
	case common.CronTrigger, common.CronWeekTrigger, common.CronMonthTrigger, common.CronCustomTrigger:
		return entity.TriggerCron
	case common.WebhookTrigger:
		return entity.TriggerWebhook
	case common.FormTrigger:
		return entity.TriggerForm
	case common.OpAnyShareSelectedFileTrigger, common.OpAnyShareSelectedFolderTrigger:
		return entity.TriggerDocument
	case common.SecurityPolicyTrigger:
		return entity.TriggerSecurityPolicy
	default:
		return entity.TriggerManually
	}
}

// 数据源类型
func (m *mgnt) triggerfromDataSource(ctx context.Context, dataSource *entity.DataSource, userid, token, ip string, triggerType entity.Trigger, runVar map[string]string, dag *entity.Dag) (bool, error) { //nolint
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	if dataSource == nil {
		return false, nil
	}
	var datas = make([]*DataSourceItem, 0)
	switch dataSource.Operator {
	case common.AnyshareDataSpecifyFiles, common.AnyshareDataSpecifyFolders:
		for _, docID := range dataSource.Parameters.DocIDs {
			attrs, err := m.efast.ConvertPath(ctx, docID, strings.TrimPrefix(token, "Bearer "), ip)
			if err != nil {
				parsedError, _err := ierrors.ExHTTPErrorParser(err)
				if _err != nil {
					return false, err
				}
				if parsedError["code"] != float64(DocNotFound) {
					return false, err
				}
				continue
			}
			datas = append(datas, &DataSourceItem{
				ID:       docID,
				Keywords: []string{attrs.Name},
			})
		}
		if len(datas) > 0 {
			err = m.createDagInstanceFromDataSource(ctx, triggerType, runVar, datas, dag)
			if err != nil {
				log.Warnf("[triggerfromDataSource] createDagInstanceFromDataSource failed: %s, dagId: %s", err.Error(), dag.ID)
			}
			return true, err
		} else {
			return false, nil
		}

	case common.AnyshareDataListFolders:
		docid := dataSource.Parameters.DocID
		docids := dataSource.Parameters.DocIDs
		depth := dataSource.Parameters.Depth
		targets := []string{}
		if depth == 0 {
			depth = 1
		}
		if docid != "" {
			targets = append(targets, docid)
		}
		for _, _id := range docids {
			if _id == common.KnowledgeDocLib || _id == common.CustomDocLib || _id == common.DepartmentDocLib {
				docLibType := strings.Split(_id, "_")[0]
				libs, libErr := m.efast.GetDocLibs(ctx, docLibType, token, ip)
				if libErr != nil {
					traceLog.WithContext(ctx).Warnf("[logic.getDataSource] GetDocLibs err, detail: %s, id: %s", libErr.Error(), _id)
					continue
				}
				for _, lib := range libs {
					targets = append(targets, lib.ID)
				}
			} else {
				targets = append(targets, _id)
			}
		}
		for _, _id := range targets {
			gerr := m.handleFoldersFromSource(ctx, depth, _id, userid, token, ip, triggerType, runVar, dag)
			if gerr != nil {
				traceLog.WithContext(ctx).Warnf("[logic.getDataSource] handleFoldersFromSource err, detail: %s, id: %s", gerr.Error(), _id)
				continue
			}
		}
		return true, nil
	case common.AnyshareDataListFiles:
		docid := dataSource.Parameters.DocID
		docids := dataSource.Parameters.DocIDs
		depth := dataSource.Parameters.Depth
		targets := []string{}
		if depth == 0 {
			depth = 1
		}
		if docid != "" {
			targets = append(targets, docid)
		}
		for _, _id := range docids {
			if _id == common.KnowledgeDocLib || _id == common.CustomDocLib || _id == common.DepartmentDocLib {
				docLibType := strings.Split(_id, "_")[0]
				libs, libErr := m.efast.GetDocLibs(ctx, docLibType, token, ip)
				if libErr != nil {
					traceLog.WithContext(ctx).Warnf("[logic.getDataSource] GetDocLibs err, detail: %s, id: %s", libErr.Error(), _id)
					continue
				}
				for _, lib := range libs {
					targets = append(targets, lib.ID)
				}
			} else {
				targets = append(targets, _id)
			}
		}
		for _, _id := range targets {
			gerr := m.handleFilesFromSource(ctx, depth, _id, userid, token, ip, triggerType, runVar, dag)
			if gerr != nil {
				traceLog.WithContext(ctx).Warnf("[logic.getDataSource] getFilesFromSource err, detail: %s, id: %s", gerr.Error(), _id)
				continue
			}
		}
		return true, nil
	case common.AnyshareDataDepartment:
		departmentIDMap := make(map[string]string)
		accessorid := dataSource.Parameters.AccessorID
		if accessorid == "00000000-0000-0000-0000-000000000000" {
			// 从根目录开始获取
			depInfos, err := m.usermgnt.GetDepartments(0)
			if err != nil {
				parsedError, _ := ierrors.ExHTTPErrorParser(err)
				return false, ierrors.NewIError(ierrors.TaskSourceInvalid, "", parsedError["detail"])
			}
			for _, depInfo := range *depInfos {
				departmentIDMap[depInfo.ID] = common.Department.ToString()
			}
		} else if accessorid != "" {
			departmentIDMap[accessorid] = common.Department.ToString()
		}

		for source, _ := range departmentIDMap {
			childDeptIDs := m.getDepartmentMembers(ctx, source)
			if err != nil {
				continue
			}
			for _, id := range childDeptIDs {
				departmentIDMap[id] = common.Department.ToString()
			}
		}
		if len(departmentIDMap) > 0 {

			nameMap, err := m.usermgnt.GetNameByAccessorIDs(departmentIDMap)

			if err != nil {
				return false, nil
			}

			for id, _ := range departmentIDMap {
				item := &DataSourceItem{
					ID: id,
				}
				if name, ok := nameMap[id]; ok {
					item.Keywords = []string{name}
				}
				datas = append(datas, item)
			}

			err = m.createDagInstanceFromDataSource(ctx, triggerType, runVar, datas, dag)
			if err != nil {
				log.Warnf("[triggerfromDataSource] createDagInstanceFromDataSource failed: %s, dagId: %s", err.Error(), dag.ID)
			}
			return true, err
		} else {
			return false, nil
		}
	case common.AnyshareDataUser:
		var sources = make([]string, 0)
		accessorid := dataSource.Parameters.AccessorID
		if accessorid == "00000000-0000-0000-0000-000000000000" {
			// 从根目录开始获取
			depInfos, err := m.usermgnt.GetDepartments(0)
			if err != nil {
				parsedError, _ := ierrors.ExHTTPErrorParser(err)
				return false, ierrors.NewIError(ierrors.TaskSourceInvalid, "", parsedError["detail"])
			}
			for _, depInfo := range *depInfos {
				sources = append(sources, depInfo.ID)
			}
		} else if accessorid != "" {
			sources = append(sources, accessorid)
		}

		userIDMap := make(map[string]string)
		for _, source := range sources {
			userIDs := m.getUserMembers(ctx, source)
			for _, id := range userIDs {
				userIDMap[id] = common.User.ToString()
			}
		}

		if len(userIDMap) > 0 {
			nameMap, err := m.usermgnt.GetNameByAccessorIDs(userIDMap)
			if err != nil {
				return false, nil
			}
			for id, _ := range userIDMap {
				item := &DataSourceItem{
					ID: id,
				}
				if name, ok := nameMap[id]; ok {
					item.Keywords = []string{name}
				}
				datas = append(datas, item)
			}
		}
		if len(datas) > 0 {
			err = m.createDagInstanceFromDataSource(ctx, triggerType, runVar, datas, dag)
			if err != nil {
				log.Warnf("[triggerfromDataSource] createDagInstanceFromDataSource failed: %s, dagId: %s", err.Error(), dag.ID)
			}
			return true, err
		} else {
			return false, nil
		}
	case common.AnyshareDataTagTree:
		tagTrees, err := m.ecotag.GetTagTrees(ctx)
		if err != nil {
			parsedError, _ := ierrors.ExHTTPErrorParser(err)
			return false, ierrors.NewIError(ierrors.TaskSourceInvalid, "", parsedError["detail"])
		}

		datas = m.getTags(tagTrees)
		if len(datas) > 0 {
			err = m.createDagInstanceFromDataSource(ctx, triggerType, runVar, datas, dag)
			if err != nil {
				log.Warnf("[triggerfromDataSource] createDagInstanceFromDataSource failed: %s, dagId: %s", err.Error(), dag.ID)
			}
			return true, err
		} else {
			return false, nil
		}
	case common.S3DataListObjects:
		// S3数据源处理
		fmt.Println("Processing S3 data source for dag: ", dag.ID)
		log.Infof("[triggerfromDataSource] Processing S3 data source for dag: %s", dag.ID)

		// 创建一个临时的 DagInstance 用于存储 S3 数据
		dagIns := &entity.DagInstance{
			DagID:     dag.ID,
			ShareData: &entity.ShareData{},
		}

		// 调用 handleS3DataSource 获取 S3 对象列表
		err := m.handleS3DataSource(ctx, dag, dagIns)
		if err != nil {
			log.Warnf("[triggerfromDataSource] handleS3DataSource failed: %s, dagId: %s", err.Error(), dag.ID)
			return false, err
		}

		if dagIns.ShareData != nil && dagIns.ShareData.Dict != nil {
			for key, value := range dagIns.ShareData.Dict {
				fmt.Printf("  Key: %v (type: %T), Value type: %T\n", key, key, value)
			}
		} else {
			fmt.Printf("  ShareData or Dict is nil\n")
		}

		// 从 ShareData 中获取 S3 对象列表
		s3Data, ok := dagIns.ShareData.Get("__0")
		if !ok {
			log.Warnf("[triggerfromDataSource] No S3 data found in ShareData for dag: %s", dag.ID)
			return false, nil
		}

		s3DataMap, ok := s3Data.(map[string]interface{})
		if !ok {
			log.Warnf("[triggerfromDataSource] Invalid S3 data format for dag: %s", dag.ID)
			return false, nil
		}

		items, ok := s3DataMap["items"]
		if !ok {
			log.Warnf("[triggerfromDataSource] No items in S3 data for dag: %s", dag.ID)
			return false, nil
		}

		// 将 S3 对象转换为 DataSourceItem
		s3Items, ok := items.([]S3DataItem)
		if !ok {
			log.Warnf("[triggerfromDataSource] Invalid S3 items format for dag: %s", dag.ID)
			return false, nil
		}

		for _, s3Item := range s3Items {
			data := map[string]interface{}{
				"id":            s3Item.ID,
				"bucket":        s3Item.Bucket,
				"key":           s3Item.Key,
				"name":          s3Item.Name,
				"size":          s3Item.Size,
				"last_modified": s3Item.LastModified,
				"etag":          s3Item.ETag,
				"download_url":  s3Item.DownloadURL,
			}
			datas = append(datas, &DataSourceItem{
				ID:       s3Item.Key,
				Keywords: []string{s3Item.Bucket, s3Item.Key},
				Data:     data,
			})
		}

		if len(datas) > 0 {
			err = m.createDagInstanceFromDataSource(ctx, triggerType, runVar, datas, dag)
			if err != nil {
				log.Warnf("[triggerfromDataSource] createDagInstanceFromDataSource failed: %s, dagId: %s", err.Error(), dag.ID)
			}
			return true, err
		} else {
			log.Warnf("[triggerfromDataSource] No S3 objects found for dag: %s", dag.ID)
			return false, nil
		}
	default:
		return false, nil
	}

}

func (m *mgnt) getDepartmentMembers(ctx context.Context, departmentID string) []string {
	var datas = make([]string, 0)
	members, err := m.usermgnt.GetDepartmentMemberIDs(departmentID)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[getDepartmentMembers] GetDepartmentMemberIDs err, detail: %s", err.Error())
	}
	deptIDs := members.DepartmentIDs
	datas = append(datas, deptIDs...)
	for _, deptID := range deptIDs {
		ids := m.getDepartmentMembers(ctx, deptID)
		datas = append(datas, ids...)
	}
	return datas
}

func (m *mgnt) getUserMembers(ctx context.Context, departmentID string) []string {
	var datas = make([]string, 0)
	members, err := m.usermgnt.GetDepartmentMemberIDs(departmentID)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[getUserMembers] GetDepartmentMemberIDs err, detail: %s", err.Error())
	}
	deptIDs := members.DepartmentIDs
	userIDs := members.UserIDs
	datas = append(datas, userIDs...)
	for _, deptID := range deptIDs {
		ids := m.getUserMembers(ctx, deptID)
		datas = append(datas, ids...)
	}
	return datas
}

func (m *mgnt) getTags(tagTrees []*drivenadapters.TagTree) []*DataSourceItem {
	var datas = make([]*DataSourceItem, 0)
	for _, tags := range tagTrees {
		datas = append(datas, &DataSourceItem{
			ID:       tags.ID,
			Keywords: []string{tags.Name},
		})
		childTagTrees := tags.ChildTags
		childTags := m.getTags(childTagTrees)
		datas = append(datas, childTags...)
	}
	return datas
}

func (m *mgnt) handleFilesFromSource(ctx context.Context, depth int, docid, userid, token, ip string, triggerType entity.Trigger, runVar map[string]string, dag *entity.Dag) error {
	var datas = make([]*DataSourceItem, 0)
	files, dirs, err := m.efast.ListDir(ctx, docid, strings.TrimPrefix(token, "Bearer "), ip)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[handleFilesFromSource] ListDir err, detail: %s", err.Error())
		parsedError, _err := ierrors.ExHTTPErrorParser(err)
		if _err != nil {
			return err
		}
		if parsedError["code"] == float64(DocNotFound) {
			m.handleTriggerError(ctx, triggerType, runVar, dag, err)
			return ierrors.NewIError(ierrors.TaskSourceNotFound, "", parsedError["detail"])
		}
		if parsedError["code"] == float64(DocNoPerm) {
			m.handleTriggerError(ctx, triggerType, runVar, dag, err)
			return ierrors.NewIError(ierrors.TaskSourceNoPerm, "", parsedError["detail"])
		}
		if parsedError["code"] == float64(TokenExpired) {
			tokenMgnt := mod.NewTokenMgnt(userid)
			newToken, _ := tokenMgnt.GetUserToken(token, userid)
			err := m.handleFilesFromSource(ctx, depth, docid, userid, newToken.Token, ip, triggerType, runVar, dag)
			if err != nil {
				traceLog.WithContext(ctx).Warnf("[logic.getFilesFromSource] getFilesFromSource err, detail: %s, id: %s", err.Error(), docid)
			}
			return nil
		}
		m.handleTriggerError(ctx, triggerType, runVar, dag, err)
		return ierrors.NewIError(ierrors.TaskSourceInvalid, "", parsedError["detail"])
	}

	for _, file := range files {
		info := file.(map[string]interface{})
		id := info["docid"].(string)
		name := info["name"].(string)
		datas = append(datas, &DataSourceItem{
			ID:       id,
			Keywords: []string{name},
		})
	}

	if len(datas) > 0 {
		err = m.createDagInstanceFromDataSource(ctx, triggerType, runVar, datas, dag)
		if err != nil {
			log.Warnf("[triggerfromDataSource] createDagInstanceFromDataSource failed: %s, dagId: %s", err.Error(), dag.ID)
			return err
		}
	}

	if depth == 1 {
		return nil
	}

	for _, dir := range dirs {
		info := dir.(map[string]interface{})
		id := info["docid"].(string)
		err := m.handleFilesFromSource(ctx, depth-1, id, userid, token, ip, triggerType, runVar, dag)
		if err != nil {
			traceLog.WithContext(ctx).Warnf("[logic.getFilesFromSource] getFilesFromSource err, detail: %s, id: %s", err.Error(), id)
			continue
		}
	}

	return nil
}

func (m *mgnt) handleFoldersFromSource(ctx context.Context, depth int, docid, userid, token, ip string, triggerType entity.Trigger, runVar map[string]string, dag *entity.Dag) error {
	var datas = make([]*DataSourceItem, 0)
	log := traceLog.WithContext(ctx)
	debug := isDebugEnabled()

	if debug {
		log.Debugf("[handleFoldersFromSource] start traversing folder: %s, depth: %d", docid, depth)
	}

	_, dirs, err := m.efast.ListDir(ctx, docid, strings.TrimPrefix(token, "Bearer "), ip)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[handleFoldersFromSource] ListDir err, detail: %s", err.Error())
		parsedError, _err := ierrors.ExHTTPErrorParser(err)
		if _err != nil {
			return err
		}
		if parsedError["code"] == float64(DocNotFound) {
			m.handleTriggerError(ctx, triggerType, runVar, dag, err)
			return ierrors.NewIError(ierrors.TaskSourceNotFound, "", parsedError["detail"])
		}
		if parsedError["code"] == float64(DocNoPerm) {
			m.handleTriggerError(ctx, triggerType, runVar, dag, err)
			return ierrors.NewIError(ierrors.TaskSourceNoPerm, "", parsedError["detail"])
		}
		if parsedError["code"] == float64(TokenExpired) {
			tokenMgnt := mod.NewTokenMgnt(userid)
			newToken, terr := tokenMgnt.GetUserToken(token, userid)
			if terr != nil {
				traceLog.WithContext(ctx).Warnf("[logic.handleFoldersFromSource] GetUserToken err, detail: %s, id: %s", terr.Error(), docid)
				return terr
			}
			err := m.handleFoldersFromSource(ctx, depth, docid, userid, newToken.Token, ip, triggerType, runVar, dag)
			if err != nil {
				traceLog.WithContext(ctx).Warnf("[logic.handleFoldersFromSource] handleFoldersFromSource err, detail: %s, id: %s", err.Error(), docid)
			}
			return nil
		}
		m.handleTriggerError(ctx, triggerType, runVar, dag, err)
		return ierrors.NewIError(ierrors.TaskSourceInvalid, "", parsedError["detail"])
	}

	for _, dir := range dirs {
		info := dir.(map[string]interface{})
		id := info["docid"].(string)
		name := info["name"].(string)
		datas = append(datas, &DataSourceItem{
			ID:       id,
			Keywords: []string{name},
		})
		if debug {
			log.Debugf("[handleFoldersFromSource] found subfolder: %s, name: %s", id, info["name"])
		}
	}

	if debug {
		log.Debugf("[handleFoldersFromSource] found %d subfolders in folder: %s", len(datas), docid)
	}

	err = m.createDagInstanceFromDataSource(ctx, triggerType, runVar, datas, dag)
	if err != nil {
		log.Warnf("[triggerfromDataSource] createDagInstanceFromDataSource failed: %s, dagId: %s", err.Error(), dag.ID)
		return err
	}

	if depth == 1 {
		if debug {
			log.Debugf("[handleFoldersFromSource] reached max depth for folder: %s", docid)
		}
		return nil
	}

	for _, item := range datas {
		if debug {
			log.Debugf("[handleFoldersFromSource] recursively traversing subfolder: %s, remaining depth: %d", item.ID, depth-1)
		}
		err := m.handleFoldersFromSource(ctx, depth-1, item.ID, userid, token, ip, triggerType, runVar, dag)
		if err != nil {
			traceLog.WithContext(ctx).Warnf("[logic.getFilesFromSource] getFilesFromSource err, detail: %s, id: %s", err.Error(), item.ID)
			continue
		}
	}

	if debug {
		log.Debugf("[handleFoldersFromSource] finished traversing folder: %s", docid)
	}
	return nil
}

// handleTriggerError 处理触发器节点执行失败
func (m *mgnt) handleTriggerError(ctx context.Context, triggerType entity.Trigger, runVar map[string]string, dag *entity.Dag, err error) error {
	dagIns, dagErr := dag.Run(ctx, triggerType, runVar)

	if dagErr != nil {
		log.Warnf("[logic.handleTriggerError] dag.Run err: %s", err.Error())
		return err
	}

	dagIns.Initial()
	dagIns.Status = entity.DagInstanceStatusFailed
	taskIns := &entity.TaskInstance{
		TaskID:     dag.Steps[0].ID,
		DagInsID:   dagIns.ID,
		Name:       dag.Steps[0].Title,
		ActionName: dag.Steps[0].Operator,
		Params:     dag.Steps[0].Parameters,
		Status:     entity.TaskInstanceStatusFailed,
		Reason:     err,
	}

	err = m.mongo.WithTransaction(ctx, func(nctx context.Context, ms mod.Store) error {
		_, dbErr := ms.CreateDagIns(nctx, dagIns)

		if dbErr != nil {
			log.Warnf("[logic.handleTriggerError] CreateDagIns err: %s", dbErr.Error())
			return err
		}

		dbErr = ms.CreateTaskIns(nctx, taskIns)
		if dbErr != nil {
			log.Warnf("[logic.handleTriggerError] CreateTaskIns err: %s", dbErr.Error())
			return err
		}

		return nil
	})

	return err
}

// isDagExist check dag  whether exist
func (m *mgnt) isDagExist(ctx context.Context, param map[string]interface{}) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	total, err := m.mongo.GetDagCount(ctx, param)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[logic.isDagExist] GetDagCount err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	if total == 0 {
		return ierrors.NewIError(ierrors.TaskNotFound, "", map[string]interface{}{"param": param})
	}
	return nil
}

// isDagInstanceExist check dag ins whether exist
func (m *mgnt) isDagInstanceExist(ctx context.Context, param map[string]interface{}) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	total, err := m.mongo.GetDagInstanceCount(ctx, param)
	if err != nil {
		traceLog.WithContext(ctx).Warnf("[logic.isDagInstanceExist] GetDagInstanceCount err, detail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	if total == 0 {
		return ierrors.NewIError(ierrors.DagInsNotFound, "", map[string]interface{}{"param": param})
	}
	return nil
}

// chargeCycle 判断节点是否为导致循环事件触发
func (m *mgnt) chargeCycle(triggerStep, step *entity.Step) bool {
	switch triggerStep.Operator {
	case common.AnyshareFileUploadTrigger:
		// 上传文件操作时,判断后续step是否有新建文件节点,存在新建节点则跳过当前执行步骤
		if step.Operator != common.AnyshareFileCreateOpt {
			return false
		}

		parameters := triggerStep.Parameters
		docIDList := parameters["docids"].([]interface{})
		// 前端创建流程如果不选择应用到子文件夹，参数不包含此字段
		inherit, ok := parameters["inherit"].(bool)
		if !ok {
			inherit = false
		}
		for _, val := range docIDList {
			_val := fmt.Sprintf("%v", val)
			if inherit && strings.HasPrefix(step.Parameters["docid"].(string), _val) ||
				!inherit && step.Parameters["docid"].(string) == _val {
				return true
			}
		}
		return false
	default:
		return false
	}
	// switch triggerStep.Operator {
	// case common.AnyshareFileCopyTrigger:
	// 	// 复制文件触发时
	// 	if step.Operator != common.AnyshareFileCopyOpt {
	// 		return false
	// 	}
	// 	if step.Parameters["destparent"] != triggerStep.Parameters["docid"] && triggerStep.Parameters["inherit"] == false {
	// 		return false
	// 	}
	// 	dest := step.Parameters["destparent"].(string)
	// 	docid := triggerStep.Parameters["docid"].(string)
	// 	if !strings.Contains(dest, docid) {
	// 		return false
	// 	}
	// 	return true
	// case common.AnyshareFolderCopyTrigger:
	// 	// 复制文件夹触发时
	// 	if step.Operator != common.AnyshareFolderCopyOpt {
	// 		return false
	// 	}
	// 	if step.Parameters["destparent"] != triggerStep.Parameters["docid"] && triggerStep.Parameters["inherit"] == false {
	// 		return false
	// 	}
	// 	dest := step.Parameters["destparent"].(string)
	// 	docid := triggerStep.Parameters["docid"].(string)
	// 	if !strings.Contains(dest, docid) {
	// 		return false
	// 	}
	// 	return true
	// case common.AnyshareFolderCreateTrigger:
	// 	// 新建文件夹触发时
	// 	if step.Operator != common.AnyshareFloderCreateOpt {
	// 		return false
	// 	}
	// 	if step.Parameters["docid"] != triggerStep.Parameters["docid"] && triggerStep.Parameters["inherit"] == false {
	// 		return false
	// 	}
	// 	dest := step.Parameters["docid"].(string)
	// 	docid := triggerStep.Parameters["docid"].(string)
	// 	if !strings.Contains(dest, docid) {
	// 		return false
	// 	}
	// 	return true
	// default:
	// 	return false
	// }
}

func (m *mgnt) UpdateTaskResults(ctx context.Context, taskId string, results map[string]interface{}, userInfo *drivenadapters.UserInfo) error {

	task, err := m.mongo.GetTaskIns(ctx, taskId)

	if err != nil {
		return err
	}

	if task.ActionName != common.WorkflowApproval {
		return ierrors.NewIError(ierrors.InvalidParameter, "", map[string]interface{}{"taskId": taskId})
	}

	if task.Status != entity.TaskInstanceStatusBlocked {
		return ierrors.NewIError(ierrors.Forbidden, "", map[string]interface{}{"taskId": taskId})
	}

	dagIns, err := m.mongo.GetDagInstance(ctx, task.DagInsID)

	if err != nil {
		return err
	}

	taskResults := utils.BsonToInterface(task.Results)

	taskResultsMap, ok := taskResults.(map[string]interface{})

	if !ok {
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	groupID, ok := taskResultsMap["group_id"].(string)

	if ok {
		users, err := m.usermgnt.GetInternalGroupMembers(groupID)

		if err != nil || !utils.IsContain(userInfo.UserID, users) {
			return ierrors.NewIError(ierrors.Forbidden, "", err)
		}

		hasChanged := false

		for key, value := range results {
			if content, ok := taskResultsMap[key].(map[string]interface{}); ok {

				if content["allowModifyByAuditor"] == true {
					content["value"] = value
				}

				hasChanged = true
			}
		}

		if hasChanged {
			task.Results = taskResultsMap
			err := m.mongo.PatchTaskIns(ctx, task)
			if err != nil {
				return ierrors.NewIError(ierrors.InternalError, "", nil)
			}

			param := utils.BsonToInterface(task.Params["contents"])
			inputContents, ok := param.([]interface{})
			if ok {

				auditMsgData := make(map[string]interface{})
				topCsfLevel := 0
				docids := make([]string, 0)
				abstractInfo := map[string]string{}

				if title, ok := taskResultsMap["title"].(string); ok {
					abstractInfo["text"] = title
				}

				msgForEmail := []string{}
				msgForLog := []string{}
				msgContent := map[string]string{}

				var permMap = map[string]string{
					"display":  common.GetLocale("display"),
					"preview":  common.GetLocale("preview"),
					"cache":    common.GetLocale("cache"),
					"download": common.GetLocale("download"),
					"create":   common.GetLocale("create"),
					"modify":   common.GetLocale("modify"),
					"delete":   common.GetLocale("delete"),
				}

				if dagIns.DagType == common.DagTypeSecurityPolicy {
					err = dagIns.LoadExtData(ctx)
					if err != nil {
						return ierrors.NewIError(ierrors.InternalError, "", nil)
					}

					source, exists := dagIns.ShareData.Get("__source")
					if exists {
						auditMsgData["source"] = source
						if m, ok := source.(map[string]interface{}); ok {
							if m["type"] == "file" || m["type"] == "folder" {
								abstractInfo["icon"] = m["type"].(string)
								n, ok := m["name"].(string)
								if ok {
									abstractInfo["text"] = n
								}
								docids = append(docids, m["id"].(string))

								msgContent["doc_name"] = fmt.Sprintf("%s:%v", common.GetLocale("doc_name"), m["name"])
								msgContent["doc_id"] = fmt.Sprintf("%s:%v", common.GetLocale("doc_id"), m["id"])
								msgContent["doc_path"] = fmt.Sprintf("%s:%v", common.GetLocale("doc_path"), m["path"])
								msgForLog = append(msgForLog, "doc_name", "doc_id", "doc_path")
							}
						}
					}
				}

				contents := make([]interface{}, 0)
				for k, content := range inputContents {
					if item, ok := content.(map[string]interface{}); ok {

						key := fmt.Sprintf("contents_%d", k)

						if c, ok := taskResultsMap[key].(map[string]interface{}); ok {
							item["value"] = c["value"]
						}

						if item["type"] == "asFile" || item["type"] == "asFolder" || item["type"] == "asDoc" {
							docID := item["value"].(string)

							attr, err := m.efast.GetDocMsg(ctx, docID)
							if err != nil {
								msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
								msgForEmail = append(msgForEmail, key)
								msgForLog = append(msgForLog, key)
								continue
							}
							docids = append(docids, docID)

							msgContent[key] = fmt.Sprintf("%v:%v", item["title"], attr.Name)
							msgForEmail = append(msgForEmail, key)
							msgForLog = append(msgForLog, key)

							if attr.CsfLevel > float64(topCsfLevel) {
								topCsfLevel = int(attr.CsfLevel)
							}
						} else if item["type"] == "asTags" {

							if value, ok := item["value"].(string); ok {
								var tags []string
								err := json.Unmarshal([]byte(value), &tags)

								if err != nil {
									msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
								} else {
									msgContent[key] = fmt.Sprintf("%v:%v", item["title"], strings.Join(tags, ", "))
								}
							} else {
								msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
							}

							msgForLog = append(msgForLog, key)
						} else if item["type"] == "asLevel" {
							msgContent[key] = fmt.Sprintf("%v:%v", item["title"], "")
							msgForLog = append(msgForLog, key)
						} else if item["type"] == "asMetadata" {
							msgContent[key] = fmt.Sprintf("%v:%v", item["title"], "")
							msgForLog = append(msgForLog, key)
						} else if item["type"] == "asPerm" {

							var perm actions.PermValue
							var bytes []byte

							if value, ok := item["value"].(string); ok {
								bytes = []byte(value)
							} else {
								bytes, _ = json.Marshal(item["value"])
							}
							err := json.Unmarshal(bytes, &perm)

							if err != nil {
								msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
							} else {
								var permStr string
								if len(perm.Deny) == 7 {
									permStr = common.GetLocale("deny_all")
								} else {
									var allowSegments []string
									for _, item := range perm.Allow {
										if s, ok := permMap[item]; ok {
											allowSegments = append(allowSegments, s)
										} else {
											allowSegments = append(allowSegments, item)
										}
									}

									permStr = strings.Join(allowSegments, ", ")

									if len(perm.Deny) > 0 {
										var denySegments []string
										for _, item := range perm.Deny {
											if s, ok := permMap[item]; ok {
												denySegments = append(denySegments, s)
											} else {
												denySegments = append(denySegments, item)
											}
										}

										permStr = fmt.Sprintf("%s (%s %s)", permStr, common.GetLocale("deny"), strings.Join(denySegments, ", "))
									}
								}

								msgContent[key] = fmt.Sprintf("%v:%v", item["title"], permStr)
							}

							msgForEmail = append(msgForEmail, key)
							msgForLog = append(msgForLog, key)

						} else if item["type"] == "asAccessorPerms" {
							msgContent[key] = fmt.Sprintf("%v:%v", item["title"], "")
							msgForLog = append(msgForLog, key)
						} else if item["type"] == "datetime" {

							var timestamp int64 = -1

							if t, ok := item["value"].(int64); ok {
								timestamp = t
							} else if t, ok := item["value"].(string); ok {
								if t != "" {
									val, err := utils.ConvertTimeStringToMsTimestamp(t)
									if err == nil {
										timestamp = val
									}
								}
							}

							if timestamp == -1 {
								msgContent[key] = fmt.Sprintf("%v:%v", item["title"], common.GetLocale("forever"))
							} else {
								msgContent[key] = fmt.Sprintf("%v:%v", item["title"], time.Unix(timestamp/1e6, 0).Format("2006-01-02 15:04:05"))
							}

							msgForEmail = append(msgForEmail, key)
							msgForLog = append(msgForLog, key)
						} else {
							msgContent[key] = fmt.Sprintf("%v:%v", item["title"], item["value"])
							msgForEmail = append(msgForEmail, key)
							msgForLog = append(msgForLog, key)
						}

						contents = append(contents, item)
					} else {
						contents = append(contents, content)
					}
				}

				auditMsgData["content"] = contents

				if len(docids) > 0 {
					configs := make([]map[string]interface{}, 0)
					configs = append(configs, map[string]interface{}{
						"accessor": map[string]string{
							"id":   groupID,
							"type": "internal_group",
						},
						"allow": []string{
							"preview",
							"download",
							"display",
						},
					})

					for _, docid := range docids {
						code, cerr := m.docshare.SetDocPerm(ctx, docid, configs)
						if cerr != nil {
							if code == 409 {
								continue
							}
							log.Warnf("[HandleAuditorsMacth] SetDocPerm failed, err: %s", cerr.Error())
							continue
						}
					}
				}

				auditMsg := map[string]interface{}{
					"apply_id": task.ID,
					"data":     auditMsgData,
					"workflow": map[string]interface{}{
						"top_csf":       topCsfLevel,
						"msg_for_email": msgForEmail,
						"msg_for_log":   msgForLog,
						"content":       msgContent,
						"abstract_info": abstractInfo,
					},
				}

				auditMsgBytes, _ := json.Marshal(auditMsg)

				err := m.executeMethods.Publish(common.TopicWorkflowUpdate, auditMsgBytes)

				if err != nil {
					log.Warnf("[logic.UpdateTaskResults] Publish err, detail: %s", err.Error())
				}

				log.Infof("[logic.UpdateTaskResults] Publish update workflow message %v", auditMsg)

			} else {
				log.Warnf("[logic.UpdateTaskResults] inputContents is not primitive.A")
			}
		}

		return nil

	} else {
		return ierrors.NewIError(ierrors.Forbidden, "", nil)
	}
}

// ListOperators  list dag steps operators
func ListOperators(steps []entity.Step) []string {
	var operators []string
	for _, step := range steps {
		_step := step
		operators = append(operators, _step.Operator)
	}
	return operators
}

// ListActions  列举已注册的动作节点
func (m *mgnt) ListActions(ctx context.Context, userInfo *drivenadapters.UserInfo) ([]map[string]interface{}, error) {
	var actionsMap = mod.ActionMap
	var acts = make([]map[string]interface{}, 0)
	// Check if user is admin
	userDetail, _, err := m.getUserDetail(userInfo.UserID, &entity.AppInfo{})
	if err != nil {
		return nil, ierrors.NewIError(ierrors.UnAuthorization, "", map[string]interface{}{"info": err.Error()})
	}
	isAdmin := utils.IsAdminRole(userDetail.Roles)
	for _, val := range actionsMap {
		var act = map[string]interface{}{}
		act["name"] = val.Name()
		if val.Name() == common.AnyshareFileMatchContentOpt {
			var operations = []string{"KEYWORD"}
			act["config"] = map[string]interface{}{
				"options": operations,
			}
			if err := m.tika.CheckFastTextAnalysys(ctx); err != nil {
				act["config"] = map[string]interface{}{
					"options": operations,
				}
				acts = append(acts, act)
				continue
			}

			templateRes, err := m.tika.GetPrivacyTemplate(ctx)
			if err != nil {
				act["config"] = map[string]interface{}{
					"options": operations,
				}
				acts = append(acts, act)
				continue
			}

			operations = append(operations, templateRes...)
			act["config"] = map[string]interface{}{
				"options": operations,
			}
		} else if val.Name() == common.AnyshareFileOCROpt {
			var enable = m.config.T4th.Enable
			config := map[string]interface{}{
				"type":   m.config.T4th.Type,
				"enable": enable,
			}
			if enable {
				if isAdmin {
					config["enable"] = true
				} else {
					res, err := m.appstore.GetWhiteListStatus(ctx, "action_ocr", userInfo.TokenID)
					if err != nil {
						config["enable"] = false
					} else {
						config["enable"] = res["enable"].(bool)
					}
				}
			}

			act["config"] = config
		} else if val.Name() == common.InternalToolPy3Opt {
			act["config"] = map[string]interface{}{
				"enable": true,
			}
			// if isAdmin {
			// 	act["config"] = map[string]interface{}{
			// 		"enable": true,
			// 	}
			// } else {
			// 	res, err := m.appstore.GetWhiteListStatus(ctx, "action_python", userInfo.TokenID)
			// 	if err != nil {
			// 		act["config"] = map[string]interface{}{
			// 			"enable": false,
			// 		}
			// 	} else {
			// 		act["config"] = map[string]interface{}{
			// 			"enable": res["enable"],
			// 		}
			// 	}
			// }
		} else if val.Name() == common.AudioTransfer {
			err := m.dependency.CheckSpeechModel(ctx)
			if err != nil {
				act["config"] = map[string]interface{}{
					"enable": false,
				}
			} else {
				act["config"] = map[string]interface{}{
					"enable": true,
				}
			}
		} else if val.Name() == common.AnyshareDocLibQuotaScaleOpt {
			isAdmin, err := m.admin.CheckAdminExistByUSerID(ctx, userInfo.UserID)
			if err != nil || !isAdmin {
				act["config"] = map[string]interface{}{
					"enable": false,
				}
			}
			if isAdmin {
				act["config"] = map[string]interface{}{
					"enable": true,
				}
			}
		}
		acts = append(acts, act)
	}
	return acts, nil
}

func (m *mgnt) renderParamsV2(taskIns *entity.TaskInstance) error {
	vmIns := vm.NewVM()
	vmIns.AddGlobals(mod.NewGlobals(taskIns.RelatedDagInstance))

	g := vm.NewGenerator(vmIns)

	rawParams := make(map[string]any)
	params := taskIns.GetParams()

	switch taskIns.ActionName {
	case common.InternalAssignOpt:
		rawParams["target"] = params["target"]
		delete(params, "target")
	case common.OpJsonTemplate:
		if template, ok := params["template"]; ok {
			rawParams["template"] = template
			delete(params, "template")
		}
	case common.InternalToolPy3Opt:
		if code, ok := params["code"]; ok {
			rawParams["code"] = code
			delete(params, "code")
		}
	}

	err := g.GenerateValue(params)

	if err != nil {
		return err
	}

	vmIns.LoadInstructions(g.Instructions)

	// 使用ShareData的GetAll方法安全地获取数据副本作为VM的Env
	env := make(map[string]interface{})
	if taskIns.RelatedDagInstance.ShareData != nil {
		env = taskIns.RelatedDagInstance.ShareData.GetAll()
	}

	vmIns.Env = env
	vmIns.Run()

	_, ret, err := vmIns.Result()

	if err != nil {
		return err
	}

	if resultMap, ok := ret.(map[string]interface{}); ok {
		if len(rawParams) > 0 {
			for k, v := range rawParams {
				resultMap[k] = v
			}
		}
		taskIns.SetParams(resultMap)
	}

	return nil
}

func (m *mgnt) renderParams(taskIns *entity.TaskInstance) error {
	data := map[string]interface{}{}

	dagInstance := taskIns.RelatedDagInstance
	if dagInstance != nil {
		data["vars"] = dagInstance.Vars
		if dagInstance.ShareData != nil {
			data["shareData"] = dagInstance.ShareData.GetAll()
		}
	}

	params := taskIns.GetParams()
	err := value.MapValue(params).Walk(func(walkContext *value.WalkContext, v interface{}) error {

		// 赋值操作不解析 target 的值
		if taskIns.ActionName == common.InternalAssignOpt && walkContext.Path() == "target" {
			return nil
		}

		if mk, ok := v.(string); ok {
			if strings.Contains(mk, "{{__") && strings.Contains(mk, "}}") {
				// n := strings.Replace(mk, "{{", "{{.shareData.", 1)
				n := strings.ReplaceAll(mk, "{{", "{{.shareData.")
				result, err := m.paramRender.Render(n, data)
				if err != nil {
					return err
				}
				walkContext.Setter(result)
			}
		}
		return nil
	})
	if err != nil {
		log.Warnf("WalkString failed: %v", err)
		return err
	}
	taskIns.SetParams(params)
	return nil
}

// isAccessible 是否具有工作流的访问权限
func (m *mgnt) isAccessible(dagAccessors *[]entity.Accessor, userAccessors []string) bool {
	for _, dagAccessor := range *dagAccessors {
		if utils.IsContain(dagAccessor.ID, userAccessors) {
			return true
		}
	}
	return false
}

// getUserDetail 获取用户信息
func (m *mgnt) getUserDetail(userID string, appInfo *entity.AppInfo) (*drivenadapters.UserInfo, *entity.Token, error) {
	var (
		tokenInfo *entity.Token
		userType  string
	)
	tokenMgnt := mod.NewTokenMgnt(userID)
	accessorID := userID

	if appInfo.Enable {
		accessorID = m.config.OAuth.ClientID
		userType = common.APP.ToString()
	} else {
		isApp, err := m.usermgnt.IsApp(accessorID)
		if err != nil {
			return nil, nil, err
		}

		userType = utils.IfNot(isApp, common.APP.ToString(), common.User.ToString())
	}

	tokenInfo, err := tokenMgnt.GetUserToken("", accessorID)
	if err != nil {
		return nil, nil, err
	}

	userDetail, err := m.usermgnt.GetUserInfoByType(userID, userType)
	if err != nil {
		log.Warnf("[logic.getUserDetail] GetUserInfoByType err, deail: %s", err.Error())
		return nil, nil, err
	}
	userDetail.AccountType = userType

	return &userDetail, tokenInfo, nil
}

// getDataSourceType
func (m *mgnt) getDataSourceType(dataSource *entity.DataSource) string {
	if dataSource == nil {
		return ""
	}
	switch dataSource.Operator {
	case common.AnyshareDataSpecifyFiles, common.AnyshareDataSpecifyFolders, common.AnyshareDataListFolders, common.AnyshareDataListFiles:
		return "doc"
	case common.AnyshareDataDepartment:
		return "dept"
	case common.AnyshareDataUser:
		return "user"
	case common.AnyshareDataTagTree:
		return "tagtree"
	case common.S3DataListObjects:
		return "s3"
	}
	return ""
}

type DataSourceItem struct {
	ID       string
	Keywords []string
	Data     map[string]interface{}
}

// createDagInstanceFromDataSource
func (m *mgnt) createDagInstanceFromDataSource(ctx context.Context, triggerType entity.Trigger, runVar map[string]string, datas []*DataSourceItem, dag *entity.Dag) error {
	dag.SetPushMessage(m.executeMethods.Publish)
	dagInstances := make([]*entity.DagInstance, 0)
	for _, d := range datas {
		// Clone runVar to avoid polluting shared map across iterations or affecting caller
		currentRunVar := make(map[string]string)
		for k, v := range runVar {
			currentRunVar[k] = v
		}
		currentRunVar["docid"] = d.ID
		if d.Data != nil {
			for k, v := range d.Data {
				currentRunVar[k] = fmt.Sprintf("%v", v)
			}
		}

		dagIns, dagErr := dag.Run(ctx, triggerType, currentRunVar, entity.WithKeyWords(d.Keywords...))
		if dagErr != nil {
			return ierrors.NewIError(ierrors.Forbidden, ierrors.DagStatusNotNormal, map[string]interface{}{"id:": dag.ID, "status": dag.Status})
		}
		dagInstances = append(dagInstances, dagIns)
	}
	// TODO 考虑分批创建

	_, err := m.mongo.BatchCreateDagIns(ctx, dagInstances)
	if err != nil {
		log.Warnf("[logic.createDagInstanceFromDataSource] BatchCreateDagIns err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", err.Error())
	}

	return nil
}

// resetPermApplyStepAppPwd 回填应用账号信息
func resetPermApplyStepAppPwd(appCountInfoMap map[string]string, oldSteps, newSteps []entity.Step) {
	for _, step := range oldSteps {
		if step.Operator == common.AnyshareFileSetPermOpt {
			appid, ok := step.Parameters["appid"]
			if ok && appid != "" {
				apppwd := step.Parameters["apppwd"]
				appCountInfoMap[fmt.Sprintf("%v", appid)] = fmt.Sprintf("%v", apppwd)
			}
		}
	}
	for _, step := range newSteps {
		if step.Operator == common.AnyshareFileSetPermOpt {
			appid, ok := step.Parameters["appid"]
			if ok && appid != "" {
				oldAppPwd, aok := appCountInfoMap[fmt.Sprintf("%v", appid)]
				newAppPwd := step.Parameters["apppwd"]
				if aok && (newAppPwd == "" || newAppPwd == nil) {
					step.Parameters["apppwd"] = oldAppPwd
				} else if aok && newAppPwd != "" && oldAppPwd != newAppPwd {
					step.Parameters["apppwd"] = newAppPwd
				}
			}
		}
	}
}

type RunWithDocParams struct {
	DocID string                 `json:"docid"`
	Data  map[string]interface{} `json:"data"`
}

func (m *mgnt) RunInstanceWithDoc(ctx context.Context, id string, params RunWithDocParams, userInfo *drivenadapters.UserInfo) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	dag, err := m.mongo.GetDag(ctx, id)

	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": id})
		}
		log.Warnf("[logic.RunInstanceWithDoc] GetDagByFields err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	accessors, gerr := m.usermgnt.GetUserAccessorIDs(userInfo.UserID)
	if gerr != nil {
		log.Warnf("[logic.RunInstanceWithDoc] GetUserAccessorIDs err, detail: %s", gerr.Error())
		return gerr
	}

	if !m.isAccessible(&dag.Accessors, accessors) {

		// 创建者不在accessors列表不允许运行
		if userInfo.UserID == dag.UserID {
			return ierrors.NewIError(ierrors.Forbidden, "", map[string]string{"dagId": id})
		}

		return ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": id})
	}

	triggerType := m.getTriggerType(dag.Steps[0].Operator)

	if triggerType != entity.TriggerDocument {
		err := ierrors.NewIError(ierrors.Forbidden, ierrors.ErrorIncorretTrigger, map[string]interface{}{
			"trigger": fmt.Sprintf("%s trigger type is not allowed to run with document", triggerType),
		})
		return err
	}

	inherit, ok := dag.Steps[0].Parameters["inherit"].(bool)
	if !ok {
		inherit = false
	}

	var triggerDirs []string

	if traiggerDir, ok := dag.Steps[0].Parameters["docid"]; ok {
		triggerDirs = append(triggerDirs, traiggerDir.(string))
	}

	if traiggerDir, ok := dag.Steps[0].Parameters["docids"]; ok {
		if docIDs, ok := normalizeutil.AsSlice(traiggerDir); ok {
			for _, docID := range docIDs {
				if docIDStr, ok := docID.(string); ok {
					triggerDirs = append(triggerDirs, docIDStr)
				}
			}
		}
	}

	var runContinue bool
	for _, triggerDir := range triggerDirs {
		if inherit && strings.HasPrefix(params.DocID, triggerDir) ||
			!inherit && utils.ComputeLevelDifference(triggerDir, params.DocID) == 1 {
			runContinue = true
			break
		}
	}

	if !runContinue {
		return ierrors.NewIError(ierrors.TaskSourceInvalid, "", map[string]string{"dagId": id, "docid": params.DocID})
	}

	userDetail, err := m.usermgnt.GetUserInfoByType(userInfo.UserID, userInfo.AccountType)
	if err != nil {
		log.Warnf("[logic.RunInstanceWithDoc] GetUserInfoByType err, deail: %s", err.Error())
		return err
	}

	metadata, err := m.efast.GetDocMsg(ctx, params.DocID)
	sourceType := "file"
	if metadata.Size == -1 {
		sourceType = "folder"
	}

	source := map[string]interface{}{
		"type": sourceType,
		"id":   metadata.DocID,
		"name": metadata.Name,
		"rev":  metadata.Rev,
		"size": metadata.Size,
		"path": metadata.Path,
	}

	bytes, _ := json.Marshal(source)

	runVar := map[string]string{
		"source":        string(bytes),
		"userid":        userInfo.UserID,
		"operator_id":   userInfo.UserID,
		"operator_name": userDetail.UserName,
		"operator_type": userInfo.AccountType,
	}

	if fields, ok := normalizeutil.AsSlice(dag.Steps[0].Parameters["fields"]); ok {
		err = ParseFields(ctx, fields, params.Data, runVar, ErrTypeV1).BuildError()
		if err != nil {
			log.Warnf("[logic.RunInstanceWithDoc] ParseFields err, deail: %s", err.Error())
			return err
		}
	}

	dag.SetPushMessage(m.executeMethods.Publish)
	dagIns, dagErr := dag.Run(ctx, triggerType, runVar, entity.WithKeyWords(metadata.Name, userDetail.UserName))
	if dagErr != nil {
		return ierrors.NewIError(ierrors.Forbidden, ierrors.DagStatusNotNormal, map[string]interface{}{"id:": dag.ID, "status": dag.Status})
	}

	_, err = m.mongo.CreateDagIns(ctx, dagIns)
	if err != nil {
		log.Warnf("[logic.RunInstance] BatchCreateDagIns err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	go func() {
		detail, extMsg := common.GetLogBody(common.TriggerTaskManually, []interface{}{dag.Name},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: userInfo,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: id,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  userInfo,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagIns.ID,
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     dag.ID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}()
	return nil
}

func ParseFields(ctx context.Context, fields []interface{}, reqData map[string]interface{}, runVar map[string]string, errType string) *ValidateError {
	vErr := &ValidateError{
		Ctx:             ctx,
		ErrType:         errType,
		PublicErrorType: libErrors.PublicErrorType,
		MainCode:        ierrors.InvalidParameter,
		MainCodeV2:      libErrors.PErrorBadRequest,
		DescriptionKey:  libErrors.PErrorBadRequest,
	}
	deps := make(map[string]bool)

	for _, f := range fields {
		field, ok := f.(map[string]interface{})
		if !ok || field["type"] != "radio" {
			continue
		}

		data, ok := normalizeutil.AsSlice(field["data"])
		if !ok {
			continue
		}

		for _, item := range data {
			radioOption, isMap := item.(map[string]interface{})
			if !isMap {
				continue
			}

			related, hasRelated := normalizeutil.AsSlice(radioOption["related"])
			if !hasRelated {
				continue
			}

			for _, fieldKey := range related {
				if key, ok := fieldKey.(string); ok {
					deps[key] = deps[key] || radioOption["value"] == reqData[field["key"].(string)]
				}
			}
		}
	}

	for _, f := range fields {
		if field, ok := f.(map[string]interface{}); ok {
			key := field["key"].(string)
			typ := field["type"].(string)

			depIsValid, hasDep := deps[key]

			required, ok := field["required"].(bool)
			required = ok && required && (!hasDep || depIsValid)

			if reqData[key] == nil {

				if required {
					vErr.Detail = map[string]interface{}{"values": fmt.Sprintf("field: %s is required", key)}
					return vErr
				}

				continue
			}

			switch typ {
			case "string", "long_string", "radio":
				if val, ok := reqData[key].(string); ok {
					runVar[key] = val
				} else {
					vErr.Detail = map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)}
					return vErr
				}

			case "number":
				if val, ok := reqData[key].(int64); ok {
					runVar[key] = fmt.Sprintf("%v", val)
				} else if val, ok := reqData[key].(float64); ok {
					runVar[key] = fmt.Sprintf("%v", val)
				} else {
					vErr.Detail = map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)}
					return vErr
				}

			case "asFile", "asFolder", "asDoc":
				if val, ok := reqData[key].(string); ok && utils.IsGNS(val) {
					runVar[key] = val
				} else {
					vErr.Detail = map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)}
					return vErr
				}

			case "asTags":
				if val, ok := reqData[key].([]interface{}); ok {
					var tags = make([]string, 0)
					for _, v := range val {
						tags = append(tags, fmt.Sprintf("%v", v))
					}
					bytes, _ := json.Marshal(tags)
					runVar[key] = string(bytes)
				} else {
					vErr.Detail = map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)}
					return vErr
				}

			case "asLevel":
				if val, ok := reqData[key].(int64); ok {
					runVar[key] = fmt.Sprintf("%v", val)
				} else if val, ok := reqData[key].(map[string]interface{}); ok {
					bytes, _ := json.Marshal(val)
					err := common.JSONSchemaValid(bytes, "values/as-level-info.json")

					if err != nil {
						vErr.Detail = map[string]interface{}{"params": err.Error()}
						return vErr
					}

					runVar[key] = string(bytes)
				} else {
					vErr.Detail = map[string]interface{}{"values": fmt.Sprintf("invalid: %s", key)}
					return vErr
				}

			case "asMetadata":
				{
					bytes, _ := json.Marshal(reqData[key])
					err := common.JSONSchemaValid(bytes, "values/as-metadata.json")

					if err != nil {
						vErr.Detail = map[string]interface{}{"params": err.Error()}
						return vErr
					}

					runVar[key] = string(bytes)
				}

			case "asPerm":
				{
					bytes, _ := json.Marshal(reqData[key])
					err := common.JSONSchemaValid(bytes, "values/as-perm.json")

					runVar[key] = string(bytes)
					if err != nil {
						vErr.Detail = map[string]interface{}{"params": err.Error()}
						return vErr
					}
				}
			case "asAccessorPerms":
				{
					bytes, _ := json.Marshal(reqData[key])
					err := common.JSONSchemaValid(bytes, "values/as-accessor-perms.json")

					runVar[key] = string(bytes)
					if err != nil {
						vErr.Detail = map[string]interface{}{"params": err.Error()}
						return vErr
					}
				}

			default:
				{
					switch v := reqData[key].(type) {
					case string, int, float64:
						runVar[key] = fmt.Sprintf("%v", v)
					case map[string]interface{}:
						bytes, _ := json.Marshal(v)
						runVar[key] = string(bytes)
					case []interface{}:
						bytes, _ := json.Marshal(v)
						runVar[key] = string(bytes)
					default:
						runVar[key] = fmt.Sprintf("%v", v)
					}
				}
			}
		}
	}
	return nil
}

// ListModelBindDags 列举与模型绑定的dag信息
func (m *mgnt) ListModelBindDags(ctx context.Context, id, userID string) ([]*DagSimpleInfo, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var dagsInfo = make([]*DagSimpleInfo, 0)
	listDagInput := &mod.ListDagInput{
		Limit:  -1,
		Offset: 50,
		Order:  -1,
		UserID: userID,
	}

	dags, err := m.mongo.ListDag(ctx, listDagInput)
	if err != nil {
		log.Warnf("[logic.ListModelBindDags] ListDag err, detail: %s", err.Error())
		return dagsInfo, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	var modelBindDagIDs []string
	for _, dag := range dags {
		if utils.IsContain(common.DocInfoEntityExtract, ListOperators(dag.Steps)) {
			modelBindDagIDs = append(modelBindDagIDs, dag.ID)
		}
	}

	if len(modelBindDagIDs) == 0 {
		return dagsInfo, nil
	}

	var fillter = bson.M{
		"_id":    bson.M{"$in": modelBindDagIDs},
		"userid": bson.M{"$eq": userID},
	}
	modelBindDags, err := m.mongo.ListDagByFields(ctx, fillter, options.FindOptions{})
	if err != nil {
		log.Warnf("[logic.ListModelBindDags] ListDagByFields err, query: %v, detail: %s", fillter, err.Error())
		return dagsInfo, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	var accessorIDs = make(map[string]string)

	for _, dag := range dags {
		accessorIDs[dag.UserID] = common.User.ToString()
	}

	accessors, _ := m.usermgnt.GetNameByAccessorIDs(accessorIDs)

	for _, dag := range modelBindDags {
		for _, step := range dag.Steps {
			if step.Parameters == nil {
				continue
			}
			modelID, ok := step.Parameters["modelid"]
			if !ok || fmt.Sprintf("%v", modelID) != id {
				continue
			}
			dagsInfo = append(dagsInfo, &DagSimpleInfo{
				ID:        dag.ID,
				Title:     dag.Name,
				CreatedAt: dag.CreatedAt,
				UpdatedAt: dag.UpdatedAt,
				Actions:   ListOperators(dag.Steps),
				Status:    fmt.Sprintf("%v", dag.Status),
				UserID:    dag.UserID,
				Creator:   accessors[dag.UserID],
			})
		}
	}

	return dagsInfo, nil
}

// GetDagTriggerConfig 获取触发器配置信息
func (m *mgnt) GetDagTriggerConfig(ctx context.Context, taskInsID, typeBy string, userInfo *drivenadapters.UserInfo) (TriggerConfig, error) {
	var (
		err           error
		triggerConfig TriggerConfig
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	taskIns, err := m.mongo.GetTaskIns(ctx, taskInsID)
	if err != nil {
		log.Warnf("[logic.GetDagTriggerConfig] GetTaskIns err, detail: %s", err.Error())
		if ierrors.IsNotFoundErr(err) {
			return triggerConfig, ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"taskId": taskInsID})
		}
		return triggerConfig, err
	}

	// 控制台的工作流使用管理员角色创建，如果普通用户调用此接口会查询不到
	queryBy := map[string]interface{}{"_id": taskIns.DagInsID}
	if typeBy == common.CreateFlowByClient {
		queryBy["userid"] = userInfo.UserID
	}

	dagIns, err := m.mongo.GetDagInstanceByFields(ctx, queryBy)
	if err != nil {
		log.Warnf("[logic.GetDagTriggerConfig] GetDagInstance err, detail: %s", err.Error())
		if ierrors.IsNotFoundErr(err) {
			return triggerConfig, ierrors.NewIError(ierrors.DagInsNotFound, "", map[string]string{"dagInsId": taskIns.DagInsID})
		}
		return triggerConfig, err
	}

	// 判断流程是否为安全策略流程，非安全策略流程则抛错
	if dagIns.Trigger != common.DagTypeSecurityPolicy && typeBy == common.CreateFlowByConsole {
		return triggerConfig, ierrors.NewIError(ierrors.DagInsNotFound, "", map[string]string{"dagInsId": taskIns.DagInsID})
	}

	dag, err := m.mongo.GetDag(ctx, dagIns.DagID)
	if err != nil {
		log.Warnf("[logic.GetDagTriggerConfig] GetDag err, detail: %s", err.Error())
		if ierrors.IsNotFoundErr(err) {
			return triggerConfig, ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": dagIns.DagID})
		}
		return triggerConfig, err
	}

	step := dag.Steps[0]
	triggerConfig.Operator = step.Operator
	if step.Parameters != nil {
		triggerConfig.Params = step.Parameters
	}
	triggerConfig.ID = dag.ID

	// 获取数据源配置信息
	tasks, err := m.mongo.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
		DagInsID: taskIns.DagInsID,
	})
	if err != nil {
		log.Warnf("[logic.GetDagTriggerConfig] ListTaskInstance err, detail: %s", err.Error())
		return triggerConfig, err
	}

	for _, task := range tasks {
		if !strings.HasPrefix(task.ActionName, "@trigger") {
			continue
		}

		resultsMap := normalizeutil.PrimitiveToMap(task.Results)
		if len(resultsMap) != 0 {
			triggerConfig.Result = resultsMap
		}

		break
	}
	return triggerConfig, nil
}

func (m *mgnt) LogDagInsResult(ctx context.Context, dagIns *entity.DagInstance) {
	dag, err := m.mongo.GetDagWithOptionalVersion(context.Background(), dagIns.DagID, dagIns.VersionID)
	if err != nil {
		log.Warnf("get dag[%s] failed: %s", dagIns.DagID, err)
		return
	}

	var (
		detail string
		extMsg string
	)

	if dagIns.DagType == common.DagTypeSecurityPolicy {

		bodyType := "RunSecurityPolicyFlowFailed"
		if dagIns.Status == entity.DagInstanceStatusSuccess {
			bodyType = "RunSecurityPolicyFlowSuccess"
		}
		detail, extMsg = common.GetLogBody(bodyType, []interface{}{dag.ID},
			[]interface{}{})

	} else {
		bodyType := common.CompleteTaskWithFailed
		if dagIns.Status == entity.DagInstanceStatusSuccess {
			bodyType = common.CompleteTaskWithSuccess
		}
		detail, extMsg = common.GetLogBody(bodyType, []interface{}{dag.Name},
			[]interface{}{})
	}

	object := map[string]interface{}{
		"type":          dag.Trigger,
		"id":            dagIns.ID,
		"dagId":         dagIns.DagID,
		"name":          dag.Name,
		"priority":      dagIns.Priority,
		"status":        dagIns.Status,
		"biz_domain_id": utils.IfNot(dag.BizDomainID == "", common.BizDomainDefaultID, dag.BizDomainID),
	}

	if len(dag.Type) != 0 {
		object["dagType"] = dag.Type
	} else {
		object["dagType"] = common.DagTypeDefault
	}

	if dagIns.EndedAt < dagIns.CreatedAt {
		endedAt := time.Now().Unix()
		object["duration"] = endedAt - dagIns.CreatedAt
	} else {
		object["duration"] = dagIns.EndedAt - dagIns.CreatedAt
	}

	varsGetter := dagIns.VarsGetter()
	userID, _ := varsGetter("operator_id")
	AccountType, _ := varsGetter("operator_type")

	var userInfo drivenadapters.UserInfo
	userInfo, err0 := drivenadapters.NewUserManagement().GetUserInfoByType(fmt.Sprintf("%v", userID), fmt.Sprintf("%v", AccountType))
	if err0 != nil {
		log.Warnf("[logic.LogDagInsResult] GetUserInfoByType failed: %s", err0.Error())
		userName, _ := varsGetter("operator_name")
		userType, _ := varsGetter("operator_type")
		userInfo = drivenadapters.UserInfo{
			UserID:      fmt.Sprintf("%v", userID),
			UserName:    fmt.Sprintf("%v", userName),
			Type:        fmt.Sprintf("%v", userType),
			AccountType: fmt.Sprintf("%v", AccountType),
		}
	}
	userInfo.VisitorType = common.InternalServiceUserType

	m.logger.LogO11y(&drivenadapters.BuildARLogParams{
		Operation:   common.ArLogEndDagIns,
		Description: detail,
		UserInfo:    &userInfo,
		Object:      object,
	}, &drivenadapters.O11yLogWriter{Logger: traceLog.NewFlowO11yLogger()})

	traceLog.WithContext(ctx).Infof("detail: %s, extMsg: %s", detail, extMsg)
	write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
	// 原AS审计日志发送逻辑
	// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
	// 	UserInfo: &userInfo,
	// 	Msg:      detail,
	// 	ExtMsg:   extMsg,
	// 	OutBizID: dagIns.ID,
	// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_WARN,
	// }, write)

	m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
		UserInfo:  &userInfo,
		Msg:       detail,
		ExtMsg:    extMsg,
		OutBizID:  dagIns.ID,
		Operation: drivenadapters.ExecuteOperation,
		ObjID:     dag.ID,
		ObjName:   dag.Name,
		LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
	}, write)
}

func (m *mgnt) CallAgent(ctx context.Context, name string, inputs map[string]interface{}, options *drivenadapters.CallAgentOptions, token string) (res *drivenadapters.CallAgentRes, ch chan *drivenadapters.CallAgentRes, err error) {
	res, ch, err = m.ad.CallAgent(ctx, name, inputs, options, token)
	return
}

const (
	envDebugKey = "DEBUG"
)

// 添加一个工具函数来检查是否开启debug模式
func isDebugEnabled() bool {
	return os.Getenv(envDebugKey) == "true"
}

func (m *mgnt) GetAgents(ctx context.Context) (res []*rds.AgentModel, err error) {
	res, err = m.agent.GetAgents(ctx)
	return
}

// Extract shared logic into a helper method
func (m *mgnt) runDagInstance(ctx context.Context, dag *entity.Dag, triggerType entity.Trigger, runVar map[string]string, userDetail *drivenadapters.UserInfo) error {
	dagIns, dagErr := dag.Run(ctx, triggerType, runVar)
	if dagErr != nil {
		return ierrors.NewIError(ierrors.Forbidden, ierrors.DagStatusNotNormal, map[string]interface{}{"id:": dag.ID, "status": dag.Status})
	}
	_, err := m.mongo.CreateDagIns(ctx, dagIns)
	if err != nil {
		log.Warnf("[logic.RunInstance] BatchCreateDagIns err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	go func() {
		detail, extMsg := common.GetLogBody(common.TriggerTaskManually, []interface{}{dag.Name},
			[]interface{}{})
		log.Infof("detail: %s, extMsg: %s", detail, extMsg)
		write := &drivenadapters.JSONLogWriter{SendFunc: m.executeMethods.Publish}
		// 原AS审计日志发送逻辑
		// m.logger.Log(drivenadapters.LogTypeASAuditLog, &drivenadapters.BuildAuditLogParams{
		// 	UserInfo: userDetail,
		// 	Msg:      detail,
		// 	ExtMsg:   extMsg,
		// 	OutBizID: dagIns.ID,
		// 	LogLevel: drivenadapters.NcTLogLevel_NCT_LL_INFO,
		// }, write)

		m.logger.Log(drivenadapters.LogTypeDIPFlowAduitLog, &drivenadapters.BuildDIPFlowAuditLog{
			UserInfo:  userDetail,
			Msg:       detail,
			ExtMsg:    extMsg,
			OutBizID:  dagIns.ID,
			Operation: drivenadapters.ExecuteOperation,
			ObjID:     dag.ID,
			ObjName:   dag.Name,
			LogLevel:  drivenadapters.NcTLogLevel_NCT_LL_INFO_Str,
		}, write)
	}()
	return nil
}

func (m *mgnt) RetryDagInstance(ctx context.Context, dagInsID string, userInfo *drivenadapters.UserInfo) error {
	dagIns, err := m.mongo.GetDagInstance(ctx, dagInsID)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return ierrors.NewIError(ierrors.TaskNotFound, "", nil)
		}
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow:      {perm.ManualExecOperation},
			common.DagTypeComboOperator: {perm.OpExecuteOperation},
			common.DagTypeDefault:       {perm.OldOnlyAdminOperation},
		},
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dagIns.DagID, userInfo, opMap)
	if err != nil {
		return err
	}

	if dagIns.Status != entity.DagInstanceStatusFailed && dagIns.Status != entity.DagInstanceStatusCancled {
		return ierrors.NewIError(ierrors.ForbiddenRetryableDagIns, "", nil)
	}

	tasks, err := m.mongo.ListTaskInstance(ctx, &mod.ListTaskInstanceInput{
		DagInsID: dagInsID,
	})
	if err != nil {
		log.Warnf("[logic.RetryDagInstance] ListTaskInstance err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	var toUpdateTaskInsIDs []string
	for _, task := range tasks {
		if task.Status != entity.TaskInstanceStatusFailed && task.Status != entity.TaskInstanceStatusCanceled {
			continue
		}
		toUpdateTaskInsIDs = append(toUpdateTaskInsIDs, task.ID)
	}

	err = m.mongo.RetryDagIns(ctx, dagInsID, toUpdateTaskInsIDs)
	if err != nil {
		log.Warnf("[logic.RetryDagInstance] RetryDagIns err, deail: %s", err.Error())
		return ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	return nil
}

// Update both RunInstance and RunCronInstance to use the helper

func (m *mgnt) RunFormInstanceV2(ctx context.Context, id string, formData map[string]interface{}, successCallback string, errorCallback string, userInfo *drivenadapters.UserInfo) (dagIns *entity.DagInstance, vmIns *mod.VMExt, err error) {
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	log := traceLog.WithContext(ctx)

	dag, err := m.mongo.GetDag(ctx, id)
	if err != nil {
		if ierrors.IsNotFoundErr(err) {
			return nil, nil, ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": id})
		}
		log.Warnf("[logic.RunFormInstanceV2] GetDagByFields err, deail: %s", err.Error())
		return nil, nil, ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	dag.SetPushMessage(m.executeMethods.Publish)

	if userInfo != nil && dag.UserID != userInfo.UserID {
		accessors, gerr := m.usermgnt.GetUserAccessorIDs(userInfo.UserID)
		if gerr != nil {
			log.Warnf("[logic.RunFormInstanceV2] GetUserAccessorIDs err, detail: %s", gerr.Error())
			return nil, nil, gerr
		}
		if !m.isAccessible(&dag.Accessors, accessors) {
			return nil, nil, ierrors.NewIError(ierrors.TaskNotFound, "", map[string]string{"dagId": id})
		}
	}

	if !dag.Published && userInfo == nil {
		return nil, nil, ierrors.NewIError(ierrors.Forbidden, "", map[string]string{"dagId": id})
	}

	if userInfo != nil {
		userDetail, err := m.usermgnt.GetUserInfoByType(userInfo.UserID, userInfo.AccountType)
		if err != nil {
			log.Warnf("[logic.RunFormInstanceV2] GetUserInfoByType err, deail: %s", err.Error())
			return nil, nil, err
		}

		userInfo.UserName = userDetail.UserName
	}

	// Check if the dag is published
	if dag.Published {
		// Use existing userInfo if available, otherwise create a default anonymous user
		if userInfo == nil {
			userInfo = &drivenadapters.UserInfo{
				UserID:   "anonymous",      // or any default ID
				UserName: "Anonymous User", // or any default name
				Type:     "anonymous",      // or any default type
			}
		}
	}

	return m.runFormInstanceVM(ctx, dag, formData, successCallback, errorCallback, "", userInfo, dag.UserID)
}

// isEmptyObject 检查一个值是否为空对象
func isEmptyObject(val interface{}) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Map:
		return v.Len() == 0
	case reflect.Slice, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return isEmptyObject(v.Elem().Interface())
	default:
		// 其他类型不认为是空对象
		return false
	}
}

func (m *mgnt) BatchGetDag(ctx context.Context, dagIDs []string, fields string, userInfo *drivenadapters.UserInfo) ([]*DagInfoOption, error) {
	var (
		err      error
		dagInfos = []*DagInfoOption{}
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	if len(fields) == 0 {
		return dagInfos, ierrors.NewIError(ierrors.InvalidParameter, "", fmt.Errorf("query fields empty"))
	}

	fieldList, err := common.FieldQuery.Accept(&common.DagQueryVisitor{Fields: fields})
	if err != nil {
		return dagInfos, ierrors.NewIError(ierrors.InvalidParameter, "", err.Error())
	}

	dags, err := m.mongo.ListDag(ctx, &mod.ListDagInput{
		UserID:      userInfo.UserID,
		DagIDs:      dagIDs,
		SelectField: fieldList,
		Type:        "all",
	})
	if err != nil {
		log.Warnf("[logic.BatchGetDag] ListDag err, deail: %s", err.Error())
		return dagInfos, err
	}

	for _, dag := range dags {
		dio := &DagInfoOption{}
		dagByte, _ := json.Marshal(dag)
		_ = json.Unmarshal(dagByte, dio)
		// 去除无用字段
		dio.CreatedAt = dio.TmpCreatedAt
		dio.UpdatedAt = dio.TmpUpdatedAt
		if len(dio.UserID) != 0 {
			dio.Creator = &drivenadapters.UserAttribute{
				ID:   dio.UserID,
				Name: userInfo.UserName,
			}
		}
		dio.TmpCreatedAt, dio.TmpUpdatedAt = 0, 0
		dio.UserID = ""
		dagInfos = append(dagInfos, dio)
	}

	return dagInfos, nil
}

// ListDagV2 修改ResourceID格式为 dagID:type 用户后续指定类型查询获取对应类型总数
func (m *mgnt) ListDagV2(ctx context.Context, param QueryParams, userInfo *drivenadapters.UserInfo) ([]*DagSimpleInfo, int64, error) {
	var (
		err   error
		total int64
		dags  []*entity.Dag
		res   = []*DagSimpleInfo{}
	)
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	// Step1: check dag whether exist
	dags, total, err = ListDagWithFilters(ctx, param,
		WithBizDomainFilter(m.bizDomain, param.BizDomainID, "", param.Type, userInfo.TokenID),
		WithPermissionFilter(m.permPolicy, userInfo, param.Type),
		WithExistenceFilter())
	if err != nil {
		return res, total, err
	}
	var accessorIDs = make(map[string]string)

	for _, dag := range dags {
		accessorIDs[dag.UserID] = common.User.ToString()
	}

	accessors, _ := m.usermgnt.GetNameByAccessorIDs(accessorIDs)

	for _, dag := range dags {
		simpleDag := &DagSimpleInfo{
			ID:          dag.ID,
			Title:       dag.Name,
			Description: dag.Description,
			CreatedAt:   dag.CreatedAt,
			UpdatedAt:   dag.UpdatedAt,
			Actions:     ListOperators(dag.Steps),
			Status:      fmt.Sprintf("%v", dag.Status),
			UserID:      dag.UserID,
			Creator:     accessors[dag.UserID],
			Trigger:     dag.Trigger,
			Type:        dag.Type,
			VersionID:   dag.VersionID,
		}
		if dag.Type == "" {
			simpleDag.Type = common.DagTypeDefault
		}
		res = append(res, simpleDag)
	}

	return res, total, nil
}

func (m *mgnt) ListDagInstanceEvents(ctx context.Context, dagID, dagInsID string, offset, limit int, userInfo *drivenadapters.UserInfo) (
	logs []*entity.DagInstanceEvent, dagIns *entity.DagInstance, total int, next int, err error) {

	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	opMap := &perm.MapOperationProvider{
		OpMap: map[string][]string{
			common.DagTypeDataFlow:      {perm.RunStatisticsOperation},
			common.DagTypeComboOperator: {perm.ViewOperation},
			common.DagTypeDefault:       {perm.OldOnlyAdminOperation},
		},
	}

	_, err = m.permCheck.CheckDagAndPerm(ctx, dagID, userInfo, opMap)
	if err != nil {
		return nil, nil, 0, offset, err
	}

	err = m.isDagInstanceExist(ctx, map[string]interface{}{"_id": dagInsID, "dagId": dagID})
	if err != nil {
		return nil, nil, 0, offset, err
	}

	dagIns, err = m.mongo.GetDagInstanceByFields(ctx, map[string]interface{}{"_id": dagInsID, "dagId": dagID})

	if err != nil {
		log.Warnf("[logic.ListDagInstanceEvents] GetDagInstanceByFields err, detail: %s", err.Error())
		return nil, nil, 0, offset, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	opt := &rds.DagInstanceEventListOptions{
		DagInstanceID: dagInsID,
		Offset:        offset,
		Limit:         limit,
		Visibilities:  []rds.DagInstanceEventVisibility{rds.DagInstanceEventVisibilityPublic},
		Fields:        rds.DagInstanceEventFieldPublic,
	}

	total, err = m.eventRepository.ListCount(ctx, opt)

	if err != nil {
		log.Warnf("[logic.ListDagInstanceEvents] ListCount err, detail: %s", err.Error())
		return nil, nil, 0, offset, ierrors.NewIError(ierrors.InternalError, "", nil)
	}

	logs, err = dagIns.ListEvents(ctx, opt)

	if err != nil {
		log.Warnf("[logic.ListDagInstanceEvents] List err, detail: %s", err.Error())
		return nil, nil, 0, offset, ierrors.NewIError(ierrors.InternalError, "", nil)
	}
	next = offset + len(logs)
	return
}

// validSubFlowInSteps 校验子流程调用信息
func (m *mgnt) validSubFlowInSteps(ctx context.Context, dagID string, stepList []map[string]interface{}) ([]string, error) {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	log := traceLog.WithContext(ctx)

	var subDagIDs []string
	for _, v := range stepList {
		opType := v["operator"].(string)
		if opType == common.SubFlowCallWorkflowOpt || opType == common.SubFlowCallDataflowOpt {
			parameters := v["parameters"].(map[string]any)
			subDagIDs = append(subDagIDs, parameters["dag_id"].(string))
		}
	}

	if len(subDagIDs) == 0 {
		return subDagIDs, nil
	}

	dags, err := m.mongo.ListDag(ctx, &mod.ListDagInput{
		DagIDs: subDagIDs,
		Type:   "all",
	})
	if err != nil {
		log.Warnf("[logic.validSubFlowInSteps] ListDag err, detail: %s", err.Error())
		return subDagIDs, libErrors.NewPublicRestError(ctx, libErrors.PErrorInternalServerError, libErrors.PErrorInternalServerError, nil)
	}

	if len(subDagIDs) != len(dags) {
		existIDs := []string{}
		for _, dag := range dags {
			existIDs = append(existIDs, dag.ID)
		}
		_, del := utils.Arrcmp(subDagIDs, existIDs)
		return subDagIDs, libErrors.NewPublicRestError(ctx, libErrors.PErrorNotFound, libErrors.PErrorNotFound, map[string]any{"ids": del})
	}

	cycle, err := m.hasCycle(ctx, dagID, subDagIDs, []string{common.DagTypeDefault, common.DagTypeDataFlow})
	if err != nil {
		return subDagIDs, err
	}

	if cycle.Cycle {
		return subDagIDs, libErrors.NewPublicRestError(ctx, libErrors.PErrorBadRequest, libErrors.PErrorLoopDetected, cycle)
	}

	return subDagIDs, nil
}
