package conversationsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestConversationSvc_DeleteByAppKey_BeginTxError(t *testing.T) {
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
	appKey := "app-key-123"

	txErr := errors.New("transaction begin failed")
	mockConversationRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, txErr)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	err := svc.DeleteByAppKey(ctx, appKey)

	assert.Error(t, err)
}
