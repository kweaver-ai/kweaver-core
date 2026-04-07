package worker

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"go.uber.org/mock/gomock"
)

func newRESTHandler() RESTHandler {
	return &restHandler{}
}

func setGinMode() func() {
	old := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(old)
	}
}

func TestWorker(t *testing.T) {
	test := setGinMode()
	defer test()
	engine := gin.New()
	engine.Use(gin.Recovery())
	group := engine.Group("/api/automation/v1")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCRestHandler := newRESTHandler()
	testCRestHandler.RegisterAPI(group)

	req := httptest.NewRequest("GET", "/api/automation/v1/worker", http.NoBody)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	result := w.Result()

	assert.Equal(t, result.StatusCode, http.StatusOK)
	if err := result.Body.Close(); err != nil {
		assert.Equal(t, err, nil)
	}
}
