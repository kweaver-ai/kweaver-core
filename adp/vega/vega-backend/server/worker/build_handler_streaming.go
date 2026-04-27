// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package worker provides background workers for VEGA Manager.
package worker

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/segmentio/kafka-go"

	"vega-backend/common"
	asynqAccess "vega-backend/drivenadapters/asynq"
	taskAccess "vega-backend/drivenadapters/build_task"
	kafkaAccess "vega-backend/drivenadapters/kafka"
	resourceAccess "vega-backend/drivenadapters/resource"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	"vega-backend/logics/dataset"
)

// getServerID generates a unique server ID based on the connector name
func getServerID(connectorName string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(connectorName))
	return h.Sum32()
}

// getServerName generates a server name based on the hostname hash
func getServerName(hostname string) string {
	h := fnv.New32a()
	h.Write([]byte(hostname))
	return fmt.Sprintf("vega-%d", h.Sum32())
}

// streamingBuildHandler handles build tasks.
type streamingBuildHandler struct {
	appSetting  *common.AppSetting
	taskAccess  interfaces.BuildTaskAccess
	resAccess   interfaces.ResourceAccess
	cs          interfaces.CatalogService
	ds          interfaces.DatasetService
	client      *asynq.Client
	httpClient  rest.HTTPClient
	kafkaAccess interfaces.KafkaAccess
	// 记录已启动的订阅，避免重复启动
	runningSubscriptions map[string]bool // key: resource ID
	// 记录每个resource对应的cancel函数，用于取消订阅
	subscriptionCancels map[string]context.CancelFunc // key: resource ID
	// 用于保护 runningSubscriptions 和 subscriptionCancels 的互斥锁
	subscriptionMutex sync.Mutex
}

// NewStreamingBuildHandler creates a new build handler.
func NewStreamingBuildHandler(appSetting *common.AppSetting) *streamingBuildHandler {
	return &streamingBuildHandler{
		appSetting:           appSetting,
		taskAccess:           taskAccess.NewBuildTaskAccess(appSetting),
		resAccess:            resourceAccess.NewResourceAccess(appSetting),
		cs:                   catalog.NewCatalogService(appSetting),
		ds:                   dataset.NewDatasetService(appSetting),
		client:               asynqAccess.NewAsynqAccess(appSetting).CreateClient(context.Background()),
		httpClient:           common.NewHTTPClient(),
		kafkaAccess:          kafkaAccess.NewKafkaAccess(appSetting),
		runningSubscriptions: make(map[string]bool),
		subscriptionCancels:  make(map[string]context.CancelFunc),
	}
}

// HandleTask handles a build task from the queue.
func (sh *streamingBuildHandler) HandleTask(ctx context.Context, task *asynq.Task) error {
	var msg interfaces.StreamingBuildTaskMessage
	if err := sonic.Unmarshal(task.Payload(), &msg); err != nil {
		logger.Errorf("Failed to unmarshal task message: %v", err)
		return err
	}

	taskID := msg.TaskID
	buildTaskInfo, err := sh.taskAccess.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get build task failed: %w", err)
	}
	if buildTaskInfo == nil {
		// Task not found, return nil
		return nil
	}
	resourceID := buildTaskInfo.ResourceID
	logger.Infof("Starting build for task: %s, resource: %s", taskID, resourceID)

	// Get resource info
	resource, err := sh.resAccess.GetByID(ctx, resourceID)
	if err != nil {
		logger.Errorf("Failed to get resource for task %s: %v", taskID, err)
		return err
	}
	if resource == nil {
		logger.Errorf("Resource not found for task %s, resourceID: %s", taskID, resourceID)
		err = sh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusFailed, "errorMsg": "resource not found"})
		if err != nil {
			return fmt.Errorf("update build task status failed: %w", err)
		}
		// Resource not found, return nil to  stop the task
		return nil
	}

	if buildTaskInfo.Status == interfaces.BuildTaskStatusInit {
		err := createLocalIndex(ctx, sh.ds, buildTaskInfo, resource)
		if err != nil {
			return fmt.Errorf("create local index failed: %w", err)
		}
		if buildTaskInfo.EmbeddingFields != "" {
			// first time, send embedding task to queue first
			err = sendEmbeddingTask(sh.client, taskID)
			if err != nil {
				return fmt.Errorf("send embedding task failed: %w", err)
			}
			logger.Infof("Embedding task sent for task %s", taskID)
		}
	}

	// Update task status to running
	_ = sh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusRunning})

	// Execute build
	err = sh.executeBuild(ctx, resource, buildTaskInfo)
	if err != nil {
		// Update task status to failed
		_ = sh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"errorMsg": err.Error()})
		return err
	}

	logger.Infof("Build stopped for task: %s, resource: %s", taskID, resourceID)
	return nil
}

// subscribeToKafkaMessages subscribes to Kafka messages for the specified resource
func (sh *streamingBuildHandler) subscribeToKafkaMessages(ctx context.Context, resource *interfaces.Resource, buildTaskInfo *interfaces.BuildTask, sourceId string) {
	resourceID := resource.ID
	// Use the connector name as the Kafka topic prefix
	topic := fmt.Sprintf("%s-%s.%s", interfaces.BUILD_PREFIX, resourceID, sourceId)
	groupID := fmt.Sprintf("%s-%s", interfaces.BUILD_PREFIX, resourceID)

	// Create Kafka reader
	reader, err := sh.kafkaAccess.NewReader(ctx, topic, groupID)
	if err != nil {
		logger.Errorf("Failed to create Kafka reader for topic %s: %v", topic, err)
		// Remove the resource from runningSubscriptions if creation fails
		sh.subscriptionMutex.Lock()
		delete(sh.runningSubscriptions, resourceID)
		sh.subscriptionMutex.Unlock()
		logger.Infof("Kafka subscription for resource %s has been stopped due to creation failure", resourceID)
		return
	}
	defer sh.kafkaAccess.CloseReader(reader)

	logger.Infof("Started Kafka subscription for topic %s with group ID %s", topic, groupID)

	fieldMap := map[string]*interfaces.Property{}
	for _, prop := range resource.SchemaDefinition {
		fieldMap[prop.Name] = prop
	}

	// Create embedding topic if needed
	embeddingTopic := ""
	if buildTaskInfo.EmbeddingFields != "" {
		embeddingTopic = fmt.Sprintf("%s-%s-embedding", interfaces.BUILD_PREFIX, resourceID)
		// Create Kafka topic if it doesn't exist
		if err := sh.kafkaAccess.CreateTopic(ctx, embeddingTopic); err != nil {
			logger.Errorf("Failed to create embedding Kafka topic: %v", err)
		}
	}

	indexName := getIndexName(resourceID, buildTaskInfo.ID)
	retryInterval := interfaces.BUILD_TASK_RETRY_INTERVAL * time.Second
	// Message processing loop
	for {
		select {
		case <-ctx.Done():
			logger.Infof("Kafka subscription context canceled, exiting")

			// Remove the resource from runningSubscriptions when subscription ends
			sh.subscriptionMutex.Lock()
			delete(sh.runningSubscriptions, resourceID)
			sh.subscriptionMutex.Unlock()
			logger.Infof("Kafka subscription for resource %s has been stopped", resourceID)
			return
		default:
			// Read message from Kafka
			msg, err := sh.kafkaAccess.ReadMessage(ctx, reader)
			if err != nil {
				logger.Errorf("Streaming task Failed to read message from Kafka: %v", err)
				time.Sleep(retryInterval)
				continue
			}
			// 打印消息的基本信息和内容
			//logger.Debugf("Received message: key=%s, value=%s", string(msg.Key), string(msg.Value))

			// Parse Kafka message to extract data
			var keyMap map[string]any
			var valueMap map[string]any

			// Check if message value is empty
			if len(msg.Value) == 0 {
				logger.Debugf("Empty message value, skipping processing")
				// Commit the message to avoid reprocessing
				if err := sh.kafkaAccess.CommitMessages(ctx, reader, msg); err != nil {
					logger.Errorf("Failed to commit message: %v", err)
				}
				continue
			}

			if err := sonic.Unmarshal(msg.Key, &keyMap); err != nil {
				logger.Errorf("Failed to unmarshal message key: %v", err)
				time.Sleep(retryInterval)
				continue
			} else if err := sonic.Unmarshal(msg.Value, &valueMap); err != nil {
				logger.Errorf("Failed to unmarshal message value: %v", err)
				time.Sleep(retryInterval)
				continue
			}
			// Extract data from the message
			if payload, ok := valueMap["payload"].(map[string]any); ok {
				op := payload["op"].(string)
				after, _ := payload["after"].(map[string]any)

				// Determine operation type
				switch op {
				case "r", "c":
					// Full snapshot or create operation
					// Create document from the after data
					document := make(map[string]any)
					for k, v := range after {
						document[k] = v
					}

					if docIDs, err := sh.ds.UpsertDocuments(ctx, indexName, []map[string]any{{"id": getOldDocID(getPrimaryKeyValue(keyMap)), "document": document}}); err != nil {
						logger.Errorf("Failed to write document to dataset: %v", err)
						time.Sleep(retryInterval)
						continue
					} else if buildTaskInfo.EmbeddingFields != "" && len(docIDs) > 0 {
						// Send document ID to Kafka for embedding
						docID := docIDs[0]
						// Create Kafka writer for embedding
						writer, err := sh.kafkaAccess.NewWriter(ctx, embeddingTopic)
						if err != nil {
							logger.Errorf("Failed to create Kafka writer for embedding: %v", err)
						} else {
							// Create message
							messageData := map[string]any{
								"document_id": docID,
							}
							messageBytes, err := sonic.Marshal(messageData)
							if err != nil {
								logger.Errorf("Failed to marshal message: %v", err)
							} else {
								// Write message to Kafka
								// Use docID + timestamp as key to avoid conflicts even if document is modified multiple times
								err = sh.kafkaAccess.WriteMessages(ctx, writer, []kafka.Message{
									{
										Key:   []byte(fmt.Sprintf("%s-%d", docID, time.Now().UnixNano())),
										Value: messageBytes,
									},
								}...)
								if err != nil {
									logger.Errorf("Failed to write message to Kafka: %v", err)
								}
							}
							// Close writer
							sh.kafkaAccess.CloseWriter(writer)
						}
					}
				case "u":
					// Update operation
					if err := sh.handleUpdateOperation(ctx, keyMap, after, fieldMap, indexName, embeddingTopic, retryInterval, buildTaskInfo, resource); err != nil {
						logger.Errorf("Failed to handle update operation: %v", err)
						time.Sleep(retryInterval)
						continue
					}
				case "d":
					// Delete operation
					if err := sh.handleDeleteOperation(ctx, keyMap, indexName, resource, retryInterval); err != nil {
						logger.Errorf("Failed to handle delete operation: %v", err)
						time.Sleep(retryInterval)
						continue
					}
				default:
					logger.Errorf("Unknown operation type: %s", op)
					time.Sleep(retryInterval)
					continue
				}
			}

			// Commit the message
			if err := sh.kafkaAccess.CommitMessages(ctx, reader, msg); err != nil {
				logger.Errorf("Failed to commit message: %v", err)
			}
			buildTaskInfo.SyncedCount++
		}
	}
}

// executeBuild executes the build logic
func (sh *streamingBuildHandler) executeBuild(ctx context.Context, resource *interfaces.Resource, buildTaskInfo *interfaces.BuildTask) error {
	// Get catalog for MySQL connection
	catalog, err := sh.cs.GetByID(ctx, resource.CatalogID, true)
	if err != nil {
		return fmt.Errorf("get catalog failed: %w", err)
	}
	if catalog == nil {
		logger.Errorf("Catalog not found for task %s, catalogID: %s", buildTaskInfo.ID, resource.CatalogID)
		err = sh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"status": interfaces.BuildTaskStatusFailed, "errorMsg": "catalog not found"})
		if err != nil {
			return fmt.Errorf("update build task status failed: %w", err)
		}
		// Catalog not found, return nil to  stop the task
		return nil
	}

	kafkaConnectSetting := sh.appSetting.KafkaConnectSetting
	// connector name 和 catalog 绑定，catalog 下多个 resource 公有一个 connector，各自订阅自己的表的 topic
	connectorName := fmt.Sprintf("%s-%s", interfaces.BUILD_PREFIX, catalog.ID)
	connectorUrl := fmt.Sprintf("%s://%s:%d/connectors", kafkaConnectSetting.Protocol, kafkaConnectSetting.Host, kafkaConnectSetting.Port)

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME: interfaces.CONTENT_TYPE_JSON,
	}

	retryInterval := interfaces.BUILD_TASK_RETRY_INTERVAL * time.Second
	for {
		// Check task status before each batch
		taskStatus, err := sh.taskAccess.GetStatus(ctx, buildTaskInfo.ID)
		if err != nil {
			logger.Errorf("Failed to get task status: %v", err)
			time.Sleep(retryInterval)
			continue
		}

		// Handle stopping status
		if taskStatus == interfaces.BuildTaskStatusStopping {
			// Task is stopping, notify the kafka connector to stop
			// only stop kafka subscription now, stop connector need get connector config first
			// TODO
			// _, _, err := sh.httpClient.Put(ctx, fmt.Sprintf("%s/%s/stop", connectorUrl, connectorName), headers, map[string]interface{}{})
			// // _, _, err := sh.httpClient.Delete(ctx, fmt.Sprintf("%s/%s", connectorUrl, connectorName), headers)
			// if err != nil {
			// 	logger.Errorf("Failed to stop kafka connector: %v", err)
			// 	time.Sleep(retryInterval)
			// 	continue
			// }
			logger.Infof("Task %s is stopping, exiting...", buildTaskInfo.ID)
			err = sh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"status": interfaces.BuildTaskStatusStopped, "syncedCount": buildTaskInfo.SyncedCount})
			if err != nil {
				return fmt.Errorf("update build task status failed: %w", err)
			}

			// Cancel Kafka subscription
			sh.subscriptionMutex.Lock()
			if cancelFn, exists := sh.subscriptionCancels[resource.ID]; exists {
				cancelFn()
				delete(sh.subscriptionCancels, resource.ID)
				logger.Infof("Canceled Kafka subscription for resource %s", resource.ID)
			}
			sh.subscriptionMutex.Unlock()

			return nil
		}

		database := catalog.ConnectorCfg["database"]
		if database == nil || database == "" {
			database = resource.Database
		}
		sourceId, err := sh.formatTableName(resource.SourceIdentifier, catalog.ConnectorType, database)
		if err != nil {
			logger.Errorf("Failed to format table name: %v", err)
			err = sh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"status": interfaces.BuildTaskStatusFailed, "errorMsg": err.Error()})
			if err != nil {
				return fmt.Errorf("update build task status failed: %w", err)
			}
			return nil
		}
		// get connector
		respCode, respBody, err := sh.httpClient.Get(ctx, fmt.Sprintf("%s/%s", connectorUrl, connectorName), nil, headers)
		if err != nil {
			logger.Errorf("Failed to get kafka connector: %v", err)
			time.Sleep(retryInterval)
			continue
		}
		if respCode == http.StatusNotFound {
			connectorBody, sourceId, err := sh.buildConnectorConfig(connectorName, catalog, resource, database, sourceId)
			if err != nil {
				logger.Errorf("Failed to build connector config: %v", err)
				err = sh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"status": interfaces.BuildTaskStatusFailed, "errorMsg": "resource not found"})
				if err != nil {
					return fmt.Errorf("update build task status failed: %w", err)
				}
				return nil
			}
			respCode, respBody, err := sh.httpClient.Post(ctx, connectorUrl, headers, connectorBody)
			if err != nil {
				logger.Errorf("Failed to create kafka connector: %v", err)
				time.Sleep(retryInterval)
				continue
			}
			if respCode != http.StatusCreated {
				logger.Errorf("Create kafka connector %s failed, status code: %d, body: %v", connectorName, respCode, respBody)
				time.Sleep(retryInterval)
				continue
			}

			logger.Infof("Create kafka connector %s success", connectorName)
			time.Sleep(retryInterval)

			// Start a goroutine to subscribe to Kafka messages if not already running
			sh.subscriptionMutex.Lock()
			if !sh.runningSubscriptions[resource.ID] {
				sh.runningSubscriptions[resource.ID] = true
				subCtx, cancel := context.WithCancel(context.Background())
				sh.subscriptionCancels[resource.ID] = cancel
				sh.subscriptionMutex.Unlock()
				go sh.subscribeToKafkaMessages(subCtx, resource, buildTaskInfo, sourceId)
			} else {
				sh.subscriptionMutex.Unlock()
			}
		} else if respCode == http.StatusOK {
			// Connector found
			config := respBody.(map[string]any)["config"].(map[string]any)
			tableIncludeList, ok := config["table.include.list"].(string)
			if !ok {
				logger.Errorf("Invalid table.include.list type: %T", config["table.include.list"])
				time.Sleep(retryInterval)
				continue
			}
			table_lists := strings.Split(tableIncludeList, ",")
			tableExist := false
			for _, table := range table_lists {
				if strings.TrimSpace(table) == sourceId {
					tableExist = true
					break
				}
			}
			if !tableExist {
				// update kafka connector config
				newTableList := tableIncludeList
				if newTableList != "" {
					newTableList += ","
				}
				newTableList += sourceId
				config["table.include.list"] = newTableList
				_, _, err = sh.httpClient.Put(ctx, fmt.Sprintf("%s/%s/config", connectorUrl, connectorName), headers, config)
				if err != nil {
					logger.Errorf("Failed to update kafka connector config: %v", err)
					time.Sleep(retryInterval)
					continue
				}
				logger.Infof("Updated kafka connector config to include table: %s", sourceId)
			}

			// check kafka connector status
			_, respBody, err = sh.httpClient.Get(ctx, fmt.Sprintf("%s/%s/status", connectorUrl, connectorName), nil, headers)
			if err != nil {
				logger.Errorf("Failed to get kafka connector status: %v", err)
				time.Sleep(retryInterval)
				continue
			}
			// Type assertion for respBody
			if statusBody, ok := respBody.(map[string]any); ok {
				// Type assertion for connector field
				if connector, ok := statusBody["connector"].(map[string]any); ok {
					if state, ok := connector["state"].(string); ok && state != "RUNNING" {
						_, _, err = sh.httpClient.Put(ctx, fmt.Sprintf("%s/%s/resume", connectorUrl, connectorName), headers, map[string]interface{}{})
						if err != nil {
							logger.Errorf("Failed to resume kafka connector: %v", err)
							time.Sleep(retryInterval)
							continue
						}
					} else {
						if tasks, ok := statusBody["tasks"].([]interface{}); ok {
							for _, task := range tasks {
								if taskMap, ok := task.(map[string]any); ok {
									if taskState, ok := taskMap["state"].(string); ok && taskState != "RUNNING" {
										if idFloat, ok := taskMap["id"].(float64); ok {
											_, _, err = sh.httpClient.Post(ctx, fmt.Sprintf("%s/%s/tasks/%d/restart", connectorUrl, connectorName, int(idFloat)), headers, map[string]interface{}{})
											if err != nil {
												logger.Errorf("Failed to restart task: %v", err)
											}
										}
									}
								}
							}
						}
					}
					time.Sleep(retryInterval * 2) // for task status check
				}
			} else {
				logger.Errorf("Invalid respBody type: %T", respBody)
				time.Sleep(retryInterval)
				continue
			}
		}
	}
}

// buildConnectorConfig builds the connector configuration
func (sh *streamingBuildHandler) buildConnectorConfig(connectorName string, catalog *interfaces.Catalog, resource *interfaces.Resource, database any, sourceId string) (map[string]any, string, error) {
	// Connector not found, create connector
	mqSetting := sh.appSetting.MQSetting
	connectorBody := map[string]any{
		"name": connectorName,
		"config": map[string]any{
			"connector.class":   interfaces.ConnectorClassMapping[catalog.ConnectorType],
			"tasks.max":         "1",
			"database.hostname": catalog.ConnectorCfg["host"],
			"database.port":     catalog.ConnectorCfg["port"],
			"database.user":     catalog.ConnectorCfg["username"],
			"database.password": catalog.ConnectorCfg["password"],
			//"column.include.list":   ?,
			"schema.history.internal.kafka.bootstrap.servers": fmt.Sprintf("%s:%d", mqSetting.MQHost, mqSetting.MQPort),
			"schema.history.internal.kafka.topic":             fmt.Sprintf("%s-schema-changes", interfaces.BUILD_PREFIX),
			"include.schema.changes":                          "true",
			"topic.prefix":                                    fmt.Sprintf("%s-%s", interfaces.BUILD_PREFIX, resource.ID),
		},
	}

	if mqSetting.Auth.Mechanism != "" && mqSetting.Auth.Username != "" && mqSetting.Auth.Password != "" {
		jaasConfig := fmt.Sprintf("org.apache.kafka.common.security.plain.PlainLoginModule required username=\"%s\" password=\"%s\";", mqSetting.Auth.Username, mqSetting.Auth.Password)
		connectorBody["config"].(map[string]any)["schema.history.internal.consumer.security.protocol"] = "SASL_PLAINTEXT"
		connectorBody["config"].(map[string]any)["schema.history.internal.consumer.sasl.mechanism"] = mqSetting.Auth.Mechanism
		connectorBody["config"].(map[string]any)["schema.history.internal.consumer.sasl.jaas.config"] = jaasConfig
		connectorBody["config"].(map[string]any)["schema.history.internal.producer.security.protocol"] = "SASL_PLAINTEXT"
		connectorBody["config"].(map[string]any)["schema.history.internal.producer.sasl.mechanism"] = mqSetting.Auth.Mechanism
		connectorBody["config"].(map[string]any)["schema.history.internal.producer.sasl.jaas.config"] = jaasConfig
		connectorBody["config"].(map[string]any)["table.include.list"] = sourceId
	}
	switch catalog.ConnectorType {
	case "mysql":
		connectorBody["config"].(map[string]any)["database.server.id"] = fmt.Sprintf("%d", getServerID(connectorName))
		connectorBody["config"].(map[string]any)["database.server.name"] = getServerName(fmt.Sprintf("%v", catalog.ConnectorCfg["host"]))
		connectorBody["config"].(map[string]any)["database.include.list"] = database
	case "postgresql":
		connectorBody["config"].(map[string]any)["database.dbname"] = database
		//connectorBody["config"].(map[string]any)["schema.include.list"] = "public" //一般用不上，table.include.list包含schema信息
		connectorBody["config"].(map[string]any)["plugin.name"] = "pgoutput"
	default:
		logger.Errorf("Unsupported connector type: %s", catalog.ConnectorType)
		return nil, "", fmt.Errorf("Unsupported connector type: %s", catalog.ConnectorType)
	}

	return connectorBody, sourceId, nil
}

// handleUpdateOperation 处理更新操作
func (sh *streamingBuildHandler) handleUpdateOperation(ctx context.Context, keyMap, after map[string]any, fieldMap map[string]*interfaces.Property, indexName, embeddingTopic string, retryInterval time.Duration, buildTaskInfo *interfaces.BuildTask, resource *interfaces.Resource) error {
	primaryKeyValues := getPrimaryKeyValue(keyMap)
	if primaryKeyValues == nil {
		return fmt.Errorf("failed to extract unique key values from keyMap")
	}
	oldDocID := getOldDocID(primaryKeyValues)

	// Create updated document from the after data
	document := make(map[string]any)
	for k, v := range after {
		document[k] = v
	}

	newDocID := getNewDocID(primaryKeyValues, document)
	if newDocID != oldDocID {
		err := sh.ds.DeleteDocument(ctx, indexName, oldDocID)
		if err != nil {
			return fmt.Errorf("failed to delete document in dataset: %w", err)
		}
	}

	_, err := sh.ds.UpsertDocuments(ctx, indexName, []map[string]any{{"id": newDocID, "document": document}})
	if err != nil {
		return fmt.Errorf("failed to update document in dataset: %w", err)
	} else if buildTaskInfo.EmbeddingFields != "" {
		// Send document ID to Kafka for embedding
		// Create Kafka writer for embedding
		writer, err := sh.kafkaAccess.NewWriter(ctx, embeddingTopic)
		if err != nil {
			return fmt.Errorf("failed to create Kafka writer for embedding: %w", err)
		}
		defer sh.kafkaAccess.CloseWriter(writer)

		// Create message
		messageData := map[string]any{
			"document_id": newDocID,
		}
		messageBytes, err := sonic.Marshal(messageData)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		// Write message to Kafka
		// Use docID + timestamp as key to avoid conflicts even if document is modified multiple times
		err = sh.kafkaAccess.WriteMessages(ctx, writer, []kafka.Message{
			{
				Key:   []byte(fmt.Sprintf("%s-%d", newDocID, time.Now().UnixNano())),
				Value: messageBytes,
			},
		}...)
		if err != nil {
			return fmt.Errorf("failed to write message to Kafka: %w", err)
		}
	}

	return nil
}

// handleDeleteOperation 处理删除操作
func (sh *streamingBuildHandler) handleDeleteOperation(ctx context.Context, keyMap map[string]any, indexName string, resource *interfaces.Resource, retryInterval time.Duration) error {
	primaryKeyValues := getPrimaryKeyValue(keyMap)
	if primaryKeyValues == nil {
		return fmt.Errorf("failed to extract unique key values from keyMap")
	}
	oldDocID := getOldDocID(primaryKeyValues)

	// Delete documents by query
	if err := sh.ds.DeleteDocument(ctx, indexName, oldDocID); err != nil {
		return fmt.Errorf("failed to delete document in dataset: %w", err)
	}

	return nil
}

// 格式化table名称
func (sh *streamingBuildHandler) formatTableName(sourceIdentifier string, connectorType string, database any) (string, error) {
	sourceId := sourceIdentifier
	switch connectorType {
	case "mysql":
		// 如果不是 db.table 格式，前面加上 dbname.
		if !strings.Contains(sourceIdentifier, ".") {
			sourceId = fmt.Sprintf("%v", database) + "." + sourceIdentifier
		}
	case "postgresql":
		// 如果是 db.schema.table 格式，去掉 db.
		if strings.Count(sourceIdentifier, ".") >= 2 {
			parts := strings.Split(sourceIdentifier, ".")
			sourceId = strings.Join(parts[1:], ".")
		} else if !strings.Contains(sourceIdentifier, ".") {
			return "", fmt.Errorf("sourceIdentifier %s is not contain database name or schema name", sourceIdentifier)
		}
	default:
		return "", fmt.Errorf("connector type %s is not supported", connectorType)
	}
	return sourceId, nil
}

// getPrimaryKeyValue 获取主键值
func getPrimaryKeyValue(keyMap map[string]any) []interfaces.KeyValue {
	keySize := len(keyMap)
	primaryKeyValues := make([]interfaces.KeyValue, 0, keySize)
	// 检查keyMap是否包含payload字段
	keyData := keyMap
	if payload, ok := keyMap["payload"].(map[string]any); ok {
		keyData = payload
	}

	keys := make([]string, 0, keySize)
	for key := range keyData {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if value, ok := keyData[key]; ok {
			primaryKeyValues = append(primaryKeyValues, interfaces.KeyValue{
				Key:   key,
				Value: value,
			})
		}
	}
	return primaryKeyValues
}
