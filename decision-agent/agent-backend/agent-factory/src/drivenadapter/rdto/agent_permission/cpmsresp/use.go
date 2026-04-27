package cpmsresp

// CheckRunResp 检查运行权限响应
type CheckRunResp struct {
	IsAllowed bool `json:"is_allowed"` // 是否有运行权限
}
