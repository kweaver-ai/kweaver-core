package otherreq

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum/builtinagentenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
)

type DolphinTplListReq struct {
	Config *daconfvalobj.Config `json:"config" binding:"required"`

	BuiltInAgentKey builtinagentenum.AgentKey `json:"built_in_agent_key"`
}

func (p *DolphinTplListReq) GetErrMsgMap() map[string]string {
	return map[string]string{
		"Config.required": `"config"不能为空`,
	}
}

func (p *DolphinTplListReq) CustomCheck() (err error) {
	//	if p.BuiltInAgentKey != "" {
	//		if err = p.BuiltInAgentKey.EnumCheck(); err != nil {
	//			err = errors.Wrap(err, "[DolphinTplListReq]: built_in_agent_key is invalid")
	//			return
	//		}
	//	}
	return
}
