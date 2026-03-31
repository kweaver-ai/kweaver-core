package agenttplresp

type PublishInfoRes struct {
	Categories []CategoryInfo `json:"categories"` // 分类列表
}

type CategoryInfo struct {
	ID   string `json:"id"`   // 分类ID
	Name string `json:"name"` // 分类名称
}
