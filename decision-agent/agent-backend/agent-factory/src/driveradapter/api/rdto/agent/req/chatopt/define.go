package chatopt

import (
	"fmt"
)

type ChatOption struct {
	EnableDependencyCache bool `json:"enable_dependency_cache"` // 是否需要使用缓存

	IsNeedHistory bool `json:"is_need_history"` // 是否需要历史上下文

	IsNeedDocRetrivalPostProcess bool `json:"is_need_doc_retrival_post_process"` // 是否需要文档检索后处理

	IsNeedProgress bool `json:"is_need_progress"` // 是否需要progress
}

func (o *ChatOption) Check(isDebugMode bool) (err error) {
	// 1. 如果is_debug_mode为true，则enable_dependency_cache必须为false
	if isDebugMode && o.EnableDependencyCache {
		return fmt.Errorf("enable_dependency_cache must be false when is_debug_mode is true")
	}

	// // 2. check chat_scenario_type
	// if err = o.ChatScenarioType.EnumCheck(); err != nil {
	// 	return
	// }

	// // 3. check chat_scenario
	// if (o.ChatScenarioType == chat_enum.ChatScenarioCustom || o.ChatScenarioType == chat_enum.ChatScenarioThirdSystem) && o.ChatScenario == "" {
	// 	return fmt.Errorf("chat_scenario must be set when chat_scenario_type is custom or third_system")
	// }

	return
}
