package cdaenum

import "github.com/pkg/errors"

// HistoryStrategy 历史对话策略
type HistoryStrategy string

const (
	// HistoryStrategyNone 无历史对话数据策略
	HistoryStrategyNone HistoryStrategy = "none"

	// HistoryStrategyCount 按历史对话数据数量策略
	HistoryStrategyCount HistoryStrategy = "count"

	// HistoryStrategyTimeWindow 按历史对话数据的时间窗口策略（预留）
	HistoryStrategyTimeWindow HistoryStrategy = "time_window"

	// HistoryStrategyToken 按历史对话数据占用token策略（预留）
	HistoryStrategyToken HistoryStrategy = "token"
)

func (h HistoryStrategy) EnumCheck() (err error) {
	switch h {
	case HistoryStrategyNone, HistoryStrategyCount, HistoryStrategyTimeWindow, HistoryStrategyToken:
		return
	default:
		err = errors.New("[HistoryStrategy]: invalid history strategy, must be one of: none, count, time_window, token")
		return
	}
}

func (h HistoryStrategy) IsNone() bool {
	return h == HistoryStrategyNone
}

func (h HistoryStrategy) IsCount() bool {
	return h == HistoryStrategyCount
}

func (h HistoryStrategy) IsTimeWindow() bool {
	return h == HistoryStrategyTimeWindow
}

func (h HistoryStrategy) IsToken() bool {
	return h == HistoryStrategyToken
}

func (h HistoryStrategy) IsEnabled() bool {
	return h != HistoryStrategyNone
}
