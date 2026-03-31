package publishedsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/publishedp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

// PubedTplDetail 获取已发布模板详情
func (svc *publishedSvc) PubedTplDetail(ctx context.Context, tplID int64) (res *pubedresp.DetailRes, err error) {
	// 1. 从数据库获取模板
	po, err := svc.publishedTplRepo.GetByTplID(ctx, tplID)
	if err != nil {
		// 检查是否是记录不存在的错误
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.PublishedTplNotFound, "此已发布的Agent模板不存在")
		}

		return
	}

	// 2. PO转EO
	eo, err := publishedp2e.PublishedTpl(ctx, po, svc.productRepo)
	if err != nil {
		err = errors.Wrapf(err, "convert po to eo")
		return
	}

	// 3. 构建响应DTO
	res = pubedresp.NewDetailRes()

	err = res.LoadFromEo(eo)
	if err != nil {
		err = errors.Wrapf(err, "load from eo")
		return
	}

	return
}
