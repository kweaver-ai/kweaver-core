package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
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

	legacyRedirectReq := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	legacyRedirectResp := httptest.NewRecorder()
	engine.ServeHTTP(legacyRedirectResp, legacyRedirectReq)
	assert.Equal(t, http.StatusNotFound, legacyRedirectResp.Code)

	indexReq := httptest.NewRequest(http.MethodGet, scalarDocsPath, nil)
	indexResp := httptest.NewRecorder()
	engine.ServeHTTP(indexResp, indexReq)
	assert.Equal(t, http.StatusOK, indexResp.Code)
	assert.Contains(t, indexResp.Body.String(), scalarJSAssetPath)
	assert.Contains(t, indexResp.Body.String(), scalarDocJSONPath)
	assert.Contains(t, indexResp.Body.String(), `rel="icon"`)
	assert.Contains(t, indexResp.Body.String(), scalarFaviconPath)
	assert.Contains(t, indexResp.Body.String(), redocDocsPath)
	assert.Contains(t, indexResp.Body.String(), "promoteScalarModelsGroup")
	assert.Contains(t, indexResp.Body.String(), "syncDocsNavHeight")
	assert.Contains(t, indexResp.Body.String(), "--docs-nav-height")
	assert.NotContains(t, indexResp.Body.String(), "cdn.jsdelivr.net")
	assert.NotContains(t, indexResp.Body.String(), "cdn.redocly.com")
	assert.NotContains(t, indexResp.Body.String(), "/swagger/index.html")

	trailingScalarReq := httptest.NewRequest(http.MethodGet, scalarDocsPath+"/", nil)
	trailingScalarResp := httptest.NewRecorder()
	engine.ServeHTTP(trailingScalarResp, trailingScalarReq)
	assert.Equal(t, http.StatusMovedPermanently, trailingScalarResp.Code)
	assert.Equal(t, scalarDocsPath, trailingScalarResp.Header().Get("Location"))

	redocIndexReq := httptest.NewRequest(http.MethodGet, redocDocsPath, nil)
	redocIndexResp := httptest.NewRecorder()
	engine.ServeHTTP(redocIndexResp, redocIndexReq)
	assert.Equal(t, http.StatusOK, redocIndexResp.Code)
	assert.Contains(t, redocIndexResp.Body.String(), redocJSAssetPath)
	assert.Contains(t, redocIndexResp.Body.String(), scalarDocsPath)
	assert.Contains(t, redocIndexResp.Body.String(), scalarDocJSONPath)
	assert.Contains(t, redocIndexResp.Body.String(), "syncDocsNavHeight")
	assert.Contains(t, redocIndexResp.Body.String(), "--docs-nav-height")
	assert.NotContains(t, redocIndexResp.Body.String(), "cdn.redocly.com")
	assert.NotContains(t, redocIndexResp.Body.String(), "fonts.googleapis.com")

	trailingRedocReq := httptest.NewRequest(http.MethodGet, redocDocsPath+"/", nil)
	trailingRedocResp := httptest.NewRecorder()
	engine.ServeHTTP(trailingRedocResp, trailingRedocReq)
	assert.Equal(t, http.StatusMovedPermanently, trailingRedocResp.Code)
	assert.Equal(t, redocDocsPath, trailingRedocResp.Header().Get("Location"))

	legacyIndexReq := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	legacyIndexResp := httptest.NewRecorder()
	engine.ServeHTTP(legacyIndexResp, legacyIndexReq)
	assert.Equal(t, http.StatusNotFound, legacyIndexResp.Code)

	legacyJSONReq := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	legacyJSONResp := httptest.NewRecorder()
	engine.ServeHTTP(legacyJSONResp, legacyJSONReq)
	assert.Equal(t, http.StatusNotFound, legacyJSONResp.Code)

	legacyYAMLReq := httptest.NewRequest(http.MethodGet, "/swagger/doc.yaml", nil)
	legacyYAMLResp := httptest.NewRecorder()
	engine.ServeHTTP(legacyYAMLResp, legacyYAMLReq)
	assert.Equal(t, http.StatusNotFound, legacyYAMLResp.Code)

	legacyFaviconReq := httptest.NewRequest(http.MethodGet, "/swagger/favicon.png", nil)
	legacyFaviconResp := httptest.NewRecorder()
	engine.ServeHTTP(legacyFaviconResp, legacyFaviconReq)
	assert.Equal(t, http.StatusNotFound, legacyFaviconResp.Code)

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

	uiReq := httptest.NewRequest(http.MethodGet, scalarJSAssetPath, nil)
	uiResp := httptest.NewRecorder()
	engine.ServeHTTP(uiResp, uiReq)
	assert.Equal(t, http.StatusOK, uiResp.Code)
	assert.NotEmpty(t, uiResp.Body.Bytes())
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

	redocReq := httptest.NewRequest(http.MethodGet, redocDocsPath, nil)
	redocResp := httptest.NewRecorder()
	engine.ServeHTTP(redocResp, redocReq)
	assert.Equal(t, http.StatusNotFound, redocResp.Code)
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

	redocReq := httptest.NewRequest(http.MethodGet, redocDocsPath, nil)
	redocResp := httptest.NewRecorder()
	engine.ServeHTTP(redocResp, redocReq)
	assert.Equal(t, http.StatusNotFound, redocResp.Code)
}

func TestCurrentRequestBaseURL_UsesForwardedHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "http://internal:13020/scalar/doc.json", nil)
	req.Header.Set("X-Forwarded-Proto", "https, http")
	req.Header.Set("X-Forwarded-Host", "api.example.com, internal:13020")
	ctx.Request = req

	assert.Equal(t, "https://api.example.com/", currentRequestBaseURL(ctx))
}

func TestCurrentRequestHostPort_UsesForwardedHeadersAndDefaultPorts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("forwarded host without port falls back to https default", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req := httptest.NewRequest(http.MethodGet, "http://internal/scalar/doc.json", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "docs.example.com")
		ctx.Request = req

		host, port := currentRequestHostPort(ctx, currentRequestScheme(ctx))
		assert.Equal(t, "docs.example.com", host)
		assert.Equal(t, "443", port)
	})

	t.Run("request host with explicit port is preserved", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req := httptest.NewRequest(http.MethodGet, "http://internal/scalar/doc.json", nil)
		req.Host = "localhost:13020"
		ctx.Request = req

		host, port := currentRequestHostPort(ctx, currentRequestScheme(ctx))
		assert.Equal(t, "localhost", host)
		assert.Equal(t, "13020", port)
	})
}
