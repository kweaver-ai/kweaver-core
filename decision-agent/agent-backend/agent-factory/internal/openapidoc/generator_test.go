package openapidoc

import (
	"context"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
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

func TestRenderScalarStaticHTMLEmbedsSpec(t *testing.T) {
	t.Parallel()

	html := RenderScalarStaticHTML([]byte(`{"openapi":"3.0.2","info":{"title":"Agent Factory API","version":"1.0.0"},"paths":{}}`))

	require.Contains(t, html, "@scalar/api-reference")
	require.Contains(t, html, "openapi-document")
	require.Contains(t, html, "URL.createObjectURL")
	require.Contains(t, html, "Agent Factory API Reference")
}

func TestBuildComparisonReportIncludesCountsAndMissingOperations(t *testing.T) {
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
        "summary": "创建agent",
        "responses": {
          "200": {
            "description": "created"
          }
        }
      },
      "get": {
        "summary": "Agent列表",
        "responses": {
          "200": {
            "description": "ok"
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
  }
}`)

	report := BuildComparisonReport(generated, baseline)

	require.Contains(t, report, "Generated paths: 1")
	require.Contains(t, report, "Baseline paths: 2")
	require.Contains(t, report, "/api/agent-factory/v1/app/{app_key}/chat/completion")
	require.Contains(t, report, "GET /api/agent-factory/v3/agent")
}

func TestBuildComparisonReportIncludesSemanticMismatches(t *testing.T) {
	t.Parallel()

	generated := mustLoadDoc(t, `{
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
    "/api/agent-factory/v3/agent/{agent_id}/copy": {
      "post": {
        "summary": "复制agent",
        "tags": ["agent"],
        "responses": {
          "201": {
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
  "security": [
    {
      "ApiKeyAuth": []
    }
  ],
  "paths": {
    "/api/agent-factory/v3/agent/{agent_id}/copy": {
      "post": {
        "summary": "复制agent为模板",
        "tags": ["agent", "模板"],
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "type": "object"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "ok"
          },
          "201": {
            "description": "created"
          }
        }
      }
    }
  }
}`)

	report := BuildComparisonReport(generated, baseline)

	require.Contains(t, report, "Summary mismatches")
	require.Contains(t, report, `generated="复制agent" baseline="复制agent为模板"`)
	require.Contains(t, report, "Tag mismatches")
	require.Contains(t, report, "generated=agent baseline=agent,模板")
	require.Contains(t, report, "Security mismatches")
	require.Contains(t, report, "generated=BearerAuth baseline=ApiKeyAuth")
	require.Contains(t, report, "Request body mismatches")
	require.Contains(t, report, "generated=false baseline=true")
	require.Contains(t, report, "Response code mismatches")
	require.Contains(t, report, "generated=201 baseline=200,201")
}

func TestNormalizePathParametersMarksPathParamsRequired(t *testing.T) {
	t.Parallel()

	doc := mustLoadDoc(t, `{
  "openapi": "3.0.2",
  "info": {
    "title": "Generated",
    "version": "1.0.0"
  },
  "paths": {
    "/api/agent-factory/v1/app/{app_key}/chat/completion": {
      "post": {
        "parameters": [
          {
            "name": "app_key",
            "in": "path",
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "ok"
          }
        }
      }
    }
  }
}`)

	NormalizePathParameters(doc)

	require.True(t, doc.Paths.Value("/api/agent-factory/v1/app/{app_key}/chat/completion").Post.Parameters[0].Value.Required)
}

func TestNormalizePathParametersAddsMissingPathParamsAndNormalizesComponents(t *testing.T) {
	t.Parallel()

	doc := mustLoadDoc(t, `{
  "openapi": "3.0.2",
  "info": {
    "title": "Generated",
    "version": "1.0.0"
  },
  "paths": {
    "/api/agent-factory/v1/app/{app_key}/chat/termination": {
      "post": {
        "responses": {
          "200": {
            "description": "ok"
          }
        }
      }
    }
  },
  "components": {
    "parameters": {
      "IDInt": {
        "name": "id",
        "in": "path",
        "schema": {
          "type": "integer"
        }
      }
    }
  }
}`)

	NormalizePathParameters(doc)

	pathParams := doc.Paths.Value("/api/agent-factory/v1/app/{app_key}/chat/termination").Parameters
	require.Len(t, pathParams, 1)
	require.Equal(t, "app_key", pathParams[0].Value.Name)
	require.True(t, pathParams[0].Value.Required)
	require.True(t, doc.Components.Parameters["IDInt"].Value.Required)
}

func TestNormalizeOperationIDsMakesDuplicatesUnique(t *testing.T) {
	t.Parallel()

	doc := mustLoadDoc(t, `{
  "openapi": "3.0.2",
  "info": {
    "title": "Generated",
    "version": "1.0.0"
  },
  "paths": {
    "/api/agent-factory/v3/agent/{agent_id}/publish-info": {
      "get": {
        "operationId": "getPublishInfo",
        "responses": {
          "200": {
            "description": "ok"
          }
        }
      }
    },
    "/api/agent-factory/v3/agent-tpl/{id}/publish-info": {
      "get": {
        "operationId": "getPublishInfo",
        "responses": {
          "200": {
            "description": "ok"
          }
        }
      }
    }
  }
}`)

	NormalizeOperationIDs(doc)

	first := doc.Paths.Value("/api/agent-factory/v3/agent/{agent_id}/publish-info").Get.OperationID
	second := doc.Paths.Value("/api/agent-factory/v3/agent-tpl/{id}/publish-info").Get.OperationID
	require.NotEqual(t, first, second)
	require.Contains(t, first, "getPublishInfo_get_")
	require.Contains(t, second, "getPublishInfo_get_")
}

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

func TestValidateOpenAPIRejectsInvalidDoc(t *testing.T) {
	t.Parallel()

	doc := &openapi3.T{
		OpenAPI: "3.0.2",
		Info:    &openapi3.Info{Title: "broken", Version: "1.0.0"},
	}

	err := ValidateOpenAPI(context.Background(), doc)
	require.Error(t, err)
}

func TestLoadOpenAPIDocSanitizesLegacySchemaFields(t *testing.T) {
	t.Parallel()

	doc, err := loadOpenAPIDoc([]byte(`{
  "openapi": "3.0.2",
  "info": {
    "title": "Manual",
    "version": "1.0.0"
  },
  "paths": {},
  "components": {
    "schemas": {
      "Legacy": {
        "type": "object",
        "properties": {
          "avatar": {
            "type": "string",
            "example": 1,
            "required": true
          },
          "progress": {
            "type": "array",
            "example": {
              "progress": []
            }
          }
        }
      }
    }
  }
}`))
	require.NoError(t, err)

	schema := doc.Components.Schemas["Legacy"].Value.Properties["avatar"].Value
	require.Equal(t, "1", schema.Example)
	require.Empty(t, schema.Required)
	require.Nil(t, doc.Components.Schemas["Legacy"].Value.Properties["progress"].Value.Example)
}

func mustLoadDoc(t *testing.T, data string) *openapi3.T {
	t.Helper()

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(strings.TrimSpace(data)))
	require.NoError(t, err)

	return doc
}
