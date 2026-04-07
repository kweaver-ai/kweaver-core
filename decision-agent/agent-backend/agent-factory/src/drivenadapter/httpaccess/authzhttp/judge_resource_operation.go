package authzhttp

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/pkg/errors"
)

// ResourceOperation 获取资源操作
func (a *authZHttpAcc) ResourceOperation(ctx context.Context, req *authzhttpreq.ResourceOperationReq) (list []*authzhttpres.ResourceOperationItem, err error) {
	list = []*authzhttpres.ResourceOperationItem{}

	url := a.privateBaseURL + "/api/authorization/v1/resource-operation"

	c := httphelper.NewHTTPClient()

	respBody, err := c.PostJSONExpect2xxByte(ctx, url, req)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.ResourceOperation http post")
		return nil, errors.Wrap(err, "发送HTTP请求失败")
	}

	err = json.Unmarshal(respBody, &list)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.ResourceOperation unmarshal response")
		return nil, errors.Wrap(err, "解析响应数据失败")
	}

	return
}

func (a *authZHttpAcc) ResourceOperationSingle(ctx context.Context, accessor *authzhttpreq.Accessor, resource *authzhttpreq.Resource) (list []*authzhttpres.ResourceOperationItem, err error) {
	req := authzhttpreq.NewResourceOperationReqSingle(accessor, resource)

	return a.ResourceOperation(ctx, req)
}

func (a *authZHttpAcc) GetResourceOpsByUid(ctx context.Context, uid string, resource *authzhttpreq.Resource) (ops []cdapmsenum.Operator, err error) {
	req := authzhttpreq.NewResourceOperationReqSingleByUid(uid, resource)

	list, err := a.ResourceOperation(ctx, req)
	if err != nil {
		return
	}

	for _, item := range list {
		ops = append(ops, item.Operation...)
	}

	return
}

func (a *authZHttpAcc) GetAgentResourceOpsByUid(ctx context.Context, uid string) (opMap map[cdapmsenum.Operator]bool, err error) {
	opMap = make(map[cdapmsenum.Operator]bool)
	resource := &authzhttpreq.Resource{
		ID:   cconstant.PmsAllFlag,
		Type: cdaenum.ResourceTypeDataAgent,
	}
	req := authzhttpreq.NewResourceOperationReqSingleByUid(uid, resource)

	list, err := a.ResourceOperation(ctx, req)
	if err != nil {
		return
	}

	for _, item := range list {
		for _, op := range item.Operation {
			opMap[op] = true
		}
	}

	return
}

func (a *authZHttpAcc) GetAgentTplResourceOpsByUid(ctx context.Context, uid string) (opMap map[cdapmsenum.Operator]bool, err error) {
	opMap = make(map[cdapmsenum.Operator]bool)

	resource := &authzhttpreq.Resource{
		ID:   cconstant.PmsAllFlag,
		Type: cdaenum.ResourceTypeDataAgentTpl,
	}
	req := authzhttpreq.NewResourceOperationReqSingleByUid(uid, resource)

	list, err := a.ResourceOperation(ctx, req)
	if err != nil {
		return
	}

	for _, item := range list {
		for _, op := range item.Operation {
			opMap[op] = true
		}
	}

	return
}
