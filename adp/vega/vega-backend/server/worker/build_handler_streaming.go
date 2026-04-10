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
	"math"
	"net/http"
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
	"vega-backend/logics/filter_condition"
)

const (
	KafkaTopicPrefix = "vega-dataset"
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
	// 用于保护 runningSubscriptions 的互斥锁
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
		// Resource not found, return nil to  stop the task
		return nil
	}

	// Update task status to running
	_ = sh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusRunning})

	// executeBuild is continuously executed task，so put embedding task to queue first
	embeddingTaskMsg := interfaces.EmbeddingBuildTaskMessage{
		TaskID: taskID,
	}
	payload, err := sonic.Marshal(embeddingTaskMsg)
	if err != nil {
		logger.Errorf("Failed to marshal embedding task message: %v", err)
	} else {
		asynqTask := asynq.NewTask(interfaces.BuildTaskTypeEmbedding, payload)
		_, err = sh.client.Enqueue(asynqTask,
			asynq.Queue(interfaces.DefaultQueue),
			asynq.MaxRetry(interfaces.DATASET_BUILD_MAX_RETRY_COUNT),
			asynq.Timeout(math.MaxInt64),                                                  // 永不超时
			asynq.Deadline(time.Unix(math.MaxInt64/1000000000, math.MaxInt64%1000000000)), // 永不过期
		)
		if err != nil {
			logger.Errorf("Failed to enqueue embedding task: %v", err)
		} else {
			logger.Infof("Embedding task queued for resource: %s", resourceID)
		}
	}

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
func (sh *streamingBuildHandler) subscribeToKafkaMessages(ctx context.Context, resource *interfaces.Resource, uniqueKeys []string) {
	// Create a context that can be canceled
	kafkaCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Use the connector name as the Kafka topic prefix
	topic := fmt.Sprintf("%s-%s.%s", KafkaTopicPrefix, resource.ID, resource.SourceIdentifier)
	groupID := fmt.Sprintf("%s-%s", KafkaTopicPrefix, resource.ID)

	// Create Kafka reader
	reader, err := sh.kafkaAccess.NewReader(kafkaCtx, topic, groupID)
	if err != nil {
		logger.Errorf("Failed to create Kafka reader for topic %s: %v", topic, err)
		// Remove the resource from runningSubscriptions if creation fails
		sh.subscriptionMutex.Lock()
		delete(sh.runningSubscriptions, resource.ID)
		sh.subscriptionMutex.Unlock()
		logger.Infof("Kafka subscription for resource %s has been stopped due to creation failure", resource.ID)
		return
	}
	defer sh.kafkaAccess.CloseReader(reader)

	logger.Infof("Started Kafka subscription for topic %s with group ID %s", topic, groupID)

	// Create a map for quick field lookup and Check if resource has vector fields that need embedding
	outputFieldsMap := make(map[string]bool)
	fieldMap := map[string]*interfaces.Property{}
	hasVectorField := false
	if resource.SchemaDefinition != nil {
		for _, prop := range resource.SchemaDefinition {
			outputFieldsMap[prop.Name] = true
			fieldMap[prop.Name] = prop
			if prop.Type == "vector" {
				hasVectorField = true
			}
		}
	}

	// Create embedding topic if needed
	embeddingTopic := ""
	if hasVectorField {
		embeddingTopic = fmt.Sprintf("%s-%s-embedding", KafkaTopicPrefix, resource.ID)
		// Create Kafka topic if it doesn't exist
		if err := sh.kafkaAccess.CreateTopic(kafkaCtx, embeddingTopic); err != nil {
			logger.Errorf("Failed to create embedding Kafka topic: %v", err)
		}
	}

	retryInterval := interfaces.DATASET_BUILD_RETRY_INTERVAL * time.Second
	// Message processing loop
	for {
		select {
		case <-kafkaCtx.Done():
			logger.Infof("Kafka subscription context canceled, exiting")
			// Remove the resource from runningSubscriptions when subscription ends
			sh.subscriptionMutex.Lock()
			delete(sh.runningSubscriptions, resource.ID)
			sh.subscriptionMutex.Unlock()
			logger.Infof("Kafka subscription for resource %s has been stopped", resource.ID)
			return
		default:
			// Read message from Kafka
			msg, err := sh.kafkaAccess.ReadMessage(kafkaCtx, reader)
			if err != nil {
				logger.Errorf("Failed to read message from Kafka: %v", err)
				// Continue to next message
				continue
			}

			// Parse Kafka message to extract data
			var keyMap map[string]any
			var valueMap map[string]any

			// Check if message value is empty
			if len(msg.Value) == 0 {
				logger.Debugf("Empty message value, skipping processing")
				// Commit the message to avoid reprocessing
				if err := sh.kafkaAccess.CommitMessages(kafkaCtx, reader, msg); err != nil {
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
			} else {
				// Extract data from the message
				if payload, ok := valueMap["payload"].(map[string]any); ok {
					after, hasAfter := payload["after"].(map[string]any)
					before, hasBefore := payload["before"].(map[string]any)

					// Determine operation type
					if hasAfter && !hasBefore {
						// Insert operation
						// Create document from the after data
						document := make(map[string]any)
						for k, v := range after {
							if outputFieldsMap[k] {
								document[k] = v
							}
						}

						// Write document to dataset
						if docIDs, err := sh.ds.CreateDocuments(kafkaCtx, resource.ID, []map[string]any{document}); err != nil {
							logger.Errorf("Failed to write document to dataset: %v", err)
							time.Sleep(retryInterval)
							continue
						} else if hasVectorField && len(docIDs) > 0 {
							// Send document ID to Kafka for embedding
							docID := docIDs[0]
							// Create Kafka writer for embedding
							writer, err := sh.kafkaAccess.NewWriter(kafkaCtx, embeddingTopic)
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
									err = sh.kafkaAccess.WriteMessages(kafkaCtx, writer, []kafka.Message{
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
					} else if hasAfter && hasBefore {
						// Update operation
						// Extract primary key values from keyMap or before data
						primaryKeyValues := make(map[string]any)
						foundAllKeys := false

						// Try from key map first
						if keyMap != nil {
							foundAllKeys = true
							for _, key := range uniqueKeys {
								if value, ok := keyMap[key]; ok {
									primaryKeyValues[key] = value
								} else {
									foundAllKeys = false
									break
								}
							}
						}

						// Try from before data if key map doesn't have all primary keys
						if !foundAllKeys && before != nil {
							foundAllKeys = true
							for _, key := range uniqueKeys {
								if value, ok := before[key]; ok {
									primaryKeyValues[key] = value
								} else {
									foundAllKeys = false
									break
								}
							}
						}

						if !foundAllKeys {
							logger.Errorf("Failed to extract all primary key values for update operation")
							time.Sleep(retryInterval)
							continue
						}

						// Construct filter conditions using unique keys
						var subConds []*interfaces.FilterCondCfg
						for key, value := range primaryKeyValues {
							subConds = append(subConds, &interfaces.FilterCondCfg{
								Name:      key,
								Operation: "eq",
								ValueOptCfg: interfaces.ValueOptCfg{
									Value:     value,
									ValueFrom: interfaces.ValueFrom_Const,
								},
							})
						}

						// Create query params with filter conditions
						params := &interfaces.ResourceDataQueryParams{
							FilterCondCfg: &interfaces.FilterCondCfg{
								Operation: "and",
								SubConds:  subConds,
							},
							Limit: 1, // Only need the first matching document
						}
						actualFilterCond, err := filter_condition.NewFilterCondition(ctx, params.FilterCondCfg, fieldMap)
						if err != nil {
							logger.Errorf("create filter condition failed: %v", err)
							time.Sleep(retryInterval)
							continue
						}
						params.ActualFilterCond = actualFilterCond

						// Query documents by unique keys
						documents, _, err := sh.ds.ListDocuments(kafkaCtx, resource, params)
						if err != nil {
							logger.Errorf("Failed to query documents by unique keys: %v", err)
							time.Sleep(retryInterval)
							continue
						}

						if len(documents) == 0 {
							logger.Errorf("No document found with unique keys")
							time.Sleep(retryInterval)
							continue
						}

						// Extract document ID from the first matching document
						docID, ok := documents[0]["_id"].(string)
						if !ok || docID == "" {
							logger.Errorf("Failed to extract document ID from queried document")
							time.Sleep(retryInterval)
							continue
						}

						// Create updated document from the after data
						document := make(map[string]any)
						// Create a map for quick unique key lookup
						uniqueKeyMap := make(map[string]bool)
						for _, key := range uniqueKeys {
							uniqueKeyMap[key] = true
						}
						for k, v := range after {
							if outputFieldsMap[k] && !uniqueKeyMap[k] {
								document[k] = v
							}
						}

						// Check if any field has changed
						hasChanges := false
						existingDoc := documents[0]
						for k, newVal := range document {
							oldVal, exists := existingDoc[k]
							if !exists || oldVal != newVal {
								hasChanges = true
								break
							}
						}

						// Only update if there are changes
						if hasChanges {
							// Update document in dataset
							if err := sh.ds.UpdateDocument(kafkaCtx, resource.ID, docID, document); err != nil {
								logger.Errorf("Failed to update document in dataset: %v", err)
								time.Sleep(retryInterval)
								continue
							} else if hasVectorField {
								// Send document ID to Kafka for embedding
								// Create Kafka writer for embedding
								writer, err := sh.kafkaAccess.NewWriter(kafkaCtx, embeddingTopic)
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
										err = sh.kafkaAccess.WriteMessages(kafkaCtx, writer, []kafka.Message{
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
						}
					} else if !hasAfter && hasBefore {
						// Delete operation
						// Extract unique key values from keyMap or before data
						uniqueKeyValues := make(map[string]any)
						foundAllKeys := false

						// Try from key map first
						if keyMap != nil {
							foundAllKeys = true
							for _, key := range uniqueKeys {
								if value, ok := keyMap[key]; ok {
									uniqueKeyValues[key] = value
								} else {
									foundAllKeys = false
									break
								}
							}
						}

						// Try from before data if key map doesn't have all unique keys
						if !foundAllKeys && before != nil {
							foundAllKeys = true
							for _, key := range uniqueKeys {
								if value, ok := before[key]; ok {
									uniqueKeyValues[key] = value
								} else {
									foundAllKeys = false
									break
								}
							}
						}

						if !foundAllKeys {
							logger.Errorf("Failed to extract all unique key values for delete operation")
							time.Sleep(retryInterval)
							continue
						}

						// Construct filter conditions using primary keys
						var subConds []*interfaces.FilterCondCfg
						for key, value := range uniqueKeyValues {
							subConds = append(subConds, &interfaces.FilterCondCfg{
								Name:      key,
								Operation: "eq",
								ValueOptCfg: interfaces.ValueOptCfg{
									Value:     value,
									ValueFrom: interfaces.ValueFrom_Const,
								},
							})
						}

						// Create query params with filter conditions
						params := &interfaces.ResourceDataQueryParams{
							FilterCondCfg: &interfaces.FilterCondCfg{
								Operation: "and",
								SubConds:  subConds,
							},
						}

						// Delete documents by query
						if err := sh.ds.DeleteDocumentsByQuery(kafkaCtx, resource, params); err != nil {
							logger.Errorf("Failed to delete documents by query: %v", err)
							time.Sleep(retryInterval)
							continue
						}
					}
				}
			}

			// Commit the message
			if err := sh.kafkaAccess.CommitMessages(kafkaCtx, reader, msg); err != nil {
				logger.Errorf("Failed to commit message: %v", err)
			}
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
		// Catalog not found, return nil to  stop the task
		return nil
	}

	// Extract primary unique keys from SourceMetadata
	uniqueKeys := []string{}
	if uniqueKeysInterface, ok := resource.SourceMetadata["unique_keys"]; ok {
		if keysInterface, ok := uniqueKeysInterface.([]interface{}); ok {
			// Handle []interface{} type
			for _, key := range keysInterface {
				if keyStr, ok := key.(string); ok {
					uniqueKeys = append(uniqueKeys, keyStr)
				}
			}
		}
	}
	if len(uniqueKeys) == 0 {
		logger.Errorf("Resource %s does not have primary keys in SourceMetadata", resource.ID)
		return nil
	}

	kafkaConnectSetting := sh.appSetting.KafkaConnectSetting
	connectorName := fmt.Sprintf("%s-%s-%s", KafkaTopicPrefix, catalog.ID, buildTaskInfo.ID)
	connectorUrl := fmt.Sprintf("%s://%s:%d/connectors", kafkaConnectSetting.Protocol, kafkaConnectSetting.Host, kafkaConnectSetting.Port)

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME: interfaces.CONTENT_TYPE_JSON,
	}

	// Start a goroutine to subscribe to Kafka messages if not already running
	sh.subscriptionMutex.Lock()
	if !sh.runningSubscriptions[resource.ID] {
		sh.runningSubscriptions[resource.ID] = true
		sh.subscriptionMutex.Unlock()
		go sh.subscribeToKafkaMessages(ctx, resource, uniqueKeys)
	} else {
		sh.subscriptionMutex.Unlock()
		logger.Infof("Kafka subscription for resource %s is already running", resource.ID)
	}

	retryInterval := interfaces.DATASET_BUILD_RETRY_INTERVAL * time.Second
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
			//_, _, err := sh.httpClient.Put(ctx, fmt.Sprintf("%s/%s/stop", connectorUrl, connectorName), headers, map[string]interface{}{})
			_, _, err := sh.httpClient.Delete(ctx, fmt.Sprintf("%s/%s", connectorUrl, connectorName), headers)
			if err != nil {
				logger.Errorf("Failed to stop kafka connector: %v", err)
				time.Sleep(retryInterval)
				continue
			}
			logger.Infof("Task %s is stopping, exiting...", buildTaskInfo.ID)
			// Update task status to stopped
			_ = sh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"status": interfaces.BuildTaskStatusStopped})
			return nil
		}

		// check kafka connector status
		respCode, respBody, err := sh.httpClient.Get(ctx, fmt.Sprintf("%s/%s/status", connectorUrl, connectorName), nil, headers)
		if err != nil {
			logger.Errorf("Failed to get kafka connector status: %v", err)
			time.Sleep(retryInterval)
			continue
		}
		if respCode == http.StatusNotFound {
			// Connector not found, create connector
			database := catalog.ConnectorCfg["database"]
			if database == nil || database == "" {
				database = resource.Database
			}
			connectorBody := map[string]any{
				"name": connectorName,
				"config": map[string]any{
					"connector.class":       interfaces.ConnectorClassMapping[catalog.ConnectorType],
					"tasks.max":             "1",
					"database.hostname":     catalog.ConnectorCfg["host"],
					"database.port":         catalog.ConnectorCfg["port"],
					"database.user":         catalog.ConnectorCfg["username"],
					"database.password":     catalog.ConnectorCfg["password"],
					"database.server.id":    fmt.Sprintf("%d", getServerID(connectorName)),
					"database.server.name":  getServerName(fmt.Sprintf("%v", catalog.ConnectorCfg["host"])),
					"database.include.list": database,
					"schema.history.internal.kafka.bootstrap.servers": fmt.Sprintf("%s:%d", sh.appSetting.MQSetting.MQHost, sh.appSetting.MQSetting.MQPort),
					"schema.history.internal.kafka.topic":             fmt.Sprintf("%s-schema-changes", KafkaTopicPrefix),
					"include.schema.changes":                          "true",
					"topic.prefix":                                    fmt.Sprintf("%s-%s", KafkaTopicPrefix, resource.ID),
				},
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
		} else if respCode == http.StatusOK {
			// Connector found, check status
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
					}
				}
			} else {
				logger.Errorf("Invalid respBody type: %T", respBody)
				time.Sleep(retryInterval)
				continue
			}
		}
	}
}
