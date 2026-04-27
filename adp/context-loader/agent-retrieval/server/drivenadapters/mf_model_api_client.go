// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/utils"
)

// API路径常量
const (
	chatCompletionsURI = "/v1/chat/completions"
	rerankURI          = "/v1/small-model/reranker"
)

// mfModelAPIClient MF-Model API统一客户端
// 提供LLM对话和向量重排序两类能力，统一使用 mf-model-api 服务
type mfModelAPIClient struct {
	logger     interfaces.Logger
	baseURL    string
	httpClient interfaces.HTTPClient
}

var (
	mfModelAPIClientOnce sync.Once
	mfModelAPIClientInst *mfModelAPIClient
)

// NewMFModelAPIClient 创建MF-Model API统一客户端单例
// 实现 DrivenMFModelAPIClient 接口
func NewMFModelAPIClient() *mfModelAPIClient {
	mfModelAPIClientOnce.Do(func() {
		conf := config.NewConfigLoader()
		mfModelAPIClientInst = &mfModelAPIClient{
			logger:     conf.GetLogger(),
			baseURL:    conf.MFModelAPI.BuildURL("/api/private/mf-model-api"),
			httpClient: rest.NewHTTPClient(),
		}
	})
	return mfModelAPIClientInst
}

// ============================================================
// DrivenLLMClient 接口实现
// ============================================================

// chatCompletionsResp Chat接口响应结构
type chatCompletionsResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Chat 非流式对话，返回完整响应内容
func (c *mfModelAPIClient) Chat(ctx context.Context, req *interfaces.LLMChatReq) (string, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, chatCompletionsURI)

	// 构建请求体
	reqBody := map[string]interface{}{
		"model":             req.Model,
		"messages":          req.Messages,
		"stream":            false, // 非流式
		"temperature":       req.Temperature,
		"top_k":             req.TopK,
		"top_p":             req.TopP,
		"frequency_penalty": req.FrequencyPenalty,
		"presence_penalty":  req.PresencePenalty,
		"max_tokens":        req.MaxTokens,
	}

	// 获取Header（统一方式）
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON

	c.logger.WithContext(ctx).Debugf("[MFModelAPIClient#Chat] URL: %s", url)

	// 调用HTTP客户端
	respCode, respBody, err := c.httpClient.Post(ctx, url, header, reqBody)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("[MFModelAPIClient#Chat] Request failed: %v", err)
		return "", fmt.Errorf("request failed: %w", err)
	}

	if respCode != http.StatusOK {
		c.logger.WithContext(ctx).Errorf("[MFModelAPIClient#Chat] Request failed with code %d", respCode)
		return "", infraErr.DefaultHTTPError(ctx, respCode, fmt.Sprintf("chat request failed with code %d", respCode))
	}

	// 解析响应
	var resp chatCompletionsResp
	resultBytes := utils.ObjectToByte(respBody)
	if err := json.Unmarshal(resultBytes, &resp); err != nil {
		c.logger.WithContext(ctx).Errorf("[MFModelAPIClient#Chat] Unmarshal failed: %v", err)
		return "", fmt.Errorf("unmarshal response failed: %w", err)
	}

	// 提取content
	if len(resp.Choices) > 0 && resp.Choices[0].Message.Content != "" {
		content := resp.Choices[0].Message.Content
		c.logger.WithContext(ctx).Debugf("[MFModelAPIClient#Chat] Response length: %d", len(content))
		return content, nil
	}

	return "", fmt.Errorf("unexpected response format: no content found")
}

// ============================================================
// DrivenRerankClient 接口实现
// ============================================================

// Rerank 对文档进行重排序
func (c *mfModelAPIClient) Rerank(ctx context.Context, query string, documents []string) (*interfaces.RerankResp, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, rerankURI)

	// 构建请求体
	reqBody := map[string]interface{}{
		"query":     query,
		"documents": documents,
		"model":     "reranker",
	}

	// 获取Header（统一方式）
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON

	c.logger.WithContext(ctx).Debugf("[MFModelAPIClient#Rerank] URL: %s, query: %s, docs count: %d",
		url, query, len(documents))

	// 调用HTTP客户端
	_, respBody, err := c.httpClient.Post(ctx, url, header, reqBody)
	if err != nil {
		c.logger.WithContext(ctx).Errorf("[MFModelAPIClient#Rerank] Request failed: %v", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// 解析响应
	var result interfaces.RerankResp
	resultBytes := utils.ObjectToByte(respBody)
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		c.logger.WithContext(ctx).Errorf("[MFModelAPIClient#Rerank] Unmarshal failed: %v", err)
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	c.logger.WithContext(ctx).Debugf("[MFModelAPIClient#Rerank] Results count: %d", len(result.Results))

	return &result, nil
}
