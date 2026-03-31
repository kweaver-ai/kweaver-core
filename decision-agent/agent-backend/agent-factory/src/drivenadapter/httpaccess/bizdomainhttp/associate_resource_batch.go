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
	associateResourceBatchPath = "/internal/api/business-system/v1/resource/batch"
)

// AssociateResourceBatch 批量资源关联
// 重复的资源会忽略（即便此资源是被另外一个域绑定），不会产生409错误
func (e *bizDomainHttpAcc) AssociateResourceBatch(ctx context.Context, req bizdomainhttpreq.AssociateResourceBatchReq) (err error) {
	uri := fmt.Sprintf("%s%s", e.privateBaseURL, associateResourceBatchPath)

	c := httphelper.NewHTTPClient()

	_, err = c.PostJSONExpect2xxByte(ctx, uri, req)
	if err != nil {
		chelper.RecordErrLogWithPos(e.logger, err, "bizDomainHttpAcc.AssociateResourceBatch http post")
		err = errors.Wrap(err, "发送HTTP请求失败")

		return
	}

	return
}
