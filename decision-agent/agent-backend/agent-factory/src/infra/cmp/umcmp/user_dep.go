package umcmp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
)

// GetUserDept 获取用户部门信息
// 【注意】此方法会自动去掉不存在的用户id，不会因为用户不存在而返回错误
func (u *Um) GetUserDept(ctx context.Context, userID string) (depts [][]ObjectBaseInfo, err error) {
	depts = make([][]ObjectBaseInfo, 0)

	args := &umarg.GetUserInfoSingleArgDto{
		UserID: userID,
		Fields: umarg.Fields{
			umarg.FieldParentDeps,
		},
	}

	_info, isNotFound, err := u.GetUserInfoSingle(ctx, args)
	if err != nil {
		return
	}

	if isNotFound {
		return
	}

	depts = _info.ParentDeps

	return
}
