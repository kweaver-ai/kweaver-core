package psdbarg

import "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/personal_space/personalspacereq"

type TplListArg struct {
	ListReq   *personalspacereq.AgentTplListReq
	CreatedBy string

	TplIDsByBd []string `json:"-"` // 根据业务域ID获取到的模板ID列表
}

func NewTplListArg(listReq *personalspacereq.AgentTplListReq, createdBy string, tplIDsByBd []string) *TplListArg {
	return &TplListArg{
		ListReq:    listReq,
		CreatedBy:  createdBy,
		TplIDsByBd: tplIDsByBd,
	}
}
