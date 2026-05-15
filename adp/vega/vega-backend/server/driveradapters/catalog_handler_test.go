// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"vega-backend/common"
	"vega-backend/interfaces"
	vmock "vega-backend/interfaces/mock"
)

func Test_CatalogRestHandler_ListCatalogs(t *testing.T) {
	Convey("Test CatalogHandler ListCatalogs\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		cs := vmock.NewMockCatalogService(mockCtrl)
		handler := MockNewRestHandler(&common.AppSetting{}, nil, cs, nil, nil, nil, nil, nil, nil, nil, nil)
		handler.RegisterPublic(engine)

		url := "/api/vega-backend/in/v1/catalogs"

		Convey("Invalid type\n", func() {
			req := httptest.NewRequest(http.MethodGet, url+"?type=unknown", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.Catalog.InvalidParameter.Type")
			So(w.Body.String(), ShouldContainSubstring, "invalid type: unknown")
		})

		Convey("Invalid health check status\n", func() {
			req := httptest.NewRequest(http.MethodGet, url+"?health_check_status=unknown", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.Catalog.InvalidParameter")
			So(w.Body.String(), ShouldContainSubstring, "invalid health_check_status: unknown")
		})

		Convey("Success list catalogs with type and health check status\n", func() {
			cs.EXPECT().List(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, params interfaces.CatalogsQueryParams) ([]*interfaces.Catalog, int64, error) {
					So(params.Type, ShouldEqual, interfaces.CatalogTypePhysical)
					So(params.HealthCheckStatus, ShouldEqual, interfaces.CatalogHealthStatusHealthy)
					return []*interfaces.Catalog{}, int64(0), nil
				})

			req := httptest.NewRequest(http.MethodGet, url+"?type=physical&health_check_status=healthy", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
