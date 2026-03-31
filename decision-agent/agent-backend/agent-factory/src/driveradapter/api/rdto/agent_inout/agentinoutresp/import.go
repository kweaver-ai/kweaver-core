package agentinoutresp

type ImportFailItem struct {
	AgentKey  string `json:"agent_key"`
	AgentName string `json:"agent_name"`
}

type ImportResp struct {
	IsSuccess bool `json:"is_success"`

	ConfigInvalid []*ImportFailItem `json:"config_invalid"` // 配置无效

	NoCreateSystemAgentPms []*ImportFailItem `json:"no_create_system_agent_pms"` // 没有创建系统agent的权限

	AgentKeyConflict []*ImportFailItem `json:"agent_key_conflict"`

	// 业务域冲突
	BizDomainConflict []*ImportFailItem `json:"biz_domain_conflict"`
}

func NewImportResp() *ImportResp {
	return &ImportResp{
		ConfigInvalid:          make([]*ImportFailItem, 0),
		NoCreateSystemAgentPms: make([]*ImportFailItem, 0),
		AgentKeyConflict:       make([]*ImportFailItem, 0),
		BizDomainConflict:      make([]*ImportFailItem, 0),
	}
}

func (r *ImportResp) HasFail() (has bool) {
	has = len(r.ConfigInvalid) > 0 || len(r.NoCreateSystemAgentPms) > 0 || len(r.AgentKeyConflict) > 0 || len(r.BizDomainConflict) > 0
	return
}

func (r *ImportResp) AddConfigInvalid(agentKey string, agentName string) {
	r.ConfigInvalid = append(r.ConfigInvalid, &ImportFailItem{
		AgentKey:  agentKey,
		AgentName: agentName,
	})
}

func (r *ImportResp) AddNoCreateSystemAgentPms(agentKey string, agentName string) {
	r.NoCreateSystemAgentPms = append(r.NoCreateSystemAgentPms, &ImportFailItem{
		AgentKey:  agentKey,
		AgentName: agentName,
	})
}

func (r *ImportResp) AddAgentKeyConflict(agentKey string, agentName string) {
	r.AgentKeyConflict = append(r.AgentKeyConflict, &ImportFailItem{
		AgentKey:  agentKey,
		AgentName: agentName,
	})
}

func (r *ImportResp) AddBizDomainConflict(agentKey string, agentName string) {
	r.BizDomainConflict = append(r.BizDomainConflict, &ImportFailItem{
		AgentKey:  agentKey,
		AgentName: agentName,
	})
}
