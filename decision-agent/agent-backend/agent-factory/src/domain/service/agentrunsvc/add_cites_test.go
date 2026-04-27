package agentsvc

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/stretchr/testify/assert"
)

func TestAddCitesToProgress_EmptyProgresses(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	ctx := context.Background()
	progresses := []*agentrespvo.Progress{}

	result := svc.addCitesToProgress(ctx, progresses, false)

	assert.Empty(t, result)
}

func TestAddCitesToProgress_NonDocQaAgent(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	ctx := context.Background()
	progresses := []*agentrespvo.Progress{
		{AgentName: "other_agent", Status: "completed", Answer: "some answer"},
	}

	result := svc.addCitesToProgress(ctx, progresses, false)

	assert.Len(t, result, 1)
	assert.Equal(t, "some answer", result[0].Answer)
}

func TestAddCitesToProgress_DocQaNotCompleted(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	ctx := context.Background()
	progresses := []*agentrespvo.Progress{
		{AgentName: "doc_qa", Status: "processing", Answer: "some answer"},
	}

	result := svc.addCitesToProgress(ctx, progresses, false)

	assert.Len(t, result, 1)
	assert.Equal(t, "some answer", result[0].Answer)
}

func TestAddCitesToProgress_DocQaCompleted_WithAnswer(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	ctx := context.Background()
	answer := map[string]interface{}{
		"full_result": map[string]interface{}{
			"text":       "result text",
			"references": []interface{}{},
		},
	}
	progresses := []*agentrespvo.Progress{
		{AgentName: "doc_qa", Status: "completed", Answer: answer},
	}

	result := svc.addCitesToProgress(ctx, progresses, false)

	assert.Len(t, result, 1)
}

func TestAddCitesToProgress_DocQaCompleted_NonSerializableAnswer(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	ctx := context.Background()
	// channel 类型不可序列化
	ch := make(chan int)
	progresses := []*agentrespvo.Progress{
		{AgentName: "doc_qa", Status: "completed", Answer: ch},
	}

	result := svc.addCitesToProgress(ctx, progresses, false)

	assert.Len(t, result, 1)
}
