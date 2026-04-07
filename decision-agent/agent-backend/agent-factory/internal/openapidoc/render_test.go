package openapidoc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderScalarStaticHTMLEmbedsSpec(t *testing.T) {
	t.Parallel()

	html := RenderScalarStaticHTML([]byte(`{"openapi":"3.0.2","info":{"title":"Agent Factory API","version":"1.0.0"},"paths":{}}`))

	require.Contains(t, html, "@scalar/api-reference")
	require.Contains(t, html, "openapi-document")
	require.Contains(t, html, "URL.createObjectURL")
	require.Contains(t, html, "Decision Agent API Reference")
	require.Contains(t, html, `rel="icon"`)
	require.Contains(t, html, "favicon.png")
}
