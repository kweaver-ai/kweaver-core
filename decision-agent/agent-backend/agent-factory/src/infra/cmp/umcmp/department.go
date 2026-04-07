package umcmp

import (
	"context"
	"fmt"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/umtypes"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

// GetDeptInfoMap 获取部门信息（可批量）
// 当err为nil时，dim为非nil
// 【注意】此方法会自动去掉不存在的部门id，不会因为部门不存在而返回错误
func (u *Um) GetDeptInfoMap(ctx context.Context, args *umarg.GetDeptInfoArgDto) (dim map[string]*umtypes.DepartmentInfo, err error) {
	var (
		loopCount int
		maxLoop   = 3
	)

	dim = make(map[string]*umtypes.DepartmentInfo)

	c := httphelper.NewHTTPClient()

Loop:
	apiURL := u.getDeptInfoApiURL(args)
	u.logger.Infof("GetDeptInfoMap apiURL: %s", apiURL)

	resp, err := c.GetExpect2xx(ctx, apiURL)

	respErr := &httphelper.CommonRespError{}

	//nolint:nestif
	if errors.As(err, &respErr) {
		loopCount++
		if loopCount > maxLoop {
			return nil, errors.Wrap(err, "获取部门信息失败")
		}

		if respErr.Code == UmNotFound && respErr.Detail != nil {
			var notExistsIDs []string
			if _notExistsIDs, ok := respErr.Detail["ids"]; ok {
				notExistsIDs = cutil.MustStrSlice2(_notExistsIDs)
			}

			// 去掉不存在的用户id
			args.DeptIds = cutil.Difference(args.DeptIds, notExistsIDs)
			if len(args.DeptIds) == 0 {
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

	var dis []*umtypes.DepartmentInfo

	err = cutil.JSON().Unmarshal([]byte(resp), &dis)
	if err != nil {
		return
	}

	// 将用户信息转换为map，key为部门id，value为部门信息
	dim = make(map[string]*umtypes.DepartmentInfo, len(dis))
	for _, di := range dis {
		dim[di.DepartmentId] = di
	}

	return
}

// GetUserDeptIDs 获取用户所属部门id（包括直接部门和间接部门）
func (u *Um) GetUserDeptIDs(ctx context.Context, userID string) (deptIDs []string, err error) {
	deptIDs = make([]string, 0)

	apiURL := fmt.Sprintf("%s/v1/users/%v/department_ids", u.getPrivateURLPrefix(), userID)
	u.logger.Infof("GetUserDeptIDs apiURL: %s", apiURL)

	c := httphelper.NewHTTPClient()

	str, err := c.GetExpect2xx(ctx, apiURL)
	if err != nil {
		return
	}

	err = cutil.JSON().Unmarshal([]byte(str), &deptIDs)

	return
}

// 和GetDeptInfoMap区别是：没有loop，不会去掉不存在的部门id
func (u *Um) GetDeptInfoMap2(ctx context.Context, args *umarg.GetDeptInfoArgDto) (deptInfoMap map[string]*umtypes.DepartmentInfo, err error) {
	deptInfoMap = make(map[string]*umtypes.DepartmentInfo)

	apiURL := u.getDeptInfoApiURL(args)
	u.logger.Infof("GetDeptInfoMap2 apiURL: %s", apiURL)

	c := httphelper.NewHTTPClient()

	str, err := c.GetExpect2xx(ctx, apiURL)
	if err != nil {
		return
	}

	deptInfos := make([]*umtypes.DepartmentInfo, 0)

	err = cutil.JSON().Unmarshal([]byte(str), &deptInfos)
	if err != nil {
		return
	}

	for _, deptInfo := range deptInfos {
		deptInfoMap[deptInfo.DepartmentId] = deptInfo
	}

	return
}

func (u *Um) getDeptInfoApiURL(args *umarg.GetDeptInfoArgDto) string {
	apiURL := fmt.Sprintf("%s/v1/departments/%s/%s", u.getPrivateURLPrefix(),
		strings.Join(args.DeptIds, ","),
		strings.Join(args.Fields.ToStrings(), ","),
	)

	return apiURL
}
