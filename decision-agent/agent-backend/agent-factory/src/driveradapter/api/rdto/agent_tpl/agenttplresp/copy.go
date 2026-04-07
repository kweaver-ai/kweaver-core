package agenttplresp

// CopyResp 复制Agent模板响应
type CopyResp struct {
	ID   int64  `json:"id"`   // 新Agent模板的ID
	Name string `json:"name"` // 新Agent模板的名称
	Key  string `json:"key"`  // 新Agent模板的标识
}
