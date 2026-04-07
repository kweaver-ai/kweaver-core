// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/mocks"
)

func TestNoopHydraIntrospect(t *testing.T) {
	convey.Convey("noopHydra.Introspect", t, func() {
		noop := &noopHydra{}

		convey.Convey("returns Active=true with Anonymous visitor type", func() {
			info, err := noop.Introspect(context.Background(), "any-token")
			convey.So(err, convey.ShouldBeNil)
			convey.So(info, convey.ShouldNotBeNil)
			convey.So(info.Active, convey.ShouldBeTrue)
			convey.So(info.VisitorTyp, convey.ShouldEqual, interfaces.Anonymous)
			convey.So(info.VisitorID, convey.ShouldBeEmpty)
		})

		convey.Convey("works with empty token", func() {
			info, err := noop.Introspect(context.Background(), "")
			convey.So(err, convey.ShouldBeNil)
			convey.So(info.Active, convey.ShouldBeTrue)
		})
	})
}

func TestIntrospect(t *testing.T) {
	convey.Convey("TestIntrospect", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		logger := mocks.NewMockLogger(ctrl)
		httpClient := mocks.NewMockHTTPClient(ctrl)
		logger.EXPECT().WithContext(gomock.Any()).Return(logger).AnyTimes()

		hydraClient := &hydra{
			adminAddress: "http://localhost:1234",
			logger:       logger,
			httpClient:   httpClient,
		}

		convey.Convey("HTTP请求错误", func() {
			logger.EXPECT().Error(gomock.Any()).Return()
			httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, fmt.Errorf("connection error"))

			_, err := hydraClient.Introspect(context.Background(), "test-token")
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("JSON序列化错误", func() {
			logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).Return()
			httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(200, "invalid-response", nil)

			_, err := hydraClient.Introspect(context.Background(), "test-token")
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("令牌无效", func() {
			httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(200, &IntrospectInfo{Active: false}, nil)

			_, err := hydraClient.Introspect(context.Background(), "test-token")
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("客户端凭证模式", func() {
			httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(200, &IntrospectInfo{
					Active:    true,
					SubID:     "client-id",
					ClientID:  "client-id",
					TokenType: "access_token",
				}, nil)

			info, err := hydraClient.Introspect(context.Background(), "test-token")
			convey.So(err, convey.ShouldBeNil)
			convey.So(info.VisitorTyp, convey.ShouldEqual, interfaces.Business)
		})
	})
}
