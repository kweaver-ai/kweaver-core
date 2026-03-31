package aigeneration

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

// prompt加载及解析

//go:embed templates/*.md
var promptTemplatesFS embed.FS

// PromptLoader 提示词加载器
type PromptLoader struct {
	DefaulTemplates    map[interfaces.PromptTemplateType]*interfaces.PromptTemplate
	MFModelManager     interfaces.MFModelManager
	AIGenerationConfig config.AIGenerationConfig
	Logger             interfaces.Logger
}

var (
	pOnce        sync.Once
	promptLoader *PromptLoader
)

// NewPromptLoader 创建新的提示词加载器
func NewPromptLoader() (*PromptLoader, error) {
	pOnce.Do(func() {
		conf := config.NewConfigLoader()
		promptLoader = &PromptLoader{
			Logger:             conf.GetLogger(),
			MFModelManager:     drivenadapters.NewMFModelManager(),
			AIGenerationConfig: conf.AIGenerationConfig,
			DefaulTemplates: map[interfaces.PromptTemplateType]*interfaces.PromptTemplate{
				interfaces.PythonFunctionGenerator: {
					Name:               string(interfaces.PythonFunctionGenerator),
					Description:        "Python函数生成Prompt模板",
					SystemPrompt:       "",
					UserPromptTemplate: "函数内容描述:%s; inputs:%v; outputs:%v;",
				},
				interfaces.MetadataParamGenerator: {
					Name:               string(interfaces.MetadataParamGenerator),
					Description:        "元数据参数生成Prompt模板",
					SystemPrompt:       "",
					UserPromptTemplate: `{"code": "%s", "inputs_json": %v, "outputs_json": %v}`,
				},
			},
		}
		// 加载默认模板文件
		if err := promptLoader.loadDefaulTemplates(); err != nil {
			panic(fmt.Errorf("failed to load prompt templates: %w", err))
		}
	})
	return promptLoader, nil
}

// loadDefaulTemplates 加载默认模板文件
func (l *PromptLoader) loadDefaulTemplates() error {
	return fs.WalkDir(promptTemplatesFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		tempName := strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(filepath.Base(path))))
		tempType := interfaces.PromptTemplateType(tempName)
		if _, ok := l.DefaulTemplates[tempType]; !ok {
			return fmt.Errorf("default prompt template %s not found", tempType)
		}

		content, err := fs.ReadFile(promptTemplatesFS, path)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", path, err)
		}
		l.DefaulTemplates[tempType].SystemPrompt = string(content)
		return nil
	})
}

// GetTemplate 获取指定类型的提示词模板
func (l *PromptLoader) GetTemplate(ctx context.Context, tempType interfaces.PromptTemplateType) (*interfaces.PromptTemplate, error) {
	temp, ok := l.DefaulTemplates[tempType]
	if !ok {
		return nil, fmt.Errorf("default prompt template %s not found", tempType)
	}
	customPrompt, err := l.loadCustomPromptTemplate(ctx, tempType)
	if err != nil {
		l.Logger.WithContext(ctx).Warnf("failed to load custom prompt: %v", err)
	}
	if customPrompt == nil || customPrompt.SystemPrompt == "" {
		// 未配置自定义提示词，使用默认提示词
		return temp, nil
	}
	// 配置了自定义提示词，使用自定义提示词
	customPrompt.UserPromptTemplate = temp.UserPromptTemplate
	customPrompt.Description = temp.Description
	return customPrompt, nil
}

// loadCustomPrompt 加载自定义配置提示词
func (l *PromptLoader) loadCustomPromptTemplate(ctx context.Context, tempType interfaces.PromptTemplateType) (*interfaces.PromptTemplate, error) {
	var promptID string
	switch tempType {
	case interfaces.PythonFunctionGenerator:
		promptID = l.AIGenerationConfig.PythonFunctionGeneratorPromptID
	case interfaces.MetadataParamGenerator:
		promptID = l.AIGenerationConfig.MetadataParamGeneratorPromptID
	}
	if promptID == "" {
		return nil, nil
	}
	// 从模型管理器获取模型配置
	promptResult, err := l.MFModelManager.GetPromptByPromptID(ctx, promptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model config: %v", err)
	}
	l.Logger.WithContext(ctx).Debugf("model manager get prompt result: %v", promptResult)
	return &interfaces.PromptTemplate{
		PromptID:     promptResult.PromptID,
		Name:         promptResult.PromptName,
		SystemPrompt: promptResult.Messages,
	}, nil
}
