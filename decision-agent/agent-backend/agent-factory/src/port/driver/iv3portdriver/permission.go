package iv3portdriver

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/rdto/agent_permission/cpmsreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/rdto/agent_permission/cpmsresp"
)

//go:generate mockgen -source=./permission.go -destination ./v3portdrivermock/permission.go -package v3portdrivermock
type IPermissionSvc interface {
	// CheckUsePermission 检查非个人空间下的某个agent是否有运行权限
	CheckUsePermission(ctx context.Context, req *cpmsreq.CheckAgentRunReq) (resp *cpmsresp.CheckRunResp, err error)

	// GetUserStatus 获取用户拥有的管理权限状态
	GetUserStatus(ctx context.Context) (resp *cpmsresp.UserStatusResp, err error)

	// GetSingleMgmtPermission 获取单个资源的管理权限
	GetSingleMgmtPermission(ctx context.Context, resouceType cdaenum.ResourceType, operator cdapmsenum.Operator) (allAllowed bool, err error)

	// InitPermission 初始化权限
	InitPermission(ctx context.Context) (err error)

	// CheckIsCustomSpaceMember 检查用户是否是自定义空间的成员
	// CheckIsCustomSpaceMember(ctx context.Context, req *cpmsreq.CheckIsCustomSpaceMemberReq) (resp *cpmsresp.CheckIsCustomSpaceMemberResp, err error)

	// GetPolicyOfAgentUse 获取智能体使用权限策略列表
	GetPolicyOfAgentUse(ctx context.Context, agentID string) (res *authzhttpres.ListPolicyRes, err error)
}
