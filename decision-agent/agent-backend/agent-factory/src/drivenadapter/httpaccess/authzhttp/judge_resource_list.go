package authzhttp

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/pkg/errors"
)

// ResourceList 资源列举
func (a *authZHttpAcc) ResourceList(ctx context.Context, req *authzhttpreq.ResourceListReq) (list []*authzhttpres.ResourceListItem, err error) {
	list = []*authzhttpres.ResourceListItem{}

	url := a.privateBaseURL + "/api/authorization/v1/resource-list"

	c := httphelper.NewHTTPClient()

	respBody, err := c.PostJSONExpect2xxByte(ctx, url, req)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.ResourceList http post")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	err = json.Unmarshal(respBody, &list)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.ResourceList unmarshal response")
		err = errors.Wrap(err, "解析响应数据失败")

		return
	}

	return
}

func (a *authZHttpAcc) GetCanUseAgentIDs(ctx context.Context, uid string) (agentIDs []string, err error) {
	// 1. 获取此uid可以使用Agent的列表
	req := authzhttpreq.NewCanUseAgentListReqByUid(uid)

	list, err := a.ResourceList(ctx, req)
	if err != nil {
		return
	}

	for _, item := range list {
		agentIDs = append(agentIDs, item.ID)
	}

	//// 2. 获取*可使用Agent的列表
	//req = authzhttpreq.NewCanUseAgentListReqByStar()
	//
	//list, err = a.ResourceList(ctx, req)
	//if err != nil {
	//	return
	//}
	//
	//for _, item := range list {
	//	agentIDs = append(agentIDs, item.ID)
	//}

	return
}
