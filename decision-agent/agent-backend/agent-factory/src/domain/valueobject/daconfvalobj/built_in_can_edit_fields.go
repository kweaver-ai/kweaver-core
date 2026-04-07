package daconfvalobj

// BuiltInCanEditFields 表示内置agent可编辑字段配置
// 当is_built_in为1时有效
// 用途：
//   - 内置agent创建者可以根据需要来设置此字段，来控制用户可以编辑的agent配置字段
//   - 前端根据此字段来控制用户可以编辑的agent配置字段
//   - agent后端编辑接口根据此字段来验证前端传过来的字段是否符合要求
type BuiltInCanEditFields struct {
	Name          bool `json:"name"`            // 名字
	Avatar        bool `json:"avatar"`          // 头像
	Profile       bool `json:"profile"`         // 描述
	InputConfig   bool `json:"input_config"`    // 输入配置
	SystemPrompt  bool `json:"system_prompt"`   // 角色指令，包括dolpin模式和非dolpin模式
	DataSourceKG  bool `json:"data_source.kg"`  // 知识来源-业务知识网络
	DataSourceDoc bool `json:"data_source.doc"` // 知识来源-文档
	// KnowledgeSource     bool `json:"knowledge_source"`      // 知识来源
	Model               bool `json:"model"`                 // 模型配置
	Skills              bool `json:"skills"`                // 技能
	OpeningRemarkConfig bool `json:"opening_remark_config"` // 开场白配置
	PresetQuestions     bool `json:"preset_questions"`      // 预设问题列表
	// skills.tools.tool_input
	SkillsToolsToolInput bool `json:"skills.tools.tool_input"` // 技能工具输入

	Memory          bool `json:"memory"`           // 长期记忆配置
	RelatedQuestion bool `json:"related_question"` // 相关问题配置
}

// ValObjCheck 验证内置agent可编辑字段配置
func (p *BuiltInCanEditFields) ValObjCheck() (err error) {
	// 目前所有字段都是布尔类型，无需特殊验证
	// 如果将来需要添加业务逻辑验证，可以在这里实现
	return
}
