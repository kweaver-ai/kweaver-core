package authzhttp

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/pkg/errors"
)

// ResourceFilter 资源过滤
func (a *authZHttpAcc) ResourceFilter(ctx context.Context, req *authzhttpreq.ResourceFilterReq) (list []*authzhttpres.ResourceListItem, err error) {
	list = []*authzhttpres.ResourceListItem{}

	url := a.privateBaseURL + "/api/authorization/v1/resource-filter"

	c := httphelper.NewHTTPClient()

	respBody, err := c.PostJSONExpect2xxByte(ctx, url, req)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.ResourceFilter http post")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	err = json.Unmarshal(respBody, &list)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.ResourceFilter unmarshal response")
		err = errors.Wrap(err, "解析响应数据失败")

		return
	}

	return
}

func (a *authZHttpAcc) FilterCanUseAgentIDs(ctx context.Context, uid string, agentIDs []string) (filteredAgentIDs []string, err error) {
	if global.GConfig.SwitchFields.DisablePmsCheck {
		filteredAgentIDs = make([]string, len(agentIDs))
		copy(filteredAgentIDs, agentIDs)

		return
	}

	req := authzhttpreq.NewFilterCanUseAgentReq(uid, agentIDs)

	list, err := a.ResourceFilter(ctx, req)
	if err != nil {
		return
	}

	for _, item := range list {
		filteredAgentIDs = append(filteredAgentIDs, item.ID)
	}

	return
}

func (a *authZHttpAcc) FilterCanUseAgentIDMap(ctx context.Context, uid string, agentIDs []string) (filteredAgentIDMap map[string]struct{}, err error) {
	filteredAgentIDMap = make(map[string]struct{}, len(agentIDs))

	agentIDs = cutil.RemoveEmptyStrFromSlice(agentIDs)
	agentIDs = cutil.DeduplGeneric(agentIDs)

	if len(agentIDs) == 0 {
		return
	}

	// 1. 过滤
	filteredAgentIDs, err := a.FilterCanUseAgentIDs(ctx, uid, agentIDs)
	if err != nil {
		return
	}

	// 2. 转换为map
	for _, item := range filteredAgentIDs {
		filteredAgentIDMap[item] = struct{}{}
	}

	return
}
