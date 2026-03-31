package daconfvalobj

// 长期记忆配置（基于对话消息进行总结，基于记忆 Agent 可以做出更智能、个性化的反应）
type MemoryCfg struct {
	IsEnabled bool `json:"is_enabled"` // 是否启用
}

// ValObjCheck 验证内置agent可编辑字段配置
func (p *MemoryCfg) ValObjCheck() (err error) {
	// 目前所有字段都是布尔类型，无需特殊验证
	// 如果将来需要添加业务逻辑验证，可以在这里实现
	return
}
