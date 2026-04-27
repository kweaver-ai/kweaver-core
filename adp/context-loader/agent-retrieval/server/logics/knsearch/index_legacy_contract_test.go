// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knsearch

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

func TestKnSearch_StripsLegacyNodesAndMessage(t *testing.T) {
	convey.Convey("TestKnSearch_StripsLegacyNodesAndMessage", t, func() {
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
			Nodes:         []*interfaces.KnSearchNode{{ObjectTypeID: "legacy-node"}},
			Message:       "legacy message",
		}
		fakeLocal := &fakeLocalSearch{resp: localResp, err: nil}

		service := &knSearchService{
			Logger:      mockLogger,
			LocalSearch: fakeLocal,
		}

		resp, err := service.KnSearch(context.Background(), &interfaces.KnSearchReq{
			Query: "测试查询",
			KnID:  "kn-001",
		})
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(resp.ObjectTypes, convey.ShouldResemble, localResp.ObjectTypes)
		convey.So(resp.Nodes, convey.ShouldBeEmpty)
		convey.So(resp.Message, convey.ShouldBeNil)
	})
}
