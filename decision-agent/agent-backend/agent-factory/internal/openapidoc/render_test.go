package openapidoc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderScalarStaticHTMLEmbedsSpec(t *testing.T) {
	t.Parallel()

	html := RenderScalarStaticHTML([]byte(`{"openapi":"3.0.2","info":{"title":"Agent Factory API","version":"1.0.0"},"paths":{}}`))

	require.Contains(t, html, "ui/scalar-api-reference.js")
	require.Contains(t, html, "openapi-document")
	require.Contains(t, html, "URL.createObjectURL")
	require.Contains(t, html, "Decision Agent API Reference")
	require.Contains(t, html, `rel="icon"`)
	require.Contains(t, html, "favicon.png")
	require.Contains(t, html, "agent-factory-redoc.html")
	require.NotContains(t, html, "cdn.jsdelivr.net")
	require.NotContains(t, html, "cdn.redocly.com")
	require.NotContains(t, html, "fonts.googleapis.com")
}

func TestRenderRedocStaticHTMLEmbedsSpec(t *testing.T) {
	t.Parallel()

	html := RenderRedocStaticHTML([]byte(`{"openapi":"3.0.2","info":{"title":"Agent Factory API","version":"1.0.0"},"paths":{}}`))

	require.Contains(t, html, "ui/redoc.standalone.js")
	require.Contains(t, html, "openapi-document")
	require.Contains(t, html, "JSON.parse")
	require.Contains(t, html, "Decision Agent API Reference")
	require.Contains(t, html, `rel="icon"`)
	require.Contains(t, html, "favicon.png")
	require.Contains(t, html, "agent-factory.html")
	require.NotContains(t, html, "cdn.jsdelivr.net")
	require.NotContains(t, html, "cdn.redocly.com")
	require.NotContains(t, html, "fonts.googleapis.com")
}
