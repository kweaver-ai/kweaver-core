package agentexecutordto

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req/chatopt"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

type AgentCallReq struct {
	ID           string `json:"id,omitempty"`
	AgentVersion string `json:"agent_version,omitempty"`
	Config       Config `json:"config,omitempty"`
	// Input  Input  `json:"input"`
	Input map[string]interface{} `json:"input"`

	UserID      string               `json:"-"`
	Token       string               `json:"-"`
	VisitorType constant.VisitorType `json:"-"`

	CallType constant.CallType `json:"-"`

	ExecutorVersion string `json:"-"` // 使用的executor版本 v1 或 v2

	XAccountID        string            `json:"-"` // 用户ID
	XAccountType      cenum.AccountType `json:"-"` // 用户类型 app/user/anonymous
	XBusinessDomainID string            `json:"-"`
	// ConversationSessionID string            `json:"conversation_session_id"`

	ChatOption chatopt.ChatOption `json:"chat_option"`

	// 新增：中断恢复信息（统一 Run 接口支持恢复执行）
	ResumeInterruptInfo *v2agentexecutordto.AgentResumeInfo `json:"-"`
}
type AgentOptions struct {
	Stream bool `json:"stream"`
	Debug  bool `json:"debug"`
	Retry  bool `json:"retry"`
	// NOTE: 一个动态运行时需要的字段，不是固定的agent配置中的数据源范围，而是传参的数据源范围
	DynamicRetrieverFields RetrieverDataSource `json:"dynamic_retriever_fields"`
	// UserDefine             map[string]interface{} `json:"user_define,omitempty"`
	Step string `json:"step"`
}
type Input struct {
	Query   string                  `json:"query"`
	File    []valueobject.TempFile  `json:"file"`
	Env     map[string]interface{}  `json:"_object"`
	Content interface{}             `json:"content,omitempty"`
	History []*comvalobj.LLMMessage `json:"history"`
	Options AgentOptions            `json:"_options,omitempty"`
	Object  map[string]interface{}  `json:"object"`
	Tool    interface{}             `json:"tool"` // ask 请求ad参数
	// 扩展字段
	ExtendedFields map[string]interface{} `json:"-"`
	// ChatMode       string                 `json:"chat_mode"`
	// ConfirmPlan bool `json:"confirm_plan"`
}

type KgSource struct {
	KgID            string              `json:"kg_id"`
	Fields          []string            `json:"fields"`
	OutputFields    []string            `json:"output_fields"`
	FieldProperties map[string][]string `json:"field_properties"`
}
type DocFields struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Source string `json:"source"`
}

type DocSource struct {
	// 本地上传文件输入
	FileSource string `json:"file_source"`
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	// AS文件
	DsID     string       `json:"ds_id,omitempty"`
	Fields   []*DocFields `json:"fields,omitempty"`
	DataSets []string     `json:"datasets,omitempty"`
	// 这个根据数据源id自动填入
	Address  string `json:"address,omitempty"`
	Port     int    `json:"port,omitempty"`
	AsUserID string `json:"as_user_id,omitempty"`
	// 标识这个数据源只用于as鉴权，不召回文件
	Disabled bool `json:"disabled"`
}

// 数据源
type RetrieverDataSource struct {
	Kg  []*KgSource  `json:"kg"`  // 图谱类型数据源
	Doc []*DocSource `json:"doc"` // 文档类型数据源
}

type Config struct {
	daconfvalobj.Config `json:",inline"`
	// NOTE: 新增字段，用于请求中透传信息
	// UserDefine map[string]string `json:"user_define"`
	AgentID        string `json:"agent_id"`
	ConversationID string `json:"conversation_id"`
	SessionID      string `json:"session_id"`
}
