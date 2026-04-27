package bizdomainhttp

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/bizdomainhttp/bizdomainhttpres"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc"
)

type mockBizDomainHttpAcc struct {
	logger icmp.Logger
}

var _ ibizdomainacc.BizDomainHttpAcc = &mockBizDomainHttpAcc{}

func NewMockBizDomainHttpAcc(logger icmp.Logger) ibizdomainacc.BizDomainHttpAcc {
	return &mockBizDomainHttpAcc{
		logger: logger,
	}
}

func (m *mockBizDomainHttpAcc) AssociateResource(ctx context.Context, req *bizdomainhttpreq.AssociateResourceReq) (err error) {
	m.logger.Infof("[MockBizDomain] AssociateResource: bdID=%s, id=%s, type=%s", req.BdID, req.ID, req.Type)
	return nil
}

func (m *mockBizDomainHttpAcc) AssociateResourceBatch(ctx context.Context, req bizdomainhttpreq.AssociateResourceBatchReq) (err error) {
	m.logger.Infof("[MockBizDomain] AssociateResourceBatch: resourceCount=%d", len(req))
	return nil
}

func (m *mockBizDomainHttpAcc) DisassociateResource(ctx context.Context, req *bizdomainhttpreq.DisassociateResourceReq) (err error) {
	m.logger.Infof("[MockBizDomain] DisassociateResource: bdID=%s, id=%s", req.BdID, req.ID)
	return nil
}

func (m *mockBizDomainHttpAcc) QueryResourceAssociations(ctx context.Context, req *bizdomainhttpreq.QueryResourceAssociationsReq) (res *bizdomainhttpres.QueryResourceAssociationsRes, err error) {
	m.logger.Infof("[MockBizDomain] QueryResourceAssociations: bdID=%s, resourceType=%s", req.BdID, req.Type)

	res = &bizdomainhttpres.QueryResourceAssociationsRes{
		Items: []*bizdomainhttpres.ResourceAssociationItem{},
	}

	return res, nil
}

func (m *mockBizDomainHttpAcc) GetAllAgentIDList(ctx context.Context, bdIDs []string) (agentIDs []string, agentID2BdIDMap map[string]string, err error) {
	m.logger.Infof("[MockBizDomain] GetAllAgentIDList: bdIDs=%v", bdIDs)

	agentIDs = []string{}
	agentID2BdIDMap = make(map[string]string)

	return agentIDs, agentID2BdIDMap, nil
}

func (m *mockBizDomainHttpAcc) GetAllAgentTplIDList(ctx context.Context, bdIDs []string) (agentIDs []string, err error) {
	m.logger.Infof("[MockBizDomain] GetAllAgentTplIDList: bdIDs=%v", bdIDs)

	agentIDs = []string{}

	return agentIDs, nil
}

func (m *mockBizDomainHttpAcc) HasResourceAssociation(ctx context.Context, req *bizdomainhttpreq.QueryResourceAssociationSingleReq) (hasAssociation bool, err error) {
	m.logger.Infof("[MockBizDomain] HasResourceAssociation: bdID=%s, id=%s", req.BdID, req.ID)
	return false, nil
}
