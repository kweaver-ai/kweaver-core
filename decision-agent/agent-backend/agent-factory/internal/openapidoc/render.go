package openapidoc

import (
	"fmt"
	"strings"
)

func renderScalarStaticHTML(specJSON []byte, scalarJSRef string) string {
	escapedSpec := escapeEmbeddedSpec(specJSON)

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
  <script>
    %s
  </script>
  <script>
    %s
  </script>
  <script type="application/json" id="openapi-document">%s</script>
  <script>
    (() => {
      const specElement = document.getElementById("openapi-document");
      const referenceElement = document.createElement("script");
      const specBlob = new Blob([specElement.textContent], { type: "application/json" });
      const specURL = URL.createObjectURL(specBlob);
      referenceElement.id = "api-reference";
      referenceElement.dataset.url = specURL;
      document.body.appendChild(referenceElement);
      window.addEventListener("pagehide", () => URL.revokeObjectURL(specURL), { once: true });
    })();
  </script>
  <script src="%s"></script>
</body>
</html>
`, DocsPageStyle(), DocsNavHTML("scalar", staticScalarPagePath, staticRedocPagePath), DocsBootstrapScript(), ScalarPageEnhancementScript(), escapedSpec, scalarJSRef)
}

func renderRedocStaticHTML(specJSON []byte, redocJSRef string) string {
	escapedSpec := escapeEmbeddedSpec(specJSON)

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
  <noscript>Redoc requires JavaScript to render the API reference.</noscript>
  <script>
    %s
  </script>
  <script type="application/json" id="openapi-document">%s</script>
  <div id="redoc-container"></div>
  <script src="%s"></script>
  <script>
    %s
  </script>
</body>
</html>
`, DocsPageStyle(), DocsNavHTML("redoc", staticScalarPagePath, staticRedocPagePath), DocsBootstrapScript(), escapedSpec, redocJSRef, RedocInitScript(`JSON.parse(document.getElementById("openapi-document").textContent)`))
}

// RenderPublicScalarStaticHTML 生成对外分发用的 Scalar HTML 页面，前端资源走国内 CDN。
func RenderPublicScalarStaticHTML(specJSON []byte) string {
	return renderScalarStaticHTML(specJSON, publicScalarJSURL)
}

// RenderRuntimeScalarStaticHTML 生成服务运行时使用的 Scalar HTML 页面，前端资源走本地 embed。
func RenderRuntimeScalarStaticHTML(specJSON []byte) string {
	return renderScalarStaticHTML(specJSON, runtimeScalarJSPath)
}

// RenderScalarStaticHTML 为兼容旧调用保留，默认返回运行时版本。
func RenderScalarStaticHTML(specJSON []byte) string {
	return RenderRuntimeScalarStaticHTML(specJSON)
}

// RenderPublicRedocStaticHTML 生成对外分发用的 Redoc HTML 页面，前端资源走国内 CDN。
func RenderPublicRedocStaticHTML(specJSON []byte) string {
	return renderRedocStaticHTML(specJSON, publicRedocJSURL)
}

// RenderRuntimeRedocStaticHTML 生成服务运行时使用的 Redoc HTML 页面，前端资源走本地 embed。
func RenderRuntimeRedocStaticHTML(specJSON []byte) string {
	return renderRedocStaticHTML(specJSON, runtimeRedocJSPath)
}

// RenderRedocStaticHTML 为兼容旧调用保留，默认返回运行时版本。
func RenderRedocStaticHTML(specJSON []byte) string {
	return RenderRuntimeRedocStaticHTML(specJSON)
}

func escapeEmbeddedSpec(specJSON []byte) string {
	return strings.ReplaceAll(string(specJSON), "</script", `<\/script`)
}
