package common

// PageByLastIntID 分页参数结构体
type PageByLastIntID struct {
	Size   int `form:"size,default=10" json:"size" binding:"numeric,max=1000"` // 每页显示数量
	LastID int `form:"last_id" json:"last_id" binding:"numeric"`               // 上一页的最后一条记录的ID
}

// GetErrMsgMap 返回错误信息映射
func (p PageByLastIntID) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Size.numeric":   `"size"的值必须是数字`,
		"Size.max":       `"size"的值不能大于1000`,
		"LastID.numeric": `"last_id"的值必须是数字`,
	}
}

// GetSize 获取每页大小
func (p PageByLastIntID) GetSize() int {
	if p.Size == 0 {
		p.Size = 10
	}

	return p.Size
}

// GetLimit 获取limit值，与size相同
func (p PageByLastIntID) GetLimit() int {
	return p.GetSize()
}
