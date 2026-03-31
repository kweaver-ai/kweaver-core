package bizdomainhttp

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/pkg/errors"
)

func (e *bizDomainHttpAcc) DisassociateResource(ctx context.Context, req *bizdomainhttpreq.DisassociateResourceReq) (err error) {
	uri := fmt.Sprintf("%s%s", e.privateBaseURL, associateResourcePath)

	c := httphelper.NewHTTPClient()

	_, err = c.DeleteExpect2xxWithQueryParams(ctx, uri, req)
	if err != nil {
		chelper.RecordErrLogWithPos(e.logger, err, "bizDomainHttpAcc.DisassociateResource http delete")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	return
}
