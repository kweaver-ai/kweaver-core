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

	"vega-backend/common"
	asynqAccess "vega-backend/drivenadapters/asynq"
	taskAccess "vega-backend/drivenadapters/build_task"
	resourceAccess "vega-backend/drivenadapters/resource"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/dataset"
	"vega-backend/logics/filter_condition"
)

// buildHandler handles build tasks.
type buildHandler struct {
	appSetting *common.AppSetting
	taskAccess interfaces.BuildTaskAccess
	resAccess  interfaces.ResourceAccess
	cs         interfaces.CatalogService
	ds         interfaces.DatasetService
	client     *asynq.Client
}

// NewBuildHandler creates a new build handler.
func NewBuildHandler(appSetting *common.AppSetting) *buildHandler {
	return &buildHandler{
		appSetting: appSetting,
		taskAccess: taskAccess.NewBuildTaskAccess(appSetting),
		resAccess:  resourceAccess.NewResourceAccess(appSetting),
		cs:         catalog.NewCatalogService(appSetting),
		ds:         dataset.NewDatasetService(appSetting),
		client:     asynqAccess.NewAsynqAccess(appSetting).CreateClient(context.Background()),
	}
}

// HandleTask handles a build task from the queue.
func (bh *buildHandler) HandleTask(ctx context.Context, task *asynq.Task) error {
	var msg interfaces.BuildTaskMessage
	if err := sonic.Unmarshal(task.Payload(), &msg); err != nil {
		logger.Errorf("Failed to unmarshal task message: %v", err)
		return err
	}

	taskID := msg.TaskID
	buildTaskInfo, err := bh.taskAccess.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("get build task failed: %w", err)
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
		return fmt.Errorf("resource not found")
	}

	// Update task status to running
	_ = bh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusRunning})

	// Execute build
	err = bh.executeBuild(ctx, resource, buildTaskInfo)
	if err != nil {
		// Update task status to failed
		_ = bh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusFailed, "errorMsg": err.Error()})
		return nil
	} else {
		// Update task status to completed
		_ = bh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusCompleted})
		// put embedding task to queue
		embeddingTaskMsg := interfaces.EmbeddingTaskMessage{
			TaskID:     taskID,
			ResourceID: resourceID,
		}
		payload, err := sonic.Marshal(embeddingTaskMsg)
		if err != nil {
			logger.Errorf("Failed to marshal embedding task message: %v", err)
		} else {
			asynqTask := asynq.NewTask(interfaces.EmbeddingTaskType, payload)
			_, err := bh.client.Enqueue(asynqTask,
				asynq.Queue("default"),
				asynq.MaxRetry(10),
				asynq.Timeout(30*time.Minute),
				asynq.Deadline(time.Now().Add(24*time.Hour)),
			)
			if err != nil {
				logger.Errorf("Failed to enqueue embedding task: %v", err)
			} else {
				logger.Infof("Embedding task queued for resource: %s", resourceID)
			}
		}
	}

	logger.Infof("Build completed for task: %s, resource: %s", taskID, resourceID)
	return nil
}

// executeBuild executes the build logic
func (bh *buildHandler) executeBuild(ctx context.Context, resource *interfaces.Resource, buildTaskInfo *interfaces.BuildTask) error {
	sourceMetadata := resource.SourceMetadata
	if sourceMetadata == nil {
		return fmt.Errorf("source metadata is nil")
	}

	lastSyncedMark := ""
	if buildTaskInfo.SyncedMark != "" {
		// is retry build, continue from SyncedMark
		lastSyncedMark = buildTaskInfo.SyncedMark
	} else if buildTaskInfo.Mode == interfaces.BuildTaskModeIncremental {
		// get batch fields from last SyncedMark
		lastTask, err := bh.taskAccess.GetLastBuildTask(ctx, resource.ID)
		if err != nil {
			return fmt.Errorf("get last build task failed: %w", err)
		}
		if lastTask != nil {
			lastSyncedMark = lastTask.SyncedMark
		}
	}

	var batchFields []string
	var lastBatchFieldValues []any
	if lastSyncedMark != "" {
		// syncMark format : {"filed1_name":field1_value,"filed2_name":field2_value}
		var syncedMark map[string]interface{}
		if err := sonic.Unmarshal([]byte(lastSyncedMark), &syncedMark); err != nil {
			logger.Errorf("Failed to unmarshal synced mark: %v", err)
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
		if buildTaskInfo.Mode == interfaces.BuildTaskModeRealtime {
			return fmt.Errorf("build task not support realtime mode now")
		}
		// Extract primary keys
		primaryKeys := []string{}
		if primaryKeysInterface, ok := sourceMetadata["primary_keys"]; ok {
			if keys, ok := primaryKeysInterface.([]string); ok {
				primaryKeys = keys
			} else if keysInterface, ok := primaryKeysInterface.([]interface{}); ok {
				// Handle []interface{} type
				for _, key := range keysInterface {
					if keyStr, ok := key.(string); ok {
						primaryKeys = append(primaryKeys, keyStr)
					}
				}
			}
		}

		// Extract index columns
		indexColumns := []string{}
		if indexColumnsInterface, ok := sourceMetadata["indices"]; ok {
			if indices, ok := indexColumnsInterface.([]string); ok {
				indexColumns = indices
			} else if indicesInterface, ok := indexColumnsInterface.([]interface{}); ok {
				// Handle []interface{} type
				for _, index := range indicesInterface {
					if indexStr, ok := index.(string); ok {
						indexColumns = append(indexColumns, indexStr)
					}
				}
			}
		}

		// Find suitable fields for batch reading
		// In a real implementation, these fields would be used for batch reading
		if len(primaryKeys) > 0 {
			// Use all primary keys as batch fields
			batchFields = primaryKeys
		} else if len(indexColumns) > 0 {
			// Use all index columns as batch fields
			batchFields = indexColumns
		} else {
			return fmt.Errorf("no primary key or index column found")
		}
	}

	// Get catalog for MySQL connection
	catalog, err := bh.cs.GetByID(ctx, resource.CatalogID, true)
	if err != nil {
		return fmt.Errorf("get catalog failed: %w", err)
	}
	if catalog == nil {
		return fmt.Errorf("catalog not found")
	}

	// Batch read data from MySQL and write to dataset
	batchSize := 1000
	firstQuery := true

	// get total rows from MySQL
	connector, err := factory.GetFactory().CreateConnectorInstance(ctx, catalog.ConnectorType, catalog.ConnectorCfg)
	if err != nil {
		logger.Errorf("Create connector instance failed: %v", err)
		return fmt.Errorf("create connector instance failed: %w", err)
	}
	if err := connector.Connect(ctx); err != nil {
		logger.Errorf("Connect failed: %v", err)
		return fmt.Errorf("connect failed: %w", err)
	}
	defer connector.Close(ctx)
	tableConnector, ok := connector.(connectors.TableConnector)
	if !ok {
		logger.Errorf("Connector is not a table connector")
		return fmt.Errorf("connector is not a table connector")
	}
	// Construct OutputFields from SchemaDefinition
	outputFields := make([]string, 0)
	batchFieldExistsMap := make(map[string]bool)

	// Extract fields from SchemaDefinition
	if resource.SchemaDefinition != nil {
		for _, prop := range resource.SchemaDefinition {
			// vector need to be processed separately
			if prop.Type == "vector" {
				continue
			}
			outputFields = append(outputFields, prop.Name)
			batchFieldExistsMap[prop.Name] = true
		}
	}

	// Add batch fields that are not in outputFields
	for _, field := range batchFields {
		if !batchFieldExistsMap[field] {
			outputFields = append(outputFields, field)
		}
	}

	syncedCount := buildTaskInfo.SyncedCount
	for {
		// Build sort fields
		sortFields := make([]*interfaces.SortField, len(batchFields))
		for i, field := range batchFields {
			sortFields[i] = &interfaces.SortField{
				Field: field,
			}
		}

		params := &interfaces.ResourceDataQueryParams{
			Limit:        batchSize,
			Sort:         sortFields,
			OutputFields: outputFields,
			NeedTotal:    firstQuery,
		}

		// Add filter condition for batch fields if we have last values
		if len(lastBatchFieldValues) > 0 && !firstQuery {
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
				logger.Errorf("Create filter condition failed: %v", err)
				return fmt.Errorf("create filter condition failed: %w", err)
			}
			params.ActualFilterCond = actualFilterCond
		}
		result, err := tableConnector.ExecuteQuery(ctx, resource, params)
		if err != nil {
			logger.Errorf("Execute query failed: %v", err)
			return fmt.Errorf("execute query failed: %w", err)
		}

		totalRows := result.Total
		readRows := int64(len(result.Rows))

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

		// Update lastBatchFieldValues with the last values in this batch
		newSyncedMark := map[string]any{}
		if readRows > 0 {
			lastItem := result.Rows[readRows-1]
			lastBatchFieldValues = make([]any, len(batchFields))
			for i, field := range batchFields {
				if value, exists := lastItem[field]; exists {
					lastBatchFieldValues[i] = value
					newSyncedMark[field] = value
				}
			}
		}
		if readRows > 0 {
			_, err = bh.ds.CreateDocuments(ctx, resource.ID, documents)
			if err != nil {
				logger.Errorf("Create documents failed: %v", err)
				return fmt.Errorf("create documents failed: %w", err)
			}
		}

		syncedCount += readRows
		// Set firstQuery to false after the first query
		if firstQuery {
			firstQuery = false
			syncedMarkStr, err := sonic.MarshalString(newSyncedMark)
			if err != nil {
				logger.Errorf("Failed to marshal synced mark: %v", err)
			}
			_ = bh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"totalCount": int64(totalRows), "syncedCount": syncedCount, "syncedMark": syncedMarkStr})
		} else {
			syncedMarkStr, err := sonic.MarshalString(newSyncedMark)
			if err != nil {
				logger.Errorf("Failed to marshal synced mark: %v", err)
			}
			_ = bh.taskAccess.UpdateStatus(ctx, buildTaskInfo.ID, map[string]interface{}{"syncedCount": syncedCount, "syncedMark": syncedMarkStr})
		}

		if len(result.Rows) < batchSize {
			break
		}
	}

	return nil
}
