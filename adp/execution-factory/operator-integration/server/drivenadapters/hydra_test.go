package drivenadapters

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestIntrospect(t *testing.T) {
	Convey("TestIntrospect", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		logger := mocks.NewMockLogger(ctrl)
		httpClient := mocks.NewMockHTTPClient(ctrl)
		logger.EXPECT().WithContext(gomock.Any()).Return(logger).AnyTimes()

		hydraClient := &hydraService{
			adminAddress: "http://localhost:1234",
			logger:       logger,
			httpClient:   httpClient,
		}
		c := &gin.Context{}
		ctx := common.SetAccountAuthContextToCtx(context.Background(), &interfaces.AccountAuthContext{
			AccountID:   "user-1",
			AccountType: interfaces.AccessorTypeUser,
		})
		c.Request = c.Request.WithContext(ctx)

		Convey("HTTP请求错误", func() {
			logger.EXPECT().Error(gomock.Any()).Return()
			httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, fmt.Errorf("connection error"))

			_, err := hydraClient.Introspect(c)
			So(err, ShouldNotBeNil)
		})

		Convey("JSON序列化错误", func() {
			logger.EXPECT().Warnf(gomock.Any(), gomock.Any()).Return()
			httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(200, "invalid-response", nil)
			_, err := hydraClient.Introspect(c)
			So(err, ShouldNotBeNil)
		})

		Convey("令牌无效", func() {
			httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(200, &IntrospectInfo{Active: false}, nil)
			_, err := hydraClient.Introspect(c)
			So(err, ShouldNotBeNil)
		})

		Convey("客户端凭证模式", func() {
			httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(200, &IntrospectInfo{
					Active:    true,
					SubID:     "client-id",
					ClientID:  "client-id",
					TokenType: "access_token",
				}, nil)

			info, err := hydraClient.Introspect(c)
			So(err, ShouldBeNil)
			So(info.VisitorTyp, ShouldEqual, interfaces.Business)
		})
		Convey("匿名用户", func() {
			httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(200, &IntrospectInfo{
					Active:    true,
					SubID:     "sub-client-id",
					ClientID:  "client-id",
					TokenType: "access_token",

					Ext: Extend{
						VisitorType: "anonymous",
					},
				}, nil)

			info, err := hydraClient.Introspect(c)
			So(err, ShouldBeNil)
			So(info.VisitorTyp, ShouldEqual, interfaces.Anonymous)
		})
		Convey("实名用户", func() {
			httpClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(200, &IntrospectInfo{
					Active:    true,
					SubID:     "sub-client-id",
					ClientID:  "client-id",
					TokenType: "access_token",
					Ext: Extend{
						VisitorType: "realname",
					},
				}, nil)

			info, err := hydraClient.Introspect(c)
			So(err, ShouldBeNil)
			So(info.VisitorTyp, ShouldEqual, interfaces.RealName)
		})
	})
}

func TestNewHydra_WhenAuthDisabled_ReturnsNoop(t *testing.T) {
	Convey("NewHydra auth disabled returns noop implementation", t, func() {
		t.Setenv("AUTH_ENABLED", "false")
		once = sync.Once{}
		h = nil
		defer func() {
			once = sync.Once{}
			h = nil
		}()
		c := &gin.Context{}
		ctx := common.SetAccountAuthContextToCtx(context.Background(), &interfaces.AccountAuthContext{
			AccountID:   "user-1",
			AccountType: interfaces.AccessorTypeUser,
		})
		c.Request = c.Request.WithContext(ctx)

		client := NewHydra()
		tokenInfo, err := client.Introspect(c)

		So(err, ShouldBeNil)
		So(tokenInfo, ShouldNotBeNil)
		So(tokenInfo.Active, ShouldBeTrue)
		So(tokenInfo.VisitorID, ShouldEqual, "user-1")
		So(tokenInfo.VisitorTyp, ShouldEqual, interfaces.RealName)
	})
}
