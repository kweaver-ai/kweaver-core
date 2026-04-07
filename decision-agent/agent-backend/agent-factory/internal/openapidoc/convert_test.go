package openapidoc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertSwagger2ToOpenAPI3AndPrefixPaths(t *testing.T) {
	t.Parallel()

	doc, err := ConvertSwagger2ToOpenAPI3([]byte(`{
  "swagger": "2.0",
  "info": {
    "title": "Agent Factory API",
    "version": "1.0"
  },
  "basePath": "/api/agent-factory",
  "paths": {
    "/v1/app/{app_key}/chat/completion": {
      "post": {
        "summary": "Agent 对话",
        "responses": {
          "200": {
            "description": "ok"
          }
        }
      }
    }
  }
}`))
	require.NoError(t, err)

	RewriteAgentFactoryPaths(doc)

	require.Equal(t, "3.0.2", doc.OpenAPI)
	require.NotNil(t, doc.Paths)
	require.NotNil(t, doc.Paths.Value("/api/agent-factory/v1/app/{app_key}/chat/completion"))
	require.Nil(t, doc.Paths.Value("/v1/app/{app_key}/chat/completion"))
}
