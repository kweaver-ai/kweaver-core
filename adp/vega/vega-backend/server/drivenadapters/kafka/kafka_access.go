// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	kAccessOnce sync.Once
	kAccess     interfaces.KafkaAccess
)

type kafkaAccess struct {
	appSetting *common.AppSetting
}

func NewKafkaAccess(appSetting *common.AppSetting) interfaces.KafkaAccess {
	kAccessOnce.Do(func() {
		kAccess = &kafkaAccess{
			appSetting: appSetting,
		}
	})

	return kAccess
}

// getSASLDialer 创建带 SASL 认证的 Dialer
func (ka *kafkaAccess) getSASLDialer() *kafka.Dialer {
	mechanism := plain.Mechanism{
		Username: ka.appSetting.MQSetting.Auth.Username,
		Password: ka.appSetting.MQSetting.Auth.Password,
	}

	return &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}
}

// getBrokerAddress 获取 broker 地址
func (ka *kafkaAccess) getBrokerAddress() string {
	return fmt.Sprintf("%s:%d", ka.appSetting.MQSetting.MQHost, ka.appSetting.MQSetting.MQPort)
}

// NewReader 创建消费者
func (ka *kafkaAccess) NewReader(ctx context.Context, topic string, groupID string) (*kafka.Reader, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{ka.getBrokerAddress()},
		Topic:       topic,
		GroupID:     groupID,
		Dialer:      ka.getSASLDialer(),
		MaxBytes:    interfaces.MAX_MESSAGE_BYTES,
		StartOffset: kafka.FirstOffset,
		// 不设置 CommitInterval，使用手动提交
	})

	logger.Debugf("Created reader for topic %s with groupID %s on cluster %s", topic, groupID, ka.appSetting.MQSetting.MQHost)
	return r, nil
}

// CloseReader 关闭消费者
func (ka *kafkaAccess) CloseReader(r *kafka.Reader) {
	if r != nil {
		if err := r.Close(); err != nil {
			logger.Errorf("Failed to close reader: %v", err)
		}
	}
}

// NewWriter 创建生产者
func (ka *kafkaAccess) NewWriter(ctx context.Context, topic string) (*kafka.Writer, error) {
	w := &kafka.Writer{
		Addr:         kafka.TCP(ka.getBrokerAddress()),
		Topic:        topic,
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		RequiredAcks: kafka.RequireAll,
		Transport: &kafka.Transport{
			SASL: plain.Mechanism{
				Username: ka.appSetting.MQSetting.Auth.Username,
				Password: ka.appSetting.MQSetting.Auth.Password,
			},
		},
	}

	logger.Debugf("Created writer for topic %s on cluster %s", topic, ka.appSetting.MQSetting.MQHost)
	return w, nil
}

// CloseWriter 关闭生产者
func (ka *kafkaAccess) CloseWriter(w *kafka.Writer) {
	if w != nil {
		if err := w.Close(); err != nil {
			logger.Errorf("Failed to close writer: %v", err)
		}
	}
}

// WriteMessages 发送消息
func (ka *kafkaAccess) WriteMessages(ctx context.Context, w *kafka.Writer, msgs ...kafka.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	logger.Debugf("Preparing to write %d messages to topic %s", len(msgs), w.Topic)

	err := w.WriteMessages(ctx, msgs...)
	if err != nil {
		logger.Errorf("Failed to write messages to topic %s: %v", w.Topic, err)
		return err
	}

	logger.Debugf("Successfully wrote %d messages to topic %s", len(msgs), w.Topic)
	return nil
}

// ReadMessage 消费消息
func (ka *kafkaAccess) ReadMessage(ctx context.Context, r *kafka.Reader) (kafka.Message, error) {
	msg, err := r.ReadMessage(ctx)
	if err != nil {
		return kafka.Message{}, err
	}
	return msg, nil
}

// CommitMessages 手动提交位移
func (ka *kafkaAccess) CommitMessages(ctx context.Context, r *kafka.Reader, msgs ...kafka.Message) error {
	if err := r.CommitMessages(ctx, msgs...); err != nil {
		logger.Errorf("Failed to commit messages: %v", err)
		return err
	}
	logger.Debugf("Successfully committed %d messages", len(msgs))
	return nil
}

// CreateTopic 创建 topic
func (ka *kafkaAccess) CreateTopic(ctx context.Context, topicName string) error {
	logger.Infof("Creating topic %s", topicName)
	// 使用带 SASL 认证的连接
	dialer := ka.getSASLDialer()
	conn, err := dialer.DialContext(ctx, "tcp", ka.getBrokerAddress())
	if err != nil {
		logger.Errorf("Failed to dial kafka with SASL: %v", err)
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		logger.Errorf("Failed to get controller: %v", err)
		return err
	}

	controllerConn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(controller.Host, fmt.Sprintf("%d", controller.Port)))
	if err != nil {
		logger.Errorf("Failed to dial controller: %v", err)
		return err
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topicName,
			NumPartitions:     -1,
			ReplicationFactor: -1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		// 忽略 topic 已存在的错误
		if err.Error() == "Topic with this name already exists" {
			logger.Infof("Topic %s already exists", topicName)
			return nil
		}
		logger.Errorf("Failed to create topic %s: %v", topicName, err)
		return err
	}

	logger.Infof("Created topic %s", topicName)
	return nil
}
