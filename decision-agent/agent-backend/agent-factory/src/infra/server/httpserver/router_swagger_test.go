package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
)

func TestRegisterSwaggerRoutes_ScalarDocsEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origConfig := global.GConfig
	global.GConfig = &conf.Config{EnableSwagger: true}

	t.Cleanup(func() {
		global.GConfig = origConfig
	})

	engine := gin.New()
	server := &httpServer{}

	server.registerSwaggerRoutes(engine)

	redirectReq := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	redirectResp := httptest.NewRecorder()
	engine.ServeHTTP(redirectResp, redirectReq)
	assert.Equal(t, http.StatusMovedPermanently, redirectResp.Code)
	assert.Equal(t, scalarDocsPath, redirectResp.Header().Get("Location"))

	indexReq := httptest.NewRequest(http.MethodGet, scalarDocsPath, nil)
	indexResp := httptest.NewRecorder()
	engine.ServeHTTP(indexResp, indexReq)
	assert.Equal(t, http.StatusOK, indexResp.Code)
	assert.Contains(t, indexResp.Body.String(), "@scalar/api-reference")
	assert.Contains(t, indexResp.Body.String(), scalarDocJSONPath)

	jsonReq := httptest.NewRequest(http.MethodGet, scalarDocJSONPath, nil)
	jsonReq.Header.Set("X-Forwarded-Proto", "https")
	jsonReq.Header.Set("X-Forwarded-Host", "docs.example.com")

	jsonResp := httptest.NewRecorder()
	engine.ServeHTTP(jsonResp, jsonReq)
	assert.Equal(t, http.StatusOK, jsonResp.Code)
	assert.Contains(t, jsonResp.Body.String(), "\"openapi\"")
	assert.Contains(t, jsonResp.Body.String(), "https://docs.example.com/")

	yamlReq := httptest.NewRequest(http.MethodGet, scalarDocYAMLPath, nil)
	yamlResp := httptest.NewRecorder()
	engine.ServeHTTP(yamlResp, yamlReq)
	assert.Equal(t, http.StatusOK, yamlResp.Code)
	assert.Contains(t, yamlResp.Body.String(), "openapi: 3.0.2")
}

func TestRegisterSwaggerRoutes_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origConfig := global.GConfig
	global.GConfig = &conf.Config{EnableSwagger: false}

	t.Cleanup(func() {
		global.GConfig = origConfig
	})

	engine := gin.New()
	server := &httpServer{}

	server.registerSwaggerRoutes(engine)

	req := httptest.NewRequest(http.MethodGet, scalarDocsPath, nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestRegisterSwaggerRoutes_NilConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	origConfig := global.GConfig
	global.GConfig = nil

	t.Cleanup(func() {
		global.GConfig = origConfig
	})

	engine := gin.New()
	server := &httpServer{}

	server.registerSwaggerRoutes(engine)

	req := httptest.NewRequest(http.MethodGet, scalarDocsPath, nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestCurrentRequestBaseURL_UsesForwardedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "http://internal:13020/swagger/doc.json", nil)
	req.Header.Set("X-Forwarded-Proto", "https, http")
	req.Header.Set("X-Forwarded-Host", "api.example.com, internal:13020")
	ctx.Request = req

	assert.Equal(t, "https://api.example.com/", currentRequestBaseURL(ctx))
}
