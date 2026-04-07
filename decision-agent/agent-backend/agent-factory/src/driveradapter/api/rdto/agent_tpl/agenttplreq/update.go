package agenttplreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/customvalidator"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type UpdateReq struct {
	Name       string               `json:"name" binding:"required,checkAgentAndTplName,max=50"` // 模板名称
	Profile    *string              `json:"profile" binding:"max=500"`                           // 模板简介
	Avatar     string               `json:"avatar"`                                              // 头像信息
	AvatarType int                  `json:"avatar_type"`                                         // 头像类型
	Config     *daconfvalobj.Config `json:"config" binding:"required"`                           // Agent配置
}

func (p *UpdateReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Name.required":             `"name"不能为空`,
		"Name.checkAgentAndTplName": customvalidator.GenAgentAndTplNameErrMsg(`"模板名称"`),
		"Name.max":                  `"name"长度不能超过50`,
		"Config.required":           `"config"不能为空`,
		"Profile.max":               `"profile"长度不能超过500`,
	}
}

func (p *UpdateReq) D2e() (eo *daconfeo.DataAgentTpl, err error) {
	// 1. 生成allowed_file_types
	err = agentconfigreq.HandleConfig(p.Config)
	if err != nil {
		err = errors.Wrap(err, "[UpdateReq]: d2eCommon failed")
		return
	}

	// 2 . dto to eo
	eo = &daconfeo.DataAgentTpl{}

	err = cutil.CopyStructUseJSON(eo, p)
	if err != nil {
		return
	}

	// 3. d2e后处理
	// agentconfigreq.D2eCommonAfterD2e(eo)

	return
}
