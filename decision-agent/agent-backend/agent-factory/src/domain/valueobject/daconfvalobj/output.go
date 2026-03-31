package daconfvalobj

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/pkg/errors"
)

// VariablesS 表示输出变量 (S: struct)
type VariablesS struct {
	AnswerVar           string   `json:"answer_var"`            // 包含最终回答的变量名
	DocRetrievalVar     string   `json:"doc_retrieval_var"`     // 包含文档检索结果的变量名
	GraphRetrievalVar   string   `json:"graph_retrieval_var"`   // 包含图谱检索结果的变量名
	RelatedQuestionsVar string   `json:"related_questions_var"` // 包含相关问题的变量名
	OtherVars           []string `json:"other_vars"`            // 其他变量数组
	MiddleOutputVars    []string `json:"middle_output_vars"`    // 中间输出变量数组
}

// Output 表示输出结果
type Output struct {
	Variables     *VariablesS                 `json:"variables"`      // 变量名
	DefaultFormat cdaenum.OutputDefaultFormat `json:"default_format"` // 默认输出格式
}

func (p *Output) ValObjCheck(isDolphinMode bool) (err error) {
	if err = p.DefaultFormat.EnumCheck(); err != nil {
		err = errors.Wrap(err, "[Output]: default_format is invalid")
		return
	}

	if p.Variables == nil {
		p.Variables = &VariablesS{}
	}

	if p.Variables.AnswerVar == "" {
		if isDolphinMode {
			err = errors.New("[Output]: answer_var is required when dolphin_mode")
			return
		}

		p.Variables.AnswerVar = "answer"
	}

	if p.Variables.DocRetrievalVar == "" {
		p.Variables.DocRetrievalVar = "doc_retrieval_res"
	}

	if p.Variables.GraphRetrievalVar == "" {
		p.Variables.GraphRetrievalVar = "graph_retrieval_res"
	}

	if p.Variables.RelatedQuestionsVar == "" {
		p.Variables.RelatedQuestionsVar = "related_questions"
	}

	return
}
