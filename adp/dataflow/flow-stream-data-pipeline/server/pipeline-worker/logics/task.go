// task 从 input_topic 消费数据，并写入到 opensearch 和 output_topic
package logics

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	libcomm "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	osutil "github.com/opensearch-project/opensearch-go/v2/opensearchutil"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-worker/interfaces"
)

type Task struct {
	appSetting *common.AppSetting
	mqAccess   interfaces.MQAccess
	osAccess   interfaces.OpenSearchAccess

	pipelineID         string
	inputTopic         string
	outputTopic        string
	errorTopic         string
	status             string
	runningChannel     chan bool
	IndexBaseInfo      *interfaces.IndexBaseInfo
	useIndexBaseInData bool
	IsDeleted          bool
	delChan            chan bool
}

func NewTask(appSetting *common.AppSetting, pipeline *interfaces.Pipeline, indexBaseInfo *interfaces.IndexBaseInfo) *Task {
	return &Task{
		appSetting:         appSetting,
		mqAccess:           MQAccess,
		osAccess:           OSAccess,
		pipelineID:         pipeline.PipelineID,
		inputTopic:         pipeline.InputTopic,
		outputTopic:        pipeline.OutputTopic,
		errorTopic:         pipeline.ErrorTopic,
		status:             interfaces.TaskStatus_Running,
		runningChannel:     make(chan bool),
		IndexBaseInfo:      indexBaseInfo,
		useIndexBaseInData: pipeline.UseIndexBaseInData,
		IsDeleted:          false,
		delChan:            make(chan bool),
	}
}

func GenerateOutputTopicName(tenant string, BaseType string) string {
	return fmt.Sprintf(interfaces.TopicOutputName, tenant, BaseType)
}

func (t *Task) String() string {
	return fmt.Sprintf("{task_pipeline_id = %s, task_input_topic = %s, task_output_topic = %s, "+
		"task_error_topic = %s, task_status = %s}", t.pipelineID, t.inputTopic, t.outputTopic, t.errorTopic, t.status)
}

func (t *Task) Run(ctx context.Context) error {
	logger.Infof("Start task %s", t)
	err := t.Execute(ctx)
	if err != nil {
		t.status = interfaces.TaskStatus_Error
		logger.Errorf("failed to Execute task %s, error: %v", t, err)
	}
	return err
}

// 停止 task
func (t *Task) Stop() error {
	switch tStatus := t.status; tStatus {
	case interfaces.TaskStatus_Stopped:
		logger.Infof("Stop a stopped task %s", t)
		return fmt.Errorf("Stop a stopped task")
	default:
		logger.Debugf("Task %s will be set to stopping status", t)
		t.status = interfaces.TaskStatus_Stopping
	}

	// 等待 executeTask 的协程结束
	<-t.runningChannel

	t.status = interfaces.TaskStatus_Stopped
	logger.Infof("Task %s has stopped", t)
	return nil
}

func (t *Task) CloseConsumer(c *kafka.Consumer) {
	go c.Close()
}

func (t *Task) CloseProducer(kp *interfaces.KafkaProducer) {
	// flush in-flight msg within queue.
	i := kp.Producer.Flush(3000)
	if i > 0 {
		logger.Warnf("There are still %d un-flushed outstanding events on task %s", i, t)
	}

	close(kp.DeliveryChan)
	logger.Debugf("Delivery channel for producer %v closed on task %s", kp.Producer, t)

	kp.Producer.Close()
	logger.Debugf("Producer %v closed on task %s", kp.Producer, t)
}

func (t *Task) Execute(ctx context.Context) error {
	groupID := ComsumerGroupID(t.appSetting.MQSetting.Tenant, t.pipelineID)
	c, err := t.mqAccess.NewConsumer(groupID)
	if err != nil {
		logger.Errorf("task failed to NewConsumer: %v", err)
		return err
	}
	defer func() {
		// 异步关闭consumer
		// 30s后kafka报：Connection setup timed out in status APIVERSION_QUERY (after 30000ms in status APIVERSION_QUERY)
		t.CloseConsumer(c)
	}()

	// 生产者发送消息使用的 delivery channel
	deliveryChan := make(chan kafka.Event, t.appSetting.ServerSetting.FlushItems)

	// 创建生产者
	txId := fmt.Sprintf("%s_%s", t.pipelineID, t.inputTopic)
	p, err := t.mqAccess.NewTransactionalProducer(txId)
	if err != nil {
		return err
	}

	kp := &interfaces.KafkaProducer{
		Producer:     p,
		DeliveryChan: deliveryChan,
	}
	// 函数结束前关闭生产者和 delivery channel
	defer t.CloseProducer(kp)

	// 订阅 topic
	err = c.Subscribe(t.inputTopic, nil)
	if err != nil {
		logger.Errorf("task failed to Subscribe: %v", err)
		return err
	}

	indexer := t.osAccess.NewBulkIndexer(FlushBytes)

	// 初始化事务
	txnOperateTimeout := time.Duration(t.appSetting.KafkaSetting.TransactionTimeoutMs) * time.Millisecond
	txnCtx, cancel := context.WithTimeout(context.Background(), txnOperateTimeout)
	defer cancel()
	err = p.InitTransactions(txnCtx)
	if err != nil {
		logger.Errorf("Task failed to initTransaction on, %v", err)
		return err
	}

	lastFlushTime := time.Now()
	currentMsgs := make([]*kafka.Message, 0, FlushItems+1)
	currentBufLen := 0

	cnt := 0
	for {
		msg, err := t.mqAccess.DoConsume(c)
		if err != nil {
			cnt++
			if cnt >= FailureThreshold {
				logger.Errorf("task failed to DoConsume, need to restart, retry[%d]: %v", cnt, err)
				return err
			}

			logger.Errorf("task failed to DoConsume: wait to retry[%d]: %v", cnt, err)
			time.Sleep(RetryInterval)
			continue
		}

		if msg != nil {
			currentMsgs = append(currentMsgs, msg)
			currentBufLen += len(msg.Value)
			if currentBufLen >= FlushBytes || len(currentMsgs) >= FlushItems {
				mqItems, docIDToKafkaMsg, err := t.PackagingMessages(currentMsgs, indexer)
				if err != nil {
					return err
				}

				currentMsgs = make([]*kafka.Message, 0, FlushItems+1)
				currentBufLen = 0

				failedMsgs, err := t.FlushMessagesToOS(ctx, kp, indexer, docIDToKafkaMsg)
				if err != nil {
					return err
				}

				if len(failedMsgs) > 0 {
					// 将失败的数据写入 error topic
					err = t.FlushMessagesToErrorTopic(failedMsgs, c, kp)
					if err != nil {
						logger.Errorf("failed to write to error topic, error: %v", err)
						return err
					}

					logger.Infof("success to write %d messages to error topic", len(failedMsgs))

				} else {

					err = t.FlushMessagesToOutputTopic(mqItems, c, kp)
					if err != nil {
						return err
					}
				}

				lastFlushTime = time.Now()

				if t.IsDeleted {
					return nil
				}
			}
		}

		if time.Since(lastFlushTime) > FlushInterval {
			if len(currentMsgs) > 0 {
				mqItems, docIDToKafkaMsg, err := t.PackagingMessages(currentMsgs, indexer)
				if err != nil {
					return err
				}

				currentMsgs = make([]*kafka.Message, 0, FlushItems+1)
				currentBufLen = 0

				failedMsgs, err := t.FlushMessagesToOS(ctx, kp, indexer, docIDToKafkaMsg)
				if err != nil {
					return err
				}

				if len(failedMsgs) > 0 {
					// 将失败的数据写入 error topic
					err = t.FlushMessagesToErrorTopic(failedMsgs, c, kp)
					if err != nil {
						logger.Errorf("failed to write to error topic, error: %v", err)
						return err
					}

					logger.Infof("success to write %d messages to error topic", len(failedMsgs))

				} else {

					err = t.FlushMessagesToOutputTopic(mqItems, c, kp)
					if err != nil {
						return err
					}
				}
			}

			lastFlushTime = time.Now()

			if t.IsDeleted {
				return nil
			}
		}
	}
}

func (t *Task) PackagingMessages(msgs []*kafka.Message, indexer interfaces.BulkIndexer) ([]*kafka.Message, map[string]*kafka.Message, error) {
	if len(msgs) == 0 {
		return []*kafka.Message{}, make(map[string]*kafka.Message), nil
	}

	osItems := make([]*osutil.BulkIndexerItem, 0, len(msgs))
	mqItems := make([]*kafka.Message, 0, len(msgs))
	docIDToKafkaMsg := make(map[string]*kafka.Message)
	mu := sync.Mutex{}

	var wg sync.WaitGroup
	for _, msg := range msgs {
		wg.Add(1)
		msg := msg
		_ = PackagePool.Submit(func() {
			processedValue, _ := processJSON(msg.Value)
			msg.Value = processedValue

			osItem, _ := t.PackagingOSMessage(msg)
			mqItem, _ := t.packagingMQMessageToOutputTopic(msg)

			mu.Lock()
			if osItem != nil {
				osItems = append(osItems, osItem)
				docIDToKafkaMsg[osItem.DocumentID] = msg
			}
			if mqItem != nil {
				mqItems = append(mqItems, mqItem)
			}
			mu.Unlock()

			wg.Done()
		})
	}
	wg.Wait()

	for _, item := range osItems {
		err := t.PackagingItem(item, indexer)
		if err != nil {
			return nil, nil, err
		}
	}

	return mqItems, docIDToKafkaMsg, nil
}

// 检查输入的字节切片是否为有效的 JSON
func isValidJSON(data []byte) bool {
	return sonic.Valid(data)
}

// 处理输入的字节切片
func processJSON(data []byte) ([]byte, error) {
	if isValidJSON(data) {
		return data, nil
	}

	// 构造新的 JSON
	newJSON := map[string]any{
		"message": string(data),
	}

	return sonic.Marshal(newJSON)
}

// 处理 msg，后续写入 opensearch
func (t *Task) PackagingOSMessage(msg *kafka.Message) (*osutil.BulkIndexerItem, error) {
	root, err := sonic.Get(msg.Value)
	if err != nil {
		logger.Errorf("failed to get sonic root: %v, raw data: %s", err, msg.Value)
		return nil, err
	}

	// 将管道 id 写入到数据里
	_, _ = root.Set("__pipeline_id", ast.NewString(t.pipelineID))

	base_type, _ := root.Get("__index_base").String()
	if base_type == "" {
		base_type = t.IndexBaseInfo.BaseType
		_, _ = root.Set("__index_base", ast.NewString(base_type))
	}
	// } else if base_type != t.IndexBaseInfo.BaseType {
	// base type 不一致比较常见，不打印日志，避免太多
	// logger.Warnf("__base_type '%s' is not match with IndexBase: %s", base_type, t.IndexBaseInfo.Name)
	// }

	// 如果useIndexBaseInData 为 true，使用数据里的 __index_base
	var indexAlias string
	if t.useIndexBaseInData {
		indexAlias = fmt.Sprintf("mdl-%s", base_type)
		// 重置 output topic
		t.outputTopic = fmt.Sprintf(interfaces.TopicOutputName, t.appSetting.MQSetting.Tenant, base_type)
	} else {
		indexAlias = fmt.Sprintf("mdl-%s", t.IndexBaseInfo.BaseType)
	}

	data_type, _ := root.Get("__data_type").String()
	if data_type == "" {
		data_type = t.IndexBaseInfo.DataType
		_, _ = root.Set("__data_type", ast.NewString(data_type))
	}
	// } else if data_type != t.IndexBaseInfo.DataType {
	// logger.Warnf("__data_type '%s' is not match with IndexBase: %s", data_type, t.IndexBaseInfo.Name)
	// }

	write_time := msg.Timestamp.UnixMilli()
	_, _ = root.Set("__write_time", ast.NewAny(write_time))
	timestamp, _ := root.Get("@timestamp").String()
	if timestamp == "" {
		_, _ = root.Set("@timestamp", ast.NewAny(write_time))
	} else if t, err := time.Parse(libcomm.RFC3339Milli, timestamp); err == nil {
		// @timestmap is a RFC3339Milli string
		_, _ = root.Set("@timestamp", ast.NewAny(t.UnixMilli()))
	} else if _, err := root.Get("@timestamp").Int64(); err != nil {
		// @timestmap is not int64 or RFC3339Milli string
		_, _ = root.Set("@timestamp", ast.NewAny(write_time))
	}

	// TODO 如果写入到数据里的索引库，获取所有索引库信息存入缓存，以此补齐数据里的category信息
	category, _ := root.Get("__category").String()
	if category == "" {
		category, _ = root.Get("category").String()

		if category == "" {
			category = t.IndexBaseInfo.Category
		} else if category != t.IndexBaseInfo.Category {
			// TODO 日志打印有点多啊
			logger.Warnf("__category '%s' is not match with IndexBase: %s", category, t.IndexBaseInfo.Name)
		}

		_, _ = root.Set("__category", ast.NewString(category))
	}

	var routing *string = nil
	switch category {
	case interfaces.CATEGORY_METRIC:
		routing, _ = t.ProcessMetricMessage(&root)
	case interfaces.CATEGORY_TRACE:
		routing, _ = t.ProcessTraceMessage(&root)
	}

	// 如果数据里有id, 用数据里的，没有id, 就生成一个
	docid, _ := root.Get("__id").String()
	if docid == "" {
		docid = fmt.Sprintf("%s[%d]@%s-%d", base_type,
			msg.TopicPartition.Partition, msg.TopicPartition.Offset, msg.Timestamp.UnixMilli())
		//logger.Debugf("docid: %s", docid)
		// 将 docid 写入到数据里
		_, _ = root.Set("__id", ast.NewString(docid))
	}

	// 可以支持数据里有换行符，不支持格式上有换行符
	// buf, _ := sonic.Marshal(&root)
	// str := strings.ReplaceAll(string(buf), "\\n", "\\\\n")

	// 使用 json.Marshal 处理数据里和格式上的特殊字符
	buf, _ := json.Marshal(&root)

	RequireAlias := true
	item := osutil.BulkIndexerItem{
		Index:        indexAlias,
		Action:       "index",
		DocumentID:   docid,
		Routing:      routing,
		Body:         strings.NewReader(string(buf)),
		RequireAlias: &RequireAlias,
	}

	return &item, nil
}

// 处理 msg, 写入 kafka output topic
func (t *Task) packagingMQMessageToOutputTopic(msg *kafka.Message) (*kafka.Message, error) {
	item := copyMessage(msg)
	headers := []kafka.Header{{Key: "__pipeline_id", Value: []byte(t.pipelineID)}}
	item.Headers = headers

	item.TopicPartition.Topic = &t.outputTopic
	item.Value = msg.Value

	return item, nil
}

// 处理 msg，写入 kafka error topic
func (t *Task) packagingMessageToErrorTopic(msg *kafka.Message) (*kafka.Message, error) {
	// copy msg, 否则会导致 msg 被修改
	item := copyMessage(msg)
	headers := []kafka.Header{{Key: "__pipeline_id", Value: []byte(t.pipelineID)}}
	item.Headers = headers

	item.TopicPartition.Topic = &t.errorTopic
	item.Value = msg.Value

	return item, nil
}

func copyMessage(msg *kafka.Message) *kafka.Message {
	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:       msg.TopicPartition.Topic,
			Partition:   msg.TopicPartition.Partition,
			Offset:      msg.TopicPartition.Offset,
			Metadata:    msg.TopicPartition.Metadata,
			Error:       msg.TopicPartition.Error,
			LeaderEpoch: msg.TopicPartition.LeaderEpoch,
		},
		Value:         msg.Value,
		Key:           msg.Key,
		Timestamp:     msg.Timestamp,
		TimestampType: msg.TimestampType,
		Opaque:        msg.Opaque,
		Headers:       msg.Headers,
	}
}

func (t *Task) ProcessMetricMessageRouting(root *ast.Node) string {
	// 设置__routing字段，值为 __tsid + "-" + @timestamp / 7200
	timestamp, err := root.Get("@timestamp").Int64()
	if err != nil {
		logger.Warnf("parse @timestamp failed: %v", err)
		return ""
	}
	routing := fmt.Sprintf("%d", timestamp/1000/7200)
	return routing
}

func (t *Task) ProcessMetricMessageLabelsStr(root *ast.Node) string {
	// 设置__labels_str字段，值为全部labels按字母序拼接，分隔符为逗号。
	labelsNode := root.Get("labels")
	if labelsNode == nil {
		promNode := root.Get("prometheus")
		if promNode == nil {
			logger.Warnf("property 'labels' does not exist" +
				" and property 'prometheus' does not exist")
			return ""
		}
		if promNode.TypeSafe() != ast.V_OBJECT {
			raw, _ := promNode.Raw()
			logger.Warnf("property 'labels' does not exist"+
				" and property 'prometheus' is not a object: %s", raw)
			return ""
		}

		labelsNode = promNode.Get("labels")
		if labelsNode != nil {
			_, _ = root.Set("labels", *labelsNode)
			_, _ = promNode.Unset("labels")
			labelsNode = root.Get("labels")
		} else {
			logger.Warnf("property 'labels' does not exist" +
				" and property 'prometheus.labels' does not exist")
			return ""
		}
	}

	if labelsNode.TypeSafe() != ast.V_OBJECT {
		raw, _ := labelsNode.Raw()
		logger.Warnf("property 'labels' is not a object: %s", raw)
		return ""
	}

	err := labelsNode.SortKeys(false)
	if err != nil {
		return ""
	}
	itr, err := labelsNode.Properties()
	if err != nil {
		return ""
	}

	labels := []string{}
	var p ast.Pair
	for itr.Next(&p) {
		value, _ := p.Value.String()
		labels = append(labels, fmt.Sprintf("%s=%s", p.Key, value))
	}
	labels_str := strings.Join(labels, ",")
	return labels_str
}

func (t *Task) ProcessMetricMessage(root *ast.Node) (*string, error) {
	labels_str := t.ProcessMetricMessageLabelsStr(root)
	// _, _ = root.Set("__labels_str", ast.NewString(labels_str))

	// __tsid
	md5Hasher := md5.New()
	md5Hasher.Write([]byte(labels_str)) // resMap["__labels_str"].(string)
	hashed := md5Hasher.Sum(nil)
	tsid := hex.EncodeToString(hashed)
	_, _ = root.Set("__tsid", ast.NewString(tsid))

	var routingPtr *string = nil
	routing := t.ProcessMetricMessageRouting(root)
	if routing != "" {
		_, _ = root.Set("__routing", ast.NewString(fmt.Sprintf("%s-%s", tsid, routing)))
		routingPtr = &routing
	}

	return routingPtr, nil
}

func (t *Task) ProcessTraceMessage(root *ast.Node) (*string, error) {
	startTimeNode := root.Get("StartTime")
	endTimeNode := root.Get("EndTime")
	durationNode := root.Get("Duration")
	if startTimeNode != nil {
		startTime, err := startTimeNode.Int64()
		if err != nil {
			return nil, err
		}

		_, _ = root.Set("@timestamp", ast.NewAny(startTime/1000/1000))
	}

	if startTimeNode != nil && endTimeNode != nil && durationNode == nil {
		startTime, err := startTimeNode.Int64()
		if err != nil {
			return nil, err
		}
		endTime, err := startTimeNode.Int64()
		if err != nil {
			return nil, err
		}
		_, _ = root.Set("@Duration", ast.NewAny(endTime-startTime))
	}

	traceIDNode := root.GetByPath("SpanContext", "TraceID")
	if traceIDNode != nil {
		traceID, _ := traceIDNode.String()
		return &traceID, nil
	}

	return nil, nil
}

func (t *Task) PackagingItem(item *osutil.BulkIndexerItem, indexer interfaces.BulkIndexer) error {
	err := indexer.Add(item)
	if err != nil {
		logger.Errorf("failed to add item to indexer: %v", err)
		return err
	}

	return nil
}

// indexer : 写入 opensearch 的请求队列
// docIDToKafkaMsg : 文档 ID 和对应的原始kafka消息
// currentMsgs : 当前批次的原始数据
func (t *Task) FlushMessagesToOS(ctx context.Context, kp *interfaces.KafkaProducer,
	indexer interfaces.BulkIndexer, docIDToKafkaMsg map[string]*kafka.Message) ([]*kafka.Message, error) {

	cnt := 0
	// 重试循环
	for {
		failedDocIDs, err := indexer.Flush(ctx)
		// 如果 Flush 成功，不需要重试
		if err == nil {
			break
		}

		cnt++
		if cnt >= FailureThreshold {
			logger.Errorf("task %s failed to indexer Flush, has retry[%d] times, will write to error topic, error: %v", t, cnt, err)

			failedMsgs := make([]*kafka.Message, 0, len(failedDocIDs))

			if len(failedDocIDs) > 0 {
				for _, docID := range failedDocIDs {
					mqItem, err := t.packagingMessageToErrorTopic(docIDToKafkaMsg[docID])
					if err != nil {
						logger.Errorf("failed to packaging message to error topic, error: %v", err)
						return nil, err
					}

					failedMsgs = append(failedMsgs, mqItem)
				}
			} else {
				// 如果 Flush 重试失败，但是没有失败的数据，则将当前批次的数据写入 error topic
				for _, msg := range docIDToKafkaMsg {
					mqItem, err := t.packagingMessageToErrorTopic(msg)
					if err != nil {
						logger.Errorf("failed to packaging message to error topic, error: %v", err)
						return nil, err
					}

					failedMsgs = append(failedMsgs, mqItem)
				}
			}

			// 不返回 error, 避免一条失败的数据导致整个任务失败
			// 操作下一批前清空请求队列
			indexer.Reset()
			return failedMsgs, nil
		}

		logger.Errorf("failed to indexer Flush, wait to retry[%d]: %v", cnt, err)
		time.Sleep(RetryInterval)
	}

	return nil, nil
}

func (t *Task) FlushMessagesToOutputTopic(msgs []*kafka.Message, c *kafka.Consumer, kp *interfaces.KafkaProducer) error {
	err := t.mqAccess.DoProduceAndCommit(kp, c, msgs)
	if err != nil {
		logger.Errorf("DoProduce to output topic %s failed, %v", t.outputTopic, err)
		return err
	}

	return nil
}

// 如果写入OpenSearch失败，则将数据写入错误topic中
func (t *Task) FlushMessagesToErrorTopic(ErrorMsgs []*kafka.Message, c *kafka.Consumer, kp *interfaces.KafkaProducer) error {
	logger.Debugf("Preparing to produce %d messages to error topic", len(ErrorMsgs))

	err := t.mqAccess.DoProduceAndCommit(kp, c, ErrorMsgs)
	if err != nil {
		logger.Errorf("DoProduce to error topic %s failed, %v", t.errorTopic, err)
		return err
	}

	return nil
}
