// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package user_mgmt

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

func newTestUserMgmtAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *userMgmtAccess {
	return &userMgmtAccess{
		appSetting:  appSetting,
		httpClient:  httpClient,
		userMgmtUrl: appSetting.UserMgmtUrl,
	}
}

func TestNewUserMgmtAccess(t *testing.T) {
	Convey("Test NewUserMgmtAccess", t, func() {
		appSetting := &common.AppSetting{
			UserMgmtUrl: "http://test-user-mgmt",
		}

		access1 := NewUserMgmtAccess(appSetting)
		access2 := NewUserMgmtAccess(appSetting)

		Convey("Should return singleton instance", func() {
			So(access1, ShouldNotBeNil)
			So(access2, ShouldEqual, access1)
		})
	})
}

func Test_userMgmtAccess_GetAccountNames(t *testing.T) {
	Convey("Test GetAccountNames", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			UserMgmtUrl: "http://test-user-mgmt",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		uma := newTestUserMgmtAccess(appSetting, mockHTTPClient)

		httpUrl := "http://test-user-mgmt/api/user-management/v2/names"

		Convey("Success getting account names", func() {
			accountInfos := []*interfaces.AccountInfo{
				{ID: "user1", Type: interfaces.ACCESSOR_TYPE_USER},
				{ID: "app1", Type: interfaces.ACCESSOR_TYPE_APP},
			}

			response := map[string]any{
				"user_names": []map[string]string{
					{"id": "user1", "name": "User One"},
				},
				"app_names": []map[string]string{
					{"id": "app1", "name": "App One"},
				},
			}
			respData, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(ctx, httpUrl, gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			err := uma.GetAccountNames(ctx, accountInfos)
			So(err, ShouldBeNil)
			So(accountInfos[0].Name, ShouldEqual, "User One")
			So(accountInfos[1].Name, ShouldEqual, "App One")
		})

		Convey("Empty account infos", func() {
			err := uma.GetAccountNames(ctx, []*interfaces.AccountInfo{})
			So(err, ShouldBeNil)
		})

		Convey("HTTP request error", func() {
			accountInfos := []*interfaces.AccountInfo{
				{ID: "user1", Type: interfaces.ACCESSOR_TYPE_USER},
			}

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(ctx, httpUrl, gomock.Any(), gomock.Any()).
				Return(0, []byte(""), errors.New("network error"))

			err := uma.GetAccountNames(ctx, accountInfos)
			So(err, ShouldNotBeNil)
		})

		Convey("Non-200 status code", func() {
			accountInfos := []*interfaces.AccountInfo{
				{ID: "user1", Type: interfaces.ACCESSOR_TYPE_USER},
			}

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(ctx, httpUrl, gomock.Any(), gomock.Any()).
				Return(http.StatusInternalServerError, []byte("error"), nil)

			err := uma.GetAccountNames(ctx, accountInfos)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid response format", func() {
			accountInfos := []*interfaces.AccountInfo{
				{ID: "user1", Type: interfaces.ACCESSOR_TYPE_USER},
			}

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(ctx, httpUrl, gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			err := uma.GetAccountNames(ctx, accountInfos)
			So(err, ShouldNotBeNil)
		})
	})
}
