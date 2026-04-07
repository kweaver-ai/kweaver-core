package openapidoc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
