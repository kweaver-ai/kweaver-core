package interfaces

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	// poll 阻塞时间
	POLL_TIMEOUT_MS = 100
)

const (
	// 设置为 1, 确保某一时刻只能发送一个请求, 避免因为 retry 导致的消息乱序
	MAX_IN_FLIGHT_REQUESTS_PER_CONNECTION = 1

	// 允许的最大消息大小(byte)
	MAX_MESSAGE_BYTES = 20971520
	// 生产者生产消息阻塞的时间
	PRODUCE_FLUSH_TIMEOUT_MS = 100
)

// 生产者结构体
type KafkaProducer struct {
	Producer     *kafka.Producer
	DeliveryChan chan kafka.Event
}

//go:generate mockgen -source ../interfaces/mq_access.go -destination ../interfaces/mock/mock_mq_access.go
type MQAccess interface {
	DoCommit(c *kafka.Consumer) error
	NewConsumer(groupID string) (*kafka.Consumer, error)
	NewTransactionalProducer(txId string) (*kafka.Producer, error)
	DoConsume(c *kafka.Consumer) (*kafka.Message, error)
	DoProduceAndCommit(kp *KafkaProducer, c *kafka.Consumer, msgs []*kafka.Message) error
}
