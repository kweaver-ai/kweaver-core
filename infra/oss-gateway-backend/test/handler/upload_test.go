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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockURLService 是 URLService 的 mock 实现
type MockURLService struct {
	mock.Mock
}

func (m *MockURLService) GetUploadURL(ctx context.Context, storageID, objectKey, method string, expires int64, useInternal bool) (*adapter.PresignedURL, error) {
	args := m.Called(ctx, storageID, objectKey, method, expires, useInternal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*adapter.PresignedURL), args.Error(1)
}

func (m *MockURLService) InitMultipartUpload(ctx context.Context, storageID, objectKey string, fileSize int64) (*service.InitMultipartResponse, error) {
	args := m.Called(ctx, storageID, objectKey, fileSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.InitMultipartResponse), args.Error(1)
}

func (m *MockURLService) GetUploadPartURLs(ctx context.Context, storageID, objectKey, uploadID string, partIDs []int, expires int64, useInternal bool) (map[int]*adapter.PresignedURL, error) {
	args := m.Called(ctx, storageID, objectKey, uploadID, partIDs, expires, useInternal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int]*adapter.PresignedURL), args.Error(1)
}

func (m *MockURLService) CompleteMultipartUpload(ctx context.Context, storageID, objectKey, uploadID string, parts []adapter.PartInfo) (*adapter.PresignedURL, error) {
	args := m.Called(ctx, storageID, objectKey, uploadID, parts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*adapter.PresignedURL), args.Error(1)
}

func (m *MockURLService) GetDownloadURL(ctx context.Context, storageID, objectKey string, expires int64, saveName string, useInternal bool) (*adapter.PresignedURL, error) {
	args := m.Called(ctx, storageID, objectKey, expires, saveName, useInternal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*adapter.PresignedURL), args.Error(1)
}

func (m *MockURLService) GetHeadURL(ctx context.Context, storageID, objectKey string, expires int64, useInternal bool) (*adapter.PresignedURL, error) {
	args := m.Called(ctx, storageID, objectKey, expires, useInternal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*adapter.PresignedURL), args.Error(1)
}

func (m *MockURLService) BatchGetHeadURL(ctx context.Context, storageID string, keys []string, expires int64, useInternal bool) (map[string]*adapter.PresignedURL, error) {
	args := m.Called(ctx, storageID, keys, expires, useInternal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*adapter.PresignedURL), args.Error(1)
}

func (m *MockURLService) GetDeleteURL(ctx context.Context, storageID, objectKey string, expires int64, useInternal bool) (*adapter.PresignedURL, error) {
	args := m.Called(ctx, storageID, objectKey, expires, useInternal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*adapter.PresignedURL), args.Error(1)
}

func TestUploadHandler_GetUploadURL_Success(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.GET("/upload/:storageId/*key", h.GetUploadURL)

	expectedURL := &adapter.PresignedURL{
		Method:  "PUT",
		URL:     "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?signature=xxx",
		Headers: map[string]string{"Content-Type": "application/octet-stream"},
	}

	mockService.On("GetUploadURL", mock.Anything, "storage123", "test/file.txt", "PUT", int64(0), false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/upload/storage123/test/file.txt", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUploadHandler_GetUploadURL_WithCustomMethod(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.GET("/upload/:storageId/*key", h.GetUploadURL)

	expectedURL := &adapter.PresignedURL{
		Method: "POST",
		URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/",
		FormField: map[string]string{
			"key":       "test/file.txt",
			"policy":    "base64policy",
			"Signature": "signature",
		},
	}

	mockService.On("GetUploadURL", mock.Anything, "storage123", "test/file.txt", "POST", int64(0), false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/upload/storage123/test/file.txt?request_method=POST", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUploadHandler_GetUploadURL_WithExpires(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.GET("/upload/:storageId/*key", h.GetUploadURL)

	expectedURL := &adapter.PresignedURL{
		Method: "PUT",
		URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?signature=xxx",
	}

	mockService.On("GetUploadURL", mock.Anything, "storage123", "test/file.txt", "PUT", int64(7200), false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/upload/storage123/test/file.txt?expires=7200", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUploadHandler_GetUploadURL_InvalidExpires(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.GET("/upload/:storageId/*key", h.GetUploadURL)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/upload/storage123/test/file.txt?expires=invalid", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadHandler_GetUploadURL_InternalRequest(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.GET("/upload/:storageId/*key", h.GetUploadURL)

	expectedURL := &adapter.PresignedURL{
		Method: "PUT",
		URL:    "https://internal-endpoint/test/file.txt?signature=xxx",
	}

	mockService.On("GetUploadURL", mock.Anything, "storage123", "test/file.txt", "PUT", int64(0), true).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/upload/storage123/test/file.txt?internal_request=true", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUploadHandler_InitMultipartUpload_Success(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.GET("/initmultiupload/:storageId/*key", h.InitMultipartUpload)

	expectedResult := &service.InitMultipartResponse{
		UploadID: "upload123",
		PartSize: int64(5242880),
		Key:      "large-file.zip",
	}

	mockService.On("InitMultipartUpload", mock.Anything, "storage123", "large-file.zip", int64(104857600)).
		Return(expectedResult, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/initmultiupload/storage123/large-file.zip?size=104857600", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUploadHandler_InitMultipartUpload_MissingSize(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.GET("/initmultiupload/:storageId/*key", h.InitMultipartUpload)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/initmultiupload/storage123/large-file.zip", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadHandler_InitMultipartUpload_InvalidSize(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.GET("/initmultiupload/:storageId/*key", h.InitMultipartUpload)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/initmultiupload/storage123/large-file.zip?size=invalid", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadHandler_GetUploadPartURLs_Success(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.POST("/uploadpart/:storageId/*key", h.GetUploadPartURLs)

	expectedURLs := map[int]*adapter.PresignedURL{
		1: {
			Method: "PUT",
			URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/large-file.zip?partNumber=1&uploadId=upload123",
		},
		2: {
			Method: "PUT",
			URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/large-file.zip?partNumber=2&uploadId=upload123",
		},
	}

	mockService.On("GetUploadPartURLs", mock.Anything, "storage123", "large-file.zip", "upload123", []int{1, 2}, int64(0), false).
		Return(expectedURLs, nil)

	reqBody := map[string]interface{}{
		"upload_id": "upload123",
		"part_id":   []int{1, 2},
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/uploadpart/storage123/large-file.zip", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUploadHandler_GetUploadPartURLs_MissingFields(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.POST("/uploadpart/:storageId/*key", h.GetUploadPartURLs)

	reqBody := map[string]interface{}{
		"upload_id": "upload123",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/uploadpart/storage123/large-file.zip", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadHandler_CompleteMultipartUpload_Success(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.POST("/completeupload/:storageId/*key", h.CompleteMultipartUpload)

	expectedURL := &adapter.PresignedURL{
		Method: "POST",
		URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/large-file.zip?uploadId=upload123",
		Headers: map[string]string{
			"Content-Type": "application/xml",
		},
		Body: "<CompleteMultipartUpload><Part><PartNumber>1</PartNumber><ETag>\"etag1\"</ETag></Part></CompleteMultipartUpload>",
	}

	mockService.On("CompleteMultipartUpload", mock.Anything, "storage123", "large-file.zip", "upload123", mock.AnythingOfType("[]adapter.PartInfo")).
		Return(expectedURL, nil)

	etagMap := map[string]string{
		"1": "\"etag1\"",
		"2": "\"etag2\"",
	}
	body, _ := json.Marshal(etagMap)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/completeupload/storage123/large-file.zip?upload_id=upload123", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestUploadHandler_CompleteMultipartUpload_MissingUploadID(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.POST("/completeupload/:storageId/*key", h.CompleteMultipartUpload)

	etagMap := map[string]string{
		"1": "\"etag1\"",
	}
	body, _ := json.Marshal(etagMap)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/completeupload/storage123/large-file.zip", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadHandler_CompleteMultipartUpload_ServiceError(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewUploadHandler(mockService)
	router := setupRouter()
	router.POST("/completeupload/:storageId/*key", h.CompleteMultipartUpload)

	mockService.On("CompleteMultipartUpload", mock.Anything, "storage123", "large-file.zip", "upload123", mock.AnythingOfType("[]adapter.PartInfo")).
		Return(nil, errors.New("completion failed"))

	etagMap := map[string]string{
		"1": "\"etag1\"",
	}
	body, _ := json.Marshal(etagMap)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/completeupload/storage123/large-file.zip?upload_id=upload123", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
