package bizdomainhttpres

// ResourceAssociationItem 资源关联项
type ResourceAssociationItem struct {
	BdID     string `json:"bd_id"`     // 业务域ID
	ID       string `json:"id"`        // 资源ID
	Type     string `json:"type"`      // 资源类型
	CreateBy string `json:"create_by"` // 创建者
}

// QueryResourceAssociationsRes 关联关系查询响应
type QueryResourceAssociationsRes struct {
	Limit  int                        `json:"limit"`  // 限制数量
	Offset int                        `json:"offset"` // 偏移量
	Total  int                        `json:"total"`  // 总数
	Items  []*ResourceAssociationItem `json:"items"`  // 关联项列表
}

func (e *QueryResourceAssociationsRes) GetItemIDs() (ids []string) {
	ids = make([]string, 0, len(e.Items))
	for _, item := range e.Items {
		ids = append(ids, item.ID)
	}

	return
}
