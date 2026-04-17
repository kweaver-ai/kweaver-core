package swagger

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj/skillvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
)

// AgentConfigCreateRes 创建 agent 的响应
type AgentConfigCreateRes = agentconfigresp.CreateRes

// AgentConfigCreateReq 创建 agent 的请求体
type AgentConfigCreateReq struct {
	Key           string             `json:"key"`             // agent 标识
	IsBuiltIn     *cdaenum.BuiltIn   `json:"is_built_in"`     // 是否内置（内部接口有效）
	IsSystemAgent *cenum.YesNoInt8   `json:"is_system_agent"` // 是否是系统智能体
	Name          string             `json:"name"`            // 名字
	Profile       string             `json:"profile"`         // 简介
	AvatarType    cdaenum.AvatarType `json:"avatar_type"`     // 头像类型
	Avatar        string             `json:"avatar"`          // 头像信息
	ProductKey    string             `json:"product_key"`     // 所属产品标识
	Config        *AgentConfigConfig `json:"config"`          // agent 配置
	CreatedBy     string             `json:"created_by"`      // 创建人 uid（内部接口有效）
}

// AgentConfigUpdateReq 编辑 agent 的请求体
type AgentConfigUpdateReq struct {
	IsBuiltIn  *cdaenum.BuiltIn   `json:"is_built_in"` // 是否内置（内部接口有效）
	Name       string             `json:"name"`        // 名字
	Profile    string             `json:"profile"`     // 描述
	ProductKey string             `json:"product_key"` // 所属产品标识
	AvatarType cdaenum.AvatarType `json:"avatar_type"` // 头像类型
	Avatar     string             `json:"avatar"`      // 头像信息
	Config     *AgentConfigConfig `json:"config"`      // agent 配置
	UpdatedBy  string             `json:"updated_by"`  // 更新人 uid（内部接口有效）
}

// AgentConfigDetailRes agent 详情响应
type AgentConfigDetailRes struct {
	ID            string             `json:"id"`              // agent id
	Key           string             `json:"key"`             // agent 标识
	IsBuiltIn     cdaenum.BuiltIn    `json:"is_built_in"`     // 是否内置
	IsSystemAgent *cenum.YesNoInt8   `json:"is_system_agent"` // 是否是系统智能体
	Name          string             `json:"name"`            // 名字
	Profile       string             `json:"profile"`         // 描述
	AvatarType    cdaenum.AvatarType `json:"avatar_type"`     // 头像类型
	Avatar        string             `json:"avatar"`          // 头像信息
	ProductKey    string             `json:"product_key"`     // 所属产品标识
	ProductName   string             `json:"product_name"`    // 所属产品名称
	Config        *AgentConfigConfig `json:"config"`          // agent 配置
	Status        cdaenum.Status     `json:"status"`          // 状态
	IsPublished   bool               `json:"is_published"`    // 是否发布过
}

// AgentConfigConfig agent 配置
type AgentConfigConfig struct {
	Input                     *daconfvalobj.Input                     `json:"input"`                       // 输入参数
	SystemPrompt              string                                  `json:"system_prompt"`               // 系统提示词
	Dolphin                   string                                  `json:"dolphin"`                     // Dolphin 语句
	DolphinEnhance            string                                  `json:"dolphin_enhance"`             // Dolphin 增强语句
	IsDolphinMode             cdaenum.DolphinMode                     `json:"is_dolphin_mode"`             // 是否是 dolphin 模式
	PreDolphin                []*daconfvalobj.DolphinTpl              `json:"pre_dolphin"`                 // 在用户自定义 dolphin 之前执行的内置 dolphin 语句
	PostDolphin               []*daconfvalobj.DolphinTpl              `json:"post_dolphin"`                // 在用户自定义 dolphin 之后执行的内置 dolphin 语句
	DataSource                *datasourcevalobj.RetrieverDataSource   `json:"data_source"`                 // 数据源
	Skills                    *AgentConfigSkills                      `json:"skills"`                      // 技能
	Llms                      []*daconfvalobj.LlmItem                 `json:"llms"`                        // LLM 配置
	IsDataFlowSetEnabled      int                                     `json:"is_data_flow_set_enabled"`    // 是否启用数据流设置
	OpeningRemarkConfig       *daconfvalobj.OpeningRemarkConfig       `json:"opening_remark_config"`       // 开场白配置
	PresetQuestions           []*daconfvalobj.PresetQuestion          `json:"preset_questions"`            // 预设问题列表
	Output                    *daconfvalobj.Output                    `json:"output"`                      // 输出结果
	BuiltInCanEditFields      *daconfvalobj.BuiltInCanEditFields      `json:"built_in_can_edit_fields"`    // 内置 agent 可编辑字段配置
	Memory                    *daconfvalobj.MemoryCfg                 `json:"memory"`                      // 长期记忆配置
	RelatedQuestion           *daconfvalobj.RelatedQuestion           `json:"related_question"`            // 相关问题配置
	PlanMode                  *daconfvalobj.PlanMode                  `json:"plan_mode"`                   // 任务规划模式配置
	NonDolphinModeConfig      *daconfvalobj.NonDolphinModeConfig      `json:"non_dolphin_mode_config"`     // 非Dolphin模式配置
	ConversationHistoryConfig *daconfvalobj.ConversationHistoryConfig `json:"conversation_history_config"` // 会话历史配置
	Metadata                  daconfvalobj.ConfigMetadata             `json:"metadata"`                    // 配置元数据
}

// AgentConfigSkills 技能配置
type AgentConfigSkills struct {
	Tools  AgentConfigSkillTools  `json:"tools"`  // 工具技能
	Agents AgentConfigSkillAgents `json:"agents"` // Agent 技能
	MCPs   AgentConfigSkillMCPs   `json:"mcps"`   // MCP 技能
	Skills AgentConfigSkillSkills `json:"skills"` // skill
}

// AgentConfigSkillTools 工具技能列表
type AgentConfigSkillTools []*AgentConfigSkillTool

// AgentConfigSkillAgents Agent 技能列表
type AgentConfigSkillAgents []*AgentConfigSkillAgent

// AgentConfigSkillMCPs MCP 技能列表
type AgentConfigSkillMCPs []*AgentConfigSkillMCP

// AgentConfigSkillSkills skill 列表
type AgentConfigSkillSkills []*AgentConfigSkillSkill

// AgentConfigSkillTool 工具技能配置
type AgentConfigSkillTool struct {
	ToolID                          string                              `json:"tool_id"`                           // 工具 ID
	ToolBoxID                       string                              `json:"tool_box_id"`                       // 工具箱 ID
	ToolTimeout                     int                                 `json:"tool_timeout"`                      // 工具调用超时时间
	ToolInput                       AgentConfigSkillToolInputs          `json:"tool_input"`                        // 输入参数配置
	Intervention                    bool                                `json:"intervention"`                      // 是否启用人工干预
	InterventionConfirmationMessage string                              `json:"intervention_confirmation_message"` // 人工干预确认消息
	ResultProcessStrategies         []skillvalobj.ResultProcessStrategy `json:"result_process_strategies"`         // 结果处理策略
}

// AgentConfigSkillToolInputs 工具输入列表
type AgentConfigSkillToolInputs []*AgentConfigSkillToolInput

// AgentConfigSkillToolInput 工具输入配置
type AgentConfigSkillToolInput struct {
	Children  AgentConfigSkillToolInputs `json:"children"`   // 子字段
	Enable    bool                       `json:"enable"`     // 是否启用
	InputDesc string                     `json:"input_desc"` // 输入描述
	InputName string                     `json:"input_name"` // 输入名称
	InputType string                     `json:"input_type"` // 输入类型
	MapType   string                     `json:"map_type"`   // 值类型
	MapValue  any                        `json:"map_value"`  // 值
}

// AgentConfigSkillAgent Agent 技能配置
type AgentConfigSkillAgent struct {
	AgentKey                        string                        `json:"agent_key"`                         // Agent Key
	AgentVersion                    string                        `json:"agent_version"`                     // Agent 版本
	AgentInput                      AgentConfigSkillAgentInputs   `json:"agent_input"`                       // 技能输入
	Intervention                    bool                          `json:"intervention"`                      // 是否启用干预
	InterventionConfirmationMessage string                        `json:"intervention_confirmation_message"` // 人工干预确认消息
	DataSourceConfig                *skillvalobj.DataSourceConfig `json:"data_source_config"`                // 数据源配置
	LlmConfig                       *skillvalobj.LLMConfig        `json:"llm_config"`                        // LLM 配置
	AgentTimeout                    int                           `json:"agent_timeout"`                     // Agent 工具调用超时时间
	CurrentPmsCheckStatus           string                        `json:"current_pms_check_status"`          // 当前此技能的使用权限状态
	CurrentIsExistsAndPublished     bool                          `json:"current_is_exists_and_published"`   // 当前是否存在并已发布
}

// AgentConfigSkillAgentInputs Agent 输入列表
type AgentConfigSkillAgentInputs []*AgentConfigSkillAgentInput

// AgentConfigSkillAgentInput Agent 输入配置
type AgentConfigSkillAgentInput struct {
	Children  AgentConfigSkillAgentInputs `json:"children"`   // 子字段
	Enable    bool                        `json:"enable"`     // 是否启用
	InputName string                      `json:"input_name"` // 输入名称
	InputType string                      `json:"input_type"` // 输入类型
	MapType   string                      `json:"map_type"`   // 值类型
	MapValue  any                         `json:"map_value"`  // 值
}

// AgentConfigSkillMCP MCP 技能配置
type AgentConfigSkillMCP struct {
	MCPServerID string `json:"mcp_server_id"` // mcp server id
}

// AgentConfigSkillSkill skill 配置
type AgentConfigSkillSkill struct {
	SkillID string `json:"skill_id"` // skill id
}
