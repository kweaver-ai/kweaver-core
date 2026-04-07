package daconfvalobj

// 任务规划模式配置
type PlanMode struct {
	IsEnabled bool `json:"is_enabled"` // 是否启用
}

func NewPlanMode(isEnabled bool) *PlanMode {
	return &PlanMode{
		IsEnabled: isEnabled,
	}
}

// ValObjCheck 验证计划模式配置
func (p *PlanMode) ValObjCheck() (err error) {
	return
}
