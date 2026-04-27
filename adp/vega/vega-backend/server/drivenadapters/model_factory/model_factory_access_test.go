// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package model_factory

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

	"vega-backend/common"
	"vega-backend/interfaces"
)

func newTestModelFactoryAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *modelFactoryAccess {
	return &modelFactoryAccess{
		appSetting:   appSetting,
		httpClient:   httpClient,
		mfManagerUrl: appSetting.MfModelManagerUrl,
		mfAPIUrl:     appSetting.MfModelApiUrl,
	}
}

func TestNewModelFactoryAccess(t *testing.T) {
	Convey("Test NewModelFactoryAccess", t, func() {
		appSetting := &common.AppSetting{
			MfModelManagerUrl: "http://test-mf-manager",
			MfModelApiUrl:     "http://test-mf-api",
		}

		access1 := NewModelFactoryAccess(appSetting)
		access2 := NewModelFactoryAccess(appSetting)

		Convey("Should return singleton instance", func() {
			So(access1, ShouldNotBeNil)
			So(access2, ShouldEqual, access1)
		})
	})
}

func Test_modelFactoryAccess_GetModelByName(t *testing.T) {
	Convey("Test GetModelByName", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			MfModelManagerUrl: "http://test-mf-manager",
			MfModelApiUrl:     "http://test-mf-api",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		mfa := newTestModelFactoryAccess(appSetting, mockHTTPClient)

		modelName := "test-model"
		// httpUrl := "http://test-mf-manager/small-model/get_by_name?model_name=test-model"

		Convey("Success getting model by name", func() {
			model := interfaces.SmallModel{
				ModelID:   "model1",
				ModelName: modelName,
			}
			respData, _ := sonic.Marshal(model)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			result, err := mfa.GetModelByName(ctx, modelName)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.ModelName, ShouldEqual, modelName)
		})

		Convey("Model not found", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNotFound, []byte(""), nil)

			result, err := mfa.GetModelByName(ctx, modelName)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("HTTP request error", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, []byte(""), errors.New("network error"))

			result, err := mfa.GetModelByName(ctx, modelName)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("HTTP status not OK and not NotFound", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, []byte("internal error"), nil)

			result, err := mfa.GetModelByName(ctx, modelName)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Unmarshal response failed", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			result, err := mfa.GetModelByName(ctx, modelName)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_modelFactoryAccess_GetVector(t *testing.T) {
	Convey("Test GetVector", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			MfModelManagerUrl: "http://test-mf-manager",
			MfModelApiUrl:     "http://test-mf-api",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		mfa := newTestModelFactoryAccess(appSetting, mockHTTPClient)

		model := &interfaces.SmallModel{
			ModelID:   "model1",
			BatchSize: 10,
			MaxTokens: 100,
		}
		words := []string{"word1", "word2", "word3"}
		// httpUrl := "http://test-mf-api/small-model/embeddings"

		Convey("Success getting vectors", func() {
			response := map[string]any{
				"data": []*interfaces.VectorResp{
					{Vector: []float32{0.1, 0.2}},
					{Vector: []float32{0.3, 0.4}},
					{Vector: []float32{0.5, 0.6}},
				},
			}
			respData, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			result, err := mfa.GetVector(ctx, model.ModelID, words)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result), ShouldEqual, 3)
		})

		Convey("Empty model name", func() {
			result, err := mfa.GetVector(ctx, "", words)
			So(err, ShouldNotBeNil)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Empty words", func() {
			result, err := mfa.GetVector(ctx, model.ModelID, []string{})
			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 0)
		})
	})
}
