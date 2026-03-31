package authzhttp

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

// ListPolicy 查询策略列表
func (a *authZHttpAcc) ListPolicy(ctx context.Context, req *authzhttpreq.ListPolicyReq, userToken string) (res *authzhttpres.ListPolicyRes, err error) {
	url := fmt.Sprintf("%s/api/authorization/v1/policy?%s",
		a.publicBaseURL, req.ToReqQuery())

	opt := httphelper.WithToken(userToken)
	c := httphelper.NewHTTPClient(opt)

	var respByte []byte
	if cenvhelper.IsAaronLocalDev() {
		respByte = a.getListPolicyMockData()
	} else {
		respByte, err = c.GetExpect2xxByte(ctx, url)
	}

	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.ListPolicy http get")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	err = cutil.JSON().Unmarshal(respByte, &res)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.ListPolicy json unmarshal")
		err = errors.Wrap(err, "解析JSON响应失败")

		return
	}

	return
}

func (a *authZHttpAcc) getListPolicyMockData() (bys []byte) {
	bys = []byte(`{
    "entries": [
        {
            "expires_at": "1970-01-01T08:00:00+08:00",
            "id": "c425f82b-0b70-4406-8b47-9baa34ffa27c",
            "resource": {
                "id": "5d494a31-f42e-451c-a132-47adb0b15410",
                "type": "agent",
                "name": "agent1"
            },
            "accessor": {
                "id": "5238483c-6bb6-11f0-87ed-fa9a8e685be1",
                "type": "role",
                "name": "mock_role_name"
            },
            "operation": {
                "allow": [
                    {
                        "id": "use",
                        "name": "使用"
                    },
                    {
                        "id": "delete",
                        "name": "删除"
                    }
                ],
                "deny": []
            },
            "condition": ""
        },
        {
            "expires_at": "1970-01-01T08:00:00+08:00",
            "id": "e463e80f-5033-48a9-a323-aa567a4d0976",
            "resource": {
                "id": "5d494a31-f42e-451c-a132-47adb0b15410",
                "type": "agent",
                "name": "agent2"
            },
            "accessor": {
                "id": "33565778-309d-11f0-92fb-1688e6ea28e2",
                "type": "user",
                "name": "01",
                "parent_deps": [
                    [
                        {
                            "id": "b41e83c8-4a54-11f0-bb07-1688e6ea28e2",
                            "name": "AISHU",
                            "type": "department"
                        },
                        {
                            "id": "b421853c-4a54-11f0-92fb-1688e6ea28e2",
                            "name": "测试部",
                            "type": "department"
                        },
                        {
                            "id": "0095fc1a-6083-11f0-9713-fa9a8e685be1",
                            "name": "态势感知",
                            "type": "department"
                        }
                    ]
                ]
            },
            "operation": {
                "allow": [
                    {
                        "id": "use",
                        "name": "使用"
                    },
                    {
                        "id": "modify",
                        "name": "编辑"
                    }
                ],
                "deny": []
            },
            "condition": ""
        }
    ],
    "total_count": 2
}`)

	return
}
