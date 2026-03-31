// Package metric
// @description: 指标模型操作接口
// @file metric.go
package metric

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/localize"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

// 常量定义
const (
	RecorderDIP    = "DIP"                        // 产品
	PackageName    = "DataOperatorHub"            // 包名
	OperationType  = "operation"                  // 操作日志类型
	ServiceName    = "agent-operator-integration" // 服务名
	ExMsgLimit     = 65535                        // 附加信息最大长度
	DefaultTimeout = 30 * time.Second             // 默认超时时间
)

// AuditLogOperationModel 审计操作日志模型
type AuditLogOperationModel struct {
	Operation   AuditLogOperationType `json:"operation" validate:"required"`          // 操作类型
	Description string                `json:"description" validate:"required"`        // 字符串描述，最大长度65,535
	OpTime      int64                 `json:"op_time" validate:"required"`            // 操作时间（通过mq上报的必需）精确到纳秒
	Operator    AuditLogOperatorInfo  `json:"operator" validate:"required"`           // 操作者信息
	Object      AuditLogObject        `json:"object,omitempty"`                       // 操作对象信息
	LogFrom     LogFrom               `json:"log_from" validate:"required"`           // 日志来源
	Detail      interface{}           `json:"detail,omitempty"`                       // 细节
	ExMsg       string                `json:"ex_msg,omitempty"`                       // 附加信息，最大长度65,535
	Level       LoggerLevel           `json:"level" validate:"required"`              // 日志级别，默认INFO
	OutBizID    string                `json:"out_biz_id" validate:"required,max=128"` // 外部唯一业务ID，用于防抖，格式不限 最长128
	Type        string                `json:"type" validate:"required"`               // 日志类型，最大长度128
}

// LogFrom 日志来源
type LogFrom struct {
	Package string      `json:"package" validate:"required"` // 大包名
	Service ServiceInfo `json:"service" validate:"required"` // 服务信息
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name string `json:"name" validate:"required"` // 服务名称
}

// LoggerLevel 日志级别
type LoggerLevel string

const (
	LoggerLevelInfo LoggerLevel = "INFO" // 信息
	LoggerLevelWarn LoggerLevel = "WARN" // 警告
)

// AuditLogObjectType 审计日志操作对象类型
type AuditLogObjectType string

const (
	AuditLogObjectOperator AuditLogObjectType = "operator" // 算子
	AuditLogObjectTool     AuditLogObjectType = "tool"     // 工具
	AuditLogObjectMCP      AuditLogObjectType = "mcp"      // mcp
)

// AuditLogObject 操作对象信息
type AuditLogObject struct {
	Type AuditLogObjectType `json:"type" validate:"required"` // 操作对象类型
	Name string             `json:"name"`                     // 操作对象名称，最大长度128
	ID   string             `json:"id"`                       // 操作对象ID，最大长度40
}

// NewAuditLogObject 创建操作对象信息
func NewAuditLogObject(typ AuditLogObjectType, name, id string) *AuditLogObject {
	return &AuditLogObject{
		Type: typ,
		Name: name,
		ID:   id,
	}
}

// AuditLogOperatoAgent 操作者代理信息
type AuditLogOperatoAgent struct {
	Type string `json:"type" validate:"required"` // 操作者客户端类型
	IP   string `json:"ip" validate:"required"`   // 操作者设备IP
	MAC  string `json:"mac" validate:"required"`  // 操作者设备mac地址
}

// AuditLogOperatorInfo 操作者信息
type AuditLogOperatorInfo struct {
	ID    string               `json:"id" validate:"required,max=40"`    // 操作者ID，最大长度40
	Name  string               `json:"name" validate:"required,max=128"` // 操作者名称，以传入数据为准，最大长度128,type为internal_service必传
	Type  AuditLogOperatorType `json:"type" validate:"required"`         // 操作者类型
	Agent AuditLogOperatoAgent `json:"agent" validate:"required"`        // 操作者代理信息
}

// AuditLogOperatorType 操作者类型
type AuditLogOperatorType string

const (
	AuthenticatedUser AuditLogOperatorType = "authenticated_user" // 实名用户
	AnonymousUser     AuditLogOperatorType = "anonymous_user"     // 匿名用户
	AppUser           AuditLogOperatorType = "app"                // 应用账户
	InternalService   AuditLogOperatorType = "internal_service"   // 内部服务
)

// Validate 校验操作者类型是否合法
func (a AuditLogOperatorType) Validate() error {
	validTypeMap := map[AuditLogOperatorType]struct{}{
		AuthenticatedUser: {},
		AnonymousUser:     {},
		AppUser:           {},
		InternalService:   {},
	}
	for t := range validTypeMap {
		if a == t {
			return nil
		}
	}
	return fmt.Errorf("invalid operator type %s", a)
}

// AuditLogOperationType 审计日志操作类型
type AuditLogOperationType string

const (
	AuditLogOperationCreate    AuditLogOperationType = "create"    // 新建
	AuditLogOperationDelete    AuditLogOperationType = "delete"    // 删除
	AuditLogOperationEdit      AuditLogOperationType = "edit"      // 编辑
	AuditLogOperationPublish   AuditLogOperationType = "publish"   // 发布
	AuditLogOperationUnpublish AuditLogOperationType = "unpublish" // 取消发布（下架）
	AuditLogOperationExecute   AuditLogOperationType = "execute"   // 执行
)

// AuditLogBuilder 审计日志构建器
type AuditLogBuilder struct {
	ts                 *localize.I18nTranslator
	logger             interfaces.Logger
	topic              string
	outboxMessageEvent interfaces.IOutboxMessageEvent
}

// NewAuditLogBuilder 创建审计日志构建器
func NewAuditLogBuilder() *AuditLogBuilder {
	return &AuditLogBuilder{
		ts:                 localize.NewI18nTranslator(config.NewConfigLoader().Project.Language),
		logger:             config.NewConfigLoader().GetLogger(),
		topic:              interfaces.AuditLogTopic,
		outboxMessageEvent: common.NewOutboxMessageEvent(),
	}
}

// AuditLogBuilderParams 审计日志构建参数
type AuditLogBuilderParams struct {
	TokenInfo    *interfaces.TokenInfo    // 令牌信息
	Accessor     *interfaces.AuthAccessor // 访问者信息
	Operation    AuditLogOperationType    // 操作类型
	Object       *AuditLogObject          // 操作对象
	Description  string                   // 描述信息
	ExMsg        string                   // 异常信息
	Detils       interface{}              // 操作细节
	OperatorType AuditLogOperatorType     // 操作者类型
}

// AuditLogToolDetil 工具操作细节
type AuditLogToolDetil struct {
	ToolID   string `json:"tool_id"`   // 工具ID
	ToolName string `json:"tool_name"` // 工具名称
}

// AuditLogToolDetils 工具操作细节
type AuditLogToolDetils struct {
	Infos         []AuditLogToolDetil `json:",inline"`
	OperationCode OperationCode
}

// NewAuditLogToolDetils 创建工具操作细节
func NewAuditLogToolDetils(operationCode OperationCode, infos []AuditLogToolDetil) *AuditLogToolDetils {
	return &AuditLogToolDetils{
		Infos:         infos,
		OperationCode: operationCode,
	}
}

// OperationCode 操作代码
type OperationCode string

// 工具附加信息
const (
	// "从算子导入工具“%s”成功"
	ImportToolFromOperator OperationCode = "import_tool_from_operator"
	// "添加工具“%s”到工具箱成功"
	AddTool OperationCode = "add_tool"
	// "编辑工具“%s”成功"
	EditTool OperationCode = "edit_tool"
	// "从工具箱移除工具“%s”成功"
	DeleteTool OperationCode = "remove_tool"
	// "更新工具状态“%s”成功"
	UpdateToolStatus OperationCode = "update_tool_status"
	// "调试工具“%s”成功",
	DebugTool OperationCode = "debug_tool"
	// "执行工具“%s”成功"
	ExecuteTool OperationCode = "execute_tool"
	// 未知操作
	UnknownOperation OperationCode = "unknown_operation"
)

func (b *AuditLogBuilder) getToolDetailsAndExMsg(param interface{}) (detils interface{}, exMsg string) {
	if param == nil {
		return
	}
	p, ok := param.(*AuditLogToolDetils)
	if !ok {
		b.logger.Errorf("invalid detils type")
		return
	}
	if len(p.Infos) == 0 {
		return
	}
	detils = map[string]interface{}{
		"tool_infos": p.Infos,
	}
	var toolNames string
	for i, info := range p.Infos {
		if info.ToolName == "" {
			continue
		}
		if i == 0 {
			toolNames += info.ToolName
			continue
		}
		toolNames += "," + info.ToolName
	}
	switch p.OperationCode {
	case ImportToolFromOperator, AddTool, EditTool, DeleteTool, UpdateToolStatus, DebugTool, ExecuteTool:
		exMsg = fmt.Sprintf(b.ts.Trans(fmt.Sprintf("audit_log.%s", p.OperationCode)), toolNames)
	case UnknownOperation:
		exMsg = fmt.Sprintf(b.ts.Trans(fmt.Sprintf("audit_log.%s", UnknownOperation)), toolNames)
	default:
		exMsg = fmt.Sprintf(b.ts.Trans(fmt.Sprintf("audit_log.%s", UnknownOperation)), toolNames)
	}
	if len(exMsg) > ExMsgLimit {
		exMsg = exMsg[:ExMsgLimit]
	}
	return
}

// GetOperatorType 获取操作者类型
func (p *AuditLogBuilderParams) GetOperatorType() error {
	if p.OperatorType != "" {
		return p.OperatorType.Validate()
	}
	if p.TokenInfo == nil {
		return fmt.Errorf("token info is nil")
	}
	var operatorType AuditLogOperatorType
	switch p.TokenInfo.VisitorTyp {
	case interfaces.RealName:
		operatorType = AuthenticatedUser
	case interfaces.Anonymous:
		operatorType = AnonymousUser
	case interfaces.Business:
		operatorType = AppUser
	default:
		operatorType = InternalService
	}
	p.OperatorType = operatorType
	return nil
}

// Build 构建审计日志模型
func (b *AuditLogBuilder) build(p *AuditLogBuilderParams) (interface{}, error) {
	if p.TokenInfo == nil {
		return nil, fmt.Errorf("token info is nil")
	}
	if p.Accessor == nil {
		return nil, fmt.Errorf("accessor is nil")
	}
	err := p.GetOperatorType()
	if err != nil {
		return nil, err
	}
	var level LoggerLevel
	switch p.Operation {
	case AuditLogOperationCreate, AuditLogOperationEdit, AuditLogOperationPublish,
		AuditLogOperationUnpublish, AuditLogOperationExecute:
		level = LoggerLevelInfo
	case AuditLogOperationDelete:
		level = LoggerLevelWarn
	default:
		return nil, fmt.Errorf("invalid operation type")
	}
	// 组织
	logObj := &AuditLogOperationModel{
		Operation:   p.Operation,
		Description: p.Description,
		OpTime:      time.Now().UnixNano(),
		Operator: AuditLogOperatorInfo{
			Type: p.OperatorType,
			Name: p.Accessor.Name,
			ID:   p.Accessor.ID,
		},
		LogFrom: LogFrom{
			Package: PackageName,
			Service: ServiceInfo{
				Name: ServiceName,
			},
		},
		Detail:   p.Detils,
		ExMsg:    p.ExMsg,
		Level:    level,
		OutBizID: uuid.New().String(),
		Type:     OperationType,
	}
	// 内部服务不记录Agent信息
	if logObj.Operator.Type == AuthenticatedUser || logObj.Operator.Type == AnonymousUser || logObj.Operator.Type == AppUser {
		logObj.Operator.Agent = AuditLogOperatoAgent{
			Type: p.TokenInfo.ClientTyp.String(),
			IP:   p.TokenInfo.LoginIP,
			MAC:  p.TokenInfo.MAC,
		}
	}
	if p.Object == nil {
		return logObj, nil
	}
	logObj.Object = *p.Object
	if logObj.Description != "" {
		return logObj, nil
	}
	logObj.Description = b.ts.Trans(fmt.Sprintf("audit_log.%s_%s", p.Operation, p.Object.Type),
		p.Object.Name)
	switch p.Object.Type {
	case AuditLogObjectOperator, AuditLogObjectMCP:
		logObj.Description = fmt.Sprintf(b.ts.Trans(fmt.Sprintf("audit_log.%s_%s", p.Object.Type, p.Operation)), p.Object.Name)
	case AuditLogObjectTool:
		logObj.Detail, logObj.ExMsg = b.getToolDetailsAndExMsg(logObj.Detail)
		logObj.Description = fmt.Sprintf(b.ts.Trans(fmt.Sprintf("audit_log.%s_%s", p.Object.Type, p.Operation)), p.Object.Name)
	}
	return logObj, nil
}

// Logger 记录审计日志
func (b *AuditLogBuilder) Logger(ctx context.Context, p *AuditLogBuilderParams) {
	if !config.GetAuthEnabled() {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}
	newCtx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel() // 确保资源释放
	logObj, err := b.build(p)
	if err != nil {
		b.logger.WithContext(newCtx).Errorf("build audit log failed: %v", err)
		return
	}
	err = b.outboxMessageEvent.Publish(newCtx, &interfaces.OutboxMessageReq{
		EventID:   uuid.New().String(),
		EventType: interfaces.OutboxMessageEventTypeAuditLog,
		Topic:     interfaces.AuditLogTopic,
		Payload:   utils.ObjectToJSON(logObj),
	})
	if err != nil {
		b.logger.WithContext(newCtx).Errorf("write audit log failed: %v", err)
	}
}
