// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

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

func newTestDataViewAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *dataViewAccess {
	return &dataViewAccess{
		appSetting: appSetting,
		httpClient: httpClient,
	}
}

func TestNewDataViewAccess(t *testing.T) {
	Convey("Test NewDataViewAccess", t, func() {
		appSetting := &common.AppSetting{
			DataViewUrl: "http://test-data-view",
		}

		access1 := NewDataViewAccess(appSetting)
		access2 := NewDataViewAccess(appSetting)

		Convey("Should return singleton instance", func() {
			So(access1, ShouldNotBeNil)
			So(access2, ShouldEqual, access1)
		})
	})
}

func Test_dataViewAccess_GetDataViewByID(t *testing.T) {
	Convey("Test GetDataViewByID", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			DataViewUrl: "http://test-data-view",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		dva := newTestDataViewAccess(appSetting, mockHTTPClient)

		viewID := "test-view-id"
		// httpUrl := "http://test-data-view/data-views/test-view-id"

		Convey("Success getting data view", func() {
			views := []*interfaces.DataView{
				{
					ViewID:   viewID,
					ViewName: "Test View",
					Fields: []*interfaces.ViewField{
						{Name: "field1", Type: "string"},
					},
				},
			}
			respData, _ := sonic.Marshal(views)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			result, err := dva.GetDataViewByID(ctx, viewID)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.ViewID, ShouldEqual, viewID)
			So(result.FieldsMap, ShouldNotBeNil)
		})

		Convey("View not found", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNotFound, []byte(""), nil)

			result, err := dva.GetDataViewByID(ctx, viewID)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("HTTP request error", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, []byte(""), errors.New("network error"))

			result, err := dva.GetDataViewByID(ctx, viewID)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Empty result array", func() {
			views := []*interfaces.DataView{}
			respData, _ := sonic.Marshal(views)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			result, err := dva.GetDataViewByID(ctx, viewID)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Non-200 status code with baseError", func() {
			baseError := rest.BaseError{
				ErrorCode:    "INTERNAL_ERROR",
				Description:  "Internal server error",
				ErrorDetails: "Something went wrong",
			}
			respData, _ := sonic.Marshal(baseError)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, respData, nil)

			result, err := dva.GetDataViewByID(ctx, viewID)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "GetDataViewByIDs failed")
		})

		Convey("Non-200 status code with invalid error format", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, []byte("invalid json"), nil)

			result, err := dva.GetDataViewByID(ctx, viewID)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Unmarshal data view failed", func() {
			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			result, err := dva.GetDataViewByID(ctx, viewID)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Success with AccountInfo in context", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "test-account-id",
				Type: "test-account-type",
				Name: "Test Account",
			}
			ctxWithAccount := context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			views := []*interfaces.DataView{
				{
					ViewID:   viewID,
					ViewName: "Test View",
					Fields: []*interfaces.ViewField{
						{Name: "field1", Type: "string"},
					},
				},
			}
			respData, _ := sonic.Marshal(views)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			result, err := dva.GetDataViewByID(ctxWithAccount, viewID)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.ViewID, ShouldEqual, viewID)
		})
	})
}

func Test_dataViewAccess_GetDataStart(t *testing.T) {
	Convey("Test GetDataStart", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			UniQueryUrl: "http://test-uni-query",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		dva := newTestDataViewAccess(appSetting, mockHTTPClient)

		viewID := "test-view-id"
		incKey := "id"
		incValue := "value1"
		limit := 10
		// httpUrl := "http://test-uni-query/data-views/test-view-id?include_view=true&timeout=5m"

		Convey("Success getting data start", func() {
			result := interfaces.ViewQueryResult{
				TotalCount:  100,
				SearchAfter: []any{"value1"},
				Entries:     []map[string]any{{"id": "1", "name": "test"}},
			}
			respData, _ := sonic.Marshal(result)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			res, err := dva.GetDataStart(ctx, viewID, incKey, incValue, limit)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.TotalCount, ShouldEqual, 100)
		})

		Convey("HTTP request error", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, []byte(""), errors.New("network error"))

			result, err := dva.GetDataStart(ctx, viewID, incKey, incValue, limit)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Non-200 status code", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, []byte("error"), nil)

			result, err := dva.GetDataStart(ctx, viewID, incKey, incValue, limit)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Unmarshal result failed", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			result, err := dva.GetDataStart(ctx, viewID, incKey, incValue, limit)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Success with empty incKey", func() {
			result := interfaces.ViewQueryResult{
				TotalCount:  100,
				SearchAfter: []any{"value1"},
				Entries:     []map[string]any{{"id": "1", "name": "test"}},
			}
			respData, _ := sonic.Marshal(result)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			res, err := dva.GetDataStart(ctx, viewID, "", nil, limit)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.TotalCount, ShouldEqual, 100)
		})

		Convey("Success with incKey but nil incValue", func() {
			result := interfaces.ViewQueryResult{
				TotalCount:  100,
				SearchAfter: []any{"value1"},
				Entries:     []map[string]any{{"id": "1", "name": "test"}},
			}
			respData, _ := sonic.Marshal(result)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			res, err := dva.GetDataStart(ctx, viewID, incKey, nil, limit)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.TotalCount, ShouldEqual, 100)
		})

		Convey("Success with AccountInfo in context", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "test-account-id",
				Type: "test-account-type",
				Name: "Test Account",
			}
			ctxWithAccount := context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			result := interfaces.ViewQueryResult{
				TotalCount:  100,
				SearchAfter: []any{"value1"},
				Entries:     []map[string]any{{"id": "1", "name": "test"}},
			}
			respData, _ := sonic.Marshal(result)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			res, err := dva.GetDataStart(ctxWithAccount, viewID, incKey, incValue, limit)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.TotalCount, ShouldEqual, 100)
		})
	})
}

func Test_dataViewAccess_GetDataNext(t *testing.T) {
	Convey("Test GetDataNext", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			UniQueryUrl: "http://test-uni-query",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		dva := newTestDataViewAccess(appSetting, mockHTTPClient)

		viewID := "test-view-id"
		searchAfter := []any{"value1"}
		limit := 10
		// httpUrl := "http://test-uni-query/data-views/test-view-id?timeout=5m"

		Convey("Success getting data next", func() {
			result := interfaces.ViewQueryResult{
				TotalCount:  100,
				SearchAfter: []any{"value2"},
				Entries:     []map[string]any{{"id": "2", "name": "test2"}},
			}
			respData, _ := sonic.Marshal(result)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			res, err := dva.GetDataNext(ctx, viewID, searchAfter, limit)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.TotalCount, ShouldEqual, 100)
		})

		Convey("HTTP request error", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, []byte(""), errors.New("network error"))

			result, err := dva.GetDataNext(ctx, viewID, searchAfter, limit)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Non-200 status code", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, []byte("error"), nil)

			result, err := dva.GetDataNext(ctx, viewID, searchAfter, limit)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Unmarshal result failed", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			result, err := dva.GetDataNext(ctx, viewID, searchAfter, limit)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("Success with AccountInfo in context", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "test-account-id",
				Type: "test-account-type",
				Name: "Test Account",
			}
			ctxWithAccount := context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			result := interfaces.ViewQueryResult{
				TotalCount:  100,
				SearchAfter: []any{"value2"},
				Entries:     []map[string]any{{"id": "2", "name": "test2"}},
			}
			respData, _ := sonic.Marshal(result)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			res, err := dva.GetDataNext(ctxWithAccount, viewID, searchAfter, limit)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res.TotalCount, ShouldEqual, 100)
		})
	})
}
