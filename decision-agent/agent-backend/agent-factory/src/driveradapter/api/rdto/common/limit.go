package common

type LimitOffset struct {
	Limit  int `form:"limit,default=10" json:"limit" binding:"numeric,max=1000"` // 每页显示数量
	Offset int `form:"offset" json:"offset" binding:"numeric"`                   // 分页偏移数量
}

func (l LimitOffset) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Limit.numeric": `"limit"的值必须是数字`,
		"Limit.max":     `"limit"的值不能大于1000`,
	}
}

func (l LimitOffset) GetLimit() int {
	if l.Limit == 0 {
		l.Limit = 10
	}

	return l.Limit
}

func (l LimitOffset) GetOffset() int {
	return l.Offset
}
