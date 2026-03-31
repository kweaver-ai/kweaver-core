package v3agentconfigsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetTx(t *testing.T) {
	t.Run("returns transaction when BeginTx succeeds", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockTx := &sql.Tx{}

		mockRepo.EXPECT().BeginTx(gomock.Any()).Return(mockTx, nil)

		svc := &dataAgentConfigSvc{
			SvcBase:      service.NewSvcBase(),
			agentTplRepo: mockRepo,
		}

		tx, err := svc.getTx(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, mockTx, tx)
	})

	t.Run("returns error when BeginTx fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		expectedErr := errors.New("database error")

		mockRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, expectedErr)

		svc := &dataAgentConfigSvc{
			SvcBase:      service.NewSvcBase(),
			agentTplRepo: mockRepo,
		}

		tx, err := svc.getTx(context.Background())

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.Contains(t, err.Error(), "开启事务失败")
	})
}
