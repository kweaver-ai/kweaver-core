package dapo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
)

type PublishedTplPo struct {
	ID   int64  `json:"id" db:"f_id"`
	Name string `json:"name" db:"f_name"`
	Key  string `json:"key" db:"f_key"`

	ProductKey string `json:"product_key" db:"f_product_key"`

	Profile *string `json:"profile" db:"f_profile"`

	AvatarType cdaenum.AvatarType `json:"avatar_type" db:"f_avatar_type"`
	Avatar     string             `json:"avatar" db:"f_avatar"`

	IsBuiltIn *cdaenum.BuiltIn `json:"is_built_in" db:"f_is_built_in"`

	Config string `json:"config" db:"f_config"`

	PublishedAt int64  `json:"published_at" db:"f_published_at"`
	PublishedBy string `json:"published_by" db:"f_published_by"`

	TplID int64 `json:"tpl_id" db:"f_tpl_id"`
}

func (p *PublishedTplPo) TableName() string {
	return "t_data_agent_config_tpl_published"
}

func (p *PublishedTplPo) SetIsBuiltIn(builtIn cdaenum.BuiltIn) {
	p.IsBuiltIn = &builtIn
}
