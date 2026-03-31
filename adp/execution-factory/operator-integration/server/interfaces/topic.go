package interfaces

/*该文件定义了Topic常量，统一记录方便统计 */

const (
	// ChannelMessage 算子集成事件Channel
	ChannelMessage = "operator_integration" // channel
)

// 监听外部事件Topic列表
const (
	// AuthResourceNameModifyTopic 资源名称变更Topic
	AuthResourceNameModifyTopic = "authorization.resource.name.modify"
	// AuditLogTopic 审计日志Topic
	AuditLogTopic = "isf.audit_log.log"
)

// 通知外部事件Topick列表

const (
	// OperatorDeleteEventTopic 算子删除事件Topic
	OperatorDeleteEventTopic = "agent_operator_integration.operator.delete"
)

// OperatorDeleteEvent 算子删除事件
type OperatorDeleteEvent struct {
	OperatorID   string                 `json:"operator_id"`
	Version      string                 `json:"version"`
	Status       BizStatus              `json:"status"`
	IsInternal   bool                   `json:"is_internal"`                                          // 是否内部算子
	IsDataSource bool                   `json:"is_data_source" form:"is_data_source" default:"false"` // 是否为数据源算子
	ExtendInfo   map[string]interface{} `json:"extend_info"`
	OperatorType OperatorType           `json:"operator_type" form:"operator_type" default:"basic" validate:"oneof=basic composite"` // 算子类型(basic/composite
	UpdateUser   string                 `json:"update_user"`
}
