package authzhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/pkg/errors"
)

func (a *authZHttpAcc) ListPolicyAll(ctx context.Context, req *authzhttpreq.ListPolicyReq, userToken string) (res *authzhttpres.ListPolicyRes, err error) {
	res = &authzhttpres.ListPolicyRes{
		Entries:    make([]*authzhttpres.PolicyEntry, 0),
		TotalCount: 0,
	}

	const maxIterations = 5

	const pageSize = 1000

	currentReq := &authzhttpreq.ListPolicyReq{
		Limit:        pageSize,
		Offset:       0,
		ResourceID:   req.ResourceID,
		ResourceType: req.ResourceType,
	}

	for i := 0; i < maxIterations; i++ {
		pageRes, pageErr := a.ListPolicy(ctx, currentReq, userToken)
		if pageErr != nil {
			err = errors.Wrapf(pageErr, "第%d次查询策略列表失败", i+1)
			return
		}

		if pageRes == nil {
			err = errors.New("查询策略列表返回空结果")
			return
		}

		// 第一次查询时设置总数
		if i == 0 {
			res.TotalCount = pageRes.TotalCount
		}

		// 收集当前页的数据
		res.Entries = append(res.Entries, pageRes.Entries...)

		// 检查是否已获取完所有数据
		// 如果总数小于等于下一页的起始偏移量，说明已经获取完所有数据
		if pageRes.TotalCount <= currentReq.Offset+currentReq.Limit {
			// 所有数据已获取完毕
			return
		}

		// 准备下一页查询
		currentReq.Offset += pageSize
	}

	// 达到最大循环次数但仍未获取完所有数据
	err = errors.Errorf("已达到最大循环次数%d次，但仍未获取完所有数据，当前已获取%d条，总计%d条",
		maxIterations, len(res.Entries), res.TotalCount)

	return
}
