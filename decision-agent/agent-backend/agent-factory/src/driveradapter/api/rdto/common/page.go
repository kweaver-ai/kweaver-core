package common

// PageSize 分页参数结构体
type PageSize struct {
	Size int `form:"size,default=10" json:"size" binding:"numeric,max=1000"` // 每页显示数量
	Page int `form:"page,default=1" json:"page" binding:"numeric,min=1"`     // 页码，从1开始
}

// GetErrMsgMap 返回错误信息映射
func (p PageSize) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Size.numeric": `"size"的值必须是数字`,
		"Size.max":     `"size"的值不能大于1000`,
		"Page.numeric": `"page"的值必须是数字`,
		"Page.min":     `"page"的值不能小于1`,
	}
}

// GetSize 获取每页大小
func (p PageSize) GetSize() int {
	if p.Size == 0 {
		p.Size = 10
	}

	return p.Size
}

// GetPage 获取页码
func (p PageSize) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}

	return p.Page
}

// GetOffset 根据page和size计算offset
func (p PageSize) GetOffset() int {
	return (p.GetPage() - 1) * p.GetSize()
}

// GetLimit 获取limit值，与size相同
func (p PageSize) GetLimit() int {
	return p.GetSize()
}

// ToLimitOffset 转换为LimitOffset结构体
func (p PageSize) ToLimitOffset() LimitOffset {
	return LimitOffset{
		Limit:  p.GetLimit(),
		Offset: p.GetOffset(),
	}
}
