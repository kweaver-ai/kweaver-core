package tplsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

func (s *dataAgentTplSvc) GetPublishInfo(ctx context.Context, id int64) (res *agenttplresp.PublishInfoRes, err error) {
	// 1. 从数据库获取模板
	pubedPo, err := s.publishedTplRepo.GetByTplID(ctx, id)
	if err != nil {
		if chelper.IsSqlNotFound(err) {
			err = capierr.NewCustom404Err(ctx, apierr.PublishedTplNotFound, "此已发布模板不存在")
			return
		}

		err = errors.Wrapf(err, "get template by id %d", id)

		return
	}

	publishedTplID := pubedPo.ID

	// if po.Status != cdaenum.StatusPublished {
	//	err = capierr.NewCustom409Err(ctx, apierr.AgentTplIsUnpublished, "模板未发布")
	//	return
	//}

	// 2. 获取分类关联信息
	pos, err := s.publishedTplRepo.GetCategoryJoinPosByTplID(ctx, nil, publishedTplID)
	if err != nil {
		err = errors.Wrapf(err, "get category join pos by tpl id %d", publishedTplID)
		return
	}

	// 3. 转换为响应格式
	categories := make([]agenttplresp.CategoryInfo, 0, len(pos))

	for _, _po := range pos {
		categories = append(categories, agenttplresp.CategoryInfo{
			ID:   _po.CategoryID,
			Name: _po.CategoryName,
		})
	}

	res = &agenttplresp.PublishInfoRes{
		Categories: categories,
	}

	return
}
