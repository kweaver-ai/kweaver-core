package permissionsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdapmsenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/authzhttp/authzhttpreq"
	"github.com/pkg/errors"
)

// updateAgentResourceType 更新Agent资源类型
func (svc *permissionSvc) updateAgentResourceType(ctx context.Context) (err error) {
	// 构建Agent资源类型的操作列表
	operations := make([]*authzhttpreq.ResourceTypeOperationItem, 0)

	// 获取所有Agent操作
	allAgentOps := cdapmsenum.GetAllAgentOperator()
	for _, op := range allAgentOps {
		opItem := buildAgentOperationItem(op)
		if opItem != nil {
			operations = append(operations, opItem)
		}
	}

	req := &authzhttpreq.ResourceTypeSetReq{
		Name:        "智能体",
		Description: "DataAgent",
		InstanceURL: "",
		DataStruct:  "string",
		Operation:   operations,
		Hidden:      false,
	}

	err = svc.authZHttp.SetResourceType(ctx, cdaenum.ResourceTypeDataAgent, req)
	if err != nil {
		err = errors.Wrapf(err, "set resource type failed")
		return
	}

	return
}

// buildAgentOperationItem 构建Agent操作项
func buildAgentOperationItem(op cdapmsenum.Operator) *authzhttpreq.ResourceTypeOperationItem {
	opID := string(op)

	var names []*authzhttpreq.ResourceTypeOperationName

	var scope []string

	switch op {
	case cdapmsenum.AgentPublish:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "发布"},
			{Language: "en-us", Value: "Publish"},
			{Language: "zh-tw", Value: "發布"},
		}
		scope = []string{"type"}
	case cdapmsenum.AgentUnpublish:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "取消发布"},
			{Language: "en-us", Value: "Unpublish"},
			{Language: "zh-tw", Value: "取消發布"},
		}
		scope = []string{"type"}
	case cdapmsenum.AgentUnpublishOtherUserAgent:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "取消发布他人的智能体"},
			{Language: "en-us", Value: "Unpublish other user's Agent"},
			{Language: "zh-tw", Value: "取消發布他人的智能體"},
		}
		scope = []string{"type"}
	case cdapmsenum.AgentPublishToBeSkillAgent:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "发布为技能智能体"},
			{Language: "en-us", Value: "Publish as a Skill Agent"},
			{Language: "zh-tw", Value: "發布為技能智能體"},
		}
		scope = []string{"type"}
	case cdapmsenum.AgentPublishToBeWebSdkAgent:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "发布为Web SDK智能体"},
			{Language: "en-us", Value: "Publish as a Web SDK Agent"},
			{Language: "zh-tw", Value: "發布為Web SDK智能體"},
		}
		scope = []string{"type"}
	case cdapmsenum.AgentPublishToBeApiAgent:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "发布为API智能体"},
			{Language: "en-us", Value: "Publish as an API Agent"},
			{Language: "zh-tw", Value: "發布為API智能體"},
		}
		scope = []string{"type"}
	case cdapmsenum.AgentPublishToBeDataFlowAgent:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "发布为数据流智能体"},
			{Language: "en-us", Value: "Publish as a Dataflow Agent"},
			{Language: "zh-tw", Value: "發布為數據流智能體"},
		}
		scope = []string{"type"}
	case cdapmsenum.AgentCreateSystemAgent:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "创建系统智能体"},
			{Language: "en-us", Value: "Create a System Agent"},
			{Language: "zh-tw", Value: "創建系統智能體"},
		}
		scope = []string{"type"}
	case cdapmsenum.AgentUse:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "使用"},
			{Language: "en-us", Value: "Use"},
			{Language: "zh-tw", Value: "使用"},
		}
		scope = []string{"type", "instance"}
	case cdapmsenum.AgentBuiltInAgentMgmt:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "管理内置智能体"},
			{Language: "en-us", Value: "Manage built-in Agent"},
			{Language: "zh-tw", Value: "管理內置智能體"},
		}
		scope = []string{"type"}
	case cdapmsenum.AgentSeeTrajectoryAnalysis:
		names = []*authzhttpreq.ResourceTypeOperationName{
			{Language: "zh-cn", Value: "查看轨迹分析"},
			{Language: "en-us", Value: "See trajectory analysis"},
			{Language: "zh-tw", Value: "查看軌跡分析"},
		}
		scope = []string{"type"}
	default:
		return nil
	}

	return &authzhttpreq.ResourceTypeOperationItem{
		ID:          opID,
		Name:        names,
		Description: "",
		Scope:       scope,
	}
}
