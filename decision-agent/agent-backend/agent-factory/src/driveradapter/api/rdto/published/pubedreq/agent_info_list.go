package pubedreq

import "github.com/pkg/errors"

type PAInfoListReq struct {
	AgentKeys []string `json:"agent_keys" binding:"required"` // 根据智能体标识查询

	NeedConfigFields []string `json:"need_config_fields"` // 需要返回的配置字段(config下的第一层子字段)
}

func NewPAInfoListReq() *PAInfoListReq {
	return &PAInfoListReq{}
}

func (p *PAInfoListReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"AgentKeys.required": "agent_keys is required",
	}
}

func (p *PAInfoListReq) ReqCheck() (err error) {
	// 1. agent_keys
	if len(p.AgentKeys) == 0 {
		err = errors.New("agent_keys length must be greater than 0")
		return
	}

	if len(p.AgentKeys) > 1000 {
		err = errors.New("agent_keys length max is 1000")
		return
	}

	// 2. need_config_fields
	if len(p.NeedConfigFields) > 0 {
		if len(p.NeedConfigFields) != 1 || p.NeedConfigFields[0] != "input" {
			err = errors.New("need_config_fields 目前只支持 [input]")
			return
		}
	}

	return
}

func (p *PAInfoListReq) HlDefaultVal() {
	if len(p.NeedConfigFields) == 0 {
		p.NeedConfigFields = []string{"input"}
	}
}
