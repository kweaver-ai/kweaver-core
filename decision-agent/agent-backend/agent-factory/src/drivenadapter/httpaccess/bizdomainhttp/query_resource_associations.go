package bizdomainhttp

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

func (e *bizDomainHttpAcc) QueryResourceAssociations(ctx context.Context, req *bizdomainhttpreq.QueryResourceAssociationsReq) (res *bizdomainhttpres.QueryResourceAssociationsRes, err error) {
	uri := fmt.Sprintf("%s%s", e.privateBaseURL, associateResourcePath)

	c := httphelper.NewHTTPClient()

	var respByte []byte

	respByte, err = c.GetExpect2xxByte(ctx, uri, req)
	if err != nil {
		chelper.RecordErrLogWithPos(e.logger, err, "bizDomainHttpAcc.QueryResourceAssociations http get")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	err = cutil.JSON().Unmarshal(respByte, &res)
	if err != nil {
		chelper.RecordErrLogWithPos(e.logger, err, "bizDomainHttpAcc.QueryResourceAssociations json unmarshal")
		err = errors.Wrap(err, "解析JSON响应失败")

		return
	}

	return
}
