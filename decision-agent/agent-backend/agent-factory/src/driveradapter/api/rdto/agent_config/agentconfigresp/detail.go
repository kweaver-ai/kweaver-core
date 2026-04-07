package agentconfigresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type DetailRes struct {
	ID            string               `json:"id"`              // agent id
	Key           string               `json:"key"`             // agent 标识
	IsBuiltIn     cdaenum.BuiltIn      `json:"is_built_in"`     // 是否内置
	IsSystemAgent *cenum.YesNoInt8     `json:"is_system_agent"` // 是否是系统agent
	Name          string               `json:"name"`            // 名字
	Profile       string               `json:"profile"`         // 描述
	AvatarType    cdaenum.AvatarType   `json:"avatar_type"`     // 头像类型
	Avatar        string               `json:"avatar"`          // 头像信息
	ProductKey    string               `json:"product_key"`     // 所属产品标识
	ProductName   string               `json:"product_name"`    // 所属产品名称
	Config        *daconfvalobj.Config `json:"config"`          // agent配置
	Status        cdaenum.Status       `json:"status"`          // 状态
	IsPublished   bool                 `json:"is_published"`    // 是否发布过
}

func NewDetailRes() *DetailRes {
	return &DetailRes{}
}

func (d *DetailRes) LoadFromEo(eo *daconfeo.DataAgent) error {
	err := cutil.CopyStructUseJSON(d, eo)
	if err != nil {
		return errors.Wrap(err, "[DetailRes]: LoadFromEo failed")
	}

	d.Config = eo.Config

	return nil
}
