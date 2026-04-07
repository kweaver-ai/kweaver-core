// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	"github.com/segmentio/kafka-go"
)

const (
	// poll 阻塞时间
	POLL_TIMEOUT_MS = 100
	// 生产者生产消息阻塞的时间
	PRODUCE_FLUSH_TIMEOUT_MS = 100

	// 设置为 5, 平衡吞吐量和顺序性
	MAX_IN_FLIGHT_REQUESTS_PER_CONNECTION = 5

	// 限制topic最大长度: 249-len("<tenant>.mdl.dr.<clusterID>.customer."),tenant最长10位, kafka本身限制长度为249个字符,clusterID为22个字符
	SRC_TOPIC_MAX_LENGTH = 200

	// 允许的最大消息大小(byte)
	MAX_MESSAGE_BYTES = 20971520

	// kafka消息留存时间, 8小时
	RETENTION_MS = "28800000"
	// kafka消息单个分区的留存大小，100M
	RETENTION_BYTES = "104857600"
)

//go:generate mockgen -source ../interfaces/kafka_access.go -destination ../interfaces/mock/mock_kafka_access.go
type KafkaAccess interface {
	NewReader(ctx context.Context, topic string, groupID string) (*kafka.Reader, error)
	NewWriter(ctx context.Context, topic string) (*kafka.Writer, error)
	WriteMessages(ctx context.Context, w *kafka.Writer, msgs ...kafka.Message) error
	ReadMessage(ctx context.Context, r *kafka.Reader) (kafka.Message, error)
	CommitMessages(ctx context.Context, r *kafka.Reader, msgs ...kafka.Message) error
	CreateTopic(ctx context.Context, topicName string) error
	CloseReader(r *kafka.Reader)
	CloseWriter(w *kafka.Writer)
}
