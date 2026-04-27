// Package interfaces 定义接口
// @file logics_aigeneration.go
// @description: AI生成相关接口
package interfaces

import (
	"context"
	"fmt"
)

// PromptTemplate 提示词模板结构
type PromptTemplate struct {
	PromptID           string `json:"prompt_id"`
	Name               string `json:"name" validate:"required"`
	Description        string `json:"description" validate:"required"`
	SystemPrompt       string `json:"system_prompt" validate:"required"`
	UserPromptTemplate string `json:"user_prompt_template" validate:"required"`
}

// FormatUserPrompt 格式化用户提示词
func (p *PromptTemplate) FormatUserPrompt(args ...interface{}) string {
	return fmt.Sprintf(p.UserPromptTemplate, args...)
}

// PromptTemplateType 提示词模板类型
type PromptTemplateType string

const (
	PythonFunctionGenerator PromptTemplateType = "python_function_generator" // Python函数生成Prompt模板
	MetadataParamGenerator  PromptTemplateType = "metadata_param_generator"  // 元数据参数生成Prompt模板
)

// FunctionAIGenerateReq 函数AI生成请求
type FunctionAIGenerateReq struct {
	Type    PromptTemplateType `uri:"type" validate:"required,oneof=python_function_generator metadata_param_generator"` // 提示词模板类型，必填
	Query   string             `json:"query"`                                                                            // 用户输入，必填
	Inputs  []ParameterDef     `json:"inputs,omitempty" form:"inputs"`                                                   // 输入参数列表
	Outputs []ParameterDef     `json:"outputs,omitempty" form:"outputs"`
	Code    string             `json:"code,omitempty" form:"code"`     // 输出参数列表
	Stream  bool               `json:"stream,omitempty" form:"stream"` // 是否流式返回
}

// Validate 校验请求参数
func (f *FunctionAIGenerateReq) Validate() error {
	switch f.Type {
	case PythonFunctionGenerator:
		if f.Query == "" {
			return fmt.Errorf("query is empty, please input a valid query")
		}
	case MetadataParamGenerator:
		if f.Code == "" {
			return fmt.Errorf("code is empty, please input a valid code")
		}
	default:
		return fmt.Errorf("template type %s is not supported, only support python_function_generator, metadata_param_generator, metadata_test_data_generator", f.Type)
	}
	return nil
}

// FunctionAIGeneratResp 函数智能生成响应
type FunctionAIGeneratResp struct {
	Content any `json:"content"` // 生成内容
}

// AIGeneratMetadataContent AI生成元数据内容
type AIGeneratMetadataContent struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	UseRule     string         `json:"use_rule"`
	Inputs      []ParameterDef `json:"inputs"`
	Outputs     []ParameterDef `json:"outputs"`
}

// AIGenerationService AI辅助生成接口
type AIGenerationService interface {
	// FunctionAIGenerate 函数智能生成
	FunctionAIGenerate(ctx context.Context, req *FunctionAIGenerateReq) (*FunctionAIGeneratResp, error)
	// FunctionAIGenerateStream 函数智能生成流式返回
	FunctionAIGenerateStream(ctx context.Context, req *FunctionAIGenerateReq) (respChan chan string, errChan chan error, err error)
	// GetPromptTemplate 获取指定类型的提示词模板
	GetPromptTemplate(ctx context.Context, tempType PromptTemplateType) (*PromptTemplate, error)
}

// GetPromptTemplateReq 获取提示词模板请求
type GetPromptTemplateReq struct {
	Type PromptTemplateType `uri:"type" validate:"required,oneof=python_function_generator metadata_param_generator"` // 提示词模板类型，必填
}
