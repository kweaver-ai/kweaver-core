package common

import "time"

// 排序相关常量
const (
	ASC        = "asc"        // 升序
	DESC       = "desc"       // 降序
	Updated_At = "updated_at" // 更新于
	UpdatedAt  = "updatedAt"  // 更新于
	Created_At = "created_at" // 创建于
	CreatedAt  = "createdAt"  // 创建于
	Started_At = "started_at" // 开始于
	Ended_At   = "ended_at"   // 结束于
	EndedAt    = "endedAt"    // 结束于
	Name       = "name"       // 排序按名称
)

// 任务状态常量
const (
	NormalStatus    = "normal"    // 任务状态启用
	SuccessStatus   = "success"   // 任务运行成功
	FailedStatus    = "failed"    // 任务运行失败
	CanceledStatus  = "canceled"  // 任务运行取消
	RunningStatus   = "running"   // 任务运行中
	ScheduledStatus = "scheduled" // 任务等待中
	StoppedStatus   = "stopped"   // 任务状态禁用
	UndoStatus      = "undo"      // 任务未运行
	BlockStatus     = "block"     // 任务阻塞中
)

// 时间相关常量
const (
	TimeFormat = "2006-01-02 15:04:05" // 时间戳转换格式
)

// 消息通道相关常量
const (
	ChannelMessage   = "automation"     // channel
	DagChannelPrefix = "automation.dag" // DAG通道前缀
)

// 任务操作类型常量
const (
	CreateTask              = "createTask"              // 创建任务
	UpdateTask              = "updateTask"              // 更新任务
	DeleteTask              = "deleteTask"              // 删除任务
	TriggerTaskManually     = "triggerTaskManually"     // 手动触发任务
	TriggerTaskCron         = "triggerTaskCron"         // 定时触发任务
	CompleteTaskWithSuccess = "completeTaskWithSuccess" // 完成任务-成功
	CompleteTaskWithFailed  = "completeTaskWithFailed"  // 完成任务-失败
	CancelRunningInstance   = "cancelRunningInstance"   // 取消执行实例
)

// 自定义节点操作常量
const (
	CreateCustomExecutor       = "createCustomExecutor"       // 创建自定义执行器
	UpdateCustomExecutor       = "updateCustomExecutor"       // 更新自定义执行器
	DeleteCustomExecutor       = "deleteCustomExecutor"       // 删除自定义执行器
	CreateCustomExecutorAction = "createCustomExecutorAction" // 创建自定义执行器动作
	UpdateCustomExecutorAction = "updateCustomExecutorAction" // 更新自定义执行器动作
	DeleteCustomExecutorAction = "deleteCustomExecutorAction" // 删除自定义执行器动作
)

// 审核相关常量
const (
	AuditType                 = "automation"       // 审核类型
	SecurityPolicyAuditPrefix = "security_policy_" // 安全策略审核前缀
	SecurityPolicyPerm        = "perm"             // 权限申请安全策略类型
	SecurityPolicyUpload      = "upload"           // 上传审核安全策略类型
	SecurityPolicyDelete      = "delete"           // 删除审核安全策略类型
)

// HTTP头部常量
const (
	Authorization   = "x-authorization" // Authorization 标识
	AnyshareAddress = "x-as-address"    // as 地址
)

// 系统用户常量
const (
	SystemSysAdmin       = "266c6a42-6131-4d62-8f39-853e7093701c" // admin管理员
	SystemAuditAdmin     = "94752844-BDD0-4B9E-8927-1CA8D427E699" // audit管理员
	SystemSecAdmin       = "4bb41612-a040-11e6-887d-005056920bea" // security管理员
	SystemOriginSysAdmin = "234562BE-88FF-4440-9BFF-447F139871A2" // 原有sys管理员ID
)

// 调度优先级常量
const (
	PriorityHighest = "highest" // 调度highest优先级
	PriorityHigh    = "high"    // 调度high优先级
	PriorityMedium  = "medium"  // 调度medium优先级
	PriorityLow     = "low"     // 调度low优先级
	PriorityLowest  = "lowest"  // 调度lowest优先级
)

// 工作流相关常量
const (
	WorkflowApprovalTaskIds = "__workflow_approval_task_ids" // 待审核任务id变量名称
	NotifyToExecutor        = "notifyToExecutor"             // 通知执行者
	StartRunDag             = "startRunDag"                  // 开始执行dag
)

// 模型训练状态常量
const (
	TrainStatusInit     = "init"     // 模型训练初始化状态
	TrainStatusFinished = "finished" // 模型训练训练完成状态
	TrainStatusFailed   = "failed"   // 模型训练失败状态
)

// AR日志类型常量
const (
	ArLogCreateDag   = "create" // ar日志类型 创建dag
	ArLogStartDagIns = "start"  // ar日志类型 运行dag
	ArLogEndDagIns   = "end"    // ar日志类型 dag运行结束
)

// 创建来源常量
const (
	CreateByLocal        = "local"    // 从本地导入标识
	CreateByTemplate     = "template" // 从流程模板新建标识
	CreateByLocalName    = "从本地导入"    // 从本地导入
	CreateByTemplateName = "从流程模板新建"  // 从流程模板新建
	CreateByDirectName   = "直接新建"     // 直接新建
	CreateFlowByClient   = "client"   // 从客户端创建工作流
	CreateFlowByConsole  = "console"  // 从控制台创建工作流
)

// 用户类型常量
const (
	AuthenticatedUserType   = "authenticated_user" // 实名用户
	AnonymousUserType       = "anonymous_user"     // 匿名用户
	InternalServiceUserType = "internal_service"   // 内部服务用户
)

// 文档库类型常量
const (
	KnowledgeDocLib  = "knowledge_doc_lib"  // 知识库
	DepartmentDocLib = "department_doc_lib" // 部门文档库
	CustomDocLib     = "custom_doc_lib"     // 自定义文档库
)

// 服务相关常量
const (
	ServiceName               = "content-automation"                // 服务名
	FlowServiceName           = "flow-automation"                   // 工作流服务名
	ErrCodeServiceName        = "FlowAutomation"                    // RestError 错误码第一段服务名
	CMS_CONFIG_SERVICE_ACCESS = "/conf/service-access/default.yaml" // cms配置路径
)

// 执行模式常量
const (
	ExecutionModeAsync = "async" // 异步
	ExecutionModeSync  = "sync"  // 同步
)

// 工作流类型常量
const (
	DagTypeSecurityPolicy = "security-policy"   // 安全策略类流程
	DagTypeDataFlow       = "data-flow"         // 数据流类型
	DagTypeDataFlowForBot = "data-flow-for-bot" // bot已绑定工作流
	DagTypeComboOperator  = "combo-operator"    // 组合算子
	DagTypeDefault        = "default"           // 默认流程类型
)

// 数据库类型常量
const (
	DBTYPEKDB = "KDB" // KDB数据库类型
)

// 数据流版本管理常量
const (
	DefaultDagVersion = "v0.0.0" // 数据流版本管理默认版本定义
)

// 权限校验资源类型常量
const (
	ReSourceTypeObservability = "observability" // 可观测性概览资源类型
)

// API路径常量
const (
	APIPREFIXV2        = "/api/automation/v2" // API v2前缀
	BizDomainDefaultID = "bd_public"          // 默认业务域ID
)

// 国际化资源路径常量
const (
	MultiResourcePath = "resource/locales" // 国际化资源路径
)

// 时间间隔和大小常量
const (
	DumpLogLockTime  = 30 * time.Second // 日志转存锁持有时间
	DefaultQuerySize = 10000            // 一次批量查询的大小
)

// ========== 枚举类型定义 ==========

// Order 排序类型
type Order string //nolint

// AccessorType 分组类型
type AccessorType string

const (
	Group      AccessorType = "group"      // 用户组
	User       AccessorType = "user"       // 普通用户
	Department AccessorType = "department" // 部门
	Contactor  AccessorType = "contactor"  // 联系人
	APP        AccessorType = "app"        // 应用账户
)

func (at AccessorType) ToString() string {
	return string(at)
}

// 发布状态枚举
const (
	Init = iota - 1
	UnPublish
	Publish
)

// 类型枚举
const (
	_ = iota
	UIEType
	TagRuleType
)
