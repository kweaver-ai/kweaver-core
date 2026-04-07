package tplsvc

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

func TestDataAgentTplSvc_Detail_PanicsWithoutAgentTplRepo(t *testing.T) {
	t.Parallel()

	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	id := int64(123)

	assert.Panics(t, func() {
		_, _ = svc.Detail(ctx, id)
	})
}

func TestDataAgentTplSvc_Detail_NotFoundError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
	}

	ctx := context.Background()
	id := int64(999)

	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, sql.ErrNoRows)

	res, err := svc.Detail(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "模板不存在")
}

func TestDataAgentTplSvc_Detail_GetByIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
	}

	ctx := context.Background()
	id := int64(123)

	dbErr := errors.New("database connection failed")
	mockAgentTplRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, dbErr)

	res, err := svc.Detail(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestDataAgentTplSvc_DetailByKey_PanicsWithoutAgentTplRepo(t *testing.T) {
	t.Parallel()

	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	key := "test-key"

	assert.Panics(t, func() {
		_, _ = svc.DetailByKey(ctx, key)
	})
}

func TestDataAgentTplSvc_DetailByKey_NotFoundError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
	}

	ctx := context.Background()
	key := "non-existent-key"

	notFoundErr := errors.New("record not found")
	mockAgentTplRepo.EXPECT().GetByKey(gomock.Any(), key).Return(nil, notFoundErr)

	res, err := svc.DetailByKey(ctx, key)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "get template by key")
}

func TestDataAgentTplSvc_DetailByKey_GetByKeyError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockAgentTplRepo,
	}

	ctx := context.Background()
	key := "test-key"

	dbErr := errors.New("database connection failed")
	mockAgentTplRepo.EXPECT().GetByKey(gomock.Any(), key).Return(nil, dbErr)

	res, err := svc.DetailByKey(ctx, key)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "get template by key")
}
