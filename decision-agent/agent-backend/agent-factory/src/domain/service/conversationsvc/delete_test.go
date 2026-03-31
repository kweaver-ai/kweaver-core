package conversationsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestConversationSvc_Delete_PanicsWithoutConversationRepo(t *testing.T) {
	svc := &conversationSvc{
		SvcBase: service.NewSvcBase(),
		// conversationRepo is nil
	}

	ctx := context.Background()
	conversationID := "conv-123"

	// This will panic because conversationRepo is nil
	assert.Panics(t, func() {
		_ = svc.Delete(ctx, conversationID)
	})
}

func TestConversationSvc_Delete_ConversationNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConversationRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockConversationMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &conversationSvc{
		SvcBase:             service.NewSvcBase(),
		conversationRepo:    mockConversationRepo,
		conversationMsgRepo: mockConversationMsgRepo,
		logger:              mockLogger,
	}

	ctx := context.Background()
	conversationID := "non-existent-conv"

	// Expect GetByID to return not found error and logger error call
	mockConversationRepo.EXPECT().GetByID(gomock.Any(), conversationID).Return(nil, sql.ErrNoRows)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	err := svc.Delete(ctx, conversationID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "数据智能体配置不存在")
}

func TestConversationSvc_Delete_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConversationRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockConversationMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &conversationSvc{
		SvcBase:             service.NewSvcBase(),
		conversationRepo:    mockConversationRepo,
		conversationMsgRepo: mockConversationMsgRepo,
		logger:              mockLogger,
	}

	ctx := context.Background()
	conversationID := "conv-123"

	dbErr := errors.New("database connection failed")
	mockConversationRepo.EXPECT().GetByID(gomock.Any(), conversationID).Return(nil, dbErr)

	err := svc.Delete(ctx, conversationID)

	assert.Error(t, err)
}

func TestConversationSvc_Delete_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConversationRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
	mockConversationMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &conversationSvc{
		SvcBase:             service.NewSvcBase(),
		conversationRepo:    mockConversationRepo,
		conversationMsgRepo: mockConversationMsgRepo,
		logger:              mockLogger,
	}

	ctx := context.Background()
	conversationID := "conv-123"

	// Return conversation successfully but fail on BeginTx
	mockConversationRepo.EXPECT().GetByID(gomock.Any(), conversationID).Return(&dapo.ConversationPO{}, nil)

	txErr := errors.New("transaction begin failed")
	mockConversationRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, txErr)

	err := svc.Delete(ctx, conversationID)

	assert.Error(t, err)
}
