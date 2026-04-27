// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knsearch

import (
	"context"
	"errors"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

// TestKnSearch_Success 测试 KnSearch 固定走本地检索成功
func TestKnSearch_Success(t *testing.T) {
	convey.Convey("TestKnSearch_Success", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		localResp := &interfaces.KnSearchLocalResponse{
			ObjectTypes:   []*interfaces.KnSearchObjectType{},
			RelationTypes: []*interfaces.KnSearchRelationType{},
			ActionTypes:   []*interfaces.KnSearchActionType{},
			Nodes:         []*interfaces.KnSearchNode{},
		}
		fakeLocal := &fakeLocalSearch{resp: localResp, err: nil}

		service := &knSearchService{
			Logger:      mockLogger,
			LocalSearch: fakeLocal,
		}

		ctx := context.Background()
		req := &interfaces.KnSearchReq{
			Query: "测试查询",
			KnID:  "kn-001",
		}

		resp, err := service.KnSearch(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
	})
}

// TestKnSearch_Error 测试 KnSearch 本地检索错误场景
func TestKnSearch_Error(t *testing.T) {
	convey.Convey("TestKnSearch_Error", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		fakeLocal := &fakeLocalSearch{resp: nil, err: errors.New("local search error")}

		service := &knSearchService{
			Logger:      mockLogger,
			LocalSearch: fakeLocal,
		}

		ctx := context.Background()
		req := &interfaces.KnSearchReq{
			Query: "测试查询",
			KnID:  "kn-001",
		}

		_, err := service.KnSearch(ctx, req)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestKnSearch_KnIDConversion 测试 KnID 转换逻辑
func TestKnSearch_KnIDConversion(t *testing.T) {
	convey.Convey("TestKnSearch_KnIDConversion", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLogger(ctrl)
		mockLogger.EXPECT().WithContext(gomock.Any()).Return(mockLogger).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

		fakeLocal := &fakeLocalSearch{
			resp: &interfaces.KnSearchLocalResponse{
				ObjectTypes:   []*interfaces.KnSearchObjectType{},
				RelationTypes: []*interfaces.KnSearchRelationType{},
				ActionTypes:   []*interfaces.KnSearchActionType{},
				Nodes:         []*interfaces.KnSearchNode{},
			},
		}

		service := &knSearchService{
			Logger:      mockLogger,
			LocalSearch: fakeLocal,
		}

		ctx := context.Background()
		req := &interfaces.KnSearchReq{
			Query: "测试查询",
			KnID:  "kn-001",
		}

		_, err := service.KnSearch(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		knIDs := req.GetKnIDs()
		convey.So(len(knIDs), convey.ShouldEqual, 1)
		convey.So(knIDs[0].KnowledgeNetworkID, convey.ShouldEqual, "kn-001")
	})
}

// fakeLocalSearch 用于单测的 IKnSearchLocalService 桩实现
type fakeLocalSearch struct {
	resp *interfaces.KnSearchLocalResponse
	err  error
}

func (f *fakeLocalSearch) Search(_ context.Context, _ *interfaces.KnSearchLocalRequest) (*interfaces.KnSearchLocalResponse, error) {
	return f.resp, f.err
}
