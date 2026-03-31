package cdaenum

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/pkg/errors"
)

// DolphinTplKey dolphin模板key
type DolphinTplKey string

const (
	DolphinTplKeyMemoryRetrieve  DolphinTplKey = "memory_retrieve"
	DolphinTplKeyTempFileProcess DolphinTplKey = "temp_file_process"
	DolphinTplKeyDocRetrieve     DolphinTplKey = "doc_retrieve"
	DolphinTplKeyGraphRetrieve   DolphinTplKey = "graph_retrieve"
	DolphinTplKeyContextOrganize DolphinTplKey = "context_organize"

	DolphinTplKeyRelatedQuestions DolphinTplKey = "related_questions"
)

func (b DolphinTplKey) EnumCheck() (err error) {
	if !cutil.ExistsGeneric([]DolphinTplKey{DolphinTplKeyMemoryRetrieve, DolphinTplKeyTempFileProcess, DolphinTplKeyDocRetrieve, DolphinTplKeyGraphRetrieve, DolphinTplKeyContextOrganize, DolphinTplKeyRelatedQuestions}, b) {
		err = errors.New("[DolphinTplKey]: invalid dolphin tpl key")
		return
	}

	return
}

func (b DolphinTplKey) String() string {
	return string(b)
}

func (b DolphinTplKey) GetName() string {
	switch b {
	case DolphinTplKeyMemoryRetrieve:
		return "记忆召回模块"
	case DolphinTplKeyTempFileProcess:
		return "临时区文件处理模块"
	case DolphinTplKeyDocRetrieve:
		return "文档召回模块"
	case DolphinTplKeyGraphRetrieve:
		return "业务知识网络召回模块"
	case DolphinTplKeyContextOrganize:
		return "上下文组织模块"
	case DolphinTplKeyRelatedQuestions:
		return "相关问题模块"
	default:
		return ""
	}
}
