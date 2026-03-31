// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package model_factory

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	"ontology-query/interfaces"
)

var (
	mfAccessOnce sync.Once
	mfAccess     interfaces.ModelFactoryAccess
)

type modelFactoryAccess struct {
	appSetting   *common.AppSetting
	httpClient   rest.HTTPClient
	mfManagerUrl string
	mfAPIUrl     string
}

// NewModelFactoryAccess 创建模型工厂访问实例
func NewModelFactoryAccess(appSetting *common.AppSetting) interfaces.ModelFactoryAccess {
	mfAccessOnce.Do(func() {
		mfAccess = &modelFactoryAccess{
			appSetting:   appSetting,
			httpClient:   common.NewHTTPClient(),
			mfManagerUrl: appSetting.ModelFactoryManagerUrl,
			mfAPIUrl:     appSetting.ModelFactoryAPIUrl,
		}
	})

	return mfAccess
}

// GetVector 根据输入字符串数组获取对应的向量数组
// 参数：
//   - ctx: 上下文对象
//   - texts: 输入字符串数组
//
// 返回：
//   - [][]float32: 等长的向量数组，每个向量对应一个输入字符串
//   - error: 错误信息
func (mfa *modelFactoryAccess) GetVector(ctx context.Context, model *interfaces.SmallModel,
	words []string) ([]cond.VectorResp, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "GetVector", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	if model == nil {
		return []cond.VectorResp{}, fmt.Errorf("model is nil")
	}
	if len(words) == 0 {
		return []cond.VectorResp{}, nil
	}

	// 构建请求URL
	httpUrl := fmt.Sprintf("%s/small-model/embeddings", mfa.mfAPIUrl)

	// 设置请求头
	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		"Content-Type":                      "application/json",
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	modelID := model.ModelID
	maxTokens := model.MaxTokens
	batchSize := model.BatchSize

	allVectorResps := make([]cond.VectorResp, 0, len(words))
	for i := 0; i < len(words); i += batchSize {
		end := i + batchSize
		if end > len(words) {
			end = len(words)
		}
		currentWords := words[i:end]
		for j := 0; j < len(currentWords); j++ {
			// 计算utf8字符长度
			runes := []rune(currentWords[j])
			if len(runes) > maxTokens {
				currentWords[j] = string(runes[:maxTokens])
			}
		}

		// 构建请求体
		requestBody := map[string]interface{}{
			"model":    "",
			"model_id": modelID,
			"input":    currentWords,
		}

		// 发送POST请求获取向量
		respCode, result, err := mfa.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, requestBody)
		logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, result, err)

		if err != nil {
			logger.Errorf("Get vector request failed: %v", err)
			return nil, fmt.Errorf("get vector request failed: %w", err)
		}

		if respCode != 200 {
			logger.Errorf("Get vector request failed with status code: %d, %s", respCode, result)
			return nil, fmt.Errorf("get vector request failed with status code: %d, %s", respCode, result)
		}

		// 解析响应数据
		var response struct {
			Data []cond.VectorResp `json:"data"`
		}

		if err := sonic.Unmarshal(result, &response); err != nil {
			logger.Errorf("Unmarshal vector response failed: %v", err)
			return nil, fmt.Errorf("unmarshal vector response failed: %w", err)
		}

		// 检查返回的向量数量是否与输入文本数量一致
		if len(response.Data) != len(currentWords) {
			logger.Errorf("Vector count mismatch: expected %d, got %d", len(currentWords), len(response.Data))
			return nil, fmt.Errorf("vector count mismatch: expected %d, got %d", len(currentWords), len(response.Data))
		}

		allVectorResps = append(allVectorResps, response.Data...)
	}

	return allVectorResps, nil
}

func (mfa *modelFactoryAccess) GetModelByID(ctx context.Context, modelID string) (*interfaces.SmallModel, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "GetModelByID",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	// 构建请求URL
	httpUrl := fmt.Sprintf("%s/small-model/get?model_id=%s", mfa.mfManagerUrl, modelID)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	// 设置请求头
	headers := map[string]string{
		"Content-Type":                      "application/json",
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	// 发送GET请求获取模型
	respCode, result, err := mfa.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)
	logger.Debugf("get [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, result, err)

	if err != nil {
		logger.Errorf("Get model request failed: %v", err)
		return nil, fmt.Errorf("get model request failed: %w", err)
	}

	if respCode == http.StatusNotFound {
		logger.Warnf("Get model request failed with status code: %d, %s", respCode, result)
		return nil, nil
	}
	if respCode != http.StatusOK {
		logger.Errorf("Get model request failed with status code: %d, %s", respCode, result)
		return nil, fmt.Errorf("get model request failed with status code: %d, %s", respCode, result)
	}

	// 解析响应数据
	smallModel := interfaces.SmallModel{}
	if err := sonic.Unmarshal(result, &smallModel); err != nil {
		logger.Errorf("Unmarshal model response failed: %v", err)
		return nil, fmt.Errorf("unmarshal model response failed: %w", err)
	}

	return &smallModel, nil
}
