package authzhttp

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/pkg/errors"
)

// OperationCheck 单个决策（通用）
func (a *authZHttpAcc) OperationCheck(ctx context.Context, req *authzhttpreq.SingleCheckReq) (response *authzhttpres.SingleCheckResult, err error) {
	url := a.privateBaseURL + "/api/authorization/v1/operation-check"

	c := httphelper.NewHTTPClient()

	respBody, err := c.PostJSONExpect2xxByte(ctx, url, req)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.OperationCheck http post")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.OperationCheck unmarshal response")
		err = errors.Wrap(err, "解析响应数据失败")

		return
	}

	return
}

// SingleAgentUseCheck 单个Agent使用权限决策(普通用户/应用账号)
func (a *authZHttpAcc) SingleAgentUseCheck(ctx context.Context, accessorID string, accessorType cenum.PmsTargetObjType, agentID string) (ok bool, err error) {
	if accessorType != cenum.PmsTargetObjTypeUser && accessorType != cenum.PmsTargetObjTypeAppAccount {
		err = errors.New("accessorType must be user or app")
		return
	}

	var req *authzhttpreq.SingleCheckReq

	if accessorType == cenum.PmsTargetObjTypeUser {
		req = authzhttpreq.NewSingleUserAgentUseCheckReq(accessorID, agentID)
	} else {
		req = authzhttpreq.NewSingleAppAccountAgentUseCheckReq(accessorID, agentID)
	}

	response, err := a.OperationCheck(ctx, req)
	if err != nil {
		return
	}

	ok = response.Result

	return
}
