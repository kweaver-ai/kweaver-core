package common

import "sync"

// TimeoutConfig 超时配置结构体
type TimeoutConfig struct {
	// 操作类型到超时秒数的映射
	OperationTimeouts map[string]int
	// 默认超时秒数
	DefaultTimeout int
}

var timeoutConfig *TimeoutConfig
var toOnce sync.Once

// getDefaultOperationTimeouts 获取默认的操作超时映射
func getDefaultOperationTimeouts() map[string]int {
	return map[string]int{
		WorkflowApproval:                  1 * 10e9,     // 审核节点，永不超时（约31.7年）
		AnyshareFileOCROpt:                1 * 10e9,     // OCR节点，永不超时
		AnyshareEleInvoiceOpt:             1 * 10e9,     // 电子发票节点，永不超时
		AnyshareIDCardOpt:                 1 * 10e9,     // 身份证节点，永不超时
		AudioTransfer:                     1 * 10e9,     // 音频转文字节点，永不超时
		IntelliinfoTranfer:                1 * 10e9,     // 图谱转换，永不超时
		OpCognitiveAssistantDocSummarize:  30 * 60,      // 大模型文档总结，30分钟
		OpCognitiveAssistantMeetSummarize: 30 * 60,      // 大模型会议总结，30分钟
		OpOpenSearchBulkUpsert:            30 * 60,      // OpenSearch批量插入，30分钟
		OpLLMChatCompletion:               2 * 60 * 60,  // LLM聊天，2小时
		InternalToolPy3Opt:                24 * 60 * 60, // Python节点，24小时
		OpContentEntity:                   24 * 60 * 60, // 内容实体，24小时
		OpEcoconfigReindex:                24 * 60 * 60, // 重新索引，24小时
		OpAnyDataCallAgent:                2 * 60 * 60,  // Agent节点，2小时
	}
}

// NewTimeoutConfig 创建指定默认超时的超时配置
func NewTimeoutConfig() *TimeoutConfig {
	toOnce.Do(func() {
		timeoutConfig = &TimeoutConfig{
			OperationTimeouts: getDefaultOperationTimeouts(),
			DefaultTimeout:    NewConfig().Server.ExecutorTimeout,
		}
	})

	return timeoutConfig
}

// GetTimeout 根据操作类型获取超时时间（秒）
// 支持前缀匹配（如 combo_operator_* 和 trigger_operator_*）
func (tc *TimeoutConfig) GetTimeout(operator string) int {
	// 检查前缀匹配
	for op, timeout := range tc.OperationTimeouts {
		if operator == op {
			return timeout
		}
	}
	return tc.DefaultTimeout
}

// SetTimeout 设置特定操作的超时时间
func (tc *TimeoutConfig) SetTimeout(operator string, timeoutSeconds int) {
	if tc.OperationTimeouts == nil {
		tc.OperationTimeouts = make(map[string]int)
	}
	tc.OperationTimeouts[operator] = timeoutSeconds
}
