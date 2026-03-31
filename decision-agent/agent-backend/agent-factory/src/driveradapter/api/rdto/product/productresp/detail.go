package productresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/producteo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type DetailRes struct {
	ID        int64  `json:"id"`         // 产品ID
	Name      string `json:"name"`       // 产品名称
	Key       string `json:"key"`        // 产品标识
	Profile   string `json:"profile"`    // 产品简介
	CreatedAt int64  `json:"created_at"` // 创建时间（时间戳，单位：ms）
	UpdatedAt int64  `json:"updated_at"` // 更新时间（时间戳，单位：ms）
}

func NewDetailRes() *DetailRes {
	return &DetailRes{}
}

func (d *DetailRes) LoadFromEo(eo *producteo.Product) (err error) {
	err = cutil.CopyStructUseJSON(d, eo)
	if err != nil {
		return
	}

	return
}
