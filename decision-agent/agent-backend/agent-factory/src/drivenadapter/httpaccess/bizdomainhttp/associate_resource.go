package bizdomainhttp

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/pkg/errors"
)

const (
	associateResourcePath = "/internal/api/business-system/v1/resource"
)

func (e *bizDomainHttpAcc) AssociateResource(ctx context.Context, req *bizdomainhttpreq.AssociateResourceReq) (err error) {
	uri := fmt.Sprintf("%s%s", e.privateBaseURL, associateResourcePath)

	c := httphelper.NewHTTPClient()

	_, err = c.PostJSONExpect2xxByte(ctx, uri, req)
	if err != nil {
		chelper.RecordErrLogWithPos(e.logger, err, "bizDomainHttpAcc.AssociateResource http post")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	return
}
