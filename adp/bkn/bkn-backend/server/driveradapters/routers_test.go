// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	bmock "bkn-backend/interfaces/mock"
)

// setGinMode 设置 Gin 为测试模式并返回恢复函数
func setGinMode() func() {
	oldMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	return func() {
		gin.SetMode(oldMode)
	}
}

func Test_RestHandler_HealthCheck(t *testing.T) {
	Convey("Test HealthCheck\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		handler := &restHandler{appSetting: &common.AppSetting{}}
		handler.RegisterPublic(engine)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
	})
}

func Test_RestHandler_VerifyOAuth_Failure(t *testing.T) {
	Convey("Test verifyOAuth returns 401 when token verification fails\n", t, func() {
		test := setGinMode()
		defer test()

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		engine := gin.New()
		engine.Use(gin.Recovery())

		as := bmock.NewMockAuthService(mockCtrl)
		handler := &restHandler{
			appSetting: &common.AppSetting{},
			as:         as,
		}
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Return(hydra.Visitor{}, errors.New("invalid token"))

		req := httptest.NewRequest(http.MethodGet, "/api/bkn-backend/v1/knowledge-networks", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		So(w.Result().StatusCode, ShouldEqual, http.StatusUnauthorized)
	})
}
