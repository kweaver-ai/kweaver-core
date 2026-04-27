package openapidoc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeOverlayMergesMetadataAndExtensions(t *testing.T) {
	t.Parallel()

	doc := mustLoadDoc(t, `{
  "openapi": "3.0.2",
  "info": {
    "title": "Base API",
    "version": "1.0.0"
  },
  "paths": {
    "/api/agent-factory/v1/ping": {
      "get": {
        "responses": {
          "200": {
            "description": "ok"
          }
        }
      }
    }
  }
}`)

	err := MergeOverlay(doc, []byte(`
info:
  title: Agent-Factory API
  description: API to access DIP
security:
  - ApiKeyAuth: []
servers:
  - url: https://{host}:{port}/
tags:
  - name: agent
    description: Agent APIs
x-tagGroups:
  - name: Agent管理（V3）
    tags:
      - agent
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: Authorization
`))
	require.NoError(t, err)

	require.Equal(t, "Agent-Factory API", doc.Info.Title)
	require.Len(t, doc.Servers, 1)
	require.Len(t, doc.Security, 1)
	require.Len(t, doc.Tags, 1)

	raw, err := doc.MarshalJSON()
	require.NoError(t, err)
	require.Contains(t, string(raw), `"x-tagGroups"`)
	require.Contains(t, string(raw), `"ApiKeyAuth"`)
}

func TestMergeMissingFromBaselineFillsGapsWithoutOverwritingExistingContent(t *testing.T) {
	t.Parallel()

	generated := mustLoadDoc(t, `{
  "openapi": "3.0.2",
  "info": {
    "title": "Generated",
    "version": "1.0.0"
  },
  "paths": {
    "/api/agent-factory/v3/agent": {
      "post": {
        "summary": "创建agent",
        "responses": {
          "200": {
            "description": "created"
          }
        }
      }
    }
  }
}`)

	baseline := mustLoadDoc(t, `{
  "openapi": "3.0.2",
  "info": {
    "title": "Manual",
    "version": "1.0.0"
  },
  "paths": {
    "/api/agent-factory/v3/agent": {
      "post": {
        "summary": "手写创建agent",
        "description": "完整说明",
        "responses": {
          "200": {
            "description": "manual created"
          },
          "400": {
            "description": "bad request"
          }
        }
      }
    },
    "/api/agent-factory/v1/app/{app_key}/chat/completion": {
      "post": {
        "summary": "Agent 对话",
        "responses": {
          "200": {
            "description": "ok"
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Error": {
        "type": "object"
      }
    }
  }
}`)

	err := MergeMissingFromBaseline(generated, baseline)
	require.NoError(t, err)

	require.Equal(t, "创建agent", generated.Paths.Value("/api/agent-factory/v3/agent").Post.Summary)
	require.Equal(t, "完整说明", generated.Paths.Value("/api/agent-factory/v3/agent").Post.Description)
	require.NotNil(t, generated.Paths.Value("/api/agent-factory/v1/app/{app_key}/chat/completion"))
	require.NotNil(t, generated.Components)
	require.Contains(t, generated.Components.Schemas, "Error")
}
