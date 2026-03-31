package agenttplresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type DetailRes struct {
	ID int64 `json:"id"` // 模板ID

	Name    string  `json:"name"`    // 模板名称
	Profile *string `json:"profile"` // 模板简介
	Key     string  `json:"key"`     // 唯一标识

	Avatar     string         `json:"avatar"`      // 头像信息
	AvatarType int            `json:"avatar_type"` // 头像类型
	Status     cdaenum.Status `json:"status"`      // 状态

	ProductKey  string `json:"product_key"`  // 产品标识
	ProductName string `json:"product_name"` // 产品名称

	IsBuiltIn *int `json:"is_built_in"` // 是否内置

	// IsSystemAgent *cenum.YesNoInt8 `json:"is_system_agent"` // 是否是系统agent

	Config *daconfvalobj.Config `json:"config"` // Agent配置（用于展示）

	CreatedAt int64  `json:"created_at"` // 创建时间
	UpdatedAt int64  `json:"updated_at"` // 更新时间
	CreatedBy string `json:"created_by"` // 创建者
	UpdatedBy string `json:"updated_by"` // 更新者

	PublishedAt int64  `json:"published_at"` // 发布时间
	PublishedBy string `json:"published_by"` // 发布者
}

func NewDetailRes() *DetailRes {
	return &DetailRes{}
}

func (d *DetailRes) LoadFromEo(eo *daconfeo.DataAgentTpl) error {
	err := cutil.CopyStructUseJSON(d, eo)
	if err != nil {
		return errors.Wrap(err, "[DetailRes]: LoadFromEo failed")
	}

	return nil
}
