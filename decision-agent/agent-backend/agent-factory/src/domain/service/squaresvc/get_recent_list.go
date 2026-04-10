package squaresvc

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/daconstant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/p2e/daconfp2e"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/publishvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squareresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"

	"github.com/pkg/errors"
)

// GetRecentAgentList implements iv3portdriver.IMarketSvc.
func (svc *squareSvc) GetRecentAgentList(ctx context.Context, req squarereq.AgentSquareRecentAgentReq) (marketAgentList squareresp.RecentListAgentResp, err error) {
	marketAgentListEmpty := make([]squareresp.RecentAgentListItem, 0)

	rt, err := svc.releaseRepo.ListRecentAgentForMarket(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "list recent agent for square failed")
	}
	// 基于访问时间对结果进行排序
	slices.SortStableFunc(rt, func(a, b *dapo.RecentVisitAgentPO) int {
		switch {
		case a.LastVisitTime.Int64 > b.LastVisitTime.Int64:
			return -1
		case a.LastVisitTime.Int64 < b.LastVisitTime.Int64:
			return 1
		default:
			return 0
		}
	})

	// 进行逻辑分页

	// 3. 根据 page 和 size 进行分页
	page := req.Page
	size := req.Size
	start := (page - 1) * size
	end := start + size

	if start >= len(rt) {
		return marketAgentListEmpty, nil // 如果起始位置超出范围，返回空列表
	}

	if end > len(rt) {
		end = len(rt)
	}

	paginatedRt := rt[start:end]

	userIDS := make([]string, 0)

	for _, agentPO := range paginatedRt {
		if agentPO.PublishUserId.String != "" {
			userIDS = append(userIDS, agentPO.PublishUserId.String)
		}

		if agentPO.UpdatedBy != "" {
			userIDS = append(userIDS, agentPO.UpdatedBy)
		}
	}

	marketAgentList = make([]squareresp.RecentAgentListItem, len(paginatedRt))

	for i, po := range paginatedRt {
		agentCfgEO := &daconfeo.DataAgent{}
		if po.AgentVersion.String == daconstant.AgentVersionUnpublished {
			agentCfgEO, err = daconfp2e.DataAgent(ctx, &po.DataAgentPo)
			if err != nil {
				return marketAgentListEmpty, errors.Wrapf(err, "daconfp2e.DataAgent(&po.DataAgentPo)")
			}

			agentCfgEO.Status = cdaenum.StatusUnpublished
			marketAgentList[i] = squareresp.RecentAgentListItem{
				DataAgent:   *agentCfgEO,
				Version:     po.AgentVersion.String,
				Description: "",
				PublishedAt: 0,
				PublishedBy: "",
				PublishInfo: publishvo.NewListPublishInfo(),
			}

			err = cutil.CopyStructUseJSON(marketAgentList[i].PublishInfo, po.PublishedToBeStruct)
			if err != nil {
				return
			}
		} else {
			agentCfgPo := &dapo.DataAgentPo{}

			err = json.Unmarshal([]byte(po.AgentConfig.String), agentCfgPo)
			if err != nil {
				return marketAgentListEmpty, errors.Wrapf(err, "[GetRecentAgentList] json.Unmarshal error")
			}

			agentCfgEO, err = daconfp2e.DataAgent(ctx, agentCfgPo)
			if err != nil {
				return marketAgentListEmpty, errors.Wrapf(err, "[GetRecentAgentList] daconfp2e.DataAgent(&agentCfgPo)")
			}

			agentCfgEO.Status = cdaenum.StatusPublished
			marketAgentList[i] = squareresp.RecentAgentListItem{
				DataAgent:   *agentCfgEO,
				Version:     po.AgentVersion.String,
				Description: po.AgentDesc.String,
				PublishedAt: po.PublishTime.Int64,
				PublishedBy: po.PublishUserId.String,
				PublishInfo: publishvo.NewListPublishInfo(),
			}

			err = cutil.CopyStructUseJSON(marketAgentList[i].PublishInfo, po.PublishedToBeStruct)
			if err != nil {
				return
			}
		}

		marketAgentList[i].Config = nil
	}

	if len(userIDS) > 0 {
		// 获取用户信息
		userFields := []string{"name"}

		users, err := svc.usermanagementHttpClient.GetUserInfoByUserID(ctx, cutil.DeduplGeneric[string](userIDS), userFields)
		if err != nil {
			svc.Logger.Warnf("get user info failed, err: %v", err)
			return marketAgentList, nil
		}

		for i, marketAgent := range marketAgentList {
			if marketAgent.PublishedBy != "" {
				if user, ok := users[marketAgent.PublishedBy]; ok {
					marketAgentList[i].PublishedByName = user.Name
				}
			}

			if marketAgent.UpdatedBy != "" {
				if user, ok := users[marketAgent.UpdatedBy]; ok {
					marketAgentList[i].UpdatedByName = user.Name
				}
			}
		}
	}

	return marketAgentList, nil
}
