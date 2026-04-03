package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"oss-gateway/internal/handler"
	"oss-gateway/internal/service"
	"oss-gateway/pkg/adapter"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorageService 是 StorageService 的 mock 实现
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) Create(ctx context.Context, req *service.CreateStorageRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) Update(ctx context.Context, storageID string, req *service.UpdateStorageRequest) error {
	args := m.Called(ctx, storageID, req)
	return args.Error(0)
}

func (m *MockStorageService) Delete(ctx context.Context, storageID string) error {
	args := m.Called(ctx, storageID)
	return args.Error(0)
}

func (m *MockStorageService) Get(ctx context.Context, storageID string) (*service.StorageResponse, error) {
	args := m.Called(ctx, storageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.StorageResponse), args.Error(1)
}

func (m *MockStorageService) List(ctx context.Context, req *service.ListStorageRequest) (*service.ListStorageResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ListStorageResponse), args.Error(1)
}

func (m *MockStorageService) CheckConnection(ctx context.Context, storageID string) error {
	args := m.Called(ctx, storageID)
	return args.Error(0)
}

func (m *MockStorageService) GetAdapter(ctx context.Context, storageID string, useInternal bool) (adapter.OSSAdapter, error) {
	args := m.Called(ctx, storageID, useInternal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(adapter.OSSAdapter), args.Error(1)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestStorageHandler_Create_Success(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.POST("/storages", h.Create)

	req := service.CreateStorageRequest{
		StorageName:     "test-storage",
		VendorType:      "OSS",
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		BucketName:      "test-bucket",
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-secret",
		Region:          "cn-hangzhou",
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*service.CreateStorageRequest")).
		Return("storage123", nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/storages", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_Create_InvalidJSON(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.POST("/storages", h.Create)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/storages", bytes.NewBufferString("{invalid json}"))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStorageHandler_Create_ValidationError_StorageNameExists(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.POST("/storages", h.Create)

	req := service.CreateStorageRequest{
		StorageName:     "existing-storage",
		VendorType:      "OSS",
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		BucketName:      "test-bucket",
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-secret",
		Region:          "cn-hangzhou",
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*service.CreateStorageRequest")).
		Return("", &service.StorageValidationError{
			Code:        "400031107",
			Description: "Storage name already exists",
		})

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/storages", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_Create_ValidationError_StorageExists(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.POST("/storages", h.Create)

	req := service.CreateStorageRequest{
		StorageName:     "test-storage",
		VendorType:      "OSS",
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		BucketName:      "existing-bucket",
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-secret",
		Region:          "cn-hangzhou",
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*service.CreateStorageRequest")).
		Return("", &service.StorageValidationError{
			Code:        "400031108",
			Description: "Storage already exists",
		})

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/storages", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_Create_ValidationError_InvalidVendorType(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.POST("/storages", h.Create)

	req := service.CreateStorageRequest{
		StorageName:     "test-storage",
		VendorType:      "INVALID",
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		BucketName:      "test-bucket",
		AccessKeyID:     "test-key-id",
		AccessKeySecret: "test-secret",
		Region:          "cn-hangzhou",
	}

	mockService.On("Create", mock.Anything, mock.AnythingOfType("*service.CreateStorageRequest")).
		Return("", &service.StorageValidationError{
			Code:        "400031110",
			Description: "Invalid vendor type",
		})

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/storages", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_Update_Success(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.PUT("/storages/:id", h.Update)

	req := service.UpdateStorageRequest{
		StorageName: "updated-storage",
	}

	mockService.On("Update", mock.Anything, "storage123", mock.AnythingOfType("*service.UpdateStorageRequest")).
		Return(nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/storages/storage123", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_Update_Error(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.PUT("/storages/:id", h.Update)

	req := service.UpdateStorageRequest{
		StorageName: "updated-storage",
	}

	mockService.On("Update", mock.Anything, "storage123", mock.AnythingOfType("*service.UpdateStorageRequest")).
		Return(errors.New("update failed"))

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/storages/storage123", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_Delete_Success(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.DELETE("/storages/:id", h.Delete)

	mockService.On("Delete", mock.Anything, "storage123").Return(nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/storages/storage123", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_Get_Success(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.GET("/storages/:id", h.Get)

	expectedResponse := &service.StorageResponse{
		StorageID:   "storage123",
		StorageName: "test-storage",
		VendorType:  "OSS",
		Endpoint:    "https://oss-cn-hangzhou.aliyuncs.com",
		BucketName:  "test-bucket",
		Region:      "cn-hangzhou",
		IsDefault:   true,
		IsEnabled:   true,
	}

	mockService.On("Get", mock.Anything, "storage123").Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/storages/storage123", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_Get_NotFound(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.GET("/storages/:id", h.Get)

	mockService.On("Get", mock.Anything, "nonexistent").Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/storages/nonexistent", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_List_Success(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.GET("/storages", h.List)

	expectedResponse := &service.ListStorageResponse{
		Count: 2,
		Data: []*service.StorageResponse{
			{
				StorageID:   "storage1",
				StorageName: "storage-1",
				VendorType:  "OSS",
			},
			{
				StorageID:   "storage2",
				StorageName: "storage-2",
				VendorType:  "OBS",
			},
		},
	}

	mockService.On("List", mock.Anything, mock.AnythingOfType("*service.ListStorageRequest")).
		Return(expectedResponse, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/storages?page=1&size=10", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_CheckConnection_Success(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.POST("/storages/:id/check", h.CheckConnection)

	mockService.On("CheckConnection", mock.Anything, "storage123").Return(nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/storages/storage123/check", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestStorageHandler_CheckConnection_Failed(t *testing.T) {
	mockService := new(MockStorageService)
	h := handler.NewStorageHandler(mockService)
	router := setupRouter()
	router.POST("/storages/:id/check", h.CheckConnection)

	mockService.On("CheckConnection", mock.Anything, "storage123").
		Return(errors.New("connection failed"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/storages/storage123/check", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
