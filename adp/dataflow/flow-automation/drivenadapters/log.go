package drivenadapters

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/field"
	spanLog "github.com/kweaver-ai/TelemetrySDK-Go/span/v2/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
)

// 常量定义
const (
	RecorderAnyShare                  = "AnyShare"
	RecorderDIP                       = "DIP"
	AnyshPackageName                  = "AnyShareMainModule"
	DIPPackageName                    = "DataFlow"
	LogTypeOperation                  = "operation"
	ClientTypeUnknown                 = "unknown"
	DefaultLoginIP                    = "127.0.0.1"
	LogObjectType                     = "data_flow"
	NcTLogLevel_NCT_LL_INFO           = 1
	NcTLogLevel_NCT_LL_WARN           = 2
	NcTLogLevel_NCT_LL_INFO_Str       = "INFO"
	NcTLogLevel_NCT_LL_WARN_Str       = "WARN"
	NcTDocOperType_NCT_DOT_AUTOMATION = 28
	// 操作类型定义
	CreateOperation  = "create"
	UpdateOperation  = "update"
	DeleteOperation  = "delete"
	ExecuteOperation = "execute"
)

// AuditLog 审计日志结构体
type auditLog struct {
	UserID         string `json:"user_id"`
	UserName       string `json:"user_name"`
	UserType       string `json:"user_type"`
	Level          int    `json:"level"`
	OpType         int    `json:"op_type"`
	Date           int64  `json:"date"`
	IP             string `json:"ip"`
	Mac            string `json:"mac"`
	Msg            string `json:"msg"`
	ExMsg          string `json:"ex_msg"`
	UserAgent      string `json:"user_agent"`
	ObjId          string `json:"obj_id"`
	AdditionalInfo string `json:"additional_info"`
	OutBizID       string `json:"out_biz_id"`
	DeptPaths      string `json:"dept_paths"`
}

type arLog struct {
	Recorder    string                 `json:"recorder"`
	Operation   string                 `json:"operation"`
	Description string                 `json:"description"`
	Object      map[string]interface{} `json:"object"`
	Operator    Operator               `json:"operator"`
	LogFrom     map[string]interface{} `json:"log_from"`
}

type dipFlowAuditLog struct {
	Operation   string                 `json:"operation"`
	Description string                 `json:"description"`
	OpTime      int64                  `json:"op_time"`
	Operator    Operator               `json:"operator"`
	Object      LogObject              `json:"object"`
	LogFrom     map[string]interface{} `json:"log_from"`
	Detail      map[string]interface{} `json:"detail,omitempty"`
	ExMsg       string                 `json:"ex_msg"`
	Level       string                 `json:"level"`
	OutBizID    string                 `json:"out_biz_id"`
	Type        string                 `json:"type"`
}

type Operator struct {
	Type           string           `json:"type"`
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Agent          *Agent           `json:"agent,omitempty"`
	DepartmentPath []DepartmentPath `json:"department_path,omitempty"`
	IsSystemOp     bool             `json:"is_system_op,omitempty"`
}

type Agent struct {
	UdID string `json:"udid,omitempty"`
	IP   string `json:"ip"`
	Type string `json:"type"`
	Mac  string `json:"mac"`
}

type LogObject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// BuildAuditLogParams 构建审计日志参数
type BuildAuditLogParams struct {
	UserInfo *UserInfo
	Msg      string
	ExtMsg   string
	OutBizID string
	LogLevel int
}

// BuildARLogParams 构建ar上报日志参数
type BuildARLogParams struct {
	Operation   string
	Description string
	UserInfo    *UserInfo
	Object      map[string]interface{}
}

type BuildDIPFlowAuditLog struct {
	UserInfo  *UserInfo
	Msg       string
	ExtMsg    string
	OutBizID  string
	LogLevel  string
	Operation string
	ObjID     string
	ObjName   string
}

// LogBuilder 日志构建器接口
type LogBuilder interface {
	Build(params interface{}) (interface{}, error)
}

// AuditLogBuilder 审计日志构建器
type AuditLogBuilder struct{}

func (b *AuditLogBuilder) Build(params interface{}) (interface{}, error) {
	p := params.(*BuildAuditLogParams)
	user := p.UserInfo
	audit := &auditLog{
		UserID:    user.UserID,
		UserName:  user.UserName,
		UserType:  common.AuthenticatedUserType,
		Level:     p.LogLevel,
		OpType:    NcTDocOperType_NCT_DOT_AUTOMATION,
		Date:      time.Now().UnixNano() / 1e3,
		IP:        user.LoginIP,
		Msg:       p.Msg,
		ExMsg:     p.ExtMsg,
		UserAgent: user.UserAgent,
		OutBizID:  p.OutBizID,
	}
	if audit.IP == "" {
		audit.IP = os.Getenv("POD_IP")
	}
	if _, ok := user.ParentDeps.([]interface{}); ok {
		audit.DeptPaths = strings.Join(user.DepartmentNames, ",")
	} else if user.UserID != "anonymous" {
		userDetail, err := NewUserManagement().GetUserInfoByType(user.UserID, user.AccountType)
		if err != nil {
			audit.DeptPaths = "未分配组"
		} else {
			audit.DeptPaths = strings.Join(userDetail.DepartmentNames, ",")
		}
	}
	return audit, nil
}

// ARLogBuilder AR日志构建器
type ARLogBuilder struct {
	Recorder string
	Package  string
	Service  string
}

func (b *ARLogBuilder) Build(params interface{}) (interface{}, error) {
	p := params.(*BuildARLogParams)
	user := p.UserInfo
	ar := &arLog{
		Recorder:    b.Recorder,
		Operation:   p.Operation,
		Description: p.Description,
		Object:      p.Object,
		Operator: Operator{
			Type:           user.VisitorType,
			ID:             user.UserID,
			Name:           user.UserName,
			DepartmentPath: user.DepartmentPaths,
			Agent: &Agent{
				UdID: user.UdID,
				IP:   user.LoginIP,
				Type: ifEmpty(user.ClientType, ClientTypeUnknown),
			},
		},
		LogFrom: map[string]interface{}{
			"package": b.Package,
			"service": map[string]interface{}{
				"name": b.Service,
				"instance": map[string]interface{}{
					"id": os.Getenv("HOSTNAME"),
				},
			},
		},
	}
	return ar, nil
}

// DIPFlowAuditLogBuilder DIPFlow日志构建器
type DIPFlowAuditLogBuilder struct{}

func (b *DIPFlowAuditLogBuilder) Build(params interface{}) (interface{}, error) {
	p := params.(*BuildDIPFlowAuditLog)
	user := p.UserInfo
	dLog := &dipFlowAuditLog{
		Operation:   p.Operation,
		Description: p.Msg,
		OpTime:      time.Now().UnixNano(),
		Operator: Operator{
			Type: user.VisitorType,
			ID:   user.UserID,
			Name: user.UserName,
		},
		Object: LogObject{
			ID:   p.ObjID,
			Name: p.ObjName,
			Type: LogObjectType,
		},
		LogFrom: map[string]interface{}{
			"package": DIPPackageName,
			"service": map[string]interface{}{
				"name": common.FlowServiceName,
			},
		},
		ExMsg:    p.ExtMsg,
		OutBizID: p.OutBizID,
		Level:    p.LogLevel,
		Type:     LogTypeOperation,
	}

	// 实名用户或匿名用户 Agent信息必填
	if user.VisitorType == common.AuthenticatedUserType || user.VisitorType == common.AnonymousUserType {
		dLog.Operator.Agent = &Agent{
			IP:   ifEmpty(user.LoginIP, os.Getenv("POD_IP")),
			Type: user.ClientType,
			Mac:  user.Mac,
		}
	}

	return dLog, nil
}

// LogWriter 日志输出器接口
type LogWriter interface {
	Write(topic string, logObj interface{}) error
}

// JSONLogWriter AS、DIP审计日志或运营日志输出器
type JSONLogWriter struct {
	SendFunc func(string, []byte) error
}

func (w *JSONLogWriter) Write(topic string, logObj interface{}) error {
	if topic == "" {
		return nil
	}
	data, err := json.Marshal(logObj)
	if err != nil {
		return err
	}
	return w.SendFunc(topic, data)
}

// O11yLogWriter 可观测性日志概览输出器
type O11yLogWriter struct {
	Logger spanLog.Logger
}

func (w *O11yLogWriter) Write(_ string, logObj interface{}) error {
	if w.Logger == nil {
		return nil
	}
	w.Logger.InfoField(field.MallocJsonField(logObj), common.FlowServiceName)
	return nil
}

type Logger interface {
	Log(logType string, params interface{}, writer LogWriter)
	LogO11y(params *BuildARLogParams, writer LogWriter)
}

// logger
type logger struct {
	builders map[string]LogBuilder
}

// 日志类型常量
const (
	LogTypeASAuditLog      = "as_audit_log"       // AS审计日志
	LogTypeASOperationLog  = "as_operation_log"   // AS运营日志
	LogTypeDIPFlowAduitLog = "dip_flow_audit_log" // DataFlow审计日志
	LogTypeO11yOverview    = "o11y_overview"      // DataFlow可观测性概览
)

// Log AS、DIP审计日志或运营日志记录器
func (l *logger) Log(logType string, params interface{}, writer LogWriter) {
	builder, ok := l.builders[logType]
	if !ok {
		return
	}
	logObj, err := builder.Build(params)
	if err != nil {
		return
	}
	var topic string
	switch logType {
	case LogTypeASAuditLog:
		topic = common.TopicAuditLog
	case LogTypeASOperationLog:
		topic = common.TopicARLog
	case LogTypeDIPFlowAduitLog:
		topic = common.TopicDIPFlowAuditLog
	default:
		topic = ""
	}
	writer.Write(topic, logObj)
}

// LogO11y 可观测性概览日志记录器
func (l *logger) LogO11y(params *BuildARLogParams, writer LogWriter) {
	builder, ok := l.builders[LogTypeO11yOverview]
	if !ok {
		return
	}
	logObj, err := builder.Build(params)
	if err != nil {
		return
	}
	writer.Write("", logObj)
}

// ifEmpty 工具函数, 如果不存在则返回默认值
func ifEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

var (
	onceLogger   sync.Once
	globalLogger Logger
)

// NewLogger 初始化日志实例
func NewLogger() Logger {
	onceLogger.Do(func() {
		builders := map[string]LogBuilder{
			LogTypeASAuditLog:      &AuditLogBuilder{},
			LogTypeASOperationLog:  &ARLogBuilder{Recorder: RecorderAnyShare, Package: AnyshPackageName, Service: common.ServiceName},
			LogTypeDIPFlowAduitLog: &DIPFlowAuditLogBuilder{},
			LogTypeO11yOverview:    &ARLogBuilder{Recorder: RecorderDIP, Package: DIPPackageName, Service: common.FlowServiceName},
		}
		globalLogger = &logger{builders: builders}
	})
	return globalLogger
}
