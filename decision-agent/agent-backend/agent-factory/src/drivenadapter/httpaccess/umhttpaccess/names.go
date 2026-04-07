package umhttpaccess

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

func (u *umHttpAcc) GetOsnNames(ctx context.Context, dto *umarg.GetOsnArgDto) (ret *umtypes.OsnInfoMapS, err error) {
	ret, err = u.um.GetOsnNamesSFG(ctx, dto)
	if err != nil {
		return nil, errors.Wrap(err, "获取组织架构names失败")
	}

	return
}

func (u *umHttpAcc) GetUserIDNameMap(ctx context.Context, userIDs []string) (idNameMap map[string]string, err error) {
	getNamesDto := &umarg.GetOsnArgDto{
		UserIDs: userIDs,
	}

	ret, err := u.GetOsnNames(ctx, getNamesDto)
	if err != nil {
		chelper.RecordErrLogWithPos(u.logger, err, "umHttpAcc.GetUserIDNameMap")
		return nil, errors.Wrap(err, "获取用户ID=\u003eName键值对失败")
	}

	idNameMap = ret.UserNameMap

	return
}

func (u *umHttpAcc) GetSingleUserName(ctx context.Context, userID string) (name string, err error) {
	idNameMap, err := u.GetUserIDNameMap(ctx, []string{userID})
	if err != nil {
		chelper.RecordErrLogWithPos(u.logger, err, "umHttpAcc.GetSingleUserName")
		return "", errors.Wrap(err, "获取单个用户名称失败")
	}

	name = idNameMap[userID]

	return
}
