package authzhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/pkg/errors"
)

// CreatePolicy 新建策略
func (a *authZHttpAcc) CreatePolicy(ctx context.Context, req []*authzhttpreq.CreatePolicyReq) (err error) {
	url := a.privateBaseURL + "/api/authorization/v1/policy"

	c := httphelper.NewHTTPClient()

	_, err = c.PostJSONExpect2xx(ctx, url, req)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.CreatePolicy http post")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	return
}
