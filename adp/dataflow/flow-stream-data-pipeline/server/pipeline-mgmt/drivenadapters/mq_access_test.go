package drivenadapters

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	. "github.com/smartystreets/goconvey/convey"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

func MockNewMQAccess() *mqAccess {
	mqa := &mqAccess{
		appSetting: &common.AppSetting{
			KafkaSetting: common.KafkaSetting{
				Retries:                     5,
				SocketTimeoutMs:             60000,
				AdminClientRequestTimeoutMs: 30000,
				RetentionTime:               "7d",
				RetentionSize:               -1,
			},
		},
	}
	return mqa
}

func Test_MQAccess_NewAdminClient(t *testing.T) {
	Convey("create new adminClient", t, func() {
		mqa := MockNewMQAccess()

		Convey("Failed to create adminClient", func() {
			patches := ApplyFuncReturn(kafka.NewAdminClient, nil, errors.New("error"))
			defer patches.Reset()

			adminClient, err := mqa.NewAdminClient()
			So(adminClient, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("Successfully create adminClient", func() {
			patches := ApplyFuncReturn(kafka.NewAdminClient, &kafka.AdminClient{}, nil)
			defer patches.Reset()

			adminClient, err := mqa.NewAdminClient()
			So(adminClient, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}

func Test_MQAccess_CreateTopic(t *testing.T) {
	Convey("create Topics", t, func() {
		mqa := MockNewMQAccess()
		adminClient, _ := mqa.NewAdminClient()
		inputTopic := fmt.Sprintf(interfaces.TopicInputName, "default", "test")
		errorTopic := fmt.Sprintf(interfaces.TopicErrorName, "default", "test")
		topicNames := []string{inputTopic, errorTopic}

		Convey("Failed to create adminClient", func() {
			patches := ApplyFuncReturn(kafka.NewAdminClient, nil, errors.New("error"))
			defer patches.Reset()

			err := mqa.CreateTopicsOrPartitions(testCtx, topicNames)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to get metadata", func() {
			patches1 := ApplyFuncReturn(kafka.NewAdminClient, adminClient, nil)
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(adminClient, "GetMetadata", nil, errors.New("error"))
			defer patches2.Reset()

			err := mqa.CreateTopicsOrPartitions(testCtx, topicNames)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to GetDefaultConfig", func() {
			patches1 := ApplyFuncReturn(kafka.NewAdminClient, adminClient, nil)
			defer patches1.Reset()

			metadata := &kafka.Metadata{
				Brokers: []kafka.BrokerMetadata{{
					ID:   1,
					Host: "",
					Port: 8080,
				}},
				Topics: map[string]kafka.TopicMetadata{},
			}

			patches2 := ApplyMethodReturn(adminClient, "GetMetadata", metadata, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(mqa, "GetDefaultConfig", 1, 1, errors.New("error"))
			defer patches3.Reset()

			err := mqa.CreateTopicsOrPartitions(testCtx, topicNames)
			So(err, ShouldNotBeNil)
		})

		Convey("invalid MQRetentionTime", func() {
			patches1 := ApplyFuncReturn(kafka.NewAdminClient, adminClient, nil)
			defer patches1.Reset()

			metadata := &kafka.Metadata{
				Brokers: []kafka.BrokerMetadata{{
					ID:   1,
					Host: "",
					Port: 8080,
				}},
				Topics: map[string]kafka.TopicMetadata{},
			}

			patches2 := ApplyMethodReturn(adminClient, "GetMetadata", metadata, nil)
			defer patches2.Reset()

			patches := ApplyMethodReturn(mqa, "GetDefaultConfig", 1, 1, nil)
			defer patches.Reset()

			mqa.appSetting.KafkaSetting.RetentionTime = "ad"

			err := mqa.CreateTopicsOrPartitions(testCtx, topicNames)
			So(err, ShouldNotBeNil)
		})

		Convey("invalid MQRetentionTime Unit", func() {
			patches1 := ApplyFuncReturn(kafka.NewAdminClient, adminClient, nil)
			defer patches1.Reset()

			metadata := &kafka.Metadata{
				Brokers: []kafka.BrokerMetadata{{
					ID:   1,
					Host: "",
					Port: 8080,
				}},
				Topics: map[string]kafka.TopicMetadata{},
			}

			patches2 := ApplyMethodReturn(adminClient, "GetMetadata", metadata, nil)
			defer patches2.Reset()

			patches := ApplyMethodReturn(mqa, "GetDefaultConfig", 1, 1, nil)
			defer patches.Reset()

			mqa.appSetting.KafkaSetting.RetentionTime = "1f"

			err := mqa.CreateTopicsOrPartitions(testCtx, topicNames)
			So(err, ShouldNotBeNil)
		})

		Convey("adminclient CreateTopics failed", func() {
			patches1 := ApplyFuncReturn(kafka.NewAdminClient, adminClient, nil)
			defer patches1.Reset()

			metadata := &kafka.Metadata{
				Brokers: []kafka.BrokerMetadata{{
					ID:   1,
					Host: "",
					Port: 8080,
				}},
				Topics: map[string]kafka.TopicMetadata{},
			}

			patches2 := ApplyMethodReturn(adminClient, "GetMetadata", metadata, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(mqa, "GetDefaultConfig", 1, 1, nil)
			defer patches3.Reset()

			mqa.appSetting.KafkaSetting.RetentionTime = "1d"

			patches4 := ApplyMethodReturn(adminClient, "CreateTopics", []kafka.TopicResult{}, errors.New("error"))
			defer patches4.Reset()

			err := mqa.CreateTopicsOrPartitions(testCtx, topicNames)
			So(err, ShouldNotBeNil)
		})

		Convey("Successfully to create topics", func() {
			patches1 := ApplyFuncReturn(kafka.NewAdminClient, adminClient, nil)
			defer patches1.Reset()

			metadata := &kafka.Metadata{
				Brokers: []kafka.BrokerMetadata{{
					ID:   1,
					Host: "",
					Port: 8080,
				}},
				Topics: map[string]kafka.TopicMetadata{},
			}

			patches2 := ApplyMethodReturn(adminClient, "GetMetadata", metadata, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(mqa, "GetDefaultConfig", 1, 1, nil)
			defer patches3.Reset()

			mqa.appSetting.KafkaSetting.RetentionTime = "1d"

			patches4 := ApplyMethodReturn(adminClient, "CreateTopics", []kafka.TopicResult{}, nil)
			defer patches4.Reset()

			err := mqa.CreateTopicsOrPartitions(testCtx, topicNames)
			So(err, ShouldBeNil)
		})

		Convey("Successfully to create partitions", func() {
			patches1 := ApplyFuncReturn(kafka.NewAdminClient, adminClient, nil)
			defer patches1.Reset()

			metadata := &kafka.Metadata{
				Brokers: []kafka.BrokerMetadata{{
					ID:   1,
					Host: "",
					Port: 8080,
				}},
				Topics: map[string]kafka.TopicMetadata{
					inputTopic: {Topic: inputTopic, Partitions: []kafka.PartitionMetadata{{ID: 0}}},
					errorTopic: {Topic: errorTopic, Partitions: []kafka.PartitionMetadata{{ID: 0}}},
				},
			}

			patches2 := ApplyMethodReturn(adminClient, "GetMetadata", metadata, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(mqa, "GetDefaultConfig", 2, 1, nil)
			defer patches3.Reset()

			mqa.appSetting.KafkaSetting.RetentionTime = "1d"

			patches4 := ApplyMethodReturn(adminClient, "CreatePartitions", []kafka.TopicResult{}, nil)
			defer patches4.Reset()

			err := mqa.CreateTopicsOrPartitions(testCtx, topicNames)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_DeleteTopic(t *testing.T) {
	Convey("Test DeleteTopic", t, func() {
		mqa := MockNewMQAccess()
		adminClient, _ := mqa.NewAdminClient()

		inputTopic := fmt.Sprintf(interfaces.TopicInputName, "default", "test")
		errorTopic := fmt.Sprintf(interfaces.TopicErrorName, "default", "test")
		topicNames := []string{inputTopic, errorTopic}

		Convey("Failed to delete topics", func() {
			patches1 := ApplyFuncReturn(kafka.NewAdminClient, adminClient, nil)
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(adminClient, "DeleteTopics", []kafka.TopicResult{}, errors.New("error"))
			defer patches2.Reset()

			err := mqa.DeleteTopics(testCtx, topicNames)
			So(err, ShouldNotBeNil)
		})

		Convey("Successfully to delete topics", func() {
			patches1 := ApplyFuncReturn(kafka.NewAdminClient, adminClient, nil)
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(adminClient, "DeleteTopics", []kafka.TopicResult{}, nil)
			defer patches2.Reset()

			err := mqa.DeleteTopics(testCtx, topicNames)
			So(err, ShouldBeNil)
		})
	})
}

func Test_KafkaAccess_GetDefaultConfig(t *testing.T) {
	Convey("Test GetDefaultConfig", t, func() {
		mqa := MockNewMQAccess()

		adminClient, _ := mqa.NewAdminClient()

		inputTopic := fmt.Sprintf(interfaces.TopicInputName, "default", "test")
		errorTopic := fmt.Sprintf(interfaces.TopicErrorName, "default", "test")

		metadata := &kafka.Metadata{
			Brokers: []kafka.BrokerMetadata{{
				ID:   1,
				Host: "",
				Port: 8080,
			}},
			Topics: map[string]kafka.TopicMetadata{
				inputTopic: {Topic: inputTopic, Partitions: []kafka.PartitionMetadata{{ID: 0}}},
				errorTopic: {Topic: errorTopic, Partitions: []kafka.PartitionMetadata{{ID: 0}}},
			},
		}

		crResult := []kafka.ConfigResourceResult{
			{
				Type: kafka.ResourceBroker,
				Name: "1",
				Config: map[string]kafka.ConfigEntryResult{
					"num.partitions": {
						Name:   "num.partitions",
						Value:  "1",
						Source: kafka.ConfigSourceDefault,
					},
					"default.replication.factor": {
						Name:   "default.replication.factor",
						Value:  "1",
						Source: kafka.ConfigSourceDefault,
					},
				},
			},
		}

		Convey("Failed to DescribeConfigs", func() {
			patches1 := ApplyMethodReturn(adminClient, "GetMetadata", metadata, nil)
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(adminClient, "DescribeConfigs", []kafka.ConfigResourceResult{}, errors.New("error"))
			defer patches2.Reset()

			_, _, err := mqa.GetDefaultConfig(testCtx, adminClient, "broker")
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to atoi num.partitions Value", func() {
			patches1 := ApplyMethodReturn(adminClient, "GetMetadata", metadata, nil)
			defer patches1.Reset()

			crResult[0].Config["num.partitions"] = kafka.ConfigEntryResult{
				Name:   "num.partitions",
				Value:  "a",
				Source: kafka.ConfigSourceDefault,
			}
			patches2 := ApplyMethodReturn(adminClient, "DescribeConfigs", crResult, nil)
			defer patches2.Reset()

			_, _, err := mqa.GetDefaultConfig(testCtx, adminClient, "broker")
			So(err, ShouldNotBeNil)
		})

		Convey("Failed to atoi default.replication.factor Value", func() {
			patches1 := ApplyMethodReturn(adminClient, "GetMetadata", metadata, nil)
			defer patches1.Reset()

			crResult[0].Config["default.replication.factor"] = kafka.ConfigEntryResult{
				Name:   "default.replication.factor",
				Value:  "a",
				Source: kafka.ConfigSourceDefault,
			}
			patches2 := ApplyMethodReturn(adminClient, "DescribeConfigs", crResult, nil)
			defer patches2.Reset()

			_, _, err := mqa.GetDefaultConfig(testCtx, adminClient, "broker")
			So(err, ShouldNotBeNil)
		})

		Convey("Successfully to GetDefaultConfig", func() {
			patches1 := ApplyMethodReturn(adminClient, "GetMetadata", metadata, nil)
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(adminClient, "DescribeConfigs", crResult, nil)
			defer patches2.Reset()

			_, _, err := mqa.GetDefaultConfig(testCtx, adminClient, "broker")
			So(err, ShouldBeNil)
		})
	})
}
