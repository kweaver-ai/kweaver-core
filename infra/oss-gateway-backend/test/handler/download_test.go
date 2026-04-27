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

func TestDownloadHandler_GetDownloadURL_Success(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDownloadHandler(mockService)
	router := setupRouter()
	router.GET("/download/:storageId/*key", h.GetDownloadURL)

	expectedURL := &adapter.PresignedURL{
		Method:  "GET",
		URL:     "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?signature=xxx",
		Headers: map[string]string{},
	}

	mockService.On("GetDownloadURL", mock.Anything, "storage123", "test/file.txt", int64(0), "", false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/download/storage123/test/file.txt", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDownloadHandler_GetDownloadURL_WithExpires(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDownloadHandler(mockService)
	router := setupRouter()
	router.GET("/download/:storageId/*key", h.GetDownloadURL)

	expectedURL := &adapter.PresignedURL{
		Method: "GET",
		URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?signature=xxx&expires=7200",
	}

	mockService.On("GetDownloadURL", mock.Anything, "storage123", "test/file.txt", int64(7200), "", false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/download/storage123/test/file.txt?expires=7200", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDownloadHandler_GetDownloadURL_WithSaveName(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDownloadHandler(mockService)
	router := setupRouter()
	router.GET("/download/:storageId/*key", h.GetDownloadURL)

	expectedURL := &adapter.PresignedURL{
		Method: "GET",
		URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file.txt?signature=xxx&save_name=download.txt",
	}

	mockService.On("GetDownloadURL", mock.Anything, "storage123", "test/file.txt", int64(0), "download.txt", false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/download/storage123/test/file.txt?save_name=download.txt", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDownloadHandler_GetDownloadURL_InvalidExpires(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDownloadHandler(mockService)
	router := setupRouter()
	router.GET("/download/:storageId/*key", h.GetDownloadURL)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/download/storage123/test/file.txt?expires=invalid", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDownloadHandler_GetDownloadURL_InternalRequest(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDownloadHandler(mockService)
	router := setupRouter()
	router.GET("/download/:storageId/*key", h.GetDownloadURL)

	expectedURL := &adapter.PresignedURL{
		Method: "GET",
		URL:    "https://internal-endpoint/test/file.txt?signature=xxx",
	}

	mockService.On("GetDownloadURL", mock.Anything, "storage123", "test/file.txt", int64(0), "", true).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/download/storage123/test/file.txt?internal_request=true", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestDownloadHandler_GetDownloadURL_ServiceError(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDownloadHandler(mockService)
	router := setupRouter()
	router.GET("/download/:storageId/*key", h.GetDownloadURL)

	mockService.On("GetDownloadURL", mock.Anything, "storage123", "test/file.txt", int64(0), "", false).
		Return(nil, errors.New("service error"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/download/storage123/test/file.txt", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestDownloadHandler_GetDownloadURL_URLEncoding(t *testing.T) {
	mockService := new(MockURLService)
	h := handler.NewDownloadHandler(mockService)
	router := setupRouter()
	router.GET("/download/:storageId/*key", h.GetDownloadURL)

	expectedURL := &adapter.PresignedURL{
		Method: "GET",
		URL:    "https://test-bucket.oss-cn-hangzhou.aliyuncs.com/test/file%20with%20spaces.txt",
	}

	mockService.On("GetDownloadURL", mock.Anything, "storage123", "test/file with spaces.txt", int64(0), "", false).
		Return(expectedURL, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/download/storage123/test/file%20with%20spaces.txt", nil)

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
