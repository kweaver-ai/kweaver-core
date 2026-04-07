package dapo

import (
	"database/sql"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
)

type PublishedToBeStruct struct {
	IsAPIAgent      int `json:"is_api_agent" db:"f_is_api_agent"`
	IsWebSDKAgent   int `json:"is_web_sdk_agent" db:"f_is_web_sdk_agent"`
	IsSkillAgent    int `json:"is_skill_agent" db:"f_is_skill_agent"`
	IsDataFlowAgent int `json:"is_data_flow_agent" db:"f_is_data_flow_agent"`
}

func (toBe *PublishedToBeStruct) SelectFieldsZero() (str string) {
	str = "0 as f_is_api_agent, 0 as f_is_web_sdk_agent, 0 as f_is_skill_agent, 0 as f_is_data_flow_agent"

	return
}

func (toBe *PublishedToBeStruct) LoadFromReleasePo(po *ReleasePO) {
	if po == nil {
		return
	}

	if po.IsAPIAgent != nil {
		toBe.IsAPIAgent = *po.IsAPIAgent
	}

	if po.IsWebSDKAgent != nil {
		toBe.IsWebSDKAgent = *po.IsWebSDKAgent
	}

	if po.IsSkillAgent != nil {
		toBe.IsSkillAgent = *po.IsSkillAgent
	}

	if po.IsDataFlowAgent != nil {
		toBe.IsDataFlowAgent = *po.IsDataFlowAgent
	}
}

type ReleasePO struct {
	ID           string `json:"id" db:"f_id"`
	AgentID      string `json:"agent_id" db:"f_agent_id"`
	AgentName    string `json:"agent_name" db:"f_agent_name"`
	AgentConfig  string `json:"agent_config" db:"f_agent_config"`
	AgentVersion string `json:"agent_version" db:"f_agent_version"`
	AgentDesc    string `json:"agent_desc" db:"f_agent_desc"`

	IsAPIAgent      *int `json:"is_api_agent" db:"f_is_api_agent"`
	IsWebSDKAgent   *int `json:"is_web_sdk_agent" db:"f_is_web_sdk_agent"`
	IsSkillAgent    *int `json:"is_skill_agent" db:"f_is_skill_agent"`
	IsDataFlowAgent *int `json:"is_data_flow_agent" db:"f_is_data_flow_agent"`

	IsToCustomSpace *int `json:"is_to_custom_space" db:"f_is_to_custom_space"`
	IsToSquare      *int `json:"is_to_square" db:"f_is_to_square"`

	IsPmsCtrl *int `json:"is_pms_ctrl" db:"f_is_pms_ctrl"`

	CreateTime int64  `json:"create_time" db:"f_create_time"`
	UpdateTime int64  `json:"update_time" db:"f_update_time"`
	CreateBy   string `json:"create_by" db:"f_create_by"`
	UpdateBy   string `json:"update_by" db:"f_update_by"`
}

// --- to be start ---
func (r *ReleasePO) IsAPIAgentBool() bool {
	if r.IsAPIAgent == nil {
		return false
	}

	return *r.IsAPIAgent == 1
}

func (r *ReleasePO) IsWebSDKAgentBool() bool {
	if r.IsWebSDKAgent == nil {
		return false
	}

	return *r.IsWebSDKAgent == 1
}

func (r *ReleasePO) IsSkillAgentBool() bool {
	if r.IsSkillAgent == nil {
		return false
	}

	return *r.IsSkillAgent == 1
}

func (r *ReleasePO) IsDataFlowAgentBool() bool {
	if r.IsDataFlowAgent == nil {
		return false
	}

	return *r.IsDataFlowAgent == 1
}

func (r *ReleasePO) ResetPublishToBes() {
	r.IsAPIAgent = new(int)
	*r.IsAPIAgent = 0
	r.IsWebSDKAgent = new(int)
	*r.IsWebSDKAgent = 0
	r.IsSkillAgent = new(int)
	*r.IsSkillAgent = 0
	r.IsDataFlowAgent = new(int)
	*r.IsDataFlowAgent = 0
}

func (r *ReleasePO) SetPublishToBes(toBes []cdaenum.PublishToBe) {
	r.ResetPublishToBes()

	for _, tobe := range toBes {
		switch tobe {
		case cdaenum.PublishToBeAPIAgent:
			r.IsAPIAgent = new(int)
			*r.IsAPIAgent = 1
		case cdaenum.PublishToBeWebSDKAgent:
			r.IsWebSDKAgent = new(int)
			*r.IsWebSDKAgent = 1
		case cdaenum.PublishToBeSkillAgent:
			r.IsSkillAgent = new(int)
			*r.IsSkillAgent = 1
		case cdaenum.PublishToBeDataFlowAgent:
			r.IsDataFlowAgent = new(int)
			*r.IsDataFlowAgent = 1
		}
	}
}

// --- to be end ---

// --- to where start ---
func (r *ReleasePO) IsToCustomSpaceBool() bool {
	if r.IsToCustomSpace == nil {
		return false
	}

	return *r.IsToCustomSpace == 1
}

func (r *ReleasePO) IsToSquareBool() bool {
	if r.IsToSquare == nil {
		return false
	}

	return *r.IsToSquare == 1
}

func (r *ReleasePO) ResetPublishToWhere() {
	r.IsToCustomSpace = new(int)
	*r.IsToCustomSpace = 0
	r.IsToSquare = new(int)
	*r.IsToSquare = 0
}

func (r *ReleasePO) SetPublishToWhere(tos []daenum.PublishToWhere) {
	r.ResetPublishToWhere()

	for _, to := range tos {
		switch to {
		case daenum.PublishToWhereCustomSpace:
			r.IsToCustomSpace = new(int)
			*r.IsToCustomSpace = 1
		case daenum.PublishToWhereSquare:
			r.IsToSquare = new(int)
			*r.IsToSquare = 1
		}
	}
}

// --- to where end ---

// --- pms ctrl start ---

func (r *ReleasePO) IsPmsCtrlBool() bool {
	if r.IsPmsCtrl == nil {
		return false
	}

	return *r.IsPmsCtrl == 1
}

func (r *ReleasePO) ResetIsPmsCtrl() {
	r.IsPmsCtrl = new(int)
	*r.IsPmsCtrl = 0
}

func (r *ReleasePO) SetIsPmsCtrl(isPmsCtrl bool) {
	r.ResetIsPmsCtrl()

	if isPmsCtrl {
		*r.IsPmsCtrl = 1
	}
}

// --- pms ctrl end ---

type ReleaseAgentPO struct {
	DataAgentPo
	AgentConfig   sql.NullString `json:"agent_config" db:"f_agent_config"`
	AgentDesc     sql.NullString `json:"agent_desc" db:"f_agent_desc"`
	AgentVersion  sql.NullString `json:"agent_version" db:"f_agent_version"`
	PublishTime   sql.NullInt64  `json:"publish_time" db:"publish_time"`
	PublishUserId sql.NullString `json:"publish_user_id" db:"publish_user_id"`
}

type RecentVisitAgentPO struct {
	ReleaseAgentPO
	LastVisitTime sql.NullInt64 `json:"last_visit_time" db:"last_visit_time"`
	PublishedToBeStruct
}

func (r *ReleasePO) TableName() string {
	return "t_data_agent_release"
}
