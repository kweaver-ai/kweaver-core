package errors

import "net/http"

// ModuleName 模块名
var ModuleName = "ContentAutomation" // 模块名

// 主错误码
var (
	UnAuthorization      = "UnAuthorization"
	Forbidden            = "Forbidden"
	NoPermission         = "NoPermission"
	InvalidParameter     = "InvalidParameter"
	TaskNotFound         = "TaskNotFound"
	DagInsNotFound       = "DagInsNotFound"
	InternalError        = "InternalError"
	InvalidUser          = "InvalidUser"
	DuplicatedName       = "DuplicatedName"
	OperationDenied      = "OperationDenied"
	TaskSourceNotFound   = "TaskSourceNotFound"
	TaskSourceNoPerm     = "TaskSourceNotPerm"
	TaskSourceInvalid    = "TaskSourceInvalid"
	FileSizeExceed       = "FileSizeExceed"
	FileIsEmpty          = "FileIsEmpty"
	FileNotFound         = "FileNotFound"
	FileTypeNotSupported = "FileTypeNotSupported"
	NotContainPageData   = "NotContainPageData"
	FileContentUnknow    = "FileContentUnknow"

	// 自定义节点错误码
	ExecutorNotFound  = "ExecutorNotFound"
	ExecutorForbidden = "ExecutorForbidden"
	DuplicatedAdmin   = "DuplicatedAdmin"

	AgentNotFound          = "AgentNotFound"
	AlarmRuleNotFound      = "AlarmRuleNotFound"
	AlarmRuleAlreadyExists = "AlarmRuleAlreadyExists"
	UnAvailable            = "UnAvailable"

	ForbiddenRetryableDagIns = "ForbiddenRetryableDagIns"

	TaskAlreayInProgress = "TaskAlreadyInProgress"
	DataSourceIsEmpty    = "DataSourceIsEmpty"
)

var ErrorsHttpCode = map[string]int{
	ExecutorForbidden: http.StatusForbidden,
	ExecutorNotFound:  http.StatusNotFound,
	AgentNotFound:     http.StatusNotFound,
}

// 辅助错误码
var (
	ErrorDepencyService   = "ErrorDepencyService" // 依赖服务出错
	ErrorIncorretTrigger  = "ErrorIncorretTrigger"
	ErrorIncorretOperator = "ErrorIncorretOperator"
	DagInsNotRunning      = "DagInsNotRunning"
	DagStatusNotNormal    = "DagStatusNotNormal"
	NumberOfTasksLimited  = "NumberOfTasksLimited"
	ServiceDisabled       = "ServiceDisabled"
	ModelTrainFailed      = "ModelTrainFailed"
	UnSupportedFileType   = "UnSupportedFileType"
	TrainingInProgress    = "TrainingInProgress"
	NonProcessManager     = "NonProcessManager"
	UnSupportedTrigger    = "UnSupportedTrigger"
	WorkflowCycleDetected = "WorkflowCycleDetected"
)

// ErrorsMsg error msg map
var ErrorsMsg = map[string]map[string][]string{
	InvalidParameter: {
		Languages[0]: {"传入参数不正确", "请检查参数"},
		Languages[1]: {"傳入參數不正確", "請檢查參數"},
		Languages[2]: {"The passed in parameter is incorrect", "please check the parameter"},
	},
	UnAuthorization: {
		Languages[0]: {"登录或授权已过期", "请重新登录或授权"},
		Languages[1]: {"登入或授權已過期", "請重新登入或授權"},
		Languages[2]: {"Login or authorization expired", "Please login or authorization again"},
	},
	InternalError: {
		Languages[0]: {"请求因服务器内部错误导致异常", "请提交工单或联系技术支持工程师"},
		Languages[1]: {"請求因服務器內部錯誤導致异常", "請提交工單或聯系技術支援工程師"},
		Languages[2]: {"The request is abnormal due to an internal error of the server", "please submit the work order or contact the technical support engineer"},
	},
	TaskNotFound: {
		Languages[0]: {"任务不存在", "请确认任务id是否正确"},
		Languages[1]: {"任務不存在", "請確認任務id是否正確"},
		Languages[2]: {"Task does not exist", "Please confirm whether the task ID is correct"},
	},
	InvalidUser: {
		Languages[0]: {"请求中包含了不存在的用户导致失败", "请核对请求中的用户是否存在"},
		Languages[1]: {"請求中包含了不存在的用戶導致失敗", "請核對請求中的用戶是否存在"},
		Languages[2]: {"The request contains a non-existent user, resulting in failure", "please check whether the user in the request exists"},
	},
	Forbidden: {
		Languages[0]: {"操作被禁止，管理员已禁止此操作", "请确认是否允许执行此操作"},
		Languages[1]: {"操作被禁止，管理員已禁止此操作", "請確認是否允許執行此操作"},
		Languages[2]: {"The operation is prohibited, and the administrator has prohibited this operation", "Please confirm whether this operation is allowed"},
	},
	NoPermission: {
		Languages[0]: {"您没有权限执行此操作", "请确认是否拥有权限执行此操作"},
		Languages[1]: {"您沒有權限執行此操作", "請確認是否擁有許可權執行此操作"},
		Languages[2]: {"You do not have permission to perform this operation", "Please confirm whether you have permission to perform this operation"},
	},
	DuplicatedName: {
		Languages[0]: {"操作失败，已存在同名任务", "请更换名称重试"},
		Languages[1]: {"操作失敗，已存在同名任務", "請更換名稱重試"},
		Languages[2]: {"Operation failed, a task with the same name already exists", "Please change the name and try again"},
	},
	OperationDenied: {
		Languages[0]: {"操作失败，不允许执行该操作", "请确认操作对象是否有效"},
		Languages[1]: {"操作失敗，不允許執行該操作", "請確認操作對像是否有效"},
		Languages[2]: {"Operation failed, the operation is not allowed", "Please confirm whether the operation object is valid"},
	},
	DagInsNotFound: {
		Languages[0]: {"任务实例不存在", "请确认任务实例id是否正确"},
		Languages[1]: {"任務實例不存在", "請確認任務實例id是否正確"},
		Languages[2]: {"Dag instance does not exist", "Please confirm whether the task instance ID is correct"},
	},
	TaskSourceNotFound: {
		Languages[0]: {"任务执行目标已不存在", "请确认任务的执行目标"},
		Languages[1]: {"執行目標已不存在", "請確認任務的執行目標"},
		Languages[2]: {"Execution target no longer exists", "Please confirm the execution target of the task"},
	},
	TaskSourceNoPerm: {
		Languages[0]: {"对任务的执行目标没有权限", "请确认任务的执行目标"},
		Languages[1]: {"對任務的執行目標沒有權限", "請確認任務的執行目標"},
		Languages[2]: {"No permissions on the execution target of the task", "Please confirm the execution target of the task"},
	},
	TaskSourceInvalid: {
		Languages[0]: {"任务执行目标无法解析", "请确认任务的执行目标"},
		Languages[1]: {"任務執行目標無法解析", "請確認任務的執行目標"},
		Languages[2]: {"The task execution target could not be resolved", "Please confirm the execution target of the task"},
	},
	FileSizeExceed: {
		Languages[0]: {"目标文件大小超出最大限制", "请确认任务的执行目标"},
		Languages[1]: {"目標文件大小超出最大限制", "請確認任務的執行目標"},
		Languages[2]: {"Object file size exceeds maximum limit", "Please confirm the execution target of the task"},
	},
	FileIsEmpty: {
		Languages[0]: {"文件不能为空", "请确认任务的执行目标"},
		Languages[1]: {"文件不能为空", "請確認任務的執行目標"},
		Languages[2]: {"File is empty", "Please confirm the execution target of the task"},
	},
	FileNotFound: {
		Languages[0]: {"文件不存在", "请确认任务的执行目标"},
		Languages[1]: {"文件不存在", "請確認任務的執行目標"},
		Languages[2]: {"File not found", "Please confirm the execution target of the task"},
	},
	NotContainPageData: {
		Languages[0]: {"目标文件没有页数信息", "请确认任务的执行目标"},
		Languages[1]: {"目標文件沒有頁數信息", "請確認任務的執行目標"},
		Languages[2]: {"The target file has no page number information", "Please confirm the execution target of the task"},
	},
	FileTypeNotSupported: {
		Languages[0]: {"目标格式不支持", "请确认任务的执行目标"},
		Languages[1]: {"目標格式不支援", "請確認任務的執行目標"},
		Languages[2]: {"The target format is not supported", "Please confirm the execution target of the task"},
	},
	FileContentUnknow: {
		Languages[0]: {"目标内容获取失败", "请确认任务的执行目标"},
		Languages[1]: {"目標內容獲取失敗", "請確認任務的執行目標"},
		Languages[2]: {"Failed to obtain target content", "Please confirm the execution target of the task"},
	},
	ExecutorNotFound: {
		Languages[0]: {"自定义节点不存在", "请检查参数"},
		Languages[1]: {"自定義節點不存在", "請確認參數"},
		Languages[2]: {"Custom action does not exist", "Please check the parameters"},
	},
	DuplicatedAdmin: {
		Languages[0]: {"操作失败，已添加管理员", "请检查参数"},
		Languages[1]: {"操作失敗，已添加管理員", "請確認參數"},
		Languages[2]: {"Operation failed, admin has been added", "Please check the parameters"},
	},
	AgentNotFound: {
		Languages[0]: {"Agent 不存在", "请检查参数"},
		Languages[1]: {"Agent 不存在", "請確認參數"},
		Languages[2]: {"Agent does not exist", "Please check the parameters"},
	},
	AlarmRuleNotFound: {
		Languages[0]: {"告警规则不存在", "请检查参数"},
		Languages[1]: {"告警規則不存在", "請確認參數"},
		Languages[2]: {"Alarm rule does not exist", "Please check the parameters"},
	},
	AlarmRuleAlreadyExists: {
		Languages[0]: {"告警规则已存在", "请检查参数"},
		Languages[1]: {"告警規則已存在", "請確認參數"},
		Languages[2]: {"Alarm rule already exists", "Please check the parameters"},
	},
	UnAvailable: {
		Languages[0]: {"服务不可用", "请检查参数"},
		Languages[1]: {"服務不可用", "請確認參數"},
		Languages[2]: {"Service unavailable", "Please check the parameters"},
	},

	ForbiddenRetryableDagIns: {
		Languages[0]: {"任务实例状态不允许重试", "仅允许重试失败或取消的任务实例"},
		Languages[1]: {"任務實例狀態不允許重試", "僅允許重試失敗或取消的任務實例"},
		Languages[2]: {"Task instance status does not allow retry", "Only failed or canceled task instances are allowed to be retried"},
	},
	UnSupportedTrigger: {
		Languages[0]: {"不允许的触发器类型", "请确认流程可配置触发器类型"},
		Languages[1]: {"不允許的觸發器類型", "請確認流程可配置觸發器類型"},
		Languages[2]: {"Unsupported trigger type", "Please confirm the configurable trigger types for the workflow"},
	},
	WorkflowCycleDetected: {
		Languages[0]: {"流程存在循环调用", "检测到流程中存在循环依赖，请检查节点间的依赖关系"},
		Languages[1]: {"流程存在循環調用", "偵測到流程中存在循環依賴，請檢查節點間的依賴關係"},
		Languages[2]: {"Process cycle detected", "Cyclic dependency found in process, please check dependencies between nodes"},
	},
	TaskAlreayInProgress: {
		Languages[0]: {"任务正在执行中", "请勿重复提交"},
		Languages[1]: {"任務正在執行中", "請勿重複提交"},
		Languages[2]: {"Task is already in progress", "Please do not submit repeatedly"},
	},
	DataSourceIsEmpty: {
		Languages[0]: {"数据源为空", "请稍后再试"},
		Languages[1]: {"數據源為空", "請稍後再試"},
		Languages[2]: {"Data source is empty", "Please try again later"},
	},
	ErrorDepencyService: {
		Languages[0]: {"服务依赖异常", "请检查相关服务状态后重试"},
		Languages[1]: {"服務依賴異常", "請檢查相關服務狀態後重試"},
		Languages[2]: {"Service dependency error", "Please check related services and try again"},
	},
}
