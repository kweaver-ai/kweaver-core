package agentexecutoraccreq

import "errors"

type AgentCacheActionType string

const (
	AgentCacheActionUpsert  AgentCacheActionType = "upsert"   // 更新或创建新的此Agent缓存
	AgentCacheActionGetInfo AgentCacheActionType = "get_info" // 获取此Agent缓存信息
)

func (c AgentCacheActionType) ToString() string {
	return string(c)
}

func (c AgentCacheActionType) EnumCheck() (err error) {
	switch c {
	case AgentCacheActionUpsert, AgentCacheActionGetInfo:
		return
	default:
		err = errors.New("Agent缓存操作类型不合法")
		return
	}
}
