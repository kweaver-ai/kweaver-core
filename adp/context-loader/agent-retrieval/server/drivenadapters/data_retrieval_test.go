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

// TestKnowledgeRerank_Success 测试 KnowledgeRerank 成功场景
func TestKnowledgeRerank_Success(t *testing.T) {
	convey.Convey("TestKnowledgeRerank_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		client := &dataRetrievalClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.KnowledgeRerankReq{
			QueryUnderstanding: &interfaces.QueryUnderstanding{
				OriginQuery: "测试查询",
			},
			KnowledgeConcepts: []*interfaces.ConceptResult{},
			Action:            interfaces.KnowledgeRerankActionVector,
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, []interface{}{
				map[string]interface{}{
					"concept_type": "object_type",
					"id":           "obj-001",
					"name":         "测试对象",
				},
			}, nil)

		results, err := client.KnowledgeRerank(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(results, convey.ShouldNotBeNil)
	})
}

// TestKnowledgeRerank_HTTPError 测试 KnowledgeRerank HTTP 错误
func TestKnowledgeRerank_HTTPError(t *testing.T) {
	convey.Convey("TestKnowledgeRerank_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &dataRetrievalClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.KnowledgeRerankReq{
			QueryUnderstanding: &interfaces.QueryUnderstanding{},
			KnowledgeConcepts:  []*interfaces.ConceptResult{},
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.KnowledgeRerank(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestKnSearch_Success 测试 KnSearch 成功场景
func TestKnSearch_Success(t *testing.T) {
	convey.Convey("TestKnSearch_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &dataRetrievalClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.KnSearchReq{
			Query:        "测试查询",
			KnID:         "kn-001",
			XAccountID:   "account-001",
			XAccountType: "user",
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"results": []interface{}{},
			}, nil)

		resp, err := client.KnSearch(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestKnSearch_HTTPError 测试 KnSearch HTTP 错误
func TestKnSearch_HTTPError(t *testing.T) {
	convey.Convey("TestKnSearch_HTTPError", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &dataRetrievalClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		req := &interfaces.KnSearchReq{
			Query: "测试查询",
			KnID:  "kn-001",
		}

		// Mock HTTP 错误
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(0, nil, errors.New("connection refused"))

		_, err := client.KnSearch(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestKnSearch_WithAllOptionalParams 测试 KnSearch 包含所有可选参数
func TestKnSearch_WithAllOptionalParams(t *testing.T) {
	convey.Convey("TestKnSearch_WithAllOptionalParams", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockHTTPClient := mocks.NewMockHTTPClient(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

		client := &dataRetrievalClient{
			logger:     mockLogger,
			baseURL:    "http://localhost:8080",
			httpClient: mockHTTPClient,
		}

		ctx := context.Background()
		onlySchema := true
		enableRerank := true
		req := &interfaces.KnSearchReq{
			Query:        "测试查询",
			KnID:         "kn-001",
			OnlySchema:   &onlySchema,
			EnableRerank: &enableRerank,
			RetrievalConfig: map[string]interface{}{
				"top_k": 10,
			},
			XAccountID:   "account-001",
			XAccountType: "user",
		}

		// Mock HTTP 成功响应
		mockHTTPClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(200, map[string]interface{}{
				"results": []interface{}{},
			}, nil)

		resp, err := client.KnSearch(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}
