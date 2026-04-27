package openapidoc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderScalarStaticHTMLEmbedsSpec(t *testing.T) {
	t.Parallel()

	html := RenderPublicScalarStaticHTML([]byte(`{"openapi":"3.0.2","info":{"title":"Agent Factory API","version":"1.0.0"},"paths":{}}`))

	require.Contains(t, html, "cdn.jsdmirror.com/npm/@scalar/api-reference@1.34.6/dist/browser/standalone.js")
	require.Contains(t, html, "openapi-document")
	require.Contains(t, html, "URL.createObjectURL")
	require.Contains(t, html, "Decision Agent API Reference")
	require.Contains(t, html, `rel="icon"`)
	require.Contains(t, html, "favicon.png")
	require.Contains(t, html, "agent-factory-redoc.html")
	require.Contains(t, html, "--docs-nav-height")
	require.Contains(t, html, "position: fixed")
	require.Contains(t, html, "syncDocsNavHeight")
	require.Contains(t, html, "promoteScalarModelsGroup")
	require.Contains(t, html, "sidebar-models")
	require.NotContains(t, html, "ui/scalar-api-reference.js")
	require.NotContains(t, html, "cdn.jsdelivr.net")
	require.NotContains(t, html, "cdn.redocly.com")
	require.NotContains(t, html, "fonts.googleapis.com")
}

func TestRenderRedocStaticHTMLEmbedsSpec(t *testing.T) {
	t.Parallel()

	html := RenderPublicRedocStaticHTML([]byte(`{"openapi":"3.0.2","info":{"title":"Agent Factory API","version":"1.0.0"},"paths":{}}`))

	require.Contains(t, html, "cdn.jsdmirror.com/npm/redoc@2.5.1/bundles/redoc.standalone.js")
	require.Contains(t, html, "openapi-document")
	require.Contains(t, html, "JSON.parse")
	require.Contains(t, html, "Decision Agent API Reference")
	require.Contains(t, html, `rel="icon"`)
	require.Contains(t, html, "favicon.png")
	require.Contains(t, html, "agent-factory.html")
	require.Contains(t, html, "--docs-nav-height")
	require.Contains(t, html, "position: fixed")
	require.Contains(t, html, "syncDocsNavHeight")
	require.NotContains(t, html, "ui/redoc.standalone.js")
	require.NotContains(t, html, "cdn.jsdelivr.net")
	require.NotContains(t, html, "cdn.redocly.com")
	require.NotContains(t, html, "fonts.googleapis.com")
}

func TestRenderRuntimeScalarStaticHTMLEmbedsSpec(t *testing.T) {
	t.Parallel()

	html := RenderRuntimeScalarStaticHTML([]byte(`{"openapi":"3.0.2","info":{"title":"Agent Factory API","version":"1.0.0"},"paths":{}}`))

	require.Contains(t, html, "ui/scalar-api-reference.js")
	require.Contains(t, html, "openapi-document")
	require.Contains(t, html, "URL.createObjectURL")
	require.NotContains(t, html, "cdn.jsdmirror.com")
	require.NotContains(t, html, "fonts.googleapis.com")
}

func TestRenderRuntimeRedocStaticHTMLEmbedsSpec(t *testing.T) {
	t.Parallel()

	html := RenderRuntimeRedocStaticHTML([]byte(`{"openapi":"3.0.2","info":{"title":"Agent Factory API","version":"1.0.0"},"paths":{}}`))

	require.Contains(t, html, "ui/redoc.standalone.js")
	require.Contains(t, html, "openapi-document")
	require.Contains(t, html, "JSON.parse")
	require.NotContains(t, html, "cdn.jsdmirror.com")
	require.NotContains(t, html, "fonts.googleapis.com")
}
