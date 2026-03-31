package umcmp

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/umcmp/dto/umret"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// SearchOrg 组织范围搜索【内部接口】
// 1、查看某个或某些用户是否在某个或某些组织结构对象下
// 2、查看某个或某些部门是否在某个或某些组织结构对象下
// http://{host}:{post}/api/user-management/v1/search-org
func (u *Um) SearchOrg(ctx context.Context,
	args *umarg.SearchOrgArgDto,
) (ret *umret.SearchOrgRetDto, err error) {
	c := httphelper.NewHTTPClient()

	if args.DepartmentIDs == nil {
		args.DepartmentIDs = []string{}
	}

	if args.UserIDs == nil {
		args.UserIDs = []string{}
	}

	umArgDto := umarg.NewSearchOrgUMArgDto(args)
	apiURL := fmt.Sprintf("%s/v1/search-org", u.getPrivateURLPrefix())
	u.logger.Infof("SearchOrg apiURL: %s", apiURL)

	resp, err := c.PostJSONExpect2xx(ctx, apiURL, umArgDto)
	if err != nil {
		return
	}

	err = cutil.JSON().Unmarshal([]byte(resp), &ret)

	return
}
