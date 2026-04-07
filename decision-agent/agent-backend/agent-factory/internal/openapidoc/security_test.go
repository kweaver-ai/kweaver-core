package openapidoc

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

func TestNormalizeSecurityCanonicalizesBearerAuthAndKeepsPublicOverrides(t *testing.T) {
	t.Parallel()

	doc := mustLoadDoc(t, `{
  "openapi": "3.0.2",
  "info": {
    "title": "Generated",
    "version": "1.0.0"
  },
  "security": [
    {
      "BearerAuth": []
    }
  ],
  "paths": {
    "/api/agent-factory/v3/agent": {
      "get": {
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "responses": {
          "200": {
            "description": "ok"
          }
        }
      }
    },
    "/api/agent-factory/v3/category": {
      "get": {
        "security": [],
        "responses": {
          "200": {
            "description": "ok"
          }
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer"
      }
    }
  }
}`)

	NormalizeSecurity(doc)

	require.Len(t, doc.Security, 1)
	require.Contains(t, doc.Security[0], "ApiKeyAuth")
	require.NotContains(t, doc.Security[0], "BearerAuth")
	require.Contains(t, doc.Components.SecuritySchemes, "ApiKeyAuth")
	require.NotContains(t, doc.Components.SecuritySchemes, "BearerAuth")
	require.Nil(t, doc.Paths.Value("/api/agent-factory/v3/agent").Get.Security)
	require.NotNil(t, doc.Paths.Value("/api/agent-factory/v3/category").Get.Security)
	require.Empty(t, *doc.Paths.Value("/api/agent-factory/v3/category").Get.Security)
}

func TestNormalizeSecurityMarshalsEmptyScopesAsArrays(t *testing.T) {
	t.Parallel()

	doc := &openapi3.T{
		OpenAPI: "3.0.2",
		Info: &openapi3.Info{
			Title:   "Security",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
		Security: openapi3.SecurityRequirements{
			openapi3.SecurityRequirement{
				"ApiKeyAuth": nil,
			},
		},
		Components: &openapi3.Components{
			SecuritySchemes: openapi3.SecuritySchemes{
				"ApiKeyAuth": &openapi3.SecuritySchemeRef{
					Value: &openapi3.SecurityScheme{
						Type:         "http",
						Scheme:       "bearer",
						BearerFormat: "JWT",
					},
				},
			},
		},
	}

	NormalizeSecurity(doc)

	data, err := MarshalOpenAPIJSON(doc)
	require.NoError(t, err)
	require.Contains(t, string(data), `"ApiKeyAuth": []`)
	require.NotContains(t, string(data), `"ApiKeyAuth": null`)
}
