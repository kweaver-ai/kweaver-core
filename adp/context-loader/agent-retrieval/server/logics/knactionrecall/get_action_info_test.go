// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knactionrecall

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

// TestGetActionInfo_QueryActionsError 测试 QueryActions 调用失败的场景
func TestGetActionInfo_QueryActionsError(t *testing.T) {
	convey.Convey("TestGetActionInfo_QueryActionsError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              &config.Config{},
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID:             "kn-001",
			AtID:             "at-001",
			InstanceIdentity: map[string]interface{}{"id": "obj-001"},
		}

		// Mock QueryActions 返回错误
		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("query actions failed"))

		_, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestGetActionInfo_ActionSourceNil 测试 ActionSource 为 nil 的场景
func TestGetActionInfo_ActionSourceNil(t *testing.T) {
	convey.Convey("TestGetActionInfo_ActionSourceNil", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              &config.Config{},
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID:             "kn-001",
			AtID:             "at-001",
			InstanceIdentity: map[string]interface{}{"id": "obj-001"},
		}

		// Mock QueryActions 返回 ActionSource 为 nil
		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			Return(&interfaces.QueryActionsResponse{
				ActionSource: nil,
				Actions:      []interfaces.ActionParams{},
			}, nil)

		resp, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp.DynamicTools), convey.ShouldEqual, 0)
	})
}

// TestGetActionInfo_ActionsEmpty 测试 Actions 为空的场景
func TestGetActionInfo_ActionsEmpty(t *testing.T) {
	convey.Convey("TestGetActionInfo_ActionsEmpty", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              &config.Config{},
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID:             "kn-001",
			AtID:             "at-001",
			InstanceIdentity: map[string]interface{}{"id": "obj-001"},
		}

		// Mock QueryActions 返回空 Actions
		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			Return(&interfaces.QueryActionsResponse{
				ActionSource: &interfaces.ActionSource{Type: interfaces.ActionSourceTypeTool},
				Actions:      []interfaces.ActionParams{},
			}, nil)

		resp, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp.DynamicTools), convey.ShouldEqual, 0)
	})
}

// TestGetActionInfo_UnsupportedType 测试不支持的 action_source 类型
func TestGetActionInfo_UnsupportedType(t *testing.T) {
	convey.Convey("TestGetActionInfo_UnsupportedType", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              &config.Config{},
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID:             "kn-001",
			AtID:             "at-001",
			InstanceIdentity: map[string]interface{}{"id": "obj-001"},
		}

		// Mock QueryActions 返回不支持的类型
		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			Return(&interfaces.QueryActionsResponse{
				ActionSource: &interfaces.ActionSource{Type: "unsupported_type"},
				Actions: []interfaces.ActionParams{
					{Parameters: map[string]interface{}{"key": "value"}},
				},
			}, nil)

		_, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestGetActionInfo_ToolType_Success 测试 Tool 类型成功路径
func TestGetActionInfo_ToolType_Success(t *testing.T) {
	convey.Convey("TestGetActionInfo_ToolType_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		cfg := &config.Config{
			OntologyQuery: config.PrivateBaseConfig{
				PrivateProtocol: "http",
				PrivateHost:     "ontology-query",
				PrivatePort:     13018,
			},
		}

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              cfg,
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID:             "kn-001",
			AtID:             "at-001",
			InstanceIdentity: map[string]interface{}{"id": "obj-001"},
		}

		// Mock QueryActions 返回 Tool 类型
		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			Return(&interfaces.QueryActionsResponse{
				ActionSource: &interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "box-001",
					ToolID: "tool-001",
				},
				Actions: []interfaces.ActionParams{
					{Parameters: map[string]interface{}{"param1": "value1"}},
				},
			}, nil)

		// Mock GetToolDetail
		mockOperatorIntegration.EXPECT().GetToolDetail(gomock.Any(), gomock.Any()).
			Return(&interfaces.GetToolDetailResponse{
				Name:        "TestTool",
				Description: "Test tool description",
				Metadata: interfaces.ToolMetadata{
					APISpec: map[string]interface{}{
						"parameters": []interface{}{
							map[string]interface{}{
								"name":     "pod_name",
								"in":       "query",
								"required": true,
								"schema":   map[string]interface{}{"type": "string"},
							},
						},
						"request_body": map[string]interface{}{
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"namespace": map[string]interface{}{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
			}, nil)

		resp, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp.DynamicTools), convey.ShouldEqual, 1)

		tool := resp.DynamicTools[0]
		convey.So(tool.Name, convey.ShouldEqual, "TestTool")
		convey.So(tool.APICallStrategy, convey.ShouldEqual, interfaces.ResultProcessStrategyKnActionRecall)
		convey.So(tool.OriginalSchema, convey.ShouldBeNil)

		// 验证 api_url 指向行动驱动执行接口
		convey.So(tool.APIURL, convey.ShouldEqual,
			"http://ontology-query:13018/api/ontology-query/in/v1/knowledge-networks/kn-001/action-types/at-001/execute")

		// 验证 parameters 顶层为 dynamic_params + _instance_identities
		params := tool.Parameters
		convey.So(params["type"], convey.ShouldEqual, "object")
		props := params["properties"].(map[string]interface{})
		convey.So(props["dynamic_params"], convey.ShouldNotBeNil)
		convey.So(props["_instance_identities"], convey.ShouldNotBeNil)

		// 验证 dynamic_params 中包含去壳后的参数
		dynamicParams := props["dynamic_params"].(map[string]interface{})
		dynamicProps := dynamicParams["properties"].(map[string]interface{})
		convey.So(dynamicProps["pod_name"], convey.ShouldNotBeNil)
		convey.So(dynamicProps["namespace"], convey.ShouldNotBeNil)

		// 验证 fixed_params 为 ActionDriverFixedParams 结构
		fixedParams, ok := tool.FixedParams.(interfaces.ActionDriverFixedParams)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(fixedParams.DynamicParams["param1"], convey.ShouldEqual, "value1")
		convey.So(len(fixedParams.InstanceIdentities), convey.ShouldEqual, 1)
		convey.So(fixedParams.InstanceIdentities[0]["id"], convey.ShouldEqual, "obj-001")
	})
}

// TestGetActionInfo_WithoutInstanceIdentity_Success 测试 _instance_identity 非必传时的成功路径
func TestGetActionInfo_WithoutInstanceIdentity_Success(t *testing.T) {
	convey.Convey("TestGetActionInfo_WithoutInstanceIdentity_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		cfg := &config.Config{
			OntologyQuery: config.PrivateBaseConfig{
				PrivateProtocol: "http",
				PrivateHost:     "ontology-query",
				PrivatePort:     13018,
			},
		}

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              cfg,
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID: "kn-001",
			AtID: "at-001",
		}

		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, actionsReq *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
				convey.So(actionsReq.InstanceIdentities, convey.ShouldNotBeNil)
				convey.So(len(actionsReq.InstanceIdentities), convey.ShouldEqual, 0)
				return &interfaces.QueryActionsResponse{
					ActionSource: &interfaces.ActionSource{
						Type:   interfaces.ActionSourceTypeTool,
						BoxID:  "box-001",
						ToolID: "tool-001",
					},
					Actions: []interfaces.ActionParams{
						{Parameters: map[string]interface{}{"param1": "value1"}},
					},
				}, nil
			})

		mockOperatorIntegration.EXPECT().GetToolDetail(gomock.Any(), gomock.Any()).
			Return(&interfaces.GetToolDetailResponse{
				Name:        "TestTool",
				Description: "Test tool description",
				Metadata: interfaces.ToolMetadata{
					APISpec: map[string]interface{}{
						"parameters": []interface{}{
							map[string]interface{}{
								"name":     "pod_name",
								"in":       "query",
								"required": true,
								"schema":   map[string]interface{}{"type": "string"},
							},
						},
					},
				},
			}, nil)

		resp, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp.DynamicTools), convey.ShouldEqual, 1)

		fixedParams, ok := resp.DynamicTools[0].FixedParams.(interfaces.ActionDriverFixedParams)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(fixedParams.DynamicParams["param1"], convey.ShouldEqual, "value1")
		convey.So(len(fixedParams.InstanceIdentities), convey.ShouldEqual, 0)
	})
}

// TestGetActionInfo_MCPType_Success 测试 MCP 类型成功路径
func TestGetActionInfo_MCPType_Success(t *testing.T) {
	convey.Convey("TestGetActionInfo_MCPType_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		cfg := &config.Config{
			OntologyQuery: config.PrivateBaseConfig{
				PrivateProtocol: "http",
				PrivateHost:     "ontology-query",
				PrivatePort:     13018,
			},
		}

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              cfg,
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID:             "kn-001",
			AtID:             "at-001",
			InstanceIdentity: map[string]interface{}{"id": "obj-001"},
		}

		// Mock QueryActions 返回 MCP 类型
		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			Return(&interfaces.QueryActionsResponse{
				ActionSource: &interfaces.ActionSource{
					Type:     interfaces.ActionSourceTypeMCP,
					McpID:    "mcp-001",
					ToolName: "test_tool",
				},
				Actions: []interfaces.ActionParams{
					{Parameters: map[string]interface{}{"param1": "value1"}},
				},
			}, nil)

		// Mock GetMCPToolDetail
		mockOperatorIntegration.EXPECT().GetMCPToolDetail(gomock.Any(), gomock.Any()).
			Return(&interfaces.GetMCPToolDetailResponse{
				Name:        "TestMCPTool",
				Description: "Test MCP tool description",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"disease_id": map[string]interface{}{"type": "string"},
					},
					"required": []interface{}{"disease_id"},
				},
			}, nil)

		resp, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp.DynamicTools), convey.ShouldEqual, 1)

		tool := resp.DynamicTools[0]
		convey.So(tool.Name, convey.ShouldEqual, "TestMCPTool")
		convey.So(tool.APICallStrategy, convey.ShouldEqual, interfaces.ResultProcessStrategyKnActionRecall)
		convey.So(tool.OriginalSchema, convey.ShouldBeNil)

		// 验证 api_url 指向行动驱动执行接口
		convey.So(tool.APIURL, convey.ShouldEqual,
			"http://ontology-query:13018/api/ontology-query/in/v1/knowledge-networks/kn-001/action-types/at-001/execute")

		// 验证 parameters 顶层为 dynamic_params + _instance_identities
		params := tool.Parameters
		convey.So(params["type"], convey.ShouldEqual, "object")
		props := params["properties"].(map[string]interface{})
		convey.So(props["dynamic_params"], convey.ShouldNotBeNil)
		convey.So(props["_instance_identities"], convey.ShouldNotBeNil)

		// 验证 dynamic_params 中包含 MCP schema 参数
		dynamicParams := props["dynamic_params"].(map[string]interface{})
		dynamicProps := dynamicParams["properties"].(map[string]interface{})
		convey.So(dynamicProps["disease_id"], convey.ShouldNotBeNil)

		// 验证 fixed_params 为 ActionDriverFixedParams 结构
		fixedParams, ok := tool.FixedParams.(interfaces.ActionDriverFixedParams)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(fixedParams.DynamicParams["param1"], convey.ShouldEqual, "value1")
		convey.So(len(fixedParams.InstanceIdentities), convey.ShouldEqual, 1)
		convey.So(fixedParams.InstanceIdentities[0]["id"], convey.ShouldEqual, "obj-001")
	})
}

// TestGetActionInfo_GetToolDetailError 测试 GetToolDetail 调用失败
func TestGetActionInfo_GetToolDetailError(t *testing.T) {
	convey.Convey("TestGetActionInfo_GetToolDetailError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		cfg := &config.Config{}

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              cfg,
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID:             "kn-001",
			AtID:             "at-001",
			InstanceIdentity: map[string]interface{}{"id": "obj-001"},
		}

		// Mock QueryActions 返回 Tool 类型
		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			Return(&interfaces.QueryActionsResponse{
				ActionSource: &interfaces.ActionSource{
					Type:   interfaces.ActionSourceTypeTool,
					BoxID:  "box-001",
					ToolID: "tool-001",
				},
				Actions: []interfaces.ActionParams{
					{Parameters: map[string]interface{}{"param1": "value1"}},
				},
			}, nil)

		// Mock GetToolDetail 返回错误
		mockOperatorIntegration.EXPECT().GetToolDetail(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("get tool detail failed"))

		_, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestGetActionInfo_GetMCPToolDetailError 测试 GetMCPToolDetail 调用失败
func TestGetActionInfo_GetMCPToolDetailError(t *testing.T) {
	convey.Convey("TestGetActionInfo_GetMCPToolDetailError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		cfg := &config.Config{}

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              cfg,
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID:             "kn-001",
			AtID:             "at-001",
			InstanceIdentity: map[string]interface{}{"id": "obj-001"},
		}

		// Mock QueryActions 返回 MCP 类型
		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			Return(&interfaces.QueryActionsResponse{
				ActionSource: &interfaces.ActionSource{
					Type:     interfaces.ActionSourceTypeMCP,
					McpID:    "mcp-001",
					ToolName: "test_tool",
				},
				Actions: []interfaces.ActionParams{
					{Parameters: map[string]interface{}{"param1": "value1"}},
				},
			}, nil)

		// Mock GetMCPToolDetail 返回错误
		mockOperatorIntegration.EXPECT().GetMCPToolDetail(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("get mcp tool detail failed"))

		_, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// 保留原有测试
// TestMCPAPIURLConstruction 测试 MCP 类型的 API URL 构造
func TestMCPAPIURLConstruction(t *testing.T) {
	testCases := []struct {
		name        string
		mcpID       string
		toolName    string
		expectedURL string
	}{
		{
			name:        "标准 MCP ID 和工具名",
			mcpID:       "ad3ca391-a598-4764-a6c8-e62b9662e87e",
			toolName:    "generate_treatment_plan",
			expectedURL: "/api/agent-retrieval/v1/mcp/proxy/ad3ca391-a598-4764-a6c8-e62b9662e87e/tools/generate_treatment_plan/call",
		},
		{
			name:        "简短 MCP ID",
			mcpID:       "mcp-001",
			toolName:    "query_data",
			expectedURL: "/api/agent-retrieval/v1/mcp/proxy/mcp-001/tools/query_data/call",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 使用与 get_action_info.go 相同的格式化逻辑
			apiURL := "/api/agent-retrieval/v1/mcp/proxy/" + tc.mcpID + "/tools/" + tc.toolName + "/call"
			if apiURL != tc.expectedURL {
				t.Errorf("API URL 构造错误\n期望: %s\n实际: %s", tc.expectedURL, apiURL)
			}
		})
	}
}

// TestMCPFixedParamsFlat 测试 MCP 类型的固定参数是扁平化结构
func TestMCPFixedParamsFlat(t *testing.T) {
	// 模拟从 Ontology Query 返回的行动参数
	actionParams := map[string]interface{}{
		"disease_id":    "disease_000001",
		"include_drugs": "true",
		"lang":          "zh",
	}

	// MCP 类型直接使用扁平化的 map
	fixedParams := actionParams

	// 验证是扁平结构（没有 header/path/query/body 分层）
	if _, hasHeader := fixedParams["header"]; hasHeader {
		t.Error("MCP fixed_params 不应该有 header 字段")
	}
	if _, hasPath := fixedParams["path"]; hasPath {
		t.Error("MCP fixed_params 不应该有 path 字段")
	}
	if _, hasQuery := fixedParams["query"]; hasQuery {
		t.Error("MCP fixed_params 不应该有 query 字段")
	}
	if _, hasBody := fixedParams["body"]; hasBody {
		t.Error("MCP fixed_params 不应该有 body 字段")
	}

	// 验证原始字段存在
	if fixedParams["disease_id"] != "disease_000001" {
		t.Error("MCP fixed_params 应该包含原始的 disease_id 字段")
	}
}

// TestActionSourceTypeMCP 测试 MCP 类型常量定义正确
func TestActionSourceTypeMCP(t *testing.T) {
	if interfaces.ActionSourceTypeMCP != "mcp" {
		t.Errorf("ActionSourceTypeMCP 应该为 'mcp', 实际为 '%s'", interfaces.ActionSourceTypeMCP)
	}
	if interfaces.ActionSourceTypeTool != "tool" {
		t.Errorf("ActionSourceTypeTool 应该为 'tool', 实际为 '%s'", interfaces.ActionSourceTypeTool)
	}
}

// TestActionSourceMCPFields 测试 ActionSource 结构体包含 MCP 相关字段
func TestActionSourceMCPFields(t *testing.T) {
	actionSource := interfaces.ActionSource{
		Type:     interfaces.ActionSourceTypeMCP,
		McpID:    "test-mcp-id",
		ToolName: "test-tool-name",
	}

	if actionSource.Type != "mcp" {
		t.Error("ActionSource.Type 应该为 'mcp'")
	}
	if actionSource.McpID != "test-mcp-id" {
		t.Error("ActionSource.McpID 应该为 'test-mcp-id'")
	}
	if actionSource.ToolName != "test-tool-name" {
		t.Error("ActionSource.ToolName 应该为 'test-tool-name'")
	}
}

// ==================== _instance_identities 合并逻辑测试 ====================

// TestGetActionInfo_InstanceIdentities_MultipleValid 测试传入多个有效 _instance_identities
func TestGetActionInfo_InstanceIdentities_MultipleValid(t *testing.T) {
	convey.Convey("TestGetActionInfo_InstanceIdentities_MultipleValid", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		cfg := &config.Config{
			OntologyQuery: config.PrivateBaseConfig{
				PrivateProtocol: "http",
				PrivateHost:     "ontology-query",
				PrivatePort:     13018,
			},
		}

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              cfg,
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID: "kn-001",
			AtID: "at-001",
			InstanceIdentities: []map[string]any{
				{"code": "A"},
				{"code": "B"},
			},
		}

		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, actionsReq *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
				convey.So(len(actionsReq.InstanceIdentities), convey.ShouldEqual, 2)
				convey.So(actionsReq.InstanceIdentities[0]["code"], convey.ShouldEqual, "A")
				convey.So(actionsReq.InstanceIdentities[1]["code"], convey.ShouldEqual, "B")
				return &interfaces.QueryActionsResponse{
					ActionSource: &interfaces.ActionSource{
						Type:   interfaces.ActionSourceTypeTool,
						BoxID:  "box-001",
						ToolID: "tool-001",
					},
					Actions: []interfaces.ActionParams{
						{Parameters: map[string]any{"param1": "value1"}},
					},
				}, nil
			})

		mockOperatorIntegration.EXPECT().GetToolDetail(gomock.Any(), gomock.Any()).
			Return(&interfaces.GetToolDetailResponse{
				Name:        "TestTool",
				Description: "Test tool description",
				Metadata: interfaces.ToolMetadata{
					APISpec: map[string]any{
						"parameters": []any{
							map[string]any{
								"name":     "pod_name",
								"in":       "query",
								"required": true,
								"schema":   map[string]any{"type": "string"},
							},
						},
					},
				},
			}, nil)

		resp, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp.DynamicTools), convey.ShouldEqual, 1)

		fixedParams, ok := resp.DynamicTools[0].FixedParams.(interfaces.ActionDriverFixedParams)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(fixedParams.InstanceIdentities), convey.ShouldEqual, 2)
		convey.So(fixedParams.InstanceIdentities[0]["code"], convey.ShouldEqual, "A")
		convey.So(fixedParams.InstanceIdentities[1]["code"], convey.ShouldEqual, "B")
	})
}

// TestGetActionInfo_InstanceIdentities_FilterEmptyMaps 测试 _instance_identities 中的空 map 被过滤
func TestGetActionInfo_InstanceIdentities_FilterEmptyMaps(t *testing.T) {
	convey.Convey("TestGetActionInfo_InstanceIdentities_FilterEmptyMaps", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		cfg := &config.Config{
			OntologyQuery: config.PrivateBaseConfig{
				PrivateProtocol: "http",
				PrivateHost:     "ontology-query",
				PrivatePort:     13018,
			},
		}

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              cfg,
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID: "kn-001",
			AtID: "at-001",
			InstanceIdentities: []map[string]any{
				{"code": "A"},
				{},
				{"code": "C"},
			},
		}

		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, actionsReq *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
				convey.So(len(actionsReq.InstanceIdentities), convey.ShouldEqual, 2)
				convey.So(actionsReq.InstanceIdentities[0]["code"], convey.ShouldEqual, "A")
				convey.So(actionsReq.InstanceIdentities[1]["code"], convey.ShouldEqual, "C")
				return &interfaces.QueryActionsResponse{
					ActionSource: &interfaces.ActionSource{
						Type:   interfaces.ActionSourceTypeTool,
						BoxID:  "box-001",
						ToolID: "tool-001",
					},
					Actions: []interfaces.ActionParams{
						{Parameters: map[string]any{"param1": "value1"}},
					},
				}, nil
			})

		mockOperatorIntegration.EXPECT().GetToolDetail(gomock.Any(), gomock.Any()).
			Return(&interfaces.GetToolDetailResponse{
				Name:        "TestTool",
				Description: "Test tool description",
				Metadata: interfaces.ToolMetadata{
					APISpec: map[string]any{
						"parameters": []any{
							map[string]any{
								"name":     "pod_name",
								"in":       "query",
								"required": true,
								"schema":   map[string]any{"type": "string"},
							},
						},
					},
				},
			}, nil)

		resp, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)

		fixedParams, ok := resp.DynamicTools[0].FixedParams.(interfaces.ActionDriverFixedParams)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(fixedParams.InstanceIdentities), convey.ShouldEqual, 2)
	})
}

// TestGetActionInfo_InstanceIdentities_PriorityOverIdentity 测试同时传两者时 _instance_identities 优先
func TestGetActionInfo_InstanceIdentities_PriorityOverIdentity(t *testing.T) {
	convey.Convey("TestGetActionInfo_InstanceIdentities_PriorityOverIdentity", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		cfg := &config.Config{
			OntologyQuery: config.PrivateBaseConfig{
				PrivateProtocol: "http",
				PrivateHost:     "ontology-query",
				PrivatePort:     13018,
			},
		}

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              cfg,
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID:             "kn-001",
			AtID:             "at-001",
			InstanceIdentity: map[string]any{"id": "should-be-ignored"},
			InstanceIdentities: []map[string]any{
				{"id": "from-identities-1"},
				{"id": "from-identities-2"},
			},
		}

		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, actionsReq *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
				convey.So(len(actionsReq.InstanceIdentities), convey.ShouldEqual, 2)
				convey.So(actionsReq.InstanceIdentities[0]["id"], convey.ShouldEqual, "from-identities-1")
				convey.So(actionsReq.InstanceIdentities[1]["id"], convey.ShouldEqual, "from-identities-2")
				return &interfaces.QueryActionsResponse{
					ActionSource: &interfaces.ActionSource{
						Type:   interfaces.ActionSourceTypeTool,
						BoxID:  "box-001",
						ToolID: "tool-001",
					},
					Actions: []interfaces.ActionParams{
						{Parameters: map[string]any{"param1": "value1"}},
					},
				}, nil
			})

		mockOperatorIntegration.EXPECT().GetToolDetail(gomock.Any(), gomock.Any()).
			Return(&interfaces.GetToolDetailResponse{
				Name:        "TestTool",
				Description: "Test tool description",
				Metadata: interfaces.ToolMetadata{
					APISpec: map[string]any{
						"parameters": []any{
							map[string]any{
								"name":     "pod_name",
								"in":       "query",
								"required": true,
								"schema":   map[string]any{"type": "string"},
							},
						},
					},
				},
			}, nil)

		resp, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)

		fixedParams, ok := resp.DynamicTools[0].FixedParams.(interfaces.ActionDriverFixedParams)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(fixedParams.InstanceIdentities), convey.ShouldEqual, 2)
		convey.So(fixedParams.InstanceIdentities[0]["id"], convey.ShouldEqual, "from-identities-1")
	})
}

// TestGetActionInfo_InstanceIdentities_AllEmpty 测试 _instance_identities 全部为空 map
func TestGetActionInfo_InstanceIdentities_AllEmpty(t *testing.T) {
	convey.Convey("TestGetActionInfo_InstanceIdentities_AllEmpty", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		cfg := &config.Config{
			OntologyQuery: config.PrivateBaseConfig{
				PrivateProtocol: "http",
				PrivateHost:     "ontology-query",
				PrivatePort:     13018,
			},
		}

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              cfg,
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID: "kn-001",
			AtID: "at-001",
			InstanceIdentities: []map[string]any{
				{},
				{},
			},
		}

		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, actionsReq *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
				convey.So(len(actionsReq.InstanceIdentities), convey.ShouldEqual, 0)
				return &interfaces.QueryActionsResponse{
					ActionSource: &interfaces.ActionSource{
						Type:   interfaces.ActionSourceTypeTool,
						BoxID:  "box-001",
						ToolID: "tool-001",
					},
					Actions: []interfaces.ActionParams{
						{Parameters: map[string]any{"param1": "value1"}},
					},
				}, nil
			})

		mockOperatorIntegration.EXPECT().GetToolDetail(gomock.Any(), gomock.Any()).
			Return(&interfaces.GetToolDetailResponse{
				Name:        "TestTool",
				Description: "Test tool description",
				Metadata: interfaces.ToolMetadata{
					APISpec: map[string]any{
						"parameters": []any{
							map[string]any{
								"name":     "pod_name",
								"in":       "query",
								"required": true,
								"schema":   map[string]any{"type": "string"},
							},
						},
					},
				},
			}, nil)

		resp, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)

		fixedParams, ok := resp.DynamicTools[0].FixedParams.(interfaces.ActionDriverFixedParams)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(fixedParams.InstanceIdentities), convey.ShouldEqual, 0)
	})
}

// TestGetActionInfo_InstanceIdentities_FallbackToIdentity 测试未传 _instance_identities 时回退到 _instance_identity
func TestGetActionInfo_InstanceIdentities_FallbackToIdentity(t *testing.T) {
	convey.Convey("TestGetActionInfo_InstanceIdentities_FallbackToIdentity", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)
		mockOperatorIntegration := mocks.NewMockDrivenOperatorIntegration(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		cfg := &config.Config{
			OntologyQuery: config.PrivateBaseConfig{
				PrivateProtocol: "http",
				PrivateHost:     "ontology-query",
				PrivatePort:     13018,
			},
		}

		service := &knActionRecallServiceImpl{
			logger:              mockLogger,
			config:              cfg,
			ontologyQuery:       mockOntologyQuery,
			operatorIntegration: mockOperatorIntegration,
		}

		ctx := context.Background()
		req := &interfaces.KnActionRecallRequest{
			KnID:             "kn-001",
			AtID:             "at-001",
			InstanceIdentity: map[string]any{"id": "legacy-obj-001"},
		}

		mockOntologyQuery.EXPECT().QueryActions(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, actionsReq *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
				convey.So(len(actionsReq.InstanceIdentities), convey.ShouldEqual, 1)
				convey.So(actionsReq.InstanceIdentities[0]["id"], convey.ShouldEqual, "legacy-obj-001")
				return &interfaces.QueryActionsResponse{
					ActionSource: &interfaces.ActionSource{
						Type:   interfaces.ActionSourceTypeTool,
						BoxID:  "box-001",
						ToolID: "tool-001",
					},
					Actions: []interfaces.ActionParams{
						{Parameters: map[string]any{"param1": "value1"}},
					},
				}, nil
			})

		mockOperatorIntegration.EXPECT().GetToolDetail(gomock.Any(), gomock.Any()).
			Return(&interfaces.GetToolDetailResponse{
				Name:        "TestTool",
				Description: "Test tool description",
				Metadata: interfaces.ToolMetadata{
					APISpec: map[string]any{
						"parameters": []any{
							map[string]any{
								"name":     "pod_name",
								"in":       "query",
								"required": true,
								"schema":   map[string]any{"type": "string"},
							},
						},
					},
				},
			}, nil)

		resp, err := service.GetActionInfo(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)

		fixedParams, ok := resp.DynamicTools[0].FixedParams.(interfaces.ActionDriverFixedParams)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(fixedParams.InstanceIdentities), convey.ShouldEqual, 1)
		convey.So(fixedParams.InstanceIdentities[0]["id"], convey.ShouldEqual, "legacy-obj-001")
	})
}
