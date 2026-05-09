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
	"math"
	"strings"
	"time"

	"vega-backend/interfaces"

	"github.com/bytedance/sonic"
	"github.com/hibiken/asynq"
	"github.com/segmentio/kafka-go"
)

func getIndexName(resourceID, buildTaskID string) string {
	return interfaces.BUILD_PREFIX + "-" + resourceID + "-" + buildTaskID
}

func getEmbeddingTopic(resourceID, buildTaskID string) string {
	return fmt.Sprintf("%s-%s-%s-embedding", interfaces.BUILD_PREFIX, resourceID, buildTaskID)
}

func getOldDocID(primaryKeyValues []interfaces.KeyValue) string {
	// 将primaryKeyValues中的所有值拼接成id
	var idBuilder strings.Builder
	for _, item := range primaryKeyValues {
		fmt.Fprintf(&idBuilder, "%v", item.Value)
		fmt.Fprintf(&idBuilder, "-")
	}
	return idBuilder.String()
}

func getNewDocID(primaryKeyValues []interfaces.KeyValue, document map[string]any) string {
	// 构造新的文档ID，确保与oldDocID的拼接顺序相同
	var newDocIDBuilder strings.Builder
	for _, item := range primaryKeyValues {
		if value, ok := document[item.Key]; ok {
			fmt.Fprintf(&newDocIDBuilder, "%v", value)
			fmt.Fprintf(&newDocIDBuilder, "-")
		}
	}
	return newDocIDBuilder.String()
}

// updateResourceIndexName updates the index name of a resource
func updateResourceIndexName(ctx context.Context, resource *interfaces.Resource, ra interfaces.ResourceAccess, ds interfaces.DatasetService, indexName string) error {
	if resource.LocalIndexName == "" {
		resource.LocalIndexName = indexName
		return ra.Update(ctx, resource)
	}

	if resource.LocalIndexName != indexName {
		err := ds.Delete(ctx, resource.LocalIndexName)
		if err != nil {
			return fmt.Errorf("delete local index failed: %w", err)
		}
		resource.LocalIndexName = indexName
		return ra.Update(ctx, resource)
	}
	return nil
}

// createLocalIndex creates a local index with vector fields for embedding
func createLocalIndex(ctx context.Context, ds interfaces.DatasetService, buildTask *interfaces.BuildTask, resource *interfaces.Resource) error {
	newResource := *resource
	newResource.ID = getIndexName(resource.ID, buildTask.ID)
	exist, err := ds.CheckExist(ctx, newResource.ID)
	if err != nil {
		return fmt.Errorf("check dataset exist failed: %w", err)
	}
	if exist {
		return nil
	}
	if buildTask.EmbeddingFields != "" {
		embeddingFields := strings.Split(buildTask.EmbeddingFields, ",")
		var newSchema []*interfaces.Property
		if resource.SchemaDefinition != nil {
			newSchema = append(newSchema, resource.SchemaDefinition...)
		}
		for _, field := range embeddingFields {
			field = strings.TrimSpace(field)
			if field != "" {
				vectorProperty := &interfaces.Property{
					Name: field + "_vector",
					Type: interfaces.DataType_Vector,
					Features: []interfaces.PropertyFeature{
						{
							FeatureType: interfaces.DataType_Vector,
							Config: map[string]any{
								"dimension": buildTask.ModelDimensions,
								"method": map[string]any{
									"name":   "hnsw",
									"engine": "lucene",
									"parameters": map[string]any{
										"ef_construction": 256,
									},
								},
							},
						},
					},
				}
				newSchema = append(newSchema, vectorProperty)
			}
		}
		newResource.SchemaDefinition = newSchema
	}
	return ds.Create(ctx, &newResource)
}

// sendEmbeddingTask sends a embedding task to the queue
func sendEmbeddingTask(client *asynq.Client, taskID string) error {
	embeddingTaskMsg := interfaces.EmbeddingBuildTaskMessage{
		TaskID: taskID,
	}
	payload, err := sonic.Marshal(embeddingTaskMsg)
	if err != nil {
		return err
	} else {
		embeddingTask := asynq.NewTask(interfaces.BuildTaskTypeEmbedding, payload)
		_, err = client.Enqueue(embeddingTask,
			asynq.Queue(interfaces.DefaultQueue),
			asynq.TaskID(fmt.Sprintf("%s-%s", interfaces.BuildTaskTypeEmbedding, taskID)),
			asynq.MaxRetry(interfaces.BUILD_TASK_MAX_RETRY_COUNT),
			asynq.Timeout(math.MaxInt64),                                                  // 永不超时
			asynq.Deadline(time.Unix(math.MaxInt64/1000000000, math.MaxInt64%1000000000)), // 永不过期
		)
		if err != nil {
			if !errors.Is(err, asynq.ErrTaskIDConflict) {
				return err
			}
		}
	}
	return nil
}

// sendEmbeddingMessage sends a document ID to Kafka for embedding
func sendEmbeddingMessage(ctx context.Context, writer *kafka.Writer, kafkaAccess interfaces.KafkaAccess, docIDs []string) error {
	for _, docID := range docIDs {
		// Create message
		messageData := map[string]any{
			"document_id": docID,
		}
		messageBytes, err := sonic.Marshal(messageData)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		// Write message to Kafka
		// Use docID + timestamp as key to avoid conflicts even if document is modified multiple times
		err = kafkaAccess.WriteMessages(ctx, writer, []kafka.Message{
			{
				Key:   []byte(fmt.Sprintf("%s-%d", docID, time.Now().UnixNano())),
				Value: messageBytes,
			},
		}...)
		if err != nil {
			return fmt.Errorf("failed to write message to Kafka: %w", err)
		}
	}
	return nil
}
