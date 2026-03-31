// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	"ontology-query/common/visitor"
	"ontology-query/interfaces"
	omock "ontology-query/interfaces/mock"
)

// setGinMode 设置 Gin 为测试模式并返回恢复函数
func setGinMode() func() {
	oldMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(oldMode)
	}
}

// MockNewRestHandler 创建用于测试的 restHandler
func MockNewRestHandler(
	appSetting *common.AppSetting,
	as interfaces.AuthService,
	ats interfaces.ActionTypeService,
	kns interfaces.KnowledgeNetworkService,
	ots interfaces.ObjectTypeService,
) *restHandler {
	return &restHandler{
		appSetting: appSetting,
		as:         as,
		ats:        ats,
		kns:        kns,
		ots:        ots,
	}
}

func Test_RestHandler_HealthCheck(t *testing.T) {
	Convey("Test RestHandler HealthCheck", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := omock.NewMockAuthService(mockCtrl)
		ats := omock.NewMockActionTypeService(mockCtrl)
		kns := omock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, as, ats, kns, ots)
		handler.RegisterPublic(engine)

		Convey("成功 - 健康检查", func() {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}

func Test_RestHandler_verifyJsonContentTypeMiddleWare(t *testing.T) {
	Convey("Test RestHandler verifyJsonContentTypeMiddleWare", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := omock.NewMockAuthService(mockCtrl)
		ats := omock.NewMockActionTypeService(mockCtrl)
		kns := omock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, as, ats, kns, ots)

		// 注册一个测试路由使用中间件
		engine.POST("/test", handler.verifyJsonContentTypeMiddleWare(), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		Convey("成功 - Content-Type为application/json", func() {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(`{}`)))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("失败 - Content-Type不是application/json", func() {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(`{}`)))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, "text/plain")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotAcceptable)
		})

		Convey("失败 - Content-Type为空", func() {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(`{}`)))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNotAcceptable)
		})
	})
}

func Test_RestHandler_verifyOAuth(t *testing.T) {
	Convey("Test RestHandler verifyOAuth", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := omock.NewMockAuthService(mockCtrl)
		ats := omock.NewMockActionTypeService(mockCtrl)
		kns := omock.NewMockKnowledgeNetworkService(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)

		handler := MockNewRestHandler(appSetting, as, ats, kns, ots)

		Convey("成功 - Token验证通过", func() {
			visitor := hydra.Visitor{
				ID:   "user1",
				Type: hydra.VisitorType_User,
			}
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(visitor, nil)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = req

			resultVisitor, err := handler.verifyOAuth(c.Request.Context(), c)
			So(err, ShouldBeNil)
			So(resultVisitor.ID, ShouldEqual, "user1")
		})

		Convey("失败 - Token验证失败", func() {
			as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, rest.NewHTTPError(context.TODO(), http.StatusUnauthorized, rest.PublicError_Unauthorized))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			c := gin.CreateTestContextOnly(w, engine)
			c.Request = req

			_, err := handler.verifyOAuth(c.Request.Context(), c)
			So(err, ShouldNotBeNil)
			So(w.Code, ShouldEqual, http.StatusUnauthorized)
		})
	})
}

func Test_GenerateVisitor(t *testing.T) {
	Convey("Test GenerateVisitor", t, func() {
		test := setGinMode()
		defer test()

		Convey("成功 - 从header生成visitor", func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_ID, "user1")
			req.Header.Set(interfaces.HTTP_HEADER_ACCOUNT_TYPE, "user")
			req.Header.Set("X-Request-MAC", "mac123")
			req.Header.Set("User-Agent", "test-agent")

			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = req

			visitor := visitor.GenerateVisitor(c)
			So(visitor.ID, ShouldEqual, "user1")
			So(string(visitor.Type), ShouldEqual, "user")
			So(visitor.Mac, ShouldEqual, "mac123")
			So(visitor.UserAgent, ShouldEqual, "test-agent")
		})

		Convey("成功 - header为空时生成默认visitor", func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = req

			visitor := visitor.GenerateVisitor(c)
			So(visitor.ID, ShouldEqual, "")
			So(visitor.TokenID, ShouldEqual, "")
		})
	})
}

func Test_NewRestHandler(t *testing.T) {
	Convey("Test NewRestHandler", t, func() {
		appSetting := &common.AppSetting{}

		Convey("成功 - 创建RestHandler", func() {
			handler := NewRestHandler(appSetting)
			So(handler, ShouldNotBeNil)

			// 测试 RegisterPublic 不会panic
			engine := gin.New()
			So(func() {
				handler.RegisterPublic(engine)
			}, ShouldNotPanic)
		})
	})
}
