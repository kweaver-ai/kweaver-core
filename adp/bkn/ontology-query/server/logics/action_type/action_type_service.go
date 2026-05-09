// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package action_type

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	"ontology-query/logics"
	"ontology-query/logics/object_type"
)

var (
	atServiceOnce sync.Once
	atService     interfaces.ActionTypeService
)

type actionTypeService struct {
	appSetting *common.AppSetting
	omAccess   interfaces.OntologyManagerAccess
	ots        interfaces.ObjectTypeService
	uAccess    interfaces.UniqueryAccess
}

func NewActionTypeService(appSetting *common.AppSetting) interfaces.ActionTypeService {
	atServiceOnce.Do(func() {
		atService = &actionTypeService{
			appSetting: appSetting,
			omAccess:   logics.OMA,
			ots:        object_type.NewObjectTypeService(appSetting),
			uAccess:    logics.UA,
		}
	})
	return atService
}

func (ats *actionTypeService) GetActionsByActionTypeID(ctx context.Context,
	query *interfaces.ActionQuery) (interfaces.Actions, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询行动类的行动数据")
	defer span.End()

	var resps interfaces.Actions

	// 1. 先获取行动类信息
	actionType, _, exists, err := ats.omAccess.GetActionType(ctx, query.KNID, query.Branch, query.ActionTypeID)
	if err != nil {
		logger.Errorf("Get Action Type error: %s", err.Error())

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("at_id").String(query.ActionTypeID))
		span.SetStatus(codes.Error, "Get Action Type error")
		span.End()
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get Action Type error: %v", err))

		return resps, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_GetObjectTypesByIDFailed).WithErrorDetails(err.Error())
	}
	if !exists {
		logger.Debugf("Action Type %d not found!", query.ActionTypeID)

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(query.ActionTypeID))
		span.SetStatus(codes.Error, "Action Type not found!")
		span.End()
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Action Type [%s] not found!", query.ActionTypeID))

		return resps, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_ObjectType_ObjectTypeNotFound)
	}

	if missing := logics.MissingActionInputDynamicParamNames(&actionType, query.DynamicParams); len(missing) > 0 {
		return resps, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyQuery_ActionType_InvalidParameter_DynamicParams).
			WithErrorDetails(fmt.Sprintf("当前请求查询的行动类[%s]所需的动态参数未完整传入，缺少参数 %v，请在请求的 dynamic_params 中填充", actionType.ATName, missing))
	}

	// 2. 检查是否绑定了对象类
	isObjectTypeBound := actionType.ObjectTypeID != ""
	var objectType interfaces.ObjectType

	if isObjectTypeBound {
		// 获取对象类信息（用于条件评估）
		var exists bool
		var err error
		objectType, exists, err = ats.omAccess.GetObjectType(ctx, query.KNID, query.Branch, actionType.ObjectTypeID)
		if err != nil {
			logger.Errorf("Get Object Type error: %s", err.Error())
			return resps, rest.NewHTTPError(ctx, http.StatusInternalServerError,
				oerrors.OntologyQuery_ObjectType_InternalError_GetObjectTypesByIDFailed).WithErrorDetails(err.Error())
		}
		if !exists {
			logger.Debugf("Object Type %s not found!", actionType.ObjectTypeID)
			return resps, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_ObjectType_ObjectTypeNotFound)
		}
	} else {
		// 未绑定对象类的情况
		logger.Infof("Action type %s has no bound object type", actionType.ATID)
		if len(query.InstanceIdentities) == 0 {
			// Case 4: 未绑定对象类 + 无 identities → 构造一个临时的虚拟实例
			logger.Infof("No identities provided, creating virtual instance for action type %s", actionType.ATID)
			virtualAction, err := buildActionFromInstanceData(map[string]any{}, &actionType, query.DynamicParams)
			if err != nil {
				logger.Errorf("Error building virtual action: %v", err)
				return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
					WithErrorDetails(err.Error())
			}

			respActions := interfaces.Actions{
				ActionSource: actionType.ActionSource,
				Actions:      []interfaces.ActionParam{virtualAction},
				TotalCount:   1,
			}

			if query.IncludeTypeInfo {
				respActions.ActionType = &actionType
			}

			return respActions, nil
		} else {
			// Case 5: 未绑定对象类 + 有 identities → 按 identities 构造实例
			logger.Infof("Constructing instances from identities for action type %s", actionType.ATID)
			actions := []interfaces.ActionParam{}
			for _, identity := range query.InstanceIdentities {
				action, err := buildActionFromInstanceData(identity, &actionType, query.DynamicParams)
				if err != nil {
					logger.Errorf("Error building action from instance data: %v", err)
					return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
						WithErrorDetails(err.Error())
				}

				actions = append(actions, action)
			}

			respActions := interfaces.Actions{
				ActionSource: actionType.ActionSource,
				Actions:      actions,
				TotalCount:   len(actions),
			}

			if query.IncludeTypeInfo {
				respActions.ActionType = &actionType
			}

			return respActions, nil
		}
	}

	// 3. 处理 add 行动类型的特殊逻辑
	if actionType.ActionType == "add" && len(query.InstanceIdentities) > 0 {
		// 先仅根据 _instance_identities 查询对象实例（不包含行动条件）
		instanceCondition := logics.BuildInstanceIdentitiesCondition(query.InstanceIdentities)
		instanceQuery := &interfaces.ObjectQueryBaseOnObjectType{
			ActualCondition: instanceCondition,
			PageQuery: interfaces.PageQuery{
				Limit:     interfaces.MAX_LIMIT,
				NeedTotal: true,
			},
			KNID:         query.KNID,
			Branch:       query.Branch,
			ObjectTypeID: actionType.ObjectTypeID,
			CommonQueryParameters: interfaces.CommonQueryParameters{
				IncludeTypeInfo:         true,
				IncludeLogicParams:      query.IncludeLogicParams,
				ExcludeSystemProperties: query.ExcludeSystemProperties,
			},
			ObjectQueryInfo: &interfaces.ObjectQueryInfo{
				InstanceIdentity: query.InstanceIdentities,
			},
		}
		instanceObjects, err := ats.ots.GetObjectsByObjectTypeID(ctx, instanceQuery)
		if err != nil {
			return resps, err
		}

		// 如果查询结果为空，将 _instance_identities 视为新实例，评估是否满足行动条件
		if len(instanceObjects.Datas) == 0 {
			// Case 2a: 都搜索不到，则按identites构造实例，再套用行动条件，满足，产生实例
			logger.Infof("No instances found by identities for add action, constructing instances and evaluating condition")
			actions := []interfaces.ActionParam{}
			for _, instanceIdentity := range query.InstanceIdentities {
				// 评估实例是否满足行动条件
				if actionType.Condition != nil {
					satisfies, err := logics.EvaluateInstanceAgainstCondition(ctx, instanceIdentity, actionType.Condition, &objectType)
					if err != nil {
						logger.Errorf("Error evaluating condition for instance[%v], error: %v", instanceIdentity, err)
						continue
					}
					if !satisfies {
						// 不满足条件，跳过
						continue
					}
				}

				// 满足条件，构造行动数据
				action, err := buildActionFromInstanceData(instanceIdentity, &actionType, query.DynamicParams)
				if err != nil {
					logger.Errorf("Error building action from instance data: %v", err)
					return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
						WithErrorDetails(err.Error())
				}

				actions = append(actions, action)
			}

			respActions := interfaces.Actions{
				ActionSource: actionType.ActionSource,
				Actions:      actions,
				TotalCount:   len(actions),
			}

			if query.IncludeTypeInfo {
				respActions.ActionType = &actionType
			}

			return respActions, nil
		}
		// Case 2b: 搜索得到，就按identites和行动条件过滤出来的实例（继续执行后续逻辑）
		logger.Infof("Instances found by identities for add action, filtering by identities and action condition")
	}

	// 4. 根据行动条件+请求的唯一标识，去请求对象类的对象实例数据（当前行动条件只能选绑定的对象类的，不能选其他类，所以当前就直接拼，认为这些条件都在作用在这个对象类上）
	// 条件转换，唯一标识换成主键过滤，各个对象之间用or连接，主键间用and连接，然后再跟行动条件and去请求对象类的对象数据
	// 可接受instance_identities为空
	condition := logics.BuildInstanceIdentitiesCondition(query.InstanceIdentities)

	if actionType.Condition != nil {
		condition = &cond.CondCfg{
			Operation: "and",
			SubConds:  []*cond.CondCfg{condition, actionType.Condition},
		}
	}

	// 5. 根据行动条件和唯一标识组成的条件检索起点对象类的对象实例
	objectQuery := &interfaces.ObjectQueryBaseOnObjectType{
		ActualCondition: condition,
		PageQuery: interfaces.PageQuery{
			Limit:     interfaces.MAX_LIMIT, // 不限制条数，要符合条件的所有,视图最大支持1w，所以就设置1w
			NeedTotal: true,
		},
		KNID:         query.KNID,
		Branch:       query.Branch,
		ObjectTypeID: actionType.ObjectTypeID,
		CommonQueryParameters: interfaces.CommonQueryParameters{
			IncludeTypeInfo:         true,
			IncludeLogicParams:      query.IncludeLogicParams,
			ExcludeSystemProperties: query.ExcludeSystemProperties,
		},
		ObjectQueryInfo: &interfaces.ObjectQueryInfo{
			InstanceIdentity: query.InstanceIdentities,
		},
	}
	objects, err := ats.ots.GetObjectsByObjectTypeID(ctx, objectQuery)
	if err != nil {
		return resps, err
	}

	// 6. 获得的对象是满足条件的对象，这些对象都应该实例化为行动
	actions := []interfaces.ActionParam{}
	for _, object := range objects.Datas {
		paramsJson := "{}"
		dynamicParamsJson := "{}"
		for _, param := range actionType.Parameters {
			switch param.ValueFrom {
			case interfaces.LOGIC_PARAMS_VALUE_FROM_PROP:
				value := object[param.Value.(string)]
				paramsJson, err = sjson.Set(paramsJson, param.Name, value)
				if err != nil {
					return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
						WithErrorDetails(fmt.Sprintf("Error setting action type[%s]'s parameter path %s: %v",
							actionType.ATName, param.Name, err.Error()))
				}
			case interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT:
				val := logics.ActionDynamicParamGetValue(query.DynamicParams, param.Name)
				dynamicParamsJson, err = sjson.Set(dynamicParamsJson, param.Name, val)
				if err != nil {
					return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
						WithErrorDetails(fmt.Sprintf("Error setting action type[%s]'s dynamic parameter path %s: %v",
							actionType.ATName, param.Name, err.Error()))
				}
			case interfaces.LOGIC_PARAMS_VALUE_FROM_CONST:
				paramsJson, err = sjson.Set(paramsJson, param.Name, param.Value)
				if err != nil {
					return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
						WithErrorDetails(fmt.Sprintf("Error setting action type[%s]'s parameter path %s: %v",
							actionType.ATName, param.Name, err.Error()))
				}
			}
		}
		params := map[string]any{}
		err = json.Unmarshal([]byte(paramsJson), &params)
		if err != nil {
			return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
				WithErrorDetails(fmt.Sprintf("failed to Unmarshal action type[%s]'s paramtersJson to map, %s",
					actionType.ATName, err.Error()))
		}

		dynamicParams := map[string]any{}
		err = json.Unmarshal([]byte(dynamicParamsJson), &dynamicParams)
		if err != nil {
			return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
				WithErrorDetails(fmt.Sprintf("failed to Unmarshal action type[%s]'s dynamicParamsJson to map, %s",
					actionType.ATName, err.Error()))
		}

		action := interfaces.ActionParam{
			Parameters:    params,
			DynamicParams: dynamicParams,
		}

		// 已经在对象数据查询是指定了排除字段，返回的已经是按排除字段处理后的数据，所以字段存在就添加。
		if _, exist := object[interfaces.SYSTEM_PROPERTY_INSTANCE_ID]; exist {
			action.InstanceID = object[interfaces.SYSTEM_PROPERTY_INSTANCE_ID]
		}
		if _, exist := object[interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY]; exist {
			action.InstanceIdentity = object[interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY]
		}
		if _, exist := object[interfaces.SYSTEM_PROPERTY_DISPLAY]; exist {
			action.Display = object[interfaces.SYSTEM_PROPERTY_DISPLAY]
		}

		// 返回的对象数据已经按查询参数生成和排除系统字段了，此时就是按需添加
		actions = append(actions, action)
	}

	respActions := interfaces.Actions{
		ActionSource: actionType.ActionSource,
		Actions:      actions,
		TotalCount:   len(actions),
	}

	if query.IncludeTypeInfo {
		respActions.ActionType = &actionType
	}

	return respActions, nil
}

// buildActionFromInstanceData builds action data from instance data
func buildActionFromInstanceData(instanceData map[string]any,
	actionType *interfaces.ActionType, requestDynamicParams map[string]any) (interfaces.ActionParam, error) {

	var action interfaces.ActionParam

	paramsJson := "{}"
	dynamicParamsJson := "{}"
	var err error

	for _, param := range actionType.Parameters {
		switch param.ValueFrom {
		case interfaces.LOGIC_PARAMS_VALUE_FROM_PROP:
			propName, ok := param.Value.(string)
			if !ok {
				return action, fmt.Errorf("parameter %s value_from is property but value is not string", param.Name)
			}
			value := instanceData[propName]
			paramsJson, err = sjson.Set(paramsJson, param.Name, value)
			if err != nil {
				return action, fmt.Errorf("error setting action type[%s]'s parameter path %s: %v",
					actionType.ATName, param.Name, err)
			}
		case interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT:
			val := logics.ActionDynamicParamGetValue(requestDynamicParams, param.Name)
			dynamicParamsJson, err = sjson.Set(dynamicParamsJson, param.Name, val)
			if err != nil {
				return action, fmt.Errorf("error setting action type[%s]'s dynamic parameter path %s: %v",
					actionType.ATName, param.Name, err)
			}
		case interfaces.LOGIC_PARAMS_VALUE_FROM_CONST:
			paramsJson, err = sjson.Set(paramsJson, param.Name, param.Value)
			if err != nil {
				return action, fmt.Errorf("error setting action type[%s]'s parameter path %s: %v",
					actionType.ATName, param.Name, err)
			}
		}
	}

	params := map[string]any{}
	err = json.Unmarshal([]byte(paramsJson), &params)
	if err != nil {
		return action, fmt.Errorf("failed to Unmarshal action type[%s]'s paramtersJson to map, %s",
			actionType.ATName, err.Error())
	}

	dynamicParams := map[string]any{}
	err = json.Unmarshal([]byte(dynamicParamsJson), &dynamicParams)
	if err != nil {
		return action, fmt.Errorf("failed to Unmarshal action type[%s]'s dynamicParamsJson to map, %s",
			actionType.ATName, err.Error())
	}

	action = interfaces.ActionParam{
		Parameters:    params,
		DynamicParams: dynamicParams,
	}

	// Set instance identity from instanceData
	if identity, exist := instanceData[interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY]; exist {
		action.InstanceIdentity = identity
	} else {
		// If not found, construct from primary keys
		identityMap := make(map[string]any)
		for k, v := range instanceData {
			identityMap[k] = v
		}
		action.InstanceIdentity = identityMap
	}

	return action, nil
}
