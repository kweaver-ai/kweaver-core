package daconfvalobj

// 相关问题配置
type RelatedQuestion struct {
	IsEnabled bool `json:"is_enabled"` // 是否启用
}

// ValObjCheck 验证相关问题配置
func (p *RelatedQuestion) ValObjCheck() (err error) {
	return
}
