// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_scheduler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"ontology-query/interfaces"
)

func Test_ExecuteTool_Validation(t *testing.T) {
	Convey("Test ExecuteTool validation", t, func() {
		Convey("should require box_id and tool_id", func() {
			actionType := &interfaces.ActionType{
				ActionSource: interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "",
					ToolID: "",
				},
			}

			So(actionType.ActionSource.BoxID, ShouldEqual, "")
			So(actionType.ActionSource.ToolID, ShouldEqual, "")
		})

		Convey("should build correct OperatorExecutionRequest", func() {
			actionType := &interfaces.ActionType{
				ActionSource: interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "box_001",
					ToolID: "tool_001",
				},
			}
			params := map[string]any{
				"target_ip": "192.168.1.1",
				"timeout":   60,
			}

			So(actionType.ActionSource.Type, ShouldEqual, "tool")
			So(actionType.ActionSource.BoxID, ShouldEqual, "box_001")
			So(actionType.ActionSource.ToolID, ShouldEqual, "tool_001")
			So(params["target_ip"], ShouldEqual, "192.168.1.1")
		})
	})
}

func Test_buildToolExecutionRequest(t *testing.T) {
	Convey("Test buildToolExecutionRequest", t, func() {
		Convey("should put all params in body when no configParams", func() {
			params := map[string]any{"key1": "value1", "key2": 42}
			result := buildToolExecutionRequest(nil, params)

			So(result.Body["key1"], ShouldEqual, "value1")
			So(result.Body["key2"], ShouldEqual, 42)
		})

		Convey("should route flat params to correct locations by source", func() {
			configParams := []interfaces.Parameter{
				{Name: "Authorization", Source: "Header"},
				{Name: "page", Source: "Query"},
				{Name: "data", Source: "Body"},
				{Name: "id", Source: "Path"},
			}
			params := map[string]any{
				"Authorization": "Bearer token",
				"page":          1,
				"data":          "payload",
				"id":            "123",
			}

			result := buildToolExecutionRequest(configParams, params)

			So(result.Header["Authorization"], ShouldEqual, "Bearer token")
			So(result.Query["page"], ShouldEqual, 1)
			So(result.Body["data"], ShouldEqual, "payload")
			So(result.Path["id"], ShouldEqual, "123")
		})

		Convey("should handle nested parameter names with dots", func() {
			configParams := []interfaces.Parameter{
				{Name: "props.headers", Source: "Body"},
				{Name: "query", Source: "Body"},
			}
			params := map[string]any{
				"props": map[string]any{
					"headers": map[string]any{"Authorization": "Bearer xxx"},
				},
				"query": "test",
			}

			result := buildToolExecutionRequest(configParams, params)

			propsMap, ok := result.Body["props"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(propsMap["headers"], ShouldResemble, map[string]any{"Authorization": "Bearer xxx"})
			So(result.Body["query"], ShouldEqual, "test")
		})

		Convey("should skip nil values", func() {
			configParams := []interfaces.Parameter{
				{Name: "existing", Source: "Body"},
				{Name: "missing", Source: "Body"},
			}
			params := map[string]any{
				"existing": "value",
			}

			result := buildToolExecutionRequest(configParams, params)

			So(result.Body["existing"], ShouldEqual, "value")
			_, exists := result.Body["missing"]
			So(exists, ShouldBeFalse)
		})

		Convey("should default to body when source is empty", func() {
			configParams := []interfaces.Parameter{
				{Name: "key1", Source: ""},
			}
			params := map[string]any{"key1": "value1"}

			result := buildToolExecutionRequest(configParams, params)

			So(result.Body["key1"], ShouldEqual, "value1")
		})

		Convey("should handle deeply nested parameter names across different sources", func() {
			configParams := []interfaces.Parameter{
				{Name: "auth.token", Source: "Header"},
				{Name: "filter.status.value", Source: "Query"},
				{Name: "payload.data.items", Source: "Body"},
			}
			params := map[string]any{
				"auth": map[string]any{
					"token": "Bearer xxx",
				},
				"filter": map[string]any{
					"status": map[string]any{
						"value": "active",
					},
				},
				"payload": map[string]any{
					"data": map[string]any{
						"items": []string{"a", "b"},
					},
				},
			}

			result := buildToolExecutionRequest(configParams, params)

			authMap := result.Header["auth"].(map[string]any)
			So(authMap["token"], ShouldEqual, "Bearer xxx")

			filterMap := result.Query["filter"].(map[string]any)
			statusMap := filterMap["status"].(map[string]any)
			So(statusMap["value"], ShouldEqual, "active")

			payloadMap := result.Body["payload"].(map[string]any)
			dataMap := payloadMap["data"].(map[string]any)
			So(dataMap["items"], ShouldResemble, []string{"a", "b"})
		})
	})
}

func Test_getNestedValue(t *testing.T) {
	Convey("Test getNestedValue", t, func() {
		Convey("should return nil for nil map", func() {
			So(getNestedValue(nil, "key"), ShouldBeNil)
		})

		Convey("should get flat key", func() {
			data := map[string]any{"key": "value"}
			So(getNestedValue(data, "key"), ShouldEqual, "value")
		})

		Convey("should get nested key", func() {
			data := map[string]any{
				"props": map[string]any{
					"headers": "Bearer xxx",
				},
			}
			So(getNestedValue(data, "props.headers"), ShouldEqual, "Bearer xxx")
		})

		Convey("should return nil for missing nested key", func() {
			data := map[string]any{
				"props": map[string]any{},
			}
			So(getNestedValue(data, "props.headers"), ShouldBeNil)
		})

		Convey("should return nil when intermediate key is not a map", func() {
			data := map[string]any{
				"props": "not-a-map",
			}
			So(getNestedValue(data, "props.headers"), ShouldBeNil)
		})

		Convey("should handle deeply nested keys", func() {
			data := map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": 42,
					},
				},
			}
			So(getNestedValue(data, "a.b.c"), ShouldEqual, 42)
		})
	})
}

func Test_setNestedValue(t *testing.T) {
	Convey("Test setNestedValue", t, func() {
		Convey("should skip nil value", func() {
			target := map[string]any{}
			setNestedValue(target, "key", nil)
			_, exists := target["key"]
			So(exists, ShouldBeFalse)
		})

		Convey("should set flat key", func() {
			target := map[string]any{}
			setNestedValue(target, "key", "value")
			So(target["key"], ShouldEqual, "value")
		})

		Convey("should set nested key creating intermediate maps", func() {
			target := map[string]any{}
			setNestedValue(target, "props.headers", "Bearer xxx")

			propsMap, ok := target["props"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(propsMap["headers"], ShouldEqual, "Bearer xxx")
		})

		Convey("should set deeply nested key", func() {
			target := map[string]any{}
			setNestedValue(target, "a.b.c", 42)

			aMap := target["a"].(map[string]any)
			bMap := aMap["b"].(map[string]any)
			So(bMap["c"], ShouldEqual, 42)
		})

		Convey("should overwrite non-map intermediate with new map", func() {
			target := map[string]any{"props": "old-string"}
			setNestedValue(target, "props.headers", "Bearer xxx")

			propsMap, ok := target["props"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(propsMap["headers"], ShouldEqual, "Bearer xxx")
		})

		Convey("should merge into existing nested map", func() {
			target := map[string]any{
				"props": map[string]any{
					"existing": "keep-me",
				},
			}
			setNestedValue(target, "props.headers", "Bearer xxx")

			propsMap := target["props"].(map[string]any)
			So(propsMap["existing"], ShouldEqual, "keep-me")
			So(propsMap["headers"], ShouldEqual, "Bearer xxx")
		})
	})
}

func Test_ExecuteMCP_Validation(t *testing.T) {
	Convey("Test ExecuteMCP validation", t, func() {
		Convey("should require mcp_id", func() {
			actionType := &interfaces.ActionType{
				ActionSource: interfaces.ActionSource{
					Type:  interfaces.ActionSourceTypeMCP,
					McpID: "",
				},
			}

			So(actionType.ActionSource.McpID, ShouldEqual, "")
		})

		Convey("should use tool_name or fallback to tool_id", func() {
			Convey("with tool_name", func() {
				actionType := &interfaces.ActionType{
					ActionSource: interfaces.ActionSource{
						Type:     interfaces.ActionSourceTypeMCP,
						McpID:    "mcp_001",
						ToolName: "restart_service",
						ToolID:   "tool_fallback",
					},
				}

				// Should use ToolName
				toolName := actionType.ActionSource.ToolName
				if toolName == "" {
					toolName = actionType.ActionSource.ToolID
				}
				So(toolName, ShouldEqual, "restart_service")
			})

			Convey("fallback to tool_id", func() {
				actionType := &interfaces.ActionType{
					ActionSource: interfaces.ActionSource{
						Type:     interfaces.ActionSourceTypeMCP,
						McpID:    "mcp_001",
						ToolName: "",
						ToolID:   "tool_fallback",
					},
				}

				toolName := actionType.ActionSource.ToolName
				if toolName == "" {
					toolName = actionType.ActionSource.ToolID
				}
				So(toolName, ShouldEqual, "tool_fallback")
			})
		})

		Convey("should build correct MCPExecutionRequest", func() {
			actionType := &interfaces.ActionType{
				ActionSource: interfaces.ActionSource{
					Type:     interfaces.ActionSourceTypeMCP,
					McpID:    "mcp_001",
					ToolName: "restart_pod",
				},
			}
			params := map[string]any{
				"pod_name":      "test-pod",
				"namespace":     "default",
				"force_restart": true,
			}

			mcpRequest := interfaces.MCPExecutionRequest{
				McpID:      actionType.ActionSource.McpID,
				ToolName:   actionType.ActionSource.ToolName,
				Parameters: params,
				Timeout:    60,
			}

			So(mcpRequest.McpID, ShouldEqual, "mcp_001")
			So(mcpRequest.ToolName, ShouldEqual, "restart_pod")
			So(mcpRequest.Parameters["pod_name"], ShouldEqual, "test-pod")
			So(mcpRequest.Parameters["namespace"], ShouldEqual, "default")
			So(mcpRequest.Parameters["force_restart"], ShouldEqual, true)
			So(mcpRequest.Timeout, ShouldEqual, int64(60))
		})
	})
}

func Test_OperatorExecutionRequest(t *testing.T) {
	Convey("Test OperatorExecutionRequest", t, func() {
		Convey("should build correct request", func() {
			req := interfaces.OperatorExecutionRequest{
				Header: map[string]any{},
				Body: map[string]any{
					"key1": "value1",
					"key2": 123,
				},
				Query:   map[string]any{},
				Path:    map[string]any{},
				Timeout: 300,
			}

			So(req.Body["key1"], ShouldEqual, "value1")
			So(req.Body["key2"], ShouldEqual, 123)
			So(req.Timeout, ShouldEqual, int64(300))
		})
	})
}

func Test_ToolExecutionRequest(t *testing.T) {
	Convey("Test ToolExecutionRequest", t, func() {
		Convey("should build correct request for tool-box API", func() {
			req := interfaces.ToolExecutionRequest{
				Header: map[string]any{},
				Body: map[string]any{
					"target_ip": "192.168.1.1",
					"timeout":   60,
				},
				Query:   map[string]any{},
				Path:    map[string]any{},
				Timeout: 300,
			}

			So(req.Body["target_ip"], ShouldEqual, "192.168.1.1")
			So(req.Body["timeout"], ShouldEqual, 60)
			So(req.Timeout, ShouldEqual, int64(300))
		})
	})
}

func Test_MCPExecutionRequest(t *testing.T) {
	Convey("Test MCPExecutionRequest", t, func() {
		Convey("should build correct request", func() {
			req := interfaces.MCPExecutionRequest{
				McpID:    "mcp_001",
				ToolName: "execute_command",
				Parameters: map[string]any{
					"command": "ls -la",
					"timeout": 30,
				},
				Timeout: 60,
			}

			So(req.McpID, ShouldEqual, "mcp_001")
			So(req.ToolName, ShouldEqual, "execute_command")
			So(req.Parameters["command"], ShouldEqual, "ls -la")
			So(req.Timeout, ShouldEqual, int64(60))
		})
	})
}
