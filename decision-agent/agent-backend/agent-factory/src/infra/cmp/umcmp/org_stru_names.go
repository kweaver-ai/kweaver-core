package umcmp

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umerr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"

	"github.com/pkg/errors"
)

// GetOsnNames 获取组织架构对象的names
//
//nolint:funlen
func (u *Um) GetOsnNames(ctx context.Context, args *umarg.GetOsnArgDto) (ret *umtypes.OsnInfoMapS, err error) {
	var (
		loopCount int
		maxLoop   = 10
		_args     = *args // 复制一份，避免修改原始参数
	)

	ret = umtypes.NewOsnInfoMapS()

	c := httphelper.NewHTTPClient()

Loop:
	umArgDto := umarg.NewGetOsnUMArgDto(&_args)

	apiURL := fmt.Sprintf("%s/v1/names", u.getPrivateURLPrefix())
	u.logger.Infof("GetOsnNames apiURL: %s", apiURL)

	var resp []byte
	if cenvhelper.IsAaronLocalDev() {
		resp, err = u.getOsnNamesMockResp()
	} else {
		resp, err = c.PostJSONExpect2xxByte(ctx, apiURL, umArgDto)
	}

	respErr := &httphelper.CommonRespError{}
	if errors.As(err, &respErr) {
		loopCount++

		// 达到最大重试次数，返回错误
		if loopCount > maxLoop {
			return nil, errors.Wrap(err, "获取name信息失败")
		}

		// 如果是用户不存在，部门不存在，组不存在，那么去掉不存在的id，重新请求
		if (respErr.Code == umerr.UserNotFound || respErr.Code == umerr.DepartmentNotFound ||
			respErr.Code == umerr.GroupNotFound) && respErr.Detail != nil {
			detailMap := respErr.Detail

			notExistsIDsInter := detailMap["ids"]
			if notExistsIDs, ok := notExistsIDsInter.([]interface{}); ok && len(notExistsIDs) > 0 {
				notExistsIDsStrSlice := make([]string, 0, len(notExistsIDs))
				for i := range notExistsIDs {
					//nolint:forcetypeassert
					notExistsIDsStrSlice = append(notExistsIDsStrSlice, notExistsIDs[i].(string))
				}

				// 去掉不存在的id
				switch respErr.Code {
				case umerr.UserNotFound:
					_args.UserIDs = cutil.Difference(_args.UserIDs, notExistsIDsStrSlice)
				case umerr.DepartmentNotFound:
					_args.DepartmentIDs = cutil.Difference(_args.DepartmentIDs, notExistsIDsStrSlice)
				case umerr.GroupNotFound:
					_args.GroupIDs = cutil.Difference(_args.GroupIDs, notExistsIDsStrSlice)
				}

				// 如果去掉不存在的id后，没有id了，直接返回
				if len(_args.UserIDs) == 0 && len(_args.DepartmentIDs) == 0 && len(_args.GroupIDs) == 0 {
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

	// 解析返回值
	var retDto umret.GetOsnRetDto

	err = cutil.JSON().Unmarshal(resp, &retDto)
	if err != nil {
		return
	}

	// 转换为umtypes.OsnInfoMapS
	ret.FromGetOsnRetDto(&retDto)

	return
}

func (u *Um) getOsnNamesMockResp() ([]byte, error) {
	mockRet := umret.GetOsnRetDto{
		UserNames: []umret.IDName{
			{
				ID:   "user1",
				Name: "用户1",
			},
			{
				ID:   "user2",
				Name: "用户2",
			},
		},
		DepartmentNames: []umret.IDName{
			{
				ID:   "department1",
				Name: "部门1",
			},
			{
				ID:   "department2",
				Name: "部门2",
			},
		},
		GroupNames: []umret.IDName{
			{
				ID:   "group1",
				Name: "组1",
			},
			{
				ID:   "group2",
				Name: "组2",
			},
		},
		AppNames: []umret.IDName{},
	}

	return cutil.JSON().Marshal(mockRet)
}
