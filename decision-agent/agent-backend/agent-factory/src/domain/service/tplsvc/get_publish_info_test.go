package tplsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDataAgentTplSvc_GetPublishInfo_PanicsWithoutPublishedTplRepo(t *testing.T) {
	t.Parallel()

	svc := &dataAgentTplSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	id := int64(123)

	assert.Panics(t, func() {
		_, _ = svc.GetPublishInfo(ctx, id)
	})
}

func TestDataAgentTplSvc_GetPublishInfo_PublishedTplNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          service.NewSvcBase(),
		publishedTplRepo: mockPublishedTplRepo,
		logger:           mockLogger,
	}

	ctx := context.Background()
	id := int64(999)

	mockPublishedTplRepo.EXPECT().GetByTplID(gomock.Any(), id).Return(nil, sql.ErrNoRows)

	res, err := svc.GetPublishInfo(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "此已发布模板不存在")
}

func TestDataAgentTplSvc_GetPublishInfo_GetByTplIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          service.NewSvcBase(),
		publishedTplRepo: mockPublishedTplRepo,
		logger:           mockLogger,
	}

	ctx := context.Background()
	id := int64(123)

	dbErr := errors.New("database connection failed")
	mockPublishedTplRepo.EXPECT().GetByTplID(gomock.Any(), id).Return(nil, dbErr)

	res, err := svc.GetPublishInfo(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestDataAgentTplSvc_GetPublishInfo_GetCategoryJoinPosByTplIDError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          service.NewSvcBase(),
		publishedTplRepo: mockPublishedTplRepo,
		logger:           mockLogger,
	}

	ctx := context.Background()
	id := int64(123)

	publishedPo := &dapo.PublishedTplPo{
		ID: 456,
	}

	mockPublishedTplRepo.EXPECT().GetByTplID(gomock.Any(), id).Return(publishedPo, nil)

	dbErr := errors.New("database connection failed")
	mockPublishedTplRepo.EXPECT().GetCategoryJoinPosByTplID(gomock.Any(), nil, publishedPo.ID).Return(nil, dbErr)

	res, err := svc.GetPublishInfo(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestDataAgentTplSvc_GetPublishInfo_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentTplSvc{
		SvcBase:          service.NewSvcBase(),
		publishedTplRepo: mockPublishedTplRepo,
		logger:           mockLogger,
	}

	ctx := context.Background()
	id := int64(123)

	publishedPo := &dapo.PublishedTplPo{
		ID: 456,
	}

	categoryJoins := []*dapo.DataAgentTplCategoryJoinPo{
		{
			CategoryID:   "cat-1",
			CategoryName: "Category 1",
		},
		{
			CategoryID:   "cat-2",
			CategoryName: "Category 2",
		},
	}

	mockPublishedTplRepo.EXPECT().GetByTplID(gomock.Any(), id).Return(publishedPo, nil)
	mockPublishedTplRepo.EXPECT().GetCategoryJoinPosByTplID(gomock.Any(), nil, publishedPo.ID).Return(categoryJoins, nil)

	res, err := svc.GetPublishInfo(ctx, id)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Categories, 2)
}
