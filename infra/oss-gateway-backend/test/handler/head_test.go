package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"oss-gateway/internal/handler"
	"oss-gateway/pkg/adapter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHeadHandler_GetHeadURL_Success(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewHeadHandler(mockService)
	router := setupRouter()
	router.GET("/head/:storageId/*key", h.GetHeadURL)

	expectedURL := &adapter.PresignedURL{
		Method:  "HEAD",
		URL:     "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?signature=xxx",
		Headers: map[string]string{},
	}

	mockService.On("GetHeadURL", mock.Anything, "storage123", "test/file.txt", int64(0), false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/head/storage123/test/file.txt", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHeadHandler_GetHeadURL_WithExpires(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewHeadHandler(mockService)
	router := setupRouter()
	router.GET("/head/:storageId/*key", h.GetHeadURL)

	expectedURL := &adapter.PresignedURL{
		Method: "HEAD",
		URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?signature=xxx",
	}

	mockService.On("GetHeadURL", mock.Anything, "storage123", "test/file.txt", int64(3600), false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/head/storage123/test/file.txt?expires=3600", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHeadHandler_GetHeadURL_InvalidExpires(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewHeadHandler(mockService)
	router := setupRouter()
	router.GET("/head/:storageId/*key", h.GetHeadURL)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/head/storage123/test/file.txt?expires=invalid", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHeadHandler_GetHeadURL_ServiceError(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewHeadHandler(mockService)
	router := setupRouter()
	router.GET("/head/:storageId/*key", h.GetHeadURL)

	mockService.On("GetHeadURL", mock.Anything, "storage123", "test/file.txt", int64(0), false).
		Return(nil, errors.New("service error"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/head/storage123/test/file.txt", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestHeadHandler_BatchGetHeadURL_Success(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewHeadHandler(mockService)
	router := setupRouter()
	router.POST("/head/:storageId", h.BatchGetHeadURL)

	expectedURLs := map[string]*adapter.PresignedURL{
		"file1.txt": {
			Method: "HEAD",
			URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/file1.txt?signature=xxx",
		},
		"file2.txt": {
			Method: "HEAD",
			URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/file2.txt?signature=xxx",
		},
	}

	mockService.On("BatchGetHeadURL", mock.Anything, "storage123", []string{"file1.txt", "file2.txt"}, int64(0), false).
		Return(expectedURLs, nil)

	reqBody := map[string]interface{}{
		"keys": []string{"file1.txt", "file2.txt"},
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/head/storage123", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHeadHandler_BatchGetHeadURL_MissingKeys(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewHeadHandler(mockService)
	router := setupRouter()
	router.POST("/head/:storageId", h.BatchGetHeadURL)

	reqBody := map[string]interface{}{}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/head/storage123", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHeadHandler_BatchGetHeadURL_TooManyKeys(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewHeadHandler(mockService)
	router := setupRouter()
	router.POST("/head/:storageId", h.BatchGetHeadURL)

	// Create 101 keys
	keys := make([]string, 101)
	for i := 0; i < 101; i++ {
		keys[i] = "file.txt"
	}

	reqBody := map[string]interface{}{
		"keys": keys,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/head/storage123", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHeadHandler_BatchGetHeadURL_ServiceError(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewHeadHandler(mockService)
	router := setupRouter()
	router.POST("/head/:storageId", h.BatchGetHeadURL)

	mockService.On("BatchGetHeadURL", mock.Anything, "storage123", []string{"file1.txt"}, int64(0), false).
		Return(nil, errors.New("service error"))

	reqBody := map[string]interface{}{
		"keys": []string{"file1.txt"},
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/head/storage123", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}
