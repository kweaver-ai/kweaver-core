package tplp2e

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/pkg/errors"
)

// DataAgentTpl PO转EO
func DataAgentTpl(ctx context.Context, _po *dapo.DataAgentTplPo, productRepo idbaccess.IProductRepo) (eo *daconfeo.DataAgentTpl, err error) {
	eo = &daconfeo.DataAgentTpl{
		Config: &daconfvalobj.Config{},
	}

	err = cutil.CopyStructUseJSON(&eo.DataAgentTplPo, _po)
	if err != nil {
		return
	}

	// 1. 解析配置
	if _po.Config != "" {
		err = cutil.JSON().UnmarshalFromString(_po.Config, &eo.Config)
		if err != nil {
			err = errors.Wrapf(err, "DataAgentTpl unmarshal config error")
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
