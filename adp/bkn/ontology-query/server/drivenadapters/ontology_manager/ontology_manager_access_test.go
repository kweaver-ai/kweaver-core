// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package ontology_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	rmock "github.com/kweaver-ai/kweaver-go-lib/rest/mock"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	"ontology-query/interfaces"
)

// newTestOntologyManagerAccess 创建用于测试的 ontologyManagerAccess，允许注入 mock HTTP 客户端
func newTestOntologyManagerAccess(appSetting *common.AppSetting, httpClient rest.HTTPClient) *ontologyManagerAccess {
	return &ontologyManagerAccess{
		appSetting:         appSetting,
		ontologyManagerUrl: appSetting.BKNBackendUrl,
		httpClient:         httpClient,
	}
}

func Test_NewOntologyManagerAccess(t *testing.T) {
	Convey("Test NewOntologyManagerAccess", t, func() {
		appSetting := &common.AppSetting{
			BKNBackendUrl: "http://test-om",
		}

		Convey("成功 - 创建单例实例", func() {
			// 重置单例
			omAccessOnce = sync.Once{}
			omAccess = nil

			access1 := NewOntologyManagerAccess(appSetting)
			access2 := NewOntologyManagerAccess(appSetting)

			So(access1, ShouldNotBeNil)
			So(access2, ShouldNotBeNil)
			So(access1, ShouldEqual, access2) // 应该是同一个实例
		})
	})
}

func Test_ontologyManagerAccess_GetObjectType(t *testing.T) {
	Convey("Test ontologyManagerAccess GetObjectType", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			BKNBackendUrl: "http://test-om",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		oma := newTestOntologyManagerAccess(appSetting, mockHTTPClient)

		ctx := context.Background()
		knID := "kn1"
		branch := "main"
		otID := "ot1"

		Convey("成功 - 获取对象类信息", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: otID,
				},
			}
			response := struct {
				ObjectTypes []interfaces.ObjectType `json:"entries"`
			}{
				ObjectTypes: []interfaces.ObjectType{objectType},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, exists, err := oma.GetObjectType(ctx, knID, branch, otID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(result.ObjectTypeWithKeyField.OTID, ShouldEqual, otID)
		})

		Convey("失败 - HTTP 请求错误", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, fmt.Errorf("http request failed"))

			result, exists, err := oma.GetObjectType(ctx, knID, branch, otID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.ObjectTypeWithKeyField.OTID, ShouldEqual, "")
		})

		Convey("失败 - 对象类不存在 (404)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNotFound, nil, nil)

			result, exists, err := oma.GetObjectType(ctx, knID, branch, otID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(result.ObjectTypeWithKeyField.OTID, ShouldEqual, "")
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
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, errorBytes, nil)

			result, exists, err := oma.GetObjectType(ctx, knID, branch, otID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.ObjectTypeWithKeyField.OTID, ShouldEqual, "")
		})

		Convey("失败 - HTTP 状态码非 200 且解析 BaseError 失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, invalidJSON, nil)

			result, exists, err := oma.GetObjectType(ctx, knID, branch, otID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.ObjectTypeWithKeyField.OTID, ShouldEqual, "")
		})

		Convey("失败 - 响应体为空", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			result, exists, err := oma.GetObjectType(ctx, knID, branch, otID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(result.ObjectTypeWithKeyField.OTID, ShouldEqual, "")
		})

		Convey("失败 - 响应体为空数组", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			response := struct {
				ObjectTypes []interfaces.ObjectType `json:"entries"`
			}{
				ObjectTypes: []interfaces.ObjectType{},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, exists, err := oma.GetObjectType(ctx, knID, branch, otID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(result.ObjectTypeWithKeyField.OTID, ShouldEqual, "")
		})

		Convey("失败 - 解析响应失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, invalidJSON, nil)

			result, exists, err := oma.GetObjectType(ctx, knID, branch, otID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.ObjectTypeWithKeyField.OTID, ShouldEqual, "")
		})

		Convey("成功 - 无账户信息", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: otID,
				},
			}
			response := struct {
				ObjectTypes []interfaces.ObjectType `json:"entries"`
			}{
				ObjectTypes: []interfaces.ObjectType{objectType},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, exists, err := oma.GetObjectType(ctx, knID, branch, otID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(result.ObjectTypeWithKeyField.OTID, ShouldEqual, otID)
		})
	})
}

func Test_ontologyManagerAccess_GetRelationType(t *testing.T) {
	Convey("Test ontologyManagerAccess GetRelationType", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			BKNBackendUrl: "http://test-om",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		oma := newTestOntologyManagerAccess(appSetting, mockHTTPClient)

		ctx := context.Background()
		knID := "kn1"
		branch := "main"
		rtID := "rt1"

		Convey("成功 - 获取关系类信息 (DIRECT类型)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			relationType := interfaces.RelationType{
				RTID: rtID,
				Type: interfaces.RELATION_TYPE_DIRECT,
				MappingRules: []interfaces.Mapping{
					{
						SourceProp: interfaces.SimpleProperty{Name: "field1"},
						TargetProp: interfaces.SimpleProperty{Name: "field2"},
					},
				},
			}
			response := struct {
				RelationTypes []interfaces.RelationType `json:"entries"`
			}{
				RelationTypes: []interfaces.RelationType{relationType},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, exists, err := oma.GetRelationType(ctx, knID, branch, rtID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(result.RTID, ShouldEqual, rtID)
			So(result.Type, ShouldEqual, interfaces.RELATION_TYPE_DIRECT)
		})

		Convey("成功 - 获取关系类信息 (DATA_VIEW类型)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			relationType := interfaces.RelationType{
				RTID: rtID,
				Type: interfaces.RELATION_TYPE_DATA_VIEW,
				MappingRules: interfaces.InDirectMapping{
					BackingDataSource: &interfaces.ResourceInfo{
						Type: "view",
						ID:   "view1",
					},
				},
			}
			response := struct {
				RelationTypes []interfaces.RelationType `json:"entries"`
			}{
				RelationTypes: []interfaces.RelationType{relationType},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, exists, err := oma.GetRelationType(ctx, knID, branch, rtID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(result.RTID, ShouldEqual, rtID)
			So(result.Type, ShouldEqual, interfaces.RELATION_TYPE_DATA_VIEW)
		})

		Convey("失败 - HTTP 请求错误", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, fmt.Errorf("http request failed"))

			result, exists, err := oma.GetRelationType(ctx, knID, branch, rtID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.RTID, ShouldEqual, "")
		})

		Convey("失败 - 关系类不存在 (404)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNotFound, nil, nil)

			result, exists, err := oma.GetRelationType(ctx, knID, branch, rtID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(result.RTID, ShouldEqual, "")
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
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, errorBytes, nil)

			result, exists, err := oma.GetRelationType(ctx, knID, branch, rtID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.RTID, ShouldEqual, "")
		})

		Convey("失败 - HTTP 状态码非 200 且解析 BaseError 失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, invalidJSON, nil)

			result, exists, err := oma.GetRelationType(ctx, knID, branch, rtID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.RTID, ShouldEqual, "")
		})

		Convey("失败 - 响应体为空", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			result, exists, err := oma.GetRelationType(ctx, knID, branch, rtID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(result.RTID, ShouldEqual, "")
		})

		Convey("失败 - 解析响应失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, invalidJSON, nil)

			result, exists, err := oma.GetRelationType(ctx, knID, branch, rtID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.RTID, ShouldEqual, "")
		})

		Convey("失败 - 响应体为空数组", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			response := struct {
				RelationTypes []interfaces.RelationType `json:"entries"`
			}{
				RelationTypes: []interfaces.RelationType{},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, exists, err := oma.GetRelationType(ctx, knID, branch, rtID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(result.RTID, ShouldEqual, "")
		})

	})
}

func Test_ontologyManagerAccess_GetActionType(t *testing.T) {
	Convey("Test ontologyManagerAccess GetActionType", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			BKNBackendUrl: "http://test-om",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		oma := newTestOntologyManagerAccess(appSetting, mockHTTPClient)

		ctx := context.Background()
		knID := "kn1"
		branch := "main"
		atID := "at1"

		Convey("成功 - 获取行动类信息", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			actionType := interfaces.ActionType{
				ATID: atID,
			}
			response := struct {
				ActionTypes []interfaces.ActionType `json:"entries"`
			}{
				ActionTypes: []interfaces.ActionType{actionType},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, rawSnapshot, exists, err := oma.GetActionType(ctx, knID, branch, atID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeTrue)
			So(result.ATID, ShouldEqual, atID)
			So(rawSnapshot, ShouldNotBeNil)
			So(rawSnapshot["id"], ShouldEqual, atID)
		})

		Convey("失败 - HTTP 请求错误", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(0, nil, fmt.Errorf("http request failed"))

			result, rawSnapshot, exists, err := oma.GetActionType(ctx, knID, branch, atID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.ATID, ShouldEqual, "")
			So(rawSnapshot, ShouldBeNil)
		})

		Convey("失败 - 行动类不存在 (404)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusNotFound, nil, nil)

			result, rawSnapshot, exists, err := oma.GetActionType(ctx, knID, branch, atID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(result.ATID, ShouldEqual, "")
			So(rawSnapshot, ShouldBeNil)
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
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, errorBytes, nil)

			result, rawSnapshot, exists, err := oma.GetActionType(ctx, knID, branch, atID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.ATID, ShouldEqual, "")
			So(rawSnapshot, ShouldBeNil)
		})

		Convey("失败 - HTTP 状态码非 200 且解析 BaseError 失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusBadRequest, invalidJSON, nil)

			result, rawSnapshot, exists, err := oma.GetActionType(ctx, knID, branch, atID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.ATID, ShouldEqual, "")
			So(rawSnapshot, ShouldBeNil)
		})

		Convey("失败 - 响应体为空", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, nil, nil)

			result, rawSnapshot, exists, err := oma.GetActionType(ctx, knID, branch, atID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(result.ATID, ShouldEqual, "")
			So(rawSnapshot, ShouldBeNil)
		})

		Convey("失败 - 解析响应失败", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			invalidJSON := []byte("invalid json")

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, invalidJSON, nil)

			result, rawSnapshot, exists, err := oma.GetActionType(ctx, knID, branch, atID)

			So(err, ShouldNotBeNil)
			So(exists, ShouldBeFalse)
			So(result.ATID, ShouldEqual, "")
			So(rawSnapshot, ShouldBeNil)
		})

		Convey("失败 - 响应体为空数组", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			response := struct {
				ActionTypes []interfaces.ActionType `json:"entries"`
			}{
				ActionTypes: []interfaces.ActionType{},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				GetNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, rawSnapshot, exists, err := oma.GetActionType(ctx, knID, branch, atID)

			So(err, ShouldBeNil)
			So(exists, ShouldBeFalse)
			So(result.ATID, ShouldEqual, "")
			So(rawSnapshot, ShouldBeNil)
		})
	})
}

func Test_ontologyManagerAccess_GetRelationTypePathsBaseOnSource(t *testing.T) {
	Convey("Test ontologyManagerAccess GetRelationTypePathsBaseOnSource", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{
			BKNBackendUrl: "http://test-om",
		}
		mockHTTPClient := rmock.NewMockHTTPClient(mockCtrl)
		oma := newTestOntologyManagerAccess(appSetting, mockHTTPClient)

		ctx := context.Background()
		knID := "kn1"
		branch := "main"
		query := interfaces.PathsQueryBaseOnSource{
			SourceObjecTypeId: "ot1",
			Direction:         interfaces.DIRECTION_FORWARD,
			PathLength:        2,
		}

		Convey("成功 - 获取关系类路径", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			typePath := interfaces.RelationTypePath{
				ID: 0,
				TypeEdges: []interfaces.TypeEdge{
					{
						RelationType: interfaces.RelationType{
							RTID: "rt1",
							Type: interfaces.RELATION_TYPE_DIRECT,
							MappingRules: []interfaces.Mapping{
								{
									SourceProp: interfaces.SimpleProperty{Name: "field1"},
									TargetProp: interfaces.SimpleProperty{Name: "field2"},
								},
							},
						},
					},
				},
			}
			response := struct {
				TypePaths []interfaces.RelationTypePath `json:"entries"`
			}{
				TypePaths: []interfaces.RelationTypePath{typePath},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := oma.GetRelationTypePathsBaseOnSource(ctx, knID, branch, query)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0].ID, ShouldEqual, 0)
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

			result, err := oma.GetRelationTypePathsBaseOnSource(ctx, knID, branch, query)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功 - 返回空路径列表", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			response := struct {
				TypePaths []interfaces.RelationTypePath `json:"entries"`
			}{
				TypePaths: []interfaces.RelationTypePath{},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := oma.GetRelationTypePathsBaseOnSource(ctx, knID, branch, query)

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
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

			result, err := oma.GetRelationTypePathsBaseOnSource(ctx, knID, branch, query)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
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

			result, err := oma.GetRelationTypePathsBaseOnSource(ctx, knID, branch, query)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
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

			result, err := oma.GetRelationTypePathsBaseOnSource(ctx, knID, branch, query)

			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
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

			result, err := oma.GetRelationTypePathsBaseOnSource(ctx, knID, branch, query)

			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功 - 获取关系类路径 (DATA_VIEW类型)", func() {
			accountInfo := interfaces.AccountInfo{
				ID:   "account1",
				Type: "user",
			}
			ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

			typePath := interfaces.RelationTypePath{
				ID: 0,
				TypeEdges: []interfaces.TypeEdge{
					{
						RelationType: interfaces.RelationType{
							RTID: "rt1",
							Type: interfaces.RELATION_TYPE_DATA_VIEW,
							MappingRules: interfaces.InDirectMapping{
								BackingDataSource: &interfaces.ResourceInfo{
									Type: "view",
									ID:   "view1",
								},
							},
						},
					},
				},
			}
			response := struct {
				TypePaths []interfaces.RelationTypePath `json:"entries"`
			}{
				TypePaths: []interfaces.RelationTypePath{typePath},
			}
			responseBytes, _ := sonic.Marshal(response)

			mockHTTPClient.EXPECT().
				PostNoUnmarshal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(http.StatusOK, responseBytes, nil)

			result, err := oma.GetRelationTypePathsBaseOnSource(ctx, knID, branch, query)

			So(err, ShouldBeNil)
			So(len(result), ShouldEqual, 1)
			So(result[0].ID, ShouldEqual, 0)
			So(result[0].TypeEdges[0].RelationType.Type, ShouldEqual, interfaces.RELATION_TYPE_DATA_VIEW)
		})
	})
}
