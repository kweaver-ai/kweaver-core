package dapo

import "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"

type ReleasePermissionPO struct {
	ID         int64                  `db:"f_id"`
	ReleaseId  string                 `db:"f_release_id"`
	ObjectId   string                 `db:"f_obj_id"`
	ObjectType cenum.PmsTargetObjType `db:"f_obj_type"`
}

func (po *ReleasePermissionPO) TableName() string {
	return "t_data_agent_release_permission"
}
