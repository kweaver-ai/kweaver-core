package openapidoc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
