package daresvo

import (
	// "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

func (r *DataAgentRes) GetExploreAnswerList() (answerList []*agentrespvo.AnswerExplore, ok bool) {
	answerList = make([]*agentrespvo.AnswerExplore, 0)

	// 1. 判断是否为 explore 类型
	isValid, err := r.isExploreType()
	if err != nil {
		return
	}

	if !isValid {
		return
	}

	// 2. 转换为 explore 类型的answerList
	err = cutil.CopyUseJSON(&answerList, r.GetFinalAnswer())
	if err != nil {
		answerList = nil
		return
	}

	ok = len(answerList) > 0

	return
}

func (r *DataAgentRes) isExploreType() (isValid bool, err error) {
	byt, err := r.GetFinalAnswerJSON()
	if err != nil {
		return
	}

	isValid, err = agentrespvo.IsExploreType(string(byt))
	if err != nil {
		return
	}

	if !isValid {
		return
	}

	return
}
