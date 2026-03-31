package pubedresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

type DetailRes struct {
	ID int64 `json:"id"` // 发布ID（模板发布表ID）

	TplID int64 `json:"tpl_id"`

	Name    string  `json:"name"`    // 模板名称
	Profile *string `json:"profile"` // 模板简介
	Key     string  `json:"key"`     // 唯一标识

	Avatar     string `json:"avatar"`      // 头像信息
	AvatarType int    `json:"avatar_type"` // 头像类型

	ProductKey  string `json:"product_key"`  // 产品标识
	ProductName string `json:"product_name"` // 产品名称

	IsBuiltIn *int `json:"is_built_in"` // 是否内置

	Config *daconfvalobj.Config `json:"config"` // Agent配置（用于展示）

	PublishedAt int64  `json:"published_at"` // 发布时间
	PublishedBy string `json:"published_by"` // 发布者
}

func NewDetailRes() *DetailRes {
	return &DetailRes{}
}

func (d *DetailRes) LoadFromEo(eo *pubedeo.PublishedTpl) error {
	err := cutil.CopyStructUseJSON(d, eo)
	if err != nil {
		return errors.Wrap(err, "[DetailRes]: LoadFromEo failed")
	}

	return nil
}
