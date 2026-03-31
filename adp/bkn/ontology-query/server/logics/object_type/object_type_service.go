// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package object_type

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/tidwall/sjson"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	"ontology-query/logics"
)

var (
	otServiceOnce sync.Once
	otService     interfaces.ObjectTypeService
)

type objectTypeService struct {
	appSetting *common.AppSetting
	aoAccess   interfaces.AgentOperatorAccess
	mfa        interfaces.ModelFactoryAccess
	omAccess   interfaces.OntologyManagerAccess
	osa        interfaces.OpenSearchAccess
	uAccess    interfaces.UniqueryAccess
}

func NewObjectTypeService(appSetting *common.AppSetting) interfaces.ObjectTypeService {
	otServiceOnce.Do(func() {
		otService = &objectTypeService{
			appSetting: appSetting,
			aoAccess:   logics.AOA,
			mfa:        logics.MFA,
			omAccess:   logics.OMA,
			osa:        logics.OSA,
			uAccess:    logics.UA,
		}
	})
	return otService
}

func (ots *objectTypeService) GetObjectsByObjectTypeID(ctx context.Context,
	query *interfaces.ObjectQueryBaseOnObjectType) (interfaces.Objects, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询对象类的对象数据")
	defer span.End()

	start := time.Now().UnixMilli()

	var resps interfaces.Objects

	objectType, exists, err := ots.omAccess.GetObjectType(ctx, query.KNID, query.Branch, query.ObjectTypeID)
	if err != nil {
		logger.Errorf("Get Object Type error: %s", err.Error())

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(query.ObjectTypeID))
		span.SetStatus(codes.Error, "Get Object Type error")
		span.End()
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get Object Type error: %v", err))

		return resps, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_GetObjectTypesByIDFailed).WithErrorDetails(err.Error())
	}
	if !exists {
		logger.Debugf("Object Type %d not found!", query.ObjectTypeID)

		// 添加异常时的 trace 属性
		span.SetAttributes(attribute.Key("model_id").String(query.ObjectTypeID))
		span.SetStatus(codes.Error, "Object Type not found!")
		span.End()
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Object Type [%s] not found!", query.ObjectTypeID))

		return resps, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_ObjectType_ObjectTypeNotFound)
	}

	// /排序字段可以是对象类的数据属性, _score

	// 3.1 处理对象类，转成view field 到 object type property的映射
	// 视图字段到对象类属性的映射
	viewFieldPropMap := map[string]string{
		interfaces.SORT_FIELD_SCORE: interfaces.SORT_FIELD_SCORE, //  _score 字段
	}
	// 对象类属性名到属性名的映射，便于case到索引查询时使用。对象索引的字段名与属性名保持一直
	indexPropMap := map[string]string{
		interfaces.SORT_FIELD_SCORE: interfaces.SORT_FIELD_SCORE, //  _score 字段
	}
	// 对象类数据属性名到对象类数据属性的映射
	propMap := map[string]cond.DataProperty{}
	for _, prop := range objectType.DataProperties {
		propMap[prop.Name] = prop
		if len(query.Properties) == 0 { // 未指定属性集时，认为是拿全部属性
			viewFieldPropMap[prop.MappedField.Name] = prop.Name
			indexPropMap[prop.Name] = prop.Name
		} else {
			for _, requestProp := range query.Properties {
				if prop.Name == requestProp {
					viewFieldPropMap[prop.MappedField.Name] = prop.Name
					indexPropMap[prop.Name] = prop.Name
				}
			}
		}
	}

	// 排序字段非空时，排序字段必须是对象类的数据属性, _score
	if len(query.Sort) > 0 {
		for _, sp := range query.Sort {
			if _, exists := indexPropMap[sp.Field]; !exists {
				return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("排序字段[%s]不是对象类的数据属性", sp.Field))
			}
		}
	}
	// 指定的属性集需在对象类的数据属性中存在
	if len(query.Properties) > 0 {
		for _, prop := range query.Properties {
			if _, exists := propMap[prop]; !exists {
				return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("指定的属性[%s]不是对象类的数据属性", prop))
			}
		}
	}

	// 对于数据属性的查询的参数校验
	if query.ObjectQueryInfo != nil {
		// 唯一标识包含主键字段
		for i, instanceIdentity := range query.ObjectQueryInfo.InstanceIdentity {
			for _, key := range objectType.PrimaryKeys {
				if _, exist := instanceIdentity[key]; !exist {
					return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
						WithErrorDetails(fmt.Sprintf("第%d个对象的实例标识字段[%s]不能为空", i+1, key))
				}
			}
		}
		// 属性列表包含对象类的数据属性和逻辑属性
		logicPropMap := make(map[string]bool)
		for _, prop := range objectType.LogicProperties {
			logicPropMap[prop.Name] = true
		}
		for _, prop := range query.ObjectQueryInfo.Properties {
			if _, exist := propMap[prop]; !exist && !logicPropMap[prop] {
				return resps, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("属性查询的属性[%s]不是对象类的属性", prop))
			}
		}
	}

	if !query.IgnoringStore && objectType.Status != nil && objectType.Status.IndexAvailable {
		// 2. 构造排序字段
		if query.Sort == nil {
			// 给默认值, 默认按 _score desc，主键 asc
			query.Sort = logics.BuildIndexSort(objectType, propMap)
		}

		// 持久化查询,转成dsl,直接查 opensearch
		err = ots.getObjectsFromObjectIndex(ctx, query, objectType, &resps, indexPropMap)
		if err != nil {
			return resps, err
		}
		// 在response里显示走的是持久化查询
		resps.SearchFromIndex = true
	} else {
		// 2. 构造排序字段
		if query.Sort == nil {
			// 给默认值, 默认按 _score desc，主键 asc
			query.Sort = logics.BuildViewSort(objectType)
		}
		// 3. 请求视图获取数据，获取指定字段
		err = ots.getObjectsFromDataView(ctx, query, objectType, &resps, viewFieldPropMap)
		if err != nil {
			return resps, err
		}
	}

	// 4. 组装逻辑属性
	if query.IncludeLogicParams && len(objectType.LogicProperties) > 0 {
		// 逐个对象处理对象的逻辑属性,并把逻辑属性设置到对象上
		err = ots.processLogicProperties(ctx, &resps, objectType)
		if err != nil {
			return resps, err
		}
	}

	// resps.Datas = objects

	if query.IncludeTypeInfo {
		resps.ObjectType = &objectType
	}

	logger.Debugf("从对象类[%s]中获取到的数据条数为[%d],耗时: %dms", objectType.OTID, len(resps.Datas), time.Now().UnixMilli()-start)

	return resps, nil
}

// 逐个对象处理对象的逻辑属性,并把逻辑属性设置到对象上
func (*objectTypeService) processLogicProperties(ctx context.Context, resps *interfaces.Objects,
	objectType interfaces.ObjectType) error {

	var err error

	// 逐个对象处理对象的逻辑属性,并把逻辑属性设置到对象上
	for i, object := range resps.Datas {
		// loop logic prop
		for _, logicProp := range objectType.LogicProperties {
			switch logicProp.Type {
			case interfaces.LOGIC_PROPERTY_TYPE_METRIC:
				filters := []interfaces.Filter{}
				dynamicParams := map[string]any{}
				for _, param := range logicProp.Parameters {
					switch param.ValueFrom {
					case interfaces.LOGIC_PARAMS_VALUE_FROM_PROP:
						value := object[param.Value.(string)]
						filters = append(filters, interfaces.Filter{
							Name:      param.Name,
							Operation: param.Operation,
							Value:     value,
						})
					case interfaces.LOGIC_PARAMS_VALUE_FROM_CONST:
						// 固定参数且
						filters = append(filters, interfaces.Filter{
							Name:      param.Name,
							Operation: "==",
							Value:     param.Value,
						})
					case interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT:
						dynamicParams[param.Name] = param
					}
				}

				mProp := interfaces.MetricProperty{
					PropertyType:    logicProp.Type,
					MappingSourceId: logicProp.DataSource.ID,
					Parameters: interfaces.MetricFilters{
						Filters: filters,
					},
					DynamicParams: dynamicParams,
				}
				resps.Datas[i][logicProp.Name] = mProp

			case interfaces.LOGIC_PROPERTY_TYPE_OPERATOR:
				paramsJson := "{}"
				dynamicParamsJson := "{}"
				for _, param := range logicProp.Parameters {
					switch param.ValueFrom {
					case interfaces.LOGIC_PARAMS_VALUE_FROM_PROP:
						value := object[param.Value.(string)]
						paramsJson, err = sjson.Set(paramsJson, param.Name, value)
						if err != nil {
							return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
								WithErrorDetails(fmt.Sprintf("Error setting logic property[%s]'s parameter path %s: %v",
									logicProp.Name, param.Name, err.Error()))
						}

					case interfaces.LOGIC_PARAMS_VALUE_FROM_CONST:
						paramsJson, err = sjson.Set(paramsJson, param.Name, param.Value)
						if err != nil {
							return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
								WithErrorDetails(fmt.Sprintf("Error setting logic property[%s]'s parameter path %s: %v",
									logicProp.Name, param.Name, err.Error()))
						}
					case interfaces.LOGIC_PARAMS_VALUE_FROM_INPUT:
						dynamicParamsJson, err = sjson.Set(dynamicParamsJson, param.Name, param)
						if err != nil {
							return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
								WithErrorDetails(fmt.Sprintf("Error setting logic property[%s]'s dynamic parameter path %s: %v",
									logicProp.Name, param.Name, err.Error()))
						}
					}
				}
				params := map[string]any{}
				err = json.Unmarshal([]byte(paramsJson), &params)
				if err != nil {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
						WithErrorDetails(fmt.Sprintf("failed to Unmarshal logic property[%s]'s paramtersJson to map, %s",
							logicProp.Name, err.Error()))
				}

				dynamicParams := map[string]any{}
				err = json.Unmarshal([]byte(dynamicParamsJson), &dynamicParams)
				if err != nil {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
						WithErrorDetails(fmt.Sprintf("failed to Unmarshal logic property[%s]'s dynamicParamsJson to map, %s",
							logicProp.Name, err.Error()))
				}

				oProp := interfaces.OperatorProperty{
					PropertyType:    logicProp.Type,
					MappingSourceId: logicProp.DataSource.ID,
					Parameters:      params,
					DynamicParams:   dynamicParams,
				}
				resps.Datas[i][logicProp.Name] = oProp

			default:
				logger.Warnf("系统支持的逻辑属性类型有[metric, operator],当前请求的逻辑属性类型为[%s]，请求将不返回逻辑属性的计算参数", logicProp.Type)
			}
		}
	}
	return nil
}

// 从视图中获取对象数据
func (ots *objectTypeService) getObjectsFromDataView(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType,
	objectType interfaces.ObjectType, resps *interfaces.Objects, fieldPropMap map[string]string) error {

	objects := []map[string]any{}
	// 构造视图的查询请求
	viewQuery := interfaces.ViewQuery{
		NeedTotal:         query.NeedTotal,
		Limit:             query.Limit,
		UseSearchAfter:    interfaces.USE_SEARCH_AFTER_TRUE,
		Sort:              query.Sort,
		SearchAfterParams: query.SearchAfterParams,
	}

	// 构造视图的过滤条件时,需要看请求的有没有knn,如果有knn,则需要对knn向量化后再请求视图
	// 重写过滤条件
	if query.ActualCondition != nil {
		rewriteCondition, err := cond.RewriteCondition(ctx, query.ActualCondition,
			logics.TransferPropsToPropMap(objectType.DataProperties),
			func(ctx context.Context, property *cond.DataProperty, word string) ([]cond.VectorResp, error) {

				return ots.handlerVector(ctx, property, word)
			})
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails(fmt.Sprintf("failed to rewrite ontology condition to view condition, %s", err.Error()))
		}
		viewQuery.Filters = rewriteCondition
	}

	if objectType.DataSource == nil || objectType.DataSource.ID == "" {
		// 视图为空，返回异常，不请求
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("对象类[%s]绑定的视图为空", objectType.OTID))
	}

	viewData, err := ots.uAccess.GetViewDataByID(ctx, objectType.DataSource.ID, viewQuery)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_GetViewDataByIDFailed).WithErrorDetails(err.Error())
	}

	// 3. 组装数据属性，处理视图数据到数据属性的映射
	// 3.2 loop view datas
	for _, col := range viewData.Datas {
		// 一行是一个对象
		object := map[string]any{}
		for k, v := range col {
			// k 是视图的字段名，v是此字段的字段值
			if propName, exists := fieldPropMap[k]; exists {
				// 字段属于请求的properties才set
				// 存在映射，则组装到对象属性中
				object[propName] = v
			}
		}

		// 为对象添加 _instance_id, _instance_identity, _display 字段
		instanceID, instanceIdentity := logics.GetObjectID(object, &objectType)
		displayValue := object[objectType.DisplayKey]

		if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_INSTANCE_ID, query.ExcludeSystemProperties) {
			object[interfaces.SYSTEM_PROPERTY_INSTANCE_ID] = instanceID
		}
		if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY, query.ExcludeSystemProperties) {
			object[interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY] = instanceIdentity
		}
		if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_DISPLAY, query.ExcludeSystemProperties) {
			object[interfaces.SYSTEM_PROPERTY_DISPLAY] = displayValue
		}

		if len(object) > 0 {
			objects = append(objects, object)
		} else {
			logger.Warnf("将视图行数据转成对象时，对象类属性映射的字段没有一个属性能正确映射到视图上，配置的字段属性映射关系为: %v", fieldPropMap)
		}
	}

	resps.TotalCount = viewData.TotalCount
	resps.SearchAfter = viewData.SearchAfter
	resps.Datas = objects

	return nil
}

// 从对象类索引中获取对象数据
func (ots *objectTypeService) getObjectsFromObjectIndex(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType,
	objectType interfaces.ObjectType, resps *interfaces.Objects, indexPropMap map[string]string) error {

	objects := []map[string]any{}

	// 构造 DSL 过滤条件
	conditionDslStr := "{}"
	if query.ActualCondition != nil {
		condtion, err := cond.NewCondition(ctx, query.ActualCondition, 1, logics.TransferPropsToPropMap(objectType.DataProperties))
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails(fmt.Sprintf("failed to new condition, %s", err.Error()))
		}

		// 转换到dsl
		conditionDslStr, err = condtion.Convert(ctx, func(ctx context.Context, property *cond.DataProperty, word string) ([]cond.VectorResp, error) {
			return ots.handlerVector(ctx, property, word)
		})
		if err != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails(fmt.Sprintf("failed to convert condition to dsl, %s", err.Error()))
		}

	}

	dsl, err := logics.BuildDslQuery(ctx, conditionDslStr, query)
	if err != nil {
		return err
	}
	// 请求opensearch
	osHits, err := ots.osa.SearchData(ctx, objectType.Status.Index, dsl)
	if err != nil {
		logger.Errorf("SearchData error: %s", err.Error())
		return rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_InternalError_SearchDataFromOpensearchFailed).
			WithErrorDetails(fmt.Sprintf("search data from opensearch error: %s", err.Error()))
	}

	// 根据NeedTotal参数决定是否查询total
	if query.NeedTotal {
		total, err := ots.GetTotal(ctx, objectType.Status.Index, dsl)
		if err != nil {
			return err
		}
		resps.TotalCount = total
	}

	// 把每行数据拼接到结果中
	for _, hit := range osHits {
		// 一行是一个对象
		object := map[string]any{}
		for k, v := range hit.Source {
			// k 是视图的字段名，v是此字段的字段值
			if propName, exists := indexPropMap[k]; exists {
				// 字段属于请求的properties才set
				// 存在映射，则组装到对象属性中
				object[propName] = v
			}
		}
		// 添加_score字段
		object[interfaces.SORT_FIELD_SCORE] = hit.Score

		// 为对象添加 _instance_id, _instance_identity, _display 字段
		instanceID, instanceIdentity := logics.GetObjectID(object, &objectType)
		displayValue := object[objectType.DisplayKey]

		if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_INSTANCE_ID, query.ExcludeSystemProperties) {
			object[interfaces.SYSTEM_PROPERTY_INSTANCE_ID] = instanceID
		}
		if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY, query.ExcludeSystemProperties) {
			object[interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY] = instanceIdentity
		}
		if !logics.ShouldExcludeSystemProperty(interfaces.SYSTEM_PROPERTY_DISPLAY, query.ExcludeSystemProperties) {
			object[interfaces.SYSTEM_PROPERTY_DISPLAY] = displayValue
		}

		if len(object) > 0 {
			objects = append(objects, object)
		} else {
			logger.Warnf("将视图行数据转成对象时，对象类属性映射的字段没有一个属性能正确映射到视图上，配置的字段属性映射关系为: %v", indexPropMap)
		}
	}

	var searchAfter []any
	if len(osHits) > 0 {
		searchAfter = osHits[len(osHits)-1].Sort
	} else {
		searchAfter = nil
	}
	resps.SearchAfter = searchAfter

	resps.Datas = objects

	return nil
}

func (ots *objectTypeService) GetTotal(ctx context.Context, index string, dsl map[string]any) (total int64, err error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: search object type total ")
	defer span.End()

	// delete(dsl, "pit")
	delete(dsl, "from")
	delete(dsl, "size")
	delete(dsl, "sort")
	delete(dsl, "track_scores")
	totalBytes, err := ots.osa.Count(ctx, index, dsl)
	if err != nil {
		span.SetStatus(codes.Error, "Search total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError).
			WithErrorDetails(err.Error())
	}

	totalNode, err := sonic.Get(totalBytes, "count")
	if err != nil {
		span.SetStatus(codes.Error, "Get total documents count failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError).
			WithErrorDetails(err.Error())
	}

	total, err = totalNode.Int64()
	if err != nil {
		span.SetStatus(codes.Error, "Convert total documents count to type int64 failed")
		return total, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_InternalError).
			WithErrorDetails(err.Error())
	}

	span.SetStatus(codes.Ok, "")
	return total, nil
}

// 对查询语句向量化
func (ots *objectTypeService) handlerVector(ctx context.Context, property *cond.DataProperty, word string) ([]cond.VectorResp, error) {

	// 不需要判断打开与否，不用knn操作符即可
	// 系统配置打开向量模型才可使用knn查询, 没打开,请求了knn,则报错
	// if !ots.appSetting.ServerSetting.DefaultSmallModelEnabled {
	// 	err := errors.New("defaultSmallModelEnabled is false, does not support knn condition")
	// 	return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
	// 		oerrors.OntologyQuery_ObjectType_InternalError_GetSmallModelByIDFailed).
	// 		WithErrorDetails(err.Error())
	// }

	// 先根据向量索引配置的小模型id获取一次模型的配置,获取不到就报错
	model, err := ots.mfa.GetModelByID(ctx, property.IndexConfig.VectorConfig.ModelID)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_GetSmallModelByIDFailed).
			WithErrorDetails(err.Error())
	}
	if model == nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound,
			oerrors.OntologyQuery_ObjectType_SmallModelNotFound).
			WithErrorDetails(fmt.Sprintf("model %s not found", property.IndexConfig.VectorConfig.ModelID))
	}
	if model.EmbeddingDim == 0 || model.BatchSize == 0 || model.MaxTokens == 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyQuery_ObjectType_InvalidParameter_SmallModel).
			WithErrorDetails(fmt.Sprintf("model %s has invalid embedding dim, batch size or max tokens", model.ModelID))
	}

	return ots.mfa.GetVector(ctx, model, []string{word})
}

func (ots *objectTypeService) GetObjectPropertyValue(ctx context.Context,
	query *interfaces.ObjectPropertyValueQuery) (interfaces.Objects, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "查询对象属性值")
	defer span.End()

	var resps interfaces.Objects

	// 1. 根据唯一标识构建过滤条件
	ukCond := logics.BuildInstanceIdentitiesCondition(query.InstanceIdentities)
	// 2. 根据唯一标识组成的条件检索对象类的对象实例
	objectQuery := &interfaces.ObjectQueryBaseOnObjectType{
		ActualCondition: ukCond,
		PageQuery: interfaces.PageQuery{
			Limit:     interfaces.MAX_LIMIT, // 不限制条数，要符合条件的所有,视图最大支持1w，所以就设置1w
			NeedTotal: true,
		},
		KNID:         query.KNID,
		Branch:       query.Branch,
		ObjectTypeID: query.ObjectTypeID,
		CommonQueryParameters: interfaces.CommonQueryParameters{
			IncludeTypeInfo:         true, // 需要把对象类信息返回
			IncludeLogicParams:      true, // 需要把逻辑属性的计算参数返回
			ExcludeSystemProperties: query.ExcludeSystemProperties,
		},
		ObjectQueryInfo: &interfaces.ObjectQueryInfo{
			InstanceIdentity: query.InstanceIdentities,
			Properties:       query.Properties,
		},
	}
	objects, err := ots.GetObjectsByObjectTypeID(ctx, objectQuery)
	if err != nil {
		return resps, err
	}

	// 把对象类的属性转成map
	dataProperties := map[string]cond.DataProperty{}
	for _, prop := range objects.ObjectType.DataProperties {
		dataProperties[prop.Name] = prop
	}
	logicProperties := map[string]*interfaces.LogicProperty{}
	for _, prop := range objects.ObjectType.LogicProperties {
		logicProperties[prop.Name] = prop
	}

	// 目标属性
	propertyNames := map[string]bool{}
	for _, propName := range query.Properties {
		propertyNames[propName] = true
	}
	// 添加主键
	for _, key := range objects.ObjectType.PrimaryKeys {
		propertyNames[key] = true
	}

	datas := make([]map[string]any, len(objects.Datas))
	// 第一步：同步处理所有对象的数据属性
	for i, object := range objects.Datas {
		newObject := make(map[string]any)
		for prop, value := range object {
			if !propertyNames[prop] {
				continue
			}
			// 数据属性直接赋值
			if _, exist := dataProperties[prop]; exist {
				newObject[prop] = value
			}
		}

		// 已经在对象数据查询是指定了排除字段，返回的已经是按排除字段处理后的数据，所以字段存在就添加。
		if _, exist := object[interfaces.SYSTEM_PROPERTY_INSTANCE_ID]; exist {
			newObject[interfaces.SYSTEM_PROPERTY_INSTANCE_ID] = object[interfaces.SYSTEM_PROPERTY_INSTANCE_ID]
		}
		if _, exist := object[interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY]; exist {
			newObject[interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY] = object[interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY]
		}
		if _, exist := object[interfaces.SYSTEM_PROPERTY_DISPLAY]; exist {
			newObject[interfaces.SYSTEM_PROPERTY_DISPLAY] = object[interfaces.SYSTEM_PROPERTY_DISPLAY]
		}

		datas[i] = newObject
	}
	// 第二步：并发处理所有对象的所有逻辑属性
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(objects.Datas)*len(logicProperties)) // 足够大的缓冲

	for i, object := range objects.Datas {
		for prop, value := range object {
			if !propertyNames[prop] {
				continue
			}

			// 只处理逻辑属性
			if logicProp, exist := logicProperties[prop]; exist {
				wg.Add(1)
				go func(objIndex int, propName string, propValue any, logicProp *interfaces.LogicProperty) {
					defer wg.Done()

					logger.Debugf("处理对象[%d]的逻辑属性: %s", i, propName)
					resultValue, err := ots.processLogicProperty(ctx, propName, propValue, logicProp, query.DynamicParams)
					if err != nil {
						errChan <- fmt.Errorf("对象[%d]的逻辑属性[%s]处理失败: %w", objIndex, propName, err)
						return
					}

					// 安全地写入结果到对应的对象
					mu.Lock()
					datas[objIndex][propName] = resultValue
					mu.Unlock()

				}(i, prop, value, logicProp)
			}
		}
	}

	// 等待所有逻辑属性处理完成
	wg.Wait()
	close(errChan)

	// 检查错误
	if len(errChan) > 0 {
		var errors []string
		for err := range errChan {
			errors = append(errors, err.Error())
		}
		return resps, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_ProcessLogicPropertiesFailed).
			WithErrorDetails(strings.Join(errors, "; "))
	}

	resps.Datas = datas
	return resps, nil

}

// processLogicProperty 处理单个逻辑属性（封装了原有的处理逻辑）
func (ots *objectTypeService) processLogicProperty(ctx context.Context,
	propName string,
	propValue any,
	logicProp *interfaces.LogicProperty,
	dynamicParams map[string]map[string]any) (any, error) {

	switch logicProp.Type {
	case interfaces.PROPERTY_TYPE_METRIC:
		return ots.handleMetricProperty(ctx, propName, propValue, logicProp, dynamicParams)
	case interfaces.PROPERTY_TYPE_OPERATOR:
		return ots.handleOperatorProperty(ctx, propName, propValue, logicProp, dynamicParams)
	default:
		logger.Warnf("不支持的逻辑属性类型: %s", logicProp.Type)
		return nil, nil
	}
}

// handleMetricProperty 处理指标类型逻辑属性
func (ots *objectTypeService) handleMetricProperty(ctx context.Context,
	propName string,
	propValue any,
	logicProp *interfaces.LogicProperty,
	dynamicParams map[string]map[string]any) (interfaces.MetricData, error) {

	// 用请求的动态参数填充。
	var (
		start     int64
		end       int64
		isInstant bool
		step      string
		err       error
	)

	// 如果只给 start、没有给 end，则end 为当前。
	// 如果只给了end，没有给start，则start为 end - 30min
	// metricValue 是从对象类中获取的对象的指标计算参数
	metricValue := propValue.(interfaces.MetricProperty)
	start = time.Now().Add(-30 * time.Minute).UnixMilli()
	end = time.Now().UnixMilli()
	isInstant = true

	// 1. 指标需要动态参数，校验必要的动态参数是否已给
	// if _, dynamicParamExist := dynamicParams[propName]; !dynamicParamExist {
	// 	// 动态参数没有，即 start end都没有，给默认值
	// 	dynamicParams[propName] = map[string]any{
	// 		"start":   start,
	// 		"end":     end,
	// 		"instant": isInstant,
	// 	}
	// }

	// 2. 解析 dynamic_params 为 MetricPropertyDynamicParams 结构
	var metricParams interfaces.MetricPropertyDynamicParams
	paramBytes, err := sonic.Marshal(dynamicParams[propName])
	if err != nil {
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyQuery_ObjectType_InvalidParameter_DynamicParams).
			WithErrorDetails(fmt.Sprintf("属性[%s]的动态参数解析失败: %v", propName, err))
	}
	if err = sonic.Unmarshal(paramBytes, &metricParams); err != nil {
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyQuery_ObjectType_InvalidParameter_DynamicParams).
			WithErrorDetails(fmt.Sprintf("属性[%s]的动态参数解析失败: %v", propName, err))
	}

	// 3. 提取时间参数（start, end, instant, step）
	if metricParams.Start != nil {
		start = *metricParams.Start
	}
	if metricParams.End != nil {
		end = *metricParams.End
		// 如果只给了end，没有给start，则start为 end - 30min
		if metricParams.Start == nil {
			start = end - 30*time.Minute.Milliseconds()
		}
	}
	if metricParams.Instant != nil {
		isInstant = *metricParams.Instant
	}
	if metricParams.Step != nil {
		step = *metricParams.Step
	}

	// 4. 读取配置的动态参数（从 metricValue.DynamicParams 中获取参数名）
	for paramK := range metricValue.DynamicParams {
		switch paramK {
		case "start", "end", "instant", "step":
			// 这些参数已经在上面处理了，跳过
			continue
		default:
			// 校验必须的动态参数是否已给
			paramValue, paramExist := dynamicParams[propName][paramK]
			if !paramExist {
				return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusBadRequest,
					oerrors.OntologyQuery_ObjectType_InvalidParameter_DynamicParams).
					WithErrorDetails(fmt.Sprintf("指标属性[%s]所需的动态参数[%s]为空", propName, paramK))
			}
			// 构建filter，需要从属性配置里拿到当前参数的配置逻辑
			operation := "=="
			for _, configParam := range logicProp.Parameters {
				if configParam.Name == paramK {
					if configParam.Operation != "" {
						operation = configParam.Operation
					}
				}
			}
			// 构建filters
			metricValue.Parameters.Filters = append(metricValue.Parameters.Filters,
				interfaces.Filter{
					Name:      paramK,
					Operation: operation,
					Value:     paramValue,
				})
		}
	}

	// 5. 组装指标数据查询的query，发起查询
	metricQuery := interfaces.MetricQuery{
		Start:              &start,
		End:                &end,
		StepStr:            &step,
		IsInstantQuery:     isInstant,
		Filters:            metricValue.Parameters.Filters,
		AnalysisDimensions: metricParams.AnalysisDimensions,
		OrderByFields:      metricParams.OrderByFields,
		HavingCondition:    metricParams.HavingCondition,
		Metrics:            metricParams.Metrics,
	}

	metricData, err := ots.uAccess.GetMetricDataByID(ctx, logicProp.DataSource.ID, metricQuery)
	if err != nil {
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_GetMetricDataByIDFailed).
			WithErrorDetails(fmt.Sprintf("属性[%s]的指标数据失败,error: %v", propName, err))
	}

	// 6. 结果赋值给属性。如果过滤出多个序列，不报错，多个序列都返回。
	return metricData, nil
}

// handleOperatorProperty 处理算子类型逻辑属性
func (ots *objectTypeService) handleOperatorProperty(ctx context.Context,
	propName string,
	propValue any,
	logicProp *interfaces.LogicProperty,
	dynamicParams map[string]map[string]any) (any, error) {

	operatorValue := propValue.(interfaces.OperatorProperty)

	// 1. 指标需要动态参数，校验必要的动态参数是否已给
	if _, dynamicParamExist := dynamicParams[propName]; !dynamicParamExist && len(operatorValue.DynamicParams) > 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyQuery_ObjectType_InvalidParameter_DynamicParams).
			WithErrorDetails(fmt.Sprintf("当前请求的逻辑属性[%s]所需的动态参数为空，需要参数%v，请在请求中填充动态参数", propName, operatorValue.DynamicParams))
	}

	// 1. 根据属性里配置的算子id先获取算子详情
	operatorInfo, err := ots.aoAccess.GetAgentOperatorByID(ctx, logicProp.DataSource.ID)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_GetAgentOperatorByIDFailed).
			WithErrorDetails(fmt.Sprintf("当前请求的逻辑属性[%s]是个算子，获取算子[%s]详情失败，error: %v",
				propName, logicProp.DataSource.ID, err))
	}

	// 2. 从算子详情中读取execution_mode，算子为同步的才执行，若为异步，报错提示不支持
	if operatorInfo.OperatorInfo.ExecutionMode != interfaces.OPERATOR_EXECUTION_MODE_SYNC {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_GetAgentOperatorByIDFailed).
			WithErrorDetails(fmt.Sprintf("当前请求的逻辑属性[%s]是个算子，算子[%s]的执行模式只支持同步，不支持异步", propName, logicProp.DataSource.ID))
	}

	// 3. 代理执行算子
	request := generateExecRequest(logicProp.Parameters, operatorValue.Parameters, dynamicParams[propName])
	operatorResult, err := ots.aoAccess.ExecuteOperator(ctx, logicProp.DataSource.ID, request)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_ExecuteOperatorFailed).
			WithErrorDetails(fmt.Sprintf("当前请求的逻辑属性[%s]是个算子，算子[%s]执行失败，error:%v",
				propName, logicProp.DataSource.ID, err))
	}

	// 4. 返回算子执行结果
	return operatorResult, nil
}

func generateExecRequest(configParams []interfaces.Parameter, parameters map[string]any,
	dynamicParams map[string]any) interfaces.OperatorExecutionRequest {

	operateExecRequest := interfaces.OperatorExecutionRequest{
		Header: map[string]any{},
		Query:  map[string]any{},
		Body:   map[string]any{},
		Path:   map[string]any{},
	}

	// 首先处理所有参数，构建基础结构
	for _, param := range configParams {
		var value any

		if param.ValueFrom == interfaces.VALUE_FROM_INPUT {
			// 动态输入参数从dynamicParameterMap获取
			value = getNestedValue(dynamicParams, param.Name)
		} else {
			// 固定值参数从parameterMap获取
			value = getNestedValue(parameters, param.Name)
		}

		// 根据source分配到不同的组
		switch strings.ToLower(param.Source) {
		case interfaces.PARAMETER_HEADER:
			setNestedValue(operateExecRequest.Header, param.Name, value)
		case interfaces.PARAMETER_QUERY:
			setNestedValue(operateExecRequest.Query, param.Name, value)
		case interfaces.PARAMETER_BODY:
			setNestedValue(operateExecRequest.Body, param.Name, value)
		case interfaces.PARAMETER_PATH:
			setNestedValue(operateExecRequest.Path, param.Name, value)
		}
	}
	return operateExecRequest
}

// getNestedValue 从map中获取嵌套字段的值
func getNestedValue(data map[string]any, key string) any {
	if data == nil {
		return nil
	}

	// 如果key包含点号，表示嵌套字段
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		current := data

		for i, part := range parts {
			if i == len(parts)-1 {
				// 最后一个部分，返回值
				return current[part]
			}

			// 中间部分，继续深入
			if next, ok := current[part].(map[string]any); ok {
				current = next
			} else {
				return nil
			}
		}
	}

	return data[key]
}

// setNestedValue 设置嵌套字段的值到map中
func setNestedValue(target map[string]any, key string, value any) {
	if value == nil {
		return
	}

	// 如果key包含点号，表示需要设置嵌套字段
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		current := target

		for i, part := range parts {
			if i == len(parts)-1 {
				// 最后一个部分，设置值
				current[part] = value
				return
			}

			// 中间部分，确保map存在
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]any)
			}

			// 类型断言，继续深入
			if next, ok := current[part].(map[string]any); ok {
				current = next
			} else {
				// 如果类型不匹配，覆盖为新的map
				current[part] = make(map[string]any)
				current = current[part].(map[string]any)
			}
		}
	} else {
		// 简单字段直接设置
		target[key] = value
	}
}
