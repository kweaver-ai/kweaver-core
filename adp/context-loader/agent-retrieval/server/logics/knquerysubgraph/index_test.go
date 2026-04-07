// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knquerysubgraph

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

// TestQueryInstanceSubgraph_Success 测试 QueryInstanceSubgraph 成功场景
func TestQueryInstanceSubgraph_Success(t *testing.T) {
	convey.Convey("TestQueryInstanceSubgraph_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		service := &knQuerySubgraphService{
			Logger:        mockLogger,
			OntologyQuery: mockOntologyQuery,
		}

		ctx := context.Background()
		req := &interfaces.QueryInstanceSubgraphReq{
			KnID: "kn-001",
			RelationTypePaths: []interface{}{
				map[string]interface{}{"source": "obj-001"},
			},
		}

		// Mock OntologyQuery 成功响应
		mockOntologyQuery.EXPECT().QueryInstanceSubgraph(gomock.Any(), gomock.Any()).
			Return(&interfaces.QueryInstanceSubgraphResp{
				Entries: []interface{}{},
			}, nil)

		resp, err := service.QueryInstanceSubgraph(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestQueryInstanceSubgraph_Error 测试 QueryInstanceSubgraph 错误场景
func TestQueryInstanceSubgraph_Error(t *testing.T) {
	convey.Convey("TestQueryInstanceSubgraph_Error", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		service := &knQuerySubgraphService{
			Logger:        mockLogger,
			OntologyQuery: mockOntologyQuery,
		}

		ctx := context.Background()
		req := &interfaces.QueryInstanceSubgraphReq{
			KnID: "kn-001",
		}

		// Mock OntologyQuery 错误
		mockOntologyQuery.EXPECT().QueryInstanceSubgraph(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("query failed"))

		_, err := service.QueryInstanceSubgraph(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestQueryInstanceSubgraph_WithEntries 测试有返回结果的场景
func TestQueryInstanceSubgraph_WithEntries(t *testing.T) {
	convey.Convey("TestQueryInstanceSubgraph_WithEntries", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockOntologyQuery := mocks.NewMockDrivenOntologyQuery(ctrl)

		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()

		service := &knQuerySubgraphService{
			Logger:        mockLogger,
			OntologyQuery: mockOntologyQuery,
		}

		ctx := context.Background()
		req := &interfaces.QueryInstanceSubgraphReq{
			KnID: "kn-001",
			RelationTypePaths: []interface{}{
				map[string]interface{}{
					"source_ot_id":  "ot-001",
					"relation_type": "rt-001",
					"target_ot_id":  "ot-002",
				},
			},
		}

		// Mock OntologyQuery 返回有结果的响应
		mockOntologyQuery.EXPECT().QueryInstanceSubgraph(gomock.Any(), gomock.Any()).
			Return(&interfaces.QueryInstanceSubgraphResp{
				Entries: []interface{}{
					map[string]interface{}{
						"source": map[string]interface{}{"id": "obj-001"},
						"target": map[string]interface{}{"id": "obj-002"},
					},
				},
			}, nil)

		resp, err := service.QueryInstanceSubgraph(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp.Entries, convey.ShouldNotBeNil)
	})
}
