// Package aigeneration 智能生成服务
// @file aigeneration.go
// @description: 智能生成服务
package aigeneration

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

// aiGenerationService 智能生成服务
type aiGenerationService struct {
	Logger           interfaces.Logger
	MFModelAPIClient interfaces.MFModelAPIClient
	PromptLoader     *PromptLoader
	LLMConfig        config.LLMConfig
}

var (
	agOnce     sync.Once
	agInstance interfaces.AIGenerationService
)

// NewAIGenerationService 创建新的智能生成服务
func NewAIGenerationService() interfaces.AIGenerationService {
	agOnce.Do(func() {
		promptLoader, err := NewPromptLoader()
		if err != nil {
			logger.Errorf("failed to create prompt loader: %v", err)
			panic(err)
		}
		conf := config.NewConfigLoader()
		agInstance = &aiGenerationService{
			Logger:           conf.GetLogger(),
			MFModelAPIClient: drivenadapters.NewMFModelAPIClient(),
			LLMConfig:        conf.AIGenerationConfig.LLMConfig,
			PromptLoader:     promptLoader,
		}
	})
	return agInstance
}

func (ag *aiGenerationService) generateChatCompletionParams(ctx context.Context, req *interfaces.FunctionAIGenerateReq) (*interfaces.ChatCompletionReq, error) {
	promptTemplate, err := ag.PromptLoader.GetTemplate(ctx, req.Type)
	if err != nil {
		ag.Logger.WithContext(ctx).Errorf("failed to get prompt template: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
		return nil, err
	}
	chatCompletionReq := &interfaces.ChatCompletionReq{
		Model:            ag.LLMConfig.Model, // 大模型名称传空，使用默认大模型
		MaxTokens:        ag.LLMConfig.MaxTokens,
		Temperature:      ag.LLMConfig.Temperature,
		TopK:             ag.LLMConfig.TopK,
		TopP:             ag.LLMConfig.TopP,
		FrequencyPenalty: ag.LLMConfig.FrequencyPenalty,
		PresencePenalty:  ag.LLMConfig.PresencePenalty,
		Messages: []interfaces.ChatCompletionMessage{
			{
				Role:    "system",
				Content: promptTemplate.SystemPrompt,
			},
		},
	}
	var userPrompt string
	switch req.Type {
	case interfaces.PythonFunctionGenerator:
		// 默认函数内容描述
		if req.Query == "" {
			req.Query = "根据输入参数和输出参数，生成符合标准的Python函数代码"
		}
		userPrompt = promptTemplate.FormatUserPrompt(req.Query, req.Inputs, req.Outputs)
	case interfaces.MetadataParamGenerator:
		userPrompt = promptTemplate.FormatUserPrompt(req.Code, req.Inputs, req.Outputs)
	}
	chatCompletionReq.Messages = append(chatCompletionReq.Messages, interfaces.ChatCompletionMessage{
		Role:    "user",
		Content: userPrompt,
	})
	return chatCompletionReq, nil
}

// AIGenerate 智能生成
func (ag *aiGenerationService) FunctionAIGenerate(ctx context.Context, req *interfaces.FunctionAIGenerateReq) (resp *interfaces.FunctionAIGeneratResp, err error) {
	chatCompletionReq, err := ag.generateChatCompletionParams(ctx, req)
	if err != nil {
		return nil, err
	}
	result, err := ag.MFModelAPIClient.ChatCompletion(ctx, chatCompletionReq)
	if err != nil {
		return nil, err
	}
	var apiGenContent string
	if len(result.Choices) > 0 {
		apiGenContent = result.Choices[0].Message.Content
	}
	if apiGenContent == "" {
		err = errors.NewHTTPError(ctx, http.StatusServiceUnavailable, errors.ErrExtFunctionAIGenerateFailed, fmt.Sprintf("ai response %v", result))
		return nil, err
	}
	resp = &interfaces.FunctionAIGeneratResp{}
	switch req.Type {
	case interfaces.PythonFunctionGenerator:
		resp.Content = apiGenContent
	case interfaces.MetadataParamGenerator:
		content := &interfaces.AIGeneratMetadataContent{}
		err = utils.StringToObject(apiGenContent, content)
		if err != nil {
			err = errors.NewHTTPError(ctx, http.StatusServiceUnavailable, errors.ErrExtFunctionAIGenerateFailed, fmt.Sprintf("ai response %v format unmarshal err: %s", result, err.Error()))
			return nil, err
		}
		resp.Content = content
	}
	return resp, nil
}

// AIGenerateStream 智能生成流式返回
func (ag *aiGenerationService) FunctionAIGenerateStream(ctx context.Context, req *interfaces.FunctionAIGenerateReq) (respChan chan string, errChan chan error, err error) {
	chatCompletionReq, err := ag.generateChatCompletionParams(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	respChan, errChan, err = ag.MFModelAPIClient.StreamChatCompletion(ctx, chatCompletionReq)
	if err != nil {
		return nil, nil, err
	}
	return respChan, errChan, nil
}

// 获取当前提示词模板
func (ag *aiGenerationService) GetPromptTemplate(ctx context.Context, tempType interfaces.PromptTemplateType) (*interfaces.PromptTemplate, error) {
	return ag.PromptLoader.GetTemplate(ctx, tempType)
}
