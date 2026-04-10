// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package worker provides background workers for VEGA Manager.
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/segmentio/kafka-go"

	"vega-backend/common"
	asynqAccess "vega-backend/drivenadapters/asynq"
	taskAccess "vega-backend/drivenadapters/build_task"
	kafkaAccess "vega-backend/drivenadapters/kafka"
	resourceAccess "vega-backend/drivenadapters/resource"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/dataset"
	"vega-backend/logics/filter_condition"
)

// batchBuildHandler handles build tasks.
type batchBuildHandler struct {
	appSetting  *common.AppSetting
	taskAccess  interfaces.BuildTaskAccess
	resAccess   interfaces.ResourceAccess
	cs          interfaces.CatalogService
	ds          interfaces.DatasetService
	client      *asynq.Client
	kafkaAccess interfaces.KafkaAccess
}

// NewBatchBuildHandler creates a new build handler.
func NewBatchBuildHandler(appSetting *common.AppSetting) *batchBuildHandler {
	return &batchBuildHandler{
		appSetting:  appSetting,
		taskAccess:  taskAccess.NewBuildTaskAccess(appSetting),
		resAccess:   resourceAccess.NewResourceAccess(appSetting),
		cs:          catalog.NewCatalogService(appSetting),
		ds:          dataset.NewDatasetService(appSetting),
		client:      asynqAccess.NewAsynqAccess(appSetting).CreateClient(context.Background()),
		kafkaAccess: kafkaAccess.NewKafkaAccess(appSetting),
	}
}

// HandleTask handles a build task from the queue.
func (bh *batchBuildHandler) HandleTask(ctx context.Context, task *asynq.Task) error {
	var msg interfaces.BatchBuildTaskMessage
	if err := sonic.Unmarshal(task.Payload(), &msg); err != nil {
		logger.Errorf("Failed to unmarshal task message: %v", err)
		return err
	}

	taskID := msg.TaskID
	buildTaskInfo, err := bh.taskAccess.GetByID(ctx, taskID)
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
	resource, err := bh.resAccess.GetByID(ctx, resourceID)
	if err != nil {
		logger.Errorf("Failed to get resource for task %s: %v", taskID, err)
		return err
	}
	if resource == nil {
		logger.Errorf("Resource not found for task %s, resourceID: %s", taskID, resourceID)
		// Resource not found, return nil to  stop the task
		return nil
	}

	// Execute build
	err = bh.executeBuild(ctx, resource, buildTaskInfo, msg.ExecuteType)
	if err != nil {
		// Update task status to failed
		logger.Errorf("Build failed for task %s: %w", taskID, err)
		_ = bh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"errorMsg": err.Error()})
		return err
	}

	logger.Infof("Build completed for task: %s, resource: %s", taskID, resourceID)
	return nil
}

// executeBuild executes the build logic
func (bh *batchBuildHandler) executeBuild(ctx context.Context, resource *interfaces.Resource, buildTaskInfo *interfaces.BuildTask, executeType string) error {
	// Update task status to running
	err := bh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"status": interfaces.BuildTaskStatusRunning})
	if err != nil {
		return fmt.Errorf("update build task status failed: %w", err)
	}

	lastSyncedMark := buildTaskInfo.SyncedMark
	if executeType == interfaces.BuildTaskExecuteTypeFull {
		lastSyncedMark = ""
	}

	var batchFields []string
	var lastBatchFieldValues []any
	if lastSyncedMark != "" {
		// syncMark format : {"filed1_name":field1_value,"filed2_name":field2_value}
		var syncedMark map[string]interface{}
		if err := sonic.Unmarshal([]byte(lastSyncedMark), &syncedMark); err != nil {
			return fmt.Errorf("failed to unmarshal synced mark: %w", err)
		}
		// Extract field names from synced mark
		batchFields = make([]string, 0, len(syncedMark))
		for field := range syncedMark {
			batchFields = append(batchFields, field)
			lastBatchFieldValues = append(lastBatchFieldValues, syncedMark[field])
		}
	}

	if len(batchFields) == 0 {
		sourceMetadata := resource.SourceMetadata
		if sourceMetadata == nil {
			return fmt.Errorf("source metadata is nil")
		}
		// Extract ordered keys
		orderedKeys := []string{}
		if orderedKeysInterface, ok := sourceMetadata["ordered_keys"]; ok {
			if keysInterface, ok := orderedKeysInterface.([]interface{}); ok {
				// Handle []interface{} type
				for _, key := range keysInterface {
					if keyStr, ok := key.(string); ok {
						orderedKeys = append(orderedKeys, keyStr)
					}
				}
			}
		}

		// Find suitable fields for batch reading
		// In a real implementation, these fields would be used for batch reading
		if len(orderedKeys) > 0 {
			// Use all ordered keys as batch fields
			batchFields = orderedKeys
		} else {
			return fmt.Errorf("no ordered key found")
		}
	}

	// Get catalog for MySQL connection
	catalog, err := bh.cs.GetByID(ctx, resource.CatalogID, true)
	if err != nil {
		return fmt.Errorf("get catalog failed: %w", err)
	}
	if catalog == nil {
		logger.Errorf("Catalog not found for task %s, catalogID: %s", buildTaskInfo.ID, resource.CatalogID)
		// Catalog not found, return nil to  stop the task
		return nil
	}

	// Batch read data from MySQL and write to dataset
	batchSize := 1000
	firstQuery := true

	// get total rows from MySQL
	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		return fmt.Errorf("create connector instance failed: %w", err)
	}
	if err := connector.Connect(ctx); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}
	defer connector.Close(ctx)
	tableConnector, ok := connector.(connectors.TableConnector)
	if !ok {
		return fmt.Errorf("connector is not a table connector")
	}
	// Construct OutputFields from SchemaDefinition
	outputFields := make([]string, 0)
	batchFieldExistsMap := make(map[string]bool)
	hasVectorField := false

	// Extract fields from SchemaDefinition
	if resource.SchemaDefinition != nil {
		for _, prop := range resource.SchemaDefinition {
			// vector need to be processed separately
			if prop.Type == "vector" {
				hasVectorField = true
				continue
			}
			outputFields = append(outputFields, prop.Name)
			batchFieldExistsMap[prop.Name] = true
		}
	}

	// Add batch fields that are not in outputFields
	// Build sort fields
	sortFields := make([]*interfaces.SortField, len(batchFields))
	for i, field := range batchFields {
		if !batchFieldExistsMap[field] {
			outputFields = append(outputFields, field)
		}
		sortFields[i] = &interfaces.SortField{
			Field: field,
		}
	}

	var writer *kafka.Writer
	if hasVectorField {
		topic := fmt.Sprintf("%s-%s-embedding", KafkaTopicPrefix, resource.ID)
		// Create Kafka writer
		writer, err = bh.kafkaAccess.NewWriter(ctx, topic)
		if err != nil {
			return fmt.Errorf("failed to create Kafka writer: %w", err)
		}

		err = bh.kafkaAccess.CreateTopic(ctx, topic)
		if err != nil {
			return fmt.Errorf("failed to create Kafka topic: %w", err)
		}
		defer bh.kafkaAccess.CloseWriter(writer)
	}

	syncedCount := buildTaskInfo.SyncedCount
	for {
		// Check task status before each batch
		taskStatus, err := bh.taskAccess.GetStatus(ctx, buildTaskInfo.ID)
		if err != nil {
			return fmt.Errorf("failed to get task status: %w", err)
		}

		// Handle stopping status
		if taskStatus == interfaces.BuildTaskStatusStopping {
			// Task is stopping, exit the loop
			logger.Infof("Task %s is stopping, exiting...", buildTaskInfo.ID)
			// Update task status to stopped
			err = bh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"status": interfaces.BuildTaskStatusStopped})
			if err != nil {
				return fmt.Errorf("update build task status failed: %w", err)
			}
			return nil
		}

		params := &interfaces.ResourceDataQueryParams{
			Limit:        batchSize,
			Sort:         sortFields,
			OutputFields: outputFields,
			NeedTotal:    firstQuery,
		}

		// Add filter condition for batch fields if we have last values
		if len(lastBatchFieldValues) > 0 {
			// Build AND condition for multiple batch fields
			subConditions := make([]*interfaces.FilterCondCfg, len(batchFields))
			for i, field := range batchFields {
				subConditions[i] = &interfaces.FilterCondCfg{
					Name:        field,
					Operation:   "gt",
					ValueOptCfg: interfaces.ValueOptCfg{Value: lastBatchFieldValues[i], ValueFrom: interfaces.ValueFrom_Const},
				}
			}
			params.FilterCondCfg = &interfaces.FilterCondCfg{
				Operation: "and",
				SubConds:  subConditions,
			}

			// Convert FilterCondCfg to ActualFilterCond
			fieldMap := map[string]*interfaces.Property{}
			if resource.SchemaDefinition != nil {
				for _, prop := range resource.SchemaDefinition {
					fieldMap[prop.Name] = prop
				}
			}
			actualFilterCond, err := filter_condition.NewFilterCondition(ctx, params.FilterCondCfg, fieldMap)
			if err != nil {
				return fmt.Errorf("create filter condition failed: %w", err)
			}
			params.ActualFilterCond = actualFilterCond
		}
		result, err := tableConnector.ExecuteQuery(ctx, resource, params)
		if err != nil {
			return fmt.Errorf("execute query failed: %w", err)
		}

		totalRows := result.Total
		readRows := len(result.Rows)

		documents := make([]map[string]any, 0, readRows)
		for _, item := range result.Rows {
			// If batch fields are not in SchemaDefinition, remove them from item
			for _, field := range batchFields {
				if !batchFieldExistsMap[field] {
					delete(item, field)
				}
			}
			documents = append(documents, item)
		}

		if readRows > 0 {
			// Update lastBatchFieldValues with the last values in this batch
			newSyncedMark := map[string]any{}
			lastItem := result.Rows[readRows-1]
			lastBatchFieldValues = make([]any, len(batchFields))
			for i, field := range batchFields {
				if value, exists := lastItem[field]; exists {
					lastBatchFieldValues[i] = value
					newSyncedMark[field] = value
				}
			}

			docIDs, err := bh.ds.CreateDocuments(ctx, resource.ID, documents)
			if err != nil {
				return fmt.Errorf("create documents failed: %w", err)
			}

			syncedCount += int64(readRows)
			// Set firstQuery to false after the first query
			updates := map[string]interface{}{"syncedCount": syncedCount}
			if firstQuery {
				firstQuery = false
				updates["totalCount"] = int64(totalRows)
			}
			if len(newSyncedMark) > 0 {
				syncedMarkStr, err := sonic.MarshalString(newSyncedMark)
				if err != nil {
					return fmt.Errorf("failed to marshal synced mark: %w", err)
				} else {
					updates["syncedMark"] = syncedMarkStr
				}
			}
			err = bh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, updates)
			if err != nil {
				return fmt.Errorf("update build task status failed: %w", err)
			}

			// Send document IDs to Kafka for embedding
			if len(docIDs) > 0 {
				// Create messages
				var messages []kafka.Message
				for _, docID := range docIDs {
					messageData := map[string]any{
						"document_id": docID,
					}
					messageBytes, err := sonic.Marshal(messageData)
					if err != nil {
						return fmt.Errorf("failed to marshal message: %w", err)
					}
					messages = append(messages, kafka.Message{
						Key:   []byte(docID),
						Value: messageBytes,
					})
				}

				// Write messages to Kafka
				if len(messages) > 0 {
					err = bh.kafkaAccess.WriteMessages(ctx, writer, messages...)
					if err != nil {
						return fmt.Errorf("failed to write messages to Kafka: %w", err)
					}
				}
			}
		}

		if readRows < batchSize {
			if hasVectorField {
				// sync complete, push a empty document to trigger embedding
				messageData := map[string]any{
					"document_id": interfaces.EmptyDocumentID,
				}
				messageBytes, err := sonic.Marshal(messageData)
				if err != nil {
					return fmt.Errorf("failed to marshal message: %w", err)
				}

				err = bh.kafkaAccess.WriteMessages(ctx, writer, []kafka.Message{
					{
						Key:   []byte(interfaces.EmptyDocumentID),
						Value: messageBytes,
					},
				}...)
				if err != nil {
					return fmt.Errorf("failed to write messages to Kafka: %w", err)
				}
			}
			break
		}
	}

	if hasVectorField {
		// put embedding task to queue
		embeddingTaskMsg := interfaces.EmbeddingBuildTaskMessage{
			TaskID: buildTaskInfo.ID,
		}
		payload, err := sonic.Marshal(embeddingTaskMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal embedding task message: %w", err)
		} else {
			asynqTask := asynq.NewTask(interfaces.BuildTaskTypeEmbedding, payload)
			_, err := bh.client.Enqueue(asynqTask,
				asynq.Queue(interfaces.DefaultQueue),
				asynq.MaxRetry(interfaces.DATASET_BUILD_MAX_RETRY_COUNT),
				asynq.Timeout(0),                // 永不超时
				asynq.Deadline(time.Unix(0, 0)), // 永不过期
			)
			if err != nil {
				return fmt.Errorf("failed to enqueue embedding task: %w", err)
			} else {
				logger.Infof("Embedding task queued for resource: %s", resource.Name)
			}
		}
	} else {
		// Update task status to completed
		err = bh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"status": interfaces.BuildTaskStatusCompleted})
		if err != nil {
			return fmt.Errorf("failed to update task status: %w", err)
		}
	}

	return nil
}
