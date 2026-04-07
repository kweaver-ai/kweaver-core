package utils

import (
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	msqclient "github.com/kweaver-ai/proton-mq-sdk-go"
)

//go:generate mockgen -package mock -source ../utils/msmq.go -destination ../mock/mock_msmq.go

// MsmqClient 消息队列服务
type MsmqClient interface {
	Subscribe(topic, channel string, cmd func([]byte) error)
	Publish(topic string, message []byte) error
}

// msmq 消息队列句柄结构
type msmq struct {
	client                   msqclient.ProtonMQClient
	pollIntervalMilliseconds int64
	maxInFlight              int
}

var (
	mqOnce   sync.Once
	mqClient MsmqClient
	config   = NewConfiger().Config()
	logger   = NewLogger()
)

const (
	BmqType  = "bmq"
	TLQHtp20 = "htp20"
)

// NewMsmqClient 加载消息队列服务
func NewMsmqClient() MsmqClient {
	mqOnce.Do(func() {
		mqClient = &msmq{
			client:                   getMQClient(),
			pollIntervalMilliseconds: 100,
			maxInFlight:              200,
		}
	})
	return mqClient
}

func getMQClient() msqclient.ProtonMQClient {
	mqClient, err := msqclient.NewProtonMQClientFromFile(config.MQConfigPath)
	if err != nil {
		logger.Panicf("newMQClient error: %v", err)
	}
	return mqClient
}

// Subscribe 订阅消息，订阅某个主题
func (m *msmq) Subscribe(topic, channel string, cmd func([]byte) error) {
	go func() {
		switch config.MQConnectorType {
		case BmqType:
		case TLQHtp20:
			err := m.client.Sub(topic, common.ChannelECron, cmd, m.pollIntervalMilliseconds, m.maxInFlight)
			logger.Errorln(err)
		default:
			err := m.client.Sub(topic, channel, cmd, m.pollIntervalMilliseconds, m.maxInFlight)
			logger.Errorln(err)
		}
	}()
}

// Publish 发布消息
func (m *msmq) Publish(topic string, message []byte) error {
	if err := m.client.Pub(topic, message); err != nil {
		logger.Errorln(err)
		return err
	}
	return nil
}
