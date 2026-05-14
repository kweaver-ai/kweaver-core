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
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"
	cond "bkn-backend/common/condition"
	"bkn-backend/interfaces"
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

func (mfa *modelFactoryAccess) GetDefaultModel(ctx context.Context) (*interfaces.SmallModel, error) {
	// 不缓存，直接get
	if mfa.appSetting.ServerSetting.DefaultSmallModelEnabled {
		defaultModelName := mfa.appSetting.ServerSetting.DefaultSmallModelName
		smallModel, err := mfa.GetModelByName(ctx, defaultModelName)
		if err != nil {
			logger.Errorf("Get default model by name[%s] failed: %v", defaultModelName, err)
			return nil, fmt.Errorf("get default model by name[%s] failed: %w", defaultModelName, err)
		}
		return smallModel, nil
	} else {
		return nil, nil
	}
}

func (mfa *modelFactoryAccess) GetModelByID(ctx context.Context, modelID string) (*interfaces.SmallModel, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetModelByID")
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
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get model failed")
		otellog.LogError(ctx, "Get model request failed", err)
		return nil, fmt.Errorf("get model request failed: %w", err)
	}

	if respCode == http.StatusNotFound {
		logger.Warnf("Get model request failed with status code: %d, %s", respCode, result)
		oteltrace.AddHttpAttrs4Ok(span, respCode)
		return nil, nil
	}
	if respCode != http.StatusOK {
		err := fmt.Errorf("get model request failed with status code: %d, %s", respCode, result)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		otellog.LogError(ctx, "Get model request failed", err)
		return nil, err
	}

	// 解析响应数据
	smallModel := interfaces.SmallModel{}
	if err := sonic.Unmarshal(result, &smallModel); err != nil {
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal model response failed")
		otellog.LogError(ctx, "Unmarshal model response failed", err)
		return nil, fmt.Errorf("unmarshal model response failed: %w", err)
	}

	oteltrace.AddHttpAttrs4Ok(span, respCode)
	return &smallModel, nil
}

func (mfa *modelFactoryAccess) GetModelByName(ctx context.Context, modelName string) (*interfaces.SmallModel, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetModelByName")
	defer span.End()

	// 构建请求URL
	httpUrl := fmt.Sprintf("%s/small-model/get_by_name?model_name=%s", mfa.mfManagerUrl, modelName)

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
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get model by name failed")
		otellog.LogError(ctx, "Get model request failed", err)
		return nil, fmt.Errorf("get model request failed: %w", err)
	}

	if respCode == http.StatusNotFound {
		logger.Warnf("Get model request failed with status code: %d, %s", respCode, result)
		oteltrace.AddHttpAttrs4Ok(span, respCode)
		return nil, nil
	}
	if respCode != http.StatusOK {
		err := fmt.Errorf("get model request failed with status code: %d, %s", respCode, result)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		otellog.LogError(ctx, "Get model request failed", err)
		return nil, err
	}

	// 解析响应数据
	smallModel := interfaces.SmallModel{}
	if err := sonic.Unmarshal(result, &smallModel); err != nil {
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal model response failed")
		otellog.LogError(ctx, "Unmarshal model response failed", err)
		return nil, fmt.Errorf("unmarshal model response failed: %w", err)
	}

	oteltrace.AddHttpAttrs4Ok(span, respCode)
	return &smallModel, nil
}

func (mfa *modelFactoryAccess) GetVector(ctx context.Context,
	model *interfaces.SmallModel, words []string) ([]*cond.VectorResp, error) {

	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetVector")
	defer span.End()

	if model == nil {
		span.SetStatus(codes.Error, "Model is nil")
		return []*cond.VectorResp{}, fmt.Errorf("model is nil")
	}
	if len(words) == 0 {
		span.SetStatus(codes.Ok, "")
		return []*cond.VectorResp{}, nil
	}

	// 构建请求URL
	httpUrl := fmt.Sprintf("%s/small-model/embeddings", mfa.mfAPIUrl)

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

	modelID := model.ModelID
	maxTokens := model.MaxTokens
	batchSize := model.BatchSize

	allVectorResps := make([]*cond.VectorResp, 0, len(words))
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
		requestBody := map[string]any{
			"model":    "",
			"model_id": modelID,
			"input":    currentWords,
		}

		// 发送POST请求获取向量
		respCode, result, err := mfa.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, requestBody)

		// 打印日志
		modelInfo, _ := sonic.Marshal(model)
		logger.Debugf("post [%s] finished, small model info: [%s], response code is [%d], error is [%v]",
			httpUrl, modelInfo, respCode, err)

		if err != nil {
			oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get vector failed")
			otellog.LogError(ctx, "Get vector request failed", err)
			return nil, fmt.Errorf("get vector request failed: %w", err)
		}

		if respCode != 200 {
			err := fmt.Errorf("get vector request failed with status code: %d, %s", respCode, result)
			oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
			otellog.LogError(ctx, "Get vector request failed", err)
			return nil, err
		}

		// 解析响应数据
		var response struct {
			Data []*cond.VectorResp `json:"data"`
		}

		if err := sonic.Unmarshal(result, &response); err != nil {
			oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal vector response failed")
			otellog.LogError(ctx, "Unmarshal vector response failed", err)
			return nil, fmt.Errorf("unmarshal vector response failed: %w", err)
		}
		logger.Debugf("vectorized result length is [%d]", len(response.Data))

		// 检查返回的向量数量是否与输入文本数量一致
		if len(response.Data) != len(currentWords) {
			err := fmt.Errorf("vector count mismatch: expected %d, got %d", len(currentWords), len(response.Data))
			span.SetStatus(codes.Error, "Vector count mismatch")
			otellog.LogError(ctx, "Vector count mismatch", err)
			return nil, err
		}

		allVectorResps = append(allVectorResps, response.Data...)
	}

	span.SetStatus(codes.Ok, "")
	return allVectorResps, nil
}
