// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package worker provides background workers for VEGA Manager.
package worker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/common"
	taskAccess "vega-backend/drivenadapters/build_task"
	kafkaAccess "vega-backend/drivenadapters/kafka"
	"vega-backend/drivenadapters/model_factory"
	resourceAccess "vega-backend/drivenadapters/resource"
	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
	opensearchConnector "vega-backend/logics/connectors/local/index/opensearch"
	"vega-backend/logics/dataset"
)

// embeddingHandler handles embedding tasks.
type embeddingHandler struct {
	appSetting  *common.AppSetting
	taskAccess  interfaces.BuildTaskAccess
	resAccess   interfaces.ResourceAccess
	ds          interfaces.DatasetService
	connector   connectors.IndexConnector
	kafkaAccess interfaces.KafkaAccess
	mfa         interfaces.ModelFactoryAccess
}

// NewEmbeddingBuildHandler creates a new embedding handler.
func NewEmbeddingBuildHandler(appSetting *common.AppSetting) *embeddingHandler {
	opensearchSetting, ok := appSetting.DepServices["opensearch"]
	if !ok {
		panic("opensearch service not found in depServices")
	}
	cfg := interfaces.ConnectorConfig{
		"host":          opensearchSetting["host"],
		"port":          opensearchSetting["port"],
		"username":      opensearchSetting["user"],
		"password":      opensearchSetting["password"],
		"index_pattern": opensearchSetting["index_pattern"],
	}
	connector, err := opensearchConnector.NewOpenSearchConnector().New(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create OpenSearch connector: %v", err))
	}
	return &embeddingHandler{
		appSetting:  appSetting,
		taskAccess:  taskAccess.NewBuildTaskAccess(appSetting),
		resAccess:   resourceAccess.NewResourceAccess(appSetting),
		ds:          dataset.NewDatasetService(appSetting),
		connector:   connector.(connectors.IndexConnector),
		kafkaAccess: kafkaAccess.NewKafkaAccess(appSetting),
		mfa:         model_factory.NewModelFactoryAccess(appSetting),
	}
}

// HandleTask handles an embedding task from the queue.
func (eh *embeddingHandler) HandleTask(ctx context.Context, task *asynq.Task) error {
	var msg interfaces.EmbeddingBuildTaskMessage
	if err := sonic.Unmarshal(task.Payload(), &msg); err != nil {
		logger.Errorf("Failed to unmarshal task message: %v", err)
		return err
	}

	taskID := msg.TaskID
	buildTaskInfo, err := eh.taskAccess.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get build task failed: %w", err)
	}
	if buildTaskInfo == nil {
		// Task not found, return nil
		return nil
	}
	logger.Infof("Starting embedding for task: %s, resource: %s", taskID, buildTaskInfo.ResourceID)

	// Get resource info
	resource, err := eh.resAccess.GetByID(ctx, buildTaskInfo.ResourceID)
	if err != nil {
		logger.Errorf("Failed to get resource for task %s: %v", taskID, err)
		return err
	}
	if resource == nil {
		logger.Errorf("Resource not found for task %s, resourceID: %s", taskID, buildTaskInfo.ResourceID)
		// Resource not found, return nil to  stop the task
		return nil
	}

	// Update task status to running
	err = eh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusRunning})
	if err != nil {
		return fmt.Errorf("update build task status failed: %w", err)
	}

	// Execute embedding
	embed_err := eh.executeEmbedding(ctx, resource, buildTaskInfo)
	logger.Infof("executeEmbedding completed")
	if embed_err != nil {
		// Update task status to failed
		err = eh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"errorMsg": embed_err.Error()})
		if err != nil {
			return fmt.Errorf("update build task status failed: %w", err)
		}
		return embed_err
	}

	// Update task status to completed
	err = eh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusCompleted})
	if err != nil {
		return fmt.Errorf("update build task status failed: %w", err)
	}

	logger.Infof("Embedding completed for task: %s, resource: %s", taskID, buildTaskInfo.ResourceID)
	return nil
}

// executeEmbedding executes the embedding logic
func (eh *embeddingHandler) executeEmbedding(ctx context.Context, resource *interfaces.Resource, buildTaskInfo *interfaces.BuildTask) error {
	// get vector fields from resource.schema_definition
	fieldsMap := make(map[string]*interfaces.Property)
	embeddingFields := strings.Split(buildTaskInfo.EmbeddingFields, ",")
	for _, prop := range resource.SchemaDefinition {
		fieldsMap[prop.Name] = prop
	}

	// Use the connector name as the Kafka topic prefix
	topic := fmt.Sprintf("%s-%s-embedding", interfaces.BUILD_PREFIX, resource.ID)
	groupID := fmt.Sprintf("%s-embedding-%s", interfaces.BUILD_PREFIX, resource.ID)

	// Create Kafka reader
	reader, err := eh.kafkaAccess.NewReader(ctx, topic, groupID)
	if err != nil {
		return fmt.Errorf("failed to create Kafka reader for topic %s: %w", topic, err)
	}
	defer eh.kafkaAccess.CloseReader(reader)

	logger.Infof("Started Kafka subscription for embedding topic %s with group ID %s", topic, groupID)
	indexName := getIndexName(resource.ID, buildTaskInfo.ID)

	// Message processing loop
	retryInterval := interfaces.BUILD_TASK_RETRY_INTERVAL * time.Second
	totalProcessed := buildTaskInfo.VectorizedCount
	const batchUpdateThreshold = 1000 // 每处理1000条消息更新一次
	lastUpdateTime := time.Now()
	updateInterval := 30 * time.Second // embedding速度慢，至少每30秒更新一次
	for {
		// Check task status before each iteration
		taskStatus, err := eh.taskAccess.GetStatus(ctx, buildTaskInfo.ID)
		if err != nil {
			logger.Errorf("Failed to get task status: %v", err)
			time.Sleep(retryInterval)
			continue
		}

		// Handle stopping status
		if taskStatus == interfaces.BuildTaskStatusStopping {
			// Task is stopping, exit the loop
			logger.Infof("Task %s is stopping, exiting...", buildTaskInfo.ID)
			// Update task status to stopped
			err := eh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"status": interfaces.BuildTaskStatusStopped, "vectorizedCount": totalProcessed})
			if err != nil {
				return fmt.Errorf("update build task status failed: %w", err)
			}
			return nil
		}

		select {
		case <-ctx.Done():
			logger.Infof("Kafka subscription context canceled, exiting")
			// 最后一次更新任务状态
			_ = eh.taskAccess.UpdateStatus(context.Background(), buildTaskInfo.ID, map[string]interface{}{"vectorizedCount": totalProcessed})
			return nil
		default:
			// 创建带超时的上下文，避免ReadMessage一直阻塞
			timeoutCtx, cancel := context.WithTimeout(context.Background(), updateInterval)
			defer cancel()

			// Read message from Kafka
			msg, err := eh.kafkaAccess.ReadMessage(timeoutCtx, reader)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					// 超时，检查是否需要更新任务状态
					if totalProcessed > buildTaskInfo.VectorizedCount && time.Since(lastUpdateTime) > updateInterval {
						_ = eh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"vectorizedCount": totalProcessed})
						lastUpdateTime = time.Now()
					}
				} else {
					logger.Errorf("Embedding task Failed to read message from Kafka: %v", err)
				}
				time.Sleep(retryInterval)
				continue
			}

			// Parse Kafka message to extract document ID
			var messageData map[string]any
			if err := sonic.Unmarshal(msg.Value, &messageData); err != nil {
				logger.Errorf("Failed to unmarshal message value: %v", err)
				time.Sleep(retryInterval)
				continue
			}

			// Get document ID from message
			docID, ok := messageData["document_id"].(string)
			if !ok || docID == "" {
				logger.Errorf("Invalid document ID in message")
				time.Sleep(retryInterval)
				continue
			}

			// Get document from dataset
			document, err := eh.ds.GetDocument(ctx, indexName, docID)
			if err != nil {
				// Check if document ID is EmptyDocumentID
				if docID == interfaces.EmptyDocumentID {
					logger.Infof("Empty document ID detected, skipping: %s", docID)

					// Update resource index name
					indexName := getIndexName(resource.ID, buildTaskInfo.ID)
					err = updateResourceIndexName(ctx, resource, eh.resAccess, eh.ds, indexName)
					if err != nil {
						logger.Errorf("Failed to update resource index name: %v", err)
						time.Sleep(retryInterval)
						continue
					}

					// Commit the message to avoid reprocessing
					_ = eh.kafkaAccess.CommitMessages(ctx, reader, msg)
					logger.Infof("CommitMessages")
					return nil
				}
				logger.Errorf("Failed to get document %s: %v", docID, err)
				time.Sleep(retryInterval)
				continue
			}

			// 处理结果并进行嵌入
			updateDoc := make(map[string]any)
			fields := []string{}
			words := []string{}
			for _, field := range embeddingFields {
				if value, exists := document[field]; exists {
					if text, ok := value.(string); ok && text != "" {
						fields = append(fields, field)
						words = append(words, text)
					}
				}
			}
			vectorResp, err := eh.mfa.GetVector(ctx, buildTaskInfo.EmbeddingModel, words)
			if err != nil || len(vectorResp) != len(words) {
				logger.Errorf("GetVector failed: %v", err)
				time.Sleep(retryInterval)
				continue
			}
			for i, field := range fields {
				if resp := vectorResp[i]; resp.Vector != nil {
					updateDoc[field+"_vector"] = resp.Vector
				}
			}

			// 保存嵌入结果
			if len(updateDoc) > 0 {
				updateReq := map[string]any{
					"id":       docID,
					"document": updateDoc,
				}
				_, err := eh.ds.UpsertDocuments(ctx, indexName, []map[string]any{updateReq})
				if err != nil {
					logger.Errorf("Update document failed: %v", err)
					time.Sleep(retryInterval)
					continue
				}
				totalProcessed++

				// 批量更新任务状态
				if totalProcessed%batchUpdateThreshold == 0 || time.Since(lastUpdateTime) > updateInterval {
					_ = eh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"vectorizedCount": totalProcessed})
					lastUpdateTime = time.Now()
				}
			}

			// Commit the message to avoid reprocessing
			if err := eh.kafkaAccess.CommitMessages(ctx, reader, msg); err != nil {
				logger.Errorf("Failed to commit message: %v", err)
			}
		}
	}
}
