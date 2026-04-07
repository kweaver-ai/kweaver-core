// 创建删除 topic
package drivenadapters

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
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

// 新建 adminClient
func (mqa *mqAccess) NewAdminClient() (*kafka.AdminClient, error) {
	adminConfig := kafka.ConfigMap{
		"bootstrap.servers":                   fmt.Sprintf("%s:%d", mqa.appSetting.MQSetting.MQHost, mqa.appSetting.MQSetting.MQPort),
		"security.protocol":                   "sasl_plaintext",
		"retries":                             mqa.appSetting.KafkaSetting.Retries,
		"socket.timeout.ms":                   mqa.appSetting.KafkaSetting.SocketTimeoutMs,
		"allow.auto.create.topics":            false,
		"sasl.mechanism":                      mqa.appSetting.MQSetting.Auth.Mechanism,
		"sasl.username":                       mqa.appSetting.MQSetting.Auth.Username,
		"sasl.password":                       mqa.appSetting.MQSetting.Auth.Password,
		"enable.ssl.certificate.verification": false,
	}

	admin, err := kafka.NewAdminClient(&adminConfig)
	if err != nil {
		logger.Errorf("Failed to create admin client: %v", err)
		return nil, err
	}

	return admin, nil
}

func (mqa *mqAccess) CloseAdminClient(adminClient *kafka.AdminClient) {
	if adminClient != nil {
		adminClient.Close()
	}
}

// 手动创建 topic和新增分区, 如果 topic 不存在，则创建; 若存在的 topic 的分区增加，则新建增加的分区
func (mqa *mqAccess) CreateTopicsOrPartitions(ctx context.Context, topicNames []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven:创建topic和新增分区", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	adminClient, err := mqa.NewAdminClient()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to create adminClient: %v", err)
		o11y.Error(ctx, errDetails)
		logger.Errorf(errDetails)
		span.SetStatus(codes.Error, "Failed to create adminClient")
		return err
	}
	defer mqa.CloseAdminClient(adminClient)

	// 获取集群元数据, 对于topic原信息，拿的是所有topic
	metadata, err := adminClient.GetMetadata(nil, true, mqa.appSetting.KafkaSetting.AdminClientRequestTimeoutMs)
	if err != nil {
		errDetails := fmt.Sprintf("Failed to get metadata, %s.", err.Error())
		o11y.Error(ctx, errDetails)
		logger.Errorf(errDetails)
		span.SetStatus(codes.Error, "Failed to get metadata")
		return err
	}

	allTopics := metadata.Topics
	resourceName := fmt.Sprint(metadata.Brokers[0].ID)
	topicSpecifications := make([]kafka.TopicSpecification, 0, 1)
	partitionsSpecifications := make([]kafka.PartitionsSpecification, 0)

	numPartitions, replicationFactor, err := mqa.GetDefaultConfig(ctx, adminClient, resourceName)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to GetDefaultConfig")
		logger.Errorf("Failed to GetDefaultConfig %s: %s", strings.Join(topicNames, ","), err.Error())
		return err
	}

	timeStr := mqa.appSetting.KafkaSetting.RetentionTime
	numberStr, unit := timeStr[:len(timeStr)-1], timeStr[len(timeStr)-1:]

	number, err := strconv.ParseInt(numberStr, 10, 64)
	if err != nil {
		return fmt.Errorf("parse retention time %s number %s err: %v", timeStr, numberStr, err)
	}
	switch unit {
	case "h":
		number = number * time.Hour.Milliseconds()
	case "d":
		number = number * 24 * time.Hour.Milliseconds()
	default:
		return fmt.Errorf("parse retention time %s unit %s err", timeStr, unit)
	}

	var retentionSize string
	if mqa.appSetting.KafkaSetting.RetentionSize != -1 {
		retentionSize = fmt.Sprintf("%d", common.GiBToBytes(int64(mqa.appSetting.KafkaSetting.RetentionSize)))
	} else {
		retentionSize = "-1"
	}

	retentionTime := fmt.Sprintf("%d", number)

	for _, topicName := range topicNames {
		if _, ok := allTopics[topicName]; !ok {
			topicSpecifications = append(topicSpecifications, kafka.TopicSpecification{
				Topic:             topicName,
				NumPartitions:     numPartitions,
				ReplicationFactor: replicationFactor,
				Config: map[string]string{
					// kafka 数据保留时长和保留大小和索引库的默认配置保持一致
					"retention.ms":    retentionTime,
					"retention.bytes": retentionSize,
				},
			})
		} else if numPartitions > len(allTopics[topicName].Partitions) {
			partitionsSpecifications = append(partitionsSpecifications, kafka.PartitionsSpecification{
				Topic:      topicName,
				IncreaseTo: numPartitions,
			})
		}
	}

	if len(topicSpecifications) > 0 {
		adminOperateTimeout := time.Duration(mqa.appSetting.KafkaSetting.AdminClientOperationTimeoutMs) * time.Millisecond
		results, err := adminClient.CreateTopics(ctx, topicSpecifications, kafka.SetAdminOperationTimeout(adminOperateTimeout))
		if err != nil {
			errDetails := fmt.Sprintf("Failed to create topics %v: %v", topicSpecifications, err)
			o11y.Error(ctx, errDetails)
			logger.Errorf(errDetails)
			span.SetStatus(codes.Error, "Failed to create topics")
			return err
		}

		for _, result := range results {
			logger.Debugf("Topic %s createdResult: %v.", result.Topic, result)
		}
	}

	if len(partitionsSpecifications) > 0 {
		logger.Debugf("Create partitions %v", partitionsSpecifications)

		adminOperateTimeout := time.Duration(mqa.appSetting.KafkaSetting.AdminClientOperationTimeoutMs) * time.Millisecond
		results, err := adminClient.CreatePartitions(ctx, partitionsSpecifications, kafka.SetAdminOperationTimeout(adminOperateTimeout))
		if err != nil {
			errDetails := fmt.Sprintf("Failed to create partitions %v: %v", partitionsSpecifications, err)
			o11y.Error(ctx, errDetails)
			logger.Errorf(errDetails)
			span.SetStatus(codes.Error, "Failed to create partitions")
			return err
		}

		for _, result := range results {
			logger.Debugf("Successfully create partitions for Topic %s createdResult: %v.", result.Topic, result)
		}
	}

	return nil
}

func (mqa *mqAccess) DeleteTopics(ctx context.Context, topicNames []string) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven:删除topic")
	defer span.End()

	adminClient, err := mqa.NewAdminClient()
	if err != nil {
		errDetails := fmt.Sprintf("Failed to create adminClient: %v", err)
		o11y.Error(ctx, errDetails)
		logger.Errorf(errDetails)
		span.SetStatus(codes.Error, "Failed to create adminClient")
		return err
	}
	defer mqa.CloseAdminClient(adminClient)

	dur := time.Duration(mqa.appSetting.KafkaSetting.AdminClientRequestTimeoutMs) * time.Millisecond
	results, err := adminClient.DeleteTopics(ctx, topicNames, kafka.SetAdminOperationTimeout(dur))
	if err != nil {
		errDetails := fmt.Sprintf("Failed to delete topic %s: %s", strings.Join(topicNames, ","), err.Error())
		o11y.Error(ctx, errDetails)
		logger.Errorf(errDetails)
		span.SetStatus(codes.Error, "Failed to delete topic")
		return err
	}

	logger.Infof("Delete topic's result, %v", results)
	return nil
}

// 获取创建配置
func (mqa *mqAccess) GetDefaultConfig(ctx context.Context, adminClient *kafka.AdminClient, resourceName string) (numPartitions, replicationFactor int, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven:获取kafka配置", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	resourceType := kafka.ResourceBroker

	dur := time.Duration(mqa.appSetting.KafkaSetting.AdminClientRequestTimeoutMs) * time.Millisecond
	// Ask cluster for the resource's current configuration
	results, err := adminClient.DescribeConfigs(ctx, []kafka.ConfigResource{
		{
			Type: resourceType,
			Name: resourceName,
		},
	}, kafka.SetAdminRequestTimeout(dur))
	if err != nil {
		detail := fmt.Sprintf("Failed to DescribeConfigs(%s, %s): %s", resourceType, resourceName, err)
		o11y.Error(ctx, detail)
		logger.Errorf(detail)
		span.SetStatus(codes.Error, "Failed to DescribeConfigs")
		return
	}

	for _, result := range results {
		logger.Debugf("kafka defaultConfig %s %s: %s:\n", result.Type, result.Name, result.Error)
		for _, entry := range result.Config {
			switch entry.Name {
			case "num.partitions":
				logger.Debugf("%60s = %-60.60s   %-20s Read-only:%v Sensitive:%v Default:%v",
					entry.Name,
					entry.Value,
					entry.Source,
					entry.IsReadOnly,
					entry.IsSensitive,
					entry.IsDefault,
				)
				numPartitions, err = strconv.Atoi(entry.Value)
				if err != nil {
					detail := fmt.Sprintf("Failed to transfer entry.Name[%s].entry.Value[%s] to int:%s", entry.Name, entry.Value, err.Error())
					o11y.Error(ctx, detail)
					logger.Errorf(detail)
					span.SetStatus(codes.Error, "Failed to transfer entry.Value")
					return
				}
			case "default.replication.factor":
				logger.Debugf("%60s = %-60.60s   %-20s Read-only:%v Sensitive:%v Default:%v",
					entry.Name,
					entry.Value,
					entry.Source,
					entry.IsReadOnly,
					entry.IsSensitive,
					entry.IsDefault,
				)
				replicationFactor, err = strconv.Atoi(entry.Value)
				if err != nil {
					detail := fmt.Sprintf("Failed to transfer entry.Name[%s].entry.Value[%s] to int:%s", entry.Name, entry.Value, err.Error())
					o11y.Error(ctx, detail)
					logger.Errorf(detail)
					span.SetStatus(codes.Error, "Failed to transfer entry.Value")
					return
				}
			}
		}
	}

	span.SetStatus(codes.Ok, "")
	return
}
