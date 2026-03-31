package bizdomainhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

func (e *bizDomainHttpAcc) HasResourceAssociation(ctx context.Context, req *bizdomainhttpreq.QueryResourceAssociationSingleReq) (hasAssociation bool, err error) {
	_req := &bizdomainhttpreq.QueryResourceAssociationsReq{
		BdID:   req.BdID,
		ID:     req.ID,
		Type:   req.Type,
		Limit:  1,
		Offset: 0,
	}

	res, err := e.QueryResourceAssociations(ctx, _req)
	if err != nil {
		chelper.RecordErrLogWithPos(e.logger, err, "bizDomainHttpAcc.HasResourceAssociation query")
		err = errors.Wrap(err, "查询资源关联关系失败")

		return
	}

	return len(res.Items) > 0, nil
}
