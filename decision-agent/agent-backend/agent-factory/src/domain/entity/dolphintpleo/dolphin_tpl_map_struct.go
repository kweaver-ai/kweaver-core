package dolphintpleo

import (
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum/builtinagentenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
)

type DolphinTplMapStruct struct {
	MemoryRetrieve  *MemoryRetrieveContent  `json:"memory_retrieve"`
	DocRetrieve     *DocRetrieveContent     `json:"doc_retrieve"`
	GraphRetrieve   *GraphRetrieveContent   `json:"graph_retrieve"`
	ContextOrganize *ContextOrganizeContent `json:"context_organize"`

	RelatedQuestions *RelatedQuestionsContent `json:"related_questions"`
}

func NewDolphinTplMapStruct() *DolphinTplMapStruct {
	memoryRetrieve := NewMemoryRetrieveContent()
	docRetrieve := NewDocRetrieveContent()
	graphRetrieve := NewGraphRetrieveContent()

	otherTplStruct := &OtherTplStruct{
		MemoryRetrieve: memoryRetrieve,
		DocRetrieve:    docRetrieve,
		GraphRetrieve:  graphRetrieve,
	}

	contextOrganize := NewContextOrganizeContent(otherTplStruct)

	relatedQuestions := NewRelatedQuestionsContent()

	return &DolphinTplMapStruct{
		MemoryRetrieve:   memoryRetrieve,
		DocRetrieve:      docRetrieve,
		GraphRetrieve:    graphRetrieve,
		ContextOrganize:  contextOrganize,
		RelatedQuestions: relatedQuestions,
	}
}

// builtInAgentKey 为内置agent的key，用于内置agent的一些特殊处理
func (s *DolphinTplMapStruct) LoadFromConfig(config *daconfvalobj.Config, builtInAgentKey builtinagentenum.AgentKey, isNeedHandleBuiltinAgent bool) {
	s.MemoryRetrieve.LoadFromConfig(config)

	isBuiltInDocQAAgent := builtInAgentKey.IsDocQA() && isNeedHandleBuiltinAgent
	s.DocRetrieve.LoadFromConfig(config, isBuiltInDocQAAgent)

	isBuiltInGraphQAAgent := builtInAgentKey.IsGraphQA() && isNeedHandleBuiltinAgent
	s.GraphRetrieve.LoadFromConfig(config, isBuiltInGraphQAAgent)

	s.ContextOrganize.LoadFromConfig(config)

	s.RelatedQuestions.LoadFromConfig(config)
}
