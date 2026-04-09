// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package permission

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

func newTestPermissionAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *permissionAccess {
	return &permissionAccess{
		appSetting:    appSetting,
		permissionUrl: appSetting.PermissionUrl,
		httpClient:    httpClient,
	}
}

func TestNewPermissionAccess(t *testing.T) {
	Convey("Test NewPermissionAccess", t, func() {
		appSetting := &common.AppSetting{
			PermissionUrl: "http://test-permission",
		}

		access1 := NewPermissionAccess(appSetting)
		access2 := NewPermissionAccess(appSetting)

		Convey("Should return singleton instance", func() {
			So(access1, ShouldNotBeNil)
			So(access2, ShouldEqual, access1)
		})
	})
}

func Test_permissionAccess_CheckPermission(t *testing.T) {
	Convey("Test CheckPermission", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			PermissionUrl: "http://test-permission",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		pa := newTestPermissionAccess(appSetting, mockHTTPClient)

		check := interfaces.PermissionCheck{
			Accessor: interfaces.PermissionAccessor{
				ID:   "user1",
				Type: interfaces.ACCESSOR_TYPE_USER,
			},
			Resource: interfaces.PermissionResource{
				ID:   "res1",
				Type: interfaces.RESOURCE_TYPE_KN,
			},
			Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
		}
		// httpUrl := "http://test-permission/operation-check"

		Convey("Success checking permission - allowed", func() {
			result := interfaces.PermissionCheckResult{
				Result: true,
			}
			respData, _ := sonic.Marshal(result)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			allowed, err := pa.CheckPermission(ctx, check)
			So(err, ShouldBeNil)
			So(allowed, ShouldBeTrue)
		})

		Convey("Success checking permission - denied", func() {
			result := interfaces.PermissionCheckResult{
				Result: false,
			}
			respData, _ := sonic.Marshal(result)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			allowed, err := pa.CheckPermission(ctx, check)
			So(err, ShouldBeNil)
			So(allowed, ShouldBeFalse)
		})

		Convey("HTTP request error", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, []byte(""), errors.New("network error"))

			allowed, err := pa.CheckPermission(ctx, check)
			So(err, ShouldNotBeNil)
			So(allowed, ShouldBeFalse)
		})

		Convey("Null response body", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			allowed, err := pa.CheckPermission(ctx, check)
			So(err, ShouldBeNil)
			So(allowed, ShouldBeFalse)
		})

		Convey("HTTP status not OK with valid error response", func() {
			permissionError := PermissionError{
				Code:        "PERMISSION_DENIED",
				Description: "Permission denied",
				Cause:       "User does not have permission",
			}
			respData, _ := sonic.Marshal(permissionError)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusForbidden, respData, nil)

			allowed, err := pa.CheckPermission(ctx, check)
			So(err, ShouldNotBeNil)
			So(allowed, ShouldBeFalse)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusForbidden)
		})

		Convey("HTTP status not OK with invalid error response", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusForbidden, []byte("invalid json"), nil)

			allowed, err := pa.CheckPermission(ctx, check)
			So(err, ShouldNotBeNil)
			So(allowed, ShouldBeFalse)
		})

		Convey("Unmarshal result failed", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			allowed, err := pa.CheckPermission(ctx, check)
			So(err, ShouldNotBeNil)
			So(allowed, ShouldBeFalse)
		})
	})
}

func Test_permissionAccess_CreateResources(t *testing.T) {
	Convey("Test CreateResources", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			PermissionUrl: "http://test-permission",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		pa := newTestPermissionAccess(appSetting, mockHTTPClient)

		policies := []interfaces.PermissionPolicy{
			{
				Accessor: interfaces.PermissionAccessor{
					ID:   "user1",
					Type: interfaces.ACCESSOR_TYPE_USER,
				},
				Resource: interfaces.PermissionResource{
					ID:   "res1",
					Type: interfaces.RESOURCE_TYPE_KN,
				},
			},
		}
		// httpUrl := "http://test-permission/policy"

		Convey("Success creating resources", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNoContent, []byte(""), nil)

			err := pa.CreateResources(ctx, policies)
			So(err, ShouldBeNil)
		})

		Convey("HTTP request error", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, []byte(""), errors.New("network error"))

			err := pa.CreateResources(ctx, policies)
			So(err, ShouldNotBeNil)
		})

		Convey("HTTP status not NoContent with valid error response", func() {
			permissionError := PermissionError{
				Code:        "CREATE_FAILED",
				Description: "Create failed",
				Cause:       "Resource already exists",
			}
			respData, _ := sonic.Marshal(permissionError)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, respData, nil)

			err := pa.CreateResources(ctx, policies)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("HTTP status not NoContent with invalid error response", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, []byte("invalid json"), nil)

			err := pa.CreateResources(ctx, policies)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_permissionAccess_DeleteResources(t *testing.T) {
	Convey("Test DeleteResources", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			PermissionUrl: "http://test-permission",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		pa := newTestPermissionAccess(appSetting, mockHTTPClient)

		resources := []interfaces.PermissionResource{
			{
				ID:   "res1",
				Type: interfaces.RESOURCE_TYPE_KN,
			},
		}
		// httpUrl := "http://test-permission/policy-delete"

		Convey("Success deleting resources", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNoContent, []byte(""), nil)

			err := pa.DeleteResources(ctx, resources)
			So(err, ShouldBeNil)
		})

		Convey("HTTP request error", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, []byte(""), errors.New("network error"))

			err := pa.DeleteResources(ctx, resources)
			So(err, ShouldNotBeNil)
		})

		Convey("HTTP status not NoContent with valid error response", func() {
			permissionError := PermissionError{
				Code:        "DELETE_FAILED",
				Description: "Delete failed",
				Cause:       "Resource not found",
			}
			respData, _ := sonic.Marshal(permissionError)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNotFound, respData, nil)

			err := pa.DeleteResources(ctx, resources)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("HTTP status not NoContent with invalid error response", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, []byte("invalid json"), nil)

			err := pa.DeleteResources(ctx, resources)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_permissionAccess_FilterResources(t *testing.T) {
	Convey("Test FilterResources", t, func() {
		ctx := context.Background()
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			PermissionUrl: "http://test-permission",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		pa := newTestPermissionAccess(appSetting, mockHTTPClient)

		filter := interfaces.PermissionResourcesFilter{
			Accessor: interfaces.PermissionAccessor{
				ID:   "user1",
				Type: interfaces.ACCESSOR_TYPE_USER,
			},
			Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
		}
		// httpUrl := "http://test-permission/resource-filter"

		Convey("Success filtering resources", func() {
			result := []interfaces.PermissionResourceOps{
				{
					ResourceID: "res1",
					Operations: []string{interfaces.OPERATION_TYPE_VIEW_DETAIL},
				},
			}
			respData, _ := sonic.Marshal(result)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, respData, nil)

			ops, err := pa.FilterResources(ctx, filter)
			So(err, ShouldBeNil)
			So(ops, ShouldNotBeNil)
			So(len(ops), ShouldEqual, 1)
		})

		Convey("HTTP request error", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, []byte(""), errors.New("network error"))

			ops, err := pa.FilterResources(ctx, filter)
			So(err, ShouldNotBeNil)
			So(len(ops), ShouldEqual, 0)
		})

		Convey("HTTP status not OK with valid error response", func() {
			permissionError := PermissionError{
				Code:        "FILTER_FAILED",
				Description: "Filter failed",
				Cause:       "Invalid filter parameters",
			}
			respData, _ := sonic.Marshal(permissionError)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, respData, nil)

			ops, err := pa.FilterResources(ctx, filter)
			So(err, ShouldNotBeNil)
			So(len(ops), ShouldEqual, 0)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("HTTP status not OK with invalid error response", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, []byte("invalid json"), nil)

			ops, err := pa.FilterResources(ctx, filter)
			So(err, ShouldNotBeNil)
			So(len(ops), ShouldEqual, 0)
		})

		Convey("Null response body", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			ops, err := pa.FilterResources(ctx, filter)
			So(err, ShouldBeNil)
			So(len(ops), ShouldEqual, 0)
		})

		Convey("Unmarshal result failed", func() {
			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, []byte("invalid json"), nil)

			ops, err := pa.FilterResources(ctx, filter)
			So(err, ShouldNotBeNil)
			So(len(ops), ShouldEqual, 0)
		})
	})
}
