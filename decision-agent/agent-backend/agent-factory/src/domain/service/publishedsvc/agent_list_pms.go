package publishedsvc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/published/pubedresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/pkg/errors"
)

// 获取有使用权限的已发布智能体
func (svc *publishedSvc) getPmsAgentPos(ctx context.Context, req *pubedreq.PubedAgentListReq) (filteredPos []*dapo.PublishedJoinPo, agentID2BdIDMap map[string]string, isLastPage bool, err error) {
	uid := chelper.GetUserIDFromCtx(ctx)

	var (
		needSize = req.Size
		okSize   = 0
	)

	if !global.GConfig.SwitchFields.DisablePmsCheck {
		if req.Size <= 1000 {
			req.Size = 1000
		}
	}

	var agentIdsByBdIds []string

	agentID2BdIDMap = make(map[string]string)
	if !global.GConfig.IsBizDomainDisabled() {
		// get all by 业务域
		agentIdsByBdIds, agentID2BdIDMap, err = svc.bizDomainHttp.GetAllAgentIDList(ctx, req.BusinessDomainIDs)
		if err != nil {
			err = errors.Wrapf(err, "[publishedSvc][GetPublishedAgentList]: get all agent id list failed")
			return
		}

		if len(agentIdsByBdIds) == 0 {
			isLastPage = true
			return
		}
	}

	for {
		// 1. 从数据库获取已发布智能体列表并通过业务域过滤
		var (
			pos []*dapo.PublishedJoinPo
		)

		pos, err = svc.getPos(ctx, req, agentIdsByBdIds)
		if err != nil {
			err = errors.Wrapf(err, "[publishedSvc][GetPublishedAgentList]: get pos failed")
			return
		}

		if len(pos) == 0 {
			isLastPage = true
			return
		}

		if len(pos) < req.Size {
			isLastPage = true
		}

		// 2. 如果禁用权限检查，直接返回
		if global.GConfig.SwitchFields.DisablePmsCheck {
			filteredPos = append(filteredPos, pos...)
			return
		}

		// 3. 过滤权限
		// 3.1 获取需要过滤的Agent ID
		var agentIds []string

		for _, po := range pos {
			if po.IsPmsCtrlBool() {
				agentIds = append(agentIds, po.ID)
			}
		}

		// 3.2 获取过滤后的、有权限的Agent ID
		var filteredAgentIdMap map[string]struct{}

		filteredAgentIdMap, err = svc.authZHttp.FilterCanUseAgentIDMap(ctx, uid, agentIds)
		if err != nil {
			err = errors.Wrapf(err, "[publishedSvc][GetPublishedAgentList]: filter can use agent ids failed")
			return
		}

		// 3.3 将有权限的Agent添加到filteredPos
		// okSize += len(filteredAgentIdMap)

		for _, po := range pos {
			if !po.IsPmsCtrlBool() {
				filteredPos = append(filteredPos, po)
				okSize++
			} else {
				if _, ok := filteredAgentIdMap[po.ID]; ok {
					filteredPos = append(filteredPos, po)
					okSize++
				}
			}
		}

		// 3.4 如果是最后一页，退出循环
		if isLastPage {
			break
		}

		// 3.5 如果过滤后的Agent数量小于需要的数量，继续循环
		if okSize < needSize {
			// req.Size = needSize - okSize
			// 增大查询数量，每次翻倍，最多10000
			req.Size *= 2

			if req.Size > 10000 {
				req.Size = 10000
			}

			if req.Marker == nil {
				req.Marker = pubedresp.NewPAListPaginationMarker()
			}

			req.Marker.LoadFromPos(pos)

			continue
		}

		break
	}

	// 3.6 如果是最后一页，且过滤后的Agent数量大于需要的数量，设置isLastPage为false
	// 这时表示下一页还有数据
	if isLastPage && okSize > needSize {
		isLastPage = false
	}

	// 4. 截取

	filteredPos = filteredPos[:cutil.MinInt(needSize, okSize)]

	return
}

func (svc *publishedSvc) getPos(ctx context.Context, req *pubedreq.PubedAgentListReq, agentIdsByBdIds []string) (pos []*dapo.PublishedJoinPo, err error) {
	pos, err = svc.pubedAgentRepo.GetPubedList(ctx, req)
	if err != nil {
		err = errors.Wrapf(err, "[publishedSvc][GetPublishedAgentList]: get published agent list failed")
		return
	}

	if agentIdsByBdIds == nil {
		return pos, nil
	}

	newPos := make([]*dapo.PublishedJoinPo, 0)

	for _, po := range pos {
		if cutil.ExistsGeneric(agentIdsByBdIds, po.ID) {
			newPos = append(newPos, po)
		}
	}

	pos = newPos

	return
}
