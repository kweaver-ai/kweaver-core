package agentsvc

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/stretchr/testify/assert"
)

func TestAgentSvc_ToolHandle_NilAnswer(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	ctx := context.Background()
	skillsItem := &conversationmsgvo.SkillsProcessItem{}
	existing := []*conversationmsgvo.SkillsProcessItem{{AgentName: "existing"}}

	dto := &toolHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{nil},
		i:                 0,
		skillsItem:        skillsItem,
		skillsProcess:     existing,
	}

	result, err := svc.toolHandle(ctx, dto)
	assert.NoError(t, err)
	// nil answer 时直接返回原 skillsProcess，不追加
	assert.Equal(t, existing, result)
}

func TestAgentSvc_ToolHandle_StringAnswer(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	ctx := context.Background()
	skillsItem := &conversationmsgvo.SkillsProcessItem{}

	dto := &toolHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{
			{Answer: "hello world", Think: "thinking..."},
		},
		i:             0,
		skillsItem:    skillsItem,
		skillsProcess: []*conversationmsgvo.SkillsProcessItem{},
	}

	result, err := svc.toolHandle(ctx, dto)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "hello world", result[0].Text)
	assert.Equal(t, "thinking...", result[0].Thinking)
}

func TestAgentSvc_ToolHandle_NonStringAnswer(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}

	ctx := context.Background()
	skillsItem := &conversationmsgvo.SkillsProcessItem{}

	dto := &toolHandleDto{
		exploreAnswerList: []*agentrespvo.AnswerExplore{
			{Answer: map[string]interface{}{"key": "value"}, Think: "thinking"},
		},
		i:             0,
		skillsItem:    skillsItem,
		skillsProcess: []*conversationmsgvo.SkillsProcessItem{},
	}

	result, err := svc.toolHandle(ctx, dto)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	// 非 string answer 会被序列化为 markdown json 格式
	assert.Contains(t, result[0].Text, "```json")
}

func TestAgentSvc_ToAnswerJSONStr_ValidAnswer(t *testing.T) {
	t.Parallel()

	svc := &agentSvc{SvcBase: service.NewSvcBase()}

	ctx := context.Background()
	skillsItem := &conversationmsgvo.SkillsProcessItem{}
	answer := &agentrespvo.AnswerExplore{
		Answer: map[string]interface{}{"result": "ok"},
	}

	dto := &toolHandleDto{skillsItem: skillsItem}

	svc.toAnswerJSONStr(ctx, dto, answer)
	assert.Contains(t, skillsItem.Text, "```json")
	assert.Contains(t, skillsItem.Text, "result")
}

func TestAgentSvc_ToAnswerJSONStr_UnmarshalableAnswer(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{SvcBase: service.NewSvcBase(), logger: mockLogger}

	ctx := context.Background()
	skillsItem := &conversationmsgvo.SkillsProcessItem{}
	// channel 类型无法被 json.Marshal
	answer := &agentrespvo.AnswerExplore{
		Answer: make(chan int),
	}

	dto := &toolHandleDto{skillsItem: skillsItem}

	svc.toAnswerJSONStr(ctx, dto, answer)
	assert.Equal(t, "", skillsItem.Text)
}
