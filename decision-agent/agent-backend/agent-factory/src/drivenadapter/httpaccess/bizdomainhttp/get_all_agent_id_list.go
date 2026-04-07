package bizdomainhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpres"
)

func (e *bizDomainHttpAcc) GetAllAgentIDList(ctx context.Context, bdIDs []string) (agentIDs []string, agentID2BdIDMap map[string]string, err error) {
	agentIDs = make([]string, 0)
	agentID2BdIDMap = make(map[string]string)

	for _, bdID := range bdIDs {
		_req := &bizdomainhttpreq.QueryResourceAssociationsReq{
			BdID:   bdID,
			Type:   cdaenum.ResourceTypeDataAgent,
			Limit:  -1,
			Offset: 0,
		}

		var res *bizdomainhttpres.QueryResourceAssociationsRes

		res, err = e.QueryResourceAssociations(ctx, _req)
		if err != nil {
			return
		}

		agentIDs = append(agentIDs, res.GetItemIDs()...)

		for _, agentID := range res.GetItemIDs() {
			agentID2BdIDMap[agentID] = bdID
		}
	}

	return
}
