package productresp

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/producteo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type ListItem struct {
	ID        int64  `json:"id"`         // 产品ID
	Name      string `json:"name"`       // 产品名称
	Key       string `json:"key"`        // 产品标识
	Profile   string `json:"profile"`    // 产品简介
	CreatedAt int64  `json:"created_at"` // 创建时间（时间戳，单位：ms）
	UpdatedAt int64  `json:"updated_at"` // 更新时间（时间戳，单位：ms）
}

type ListRes struct {
	Entries []*ListItem `json:"entries"` // 产品列表
	Total   int         `json:"total"`   // 产品总数
}

func NewListRes() *ListRes {
	return &ListRes{
		Entries: make([]*ListItem, 0),
	}
}

func (l *ListRes) LoadFromEo(eos []*producteo.Product) (err error) {
	l.Entries = make([]*ListItem, 0, len(eos))

	for _, eo := range eos {
		item := &ListItem{}

		err = cutil.CopyStructUseJSON(item, eo)
		if err != nil {
			return
		}

		l.Entries = append(l.Entries, item)
	}

	return
}
