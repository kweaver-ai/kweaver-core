package httpserver

import (
	"net"
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
	scalarFaviconPath = "/swagger/favicon.png"
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
	engine.GET(scalarFaviconPath, func(c *gin.Context) {
		c.Data(http.StatusOK, "image/png", apidocs.AgentFactoryFaviconPNG)
	})
}

func renderScalarPage(specURL string) string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Decision Agent API Reference</title>
  <link rel="icon" type="image/png" href="favicon.png" />
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
	docJSON, err := sjson.SetBytes(apidocs.AgentFactoryJSON, "servers", runtimeServers(c))
	if err != nil {
		return apidocs.AgentFactoryJSON
	}

	return docJSON
}

// runtimeServers 为运行时文档构造可编辑的 server 模板，并用当前请求作为默认值。
func runtimeServers(c *gin.Context) []map[string]any {
	scheme := currentRequestScheme(c)
	host, port := currentRequestHostPort(c, scheme)

	return []map[string]any{
		{
			"url":         scheme + "://{host}:{port}/",
			"description": "Current service endpoint (editable)",
			"variables": map[string]any{
				"host": map[string]any{
					"default":     host,
					"description": "Host or IP",
				},
				"port": map[string]any{
					"default":     port,
					"description": "Port",
				},
			},
		},
	}
}

func currentRequestBaseURL(c *gin.Context) string {
	scheme := currentRequestScheme(c)
	host := firstHeaderValue(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = c.Request.Host
	}

	return scheme + "://" + host + "/"
}

// currentRequestScheme 提取当前请求实际对外暴露的协议。
func currentRequestScheme(c *gin.Context) string {
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

	return scheme
}

// currentRequestHostPort 提取当前请求对外主机名和端口，并在缺失端口时补默认值。
func currentRequestHostPort(c *gin.Context, scheme string) (string, string) {
	hostHeader := firstHeaderValue(c.GetHeader("X-Forwarded-Host"))
	if hostHeader == "" {
		hostHeader = c.Request.Host
	}

	return splitHostPort(hostHeader, scheme)
}

// splitHostPort 解析 host:port，并在端口缺失时补协议默认端口。
func splitHostPort(hostport string, scheme string) (string, string) {
	defaultPort := defaultPortForScheme(scheme)
	if hostport == "" {
		return "localhost", defaultPort
	}

	if strings.HasPrefix(hostport, "[") {
		if host, port, err := net.SplitHostPort(hostport); err == nil {
			return host, port
		}
	}

	if strings.Count(hostport, ":") == 1 {
		if host, port, err := net.SplitHostPort(hostport); err == nil {
			return host, port
		}
	}

	return hostport, defaultPort
}

// defaultPortForScheme 返回协议对应的常见默认端口。
func defaultPortForScheme(scheme string) string {
	if strings.EqualFold(scheme, "https") {
		return "443"
	}

	return "80"
}

func firstHeaderValue(value string) string {
	if value == "" {
		return ""
	}

	return strings.TrimSpace(strings.Split(value, ",")[0])
}
