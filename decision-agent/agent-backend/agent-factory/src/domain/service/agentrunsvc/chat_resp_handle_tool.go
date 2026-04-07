package agentsvc

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
)

type toolHandleDto struct {
	exploreAnswerList []*agentrespvo.AnswerExplore
	i                 int
	skillsItem        *conversationmsgvo.SkillsProcessItem
	skillsProcess     []*conversationmsgvo.SkillsProcessItem
}

func (agentSvc *agentSvc) toolHandle(ctx context.Context, dto *toolHandleDto) (skillsProcess []*conversationmsgvo.SkillsProcessItem, err error) {
	_answer := dto.exploreAnswerList[dto.i]

	if _answer == nil {
		return dto.skillsProcess, nil
	}

	// 直接输出 answer 回答
	// 如果 answer 不是string 转成string
	if _, ok := _answer.Answer.(string); !ok {
		agentSvc.toAnswerJSONStr(ctx, dto, _answer)
	} else {
		dto.skillsItem.Text = _answer.Answer.(string)
	}

	dto.skillsItem.Thinking = _answer.Think

	dto.skillsProcess = append(dto.skillsProcess, dto.skillsItem)

	return dto.skillsProcess, nil
}

func (agentSvc *agentSvc) toAnswerJSONStr(ctx context.Context, dto *toolHandleDto, _answer *agentrespvo.AnswerExplore) {
	// 使用解析后的对象进行格式化
	answerTmp, err := json.MarshalIndent(_answer.Answer, "", "    ")
	if err != nil {
		agentSvc.logger.Errorf("toolHandle MarshalIndent  err: %v", err)

		dto.skillsItem.Text = ""
	} else {
		// 帮我给 skillsItem.Text 加一个 markdown 的格式
		dto.skillsItem.Text = "```json\n" + string(answerTmp) + "\n```"
	}
}
