// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

const (
	SMALL_MODEL_TYPE_EMBEDDING = "embedding"
	DEFAULT_EMBEDDING_MODEL    = "embedding"
)

type VectorResp struct {
	Object string    `json:"object"`
	Vector []float32 `json:"embedding"`
	Index  int       `json:"index"`
}

type SmallModel struct {
	ModelID      string `json:"model_id"`
	ModelName    string `json:"model_name"`
	ModelType    string `json:"model_type"`
	EmbeddingDim int    `json:"embedding_dim"`
	BatchSize    int    `json:"batch_size"`
	MaxTokens    int    `json:"max_tokens"`
}

// ModelFactoryAccess 定义模型工厂相关的访问接口
//
//go:generate mockgen -source ../interfaces/model_factory_access.go -destination ../interfaces/mock/mock_model_factory_access.go
type ModelFactoryAccess interface {
	GetModelByName(ctx context.Context, modelName string) (*SmallModel, error)

	GetVector(ctx context.Context, modelName string, words []string) ([]*VectorResp, error)
}
