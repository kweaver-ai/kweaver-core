package umarg

type GetAppListArgDto struct {
	// limit  1~1000(根据文档中来的)
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
	Keyword string `json:"keyword,omitempty"` // 搜索内容。即应用账户名。输入则返回该账户名模糊匹配的所有记录。不输入返回应用账户列表所有记录。（接口文档中的描述）
}
