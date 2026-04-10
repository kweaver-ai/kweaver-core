package publishedp2e

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/locale"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/pubedeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc"
)

// PublishedTplListEo PO转EO
func PublishedTplListEo(ctx context.Context, _po *dapo.PublishedTplPo) (eo *pubedeo.PublishedTplListEo, err error) {
	eo = &pubedeo.PublishedTplListEo{}

	err = cutil.CopyStructUseJSON(&eo.PublishedTplPo, _po)
	if err != nil {
		return
	}

	return
}

// PublishedTplListEos 批量PO转EO
func PublishedTplListEos(ctx context.Context, _pos []*dapo.PublishedTplPo, umHttp iumacc.UmHttpAcc) (eos []*pubedeo.PublishedTplListEo, err error) {
	eos = make([]*pubedeo.PublishedTplListEo, 0, len(_pos))

	userIDs := make([]string, 0, len(_pos))
	for i := range _pos {
		userIDs = append(userIDs, _pos[i].PublishedBy)
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
		var eo *pubedeo.PublishedTplListEo

		if eo, err = PublishedTplListEo(ctx, _pos[i]); err != nil {
			return
		}

		if eo.PublishedBy != "" {
			userName, ok := ret.UserNameMap[eo.PublishedBy]
			if !ok {
				userName = unknownUserName
			}

			eo.PublishedByName = userName
		}

		eos = append(eos, eo)
	}

	return
}
