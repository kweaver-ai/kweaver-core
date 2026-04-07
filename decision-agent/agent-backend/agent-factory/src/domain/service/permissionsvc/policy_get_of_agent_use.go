package permissionsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpres"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/pkg/errors"
)

// GetPolicyOfAgentUse 获取智能体使用权限策略列表
func (svc *permissionSvc) GetPolicyOfAgentUse(ctx context.Context, agentID string) (res *authzhttpres.ListPolicyRes, err error) {
	// 1. 构造请求参数
	req := &authzhttpreq.ListPolicyReq{
		ResourceID:   agentID,
		ResourceType: cdaenum.ResourceTypeDataAgent,
	}

	// 2. 调用权限平台获取策略列表
	token := chelper.GetUserTokenFromCtx(ctx)

	res, err = svc.authZHttp.ListPolicyAll(ctx, req, token)
	if err != nil {
		err = errors.Wrapf(err, "获取智能体[%s]使用权限策略失败", agentID)
		return
	}

	if res == nil {
		err = errors.New("获取智能体使用权限策略返回空结果")
		return
	}

	// 3. 进行filter过滤
	// 3.1 过滤过期策略
	err = res.FilterByExpiresAt()
	if err != nil {
		err = errors.Wrapf(err, "[GetPolicyOfAgentUse][FilterByExpiresAt] failed")
		return
	}

	// 3.2 过滤操作（只保留“使用”权限）
	err = res.FilterByOperation(cdapmsenum.AgentUse)
	if err != nil {
		err = errors.Wrapf(err, "[GetPolicyOfAgentUse][FilterByOperation] failed")
		return
	}

	return
}
