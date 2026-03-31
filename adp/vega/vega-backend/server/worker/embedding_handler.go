// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package worker provides background workers for VEGA Manager.
package worker

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/common"
	taskAccess "vega-backend/drivenadapters/build_task"
	resourceAccess "vega-backend/drivenadapters/resource"
	"vega-backend/interfaces"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connectors"
	opensearchConnector "vega-backend/logics/connectors/local/index/opensearch"
	"vega-backend/logics/dataset"
)

// embeddingHandler handles embedding tasks.
type embeddingHandler struct {
	appSetting *common.AppSetting
	taskAccess interfaces.BuildTaskAccess
	resAccess  interfaces.ResourceAccess
	cs         interfaces.CatalogService
	ds         interfaces.DatasetService
	connector  connectors.IndexConnector
}

// NewEmbeddingHandler creates a new embedding handler.
func NewEmbeddingHandler(appSetting *common.AppSetting) *embeddingHandler {
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
		appSetting: appSetting,
		taskAccess: taskAccess.NewBuildTaskAccess(appSetting),
		resAccess:  resourceAccess.NewResourceAccess(appSetting),
		cs:         catalog.NewCatalogService(appSetting),
		ds:         dataset.NewDatasetService(appSetting),
		connector:  connector.(connectors.IndexConnector),
	}
}

// HandleTask handles an embedding task from the queue.
func (eh *embeddingHandler) HandleTask(ctx context.Context, task *asynq.Task) error {
	var msg interfaces.EmbeddingTaskMessage
	if err := sonic.Unmarshal(task.Payload(), &msg); err != nil {
		logger.Errorf("Failed to unmarshal task message: %v", err)
		return err
	}

	taskID := msg.TaskID
	resourceID := msg.ResourceID
	logger.Infof("Starting embedding for task: %s, resource: %s", taskID, resourceID)

	// Get resource info
	resource, err := eh.resAccess.GetByID(ctx, resourceID)
	if err != nil {
		logger.Errorf("Failed to get resource for task %s: %v", taskID, err)
		return err
	}
	if resource == nil {
		logger.Errorf("Resource not found for task %s, resourceID: %s", taskID, resourceID)
		return fmt.Errorf("resource not found")
	}

	// Update task status to running
	_ = eh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusRunning})

	// Execute embedding
	embed_err := eh.executeEmbedding(ctx, resource, taskID)
	if embed_err != nil {
		// Update task status to failed
		_ = eh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusFailed, "errorMsg": embed_err.Error()})
		return embed_err
	}

	// Update task status to completed
	_ = eh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusCompleted})

	logger.Infof("Embedding completed for task: %s, resource: %s", taskID, resourceID)
	return nil
}

// executeEmbedding executes the embedding logic
func (eh *embeddingHandler) executeEmbedding(ctx context.Context, resource *interfaces.Resource, taskID string) error {
	sourceMetadata := resource.SourceMetadata
	if sourceMetadata == nil {
		return fmt.Errorf("source metadata is nil")
	}

	// get vector fields from resource.schema_definition
	embeddingFields := []string{}
	sourceFields := []string{}
	for _, prop := range resource.SchemaDefinition {
		if prop.Type == interfaces.DataType_Vector {
			embeddingFields = append(embeddingFields, prop.Name)
			sourceFields = append(sourceFields, prop.Features[0].RefProperty)
		}
	}

	if len(embeddingFields) == 0 {
		return fmt.Errorf("no embedding fields specified")
	}

	// 批量读取数据并进行嵌入处理
	batchSize := 100
	firstQuery := true
	totalProcessed := int64(0)

	if err := eh.connector.Connect(ctx); err != nil {
		logger.Errorf("Connect failed: %v", err)
		return fmt.Errorf("connect failed: %w", err)
	}
	defer eh.connector.Close(ctx)

	for {
		// 从opensearch中查询embeddingFields[0]值为空的记录，用记录中的sourceFields字段生成向量来填充embeddingFields字段
		params := &interfaces.ResourceDataQueryParams{
			Limit:        batchSize,
			OutputFields: sourceFields,
			NeedTotal:    firstQuery,
		}

		// 添加过滤条件
		params.FilterCondCfg = &interfaces.FilterCondCfg{
			Name:      embeddingFields[0],
			Operation: "empty",
		}

		// 执行查询
		result, err := eh.connector.ExecuteQuery(ctx, resource.ID, resource, params)
		if err != nil {
			logger.Errorf("Execute query failed: %v", err)
			return fmt.Errorf("execute query failed: %w", err)
		}

		if firstQuery {
			firstQuery = false
		}

		// 处理结果并进行嵌入
		updateRequests := make([]map[string]any, 0, len(result.Rows))
		for _, item := range result.Rows {
			// 对需要嵌入的字段进行处理
			for _, field := range embeddingFields {
				if value, exists := item[field]; exists {
					if _, ok := value.(string); ok {
						// 这里应该调用嵌入服务生成向量
						// 示例：vector := embeddingService.Embed(text)
						// 为了演示，我们使用一个模拟的向量
						// 256-dimensional vector
						vector := make([]float64, 256)
						for i := range vector {
							vector[i] = 0.1 + float64(i%10)/10.0
						}
						item[field+"_embedding"] = vector
					}
				}
			}
			// 检查是否有 _id 字段
			if docID, ok := item["_id"].(string); ok {
				// 移除 _id 字段，因为它不应该在文档体中
				delete(item, "_id")
				// 创建符合 UpdateDocuments 要求的更新请求
				updateReq := map[string]any{
					"id":       docID,
					"document": item,
				}
				updateRequests = append(updateRequests, updateReq)
			}
		}

		// 保存嵌入结果
		if len(updateRequests) > 0 {
			err = eh.ds.UpdateDocuments(ctx, resource.ID, updateRequests)
			if err != nil {
				logger.Errorf("Update documents failed: %v", err)
				return fmt.Errorf("update documents failed: %w", err)
			}
			totalProcessed += int64(len(updateRequests))
		}

		// 更新任务状态
		_ = eh.taskAccess.UpdateStatus(ctx, taskID, map[string]interface{}{"status": interfaces.BuildTaskStatusRunning, "totalCount": result.Total, "syncedCount": totalProcessed})

		if len(result.Rows) < batchSize {
			break
		}
	}

	// 返回成功
	return nil
}
