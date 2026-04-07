package umarg

type GetUserGroupListArgDto struct {
	// limit  1~1000(根据文档中来的)
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
	Keyword string `json:"keyword,omitempty"` // 搜索内容。即用户组名。若不传参数表示返回搜索范围下所有用户组。（接口文档中的描述）
}
