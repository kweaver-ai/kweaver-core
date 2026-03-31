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

// TestQueryObjectInstances_Success 测试 QueryObjectInstances 成功场景
func TestQueryObjectInstances_Success(t *testing.T) {
	convey.Convey("TestQueryObjectInstances_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &ontologyQueryClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-query",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryObjectInstancesReq{
			KnID:  "kn-001",
			OtID:  "ot-001",
			Limit: 10,
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"datas":       []interface{}{},
				"object_type": map[string]interface{}{},
			}, nil)

		resp, err := client.QueryObjectInstances(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestQueryObjectInstances_HTTPError 测试 QueryObjectInstances HTTP 错误
func TestQueryObjectInstances_HTTPError(t *testing.T) {
	convey.Convey("TestQueryObjectInstances_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyQueryClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-query",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryObjectInstancesReq{
			KnID:  "kn-001",
			OtID:  "ot-001",
			Limit: 10,
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.QueryObjectInstances(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestQueryLogicProperties_Success 测试 QueryLogicProperties 成功场景
func TestQueryLogicProperties_Success(t *testing.T) {
	convey.Convey("TestQueryLogicProperties_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyQueryClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-query",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryLogicPropertiesReq{
			KnID:               "kn-001",
			OtID:               "ot-001",
			InstanceIdentities: []map[string]interface{}{{"id": "obj-001"}},
			Properties:         []string{"prop1"},
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"datas": []interface{}{
					map[string]interface{}{"prop1": "value1"},
				},
			}, nil)

		resp, err := client.QueryLogicProperties(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp.Datas), convey.ShouldEqual, 1)
	})
}

// TestQueryLogicProperties_HTTPError 测试 QueryLogicProperties HTTP 错误
func TestQueryLogicProperties_HTTPError(t *testing.T) {
	convey.Convey("TestQueryLogicProperties_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyQueryClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-query",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryLogicPropertiesReq{
			KnID: "kn-001",
			OtID: "ot-001",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.QueryLogicProperties(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestQueryActions_Success 测试 QueryActions 成功场景
func TestQueryActions_Success(t *testing.T) {
	convey.Convey("TestQueryActions_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyQueryClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-query",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryActionsRequest{
			KnID:               "kn-001",
			AtID:               "at-001",
			InstanceIdentities: []map[string]interface{}{{"id": "obj-001"}},
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"action_source": map[string]interface{}{
					"type":    "tool",
					"box_id":  "box-001",
					"tool_id": "tool-001",
				},
				"actions": []interface{}{
					map[string]interface{}{
						"parameters": map[string]interface{}{"key": "value"},
					},
				},
				"total_count": 1,
				"overall_ms":  100,
			}, nil)

		resp, err := client.QueryActions(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp.ActionSource, convey.ShouldNotBeNil)
		convey.So(resp.ActionSource.Type, convey.ShouldEqual, "tool")
	})
}

// TestQueryActions_HTTPError 测试 QueryActions HTTP 错误
func TestQueryActions_HTTPError(t *testing.T) {
	convey.Convey("TestQueryActions_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyQueryClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-query",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryActionsRequest{
			KnID: "kn-001",
			AtID: "at-001",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.QueryActions(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestQueryInstanceSubgraph_Success 测试 QueryInstanceSubgraph 成功场景
func TestQueryInstanceSubgraph_Success(t *testing.T) {
	convey.Convey("TestQueryInstanceSubgraph_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyQueryClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-query",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryInstanceSubgraphReq{
			KnID: "kn-001",
			RelationTypePaths: []interface{}{
				map[string]interface{}{"source": "obj-001"},
			},
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"entries": []interface{}{},
			}, nil)

		resp, err := client.QueryInstanceSubgraph(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestQueryInstanceSubgraph_HTTPError 测试 QueryInstanceSubgraph HTTP 错误
func TestQueryInstanceSubgraph_HTTPError(t *testing.T) {
	convey.Convey("TestQueryInstanceSubgraph_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &ontologyQueryClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080/api/ontology-query",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.QueryInstanceSubgraphReq{
			KnID: "kn-001",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.QueryInstanceSubgraph(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}
