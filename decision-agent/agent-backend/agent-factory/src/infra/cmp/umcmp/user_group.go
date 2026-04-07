package umcmp

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umerr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"

	"github.com/pkg/errors"
)

// GetGroupMembers 获取用户组成员（批量）
// 【注意】此方法会自动去掉不存在的用户组id，不会因为用户组不存在而返回错误
func (u *Um) GetGroupMembers(ctx context.Context, args *umarg.GetGroupMembersArgDto) (
	ret *umret.GetGroupMembersRetDto, err error,
) {
	var (
		loopCount int
		maxLoop   = 3
	)

	ret = umret.NewGetGroupMembersRetDto()

	if len(args.GroupIDs) == 0 {
		err = errors.New("[GetGroupMembers]group_ids不能为空")
		return
	}

	c := httphelper.NewHTTPClient()

Loop:
	umArgDto := umarg.NewGetGroupMembersUMArgDto(args)

	apiURL := fmt.Sprintf("%s/v1/group-members", u.getPrivateURLPrefix())
	u.logger.Infof("GetGroupMembers apiURL: %s", apiURL)

	resp, err := c.PostJSONExpect2xx(ctx, apiURL, umArgDto)

	respErr := &httphelper.CommonRespError{}
	if errors.As(err, &respErr) {
		loopCount++
		if loopCount > maxLoop {
			return nil, errors.Wrap(err, "获取用户组成员失败")
		}

		if respErr.Code == umerr.GroupNotFound && respErr.Detail != nil {
			detailMap := respErr.Detail
			notExistsGIDsInter := detailMap["ids"]

			if notExistsUIDs, ok := notExistsGIDsInter.([]interface{}); ok && len(notExistsUIDs) > 0 {
				notExistsGIDsStrSlice := make([]string, 0, len(notExistsUIDs))
				for i := range notExistsUIDs {
					//nolint:forcetypeassert
					notExistsGIDsStrSlice = append(notExistsGIDsStrSlice, notExistsUIDs[i].(string))
				}
				// 去掉不存在的用户组ID
				args.GroupIDs = cutil.Difference(args.GroupIDs, notExistsGIDsStrSlice)

				if len(args.GroupIDs) == 0 {
					// 当去掉不存在的后为空时，返回
					err = nil
					return
				}

				goto Loop
			}
		}
	}

	if err != nil {
		return
	}

	err = cutil.JSON().Unmarshal([]byte(resp), &ret)
	if err != nil {
		return
	}

	return
}
