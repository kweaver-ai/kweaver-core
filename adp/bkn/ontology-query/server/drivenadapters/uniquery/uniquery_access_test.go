// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package uniquery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	"ontology-query/interfaces"
)

// newTestUniqueryAccess 创建用于测试的 uniqueryAccess，允许注入 mock HTTP 客户端
func newTestUniqueryAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *uniqueryAccess {
	return &uniqueryAccess{
		appSetting:  appSetting,
		uniqueryUrl: appSetting.UniQueryUrl,
		httpClient:  httpClient,
	}
}

func Test_NewUniqueryAccess(t *testing.T) {
	Convey("Test NewUniqueryAccess", t, func() {
		appSetting := &common.AppSetting{
			UniQueryUrl: "http://test-uniquery",
		}

		Convey("成功 - 创建单例实例", func() {
			// 重置单例
			uAccessOnce = sync.Once{}
			uAccess = nil

			access1 := NewUniqueryAccess(appSetting)
			access2 := NewUniqueryAccess(appSetting)

			So(access1, ShouldNotBeNil)
			So(access2, ShouldNotBeNil)
			So(access1, ShouldEqual, access2) // 应该是同一个实例
		})
	})
}

func Test_uniqueryAccess_GetViewDataByID(t *testing.T) {
	Convey("Test uniqueryAccess GetViewDataByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			UniQueryUrl: "http://test-uniquery",
			ServerSetting: common.ServerSetting{
				ViewDataTimeout: "30s",
			},
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		ua := newTestUniqueryAccess(appSetting, mockHTTPClient)

		ctx := context.Background()
		viewID := "view1"
		viewRequest := interfaces.ViewQuery{
			NeedTotal: true,
			Limit:     10,
		}

		Convey("成功 - 获取视图数据", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			viewData := interfaces.ViewData{
				Datas: []map[string]any{
					{"id": "1", "name": "test"},
				},
				TotalCount: 1,
			}
			responseBytes, _ := json.Marshal(viewData)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := ua.GetViewDataByID(ctx, viewID, viewRequest)

			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			So(result.TotalCount, ShouldEqual, 1)
		})

		Convey("失败 - HTTP 请求错误", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, fmt.Errorf("http request failed"))

			result, err := ua.GetViewDataByID(ctx, viewID, viewRequest)

			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("失败 - HTTP 状态码非 200", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			baseError := rest.BaseError{
				ErrorCode:   "ERROR_CODE",
				Description: "Error description",
			}
			errorBytes, _ := json.Marshal(baseError)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, errorBytes, nil)

			result, err := ua.GetViewDataByID(ctx, viewID, viewRequest)

			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("失败 - HTTP 状态码非 200 且解析 BaseError 失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, invalidJSON, nil)

			result, err := ua.GetViewDataByID(ctx, viewID, viewRequest)

			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("失败 - 响应体为空", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			result, err := ua.GetViewDataByID(ctx, viewID, viewRequest)

			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("失败 - 解析响应失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			expectedURL := "http://test-uniquery/data-views/view1?timeout=30s"
			expectedHeaders := map[string]string{
				interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
				interfaces.HTTP_HEADER_METHOD_OVERRIDE: http.MethodGet,
				interfaces.HTTP_HEADER_ACCOUNT_ID:      accountInfo.ID,
				interfaces.HTTP_HEADER_ACCOUNT_TYPE:    accountInfo.Type,
			}

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(ctx, expectedURL, expectedHeaders, viewRequest).
				Return(http.StatusOK, invalidJSON, nil)

			result, err := ua.GetViewDataByID(ctx, viewID, viewRequest)

			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("失败 - 超时情况 (need_total=true && total_count>0 && len(datas)==0 && search_after!=nil)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			viewRequestWithTotal := interfaces.ViewQuery{
				NeedTotal: true,
				Limit:     10,
			}

			viewData := interfaces.ViewData{
				Datas:       []map[string]any{},
				TotalCount:  100,
				SearchAfter: []any{"value1"},
			}
			responseBytes, _ := json.Marshal(viewData)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := ua.GetViewDataByID(ctx, viewID, viewRequestWithTotal)

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "timeout")
			So(len(result.Datas), ShouldEqual, 0)
		})
	})
}

func Test_uniqueryAccess_GetMetricDataByID(t *testing.T) {
	Convey("Test uniqueryAccess GetMetricDataByID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			UniQueryUrl: "http://test-uniquery",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		ua := newTestUniqueryAccess(appSetting, mockHTTPClient)

		ctx := context.Background()
		metricID := "metric1"
		metricRequest := interfaces.MetricQuery{
			Start:   intPtr(int64(1000)),
			End:     intPtr(int64(2000)),
			StepStr: stringPtr("1m"),
		}

		Convey("成功 - 获取指标数据", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			metricData := interfaces.MetricData{
				Model: interfaces.MetricModel{
					UnitType: "count",
					Unit:     "个",
				},
				Datas: []interfaces.Data{
					{
						Labels: map[string]string{"label1": "value1"},
						Times:  []interface{}{1000, 2000},
						Values: []interface{}{10, 20},
					},
				},
				Step: "1m",
			}
			responseBytes, _ := json.Marshal(metricData)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := ua.GetMetricDataByID(ctx, metricID, metricRequest)

			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			So(result.Step, ShouldEqual, "1m")
		})

		Convey("失败 - HTTP 请求错误", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, fmt.Errorf("http request failed"))

			result, err := ua.GetMetricDataByID(ctx, metricID, metricRequest)

			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("失败 - HTTP 状态码非 200", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			baseError := rest.BaseError{
				ErrorCode:   "ERROR_CODE",
				Description: "Error description",
			}
			errorBytes, _ := json.Marshal(baseError)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, errorBytes, nil)

			result, err := ua.GetMetricDataByID(ctx, metricID, metricRequest)

			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("失败 - HTTP 状态码非 200 且解析 BaseError 失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, invalidJSON, nil)

			result, err := ua.GetMetricDataByID(ctx, metricID, metricRequest)

			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("失败 - 响应体为空", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			result, err := ua.GetMetricDataByID(ctx, metricID, metricRequest)

			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("失败 - 解析响应失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, invalidJSON, nil)

			result, err := ua.GetMetricDataByID(ctx, metricID, metricRequest)

			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})
	})
}

// 辅助函数
func intPtr(i int64) *int64 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
