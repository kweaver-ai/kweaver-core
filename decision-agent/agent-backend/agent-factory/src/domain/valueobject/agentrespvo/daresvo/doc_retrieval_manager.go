package daresvo

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/chat_enum/chatresenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/pkg/errors"
)

// DocRetrievalManager 文档召回结果管理器
type DocRetrievalManager struct {
	strategies []DocRetrievalResultStrategy
}

func NewDocRetrievalManager() *DocRetrievalManager {
	manager := &DocRetrievalManager{
		strategies: make([]DocRetrievalResultStrategy, 0),
	}

	// 注册默认策略（按优先级排序）
	manager.RegisterStrategy(NewStandardDocRetrievalStrategy())

	return manager
}

func (m *DocRetrievalManager) RegisterStrategy(strategy DocRetrievalResultStrategy) {
	m.strategies = append(m.strategies, strategy)
}

func (m *DocRetrievalManager) ProcessResult(answer interface{}, strategyName chatresenum.DocRetrievalStrategy) (agentrespvo.DocRetrievalAnswer, error) {
	for _, strategy := range m.strategies {
		if strategy.GetStrategyName() == strategyName {
			res, err := strategy.Process(answer)
			if err != nil {
				return agentrespvo.DocRetrievalAnswer{}, err
			}

			return res, nil
		}
	}

	return agentrespvo.DocRetrievalAnswer{}, errors.Errorf("未找到合适的处理策略，strategyName: %s", strategyName)
}
