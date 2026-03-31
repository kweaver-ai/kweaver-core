package categorysvc

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewCategorySvc(t *testing.T) {
	t.Parallel()

	svc := NewCategorySvc()

	assert.NotNil(t, svc)
}

func TestCategorySvc_List(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := idbaccessmock.NewMockICategoryRepo(ctrl)

	svc := &categorySvc{
		categoryRepo: mockRepo,
	}

	// Mock successful response
	mockCategories := []*dapo.CategoryPO{
		{ID: "1", Name: "Category 1", Description: "Description 1"},
		{ID: "2", Name: "Category 2", Description: "Description 2"},
	}
	mockRepo.EXPECT().List(ctx, nil).Return(mockCategories, nil)

	result, err := svc.List(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "1", result[0].ID)
	assert.Equal(t, "Category 1", result[0].Name)
	assert.Equal(t, "Description 1", result[0].Description)
}

func TestCategorySvc_List_Empty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := idbaccessmock.NewMockICategoryRepo(ctrl)

	svc := &categorySvc{
		categoryRepo: mockRepo,
	}

	// Mock empty response
	mockRepo.EXPECT().List(ctx, nil).Return([]*dapo.CategoryPO{}, nil)

	result, err := svc.List(ctx)

	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestCategorySvc_List_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := idbaccessmock.NewMockICategoryRepo(ctrl)

	svc := &categorySvc{
		categoryRepo: mockRepo,
	}

	// Mock error response
	mockRepo.EXPECT().List(ctx, nil).Return(nil, assert.AnError)

	result, err := svc.List(ctx)

	assert.Error(t, err)
	assert.NotNil(t, result) // List returns empty list, not nil
	assert.Len(t, result, 0)
}
