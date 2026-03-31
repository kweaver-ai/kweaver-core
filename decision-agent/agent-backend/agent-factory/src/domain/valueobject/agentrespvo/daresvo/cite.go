package daresvo

import (
	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/chat_enum/chatresenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/pkg/errors"
)

func (r *DataAgentRes) DocRetrievalAnswerAndCites() (answer string, cites []*agentrespvo.AnswerCite, err error) {
	docRetrieval, err := r.GetDocRetrieval()
	if err != nil {
		err = errors.Wrapf(err, "[DocRetrievalAnswerAndCites] GetDocRetrieval error:%v", err)
		return
	}

	if docRetrieval == nil {
		return
	}

	answer, cites = docRetrieval.AnswerAndCites()

	return
}

func (r *DataAgentRes) GetDocRetrieval() (docRetrievalRes *agentrespvo.DocRetrievalRes, err error) {
	bys, err := r.docRetrievalVarHelper.GetSingleFieldJSON()
	if err != nil {
		err = errors.Wrapf(err, "[GetDocRetrieval] GetSingleFieldJSON error:%v,res: %s", err, string(bys))
		return
	}

	if len(bys) == 0 {
		return
	}
	// NOTE: 如果docRetrievalRes为空，则返回空，工具刚开始，还没返回结果的时候可能为空字符串，是个string类型，无法进行Unmarshal
	if string(bys) == "\"\"" {
		return
	}

	err = sonic.Unmarshal(bys, &docRetrievalRes)
	if err != nil {
		err = errors.Wrapf(err, "[GetDocRetrieval] Unmarshal error:%v,res: %s", err, string(bys))
		return
	}

	manager := NewDocRetrievalManager()

	docRetrievalRes.Answer, err = manager.ProcessResult(docRetrievalRes.Answer, chatresenum.DocRetrievalStrategyStandard)
	if err != nil {
		err = errors.Wrapf(err, "[GetDocRetrieval] ProcessResult error:%v,res: %s", err, string(bys))
		return
	}

	return
}
