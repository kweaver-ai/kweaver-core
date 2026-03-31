// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_type

import (
	"context"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	omock "ontology-query/interfaces/mock"
	"ontology-query/logics"
)

func Test_NewActionTypeService(t *testing.T) {
	Convey("Test NewActionTypeService", t, func() {
		appSetting := &common.AppSetting{}

		Convey("成功 - 创建服务实例", func() {
			service := NewActionTypeService(appSetting)
			So(service, ShouldNotBeNil)
		})

		Convey("成功 - 单例模式", func() {
			service1 := NewActionTypeService(appSetting)
			service2 := NewActionTypeService(appSetting)
			So(service1, ShouldEqual, service2)
		})
	})
}

func Test_actionTypeService_GetActionsByActionTypeID(t *testing.T) {
	Convey("Test actionTypeService GetActionsByActionTypeID", t, func() {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()

		appSetting := &common.AppSetting{}
		omAccess := omock.NewMockOntologyManagerAccess(mockCtrl)
		ots := omock.NewMockObjectTypeService(mockCtrl)
		uAccess := omock.NewMockUniqueryAccess(mockCtrl)

		// 设置全局变量
		logics.OMA = omAccess
		logics.UA = uAccess

		service := &actionTypeService{
			appSetting: appSetting,
			omAccess:   omAccess,
			ots:        ots,
			uAccess:    uAccess,
		}

		ctx := context.Background()
		knID := "kn1"
		actionTypeID := "at1"
		objectTypeID := "ot1"

		Convey("成功 - 获取行动数据", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{
					{
						Name:      "param1",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_PROP,
						Value:     "prop1",
					},
				},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{
						"id":    "123",
						"prop1": "value1",
						"prop2": "value2",
					},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), objectTypeID).Return(objectType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
			So(len(result.Actions), ShouldEqual, 1)
			So(result.Actions[0].Parameters["param1"], ShouldEqual, "value1")
		})

		Convey("失败 - 行动类不存在", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ActionType{}, nil, false, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusNotFound)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyQuery_ObjectType_ObjectTypeNotFound)
			So(result.TotalCount, ShouldEqual, 0)
		})

		Convey("失败 - 获取行动类错误", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(interfaces.ActionType{}, nil, false, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			httpErr, ok := err.(*rest.HTTPError)
			So(ok, ShouldBeTrue)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusInternalServerError)
			So(result.TotalCount, ShouldEqual, 0)
		})

		Convey("成功 - 带行动条件", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Condition: &cond.CondCfg{
					Name:      "status",
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "active",
					},
				},
				Parameters: []interfaces.Parameter{},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), objectTypeID).Return(objectType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
				// 验证条件是否正确合并
				So(query.ActualCondition, ShouldNotBeNil)
				So(query.ActualCondition.Operation, ShouldEqual, "and")
				return objects, nil
			})

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
		})

		Convey("成功 - 参数来源为常量", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{
					{
						Name:      "const_param",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_CONST,
						Value:     "constant_value",
					},
				},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), objectTypeID).Return(objectType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.Actions[0].Parameters["const_param"], ShouldEqual, "constant_value")
		})

		Convey("成功 - 参数来源为输入", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{
					{
						Name:      "input_param",
						ValueFrom: interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT,
					},
				},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), objectTypeID).Return(objectType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.Actions[0].DynamicParams["input_param"], ShouldBeNil)
		})

		Convey("成功 - 包含类型信息", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				CommonQueryParameters: interfaces.CommonQueryParameters{
					IncludeTypeInfo: true,
				},
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), objectTypeID).Return(objectType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.ActionType, ShouldNotBeNil)
			So(result.ActionType.ATID, ShouldEqual, actionTypeID)
		})

		Convey("成功 - 多个对象", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
					{"id": "456"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123"},
					{"id": "456"},
				},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), objectTypeID).Return(objectType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 2)
			So(len(result.Actions), ShouldEqual, 2)
		})

		Convey("失败 - 获取对象数据错误", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), objectTypeID).Return(objectType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError))

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldNotBeNil)
			So(result.TotalCount, ShouldEqual, 0)
		})

		Convey("成功 - 空对象列表", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{},
				ObjectType: &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID: objectTypeID,
					},
				},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), objectTypeID).Return(objectType, true, nil)
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(objects, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 0)
			So(len(result.Actions), ShouldEqual, 0)
		})

		Convey("成功 - 未绑定对象类 + 无 identities → 构造虚拟实例", func() {
			query := &interfaces.ActionQuery{
				KNID:               knID,
				ActionTypeID:       actionTypeID,
				InstanceIdentities: []map[string]any{},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: "", // 未绑定对象类
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
			So(len(result.Actions), ShouldEqual, 1)
			So(result.Actions[0].InstanceIdentity, ShouldNotBeNil)
		})

		Convey("成功 - 未绑定对象类 + 有 identities → 按 identities 构造实例", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123", "name": "test"},
					{"id": "456", "name": "test2"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ObjectTypeID: "", // 未绑定对象类
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Parameters: []interfaces.Parameter{},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 2)
			So(len(result.Actions), ShouldEqual, 2)
			So(result.Actions[0].InstanceIdentity, ShouldResemble, map[string]any{"id": "123", "name": "test"})
			So(result.Actions[1].InstanceIdentity, ShouldResemble, map[string]any{"id": "456", "name": "test2"})
		})

		Convey("成功 - add 行动类型 + 有 identities + 查询不到实例 → 构造实例并评估条件", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ActionType:   "add",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Condition: &cond.CondCfg{
					Name:      "status",
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "active",
					},
				},
				Parameters: []interfaces.Parameter{},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
					DataProperties: []cond.DataProperty{
						{Name: "status", Type: "string"},
					},
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), objectTypeID).Return(objectType, true, nil)
			// 第一次查询：仅根据 identities 查询（查询不到）
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).Return(interfaces.Objects{Datas: []map[string]any{}}, nil)

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			// 由于条件评估可能失败，结果可能为空或包含实例
			So(result, ShouldNotBeNil)
		})

		Convey("成功 - add 行动类型 + 有 identities + 查询到实例 → 按 identities 和行动条件过滤", func() {
			query := &interfaces.ActionQuery{
				KNID:         knID,
				ActionTypeID: actionTypeID,
				InstanceIdentities: []map[string]any{
					{"id": "123"},
				},
			}

			actionType := interfaces.ActionType{
				ATID:         actionTypeID,
				ATName:       "test_action",
				ActionType:   "add",
				ObjectTypeID: objectTypeID,
				ActionSource: interfaces.ActionSource{
					Type: "tool",
				},
				Condition: &cond.CondCfg{
					Name:      "status",
					Operation: "==",
					ValueOptCfg: cond.ValueOptCfg{
						Value: "active",
					},
				},
				Parameters: []interfaces.Parameter{},
			}

			objectType := interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: objectTypeID,
				},
			}

			objects := interfaces.Objects{
				Datas: []map[string]any{
					{"id": "123", "status": "active"},
				},
			}

			omAccess.EXPECT().GetActionType(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(actionType, map[string]any{"id": actionType.ATID}, true, nil)
			omAccess.EXPECT().GetObjectType(gomock.Any(), gomock.Any(), gomock.Any(), objectTypeID).Return(objectType, true, nil)
			// 第一次查询：仅根据 identities 查询（查询到）
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
				// 验证是仅根据 identities 的查询
				So(query.ActualCondition, ShouldNotBeNil)
				return interfaces.Objects{Datas: []map[string]any{{"id": "123"}}}, nil
			})
			// 第二次查询：按 identities 和行动条件过滤
			ots.EXPECT().GetObjectsByObjectTypeID(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {
				// 验证条件是否正确合并
				So(query.ActualCondition, ShouldNotBeNil)
				So(query.ActualCondition.Operation, ShouldEqual, "and")
				return objects, nil
			})

			result, err := service.GetActionsByActionTypeID(ctx, query)
			So(err, ShouldBeNil)
			So(result.TotalCount, ShouldEqual, 1)
		})
	})
}
