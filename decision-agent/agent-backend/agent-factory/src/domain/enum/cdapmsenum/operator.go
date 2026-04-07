package cdapmsenum

import (
	"errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type Operator string

const (
	// --- 1. Agent start---
	AgentPublish                 Operator = "publish"                    // 发布Agent
	AgentUnpublish               Operator = "unpublish"                  // 取消发布Agent
	AgentUnpublishOtherUserAgent Operator = "unpublish_other_user_agent" // 取消发布其他用户的Agent

	AgentPublishToBeSkillAgent    Operator = "publish_to_be_skill_agent"     // 发布为技能Agent
	AgentPublishToBeWebSdkAgent   Operator = "publish_to_be_web_sdk_agent"   // 发布为Web SDK Agent
	AgentPublishToBeApiAgent      Operator = "publish_to_be_api_agent"       // 发布为API Agent
	AgentPublishToBeDataFlowAgent Operator = "publish_to_be_data_flow_agent" // 发布为数据流Agent

	AgentCreateSystemAgent Operator = "create_system_agent" // 创建系统Agent
	AgentBuiltInAgentMgmt  Operator = "mgnt_built_in_agent" // 内置Agent管理

	// see_trajectory_analysis
	AgentSeeTrajectoryAnalysis Operator = "see_trajectory_analysis" // 查看轨迹分析

	AgentUse Operator = "use" // 使用Agent

	// --- 1. Agent end---

	// --- 2. Agent Template start---

	AgentTplPublish                    Operator = "publish"                        // 发布Agent模板
	AgentTplUnpublish                  Operator = "unpublish"                      // 取消发布Agent模板
	AgentTplUnpublishOtherUserAgentTpl Operator = "unpublish_other_user_agent_tpl" // 取消发布其他用户Agent模板

	// --- 2. Agent Template end---
)

func (e Operator) String() string {
	return string(e)
}

func (e Operator) EnumCheck() (err error) {
	if !cutil.ExistsGeneric([]Operator{AgentPublish, AgentUnpublish, AgentPublishToBeSkillAgent, AgentPublishToBeWebSdkAgent, AgentPublishToBeApiAgent, AgentPublishToBeDataFlowAgent, AgentCreateSystemAgent, AgentUnpublishOtherUserAgent, AgentUse, AgentTplPublish, AgentTplUnpublish, AgentTplUnpublishOtherUserAgentTpl, AgentBuiltInAgentMgmt, AgentSeeTrajectoryAnalysis}, e) {
		err = errors.New("[Operator]: invalid operator")
		return
	}

	return
}

func GetAllOperator() []Operator {
	return []Operator{
		AgentPublish,
		AgentUnpublish,
		AgentPublishToBeSkillAgent,
		AgentPublishToBeWebSdkAgent,
		AgentPublishToBeApiAgent,
		AgentPublishToBeDataFlowAgent,
		AgentCreateSystemAgent,
		AgentUnpublishOtherUserAgent,

		AgentUse,
		AgentBuiltInAgentMgmt,
		AgentSeeTrajectoryAnalysis,

		AgentTplPublish,
		AgentTplUnpublish,
		AgentTplUnpublishOtherUserAgentTpl,
	}
}

func GetAllAgentMgmtOperator() []Operator {
	return []Operator{
		AgentPublish,
		AgentUnpublish,
		AgentPublishToBeSkillAgent,
		AgentPublishToBeWebSdkAgent,
		AgentPublishToBeApiAgent,
		AgentPublishToBeDataFlowAgent,
		AgentCreateSystemAgent,
		AgentUnpublishOtherUserAgent,
		AgentBuiltInAgentMgmt,
		AgentSeeTrajectoryAnalysis,
	}
}

func GetAllAgentUseOperator() []Operator {
	return []Operator{
		AgentUse,
	}
}

func GetAllAgentOperator() []Operator {
	return append(GetAllAgentMgmtOperator(), GetAllAgentUseOperator()...)
}

func GetAllAgentTplOperator() []Operator {
	return []Operator{
		AgentTplPublish,
		AgentTplUnpublish,
		AgentTplUnpublishOtherUserAgentTpl,
	}
}
