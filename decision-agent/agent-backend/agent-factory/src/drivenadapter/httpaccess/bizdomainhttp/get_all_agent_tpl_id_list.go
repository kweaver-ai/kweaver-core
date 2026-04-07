package bizdomainhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpres"
)

func (e *bizDomainHttpAcc) GetAllAgentTplIDList(ctx context.Context, bdIDs []string) (agentIDs []string, err error) {
	for _, bdID := range bdIDs {
		_req := &bizdomainhttpreq.QueryResourceAssociationsReq{
			BdID:   bdID,
			Type:   cdaenum.ResourceTypeDataAgentTpl,
			Limit:  -1,
			Offset: 0,
		}

		var res *bizdomainhttpres.QueryResourceAssociationsRes

		res, err = e.QueryResourceAssociations(ctx, _req)
		if err != nil {
			return
		}

		agentIDs = append(agentIDs, res.GetItemIDs()...)
	}

	return
}
