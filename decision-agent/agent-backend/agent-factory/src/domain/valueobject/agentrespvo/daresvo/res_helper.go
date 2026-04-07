package daresvo

import (
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentconfigvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/pkg/errors"
)

type VarFieldType string

const (
	VarFieldTypeFinalAnswer      VarFieldType = "answer"
	VarFieldTypeDocRetrieval     VarFieldType = "doc_retrieval"
	VarFieldTypeGraphRetrieval   VarFieldType = "graph_retrieval"
	VarFieldTypeRelatedQuestions VarFieldType = "related_questions"
	VarFieldTypeOther            VarFieldType = "other_vars"
	VarFieldTypeMiddleOutputVars VarFieldType = "middle_output_vars"
)

type ResHelper struct {
	Answer           *agentrespvo.AnswerS
	outputVariablesS *agentconfigvo.OutputVariablesS

	varField VarFieldType
}

func NewResHelper(answer *agentrespvo.AnswerS, outputVariablesS *agentconfigvo.OutputVariablesS, varField VarFieldType) *ResHelper {
	return &ResHelper{
		Answer:           answer,
		outputVariablesS: outputVariablesS,
		varField:         varField,
	}
}

func (r *ResHelper) getSingleFieldVal() interface{} {
	var varField string
	switch r.varField {
	case VarFieldTypeFinalAnswer:
		varField = r.outputVariablesS.AnswerVar
	case VarFieldTypeDocRetrieval:
		varField = r.outputVariablesS.DocRetrievalVar
	case VarFieldTypeGraphRetrieval:
		varField = r.outputVariablesS.GraphRetrievalVar
	case VarFieldTypeRelatedQuestions:
		varField = r.outputVariablesS.RelatedQuestionsVar
	default:
		panic(fmt.Sprintf("not support varField: %s", r.varField))
	}

	value, ok := r.Answer.GetField(varField)
	if !ok {
		return nil
	}

	return value
}

func (r *ResHelper) GetOtherVarsMap() (m map[string]interface{}, err error) {
	fields := r.outputVariablesS.OtherVars
	m = r.Answer.GetFields(fields)

	return
}

func (r *ResHelper) GetMiddleOutputVarsMap() (m map[string]interface{}, err error) {
	fields := r.outputVariablesS.MiddleOutputVars
	m = r.Answer.GetFields(fields)

	return
}

func (r *ResHelper) GetSingleFieldJSON() (byt []byte, err error) {
	answer := r.getSingleFieldVal()

	if answer == nil {
		return
	}
	// NOTE: 暴露错误
	if answewrMap, ok := answer.(map[string]interface{}); ok {
		if _, ok := answewrMap["error_code"]; ok {
			err = errors.Wrapf(err, "[GetSingleFieldJSON] failed, error :%v", answer)
			return
		}
	}

	byt, err = sonic.Marshal(answer)
	if err != nil {
		err = errors.Wrapf(err, "GetSingleFieldJSON error:%v,res: %s\n", err, string(byt))
	}

	return
}
