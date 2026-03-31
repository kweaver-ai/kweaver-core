package authzhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/pkg/errors"
)

// DeletePolicy 策略删除
func (a *authZHttpAcc) DeletePolicy(ctx context.Context, req *authzhttpreq.PolicyDeleteParams) (err error) {
	url := a.privateBaseURL + "/api/authorization/v1/policy-delete"

	c := httphelper.NewHTTPClient()

	_, err = c.PostJSONExpect2xx(ctx, url, req)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.DeletePolicy http post")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	return
}

// DeletePolicyAgentUsePms 删除Agent使用权限
func (a *authZHttpAcc) DeleteAgentPolicy(ctx context.Context, agentID string) (err error) {
	req := authzhttpreq.NewPolicyAgentDeleteReq(agentID)

	err = a.DeletePolicy(ctx, req)
	if err != nil {
		return
	}

	return
}
