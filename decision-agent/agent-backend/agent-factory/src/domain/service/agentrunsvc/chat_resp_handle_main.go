package agentsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
)

type mainHandleDto struct {
	skillsItem        *conversationmsgvo.SkillsProcessItem
	exploreAnswerList []*agentrespvo.AnswerExplore
	i                 int
	mainThinking      *string
	skillsProcess     []*conversationmsgvo.SkillsProcessItem
}

func (agentSvc *agentSvc) mainHandle(ctx context.Context, dto *mainHandleDto) []*conversationmsgvo.SkillsProcessItem {
	_answer := dto.exploreAnswerList[dto.i]

	if _answer == nil {
		return dto.skillsProcess
	}

	dto.skillsItem.Text = _answer.Answer.(string)
	dto.skillsItem.Thinking = _answer.Think

	*dto.mainThinking += _answer.Think

	if len(dto.skillsItem.Text) != 0 || len(dto.skillsItem.Thinking) != 0 {
		dto.skillsProcess = append(dto.skillsProcess, dto.skillsItem)
	}

	return dto.skillsProcess
}
