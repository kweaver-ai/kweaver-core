package dapo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type ReleasePartPo struct {
	ReleaseID   string `json:"release_id" db:"f_id"`
	AgentConfig string `json:"agent_config" db:"f_agent_config"`

	PublishDesc string `json:"publish_desc" db:"f_agent_desc"`
	Version     string `json:"version" db:"f_agent_version"`
	PublishedAt int64  `json:"published_at" db:"f_update_time"`
	PublishedBy string `json:"published_by" db:"f_update_by"`
	// f_is_pms_ctrl
	IsPmsCtrl int `json:"is_pms_ctrl" db:"f_is_pms_ctrl"`

	PublishedToBeStruct
}

func (p *ReleasePartPo) IsPmsCtrlBool() (isPmsCtrl bool) {
	isPmsCtrl = p.IsPmsCtrl == 1
	return
}

type PublishedJoinPo struct {
	DataAgentPo
	ReleasePartPo
}

func (p *PublishedJoinPo) LoadFromReleasePartPo(po *ReleasePartPo) (err error) {
	err = cutil.JSON().Unmarshal([]byte(po.AgentConfig), &p.DataAgentPo)
	if err != nil {
		return
	}

	err = cutil.CopyStructUseJSON(&p.ReleasePartPo, po)

	return
}
