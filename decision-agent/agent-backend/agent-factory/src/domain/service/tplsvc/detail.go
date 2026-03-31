package tplsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/p2e/tplp2e"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

func (s *dataAgentTplSvc) Detail(ctx context.Context, id int64) (res *agenttplresp.DetailRes, err error) {
	// 1. 从数据库获取模板
	po, err := s.agentTplRepo.GetByID(ctx, id)
	if err != nil {
		// 检查是否是记录不存在的错误
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.AgentTplNotFound, "模板不存在")
		}

		return
	}

	// 2. PO转EO
	eo, err := tplp2e.DataAgentTpl(ctx, po, s.productRepo)
	if err != nil {
		err = errors.Wrapf(err, "convert po to eo")
		return
	}

	// 3. 构建响应DTO
	res = agenttplresp.NewDetailRes()

	err = res.LoadFromEo(eo)
	if err != nil {
		err = errors.Wrapf(err, "load from eo")
		return
	}

	return
}

func (s *dataAgentTplSvc) DetailByKey(ctx context.Context, key string) (res *agenttplresp.DetailRes, err error) {
	// 1. 从数据库获取模板
	po, err := s.agentTplRepo.GetByKey(ctx, key)
	if err != nil {
		err = errors.Wrapf(err, "get template by key %s", key)
		return
	}

	// 2. PO转EO
	eo, err := tplp2e.DataAgentTpl(ctx, po, s.productRepo)
	if err != nil {
		err = errors.Wrapf(err, "convert po to eo")
		return
	}

	// 3. 构建响应DTO
	res = agenttplresp.NewDetailRes()

	err = res.LoadFromEo(eo)
	if err != nil {
		err = errors.Wrapf(err, "load from eo")
		return
	}

	return
}
