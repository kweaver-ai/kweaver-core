package agentconfigresp

// CopyResp 复制Agent响应
type CopyResp struct {
	ID      string `json:"id"`      // 新Agent的ID
	Name    string `json:"name"`    // 新Agent的名称
	Key     string `json:"key"`     // 新Agent的标识
	Version string `json:"version"` // 新Agent的版本（通常是"unpublished"）
}

// Copy2TplResp 复制Agent为模板响应
type Copy2TplResp struct {
	ID   int64  `json:"id"`   // 新模板的ID
	Name string `json:"name"` // 新模板的名称
	Key  string `json:"key"`  // 新模板的标识
}
