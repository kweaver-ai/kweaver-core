package agentsvc

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
)

type agentToolHandleDto struct {
	exploreAnswerList []*agentrespvo.AnswerExplore
	i                 int
	skillsItem        *conversationmsgvo.SkillsProcessItem
	skillsProcess     []*conversationmsgvo.SkillsProcessItem
}

func (agentSvc *agentSvc) agentToolHandle(ctx context.Context, dto *agentToolHandleDto) (skillsProcess2 []*conversationmsgvo.SkillsProcessItem, err error) {
	_answer := dto.exploreAnswerList[dto.i]

	if _answer == nil {
		return dto.skillsProcess, nil
	}

	// 直接输出 answer 回答
	// 如果 answer 不是string 转成string
	if _, ok := _answer.Answer.(string); !ok {
		toolDto := &toolHandleDto{
			exploreAnswerList: dto.exploreAnswerList,
			i:                 dto.i,
			skillsItem:        dto.skillsItem,
			skillsProcess:     dto.skillsProcess,
		}
		agentSvc.toAnswerJSONStr(ctx, toolDto, _answer)
	} else {
		dto.skillsItem.Text = _answer.Answer.(string)
	}

	dto.skillsItem.Thinking = _answer.Think

	dto.skillsProcess = append(dto.skillsProcess, dto.skillsItem)

	return dto.skillsProcess, nil
}
