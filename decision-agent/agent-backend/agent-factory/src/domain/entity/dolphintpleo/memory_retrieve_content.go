package dolphintpleo

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
)

type MemoryRetrieveContent struct {
	// Queries        string `json:"queries"`
	RelevantMemory string `json:"relevant_memory"`
	Other          string `json:"other"`
	IsEnable       bool   `json:"is_enable"`
}

func NewMemoryRetrieveContent() *MemoryRetrieveContent {
	return &MemoryRetrieveContent{}
}

func (m *MemoryRetrieveContent) LoadFromConfig(config *daconfvalobj.Config) {
	if config.MemoryCfg != nil && config.MemoryCfg.IsEnabled {
		// m.Queries = config.Input.Fields.GenNotFileDolphinStr()
		m.RelevantMemory = `
@search_memory(query=$query, user_id=$header['x-account-id'], limit=50, threshold=0.5) -> relevant_memories
json.dumps($relevant_memories["answer"]["result"]) -> memory_str
$_history + [{"role": "system", "content": "Relevant memories: " + $memory_str}] -> _history
`
		m.Other += `
"" -> relevant_memories
"" -> memory_str
`

		m.IsEnable = true
	}
}

func (m *MemoryRetrieveContent) ToString() (str string) {
	// if m.Queries != "" {
	// 	str += m.Queries
	// }
	if m.RelevantMemory != "" {
		str += m.RelevantMemory
	}

	if m.Other != "" {
		str += m.Other
	}

	return
}

func (m *MemoryRetrieveContent) ToDolphinTplEo() *DolphinTplEo {
	key := cdaenum.DolphinTplKeyMemoryRetrieve

	return &DolphinTplEo{
		Key:   key,
		Name:  key.GetName(),
		Value: m.ToString(),
	}
}
