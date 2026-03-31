package umhttpaccess

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

// GetUserUserGroupIDs 获取用户的用户组ID列表
func (u *umHttpAcc) GetUserUserGroupIDs(ctx context.Context, userID string) (userGroupIDs []string, err error) {
	dto := &umarg.GetUserInfoArgDto{
		UserIds: []string{userID},
		Fields: umarg.Fields{
			umarg.FieldGroups,
		},
	}

	var uim umcmp.UserInfoMap

	uim, err = u.um.GetUserInfo(ctx, dto)
	if err != nil {
		chelper.RecordErrLogWithPos(u.logger, err, "umHttpAcc.GetUserUserGroupIDs")
		return nil, errors.Wrap(err, "[GetUserUserGroupIDs]:获取用户的用户组信息失败")
	}

	userGroupIDs = make([]string, 0)

	for _, ui := range uim {
		for _, group := range ui.Groups {
			userGroupIDs = append(userGroupIDs, group.ID)
		}
	}

	return
}
