package daresvo

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentconfigvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
	"github.com/pkg/errors"
)

var DefOutputConf = &agentconfigvo.OutputVariablesS{
	AnswerVar:           "answer",
	DocRetrievalVar:     "doc_retrieval_res",
	GraphRetrievalVar:   "graph_retrieval_res",
	RelatedQuestionsVar: "related_questions",
	// OtherVars:           []string{"search_querys", "search_results"},
	OtherVars: []string{},
}

// DataAgentRes 表示数据代理响应的结构体
type DataAgentRes struct {
	Answer *agentrespvo.AnswerS `json:"answer"`
	// UserDefine map[string]interface{} `json:"user_define,omitempty"`
	InterruptInfo *v2agentexecutordto.ToolInterruptInfo `json:"interrupt_info,omitempty"`
	AgentRunID    string                                `json:"agent_run_id,omitempty"` // Agent 运行 ID（从 Executor 返回）
	Status        string                                `json:"status"`
	Error         interface{}                           `json:"error"`

	finalAnswerVarHelper      *ResHelper
	docRetrievalVarHelper     *ResHelper
	graphRetrievalVarHelper   *ResHelper
	relatedQuestionsVarHelper *ResHelper
	otherVarsHelper           *ResHelper
	middleOutputVarsHelper    *ResHelper
}

func NewDataAgentRes(_ context.Context, data []byte, outputVariablesS *agentconfigvo.OutputVariablesS) (*DataAgentRes, error) {
	var err error

	r := &DataAgentRes{
		Answer: agentrespvo.NewAnswerS(),
		// UserDefine: make(map[string]interface{}),
	}

	err = r.loadFromMessage(data)
	if err != nil {
		// panic(err)
		err = errors.Wrapf(err, "loadFromMessage error: %s", string(data))
		return nil, err
	}

	r.finalAnswerVarHelper = NewResHelper(r.Answer, outputVariablesS, VarFieldTypeFinalAnswer)
	r.docRetrievalVarHelper = NewResHelper(r.Answer, outputVariablesS, VarFieldTypeDocRetrieval)
	r.graphRetrievalVarHelper = NewResHelper(r.Answer, outputVariablesS, VarFieldTypeGraphRetrieval)
	r.relatedQuestionsVarHelper = NewResHelper(r.Answer, outputVariablesS, VarFieldTypeRelatedQuestions)
	r.otherVarsHelper = NewResHelper(r.Answer, outputVariablesS, VarFieldTypeOther)
	r.middleOutputVarsHelper = NewResHelper(r.Answer, outputVariablesS, VarFieldTypeMiddleOutputVars)

	return r, nil
}

func (r *DataAgentRes) loadFromMessage(msg []byte) (err error) {
	err = sonic.Unmarshal(msg, r)
	return
}

func (r *DataAgentRes) GetAnswerHelper() *ResHelper {
	return r.finalAnswerVarHelper
}
