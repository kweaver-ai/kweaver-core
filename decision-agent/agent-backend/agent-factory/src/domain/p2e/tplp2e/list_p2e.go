package tplp2e

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/locale"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/daconfeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc"
)

// AgentTplListEo PO转EO
func AgentTplListEo(ctx context.Context, _po *dapo.DataAgentTplPo) (eo *daconfeo.DataAgentTplListEo, err error) {
	eo = &daconfeo.DataAgentTplListEo{}

	err = cutil.CopyStructUseJSON(&eo.DataAgentTplPo, _po)
	if err != nil {
		return
	}

	return
}

// AgentTpls 批量PO转EO
func AgentTplListEos(ctx context.Context, _pos []*dapo.DataAgentTplPo, umHttp iumacc.UmHttpAcc) (eos []*daconfeo.DataAgentTplListEo, err error) {
	eos = make([]*daconfeo.DataAgentTplListEo, 0, len(_pos))

	userIDs := make([]string, 0, len(_pos))
	for i := range _pos {
		userIDs = append(userIDs, _pos[i].CreatedBy)
		userIDs = append(userIDs, _pos[i].UpdatedBy)
		userIDs = append(userIDs, _pos[i].GetPublishedByString())
	}

	arg := &umarg.GetOsnArgDto{
		UserIDs: userIDs,
	}

	ret := umtypes.NewOsnInfoMapS()

	if cenvhelper.IsLocalDev() {
		// 本地开发环境模拟数据
		for _, userID := range arg.UserIDs {
			ret.UserNameMap[userID] = userID + "_name"
		}
	} else {
		if umHttp == nil {
			panic("umHttp cannot be nil in non-local dev environment")
		}

		ret, err = umHttp.GetOsnNames(ctx, arg)
		if err != nil {
			return
		}
	}

	unknownUserName := locale.GetI18nByCtx(ctx, locale.UnknownUser)

	for i := range _pos {
		var eo *daconfeo.DataAgentTplListEo

		if eo, err = AgentTplListEo(ctx, _pos[i]); err != nil {
			return
		}

		if eo.CreatedBy != "" {
			userName, ok := ret.UserNameMap[eo.CreatedBy]
			if !ok {
				userName = unknownUserName
			}

			eo.CreatedByName = userName
		}

		if eo.UpdatedBy != "" {
			userName, ok := ret.UserNameMap[eo.UpdatedBy]
			if !ok {
				userName = unknownUserName
			}

			eo.UpdatedByName = userName
		}

		if eo.GetPublishedByString() != "" {
			userName, ok := ret.UserNameMap[eo.GetPublishedByString()]
			if !ok {
				userName = unknownUserName
			}

			eo.PublishedByName = userName
		}

		eos = append(eos, eo)
	}

	return
}
