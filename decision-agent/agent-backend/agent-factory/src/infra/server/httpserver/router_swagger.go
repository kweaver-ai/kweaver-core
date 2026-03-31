package httpserver

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apidocs "github.com/kweaver-ai/decision-agent/agent-factory/docs/api"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/tidwall/sjson"
)

const (
	scalarDocsPath    = "/swagger/index.html"
	scalarDocJSONPath = "/swagger/doc.json"
	scalarDocYAMLPath = "/swagger/doc.yaml"
	scalarCDNURL      = "https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.34.6"
)

// registerSwaggerRoutes 注册 API 文档路由
func (s *httpServer) registerSwaggerRoutes(engine *gin.Engine) {
	if global.GConfig == nil || !global.GConfig.EnableSwagger {
		return
	}

	engine.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, scalarDocsPath)
	})
	engine.GET("/swagger/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, scalarDocsPath)
	})
	engine.GET("/scalar", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, scalarDocsPath)
	})
	engine.GET("/scalar/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, scalarDocsPath)
	})
	engine.GET(scalarDocsPath, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(renderScalarPage(scalarDocJSONPath)))
	})
	engine.GET(scalarDocJSONPath, func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", renderOpenAPIDocJSON(c))
	})
	engine.GET(scalarDocYAMLPath, func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", apidocs.AgentFactoryYAML)
	})
}

func renderScalarPage(specURL string) string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Agent Factory API Reference</title>
  <style>
    body {
      margin: 0;
      padding: 0;
    }
  </style>
</head>
<body>
  <noscript>Scalar requires JavaScript to render the API reference.</noscript>
  <script id="api-reference" data-url="` + specURL + `"></script>
  <script src="` + scalarCDNURL + `"></script>
</body>
</html>`
}

func renderOpenAPIDocJSON(c *gin.Context) []byte {
	docJSON, err := sjson.SetBytes(apidocs.AgentFactoryJSON, "servers", []map[string]string{
		{
			"url":         currentRequestBaseURL(c),
			"description": "Current service endpoint",
		},
	})
	if err != nil {
		return apidocs.AgentFactoryJSON
	}

	return docJSON
}

func currentRequestBaseURL(c *gin.Context) string {
	scheme := firstHeaderValue(c.GetHeader("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = firstHeaderValue(c.GetHeader("X-Forwarded-Scheme"))
	}

	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	host := firstHeaderValue(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = c.Request.Host
	}

	return scheme + "://" + host + "/"
}

func firstHeaderValue(value string) string {
	if value == "" {
		return ""
	}

	return strings.TrimSpace(strings.Split(value, ",")[0])
}
