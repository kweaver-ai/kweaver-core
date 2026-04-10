package publishedsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
)

func TestPublishedSvc_PubedTplDetail_NotFound(t *testing.T) {
	t.Parallel()

	initCGlobalConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	svc := &publishedSvc{
		publishedTplRepo: mockPublishedTplRepo,
		productRepo:      mockProductRepo,
	}

	mockPublishedTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(123)).
		Return(nil, sql.ErrNoRows)

	res, err := svc.PubedTplDetail(context.Background(), 123)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestPublishedSvc_PubedTplDetail_ProductRepoError(t *testing.T) {
	t.Parallel()

	initCGlobalConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	svc := &publishedSvc{
		publishedTplRepo: mockPublishedTplRepo,
		productRepo:      mockProductRepo,
	}

	mockPublishedTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(123)).
		Return(&dapo.PublishedTplPo{
			ID:         1,
			TplID:      123,
			Name:       "tpl-1",
			ProductKey: "p-key-1",
			Config:     "{}",
		}, nil)
	mockProductRepo.EXPECT().GetByKey(gomock.Any(), "p-key-1").
		Return(nil, errors.New("query product failed"))

	res, err := svc.PubedTplDetail(context.Background(), 123)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "get product name error")
}

func TestPublishedSvc_PubedTplDetail_Success(t *testing.T) {
	t.Parallel()

	initCGlobalConfig(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublishedTplRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	svc := &publishedSvc{
		publishedTplRepo: mockPublishedTplRepo,
		productRepo:      mockProductRepo,
	}

	mockPublishedTplRepo.EXPECT().GetByTplID(gomock.Any(), int64(123)).
		Return(&dapo.PublishedTplPo{
			ID:         1,
			TplID:      123,
			Name:       "tpl-1",
			ProductKey: "p-key-1",
			Config:     "{}",
		}, nil)
	mockProductRepo.EXPECT().GetByKey(gomock.Any(), "p-key-1").
		Return(&dapo.ProductPo{
			Name: "product-1",
		}, nil)

	res, err := svc.PubedTplDetail(context.Background(), 123)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int64(123), res.TplID)
	assert.Equal(t, "product-1", res.ProductName)
}
