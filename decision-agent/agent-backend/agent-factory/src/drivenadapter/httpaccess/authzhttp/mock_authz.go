package authzhttp

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iauthzacc"
)

type mockAuthZHttpAcc struct {
	logger icmp.Logger
}

var _ iauthzacc.AuthZHttpAcc = &mockAuthZHttpAcc{}

func NewMockAuthZHttpAcc(logger icmp.Logger) iauthzacc.AuthZHttpAcc {
	return &mockAuthZHttpAcc{
		logger: logger,
	}
}

func (m *mockAuthZHttpAcc) ResourceList(ctx context.Context, req *authzhttpreq.ResourceListReq) (list []*authzhttpres.ResourceListItem, err error) {
	m.logger.Infof("[MockAuthZ] ResourceList: accessorID=%s, resourceType=%s", req.Accessor.ID, req.Resource.Type)

	list = []*authzhttpres.ResourceListItem{}

	return list, nil
}

func (m *mockAuthZHttpAcc) ResourceFilter(ctx context.Context, req *authzhttpreq.ResourceFilterReq) (list []*authzhttpres.ResourceListItem, err error) {
	m.logger.Infof("[MockAuthZ] ResourceFilter: accessorID=%s, resourceCount=%d", req.Accessor.ID, len(req.Resources))

	list = []*authzhttpres.ResourceListItem{}

	return list, nil
}

func (m *mockAuthZHttpAcc) ResourceOperation(ctx context.Context, req *authzhttpreq.ResourceOperationReq) (list []*authzhttpres.ResourceOperationItem, err error) {
	m.logger.Infof("[MockAuthZ] ResourceOperation: accessorID=%s, resourceCount=%d", req.Accessor.ID, len(req.Resources))

	list = []*authzhttpres.ResourceOperationItem{}

	return list, nil
}

func (m *mockAuthZHttpAcc) GetCanUseAgentIDs(ctx context.Context, uid string) (agentIDs []string, err error) {
	m.logger.Infof("[MockAuthZ] GetCanUseAgentIDs: uid=%s", uid)

	agentIDs = []string{}

	return agentIDs, nil
}

func (m *mockAuthZHttpAcc) FilterCanUseAgentIDs(ctx context.Context, uid string, agentIDs []string) (filteredAgentIDs []string, err error) {
	m.logger.Infof("[MockAuthZ] FilterCanUseAgentIDs: uid=%s, agentCount=%d", uid, len(agentIDs))

	filteredAgentIDs = []string{}

	return filteredAgentIDs, nil
}

func (m *mockAuthZHttpAcc) FilterCanUseAgentIDMap(ctx context.Context, uid string, agentIDs []string) (filteredAgentIDMap map[string]struct{}, err error) {
	m.logger.Infof("[MockAuthZ] FilterCanUseAgentIDMap: uid=%s, agentCount=%d", uid, len(agentIDs))

	filteredAgentIDMap = make(map[string]struct{})

	return filteredAgentIDMap, nil
}

func (m *mockAuthZHttpAcc) OperationCheck(ctx context.Context, req *authzhttpreq.SingleCheckReq) (result *authzhttpres.SingleCheckResult, err error) {
	m.logger.Infof("[MockAuthZ] OperationCheck: accessorID=%s, resourceID=%s", req.Accessor.ID, req.Resource.ID)

	result = &authzhttpres.SingleCheckResult{
		Result: true,
	}

	return result, nil
}

func (m *mockAuthZHttpAcc) SingleAgentUseCheck(ctx context.Context, accessorID string, accessorType cenum.PmsTargetObjType, agentID string) (ok bool, err error) {
	m.logger.Infof("[MockAuthZ] SingleAgentUseCheck: accessorID=%s, agentID=%s", accessorID, agentID)
	return true, nil
}

func (m *mockAuthZHttpAcc) GetAgentResourceOpsByUid(ctx context.Context, uid string) (opMap map[cdapmsenum.Operator]bool, err error) {
	m.logger.Infof("[MockAuthZ] GetAgentResourceOpsByUid: uid=%s", uid)

	opMap = map[cdapmsenum.Operator]bool{}

	return opMap, nil
}

func (m *mockAuthZHttpAcc) GetAgentTplResourceOpsByUid(ctx context.Context, uid string) (opMap map[cdapmsenum.Operator]bool, err error) {
	m.logger.Infof("[MockAuthZ] GetAgentTplResourceOpsByUid: uid=%s", uid)

	opMap = map[cdapmsenum.Operator]bool{}

	return opMap, nil
}

func (m *mockAuthZHttpAcc) CreatePolicy(ctx context.Context, req []*authzhttpreq.CreatePolicyReq) (err error) {
	m.logger.Infof("[MockAuthZ] CreatePolicy: policyCount=%d", len(req))
	return nil
}

func (m *mockAuthZHttpAcc) GrantAgentUsePmsForSingleAccessor(ctx context.Context, accessor *authzhttpreq.PolicyAccessor, agentID string, agentName string) (err error) {
	m.logger.Infof("[MockAuthZ] GrantAgentUsePmsForSingleAccessor: accessorID=%s, agentID=%s", accessor.ID, agentID)
	return nil
}

func (m *mockAuthZHttpAcc) GrantAgentUsePmsForAccessors(ctx context.Context, accessors []*authzhttpreq.PolicyAccessor, agentID string, agentName string) (err error) {
	m.logger.Infof("[MockAuthZ] GrantAgentUsePmsForAccessors: accessorCount=%d, agentID=%s", len(accessors), agentID)
	return nil
}

func (m *mockAuthZHttpAcc) GrantAgentUsePmsForAppAdmin(ctx context.Context) (err error) {
	m.logger.Infof("[MockAuthZ] GrantAgentUsePmsForAppAdmin")
	return nil
}

func (m *mockAuthZHttpAcc) GrantMgmtPmsForAppAdmin(ctx context.Context) (err error) {
	m.logger.Infof("[MockAuthZ] GrantMgmtPmsForAppAdmin")
	return nil
}

func (m *mockAuthZHttpAcc) DenyAgentUsePmsForAllAccessor(ctx context.Context, agentID string, agentName string) (err error) {
	m.logger.Infof("[MockAuthZ] DenyAgentUsePmsForAllAccessor: agentID=%s", agentID)
	return nil
}

func (m *mockAuthZHttpAcc) DeletePolicy(ctx context.Context, req *authzhttpreq.PolicyDeleteParams) (err error) {
	m.logger.Infof("[MockAuthZ] DeletePolicy")
	return nil
}

func (m *mockAuthZHttpAcc) DeleteAgentPolicy(ctx context.Context, agentID string) (err error) {
	m.logger.Infof("[MockAuthZ] DeleteAgentPolicy: agentID=%s", agentID)
	return nil
}

func (m *mockAuthZHttpAcc) ListPolicy(ctx context.Context, req *authzhttpreq.ListPolicyReq, userToken string) (res *authzhttpres.ListPolicyRes, err error) {
	m.logger.Infof("[MockAuthZ] ListPolicy")

	res = &authzhttpres.ListPolicyRes{}

	return res, nil
}

func (m *mockAuthZHttpAcc) ListPolicyAll(ctx context.Context, req *authzhttpreq.ListPolicyReq, userToken string) (res *authzhttpres.ListPolicyRes, err error) {
	m.logger.Infof("[MockAuthZ] ListPolicyAll")

	res = &authzhttpres.ListPolicyRes{}

	return res, nil
}

func (m *mockAuthZHttpAcc) SetResourceType(ctx context.Context, resourceTypeID cdaenum.ResourceType, req *authzhttpreq.ResourceTypeSetReq) (err error) {
	m.logger.Infof("[MockAuthZ] SetResourceType: resourceType=%s", resourceTypeID)
	return nil
}
