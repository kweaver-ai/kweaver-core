package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"oss-gateway/internal/handler"
	"oss-gateway/pkg/adapter"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteHandler_GetDeleteURL_Success(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDeleteHandler(mockService)
	router := setupRouter()
	router.GET("/delete/:storageId/*key", h.GetDeleteURL)

	expectedURL := &adapter.PresignedURL{
		Method:  "DELETE",
		URL:     "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?signature=xxx",
		Headers: map[string]string{},
	}

	mockService.On("GetDeleteURL", mock.Anything, "storage123", "test/file.txt", int64(0), false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/delete/storage123/test/file.txt", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteHandler_GetDeleteURL_WithExpires(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDeleteHandler(mockService)
	router := setupRouter()
	router.GET("/delete/:storageId/*key", h.GetDeleteURL)

	expectedURL := &adapter.PresignedURL{
		Method: "DELETE",
		URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?signature=xxx&expires=1800",
	}

	mockService.On("GetDeleteURL", mock.Anything, "storage123", "test/file.txt", int64(1800), false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/delete/storage123/test/file.txt?expires=1800", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteHandler_GetDeleteURL_InvalidExpires(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDeleteHandler(mockService)
	router := setupRouter()
	router.GET("/delete/:storageId/*key", h.GetDeleteURL)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/delete/storage123/test/file.txt?expires=invalid", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteHandler_GetDeleteURL_InternalRequest(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDeleteHandler(mockService)
	router := setupRouter()
	router.GET("/delete/:storageId/*key", h.GetDeleteURL)

	expectedURL := &adapter.PresignedURL{
		Method: "DELETE",
		URL:    "https://internal-endpoint/test/file.txt?signature=xxx",
	}

	mockService.On("GetDeleteURL", mock.Anything, "storage123", "test/file.txt", int64(0), true).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/delete/storage123/test/file.txt?internal_request=true", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteHandler_GetDeleteURL_ServiceError(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDeleteHandler(mockService)
	router := setupRouter()
	router.GET("/delete/:storageId/*key", h.GetDeleteURL)

	mockService.On("GetDeleteURL", mock.Anything, "storage123", "test/file.txt", int64(0), false).
		Return(nil, errors.New("service error"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/delete/storage123/test/file.txt", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestDeleteHandler_GetDeleteURL_URLEncoding(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDeleteHandler(mockService)
	router := setupRouter()
	router.GET("/delete/:storageId/*key", h.GetDeleteURL)

	expectedURL := &adapter.PresignedURL{
		Method: "DELETE",
		URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file%20with%20spaces.txt",
	}

	mockService.On("GetDeleteURL", mock.Anything, "storage123", "test/file with spaces.txt", int64(0), false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/delete/storage123/test/file%20with%20spaces.txt", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
