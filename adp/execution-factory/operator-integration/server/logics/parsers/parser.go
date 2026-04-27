package parsers

import (
	"context"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

// Parser 解析器接口
type Parser interface {
	// Type 返回解析器处理的元数据类型
	Type() interfaces.MetadataType
	// Parse 解析原始输入为元数据
	Parse(ctx context.Context, input any) ([]interfaces.IMetadataDB, error)
	// GetAllContent 获取所有内容
	GetAllContent(ctx context.Context, input any) (any, error)
}

// Registry 解析器注册中心
type Registry struct {
	mu      sync.RWMutex
	parsers map[interfaces.MetadataType]Parser
	Logger  interfaces.Logger
}

var (
	mrSync sync.Once
	mr     *Registry
)

// NewRegistry 创建解析器注册中心
func NewRegistry() *Registry {
	mrSync.Do(func() {
		conf := config.NewConfigLoader()
		mr = &Registry{
			Logger:  conf.GetLogger(),
			parsers: make(map[interfaces.MetadataType]Parser),
		}
		err := mr.Register(&openAPIParser{
			Logger: conf.GetLogger(),
		})
		if err != nil {
			panic(err)
		}
		err = mr.Register(&pythonFunctionParser{
			Logger:    conf.GetLogger(),
			Validator: validator.NewValidator(),
		})
		if err != nil {
			panic(err)
		}
	})
	return mr
}

// Register 注册解析器
func (r *Registry) Register(parser Parser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	metaType := parser.Type()
	if _, exists := r.parsers[metaType]; exists {
		return fmt.Errorf("parser for type %s already registered", metaType)
	}

	r.parsers[metaType] = parser
	return nil
}

// Get 获取解析器
func (r *Registry) Get(metaType interfaces.MetadataType) (Parser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	parser, exists := r.parsers[metaType]
	if !exists {
		return nil, fmt.Errorf("parser for type %s not found", metaType)
	}

	return parser, nil
}

// MustGet 获取解析器（不存在时 panic）
func (r *Registry) MustGet(metaType interfaces.MetadataType) Parser {
	parser, err := r.Get(metaType)
	if err != nil {
		panic(err)
	}
	return parser
}

// List 列出所有已注册的解析器类型
func (r *Registry) List() []interfaces.MetadataType {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]interfaces.MetadataType, 0, len(r.parsers))
	for t := range r.parsers {
		result = append(result, t)
	}
	return result
}
