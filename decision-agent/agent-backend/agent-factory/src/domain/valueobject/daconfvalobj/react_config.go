package daconfvalobj

// ReactConfig ReAct 模式配置。
type ReactConfig struct {
	DisableHistoryInAConversation bool `json:"disable_history_in_a_conversation"` // 是否禁用单次会话的历史记录
	DisableLLMCache               bool `json:"disable_llm_cache"`                 // 是否禁用LLM缓存
}

// ValObjCheck 验证 ReAct 模式配置。
func (p *ReactConfig) ValObjCheck() (err error) {
	return
}
