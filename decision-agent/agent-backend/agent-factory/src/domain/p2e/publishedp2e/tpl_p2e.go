package publishedp2e

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/pkg/errors"
)

// PublishedTpl PO转EO
func PublishedTpl(ctx context.Context, _po *dapo.PublishedTplPo, productRepo idbaccess.IProductRepo) (eo *pubedeo.PublishedTpl, err error) {
	eo = &pubedeo.PublishedTpl{
		Config: &daconfvalobj.Config{},
	}

	err = cutil.CopyStructUseJSON(&eo.PublishedTplPo, _po)
	if err != nil {
		return
	}

	// 1. 解析配置
	if _po.Config != "" {
		err = cutil.JSON().UnmarshalFromString(_po.Config, &eo.Config)
		if err != nil {
			err = errors.Wrapf(err, "PublishedTpl unmarshal config error")
			return
		}
	}

	// 2. 获取产品名称
	if _po.ProductKey != "" {
		var po *dapo.ProductPo

		po, err = productRepo.GetByKey(ctx, _po.ProductKey)
		if err != nil {
			if chelper.IsSqlNotFound(err) {
				err = nil
			} else {
				err = errors.Wrapf(err, "get product name error")
				return
			}
		}

		eo.ProductName = po.Name
	}

	return
}
