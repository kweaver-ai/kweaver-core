package observabilitysvc

import (
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
)

// 辅助函数用于安全的类型断言
func safeGetString(m map[string]any, key string) string {
	if val, exists := m[key]; exists && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}

	return ""
}

func safeGetBool(m map[string]any, key string) bool {
	if val, exists := m[key]; exists && val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	return false
}

func safeGetFloat64(m map[string]any, key string) float64 {
	if val, exists := m[key]; exists && val != nil {
		if f, ok := val.(float64); ok {
			return f
		}
	}

	return 0.0
}

func safeGetInt64(m map[string]any, key string) int64 {
	if val, exists := m[key]; exists && val != nil {
		if f, ok := val.(float64); ok {
			return int64(f)
		}
	}

	return 0
}

func safeParseSkillInfo(data any) *agentrespvo.SkillInfo {
	if data == nil {
		return nil
	}

	if skillMap, ok := data.(map[string]any); ok {
		return &agentrespvo.SkillInfo{
			Type:    safeGetString(skillMap, "type"),
			Name:    safeGetString(skillMap, "name"),
			Args:    safeParseArgs(skillMap["args"]),
			Checked: safeGetBool(skillMap, "checked"),
		}
	}

	return nil
}

func safeParseArgs(data any) []agentrespvo.Arg {
	if data == nil {
		return nil
	}

	if argsArray, ok := data.([]any); ok {
		var args []agentrespvo.Arg

		for _, item := range argsArray {
			if argMap, ok := item.(map[string]any); ok {
				args = append(args, agentrespvo.Arg{
					Name:  safeGetString(argMap, "name"),
					Value: argMap["value"],
					Type:  safeGetString(argMap, "type"),
				})
			}
		}

		return args
	}

	return nil
}

func safeParseTokenUsage(data any) agentrespvo.TokenUsage {
	if data == nil {
		return agentrespvo.TokenUsage{}
	}

	if usageMap, ok := data.(map[string]any); ok {
		return agentrespvo.TokenUsage{
			PromptTokens:       safeGetInt64(usageMap, "prompt_tokens"),
			CompletionTokens:   safeGetInt64(usageMap, "completion_tokens"),
			TotalTokens:        safeGetInt64(usageMap, "total_tokens"),
			PromptTokenDetails: safeParsePromptTokenDetails(usageMap["prompt_tokens_details"]),
		}
	}

	return agentrespvo.TokenUsage{}
}

func safeParsePromptTokenDetails(data any) agentrespvo.PromptTokenDetails {
	if data == nil {
		return agentrespvo.PromptTokenDetails{}
	}

	if detailsMap, ok := data.(map[string]any); ok {
		return agentrespvo.PromptTokenDetails{
			CachedTokens:   safeGetInt64(detailsMap, "cached_tokens"),
			UncachedTokens: safeGetInt64(detailsMap, "uncached_tokens"),
		}
	}

	return agentrespvo.PromptTokenDetails{}
}

// formatTimeToISO8601 将毫秒时间戳转换为ISO 8601格式字符串
func formatTimeToISO8601(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	// 将毫秒转换为秒
	seconds := timestamp / 1000
	// 将毫秒转换为纳秒
	nanoseconds := (timestamp % 1000) * 1000000

	timeObj := time.Unix(seconds, nanoseconds)

	return timeObj.Format(time.RFC3339)
}
