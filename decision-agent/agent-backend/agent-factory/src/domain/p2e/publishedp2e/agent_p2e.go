package publishedp2e

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/locale"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc"
	"github.com/pkg/errors"
)

func PublishedAgents(ctx context.Context, _pos []*dapo.PublishedJoinPo, umHttp iumacc.UmHttpAcc, isUnmarshalConfig bool) (eos []*pubedeo.PublishedAgentEo, err error) {
	eos = make([]*pubedeo.PublishedAgentEo, 0, len(_pos))

	userIDs := make([]string, 0, len(_pos))
	for i := range _pos {
		// userIDs = append(userIDs, _pos[i].CreatedBy)
		// userIDs = append(userIDs, _pos[i].UpdatedBy)
		userIDs = append(userIDs, _pos[i].ReleasePartPo.PublishedBy)
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
		ret, err = umHttp.GetOsnNames(ctx, arg)
		if err != nil {
			return
		}
	}

	unknownUserName := locale.GetI18nByCtx(ctx, locale.UnknownUser)

	for i := range _pos {
		var eo *pubedeo.PublishedAgentEo

		if eo, err = PublishedAgent(ctx, _pos[i], isUnmarshalConfig); err != nil {
			return
		}

		//// 设置用户名称
		// if eo.CreatedBy != "" {
		//	eo.CreatedByName = ret.UserNameMap[eo.CreatedBy]
		//}
		//
		//if eo.UpdatedBy != "" {
		//	eo.UpdatedByName = ret.UserNameMap[eo.UpdatedBy]
		//}

		if eo.ReleasePartPo.PublishedBy != "" {
			userName, ok := ret.UserNameMap[eo.ReleasePartPo.PublishedBy]
			if !ok {
				userName = unknownUserName
			}

			eo.PublishedByName = userName
		}

		eos = append(eos, eo)
	}

	return
}

func PublishedAgent(ctx context.Context, _po *dapo.PublishedJoinPo, isUnmarshalConfig bool) (eo *pubedeo.PublishedAgentEo, err error) {
	eo = &pubedeo.PublishedAgentEo{}

	err = cutil.CopyStructUseJSON(&eo.PublishedJoinPo, _po)
	if err != nil {
		return
	}

	// 1. 解析配置
	if _po.Config != "" && isUnmarshalConfig {
		err = cutil.JSON().UnmarshalFromString(_po.Config, &eo.Config)
		if err != nil {
			err = errors.Wrapf(err, "PublishedAgent unmarshal config error")
			return
		}
	}

	return
}
