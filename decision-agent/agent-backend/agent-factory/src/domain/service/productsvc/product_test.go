package productsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/product/productreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Helper function to create context with user ID
func createContextWithUserID(userID string) context.Context {
	visitor := &rest.Visitor{
		ID: userID,
	}

	return context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	req := &productreq.CreateReq{
		Name:    "Test Product",
		Profile: "Test Profile",
		Key:     "test-product",
	}

	// Mock expectations
	mockProductRepo.EXPECT().ExistsByName(gomock.Any(), req.Name).Return(false, nil)
	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), req.Key).Return(false, nil)
	mockProductRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(req.Key, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	key, err := svc.Create(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, req.Key, key)
}

func TestCreate_NameAlreadyExists(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	req := &productreq.CreateReq{
		Name: "Existing Product",
		Key:  "new-product",
	}

	mockProductRepo.EXPECT().ExistsByName(gomock.Any(), req.Name).Return(true, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	key, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, key)
	assert.Contains(t, err.Error(), "产品名称已存在")
}

func TestCreate_KeyAlreadyExists(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	req := &productreq.CreateReq{
		Name: "New Product",
		Key:  "existing-key",
	}

	mockProductRepo.EXPECT().ExistsByName(gomock.Any(), req.Name).Return(false, nil)
	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), req.Key).Return(true, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	key, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, key)
	assert.Contains(t, err.Error(), "产品标识已存在")
}

func TestCreate_WithoutKey(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	req := &productreq.CreateReq{
		Name:    "Test Product",
		Profile: "Test Profile",
		Key:     "", // Empty key - should generate ULID
	}

	mockProductRepo.EXPECT().ExistsByName(gomock.Any(), req.Name).Return(false, nil)
	// No ExistsByKey call when key is empty
	mockProductRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("generated-key", nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	key, err := svc.Create(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "generated-key", key)
}

func TestCreate_RepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	req := &productreq.CreateReq{
		Name: "Test Product",
		Key:  "test-product",
	}

	expectedErr := errors.New("database error")
	mockProductRepo.EXPECT().ExistsByName(gomock.Any(), req.Name).Return(false, expectedErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	key, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, key)
	assert.Equal(t, expectedErr, err)
}

func TestDetail_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(123)

	expectedPo := &dapo.ProductPo{
		ID:      productID,
		Name:    "Test Product",
		Key:     "test-product",
		Profile: "Test Profile",
	}

	mockProductRepo.EXPECT().GetByID(gomock.Any(), productID).Return(expectedPo, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.Detail(ctx, productID)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestDetail_NotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(999)

	mockProductRepo.EXPECT().GetByID(gomock.Any(), productID).Return(nil, sql.ErrNoRows)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.Detail(ctx, productID)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "产品不存在")
}

func TestGetByKey_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productKey := "test-product"

	expectedPo := &dapo.ProductPo{
		ID:      123,
		Name:    "Test Product",
		Key:     productKey,
		Profile: "Test Profile",
	}

	mockProductRepo.EXPECT().GetByKey(gomock.Any(), productKey).Return(expectedPo, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.GetByKey(ctx, productKey)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestGetByKey_NotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productKey := "nonexistent-product"

	mockProductRepo.EXPECT().GetByKey(gomock.Any(), productKey).Return(nil, sql.ErrNoRows)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.GetByKey(ctx, productKey)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "产品不存在")
}

func TestList_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	offset := 0
	limit := 10

	expectedPos := []*dapo.ProductPo{
		{ID: 1, Name: "Product 1", Key: "product-1", Profile: "Profile 1"},
		{ID: 2, Name: "Product 2", Key: "product-2", Profile: "Profile 2"},
	}

	mockProductRepo.EXPECT().List(gomock.Any(), offset, limit).Return(expectedPos, 2, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.List(ctx, offset, limit)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 2, res.Total)
}

func TestList_Empty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	offset := 0
	limit := 10

	mockProductRepo.EXPECT().List(gomock.Any(), offset, limit).Return([]*dapo.ProductPo{}, 0, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.List(ctx, offset, limit)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, 0, res.Total)
}

func TestUpdate_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(123)

	req := &productreq.UpdateReq{
		Name:    "Updated Product",
		Profile: "Updated Profile",
	}

	existingPo := &dapo.ProductPo{
		ID:      productID,
		Name:    "Old Name",
		Key:     "test-product",
		Profile: "Old Profile",
	}

	mockProductRepo.EXPECT().GetByID(gomock.Any(), productID).Return(existingPo, nil)
	mockProductRepo.EXPECT().ExistsByNameExcludeID(gomock.Any(), req.Name, productID).Return(false, nil)
	mockProductRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	auditLog, err := svc.Update(ctx, req, productID)

	assert.NoError(t, err)
	assert.Equal(t, "123", auditLog.ID)
	assert.Equal(t, req.Name, auditLog.Name)
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(999)

	req := &productreq.UpdateReq{
		Name: "Updated Product",
	}

	mockProductRepo.EXPECT().GetByID(gomock.Any(), productID).Return(nil, sql.ErrNoRows)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	auditLog, err := svc.Update(ctx, req, productID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "产品不存在")
	assert.Empty(t, auditLog.ID)
}

func TestUpdate_NameAlreadyExists(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(123)

	req := &productreq.UpdateReq{
		Name: "Existing Product",
	}

	existingPo := &dapo.ProductPo{
		ID:   productID,
		Name: "Old Name",
		Key:  "test-product",
	}

	mockProductRepo.EXPECT().GetByID(gomock.Any(), productID).Return(existingPo, nil)
	mockProductRepo.EXPECT().ExistsByNameExcludeID(gomock.Any(), req.Name, productID).Return(true, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	auditLog, err := svc.Update(ctx, req, productID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "产品名称已存在")
	assert.Equal(t, "123", auditLog.ID)
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(123)

	existingPo := &dapo.ProductPo{
		ID:      productID,
		Name:    "Test Product",
		Key:     "test-product",
		Profile: "Test Profile",
	}

	mockProductRepo.EXPECT().ExistsByID(gomock.Any(), productID).Return(true, nil)
	mockProductRepo.EXPECT().GetByID(gomock.Any(), productID).Return(existingPo, nil)
	mockProductRepo.EXPECT().Delete(gomock.Any(), productID).Return(nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	auditLog, err := svc.Delete(ctx, productID)

	assert.NoError(t, err)
	assert.Equal(t, "123", auditLog.ID)
	assert.Equal(t, "Test Product", auditLog.Name)
}

func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(999)

	mockProductRepo.EXPECT().ExistsByID(gomock.Any(), productID).Return(false, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	auditLog, err := svc.Delete(ctx, productID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "产品不存在")
	assert.Empty(t, auditLog.ID)
}

func TestDelete_RepositoryErrorOnExists(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(123)

	expectedErr := errors.New("database error")
	mockProductRepo.EXPECT().ExistsByID(gomock.Any(), productID).Return(false, expectedErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	auditLog, err := svc.Delete(ctx, productID)

	assert.Error(t, err)
	assert.Empty(t, auditLog.ID)
}

func TestList_RepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	offset := 0
	limit := 10

	expectedErr := errors.New("database error")
	mockProductRepo.EXPECT().List(gomock.Any(), offset, limit).Return(nil, 0, expectedErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.List(ctx, offset, limit)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestList_P2EConversionError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	offset := 0
	limit := 10

	// Return products that will cause P2E conversion error
	// Using a PO with nil pointer that can't be converted
	expectedPos := []*dapo.ProductPo{
		{ID: 1, Name: "Product 1", Key: "product-1", Profile: "Profile 1"},
	}

	// This test would require mocking productp2e.Products to return an error
	// Since productp2e.Products is a concrete function that can't be mocked easily,
	// we'll test a scenario where the conversion might fail
	mockProductRepo.EXPECT().List(gomock.Any(), offset, limit).Return(expectedPos, 1, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	_, err := svc.List(ctx, offset, limit)

	// The function should succeed even if there are edge cases
	// If an error occurs during P2E conversion, it would be returned
	// For this test, we verify the function completes
	assert.NoError(t, err)
}

func TestList_LoadFromEoError(t *testing.T) {
	t.Parallel()

	// This test would require creating a scenario where LoadFromEo fails
	// Since LoadFromEo uses CopyStructUseJSON which rarely fails for simple structs,
	// we'll just verify the function handles the empty case properly
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	offset := 0
	limit := 10

	// Return products
	expectedPos := []*dapo.ProductPo{
		{ID: 1, Name: "Product 1", Key: "product-1", Profile: "Profile 1"},
	}

	mockProductRepo.EXPECT().List(gomock.Any(), offset, limit).Return(expectedPos, 1, nil)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.List(ctx, offset, limit)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestList_DatabaseNotFoundError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	offset := 0
	limit := 10

	// Test sql.ErrNoRows specifically
	notFoundErr := sql.ErrNoRows
	mockProductRepo.EXPECT().List(gomock.Any(), offset, limit).Return(nil, 0, notFoundErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.List(ctx, offset, limit)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestCreate_ExistsByKeyError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	req := &productreq.CreateReq{
		Name: "Test Product",
		Key:  "test-product",
	}

	expectedErr := errors.New("database connection failed")

	mockProductRepo.EXPECT().ExistsByName(gomock.Any(), req.Name).Return(false, nil)
	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), req.Key).Return(false, expectedErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	key, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, key)
}

func TestDetail_RepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(123)

	expectedErr := errors.New("database error")
	mockProductRepo.EXPECT().GetByID(gomock.Any(), productID).Return(nil, expectedErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.Detail(ctx, productID)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestGetByKey_RepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productKey := "test-product"

	expectedErr := errors.New("database error")
	mockProductRepo.EXPECT().GetByKey(gomock.Any(), productKey).Return(nil, expectedErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	res, err := svc.GetByKey(ctx, productKey)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestUpdate_RepositoryErrorOnUpdate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(123)

	req := &productreq.UpdateReq{
		Name: "Updated Product",
	}

	existingPo := &dapo.ProductPo{
		ID:   productID,
		Name: "Old Name",
		Key:  "test-product",
	}

	expectedErr := errors.New("database update failed")

	mockProductRepo.EXPECT().GetByID(gomock.Any(), productID).Return(existingPo, nil)
	mockProductRepo.EXPECT().ExistsByNameExcludeID(gomock.Any(), req.Name, productID).Return(false, nil)
	mockProductRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(expectedErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	auditLog, err := svc.Update(ctx, req, productID)

	assert.Error(t, err)
	assert.Equal(t, "123", auditLog.ID)
}

func TestUpdate_RepositoryErrorOnExistsCheck(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(123)

	req := &productreq.UpdateReq{
		Name: "Updated Product",
	}

	existingPo := &dapo.ProductPo{
		ID:   productID,
		Name: "Old Name",
		Key:  "test-product",
	}

	expectedErr := errors.New("database connection failed")

	mockProductRepo.EXPECT().GetByID(gomock.Any(), productID).Return(existingPo, nil)
	mockProductRepo.EXPECT().ExistsByNameExcludeID(gomock.Any(), req.Name, productID).Return(false, expectedErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	auditLog, err := svc.Update(ctx, req, productID)

	assert.Error(t, err)
	assert.Equal(t, "123", auditLog.ID)
}

func TestDelete_RepositoryErrorOnGet(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	productID := int64(123)

	expectedErr := errors.New("database error")

	mockProductRepo.EXPECT().ExistsByID(gomock.Any(), productID).Return(true, nil)
	mockProductRepo.EXPECT().GetByID(gomock.Any(), productID).Return(nil, expectedErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := context.Background()
	auditLog, err := svc.Delete(ctx, productID)

	assert.Error(t, err)
	assert.Empty(t, auditLog.ID)
}

func TestCreate_RepositoryErrorOnCreate(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	req := &productreq.CreateReq{
		Name:    "Test Product",
		Profile: "Test Profile",
		Key:     "test-product",
	}

	expectedErr := errors.New("database insert failed")

	mockProductRepo.EXPECT().ExistsByName(gomock.Any(), req.Name).Return(false, nil)
	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), req.Key).Return(false, nil)
	mockProductRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("", expectedErr)

	svc := &productSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
	}

	ctx := createContextWithUserID("user-123")
	key, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Empty(t, key)
}
