// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	cond "ontology-query/common/condition"
)

const (
	SMALL_MODEL_TYPE_EMBEDDING = "embedding"
)

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
	// GetVector 根据输入字符串数组获取对应的向量数组
	// 参数：
	//   - ctx: 上下文对象
	//   - texts: 输入字符串数组
	// 返回：
	//   - [][]float32: 等长的向量数组，每个向量对应一个输入字符串
	//   - error: 错误信息
	GetVector(ctx context.Context, model *SmallModel, words []string) ([]cond.VectorResp, error)

	GetModelByID(ctx context.Context, modelID string) (*SmallModel, error)
}
