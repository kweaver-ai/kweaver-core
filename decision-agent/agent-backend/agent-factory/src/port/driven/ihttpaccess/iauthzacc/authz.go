package iauthzacc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

//go:generate mockgen -source=./authz.go -destination ./authzaccmock/authz_mock.go -package authzaccmock
type AuthZHttpAcc interface {
	// ---- 1. 策略决策接口 start ----

	ResourceList(ctx context.Context, req *authzhttpreq.ResourceListReq) (list []*authzhttpres.ResourceListItem, err error)
	ResourceFilter(ctx context.Context, req *authzhttpreq.ResourceFilterReq) (list []*authzhttpres.ResourceListItem, err error)
	ResourceOperation(ctx context.Context, req *authzhttpreq.ResourceOperationReq) (list []*authzhttpres.ResourceOperationItem, err error)

	GetCanUseAgentIDs(ctx context.Context, uid string) (agentIDs []string, err error)

	FilterCanUseAgentIDs(ctx context.Context, uid string, agentIDs []string) (filteredAgentIDs []string, err error)
	FilterCanUseAgentIDMap(ctx context.Context, uid string, agentIDs []string) (filteredAgentIDMap map[string]struct{}, err error)

	// ---1.1 单个决策接口 start---
	OperationCheck(ctx context.Context, req *authzhttpreq.SingleCheckReq) (result *authzhttpres.SingleCheckResult, err error)
	SingleAgentUseCheck(ctx context.Context, accessorID string, accessorType cenum.PmsTargetObjType, agentID string) (ok bool, err error)
	// ---1.1 单个决策接口 end---

	// ---1.2 资源操作接口 start---
	GetAgentResourceOpsByUid(ctx context.Context, uid string) (opMap map[cdapmsenum.Operator]bool, err error)
	GetAgentTplResourceOpsByUid(ctx context.Context, uid string) (opMap map[cdapmsenum.Operator]bool, err error)
	// ---1.2 资源操作接口 end---

	// ---- 1. 策略决策接口 end ----

	// ---- 2. 策略配置接口 start ----
	// --- 2.1 create start---
	// 通用创建策略接口
	CreatePolicy(ctx context.Context, req []*authzhttpreq.CreatePolicyReq) (err error)

	// 为单个访问者授予Agent使用权限
	GrantAgentUsePmsForSingleAccessor(ctx context.Context, accessor *authzhttpreq.PolicyAccessor, agentID string, agentName string) (err error)

	// 为多个访问者授予Agent使用权限
	GrantAgentUsePmsForAccessors(ctx context.Context, accessors []*authzhttpreq.PolicyAccessor, agentID string, agentName string) (err error)

	// 给应用管理员授予Agent使用权限
	GrantAgentUsePmsForAppAdmin(ctx context.Context) (err error)

	// 给应用管理员授予Agent管理权限
	GrantMgmtPmsForAppAdmin(ctx context.Context) (err error)

	// 给所有访问者拒绝某个Agent的使用权限
	DenyAgentUsePmsForAllAccessor(ctx context.Context, agentID string, agentName string) (err error)
	// --- 2.1 create end---

	// --- 2.2 delete start---
	// 通用删除策略接口
	DeletePolicy(ctx context.Context, req *authzhttpreq.PolicyDeleteParams) (err error)

	// 删除Agent使用权限
	DeleteAgentPolicy(ctx context.Context, agentID string) (err error)
	// --- 2.2 delete end---

	// ---- 2. 策略配置接口 end ----

	// ---- 3. 策略查询接口 start ----
	ListPolicy(ctx context.Context, req *authzhttpreq.ListPolicyReq, userToken string) (res *authzhttpres.ListPolicyRes, err error)
	ListPolicyAll(ctx context.Context, req *authzhttpreq.ListPolicyReq, userToken string) (res *authzhttpres.ListPolicyRes, err error)
	// ---- 3. 策略查询接口 end ----

	// ---- 4. 资源类型配置接口 start ----
	// 设置资源类型（私有接口）
	SetResourceType(ctx context.Context, resourceTypeID cdaenum.ResourceType, req *authzhttpreq.ResourceTypeSetReq) (err error)
	// ---- 4. 资源类型配置接口 end ----
}
