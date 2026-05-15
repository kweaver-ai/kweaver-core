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

func Test_ResourceRestHandler_ListResources(t *testing.T) {
	Convey("Test ResourceHandler ListResources\n", t, func() {
		test := setGinMode()
		defer test()

		engine := gin.New()
		engine.Use(gin.Recovery())

		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		rs := vmock.NewMockResourceService(mockCtrl)
		handler := MockNewRestHandler(&common.AppSetting{}, nil, nil, rs, nil, nil, nil, nil, nil, nil, nil)
		handler.RegisterPublic(engine)

		url := "/api/vega-backend/in/v1/resources"

		Convey("Invalid category\n", func() {
			req := httptest.NewRequest(http.MethodGet, url+"?category=unknown", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.Resource.InvalidParameter")
			So(w.Body.String(), ShouldContainSubstring, "invalid category: unknown")
		})

		Convey("Invalid status\n", func() {
			req := httptest.NewRequest(http.MethodGet, url+"?status=unknown", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldContainSubstring, "VegaBackend.Resource.InvalidParameter")
			So(w.Body.String(), ShouldContainSubstring, "invalid status: unknown")
		})

		Convey("Success list resources with category and status\n", func() {
			rs.EXPECT().List(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, params interfaces.ResourcesQueryParams) ([]*interfaces.Resource, int64, error) {
					So(params.Category, ShouldEqual, interfaces.ResourceCategoryDataset)
					So(params.Status, ShouldEqual, interfaces.ResourceStatusActive)
					return []*interfaces.Resource{}, int64(0), nil
				})

			req := httptest.NewRequest(http.MethodGet, url+"?category=dataset&status=active", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			So(w.Result().StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
