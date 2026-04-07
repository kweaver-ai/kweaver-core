// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

// TestGetToolDetail_Success 测试 GetToolDetail 成功场景
func TestGetToolDetail_Success(t *testing.T) {
	convey.Convey("TestGetToolDetail_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &operatorIntegrationClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/agent-operator-integration",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.GetToolDetailRequest{
			BoxID:  "box-001",
			ToolID: "tool-001",
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"tool_id":     "tool-001",
				"name":        "测试工具",
				"description": "工具描述",
				"status":      "enabled",
				"metadata": map[string]interface{}{
					"version":  "1.0.0",
					"api_spec": map[string]interface{}{},
					"path":     "/test",
					"method":   "POST",
				},
			}, nil)

		resp, err := client.GetToolDetail(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp.ToolID, convey.ShouldEqual, "tool-001")
		convey.So(resp.Name, convey.ShouldEqual, "测试工具")
	})
}

// TestGetToolDetail_HTTPError 测试 GetToolDetail HTTP 错误
func TestGetToolDetail_HTTPError(t *testing.T) {
	convey.Convey("TestGetToolDetail_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &operatorIntegrationClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/agent-operator-integration",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.GetToolDetailRequest{
			BoxID:  "box-001",
			ToolID: "tool-001",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.GetToolDetail(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestGetMCPToolDetail_Success 测试 GetMCPToolDetail 成功场景
func TestGetMCPToolDetail_Success(t *testing.T) {
	convey.Convey("TestGetMCPToolDetail_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &operatorIntegrationClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/agent-operator-integration",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.GetMCPToolDetailRequest{
			McpID:    "mcp-001",
			ToolName: "test_tool",
		}

		// Mock HTTP 成功响应 - 返回工具列表
		mockHTTPClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"tools": []interface{}{
					map[string]interface{}{
						"name":        "test_tool",
						"description": "测试 MCP 工具",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"param1": map[string]interface{}{"type": "string"},
							},
						},
					},
					map[string]interface{}{
						"name":        "other_tool",
						"description": "其他工具",
					},
				},
			}, nil)

		resp, err := client.GetMCPToolDetail(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp.Name, convey.ShouldEqual, "test_tool")
	})
}

// TestGetMCPToolDetail_NotFound 测试 GetMCPToolDetail 工具未找到
func TestGetMCPToolDetail_NotFound(t *testing.T) {
	convey.Convey("TestGetMCPToolDetail_NotFound", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &operatorIntegrationClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/agent-operator-integration",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.GetMCPToolDetailRequest{
			McpID:    "mcp-001",
			ToolName: "nonexistent_tool",
		}

		// Mock HTTP 成功响应 - 返回空工具列表
		mockHTTPClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"tools": []interface{}{
					map[string]interface{}{
						"name":        "other_tool",
						"description": "其他工具",
					},
				},
			}, nil)

		_, err := client.GetMCPToolDetail(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestGetMCPToolDetail_HTTPError 测试 GetMCPToolDetail HTTP 错误
func TestGetMCPToolDetail_HTTPError(t *testing.T) {
	convey.Convey("TestGetMCPToolDetail_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &operatorIntegrationClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/agent-operator-integration",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.GetMCPToolDetailRequest{
			McpID:    "mcp-001",
			ToolName: "test_tool",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.GetMCPToolDetail(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestCallMCPTool_Success 测试 CallMCPTool 成功场景
func TestCallMCPTool_Success(t *testing.T) {
	convey.Convey("TestCallMCPTool_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &operatorIntegrationClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/agent-operator-integration",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.CallMCPToolRequest{
			McpID:    "mcp-001",
			ToolName: "test_tool",
			Parameters: map[string]interface{}{
				"param1": "value1",
			},
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"result": "success",
				"data":   "test data",
			}, nil)

		resp, err := client.CallMCPTool(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp["result"], convey.ShouldEqual, "success")
	})
}

// TestCallMCPTool_HTTPError 测试 CallMCPTool HTTP 错误
func TestCallMCPTool_HTTPError(t *testing.T) {
	convey.Convey("TestCallMCPTool_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &operatorIntegrationClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/agent-operator-integration",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.CallMCPToolRequest{
			McpID:    "mcp-001",
			ToolName: "test_tool",
			Parameters: map[string]interface{}{
				"param1": "value1",
			},
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.CallMCPTool(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}
