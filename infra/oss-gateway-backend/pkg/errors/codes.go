package errors

// ErrorCode 错误码结构
type ErrorCode struct {
	Code        string // 错误码
	Message     string // 错误消息（简短）
	Description string // 错误描述模板（可使用 %s 等占位符）
	Solution    string // 解决方案
}

// 错误码命名规则：
// 格式：XYYZZZNNN
// X   - HTTP 状态码首位 (4=客户端错误, 5=服务端错误)
// YY  - HTTP 状态码后两位 (00=通用, 01=参数, 31=业务)
// ZZZ - 业务模块编码 (000=通用, 107=存储, 108=存储重复...)
// NNN - 具体错误序号

// ==================== 通用错误码 (400000-400099) ====================

var (
	// BadRequest 400000 - 错误的请求
	BadRequest = ErrorCode{
		Code:        "400000",
		Message:     "Bad request",
		Description: "%s",
		Solution:    "Please check the request parameters",
	}

	// InvalidParam 400001 - 无效参数
	InvalidParam = ErrorCode{
		Code:        "400001",
		Message:     "Invalid parameter",
		Description: "Invalid parameter: %s",
		Solution:    "Please check the parameter format",
	}

	// TooManyKeys 400002 - 请求键过多
	TooManyKeys = ErrorCode{
		Code:        "400002",
		Message:     "Too many keys",
		Description: "Maximum %d keys allowed",
		Solution:    "Please reduce the number of keys",
	}
)

// ==================== 资源未找到错误码 (404000-404099) ====================

var (
	// NotFound 404000 - 资源未找到
	NotFound = ErrorCode{
		Code:        "404000",
		Message:     "Not found",
		Description: "%s",
		Solution:    "Please check the resource ID",
	}

	// StorageNotFound 404001 - 存储未找到
	StorageNotFound = ErrorCode{
		Code:        "404001",
		Message:     "Storage not found",
		Description: "The specified storage does not exist",
		Solution:    "Please check the storage ID",
	}
)

// ==================== 服务器错误码 (500000-500099) ====================

var (
	// InternalError 500000 - 内部服务器错误
	InternalError = ErrorCode{
		Code:        "500000",
		Message:     "Internal server error",
		Description: "%s",
		Solution:    "Please contact the administrator",
	}

	// ConnectionFailed 500001 - 连接失败
	ConnectionFailed = ErrorCode{
		Code:        "500001",
		Message:     "Connection failed",
		Description: "Failed to connect to the storage service",
		Solution:    "Please check the storage configuration and network connectivity",
	}

	// ServiceNotReady 503000 - 服务未就绪
	ServiceNotReady = ErrorCode{
		Code:        "503000",
		Message:     "Service not ready",
		Description: "Some services are not ready",
		Solution:    "Please wait a moment and try again",
	}
)

// ==================== 存储业务错误码 (400031000-400031999) ====================

var (
	// StorageNameExists 400031107 - 存储名称已存在
	// 规则：400 (Bad Request) + 031 (业务错误) + 107 (存储名称)
	StorageNameExists = ErrorCode{
		Code:        "400031107",
		Message:     "Storage name already exists",
		Description: "Storage name(%s) already exists",
		Solution:    "Please use a different storage name",
	}

	// StorageExists 400031108 - 存储已存在（bucket+host 或 bucket+siteId）
	// 规则：400 (Bad Request) + 031 (业务错误) + 108 (存储重复)
	StorageExists = ErrorCode{
		Code:        "400031108",
		Message:     "Storage already exists",
		Description: "%s", // 动态描述：Bucket(xxx) with host(xxx) already exists
		Solution:    "This bucket has already been added with the same host or site ID",
	}

	// StorageDisabled 400031109 - 存储已禁用
	StorageDisabled = ErrorCode{
		Code:        "400031109",
		Message:     "Storage is disabled",
		Description: "The storage is currently disabled",
		Solution:    "Please enable the storage first",
	}

	// InvalidVendorType 400031110 - 无效的供应商类型
	InvalidVendorType = ErrorCode{
		Code:        "400031110",
		Message:     "Invalid vendor type",
		Description: "Invalid vendor type: %s",
		Solution:    "Supported vendor types: OSS, OBS, ECEPH, TOS",
	}

	// InvalidEndpoint 400031111 - 无效的端点
	InvalidEndpoint = ErrorCode{
		Code:        "400031111",
		Message:     "Invalid endpoint",
		Description: "Endpoint must start with http:// or https://",
		Solution:    "Please provide a valid endpoint URL",
	}

	// DefaultStorageExists 400031112 - 默认存储已存在
	DefaultStorageExists = ErrorCode{
		Code:        "400031112",
		Message:     "Default storage already exists",
		Description: "A default storage already exists: %s",
		Solution:    "Please disable the current default storage before setting a new one",
	}
)

// ==================== 认证授权错误码 (401000-403999) ====================

var (
	// Unauthorized 401000 - 未授权
	Unauthorized = ErrorCode{
		Code:        "401000",
		Message:     "Unauthorized",
		Description: "Authentication required",
		Solution:    "Please provide valid credentials",
	}

	// Forbidden 403000 - 禁止访问
	Forbidden = ErrorCode{
		Code:        "403000",
		Message:     "Forbidden",
		Description: "You don't have permission to access this resource",
		Solution:    "Please contact the administrator for access",
	}
)

// GetErrorByCode 根据错误码获取错误信息
func GetErrorByCode(code string) *ErrorCode {
	errorMap := map[string]*ErrorCode{
		"400000":    &BadRequest,
		"400001":    &InvalidParam,
		"400002":    &TooManyKeys,
		"404000":    &NotFound,
		"404001":    &StorageNotFound,
		"500000":    &InternalError,
		"500001":    &ConnectionFailed,
		"503000":    &ServiceNotReady,
		"400031107": &StorageNameExists,
		"400031108": &StorageExists,
		"400031109": &StorageDisabled,
		"400031110": &InvalidVendorType,
		"400031111": &InvalidEndpoint,
		"400031112": &DefaultStorageExists,
		"401000":    &Unauthorized,
		"403000":    &Forbidden,
	}
	return errorMap[code]
}
