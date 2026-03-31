package authzhttp

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/pkg/errors"
)

// SetResourceType 设置资源类型（私有接口）
func (a *authZHttpAcc) SetResourceType(ctx context.Context, resourceTypeID cdaenum.ResourceType, req *authzhttpreq.ResourceTypeSetReq) (err error) {
	url := fmt.Sprintf("%s/api/authorization/v1/resource_type/%s", a.privateBaseURL, resourceTypeID.String())

	c := httphelper.NewHTTPClient()

	_, err = c.PutJSONExpect2xx(ctx, url, req)
	if err != nil {
		chelper.RecordErrLogWithPos(a.logger, err, "authZHttpAcc.SetResourceType http put")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	return
}
