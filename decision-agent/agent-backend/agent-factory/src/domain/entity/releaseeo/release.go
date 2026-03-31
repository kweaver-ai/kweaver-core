package releaseeo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
)

type ReleaseEO struct {
	ID             string                  `json:"id"`
	AgentID        string                  `json:"agent_id"`
	UserID         string                  `json:"user_id"`
	AgentConfig    string                  `json:"agent_config"`
	AgentVersion   string                  `json:"agent_version"`
	AgentDesc      string                  `json:"agent_desc"`
	PublishToBes   []cdaenum.PublishToBe   `json:"publish_to_bes"`
	PublishToWhere []daenum.PublishToWhere `json:"publish_to_where"`
	IsPmsCtrl      int                     `json:"is_pms_ctrl"`
}

// ReleaseDAConfWrapperEO
// ReleaseDAConfWrapperEO: release agent config wrapper eo
type ReleaseDAConfWrapperEO struct {
	ReleaseEO
	Config *daconfvalobj.Config `json:"config"`
}

func (e *ReleaseEO) IsPmsCtrlBool() bool {
	return e.IsPmsCtrl == 1
}

func (e *ReleaseEO) SetIsPmsCtrl(isPmsCtrl bool) {
	e.IsPmsCtrl = 0

	if isPmsCtrl {
		e.IsPmsCtrl = 1
	}
}
