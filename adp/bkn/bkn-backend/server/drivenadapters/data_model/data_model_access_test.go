// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_model

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

// newTestDataModelAccess creates a test instance with mock HTTP client
func newTestDataModelAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *dataModelAccess {
	return &dataModelAccess{
		appSetting: appSetting,
		httpClient: httpClient,
	}
}

func TestNewDataModelAccess(t *testing.T) {
	Convey("Test NewDataModelAccess", t, func() {
		appSetting := &common.AppSetting{
			DataModelUrl: "http://test-data-model",
		}

		access1 := NewDataModelAccess(appSetting)
		access2 := NewDataModelAccess(appSetting)

		Convey("Should return singleton instance", func() {
			So(access1, ShouldNotBeNil)
			So(access2, ShouldEqual, access1)
		})
	})
}

func Test_dataModelAccess_GetMetricModelByID(t *testing.T) {
	Convey("Test GetMetricModelByID", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			DataModelUrl: "http://test-data-model",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)

		dda := newTestDataModelAccess(appSetting, mockHTTPClient)

		modelID := "test-model-id"
		// httpUrl := "http://test-data-model/metric-models/test-model-id"

		Convey("Success getting metric model", func() {
			models := []*interfaces.MetricModel{
				{
					ModelID:   modelID,
					ModelName: "Test Model",
					GroupID:   "group1",
					GroupName: "Group 1",
					FieldsMap: make(map[string]interfaces.Field),
				},
			}
			respData, _ := sonic.Marshal(models)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			result, err := dda.GetMetricModelByID(ctx, modelID)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.ModelID, ShouldEqual, modelID)
		})

		Convey("Model not found", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNotFound, []byte(""), nil)

			result, err := dda.GetMetricModelByID(ctx, modelID)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Empty result array", func() {
			models := []*interfaces.MetricModel{}
			respData, _ := sonic.Marshal(models)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			result, err := dda.GetMetricModelByID(ctx, modelID)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("HTTP request error", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, []byte(""), errors.New("network error"))

			result, err := dda.GetMetricModelByID(ctx, modelID)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Non-200 status code with error", func() {
			baseError := rest.BaseError{
				ErrorCode:    "INTERNAL_ERROR",
				Description:  "Internal server error",
				ErrorDetails: "Something went wrong",
			}
			respData, _ := sonic.Marshal(baseError)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, respData, nil)

			result, err := dda.GetMetricModelByID(ctx, modelID)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Non-200 status code with invalid error format", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, []byte("invalid json"), nil)

			result, err := dda.GetMetricModelByID(ctx, modelID)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Invalid response data", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			result, err := dda.GetMetricModelByID(ctx, modelID)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
