package httpserver

import (
	"encoding/json"
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
	assert.Contains(t, indexResp.Body.String(), `rel="icon"`)
	assert.Contains(t, indexResp.Body.String(), "favicon.png")

	jsonReq := httptest.NewRequest(http.MethodGet, scalarDocJSONPath, nil)
	jsonReq.Header.Set("X-Forwarded-Proto", "https")
	jsonReq.Header.Set("X-Forwarded-Host", "docs.example.com")

	jsonResp := httptest.NewRecorder()
	engine.ServeHTTP(jsonResp, jsonReq)
	assert.Equal(t, http.StatusOK, jsonResp.Code)
	assert.Contains(t, jsonResp.Body.String(), "\"openapi\"")

	var doc map[string]any

	assert.NoError(t, json.Unmarshal(jsonResp.Body.Bytes(), &doc))

	servers, ok := doc["servers"].([]any)
	if assert.True(t, ok) && assert.Len(t, servers, 1) {
		serverItem, ok := servers[0].(map[string]any)
		if assert.True(t, ok) {
			assert.Equal(t, "https://{host}:{port}/", serverItem["url"])
			assert.Equal(t, "Current service endpoint (editable)", serverItem["description"])

			variables, ok := serverItem["variables"].(map[string]any)
			if assert.True(t, ok) {
				hostVar, ok := variables["host"].(map[string]any)
				if assert.True(t, ok) {
					assert.Equal(t, "docs.example.com", hostVar["default"])
				}

				portVar, ok := variables["port"].(map[string]any)
				if assert.True(t, ok) {
					assert.Equal(t, "443", portVar["default"])
				}
			}
		}
	}

	security, ok := doc["security"].([]any)
	if assert.True(t, ok) && assert.Len(t, security, 1) {
		requirement, ok := security[0].(map[string]any)
		if assert.True(t, ok) {
			scopes, ok := requirement["ApiKeyAuth"].([]any)
			assert.True(t, ok)
			assert.Empty(t, scopes)
		}
	}

	yamlReq := httptest.NewRequest(http.MethodGet, scalarDocYAMLPath, nil)
	yamlResp := httptest.NewRecorder()
	engine.ServeHTTP(yamlResp, yamlReq)
	assert.Equal(t, http.StatusOK, yamlResp.Code)
	assert.Contains(t, yamlResp.Body.String(), "openapi: 3.0.2")

	faviconReq := httptest.NewRequest(http.MethodGet, scalarFaviconPath, nil)
	faviconResp := httptest.NewRecorder()
	engine.ServeHTTP(faviconResp, faviconReq)
	assert.Equal(t, http.StatusOK, faviconResp.Code)
	assert.Equal(t, "image/png", faviconResp.Header().Get("Content-Type"))
	assert.NotEmpty(t, faviconResp.Body.Bytes())
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

func TestCurrentRequestHostPort_UsesForwardedHeadersAndDefaultPorts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("forwarded host without port falls back to https default", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req := httptest.NewRequest(http.MethodGet, "http://internal/swagger/doc.json", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "docs.example.com")
		ctx.Request = req

		host, port := currentRequestHostPort(ctx, currentRequestScheme(ctx))
		assert.Equal(t, "docs.example.com", host)
		assert.Equal(t, "443", port)
	})

	t.Run("request host with explicit port is preserved", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req := httptest.NewRequest(http.MethodGet, "http://internal/swagger/doc.json", nil)
		req.Host = "localhost:13020"
		ctx.Request = req

		host, port := currentRequestHostPort(ctx, currentRequestScheme(ctx))
		assert.Equal(t, "localhost", host)
		assert.Equal(t, "13020", port)
	})
}
