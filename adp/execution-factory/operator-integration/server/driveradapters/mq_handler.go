package driveradapters

import (
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/mq"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/toolbox"
)

type mqHandler struct {
	MQClient            mq.MQClient
	ToolboxEventHandler interfaces.ToolBoxEventHandler
	Logger              interfaces.Logger
}

var (
	mqOnce            sync.Once
	mqHandlerInstance interfaces.MQHandler
)

// NewMQHandler 创建MQ处理接口
func NewMQHandler() interfaces.MQHandler {
	mqOnce.Do(func() {
		conf := config.NewConfigLoader()
		mqHandlerInstance = &mqHandler{
			MQClient:            mq.NewMQClient(),
			ToolboxEventHandler: toolbox.NewToolServiceImpl(),
			Logger:              conf.GetLogger(),
		}
	})
	return mqHandlerInstance
}

// 待处理Topic列表
var pendingTopics = []string{
	interfaces.OperatorDeleteEventTopic,
}

// Subscribe 订阅事件
func (h *mqHandler) Subscribe() {
	for _, topic := range pendingTopics {
		switch topic {
		case interfaces.OperatorDeleteEventTopic:
			h.MQClient.Subscribe(topic, interfaces.ChannelMessage, h.ToolboxEventHandler.HandleOperatorDeleteEvent)
		default:
			h.Logger.Errorf("unknown topic: %s", topic)
		}
	}
}
