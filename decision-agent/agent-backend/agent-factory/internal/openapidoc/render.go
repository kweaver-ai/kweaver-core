package openapidoc

import "strings"

// RenderScalarStaticHTML 生成内嵌 OpenAPI JSON 的静态 Scalar HTML 页面。
func RenderScalarStaticHTML(specJSON []byte) string {
	escapedSpec := strings.ReplaceAll(string(specJSON), "</script", `<\/script`)

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
  <script type="application/json" id="openapi-document">` + escapedSpec + `</script>
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
  <script src="` + scalarCDNURL + `"></script>
</body>
</html>
`
}
