// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

// TestSortByRerankAndMatchScore 测试 sortByRerankAndMatchScore 函数
func TestSortByRerankAndMatchScore(t *testing.T) {
	convey.Convey("TestSortByRerankAndMatchScore", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)

		service := &knRetrievalServiceImpl{
			logger: mockLogger,
		}

		convey.Convey("按 RerankScore 降序、相同时按 MatchScore 降序", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptID: "1", ConceptName: "Concept1", RerankScore: 0.8, MatchScore: 10},
				{ConceptID: "2", ConceptName: "Concept2", RerankScore: 0, MatchScore: 5},
				{ConceptID: "3", ConceptName: "Concept3", RerankScore: 0.5, MatchScore: 8},
				{ConceptID: "4", ConceptName: "Concept4", RerankScore: 0, MatchScore: 3},
			}

			result := service.sortByRerankAndMatchScore(concepts)
			convey.So(len(result), convey.ShouldEqual, 4)
			convey.So(result[0].ConceptID, convey.ShouldEqual, "1") // 0.8 最高
			convey.So(result[1].ConceptID, convey.ShouldEqual, "3") // 0.5
			convey.So(result[2].ConceptID, convey.ShouldEqual, "2") // 0, MatchScore 5 > 3
			convey.So(result[3].ConceptID, convey.ShouldEqual, "4")
		})

		convey.Convey("全部 RerankScore 为 0 时保留全部并按 MatchScore 降序", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptID: "1", RerankScore: 0, MatchScore: 1},
				{ConceptID: "2", RerankScore: 0, MatchScore: 2},
			}

			result := service.sortByRerankAndMatchScore(concepts)
			convey.So(len(result), convey.ShouldEqual, 2)
			convey.So(result[0].ConceptID, convey.ShouldEqual, "2")
			convey.So(result[1].ConceptID, convey.ShouldEqual, "1")
		})

		convey.Convey("空输入返回空", func() {
			result := service.sortByRerankAndMatchScore(nil)
			convey.So(result, convey.ShouldBeNil)
		})

		convey.Convey("空切片返回空切片", func() {
			result := service.sortByRerankAndMatchScore([]*interfaces.ConceptResult{})
			convey.So(result, convey.ShouldNotBeNil)
			convey.So(len(result), convey.ShouldEqual, 0)
		})

		convey.Convey("单元素直接返回", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptID: "1", RerankScore: 0.5, MatchScore: 10},
			}
			result := service.sortByRerankAndMatchScore(concepts)
			convey.So(len(result), convey.ShouldEqual, 1)
			convey.So(result[0].ConceptID, convey.ShouldEqual, "1")
		})

		convey.Convey("RerankScore 相同则按 MatchScore 降序", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptID: "a", RerankScore: 0.5, MatchScore: 1},
				{ConceptID: "b", RerankScore: 0.5, MatchScore: 3},
				{ConceptID: "c", RerankScore: 0.5, MatchScore: 2},
			}
			result := service.sortByRerankAndMatchScore(concepts)
			convey.So(len(result), convey.ShouldEqual, 3)
			convey.So(result[0].ConceptID, convey.ShouldEqual, "b") // MatchScore 3
			convey.So(result[1].ConceptID, convey.ShouldEqual, "c") // MatchScore 2
			convey.So(result[2].ConceptID, convey.ShouldEqual, "a") // MatchScore 1
		})

		convey.Convey("MatchScore 为零值(float64)时正常排序不 panic", func() {
			// MatchScore/RerankScore 为 float64，在 Go 中不能为 nil，零值为 0
			concepts := []*interfaces.ConceptResult{
				{ConceptID: "1", RerankScore: 0.5, MatchScore: 0},
				{ConceptID: "2", RerankScore: 0.5, MatchScore: 0},
				{ConceptID: "3", RerankScore: 0.3, MatchScore: 0},
			}
			result := service.sortByRerankAndMatchScore(concepts)
			convey.So(len(result), convey.ShouldEqual, 3)
			convey.So(result[0].RerankScore, convey.ShouldEqual, 0.5)
			convey.So(result[1].RerankScore, convey.ShouldEqual, 0.5)
			convey.So(result[2].RerankScore, convey.ShouldEqual, 0.3)
		})

		convey.Convey("RerankScore 与 MatchScore 都相同时保持原顺序", func() {
			concepts := []*interfaces.ConceptResult{
				{ConceptID: "first", RerankScore: 0.5, MatchScore: 5},
				{ConceptID: "second", RerankScore: 0.5, MatchScore: 5},
			}
			result := service.sortByRerankAndMatchScore(concepts)
			convey.So(len(result), convey.ShouldEqual, 2)
			convey.So(result[0].ConceptID, convey.ShouldBeIn, "first", "second")
			convey.So(result[1].ConceptID, convey.ShouldBeIn, "first", "second")
		})
	})
}

// TestRerankByDataRetrieval_DefaultAction 测试 rerankByDataRetrieval default action 场景
func TestRerankByDataRetrieval_DefaultAction(t *testing.T) {
	convey.Convey("TestRerankByDataRetrieval_DefaultAction", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		service := &knRetrievalServiceImpl{
			logger:        mockLogger,
			dataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		queryUnderstanding := &interfaces.QueryUnderstanding{
			OriginQuery: "测试查询",
		}

		concepts := []*interfaces.ConceptResult{
			{ConceptID: "1", ConceptName: "Concept1", RerankScore: 0.8},
			{ConceptID: "2", ConceptName: "Concept2", RerankScore: 0.5},
		}

		// default action 不调用 KnowledgeRerank
		result, err := service.rerankByDataRetrieval(ctx, queryUnderstanding, concepts, interfaces.KnowledgeRerankActionDefault, 10)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(result), convey.ShouldEqual, 2)
	})
}

// TestRerankByDataRetrieval_VectorAction 测试 rerankByDataRetrieval vector action 场景
func TestRerankByDataRetrieval_VectorAction(t *testing.T) {
	convey.Convey("TestRerankByDataRetrieval_VectorAction", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		service := &knRetrievalServiceImpl{
			logger:        mockLogger,
			dataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		queryUnderstanding := &interfaces.QueryUnderstanding{
			OriginQuery: "测试查询",
		}

		concepts := []*interfaces.ConceptResult{
			{ConceptID: "1", ConceptName: "Concept1", RerankScore: 0.8},
		}

		// Mock KnowledgeRerank 调用
		mockDataRetrieval.EXPECT().KnowledgeRerank(gomock.Any(), gomock.Any()).
			Return([]*interfaces.ConceptResult{
				{ConceptID: "1", ConceptName: "Concept1", RerankScore: 0.9},
			}, nil)

		result, err := service.rerankByDataRetrieval(ctx, queryUnderstanding, concepts, interfaces.KnowledgeRerankActionVector, 10)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(result), convey.ShouldEqual, 1)
	})
}

// TestRerankByDataRetrieval_Error 测试 rerankByDataRetrieval 错误降级场景
func TestRerankByDataRetrieval_Error(t *testing.T) {
	convey.Convey("TestRerankByDataRetrieval_Error", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		service := &knRetrievalServiceImpl{
			logger:        mockLogger,
			dataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		queryUnderstanding := &interfaces.QueryUnderstanding{
			OriginQuery: "测试查询",
		}

		concepts := []*interfaces.ConceptResult{
			{ConceptID: "1", ConceptName: "Concept1", RerankScore: 0.5}, // 添加非零分数，避免被过滤
		}

		// Mock KnowledgeRerank 错误
		mockDataRetrieval.EXPECT().KnowledgeRerank(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("rerank failed"))

		// 期待调用 Warnf 记录降级日志（2个参数：格式字符串 + err）
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger)
		mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any())

		// 修改期待：rerank 失败时应降级返回原始数据，不返回错误
		result, err := service.rerankByDataRetrieval(ctx, queryUnderstanding, concepts, interfaces.KnowledgeRerankActionVector, 10)
		convey.So(err, convey.ShouldBeNil) // 降级后不应返回错误
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(len(result), convey.ShouldEqual, 1) // 返回原始概念列表
		convey.So(result[0].ConceptID, convey.ShouldEqual, "1")
	})
}

// TestRerankByDataRetrieval_WithLimit 测试 rerankByDataRetrieval 分页限制
func TestRerankByDataRetrieval_WithLimit(t *testing.T) {
	convey.Convey("TestRerankByDataRetrieval_WithLimit", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockDataRetrieval := mocks.NewMockDataRetrieval(ctrl)

		service := &knRetrievalServiceImpl{
			logger:        mockLogger,
			dataRetrieval: mockDataRetrieval,
		}

		ctx := context.Background()
		queryUnderstanding := &interfaces.QueryUnderstanding{
			OriginQuery: "测试查询",
		}

		concepts := []*interfaces.ConceptResult{
			{ConceptID: "1", RerankScore: 0.9},
			{ConceptID: "2", RerankScore: 0.8},
			{ConceptID: "3", RerankScore: 0.7},
			{ConceptID: "4", RerankScore: 0.6},
			{ConceptID: "5", RerankScore: 0.5},
		}

		// limit=2 只返回前 2 个
		result, err := service.rerankByDataRetrieval(ctx, queryUnderstanding, concepts, interfaces.KnowledgeRerankActionDefault, 2)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(result), convey.ShouldEqual, 2)
	})
}
