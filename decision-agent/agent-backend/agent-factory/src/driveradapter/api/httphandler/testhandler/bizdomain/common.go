package bizdomain

// TestBizDomainHttpRequest 测试请求结构体
type TestBizDomainHttpRequest struct {
	AgentID string `json:"agent_id" validate:"required"` // 代理ID
}
