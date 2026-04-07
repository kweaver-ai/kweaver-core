package chat_enum

import "errors"

type ChatScenarioType string

const (
	ChatScenarioADPChatPage   ChatScenarioType = "ADP_chat_page"   // adp chat页面
	ChatScenarioADPAgentDebug ChatScenarioType = "ADP_agent_debug" // agent配置页面的debug
	ChatScenarioADPDataFlow   ChatScenarioType = "ADP_data_flow"   // adp数据流

	ChatScenarioThirdSystem ChatScenarioType = "third_system" // 第三方系统

	ChatScenarioCustom ChatScenarioType = "custom" // 自定义场景
)

func (c ChatScenarioType) ToString() string {
	return string(c)
}

func (c ChatScenarioType) EnumCheck() (err error) {
	switch c {
	case ChatScenarioADPChatPage, ChatScenarioADPAgentDebug, ChatScenarioADPDataFlow,
		ChatScenarioThirdSystem, ChatScenarioCustom:
		return
	default:
		err = errors.New("对话场景类型不合法")
		return
	}
}
