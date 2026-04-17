package httpserver

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	apidocs "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/server/apidocs"
	"github.com/tidwall/sjson"
)

const (
	scalarDocsPath    = "/swagger/index.html"
	redocDocsPath     = "/redoc/index.html"
	scalarDocJSONPath = "/swagger/doc.json"
	scalarDocYAMLPath = "/swagger/doc.yaml"
	scalarFaviconPath = "/swagger/favicon.png"
	apidocsUIPath     = "/apidocs-ui"
	scalarJSAssetPath = apidocsUIPath + "/scalar-api-reference.js"
	redocJSAssetPath  = apidocsUIPath + "/redoc.standalone.js"
)

// registerSwaggerRoutes 注册 API 文档路由
func (s *httpServer) registerSwaggerRoutes(engine *gin.Engine) {
	if global.GConfig == nil || !global.GConfig.EnableSwagger {
		return
	}

	engine.StaticFS(apidocsUIPath, apidocs.UIAssetsFileSystem())

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
	engine.GET("/redoc", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, redocDocsPath)
	})
	engine.GET("/redoc/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, redocDocsPath)
	})
	engine.GET(scalarDocsPath, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(renderScalarPage(scalarDocJSONPath)))
	})
	engine.GET(redocDocsPath, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(renderRedocPage(scalarDocJSONPath)))
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
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Decision Agent API Reference</title>
  <link rel="icon" type="image/png" href="favicon.png" />
  <style>
    %s
  </style>
</head>
<body>
  %s
  <noscript>Scalar requires JavaScript to render the API reference.</noscript>
  <script id="api-reference" data-url="%s"></script>
  <script src="%s"></script>
</body>
</html>`, docsPageStyle(), runtimeDocsNavHTML("scalar"), specURL, scalarJSAssetPath)
}

func renderRedocPage(specURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Decision Agent API Reference</title>
  <link rel="icon" type="image/png" href="%s" />
  <style>
    %s
  </style>
</head>
<body>
  %s
  <noscript>Redoc requires JavaScript to render the API reference.</noscript>
  <div id="redoc-container"></div>
  <script src="%s"></script>
  <script>
    (() => {
      const container = document.getElementById("redoc-container");
      Redoc.init("%s", {
        hideDownloadButton: false,
        scrollYOffset: 64,
      }, container);
    })();
  </script>
</body>
</html>`, scalarFaviconPath, docsPageStyle(), runtimeDocsNavHTML("redoc"), redocJSAssetPath, specURL)
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

func runtimeDocsNavHTML(active string) string {
	scalarAttrs := `href="` + scalarDocsPath + `"`
	redocAttrs := `href="` + redocDocsPath + `"`

	if active == "scalar" {
		scalarAttrs += ` aria-current="page"`
	}
	if active == "redoc" {
		redocAttrs += ` aria-current="page"`
	}

	return `<header class="docs-nav">
  <div class="docs-nav__title">Decision Agent API Reference</div>
  <nav class="docs-nav__links">
    <a ` + scalarAttrs + `>Scalar</a>
    <a ` + redocAttrs + `>Redoc</a>
  </nav>
</header>`
}

func docsPageStyle() string {
	return `:root {
      color-scheme: light;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    }

    * {
      box-sizing: border-box;
    }

    body {
      margin: 0;
      padding: 0;
      background: #f5f7fb;
    }

    .docs-nav {
      position: sticky;
      top: 0;
      z-index: 10;
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
      padding: 16px 24px;
      background: rgba(15, 23, 42, 0.94);
      color: #f8fafc;
      backdrop-filter: blur(12px);
    }

    .docs-nav__title {
      font-size: 15px;
      font-weight: 600;
      letter-spacing: 0.01em;
    }

    .docs-nav__links {
      display: flex;
      gap: 12px;
      flex-wrap: wrap;
    }

    .docs-nav__links a {
      color: #cbd5e1;
      text-decoration: none;
      padding: 8px 12px;
      border-radius: 999px;
      border: 1px solid rgba(148, 163, 184, 0.32);
      transition: background-color 0.2s ease, color 0.2s ease, border-color 0.2s ease;
    }

    .docs-nav__links a[aria-current="page"] {
      color: #0f172a;
      background: #f8fafc;
      border-color: #f8fafc;
    }

    #redoc-container {
      min-height: calc(100vh - 64px);
    }

    @media (max-width: 720px) {
      .docs-nav {
        padding: 14px 16px;
        align-items: flex-start;
        flex-direction: column;
      }
    }`
}
