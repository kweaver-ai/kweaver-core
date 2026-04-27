package boot

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service/inject/v3/dainject"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/bddbacc/bdagentdbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/bddbacc/bdagenttpldbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/daconfdbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/daconftpldbacc"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
)

// initBizDomainRel 初始化业务域关联关系
func initBizDomainRel() (err error) {
	if global.GConfig.IsBizDomainDisabled() {
		return
	}

	bizDomainSvc := dainject.NewBizDomainSvc()
	ctx := context.Background()

	// 1. 初始化agent的业务域关联
	agentRepo := daconfdbacc.NewDataAgentRepo()
	bdAgentRelRepo := bdagentdbacc.NewBizDomainAgentRelRepo()

	err = bizDomainSvc.InitBizDomainAgentRel(ctx, agentRepo, bdAgentRelRepo)
	if err != nil {
		return
	}

	// 2. 初始化agent模板的业务域关联
	agentTplRepo := daconftpldbacc.NewDataAgentTplRepo()
	bdAgentTplRelRepo := bdagenttpldbacc.NewBizDomainAgentTplRelRepo()

	err = bizDomainSvc.InitBizDomainAgentTplRel(ctx, agentTplRepo, bdAgentTplRelRepo)
	if err != nil {
		return
	}

	return
}
