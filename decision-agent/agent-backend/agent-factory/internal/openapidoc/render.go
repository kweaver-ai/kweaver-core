package openapidoc

import (
	"fmt"
	"strings"
)

// RenderScalarStaticHTML 生成内嵌 OpenAPI JSON 的静态 Scalar HTML 页面。
func RenderScalarStaticHTML(specJSON []byte) string {
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
`, staticDocsPageStyle(), staticDocsNavHTML("scalar"), escapedSpec, staticScalarJSPath)
}

// RenderRedocStaticHTML 生成内嵌 OpenAPI JSON 的静态 Redoc HTML 页面。
func RenderRedocStaticHTML(specJSON []byte) string {
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
  <script type="application/json" id="openapi-document">%s</script>
  <div id="redoc-container"></div>
  <script src="%s"></script>
  <script>
    (() => {
      const spec = JSON.parse(document.getElementById("openapi-document").textContent);
      const container = document.getElementById("redoc-container");
      Redoc.init(spec, {
        hideDownloadButton: false,
        scrollYOffset: 64,
      }, container);
    })();
  </script>
</body>
</html>
`, staticDocsPageStyle(), staticDocsNavHTML("redoc"), escapedSpec, staticRedocJSPath)
}

func escapeEmbeddedSpec(specJSON []byte) string {
	return strings.ReplaceAll(string(specJSON), "</script", `<\/script`)
}

func staticDocsNavHTML(active string) string {
	scalarAttrs := `href="` + staticScalarPagePath + `"`
	redocAttrs := `href="` + staticRedocPagePath + `"`

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

func staticDocsPageStyle() string {
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
