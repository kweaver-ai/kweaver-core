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

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

// TestParseResultFromAgentV1Answer 测试 parseResultFromAgentV1Answer 函数
func TestParseResultFromAgentV1Answer(t *testing.T) {
	convey.Convey("TestParseResultFromAgentV1Answer", t, func() {
		convey.Convey("正常 JSON", func() {
			input := `{"key": "value"}`
			result, err := parseResultFromAgentV1Answer(input)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldEqual, `{"key": "value"}`)
		})

		convey.Convey("带前缀文本的 JSON", func() {
			input := `Here is the result: {"key": "value"}`
			result, err := parseResultFromAgentV1Answer(input)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldEqual, `{"key": "value"}`)
		})

		convey.Convey("带转义字符的 JSON", func() {
			input := `{"key": "value with \\n newline"}`
			result, err := parseResultFromAgentV1Answer(input)
			convey.So(err, convey.ShouldBeNil)
			convey.So(result, convey.ShouldContainSubstring, "value with")
		})

		convey.Convey("无效格式 - 无大括号", func() {
			input := `no json here`
			_, err := parseResultFromAgentV1Answer(input)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("空字符串", func() {
			_, err := parseResultFromAgentV1Answer("")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestParseMetricMissingParamsFromError 测试 parseMetricMissingParamsFromError 函数
func TestParseMetricMissingParamsFromError(t *testing.T) {
	convey.Convey("TestParseMetricMissingParamsFromError", t, func() {
		convey.Convey("正常错误消息", func() {
			result := parseMetricMissingParamsFromError("test_prop", "缺少时间参数")
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(result.Property, convey.ShouldEqual, "test_prop")
			convey.So(result.ErrorMsg, convey.ShouldEqual, "缺少时间参数")
		})

		convey.Convey("空错误消息", func() {
			result := parseMetricMissingParamsFromError("test_prop", "")
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(result.Property, convey.ShouldEqual, "test_prop")
			convey.So(result.ErrorMsg, convey.ShouldEqual, "")
		})
	})
}

// TestParseOperatorMissingParamsFromError 测试 parseOperatorMissingParamsFromError 函数
func TestParseOperatorMissingParamsFromError(t *testing.T) {
	convey.Convey("TestParseOperatorMissingParamsFromError", t, func() {
		result := parseOperatorMissingParamsFromError("test_prop", "缺少参数")
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.Property, convey.ShouldEqual, "test_prop")
		convey.So(result.ErrorMsg, convey.ShouldEqual, "缺少参数")
	})
}

// TestAPIChat_Success 测试 APIChat 成功场景
func TestAPIChat_Success(t *testing.T) {
	convey.Convey("TestAPIChat_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &agentClient{
			logger:      mockLogger,
			baseURL:     "http://localhost:8080/api/agent-factory",
			httpClient:  mockHTTPClient,
			DeployAgent: config.DeployAgentConfig{},
		}

		ctx := context.Background()
		req := &interfaces.ChatRequest{
			AgentKey: "test-agent",
			Query:    "测试问题",
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"message": map[string]interface{}{
					"content": map[string]interface{}{},
				},
			}, nil)

		resp, err := client.APIChat(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestAPIChat_HTTPError 测试 APIChat HTTP 错误
func TestAPIChat_HTTPError(t *testing.T) {
	convey.Convey("TestAPIChat_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &agentClient{
			logger:      mockLogger,
			baseURL:     "http://localhost:8080/api/agent-factory",
			httpClient:  mockHTTPClient,
			DeployAgent: config.DeployAgentConfig{},
		}

		ctx := context.Background()
		req := &interfaces.ChatRequest{
			AgentKey: "test-agent",
			Query:    "测试问题",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.APIChat(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}
