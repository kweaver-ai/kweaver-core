package drivenadapters

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-worker/interfaces"
)

var (
	kaOnce sync.Once
	ka     interfaces.MQAccess
)

type mqAccess struct {
	appSetting *common.AppSetting
}

func NewMQAccess(appSetting *common.AppSetting) interfaces.MQAccess {
	kaOnce.Do(func() {
		ka = &mqAccess{
			appSetting: appSetting,
		}
	})

	return ka
}

// 提交该批量全部位移
func (ka *mqAccess) DoCommit(c *kafka.Consumer) error {

	partions, err := c.Assignment()
	if err != nil {
		logger.Errorf("KafkaAccess DoCommit Failed to Assignment: %v", err)
		return err
	}
	pos, err := c.Position(partions)
	if err != nil {
		logger.Errorf("KafkaAccess DoCommit Failed to Position: %v", err)
		return err
	}
	_, err = c.CommitOffsets(pos)
	if err != nil {
		logger.Errorf("KafkaAccess DoCommit Failed to CommitOffsets: %v", err)
		return err
	}
	logger.Debugf("KafkaAccess DoCommit Position %+v\n", pos)
	return nil
}

func (kAccess *mqAccess) CloseAdminClient(adminClient *kafka.AdminClient) {
	if adminClient != nil {
		adminClient.Close()
	}
}

func (kAccess *mqAccess) DoProduceAndCommit(kp *interfaces.KafkaProducer, c *kafka.Consumer, msgs []*kafka.Message) error {
	err := kp.Producer.BeginTransaction()
	if err != nil {
		logger.Errorf("Begin transaction failed, %v", err)
		return err
	}

	// logger.Debugf("Preparing to produce %d messages", len(msgs))

	// // 生产者发送消息使用的 delivery channel
	// deliveryChan := make(chan kafka.Event, kAccess.appSetting.ServerSetting.FlushItems)
	// defer close(deliveryChan)

	for _, record := range msgs {
		err = kp.Producer.Produce(record, kp.DeliveryChan)
		if err != nil {
			logger.Errorf("Send message failed, %v", err)
			abortErr := kAccess.abortTxn(err, kp.Producer)
			if abortErr != nil {
				return abortErr
			}
			return err
		}
	}

	for i := 0; i < len(msgs); i++ {
		event, ok := <-kp.DeliveryChan
		if !ok {
			logger.Error("kafka produce message to topic fail because of delivery chan is closed")
			return fmt.Errorf("delivery chan of kafka producer is closed")
		}

		message := event.(*kafka.Message)

		if message.TopicPartition.Error != nil {
			logger.Errorf("Delivery failed: %v", message.TopicPartition.Error)
			return message.TopicPartition.Error
		}
	}

	kp.Producer.Flush(interfaces.PRODUCE_FLUSH_TIMEOUT_MS)

	// 获取当前消费者消费的分区
	partitions, err := c.Assignment()
	if err != nil {
		logger.Errorf("Get conusmer assignment failed, %v", err)
		return err
	}
	// 获取给定分区的要提交的位移
	position, err := getConsumerPosition(c, partitions)
	if err != nil {
		logger.Errorf("Get conusmer position failed, %v", err)
		return err
	}
	consumerMetadata, err := c.GetConsumerGroupMetadata()
	if err != nil {
		logger.Errorf("Get consumer group metadata failed: %v", err)
		return err
	}

	txnOperateTimeout := time.Duration(kAccess.appSetting.KafkaSetting.TransactionTimeoutMs) * time.Millisecond
	sendOffsetsCtx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()

	err = kp.Producer.SendOffsetsToTransaction(sendOffsetsCtx, position, consumerMetadata)
	if err != nil {
		logger.Errorf("Send offsets to transaction failed, %v", err)
		abortErr := kAccess.abortTxn(err, kp.Producer)
		if abortErr != nil {
			return abortErr
		}
		return err
	}

	commitCtx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()

	err = kp.Producer.CommitTransaction(commitCtx)
	if err != nil {
		logger.Errorf("Committed transaction failed, %s", err)
		abortErr := kAccess.abortTxn(err, kp.Producer)
		if abortErr != nil {
			return abortErr
		}

		return err
	}

	return nil
}

// 创建消费者
func (kAccess *mqAccess) NewConsumer(groupID string) (*kafka.Consumer, error) {
	consumerConfig := kafka.ConfigMap{
		"bootstrap.servers":                   fmt.Sprintf("%s:%d", kAccess.appSetting.MQSetting.MQHost, kAccess.appSetting.MQSetting.MQPort),
		"security.protocol":                   "sasl_plaintext",
		"group.id":                            groupID,
		"enable.auto.commit":                  false,
		"auto.offset.reset":                   kAccess.appSetting.KafkaSetting.AutoOffsetReset,
		"isolation.level":                     "read_committed",
		"session.timeout.ms":                  kAccess.appSetting.KafkaSetting.SessionTimeoutMs,
		"socket.timeout.ms":                   kAccess.appSetting.KafkaSetting.SocketTimeoutMs,
		"socket.keepalive.enable":             true,
		"heartbeat.interval.ms":               kAccess.appSetting.KafkaSetting.HeartbeatIntervalMs,
		"max.poll.interval.ms":                kAccess.appSetting.KafkaSetting.MaxPollIntervalMs,
		"allow.auto.create.topics":            false,
		"sasl.mechanism":                      kAccess.appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":                       kAccess.appSetting.MQSetting.Auth.Username,
		"sasl.password":                       kAccess.appSetting.MQSetting.Auth.Password,
		"enable.ssl.certificate.verification": false,
	}

	c, err := kafka.NewConsumer(&consumerConfig)
	if err != nil {
		logger.Errorf("Failed to create consumer: %v", err)
		return nil, err
	}

	logger.Debugf("Create %s consumer %v on cluster %s", groupID, c, kAccess.appSetting.MQSetting.MQHost)

	return c, nil
}

// 创建事务生产者
func (kAccess *mqAccess) NewTransactionalProducer(txId string) (*kafka.Producer, error) {
	producerConfig := kafka.ConfigMap{
		"client.id":                             txId,
		"bootstrap.servers":                     fmt.Sprintf("%s:%d", kAccess.appSetting.MQSetting.MQHost, kAccess.appSetting.MQSetting.MQPort),
		"security.protocol":                     "sasl_plaintext",
		"acks":                                  "all",
		"transactional.id":                      txId,
		"enable.idempotence":                    true,
		"retries":                               kAccess.appSetting.KafkaSetting.Retries,
		"max.in.flight.requests.per.connection": interfaces.MAX_IN_FLIGHT_REQUESTS_PER_CONNECTION,
		"message.max.bytes":                     interfaces.MAX_MESSAGE_BYTES,
		"retry.backoff.ms":                      kAccess.appSetting.KafkaSetting.RetryBackoffMs,
		"transaction.timeout.ms":                kAccess.appSetting.KafkaSetting.TransactionTimeoutMs,
		"socket.keepalive.enable":               true,
		"allow.auto.create.topics":              false,
		"sasl.mechanism":                        kAccess.appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":                         kAccess.appSetting.MQSetting.Auth.Username,
		"sasl.password":                         kAccess.appSetting.MQSetting.Auth.Password,
		"enable.ssl.certificate.verification":   false,
	}

	p, err := kafka.NewProducer(&producerConfig)
	if err != nil {
		logger.Errorf("Create transactional producer failed, %v", err)
		return nil, err
	}

	// Listen to all the client instance-level errors.
	// It's important to read these errors too otherwise the events channel will eventually fill up
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case kafka.Error:
				// Generic client instance-level errors, such as broker connection failures, authentication issues, etc.
				// These errors should generally be considered informational as the underlying client will automatically try to
				// recover from any errors encountered, the application does not need to take action on them.
				logger.Errorf("Kafka error, error msg: %v", ev)
			default:
				// default do nothing, won't record log
			}
		}
	}()

	logger.Debugf("Create %s producer %v on cluster %s", txId, p, kAccess.appSetting.MQSetting.MQHost)
	return p, nil
}

// 执行消费 返回 message 或者 err
func (kAccess *mqAccess) DoConsume(c *kafka.Consumer) (*kafka.Message, error) {
	ev := c.Poll(interfaces.POLL_TIMEOUT_MS)
	if ev == nil {
		return nil, nil
	}

	switch e := ev.(type) {
	case *kafka.Message:
		return e, nil

	case kafka.Error:
		err := errors.New(e.String())
		return nil, err

	default:
		logger.Debugf("Ignored %v", e)
	}
	return nil, nil
}

// 获取消费者需要提交的位移
func getConsumerPosition(c *kafka.Consumer, partitions []kafka.TopicPartition) ([]kafka.TopicPartition, error) {
	position, err := c.Position(partitions)
	if err != nil {
		logger.Errorf("Failed to get consumer offsets, %v", err)
		return nil, err
	}

	return position, nil
}

// 中止事务
func (kAccess *mqAccess) abortTxn(err error, producer *kafka.Producer) error {
	txnOperateTimeout := time.Duration(kAccess.appSetting.KafkaSetting.TransactionTimeoutMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()

	logger.Infof("Current 'TxnRequiresAbort' result is %v", err.(kafka.Error).TxnRequiresAbort())

	if err.(kafka.Error).TxnRequiresAbort() {
		err = producer.AbortTransaction(ctx)
		logger.Infof("Abort transaction for %v", producer)
		if err != nil {
			if err.(kafka.Error).Code() == kafka.ErrState {
				logger.Infof("No transaction in progress, ignore the error")
				err = nil
			} else {
				logger.Errorf("Abort transaction for %v failed: %v", producer, err)
			}
		}
	}

	return err
}
