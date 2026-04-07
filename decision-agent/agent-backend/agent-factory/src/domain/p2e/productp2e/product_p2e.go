package productp2e

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/producteo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// Product 将持久化对象转换为产品实体对象
func Product(po *dapo.ProductPo) (eo *producteo.Product, err error) {
	if po == nil {
		return
	}

	eo = &producteo.Product{}
	err = cutil.CopyStructUseJSON(eo, po)

	return
}

// Products 批量PO转EO
func Products(ctx context.Context, _pos []*dapo.ProductPo) (eos []*producteo.Product, err error) {
	eos = make([]*producteo.Product, 0, len(_pos))

	for i := range _pos {
		var eo *producteo.Product

		if eo, err = Product(_pos[i]); err != nil {
			return
		}

		eos = append(eos, eo)
	}

	return
}
