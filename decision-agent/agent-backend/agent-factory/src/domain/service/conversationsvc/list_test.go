package conversationsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestConversationSvc_List_PanicsWithoutConversationRepo(t *testing.T) {
	t.Parallel()

	svc := &conversationSvc{
		SvcBase: service.NewSvcBase(),
		// conversationRepo is nil
	}

	ctx := context.Background()
	req := conversationreq.ListReq{
		AgentAPPKey: "test-app-key",
	}

	assert.Panics(t, func() {
		_, _, _ = svc.List(ctx, req)
	})
}

func TestConversationSvc_ListByAgentID_PanicsWithoutConversationRepo(t *testing.T) {
	t.Parallel()

	svc := &conversationSvc{
		SvcBase: service.NewSvcBase(),
		// conversationRepo is nil
	}

	ctx := context.Background()
	agentID := "agent-123"

	assert.Panics(t, func() {
		_, _, _ = svc.ListByAgentID(ctx, agentID, "", 1, 10, 0, 0)
	})
}

func TestConversationSvc_List_DatabaseError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConversationRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockConversationMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)

	svc := &conversationSvc{
		SvcBase:             service.NewSvcBase(),
		conversationRepo:    mockConversationRepo,
		conversationMsgRepo: mockConversationMsgRepo,
	}

	ctx := context.Background()
	req := conversationreq.ListReq{
		AgentAPPKey: "test-app-key",
		Title:       "test title",
	}

	dbErr := errors.New("database connection failed")
	mockConversationRepo.EXPECT().List(gomock.Any(), req).Return(nil, int64(0), dbErr)

	result, count, err := svc.List(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get conversation list error")
	assert.Empty(t, result)
	assert.Zero(t, count)
}

func TestConversationSvc_ListByAgentID_DatabaseError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConversationRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockConversationMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)

	svc := &conversationSvc{
		SvcBase:             service.NewSvcBase(),
		conversationRepo:    mockConversationRepo,
		conversationMsgRepo: mockConversationMsgRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"

	dbErr := errors.New("database error")
	mockConversationRepo.EXPECT().ListByAgentID(gomock.Any(), agentID, "", 1, 10).Return(nil, int64(0), dbErr)

	result, count, err := svc.ListByAgentID(ctx, agentID, "", 1, 10, 0, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get conversation list by agentID error")
	assert.Nil(t, result)
	assert.Zero(t, count)
}
