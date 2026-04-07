package agentconfigresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// BatchFieldsResp 批量获取agent指定字段的响应
type BatchFieldsResp struct {
	Entries []*AgentFieldsItem `json:"entries"`
}

// BatchFieldsRespField 返回的字段
// 目前有：
// - name：agent名称
type BatchFieldsRespField struct {
	Name string `json:"name"` // agent 名称
}

func NewBatchFieldsRespField() *BatchFieldsRespField {
	return &BatchFieldsRespField{}
}

// AgentFieldsItem agent字段项
type AgentFieldsItem struct {
	AgentID string                `json:"agent_id" description:"agent id"`
	Field   *BatchFieldsRespField `json:"field"`
}

// NewBatchFieldsResp 创建批量获取agent字段的响应
func NewBatchFieldsResp() *BatchFieldsResp {
	return &BatchFieldsResp{
		Entries: make([]*AgentFieldsItem, 0),
	}
}

// LoadFromAgentPOs 从数据持久化对象加载数据
func (resp *BatchFieldsResp) LoadFromAgentPOs(pos []*dapo.DataAgentPo, requestedFields []agentconfigreq.BatchFieldsReqField) error {
	resp.Entries = make([]*AgentFieldsItem, 0, len(pos))

	for _, po := range pos {
		item := &AgentFieldsItem{
			AgentID: po.ID,
			Field:   NewBatchFieldsRespField(),
		}

		// 根据请求的字段填充数据
		for _, field := range requestedFields {
			switch field {
			case agentconfigreq.BatchFieldsReqFieldName:
				item.Field.Name = po.Name
			}
		}

		resp.Entries = append(resp.Entries, item)
	}

	return nil
}
