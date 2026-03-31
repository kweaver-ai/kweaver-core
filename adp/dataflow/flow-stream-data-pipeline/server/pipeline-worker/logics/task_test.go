package logics

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	libmq "github.com/kweaver-ai/kweaver-go-lib/mq"
	osutil "github.com/opensearch-project/opensearch-go/v2/opensearchutil"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"flow-stream-data-pipeline/common"
	access "flow-stream-data-pipeline/pipeline-worker/drivenadapters"
	"flow-stream-data-pipeline/pipeline-worker/interfaces"
	fmock "flow-stream-data-pipeline/pipeline-worker/interfaces/mock"
)

func NewTestTask(mockCtl *gomock.Controller, mqMock interfaces.MQAccess,
	osaMock interfaces.OpenSearchAccess) *Task {

	taskTest := &Task{
		appSetting: &common.AppSetting{
			MQSetting: libmq.MQSetting{
				Tenant: "default",
			},
		},
		mqAccess:       mqMock,
		osAccess:       osaMock,
		pipelineID:     "1234567890abcdefg",
		inputTopic:     "default.sdp.test.input",
		outputTopic:    "default.mdl.process.test",
		errorTopic:     "default.sdp.test.error",
		status:         "running",
		runningChannel: make(chan bool),

		IndexBaseInfo: &interfaces.IndexBaseInfo{
			BaseType: "test",
			Name:     "test",
			DataType: "test",
			Category: "metric",
		},
		IsDeleted: false,
	}
	return taskTest
}

func TestTask_Run(t *testing.T) {
	Convey("Test Task Run", t, func() {
		ctx := context.Background()

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mqMock := fmock.NewMockMQAccess(mockCtl)
		osaMock := fmock.NewMockOpenSearchAccess(mockCtl)
		taskTest := NewTestTask(mockCtl, mqMock, osaMock)

		Convey("Execute failed", func() {
			patches := ApplyMethodReturn(taskTest, "Execute", errors.New("error"))
			defer patches.Reset()

			err := taskTest.Run(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			patches := ApplyMethodReturn(taskTest, "Execute", nil)
			defer patches.Reset()

			err := taskTest.Run(ctx)
			So(err, ShouldBeNil)
		})
	})
}

func TestTask_Stop(t *testing.T) {
	Convey("Test Task Stop", t, func() {
		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mqMock := fmock.NewMockMQAccess(mockCtl)
		osaMock := fmock.NewMockOpenSearchAccess(mockCtl)
		taskTest := NewTestTask(mockCtl, mqMock, osaMock)

		Convey("task is stopped", func() {
			patches := ApplyMethodReturn(taskTest, "Stop", nil)
			defer patches.Reset()

			err := taskTest.Stop()
			So(err, ShouldBeNil)
		})

		Convey("task is running", func() {
			patches := ApplyMethodReturn(taskTest, "Stop", errors.New("error"))
			defer patches.Reset()

			err := taskTest.Stop()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestTask_Execute(t *testing.T) {
	Convey("Test Task Execute", t, func() {
		ctx := context.Background()
		FailureThreshold = 2
		RetryInterval = 1 * time.Second
		FlushBytes = 10
		// FlushIntervalSec = 0

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mqMock := fmock.NewMockMQAccess(mockCtl)
		osaMock := fmock.NewMockOpenSearchAccess(mockCtl)
		taskTest := NewTestTask(mockCtl, mqMock, osaMock)

		c, _ := kafka.NewConsumer(&kafka.ConfigMap{})

		p, _ := kafka.NewProducer(&kafka.ConfigMap{})

		msg := &kafka.Message{
			Timestamp: time.Now(),
			Value:     []byte("{\"__index_base\":\"1234567890abcdefg\"}"),
		}
		errorTopic := "error"
		failedMsg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &errorTopic},
			Timestamp:      time.Now(),
			Value:          []byte("{\"__index_base\":\"1234567890abcdefg\"}"),
		}

		msgs := []*kafka.Message{msg}
		failedMsgs := []*kafka.Message{failedMsg}

		indexer := &access.BulkIndexer{
			Buf: bytes.NewBuffer(make([]byte, 0, FlushBytes)),
			Aux: make([]byte, 0, 512),
		}

		patches := ApplyMethodReturn(taskTest, "CloseConsumer")
		defer patches.Reset()

		Convey("NewConsumer failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, errors.New("error"))

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("NewTransactionalProducer failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			mqMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, errors.New("error"))

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("Subscribe failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			mqMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)
			patches := ApplyMethodReturn(c, "Subscribe", errors.New("error"))
			defer patches.Reset()

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("DoConsume failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			mqMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)

			patches := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches.Reset()

			osaMock.EXPECT().NewBulkIndexer(FlushBytes).Return(indexer)
			mqMock.EXPECT().DoConsume(c).Return(nil, errors.New("error")).AnyTimes()

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("PackagingMessages failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			mqMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)
			patches1 := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches1.Reset()

			osaMock.EXPECT().NewBulkIndexer(FlushBytes).Return(indexer)
			mqMock.EXPECT().DoConsume(c).Return(msg, nil).AnyTimes()

			patches2 := ApplyMethodReturn(taskTest, "PackagingMessages", msgs, nil, errors.New("error"))
			defer patches2.Reset()

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("indexer.BufLen() >= FlushBytes, FlushMessagesToOS failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			mqMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)

			patches1 := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches1.Reset()

			osaMock.EXPECT().NewBulkIndexer(FlushBytes).Return(indexer)
			mqMock.EXPECT().DoConsume(c).Return(msg, nil).AnyTimes()

			patches2 := ApplyMethodReturn(taskTest, "FlushMessagesToOS", nil, errors.New("error"))
			defer patches2.Reset()

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("indexer.BufLen() >= FlushBytes, FlushMessagesToErrorTopic failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			mqMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)

			patches1 := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches1.Reset()

			osaMock.EXPECT().NewBulkIndexer(FlushBytes).Return(indexer)
			mqMock.EXPECT().DoConsume(c).Return(msg, nil).AnyTimes()

			patches2 := ApplyMethodReturn(taskTest, "FlushMessagesToOS", failedMsgs, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(taskTest, "FlushMessagesToErrorTopic", errors.New("error"))
			defer patches3.Reset()

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("indexer.BufLen() >= FlushBytes, FlushMessagesToOutputTopic failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			mqMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)

			patches1 := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches1.Reset()

			osaMock.EXPECT().NewBulkIndexer(FlushBytes).Return(indexer)
			mqMock.EXPECT().DoConsume(c).Return(msg, nil).AnyTimes()

			patches2 := ApplyMethodReturn(taskTest, "FlushMessagesToOS", nil, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(taskTest, "FlushMessagesToOutputTopic", errors.New("error"))
			defer patches3.Reset()

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("time.Since(lastFlushTime) > FlushIntervalSec, FlushMessagesToOS failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			mqMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)
			patches := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches.Reset()

			FlushBytes = 100000
			osaMock.EXPECT().NewBulkIndexer(FlushBytes).Return(indexer)

			cnt := 0
			mqMock.EXPECT().DoConsume(c).DoAndReturn(
				func(*kafka.Consumer) (*kafka.Message, error) {
					if cnt == 0 {
						cnt++
						return nil, errors.New("error")
					} else {
						return msg, nil
					}
				}).AnyTimes()

			patches2 := ApplyMethodReturn(taskTest, "FlushMessagesToOS", nil, errors.New("error"))
			defer patches2.Reset()

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("time.Since(lastFlushTime) > FlushIntervalSec, FlushMessagesToErrorTopic failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			mqMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)
			patches := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches.Reset()

			FlushBytes = 100000
			osaMock.EXPECT().NewBulkIndexer(FlushBytes).Return(indexer)

			cnt := 0
			mqMock.EXPECT().DoConsume(c).DoAndReturn(
				func(*kafka.Consumer) (*kafka.Message, error) {
					if cnt == 0 {
						cnt++
						return nil, errors.New("error")
					} else {
						return msg, nil
					}
				}).AnyTimes()

			patches2 := ApplyMethodReturn(taskTest, "FlushMessagesToOS", failedMsgs, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(taskTest, "FlushMessagesToErrorTopic", errors.New("error"))
			defer patches3.Reset()

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})

		Convey("time.Since(lastFlushTime) > FlushIntervalSec, FlushMessagesToOutputTopic failed", func() {
			mqMock.EXPECT().NewConsumer(gomock.Any()).Return(c, nil)
			mqMock.EXPECT().NewTransactionalProducer(gomock.Any()).Return(p, nil)
			patches := ApplyMethodReturn(c, "Subscribe", nil)
			defer patches.Reset()

			FlushBytes = 100000
			osaMock.EXPECT().NewBulkIndexer(FlushBytes).Return(indexer)

			cnt := 0
			mqMock.EXPECT().DoConsume(c).DoAndReturn(
				func(*kafka.Consumer) (*kafka.Message, error) {
					if cnt == 0 {
						cnt++
						return nil, errors.New("error")
					} else {
						return msg, nil
					}
				}).AnyTimes()

			patches2 := ApplyMethodReturn(taskTest, "FlushMessagesToOS", nil, nil)
			defer patches2.Reset()

			patches3 := ApplyMethodReturn(taskTest, "FlushMessagesToOutputTopic", errors.New("error"))
			defer patches3.Reset()

			err := taskTest.Execute(ctx)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestTask_PackagingMessages(t *testing.T) {
	Convey("Test Task PackagingMessages", t, func() {

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mqMock := fmock.NewMockMQAccess(mockCtl)
		osaMock := fmock.NewMockOpenSearchAccess(mockCtl)
		taskTest := NewTestTask(mockCtl, mqMock, osaMock)

		indexer := &access.BulkIndexer{
			Buf: bytes.NewBuffer(make([]byte, 0, FlushBytes)),
			Aux: make([]byte, 0, 512),
		}

		Convey("msgs is empty", func() {
			msgs := []*kafka.Message{}
			_, _, err := taskTest.PackagingMessages(msgs, indexer)
			So(err, ShouldBeNil)
		})

		Convey("json format error", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\""),
			}
			msgs := []*kafka.Message{msg}

			_, _, err := taskTest.PackagingMessages(msgs, indexer)
			So(err, ShouldBeNil)
		})

		Convey("PackagingItem failed", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\"}"),
			}
			msgs := []*kafka.Message{msg}

			patches := ApplyMethodReturn(taskTest, "PackagingItem", errors.New("error"))
			defer patches.Reset()

			_, _, err := taskTest.PackagingMessages(msgs, indexer)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\"}"),
			}
			msgs := []*kafka.Message{msg}

			patches := ApplyMethodReturn(taskTest, "PackagingItem", nil)
			defer patches.Reset()

			_, _, err := taskTest.PackagingMessages(msgs, indexer)
			So(err, ShouldBeNil)
		})
	})
}

func TestTask_isValidJson(t *testing.T) {
	Convey("Test Task isValidJson", t, func() {
		Convey("json format error", func() {
			data := []byte(`{"h":"b}`)
			res := isValidJSON(data)
			So(res, ShouldBeFalse)
		})

		Convey("json format ok", func() {
			data := []byte(`{"h":"b"}`)
			res := isValidJSON(data)
			So(res, ShouldBeTrue)
		})
	})
}

func TestTask_processJSON(t *testing.T) {
	Convey("Test Task processJSON", t, func() {
		Convey("process json ok", func() {
			testData1 := []byte(`{"key": "value"}`)
			result1, err := processJSON(testData1)

			So(err, ShouldBeNil)
			So(string(result1), ShouldEqual, string(testData1))

		})

		Convey("process json error", func() {
			testData2 := []byte("not a valid json")
			result2, err := processJSON(testData2)

			expected := map[string]any{"message": string(testData2)}
			expectedRes, _ := sonic.Marshal(expected)

			So(err, ShouldBeNil)
			So(string(result2), ShouldEqual, string(expectedRes))
		})
	})
}

func TestTask_PackagingOSMessage(t *testing.T) {
	Convey("Test Task PackagingOSMessage", t, func() {

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mqMock := fmock.NewMockMQAccess(mockCtl)
		osaMock := fmock.NewMockOpenSearchAccess(mockCtl)
		taskTest := NewTestTask(mockCtl, mqMock, osaMock)

		Convey("json format error", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value:     []byte("{\"__index_base\":\"1234567890abcdefg\""),
			}

			_, err := taskTest.PackagingOSMessage(msg)
			So(err, ShouldNotBeNil)
		})

		Convey("success, category is null", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value: []byte(`
				{
					"__data_type":"1234567890abcdefg",
					"__index_base":"1234567890abcdefg",
					"__category": ""
				}`),
			}

			_, err := taskTest.PackagingOSMessage(msg)
			So(err, ShouldBeNil)
		})

		Convey("success, category is metric", func() {
			msg := &kafka.Message{
				Timestamp: time.Now(),
				Value: []byte(`
				{
					"__data_type":"1234567890abcdefg",
					"__index_base":"1234567890abcdefg",
					"__category": "metric"
				}`),
			}

			patches := ApplyMethodReturn(taskTest, "ProcessMetricMessage", nil, nil)
			defer patches.Reset()

			_, err := taskTest.PackagingOSMessage(msg)
			So(err, ShouldBeNil)
		})
	})
}

func TestTask_ProcessMetricMessageRouting(t *testing.T) {
	Convey("Test Task ProcessMetricMessageRouting", t, func() {

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mqMock := fmock.NewMockMQAccess(mockCtl)
		osaMock := fmock.NewMockOpenSearchAccess(mockCtl)
		taskTest := NewTestTask(mockCtl, mqMock, osaMock)

		Convey("@timestamp format error", func() {
			root, _ := sonic.Get([]byte(
				`{
					"@timestamp":"1234567890abcdefg"
				}`,
			))

			routing := taskTest.ProcessMetricMessageRouting(&root)
			So(routing, ShouldBeEmpty)
		})

		Convey("@timestamp format is RFC3339", func() {
			root, _ := sonic.Get([]byte(
				`{
					"@timestamp":"2024-06-17T09:56:00.000+08:00"
				}`,
			))

			routing := taskTest.ProcessMetricMessageRouting(&root)
			So(routing, ShouldBeEmpty)
		})

		Convey("success", func() {
			root, _ := sonic.Get([]byte(
				`{
					"@timestamp":1718585760000
				}`,
			))

			routing := taskTest.ProcessMetricMessageRouting(&root)
			So(routing, ShouldEqual, "238692")
		})
	})
}

func TestTask_ProcessMetricMessageLabelsStr(t *testing.T) {
	Convey("Test Task ProcessMetricMessageLabelsStr", t, func() {

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mqMock := fmock.NewMockMQAccess(mockCtl)
		osaMock := fmock.NewMockOpenSearchAccess(mockCtl)
		taskTest := NewTestTask(mockCtl, mqMock, osaMock)

		Convey("labels does not exist and prometheus does not exist", func() {
			root, _ := sonic.Get([]byte(
				`{}`,
			))

			labels_str := taskTest.ProcessMetricMessageLabelsStr(&root)
			So(labels_str, ShouldBeEmpty)
		})

		Convey("labels does not exist and prometheus is not an object", func() {
			root, _ := sonic.Get([]byte(
				`{
					"prometheus":"123"
				}`,
			))

			labels_str := taskTest.ProcessMetricMessageLabelsStr(&root)
			So(labels_str, ShouldBeEmpty)
		})

		Convey("labels does not exist and prometheus.labels does not exist", func() {
			root, _ := sonic.Get([]byte(
				`{
					"prometheus":{}
				}`,
			))

			labels_str := taskTest.ProcessMetricMessageLabelsStr(&root)
			So(labels_str, ShouldBeEmpty)
		})

		Convey("labels is not an object", func() {
			root, _ := sonic.Get([]byte(
				`{
					"labels": "123"
				}`,
			))

			labels_str := taskTest.ProcessMetricMessageLabelsStr(&root)
			So(labels_str, ShouldBeEmpty)
		})

		Convey("labels does not exist and prometheus.labels is not an object", func() {
			root, _ := sonic.Get([]byte(
				`{
					"prometheus": {
						"labels": "123"
					}
				}`,
			))

			labels_str := taskTest.ProcessMetricMessageLabelsStr(&root)
			So(labels_str, ShouldBeEmpty)
		})

		Convey("SortKeys failed", func() {
			root, _ := sonic.Get([]byte(
				`{
					"prometheus":{
						"labels": {
							"name": "123",
							"host": "234",
							"aaa": 123
						}
					}
				}`,
			))

			patches := ApplyMethodReturn(&root, "SortKeys", errors.New("error"))
			defer patches.Reset()

			labels_str := taskTest.ProcessMetricMessageLabelsStr(&root)
			So(labels_str, ShouldBeEmpty)
		})

		Convey("Properties failed", func() {
			root, _ := sonic.Get([]byte(
				`{
					"labels": {
						"name": "123",
						"host": "234",
						"aaa": 123
					}
				}`,
			))

			patches := ApplyMethodReturn(&root, "Properties", ast.ObjectIterator{}, errors.New("error"))
			defer patches.Reset()

			labels_str := taskTest.ProcessMetricMessageLabelsStr(&root)
			So(labels_str, ShouldBeEmpty)
		})

		Convey("success", func() {
			root, _ := sonic.Get([]byte(
				`{
					"labels": {
						"name": "123",
						"host": "234",
						"aaa": 123
					}
				}`,
			))

			labels_str := taskTest.ProcessMetricMessageLabelsStr(&root)
			So(labels_str, ShouldEqual, "aaa=123,host=234,name=123")
		})
	})
}

func TestTask_ProcessMetricMessage(t *testing.T) {
	Convey("Test Task TestTask_ProcessMetricMessage", t, func() {

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mqMock := fmock.NewMockMQAccess(mockCtl)
		osaMock := fmock.NewMockOpenSearchAccess(mockCtl)
		taskTest := NewTestTask(mockCtl, mqMock, osaMock)

		Convey("success", func() {
			root, _ := sonic.Get([]byte(
				`{
					"@timestamp":"2024-06-17T09:56:00.000+08:00",
					"labels": {
						"name": "123",
						"host": "234",
						"aaa": 123
					}
				}`,
			))

			patches1 := ApplyMethodReturn(taskTest, "ProcessMetricMessageRouting", "954771")
			defer patches1.Reset()

			patches2 := ApplyMethodReturn(taskTest, "ProcessMetricMessageLabelsStr", "aaa=123,host=234,name=123")
			defer patches2.Reset()

			routingPtr, err := taskTest.ProcessMetricMessage(&root)
			So(err, ShouldBeNil)
			So(routingPtr, ShouldNotBeNil)
		})
	})
}

func TestTask_PackagingItem(t *testing.T) {
	Convey("Test Task PackagingMessage", t, func() {

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mqMock := fmock.NewMockMQAccess(mockCtl)
		osaMock := fmock.NewMockOpenSearchAccess(mockCtl)
		taskTest := NewTestTask(mockCtl, mqMock, osaMock)

		indexer := &access.BulkIndexer{
			Buf: bytes.NewBuffer(make([]byte, 0, FlushBytes)),
			Aux: make([]byte, 0, 512),
		}
		item := &osutil.BulkIndexerItem{}

		Convey("indexer.Add failed", func() {
			patches1 := ApplyMethodReturn(indexer, "Add", errors.New("error"))
			defer patches1.Reset()

			err := taskTest.PackagingItem(item, indexer)
			So(err, ShouldNotBeNil)
		})

		Convey("success", func() {
			patches1 := ApplyMethodReturn(indexer, "Add", nil)
			defer patches1.Reset()

			err := taskTest.PackagingItem(item, indexer)
			So(err, ShouldBeNil)
		})
	})
}

func TestTask_FlushMessages(t *testing.T) {
	Convey("Test Task FlushMessages", t, func() {
		ctx := context.Background()
		FailureThreshold = 2

		mockCtl := gomock.NewController(t)
		defer mockCtl.Finish()

		mqMock := fmock.NewMockMQAccess(mockCtl)
		osaMock := fmock.NewMockOpenSearchAccess(mockCtl)
		taskTest := NewTestTask(mockCtl, mqMock, osaMock)

		kp := &interfaces.KafkaProducer{}

		indexer := &access.BulkIndexer{
			Buf: bytes.NewBuffer(make([]byte, 0, FlushBytes)),
			Aux: make([]byte, 0, 512),
		}
		msg := &kafka.Message{
			Timestamp: time.Now(),
			Value:     []byte("{\"__index_base\":\"1234567890abcdefg\"}"),
		}

		docIDToKafkaMsg := map[string]*kafka.Message{
			"123": msg,
		}
		// msgs := []*kafka.Message{msg}
		_, _ = taskTest.PackagingOSMessage(msg)

		Convey("indexer.Flush failed", func() {
			patches := ApplyMethodReturn(indexer, "Flush", []string{"123"}, errors.New("error"))
			defer patches.Reset()

			patches1 := ApplyMethodReturn(taskTest, "FlushMessagesToErrorTopic", nil)
			defer patches1.Reset()

			_, err := taskTest.FlushMessagesToOS(ctx, kp, indexer, docIDToKafkaMsg)
			So(err, ShouldBeNil)
		})

		// Convey("kAccess.DoCommit failed", func() {
		// 	patches := ApplyMethodReturn(indexer, "Flush", nil)
		// 	defer patches.Reset()

		// 	mqMock.EXPECT().DoCommit(gomock.Any()).Return(errors.New("error")).AnyTimes()

		// 	err := taskTest.FlushMessagesToOS(ctx, c, indexer)
		// 	So(err, ShouldNotBeNil)
		// })

		Convey("success", func() {
			patches := ApplyMethodReturn(indexer, "Flush", []string{}, nil)
			defer patches.Reset()

			// mqMock.EXPECT().DoCommit(gomock.Any()).Return(nil)

			_, err := taskTest.FlushMessagesToOS(ctx, kp, indexer, docIDToKafkaMsg)
			So(err, ShouldBeNil)
		})
	})
}
