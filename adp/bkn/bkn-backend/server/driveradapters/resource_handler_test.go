// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	"bkn-backend/interfaces"
	bmock "bkn-backend/interfaces/mock"
)

func MockNewResourceRestHandler(appSetting *common.AppSetting,
	as interfaces.AuthService,
	kns interfaces.KNService) (r *restHandler) {

	r = &restHandler{
		appSetting: appSetting,
		as:         as,
		kns:        kns,
	}
	return r
}

func Test_RestHandler_ListResources(t *testing.T) {
	Convey("Test RestHandler ListResources\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		as := bmock.NewMockAuthService(mockCtrl)
		kns := bmock.NewMockKNService(mockCtrl)

		handler := MockNewResourceRestHandler(appSetting, as, kns)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().Return(hydra.Visitor{}, nil)

		url := "/api/bkn-backend/v1/resources"

		Convey("Success ListResources with KN type\n", func() {
			kns.EXPECT().ListKnSrcs(gomock.Any(), gomock.Any()).Return([]interfaces.Resource{}, 0, nil)

			req := httptest.NewRequest(http.MethodGet, url+"?resource_type="+interfaces.RESOURCE_TYPE_KN, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("Success ListResources with unknown type\n", func() {
			req := httptest.NewRequest(http.MethodGet, url+"?resource_type=unknown", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			// 默认情况下不返回错误，只是不处理
			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
