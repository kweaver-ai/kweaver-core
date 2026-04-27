package cdaenum

import "github.com/pkg/errors"

type PublishToBe string

const (
	// 发布为 API
	PublishToBeAPIAgent PublishToBe = "api_agent"
	// 发布为 Web SDK
	PublishToBeWebSDKAgent PublishToBe = "web_sdk_agent"
	// 发布为 技能
	PublishToBeSkillAgent PublishToBe = "skill_agent"

	// 发布为 数据流 agent
	PublishToBeDataFlowAgent PublishToBe = "data_flow_agent"
)

func (t PublishToBe) EnumCheck() (err error) {
	if t != PublishToBeAPIAgent && t != PublishToBeWebSDKAgent && t != PublishToBeSkillAgent && t != PublishToBeDataFlowAgent {
		err = errors.New("[PublishToBe]: invalid publish_to_be")
		return
	}

	return
}
