package interfaces

import (
	"context"
)

const (
	// kafka消息留存时间, 8小时
	RETENTION_MS = "28800000"
	// kafka消息单个分区的留存大小，100M
	RETENTION_BYTES = "104857600"
)

var (
	TopicInputName  = "%s.sdp.%s.input"
	TopicOutputName = "%s.mdl.process.%s"
	TopicErrorName  = "%s.sdp.%s.error"
)

type TopicMetadata struct {
	TopicName       string
	PartitionsCount int
}

//go:generate mockgen -source ../interfaces/mq_access.go -destination ../interfaces/mock/mock_mq_access.go
type MQAccess interface {
	CreateTopicsOrPartitions(ctx context.Context, topicNames []string) error
	DeleteTopics(ctx context.Context, topicNames []string) error
}
