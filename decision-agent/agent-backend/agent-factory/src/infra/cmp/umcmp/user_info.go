package umcmp

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

// GetUserInfo 获取用户信息（可批量）
// 当err为nil时，uim为非nil
// 【注意】此方法会自动去掉不存在的用户id，不会因为用户不存在而返回错误
func (u *Um) GetUserInfo(ctx context.Context, args *umarg.GetUserInfoArgDto) (uim UserInfoMap, err error) {
	var (
		loopCount int
		maxLoop   = 3
	)

	uim = make(UserInfoMap)

	c := httphelper.NewHTTPClient()

Loop:
	apiURL := u.getUserInfoApiURL(args)
	u.logger.Infof("GetUserInfo apiURL: %s", apiURL)

	resp, err := c.GetExpect2xx(ctx, apiURL)

	respErr := &httphelper.CommonRespError{}

	//nolint:nestif
	if errors.As(err, &respErr) {
		loopCount++
		if loopCount > maxLoop {
			return nil, errors.Wrap(err, "获取用户信息失败")
		}

		if respErr.Code == UmNotFound && respErr.Detail != nil {
			var notExistsIDs []string
			if _notExistsIDs, ok := respErr.Detail["ids"]; ok {
				notExistsIDs = cutil.MustStrSlice2(_notExistsIDs)
			}

			// 去掉不存在的用户id
			args.UserIds = cutil.Difference(args.UserIds, notExistsIDs)
			if len(args.UserIds) == 0 {
				// 当去掉不存在的后为空时，返回
				err = nil
				return
			}

			goto Loop
		}
	}

	if err != nil {
		return
	}

	var uis UserInfos

	err = cutil.JSON().Unmarshal([]byte(resp), &uis)
	if err != nil {
		return
	}

	// 将用户信息转换为map，key为用户id，value为用户信息
	uim = make(UserInfoMap, len(uis))
	for _, ui := range uis {
		uim[ui.Id] = ui
	}

	return
}

// GetUserName 获取单个用户名称
func (u *Um) GetUserName(ctx context.Context, userID string) (name string, isNotFound bool, err error) {
	_args := &umarg.GetUserInfoArgDto{
		UserIds: []string{userID},
		Fields:  umarg.Fields{umarg.FieldName},
	}

	c := httphelper.NewHTTPClient()

	apiURL := u.getUserInfoApiURL(_args)
	u.logger.Infof("GetUserName apiURL: %s", apiURL)

	resp, err := c.GetExpect2xx(ctx, apiURL)

	respErr := &httphelper.CommonRespError{}
	if errors.As(err, &respErr) {
		if respErr.Code == UmNotFound {
			// 没找到用户，修改标志，然后返回
			isNotFound = true
			return
		}
	}

	var uis UserInfos

	err = cutil.JSON().Unmarshal([]byte(resp), &uis)
	if err != nil {
		return
	}

	if len(uis) != 1 {
		err = errors.New("[GetUserName] 获取名称错误")
		return
	}

	name = uis[0].Name

	return
}

// GetUserEnableStatus 获取用户禁用状态(可批量)
func (u *Um) GetUserEnableStatus(ctx context.Context,
	args *umarg.GetUserEnableStatusArgDto,
) (uem umret.UserEnabledMap, err error) {
	_args := &umarg.GetUserInfoArgDto{
		UserIds: args.UserIds,
		Fields:  umarg.Fields{umarg.FieldEnabled},
	}

	uim, err := u.GetUserInfo(ctx, _args)
	if err != nil {
		return
	}

	uem = make(umret.UserEnabledMap, len(uim))
	for _, ui := range uim {
		uem[ui.Id] = ui.Enabled
	}

	return
}

// GetUserInfoSingle 获取单个用户信息
func (u *Um) GetUserInfoSingle(ctx context.Context,
	args *umarg.GetUserInfoSingleArgDto,
) (info UserInfo, isNotFound bool, err error) {
	_args := &umarg.GetUserInfoArgDto{
		UserIds: []string{args.UserID},
		Fields:  args.Fields,
	}

	c := httphelper.NewHTTPClient()

	apiURL := u.getUserInfoApiURL(_args)
	resp, err := c.GetExpect2xx(ctx, apiURL)

	respErr := &httphelper.CommonRespError{}
	if errors.As(err, &respErr) {
		if respErr.Code == UmNotFound {
			// 没找到用户，修改标志，然后返回
			isNotFound = true
			err = nil

			return
		}
	}

	if err != nil {
		return
	}

	var uis UserInfos

	err = cutil.JSON().Unmarshal([]byte(resp), &uis)
	if err != nil {
		return
	}

	if len(uis) != 1 {
		err = errors.New("[GetUserInfoSingle] 获取用户信息错误")
		return
	}

	info = *uis[0]

	return
}

//nolint:stylecheck
func (u *Um) getUserInfoApiURL(args *umarg.GetUserInfoArgDto) string {
	// http://{host}:{post}/api/user-management/v1/users/{user_ids}/{fields} 【内部接口】
	apiURL := fmt.Sprintf("%s/v1/users/%s/%s", u.getPrivateURLPrefix(),
		strings.Join(args.UserIds, ","),
		strings.Join(args.Fields.ToStrings(), ","),
	)

	return apiURL
}
