package agentexecutoraccres

type AgentCacheManageResp struct {
	CacheID   string                 `json:"cache_id"`
	TTL       int                    `json:"ttl"`
	CreatedAt string                 `json:"created_at"`
	CacheData map[string]interface{} `json:"cache_data"`
}
