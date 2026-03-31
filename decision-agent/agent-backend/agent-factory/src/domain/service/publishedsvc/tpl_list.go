package publishedsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/publishedp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/pkg/errors"
)

// GetPubedTplList 获取已发布模板列表
func (svc *publishedSvc) GetPubedTplList(ctx context.Context, req *pubedreq.PubedTplListReq) (res *pubedresp.PublishedAgentTplListResp, err error) {
	res = pubedresp.NewPublishedAgentTplListResp()

	// 计算分页参数
	// offset := req.GetOffset()
	// limit := req.GetLimit()

	// 1. 从数据库获取已发布模板列表
	needSize := req.Size
	req.Size += 1

	var tplIDsByBd []string

	if !global.GConfig.IsBizDomainDisabled() {
		// 1.1 获取业务域ID
		bdID := chelper.GetBizDomainIDFromCtx(ctx)

		// 1.2 获取业务域ID对应的模板ID列表
		tplIDsByBd, err = svc.bizDomainHttp.GetAllAgentTplIDList(ctx, []string{bdID})
		if err != nil {
			err = errors.Wrapf(err, "[publishedSvc][GetPubTplList]: bizDomainHttp.GetAllAgentTplIDList failed")
			return
		}

		if len(tplIDsByBd) == 0 {
			res.IsLastPage = true
			return
		}

		req.TplIDsByBd = tplIDsByBd
	}

	// 1.3 从数据库获取已发布模板列表
	pos, err := svc.publishedTplRepo.GetPubTplList(ctx, req)
	if err != nil {
		err = errors.Wrapf(err, "[publishedSvc][GetPubTplList]: publishedTplRepo.GetPubTplList failed")
		return
	}

	if len(pos) == 0 {
		return
	}

	// 2. 如果pos长度大于needSize，说明还有下一页
	if len(pos) > needSize {
		pos = pos[:needSize]
		res.IsLastPage = false
	} else {
		res.IsLastPage = true
	}

	// 3. pos to eos
	eos, err := publishedp2e.PublishedTplListEos(ctx, pos, svc.umHttp)
	if err != nil {
		err = errors.Wrapf(err, "[publishedSvc][GetPubTplList]: convert published agent template list failed")
		return
	}

	// 4. 转换为响应格式

	err = res.LoadFromEos(eos)

	return
}
