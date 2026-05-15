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

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"vega-backend/common"
	"vega-backend/interfaces"
	vmock "vega-backend/interfaces/mock"
)

func Test_ConnectorTypeRestHandler_UpdateConnectorType(t *testing.T) {
	Convey("Test ConnectorTypeHandler UpdateConnectorType\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		as := vmock.NewMockAuthService(mockCtrl)
		cts := vmock.NewMockConnectorTypeService(mockCtrl)
		handler := MockNewRestHandler(&common.AppSetting{}, as, nil, nil, nil, nil, cts, nil, nil, nil, nil)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().
			Return(hydra.Visitor{ID: "u1", Type: hydra.VisitorType_User}, nil)

		tp := "mysql"
		url := "/api/vega-backend/v1/connector-types/" + tp

		Convey("Body type mismatch\n", func() {
			reqData := interfaces.ConnectorTypeReq{
				Type:     "postgres",
				Name:     "MySQL",
				Mode:     interfaces.ConnectorModeLocal,
				Category: interfaces.ConnectorCategoryTable,
			}
			reqParamByte, _ := sonic.Marshal(reqData)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusConflict)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.ConnectorType.TypeMismatch")
		})

		Convey("Success update connector type\n", func() {
			reqData := interfaces.ConnectorTypeReq{
				Type:     tp,
				Name:     "MySQL",
				Mode:     interfaces.ConnectorModeLocal,
				Category: interfaces.ConnectorCategoryTable,
			}
			cts.EXPECT().GetByType(gomock.Any(), tp).
				Return(&interfaces.ConnectorType{
					Type:     tp,
					Name:     "MySQL",
					Mode:     interfaces.ConnectorModeLocal,
					Category: interfaces.ConnectorCategoryTable,
				}, nil)
			cts.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(reqData)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Success update fileset connector type\n", func() {
			reqData := interfaces.ConnectorTypeReq{
				Type:     tp,
				Name:     "AnyShare",
				Mode:     interfaces.ConnectorModeLocal,
				Category: interfaces.ConnectorCategoryFileset,
			}
			cts.EXPECT().GetByType(gomock.Any(), tp).
				Return(&interfaces.ConnectorType{
					Type:     tp,
					Name:     "AnyShare",
					Mode:     interfaces.ConnectorModeLocal,
					Category: interfaces.ConnectorCategoryFileset,
				}, nil)
			cts.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

			reqParamByte, _ := sonic.Marshal(reqData)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("Body type omitted\n", func() {
			reqData := map[string]any{
				"name":     "MySQL",
				"mode":     interfaces.ConnectorModeLocal,
				"category": interfaces.ConnectorCategoryTable,
			}
			reqParamByte, _ := sonic.Marshal(reqData)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(reqParamByte))
			req.Header.Set(interfaces.CONTENT_TYPE_NAME, interfaces.CONTENT_TYPE_JSON)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.ConnectorType.InvalidParameter.Type")
		})
	})
}

func Test_ConnectorTypeRestHandler_ListConnectorTypes(t *testing.T) {
	Convey("Test ConnectorTypeHandler ListConnectorTypes\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		as := vmock.NewMockAuthService(mockCtrl)
		cts := vmock.NewMockConnectorTypeService(mockCtrl)
		handler := MockNewRestHandler(&common.AppSetting{}, as, nil, nil, nil, nil, cts, nil, nil, nil, nil)
		handler.RegisterPublic(engine)

		as.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).AnyTimes().
			Return(hydra.Visitor{ID: "u1", Type: hydra.VisitorType_User}, nil)

		url := "/api/vega-backend/v1/connector-types"

		Convey("Invalid enabled\n", func() {
			req := httptest.NewRequest(http.MethodGet, url+"?enabled=maybe", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.ConnectorType.InvalidParameter")
			So(w.Body.String(), ShouldContainSubstring, "invalid enabled: maybe")
		})

		Convey("Invalid mode\n", func() {
			req := httptest.NewRequest(http.MethodGet, url+"?mode=unknown", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.ConnectorType.InvalidParameter.Mode")
			So(w.Body.String(), ShouldContainSubstring, "invalid mode: unknown")
		})

		Convey("Invalid category\n", func() {
			req := httptest.NewRequest(http.MethodGet, url+"?category=unknown", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.ConnectorType.InvalidParameter.Category")
			So(w.Body.String(), ShouldContainSubstring, "invalid category: unknown")
		})

		Convey("Success list connector types with mode category and enabled\n", func() {
			cts.EXPECT().List(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, params interfaces.ConnectorTypesQueryParams) ([]*interfaces.ConnectorType, int64, error) {
					So(params.Mode, ShouldEqual, interfaces.ConnectorModeLocal)
					So(params.Category, ShouldEqual, interfaces.ConnectorCategoryFileset)
					So(params.Enabled, ShouldNotBeNil)
					So(*params.Enabled, ShouldBeTrue)
					return []*interfaces.ConnectorType{}, int64(0), nil
				})

			req := httptest.NewRequest(http.MethodGet, url+"?mode=local&category=fileset&enabled=true", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
