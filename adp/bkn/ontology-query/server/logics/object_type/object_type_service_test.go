// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package object_type

import (
	"context"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	omock "ontology-query/interfaces/mock"
	"ontology-query/logics"
)

func Test_NewObjectTypeService(t *testing.T) {
	Convey("Test NewObjectTypeService", t, func() {
		appSetting := &common.AppSetting{}

		Convey("成功 - 创建服务实例", func() {
			service := NewObjectTypeService(appSetting)
			So(service, ShouldNotBeNil)
		})

		Convey("成功 - 单例模式", func() {
			service1 := NewObjectTypeService(appSetting)
			service2 := NewObjectTypeService(appSetting)
			So(service1, ShouldEqual, service2)
		})
	})
}

func Test_objectTypeService_GetObjectsByObjectTypeID(t *testing.T) {
	Convey("Test objectTypeService GetObjectsByObjectTypeID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		omAccess := omock.NewMockOntologyManagerAccess(mockCtrl)
		osa := omock.NewMockOpenSearchAccess(mockCtrl)
		uAccess := omock.NewMockUniqueryAccess(mockCtrl)
		mfa := omock.NewMockModelFactoryAccess(mockCtrl)
		aoAccess := omock.NewMockAgentOperatorAccess(mockCtrl)

		logics.OMA = omAccess
		logics.OSA = osa
		logics.UA = uAccess
		logics.MFA = mfa
		logics.AOA = aoAccess

		service := &objectTypeService{
			appSetting: appSetting,
			omAccess:   omAccess,
			osa:        osa,
			uAccess:    uAccess,
			mfa:        mfa,
			aoAccess:   aoAccess,
		}

		ctx := context.Background()
		knID := "kn1"
		branch := "main"
		objectTypeID := "ot1"

		Convey("失败 - 对象类不存在", func() {
			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ObjectType{}, false, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
			So(result.Datas, ShouldBeNil)
		})

		Convey("失败 - 获取对象类错误", func() {
			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ObjectType{}, false, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.Datas, ShouldBeNil)
		})

		Convey("失败 - 无效的排序字段", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{Name: "prop1"},
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				PageQuery: interfaces.PageQuery{
					Sort: []*interfaces.SortParams{
						{Field: "invalid_field"},
					},
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(result.Datas, ShouldBeNil)
		})

		Convey("失败 - 无效的属性", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{Name: "prop1"},
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				Properties:   []string{"invalid_prop"},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(result.Datas, ShouldBeNil)
		})

		Convey("成功 - 从索引查询", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{Name: "prop1"},
					},
					PrimaryKeys: []string{"id"},
				},
				Status: &interfaces.ObjectTypeStatus{
					IndexAvailable: true,
					Index:          "index1",
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				PageQuery: interfaces.PageQuery{
					Limit:     10,
					NeedTotal: true,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.Hit{
				{
					Source: map[string]interface{}{
						"prop1": "value1",
					},
					Score: 1.0,
				},
			}, nil)
			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(`{"count":1}`), nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.SearchFromIndex, ShouldBeTrue)
			So(len(result.Datas), ShouldEqual, 1)
		})

		Convey("失败 - 从索引查询时GetTotal失败", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{Name: "prop1"},
					},
					PrimaryKeys: []string{"id"},
				},
				Status: &interfaces.ObjectTypeStatus{
					IndexAvailable: true,
					Index:          "index1",
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				PageQuery: interfaces.PageQuery{
					Limit:     10,
					NeedTotal: true,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.Hit{
				{
					Source: map[string]interface{}{
						"prop1": "value1",
					},
					Score: 1.0,
				},
			}, nil)
			osa.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.Datas, ShouldBeNil)
		})

		Convey("成功 - 从索引查询不需要总数", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{Name: "prop1"},
					},
					PrimaryKeys: []string{"id"},
				},
				Status: &interfaces.ObjectTypeStatus{
					IndexAvailable: true,
					Index:          "index1",
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				PageQuery: interfaces.PageQuery{
					Limit:     10,
					NeedTotal: false,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return([]interfaces.Hit{
				{
					Source: map[string]interface{}{
						"prop1": "value1",
					},
					Score: 1.0,
				},
			}, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.SearchFromIndex, ShouldBeTrue)
			So(len(result.Datas), ShouldEqual, 1)
		})

		Convey("成功 - 从视图查询", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					PrimaryKeys: []string{"id"},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{"field1": "value1"},
				},
				TotalCount: 1,
			}, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.SearchFromIndex, ShouldBeFalse)
			So(len(result.Datas), ShouldEqual, 1)
		})

		Convey("失败 - 重写条件失败", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					PrimaryKeys: []string{"id"},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				ActualCondition: &cond.CondCfg{
					Name:      "invalid_field",
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "value1",
					},
				},
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(result.Datas, ShouldBeNil)
		})

		Convey("成功 - 包含逻辑属性", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
					PrimaryKeys: []string{"id"},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name: "logic_prop1",
							Type: interfaces.LOGIC_PROPERTY_TYPE_METRIC,
							DataSource: &interfaces.ResourceInfo{
								Type: interfaces.LOGIC_PROPERTY_TYPE_METRIC,
								ID:   "metric1",
							},
							Parameters: []interfaces.Parameter{
								{
									Name:      "param1",
									ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
									Value:     "prop1",
								},
							},
						},
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				CommonQueryParameters: interfaces.CommonQueryParameters{
					IncludeLogicParams: true,
				},
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{"field1": "value1"},
				},
				TotalCount: 1,
			}, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			So(result.Datas[0]["logic_prop1"], ShouldNotBeNil)
		})

		Convey("失败 - 唯一标识缺少主键字段", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{Name: "prop1"},
					},
					PrimaryKeys: []string{"id", "name"},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				ObjectQueryInfo: &interfaces.ObjectQueryInfo{
					InstanceIdentity: []map[string]any{
						{"id": "123"}, // 缺少 name
					},
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(result.Datas, ShouldBeNil)
		})

		Convey("失败 - 属性查询的属性不存在", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{Name: "prop1"},
					},
					PrimaryKeys: []string{"id"},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				ObjectQueryInfo: &interfaces.ObjectQueryInfo{
					InstanceIdentity: []map[string]any{
						{"id": "123"},
					},
					Properties: []string{"invalid_prop"},
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(result.Datas, ShouldBeNil)
		})

		Convey("成功 - IgnoringStore=true时从视图查询", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					PrimaryKeys: []string{"id"},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
				Status: &interfaces.ObjectTypeStatus{
					IndexAvailable: true,
					Index:          "index1",
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				CommonQueryParameters: interfaces.CommonQueryParameters{
					IgnoringStore: true,
				},
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{"field1": "value1"},
				},
				TotalCount: 1,
			}, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.SearchFromIndex, ShouldBeFalse)
			So(len(result.Datas), ShouldEqual, 1)
		})

		Convey("失败 - 视图数据源为空", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					PrimaryKeys: []string{"id"},
					DataSource:  nil, // 数据源为空
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(result.Datas, ShouldBeNil)
		})

		Convey("失败 - 视图数据源ID为空", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					PrimaryKeys: []string{"id"},
					DataSource: &interfaces.ResourceInfo{
						ID: "", // ID为空
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
			So(result.Datas, ShouldBeNil)
		})

		Convey("失败 - 获取视图数据失败", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					PrimaryKeys: []string{"id"},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ViewData{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.Datas, ShouldBeNil)
		})

		Convey("失败 - 从索引查询失败", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{Name: "prop1"},
					},
					PrimaryKeys: []string{"id"},
				},
				Status: &interfaces.ObjectTypeStatus{
					IndexAvailable: true,
					Index:          "index1",
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			osa.EXPECT().SearchData(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.Datas, ShouldBeNil)
		})

		Convey("成功 - IncludeTypeInfo=true", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					PrimaryKeys: []string{"id"},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				CommonQueryParameters: interfaces.CommonQueryParameters{
					IncludeTypeInfo: true,
				},
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{"field1": "value1"},
				},
				TotalCount: 1,
			}, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.ObjectType, ShouldNotBeNil)
			So(result.ObjectType.OTID, ShouldEqual, objectTypeID)
		})

		Convey("成功 - 包含LOGIC_PARAMS_VALUE_FROM_CONST的逻辑属性", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
					PrimaryKeys: []string{"id"},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name: "logic_prop1",
							Type: interfaces.LOGIC_PROPERTY_TYPE_METRIC,
							DataSource: &interfaces.ResourceInfo{
								ID: "metric1",
							},
							Parameters: []interfaces.Parameter{
								{
									Name:      "param1",
									ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
									Value:     "const_value",
								},
							},
						},
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				CommonQueryParameters: interfaces.CommonQueryParameters{
					IncludeLogicParams: true,
				},
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{"field1": "value1"},
				},
				TotalCount: 1,
			}, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			So(result.Datas[0]["logic_prop1"], ShouldNotBeNil)
		})

		Convey("成功 - 包含LOGIC_PARAMS_VALUE_FROM_INPUT的逻辑属性", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
					PrimaryKeys: []string{"id"},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name: "logic_prop1",
							Type: interfaces.LOGIC_PROPERTY_TYPE_METRIC,
							DataSource: &interfaces.ResourceInfo{
								ID: "metric1",
							},
							Parameters: []interfaces.Parameter{
								{
									Name:      "param1",
									ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
								},
							},
						},
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				CommonQueryParameters: interfaces.CommonQueryParameters{
					IncludeLogicParams: true,
				},
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{"field1": "value1"},
				},
				TotalCount: 1,
			}, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			So(result.Datas[0]["logic_prop1"], ShouldNotBeNil)
		})

		Convey("成功 - 包含LOGIC_PROPERTY_TYPE_OPERATOR的逻辑属性", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
					PrimaryKeys: []string{"id"},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name: "logic_prop1",
							Type: interfaces.LOGIC_PROPERTY_TYPE_OPERATOR,
							DataSource: &interfaces.ResourceInfo{
								ID: "operator1",
							},
							Parameters: []interfaces.Parameter{
								{
									Name:      "param1",
									ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
									Value:     "prop1",
								},
							},
						},
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				CommonQueryParameters: interfaces.CommonQueryParameters{
					IncludeLogicParams: true,
				},
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{"field1": "value1"},
				},
				TotalCount: 1,
			}, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			So(result.Datas[0]["logic_prop1"], ShouldNotBeNil)
		})

		Convey("成功 - 不支持的逻辑属性类型", func() {
			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{
							Name: "prop1",
							MappedField: cond.Field{
								Name: "field1",
							},
						},
					},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
					PrimaryKeys: []string{"id"},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name: "logic_prop1",
							Type: "unsupported_type",
							DataSource: &interfaces.ResourceInfo{
								ID: "resource1",
							},
						},
					},
				},
			}

			query := &interfaces.ObjectQueryBaseOnObjectType{
				KNID:         knID,
				Branch:       branch,
				ObjectTypeID: objectTypeID,
				CommonQueryParameters: interfaces.CommonQueryParameters{
					IncludeLogicParams: true,
				},
				PageQuery: interfaces.PageQuery{
					Limit: 10,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{"field1": "value1"},
				},
				TotalCount: 1,
			}, nil)

			result, err := service.GetObjectsByObjectTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			// 不支持的逻辑属性类型不会添加到结果中
		})
	})
}

func Test_objectTypeService_GetTotal(t *testing.T) {
	Convey("Test objectTypeService GetTotal", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		service := &objectTypeService{
			osa: omock.NewMockOpenSearchAccess(mockCtrl),
		}

		ctx := context.Background()
		index := "index1"
		dsl := map[string]any{
			"query": map[string]any{
				"match_all": map[string]any{},
			},
			"from": 0,
			"size": 10,
			"sort": []any{},
		}

		Convey("成功 - 获取总数", func() {
			mockOSA := service.osa.(*omock.MockOpenSearchAccess)
			mockOSA.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(`{"count":100}`), nil)

			result, err := service.GetTotal(ctx, index, dsl)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 100)
		})

		Convey("失败 - Count错误", func() {
			mockOSA := service.osa.(*omock.MockOpenSearchAccess)
			mockOSA.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.GetTotal(ctx, index, dsl)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})

		Convey("失败 - 无效JSON", func() {
			mockOSA := service.osa.(*omock.MockOpenSearchAccess)
			mockOSA.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(`invalid json`), nil)

			result, err := service.GetTotal(ctx, index, dsl)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})

		Convey("失败 - 获取count字段失败", func() {
			mockOSA := service.osa.(*omock.MockOpenSearchAccess)
			mockOSA.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(`{"total":100}`), nil)

			result, err := service.GetTotal(ctx, index, dsl)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})

		Convey("失败 - 转换为int64失败", func() {
			mockOSA := service.osa.(*omock.MockOpenSearchAccess)
			mockOSA.EXPECT().Count(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(`{"count":"not_a_number"}`), nil)

			result, err := service.GetTotal(ctx, index, dsl)
			So(err, ShouldNotBeNil)
			So(result, ShouldEqual, 0)
		})
	})
}

func Test_objectTypeService_GetObjectPropertyValue(t *testing.T) {
	Convey("Test objectTypeService GetObjectPropertyValue", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		omAccess := omock.NewMockOntologyManagerAccess(mockCtrl)
		osa := omock.NewMockOpenSearchAccess(mockCtrl)
		uAccess := omock.NewMockUniqueryAccess(mockCtrl)
		mfa := omock.NewMockModelFactoryAccess(mockCtrl)
		aoAccess := omock.NewMockAgentOperatorAccess(mockCtrl)

		logics.OMA = omAccess
		logics.OSA = osa
		logics.UA = uAccess
		logics.MFA = mfa
		logics.AOA = aoAccess

		service := &objectTypeService{
			appSetting: appSetting,
			omAccess:   omAccess,
			osa:        osa,
			uAccess:    uAccess,
			mfa:        mfa,
			aoAccess:   aoAccess,
		}

		ctx := context.Background()

		Convey("成功 - 获取对象属性值", func() {
			query := &interfaces.ObjectPropertyValueQuery{
				KNID:         "kn1",
				Branch:       "main",
				ObjectTypeID: "ot1",
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
				Properties: []string{"prop1"},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataProperties: []cond.DataProperty{
						{Name: "id", MappedField: cond.Field{Name: "id"}},
						{Name: "prop1", MappedField: cond.Field{Name: "prop1"}},
					},
					PrimaryKeys: []string{"id"},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
			}

			// GetObjectPropertyValue 内部会调用 GetObjectsByObjectTypeID
			// GetObjectsByObjectTypeID 需要这些依赖
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{"id": "123", "prop1": "value1"},
				},
				TotalCount: 1,
			}, nil)

			result, err := service.GetObjectPropertyValue(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			So(result.Datas[0]["prop1"], ShouldEqual, "value1")
			So(result.Datas[0]["id"], ShouldEqual, "123")
		})

		Convey("成功 - 包含逻辑属性", func() {
			query := &interfaces.ObjectPropertyValueQuery{
				KNID:         "kn1",
				Branch:       "main",
				ObjectTypeID: "ot1",
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
				Properties: []string{"prop1", "logic_prop1"},
				DynamicParams: map[string]map[string]any{
					"logic_prop1": {
						"start":   1234567890,
						"end":     1234567890,
						"instant": true,
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataProperties: []cond.DataProperty{
						{Name: "id", MappedField: cond.Field{Name: "id"}},
						{Name: "prop1", MappedField: cond.Field{Name: "prop1"}},
					},
					PrimaryKeys: []string{"id"},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name: "logic_prop1",
							Type: interfaces.LOGIC_PROPERTY_TYPE_METRIC,
							Parameters: []interfaces.Parameter{
								{
									Name:      "param1",
									ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
									Value:     "prop1",
								},
							},
							DataSource: &interfaces.ResourceInfo{
								ID: "metric1",
							},
						},
					},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{
						"id":    "123",
						"prop1": "value1",
						"logic_prop1": interfaces.MetricProperty{
							PropertyType:    interfaces.LOGIC_PROPERTY_TYPE_METRIC,
							MappingSourceId: "metric1",
							Parameters: interfaces.MetricFilters{
								Filters: []interfaces.Filter{
									{
										Name:      "param1",
										Operation: "==",
										Value:     "value1",
									},
								},
							},
							DynamicParams: map[string]any{},
						},
					},
				},
				TotalCount: 1,
			}, nil)
			uAccess.EXPECT().GetMetricDataByID(gomock.Any(), "metric1", gomock.Any()).Return(interfaces.MetricData{
				Datas: []interfaces.Data{
					{
						Labels: map[string]string{"param1": "value1"},
						Values: []interface{}{100},
						Times:  []interface{}{1234567890},
					},
				},
			}, nil)

			result, err := service.GetObjectPropertyValue(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
			So(result.Datas[0]["prop1"], ShouldEqual, "value1")
		})

		Convey("失败 - 获取对象错误", func() {
			query := &interfaces.ObjectPropertyValueQuery{
				KNID:         "kn1",
				Branch:       "main",
				ObjectTypeID: "ot1",
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
				Properties: []string{"prop1"},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ObjectType{}, false, nil)

			result, err := service.GetObjectPropertyValue(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("成功 - 包含逻辑属性且处理成功", func() {
			query := &interfaces.ObjectPropertyValueQuery{
				KNID:         "kn1",
				Branch:       "main",
				ObjectTypeID: "ot1",
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
				Properties: []string{"logic_prop1"},
				DynamicParams: map[string]map[string]any{
					"logic_prop1": {
						"start":   int64(1234567890),
						"end":     int64(1234567890),
						"instant": true,
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataProperties: []cond.DataProperty{
						{Name: "id", MappedField: cond.Field{Name: "id"}},
					},
					PrimaryKeys: []string{"id"},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name: "logic_prop1",
							Type: interfaces.LOGIC_PROPERTY_TYPE_METRIC,
							Parameters: []interfaces.Parameter{
								{
									Name:      "param1",
									ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
									Value:     "id",
								},
							},
							DataSource: &interfaces.ResourceInfo{
								ID: "metric1",
							},
						},
					},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{
						"id": "123",
						"logic_prop1": interfaces.MetricProperty{
							PropertyType:    interfaces.LOGIC_PROPERTY_TYPE_METRIC,
							MappingSourceId: "metric1",
							Parameters: interfaces.MetricFilters{
								Filters: []interfaces.Filter{
									{
										Name:      "param1",
										Operation: "==",
										Value:     "123",
									},
								},
							},
							DynamicParams: map[string]any{
								"start":   int64(1234567890),
								"end":     int64(1234567890),
								"instant": true,
							},
						},
					},
				},
				TotalCount: 1,
			}, nil)
			uAccess.EXPECT().GetMetricDataByID(gomock.Any(), "metric1", gomock.Any()).Return(interfaces.MetricData{
				Datas: []interfaces.Data{
					{
						Labels: map[string]string{"param1": "123"},
						Values: []interface{}{100},
						Times:  []interface{}{1234567890},
					},
				},
			}, nil)

			result, err := service.GetObjectPropertyValue(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
		})

		Convey("成功 - 包含算子类型逻辑属性", func() {
			query := &interfaces.ObjectPropertyValueQuery{
				KNID:         "kn1",
				Branch:       "main",
				ObjectTypeID: "ot1",
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
				Properties: []string{"logic_prop1"},
				DynamicParams: map[string]map[string]any{
					"logic_prop1": {
						"param1": "value1",
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataProperties: []cond.DataProperty{
						{Name: "id", MappedField: cond.Field{Name: "id"}},
					},
					PrimaryKeys: []string{"id"},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name: "logic_prop1",
							Type: interfaces.LOGIC_PROPERTY_TYPE_OPERATOR,
							Parameters: []interfaces.Parameter{
								{
									Name:      "param1",
									ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
									Source:    interfaces.PARAMETER_BODY,
								},
							},
							DataSource: &interfaces.ResourceInfo{
								ID: "operator1",
							},
						},
					},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
			}

			operatorInfo := interfaces.AgentOperator{
				OperatorId: "operator1",
				OperatorInfo: interfaces.OperatorInfo{
					ExecutionMode: interfaces.OPERATOR_EXECUTION_MODE_SYNC,
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{
						"id": "123",
						"logic_prop1": interfaces.OperatorProperty{
							PropertyType:    interfaces.LOGIC_PROPERTY_TYPE_OPERATOR,
							MappingSourceId: "operator1",
							Parameters:      map[string]any{},
							DynamicParams: map[string]any{
								"param1": "value1",
							},
						},
					},
				},
				TotalCount: 1,
			}, nil)
			aoAccess.EXPECT().GetAgentOperatorByID(gomock.Any(), "operator1").Return(operatorInfo, nil)
			aoAccess.EXPECT().ExecuteOperator(gomock.Any(), "operator1", gomock.Any()).Return(map[string]any{"result": "success"}, nil)

			result, err := service.GetObjectPropertyValue(ctx, query)
			So(err, ShouldBeNil)
			So(len(result.Datas), ShouldEqual, 1)
		})

		Convey("失败 - 算子执行模式不是同步", func() {
			query := &interfaces.ObjectPropertyValueQuery{
				KNID:         "kn1",
				Branch:       "main",
				ObjectTypeID: "ot1",
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
				Properties: []string{"logic_prop1"},
				DynamicParams: map[string]map[string]any{
					"logic_prop1": {
						"param1": "value1",
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataProperties: []cond.DataProperty{
						{Name: "id", MappedField: cond.Field{Name: "id"}},
					},
					PrimaryKeys: []string{"id"},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name: "logic_prop1",
							Type: interfaces.LOGIC_PROPERTY_TYPE_OPERATOR,
							Parameters: []interfaces.Parameter{
								{
									Name:      "param1",
									ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
									Source:    interfaces.PARAMETER_BODY,
								},
							},
							DataSource: &interfaces.ResourceInfo{
								ID: "operator1",
							},
						},
					},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
			}

			operatorInfo := interfaces.AgentOperator{
				OperatorId: "operator1",
				OperatorInfo: interfaces.OperatorInfo{
					ExecutionMode: "async", // 异步模式
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{
						"id": "123",
						"logic_prop1": interfaces.OperatorProperty{
							PropertyType:    interfaces.LOGIC_PROPERTY_TYPE_OPERATOR,
							MappingSourceId: "operator1",
							Parameters:      map[string]any{},
							DynamicParams: map[string]any{
								"param1": "value1",
							},
						},
					},
				},
				TotalCount: 1,
			}, nil)
			aoAccess.EXPECT().GetAgentOperatorByID(gomock.Any(), "operator1").Return(operatorInfo, nil)

			result, err := service.GetObjectPropertyValue(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})

		Convey("失败 - 算子缺少动态参数", func() {
			query := &interfaces.ObjectPropertyValueQuery{
				KNID:         "kn1",
				Branch:       "main",
				ObjectTypeID: "ot1",
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
				Properties:    []string{"logic_prop1"},
				DynamicParams: map[string]map[string]any{}, // 缺少动态参数
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataProperties: []cond.DataProperty{
						{Name: "id", MappedField: cond.Field{Name: "id"}},
					},
					PrimaryKeys: []string{"id"},
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name: "logic_prop1",
							Type: interfaces.LOGIC_PROPERTY_TYPE_OPERATOR,
							Parameters: []interfaces.Parameter{
								{
									Name:      "param1",
									ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
									Source:    interfaces.PARAMETER_BODY,
								},
							},
							DataSource: &interfaces.ResourceInfo{
								ID: "operator1",
							},
						},
					},
					DataSource: &interfaces.ResourceInfo{
						ID: "view1",
					},
				},
			}

			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(objectType, true, nil)
			uAccess.EXPECT().GetViewDataByID(gomock.Any(), "view1", gomock.Any()).Return(interfaces.ViewData{
				Datas: []map[string]any{
					{
						"id": "123",
						"logic_prop1": interfaces.OperatorProperty{
							PropertyType:    interfaces.LOGIC_PROPERTY_TYPE_OPERATOR,
							MappingSourceId: "operator1",
							Parameters:      map[string]any{},
							DynamicParams: map[string]any{
								"param1": "value1", // 需要动态参数
							},
						},
					},
				},
				TotalCount: 1,
			}, nil)

			result, err := service.GetObjectPropertyValue(ctx, query)
			So(err, ShouldNotBeNil)
			So(len(result.Datas), ShouldEqual, 0)
		})
	})
}

func Test_getNestedValue(t *testing.T) {
	Convey("Test getNestedValue", t, func() {
		Convey("成功 - 简单字段", func() {
			data := map[string]any{
				"key1": "value1",
			}
			result := getNestedValue(data, "key1")
			So(result, ShouldEqual, "value1")
		})

		Convey("成功 - 嵌套字段", func() {
			data := map[string]any{
				"level1": map[string]any{
					"level2": "value",
				},
			}
			result := getNestedValue(data, "level1.level2")
			So(result, ShouldEqual, "value")
		})

		Convey("成功 - data为nil", func() {
			result := getNestedValue(nil, "key1")
			So(result, ShouldBeNil)
		})

		Convey("成功 - 字段不存在", func() {
			data := map[string]any{}
			result := getNestedValue(data, "nonexistent")
			So(result, ShouldBeNil)
		})

		Convey("成功 - 嵌套路径不存在", func() {
			data := map[string]any{
				"level1": "not_a_map",
			}
			result := getNestedValue(data, "level1.level2")
			So(result, ShouldBeNil)
		})
	})
}

func Test_setNestedValue(t *testing.T) {
	Convey("Test setNestedValue", t, func() {
		Convey("成功 - 简单字段", func() {
			target := make(map[string]any)
			setNestedValue(target, "key1", "value1")
			So(target["key1"], ShouldEqual, "value1")
		})

		Convey("成功 - 嵌套字段", func() {
			target := make(map[string]any)
			setNestedValue(target, "level1.level2", "value")
			So(target["level1"], ShouldNotBeNil)
			level1, ok := target["level1"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(level1["level2"], ShouldEqual, "value")
		})

		Convey("成功 - value为nil", func() {
			target := make(map[string]any)
			setNestedValue(target, "key1", nil)
			_, exists := target["key1"]
			So(exists, ShouldBeFalse)
		})

		Convey("成功 - 深层嵌套", func() {
			target := make(map[string]any)
			setNestedValue(target, "a.b.c", "value")
			a, _ := target["a"].(map[string]any)
			b, _ := a["b"].(map[string]any)
			So(b["c"], ShouldEqual, "value")
		})
	})
}

func Test_generateExecRequest(t *testing.T) {
	Convey("Test generateExecRequest", t, func() {
		Convey("成功 - Header参数", func() {
			configParams := []interfaces.Parameter{
				{
					Name:      "header_param",
					Source:    interfaces.PARAMETER_HEADER,
					ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
				},
			}
			parameters := map[string]any{
				"header_param": "header_value",
			}
			dynamicParams := map[string]any{}

			result := generateExecRequest(configParams, parameters, dynamicParams)
			So(result.Header["header_param"], ShouldEqual, "header_value")
		})

		Convey("成功 - Query参数", func() {
			configParams := []interfaces.Parameter{
				{
					Name:      "query_param",
					Source:    interfaces.PARAMETER_QUERY,
					ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
				},
			}
			parameters := map[string]any{
				"query_param": "query_value",
			}
			dynamicParams := map[string]any{}

			result := generateExecRequest(configParams, parameters, dynamicParams)
			So(result.Query["query_param"], ShouldEqual, "query_value")
		})

		Convey("成功 - Body参数", func() {
			configParams := []interfaces.Parameter{
				{
					Name:      "body_param",
					Source:    interfaces.PARAMETER_BODY,
					ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
				},
			}
			parameters := map[string]any{
				"body_param": "body_value",
			}
			dynamicParams := map[string]any{}

			result := generateExecRequest(configParams, parameters, dynamicParams)
			So(result.Body["body_param"], ShouldEqual, "body_value")
		})

		Convey("成功 - Path参数", func() {
			configParams := []interfaces.Parameter{
				{
					Name:      "path_param",
					Source:    interfaces.PARAMETER_PATH,
					ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
				},
			}
			parameters := map[string]any{
				"path_param": "path_value",
			}
			dynamicParams := map[string]any{}

			result := generateExecRequest(configParams, parameters, dynamicParams)
			So(result.Path["path_param"], ShouldEqual, "path_value")
		})

		Convey("成功 - 动态参数", func() {
			configParams := []interfaces.Parameter{
				{
					Name:      "dynamic_param",
					Source:    interfaces.PARAMETER_BODY,
					ValueFrom: interfaces.VALUE_FROM_INPUT,
				},
			}
			parameters := map[string]any{}
			dynamicParams := map[string]any{
				"dynamic_param": "dynamic_value",
			}

			result := generateExecRequest(configParams, parameters, dynamicParams)
			So(result.Body["dynamic_param"], ShouldEqual, "dynamic_value")
		})

		Convey("成功 - 嵌套参数", func() {
			configParams := []interfaces.Parameter{
				{
					Name:      "nested.param",
					Source:    interfaces.PARAMETER_BODY,
					ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
				},
			}
			parameters := map[string]any{
				"nested": map[string]any{
					"param": "nested_value",
				},
			}
			dynamicParams := map[string]any{}

			result := generateExecRequest(configParams, parameters, dynamicParams)
			nested, ok := result.Body["nested"].(map[string]any)
			So(ok, ShouldBeTrue)
			So(nested["param"], ShouldEqual, "nested_value")
		})
	})
}
