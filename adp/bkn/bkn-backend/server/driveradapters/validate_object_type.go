// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dlclark/regexp2"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func ValidateObjectTypes(ctx context.Context, knID string, objectTypes []*interfaces.ObjectType, strictMode bool) error {
	tmpNameMap := make(map[string]any)
	idMap := make(map[string]any)
	for i := 0; i < len(objectTypes); i++ {
		// 校验导入模型时模块是否是对象类
		if objectTypes[i].ModuleType != "" && objectTypes[i].ModuleType != interfaces.MODULE_TYPE_OBJECT_TYPE {
			return rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_InvalidParameter_ModuleType).
				WithErrorDetails("Object type name is not 'object_type'")
		}

		// 0.校验请求体中多个模型 ID 是否重复
		otID := objectTypes[i].OTID
		if _, ok := idMap[otID]; !ok || otID == "" {
			idMap[otID] = nil
		} else {
			errDetails := fmt.Sprintf("ObjectType ID '%s' already exists in the request body", otID)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_Duplicated_IDInFile).
				WithDescription(map[string]any{"ObjectTypeID": otID}).
				WithErrorDetails(errDetails)
		}

		// 1. 校验 对象类必要创建参数的合法性, 非空、长度、是枚举值
		err := ValidateObjectType(ctx, objectTypes[i], strictMode)
		if err != nil {
			return err
		}

		// 2. 校验 请求体中对象类名称重复性
		if _, ok := tmpNameMap[objectTypes[i].OTName]; !ok {
			tmpNameMap[objectTypes[i].OTName] = nil
		} else {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_Duplicated_Name)
		}

		objectTypes[i].KNID = knID
	}
	return nil
}

// 对象类必要创建参数的非空校验。
func ValidateObjectType(ctx context.Context, objectType *interfaces.ObjectType, strictMode bool) error {
	// 校验id的合法性
	err := validateID(ctx, objectType.OTID)
	if err != nil {
		return err
	}

	// 校验名称合法性
	// 去掉名称的前后空格
	objectType.OTName = strings.TrimSpace(objectType.OTName)
	err = validateObjectName(ctx, objectType.OTName, interfaces.MODULE_TYPE_OBJECT_TYPE)
	if err != nil {
		return err
	}

	// 若输入了 tags，校验 tags 的合法性
	err = ValidateTags(ctx, objectType.Tags)
	if err != nil {
		return err
	}

	// 去掉tag前后空格以及数组去重
	objectType.Tags = libCommon.TagSliceTransform(objectType.Tags)

	// 校验 data_source.type 非空时，只支持 data_view、resource
	if objectType.DataSource != nil && objectType.DataSource.Type != "" {
		if objectType.DataSource.Type != interfaces.DATA_SOURCE_TYPE_DATA_VIEW &&
			objectType.DataSource.Type != interfaces.DATA_SOURCE_TYPE_RESOURCE {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]数据来源类型[%s]不支持, 只支持 data_view、resource", objectType.OTName, objectType.DataSource.Type))
		}
	}

	if len(objectType.DataProperties) > interfaces.MAX_PROPERTY_NUM {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("对象类[%s]数据属性数[%d]超过最大限制[%d]", objectType.OTName, len(objectType.DataProperties), interfaces.MAX_PROPERTY_NUM))
	}

	if len(objectType.LogicProperties) > interfaces.MAX_PROPERTY_NUM {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性数[%d]超过最大限制[%d]", objectType.OTName, len(objectType.LogicProperties), interfaces.MAX_PROPERTY_NUM))
	}

	// 属性名只包含小写英文字母和数字和下划线(_)和连字符(-)，且不能以下划线开头，不能超过40个字符
	dataPropMap := map[string]*interfaces.DataProperty{}
	for _, prop := range objectType.DataProperties {
		err = ValidateDataProperty(ctx, prop, strictMode)
		if err != nil {
			return err
		}

		dataPropMap[prop.Name] = prop
	}

	// 校验主键非空
	if len(objectType.PrimaryKeys) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_ObjectType_NullParameter_PrimaryKeys)
	} else {
		for _, pKey := range objectType.PrimaryKeys {
			prop, ok := dataPropMap[pKey]
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]主键[%s]不存在", objectType.OTName, pKey))
			}

			// primary_keys：主键的属性的类型只能是: integer, unsigned integer, string, text
			if !interfaces.ValidPrimaryKeyTypes[prop.Type] {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]主键[%s]类型[%s]无效，只支持 integer, unsigned integer, string", objectType.OTName, pKey, prop.Type))
			}
		}
	}

	if objectType.DisplayKey == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_ObjectType_NullParameter_DisplayKey)
	} else {
		prop, ok := dataPropMap[objectType.DisplayKey]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]显示键[%s]不存在", objectType.OTName, objectType.DisplayKey))
		}

		// display_key： 标题的属性的类型支持： integer, unsigned integer, float, decimal, string, text, date, timestamp, time, datetime, boolean
		if !interfaces.ValidDisplayKeyTypes[prop.Type] {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]显示键[%s]类型[%s]无效，只支持 integer, unsigned integer, float, decimal, string, text, date, timestamp, time, datetime, boolean", objectType.OTName, objectType.DisplayKey, prop.Type))
		}
	}

	if objectType.IncrementalKey != "" {
		if field, ok := dataPropMap[objectType.IncrementalKey]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]增量键[%s]不存在", objectType.OTName, objectType.IncrementalKey))
		} else {
			switch field.Type {
			case "integer", "datetime", "timestamp":
			default:
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("不支持的对象类[%s]增量键[%s]类型[%s]", objectType.OTName, field.Name, field.Type))
			}
		}
	}

	// 当逻辑属性是指标模型时，初始化3个请求参数, instant start end
	IfSystemGen := true
	for i, prop := range objectType.LogicProperties {
		// 校验属性名的合法性,与id的规则不同，属性名还支持大写字母
		err := ValidatePropertyName(ctx, prop.Name)
		if err != nil {
			return err
		}

		// 校验displayName
		err = validateObjectName(ctx, prop.DisplayName, interfaces.MODULE_TYPE_OBJECT_TYPE)
		if err != nil {
			return err
		}

		// logic_property: type and data_source (type, id) are required; types must match (metric / operator)
		if prop.Type == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]类型不能为空", objectType.OTName, prop.Name))
		}
		if prop.Type != interfaces.LOGIC_PROPERTY_TYPE_METRIC && prop.Type != interfaces.LOGIC_PROPERTY_TYPE_OPERATOR {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]类型[%s]无效，只支持 metric, operator", objectType.OTName, prop.Name, prop.Type))
		}
		if prop.DataSource == nil || prop.DataSource.Type == "" || prop.DataSource.ID == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的数据来源(type、id)不能为空", objectType.OTName, prop.Name))
		}
		if !interfaces.ValidLogicSourceTypes[prop.DataSource.Type] {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的数据资源类型[%s]无效，只支持 metric, operator", objectType.OTName, prop.Name, prop.DataSource.Type))
		}
		if prop.Type != prop.DataSource.Type {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的数据类型[%s]与其所绑定的数据资源类型[%s]不一致",
					objectType.OTName, prop.Name, prop.Type, prop.DataSource.Type))
		}

		//  logic_property.parameters 非空时：参数名称非空
		if len(prop.Parameters) > 0 {
			for _, param := range prop.Parameters {
				if param.Name == "" {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
						WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的参数名称不能为空", objectType.OTName, prop.Name))
				}
			}
		}

		// 如果逻辑属性是指标模型
		if prop.Type == interfaces.PROPERTY_TYPE_METRIC {
			// 如果不存在指标相关的参数，那么就追加
			paramMap := map[string]interfaces.Parameter{}
			for _, param := range prop.Parameters {
				paramMap[param.Name] = param
			}
			hasInstant := false
			hasStart := false
			hasEnd := false
			hasStep := false
			for pName := range paramMap {
				switch pName {
				case "instant":
					hasInstant = true
				case "start":
					hasStart = true
				case "end":
					hasEnd = true
				case "step":
					hasStep = true
				}
			}
			params := []interfaces.Parameter{}
			if !hasInstant {
				params = append(params, interfaces.Parameter{
					Name:        "instant",
					Type:        "boolean",
					Operation:   "==",
					ValueFrom:   interfaces.VALUE_FROM_INPUT,
					IfSystemGen: &IfSystemGen,
				})
			}
			if !hasStart {
				params = append(params, interfaces.Parameter{
					Name:        "start",
					Type:        "integer",
					Operation:   "==",
					ValueFrom:   interfaces.VALUE_FROM_INPUT,
					IfSystemGen: &IfSystemGen,
				})
			}
			if !hasEnd {
				params = append(params, interfaces.Parameter{
					Name:        "end",
					Type:        "integer",
					Operation:   "==",
					ValueFrom:   interfaces.VALUE_FROM_INPUT,
					IfSystemGen: &IfSystemGen,
				})
			}
			if !hasStep {
				params = append(params, interfaces.Parameter{
					Name:        "step",
					Type:        "string",
					Operation:   "==",
					ValueFrom:   interfaces.VALUE_FROM_INPUT,
					IfSystemGen: &IfSystemGen,
				})
			}
			objectType.LogicProperties[i].Parameters = append(objectType.LogicProperties[i].Parameters, params...)
		}
	}

	return nil
}

func ValidatePropertyName(ctx context.Context, name string) error {
	if name == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_NullParameter_PropertyName)
	}
	//  id，只包含大小写英文字母和数字和下划线(_)和连字符(-)，且不能以下划线开头，不能超过40个字符
	re := regexp2.MustCompile(interfaces.RegexPattern_Property_Name, regexp2.RE2)
	match, err := re.MatchString(name)
	if err != nil || !match {
		errDetails := `The property name can contain only letters, digits and underscores(_),
			it cannot start with underscores and cannot exceed 40 characters`
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter_PropertyName).
			WithErrorDetails(errDetails)
	}
	return nil
}

func ValidateDataProperties(ctx context.Context, propertyNames []string, dataProperties []*interfaces.DataProperty, strictMode bool) error {
	if len(propertyNames) != len(dataProperties) {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails("PropertyNames and DataProperties length not equal")
		return httpErr
	}

	propertyNameMap := map[string]string{}
	for _, propertyName := range propertyNames {
		propertyNameMap[propertyName] = propertyName
	}
	for _, prop := range dataProperties {
		if _, ok := propertyNameMap[prop.Name]; !ok {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("DataProperty %s not in URL", prop.Name))
			return httpErr
		}

		err := ValidateDataProperty(ctx, prop, strictMode)
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateDataProperty(ctx context.Context, dataProperty *interfaces.DataProperty, strictMode bool) error {
	// 校验属性名的合法性,与id的规则不同，属性名还支持大写字母
	err := ValidatePropertyName(ctx, dataProperty.Name)
	if err != nil {
		return err
	}

	err = validateObjectName(ctx, dataProperty.DisplayName,
		interfaces.MODULE_TYPE_OBJECT_TYPE)
	if err != nil {
		return err
	}

	// data_property.type： 非空时，需是有效的类型：integer, unsigned integer, float, decimal, string, text, date, timestamp, time, datetime, boolean, binary, json, vector, point, shape, ip。
	if dataProperty.Type != "" {
		if !interfaces.ValidDataPropertyTypes[dataProperty.Type] {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("数据属性[%s]类型[%s]无效，只支持 integer, unsigned integer, float, decimal, string, text, date, timestamp, time, datetime, boolean, binary, json, vector, point, shape, ip",
					dataProperty.Name, dataProperty.Type))
		}
	}

	// data_property.mapped_field：非空时，name 非空
	if dataProperty.MappedField != nil && dataProperty.MappedField.Name == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("数据属性[%s]的映射字段名称不能为空", dataProperty.Name))
	}

	if dataProperty.IndexConfig != nil {
		err = ValidateIndexConfig(ctx, *dataProperty.IndexConfig, strictMode)
		if err != nil {
			return err
		}
	}

	return nil
}

func ValidateIndexConfig(ctx context.Context, indexConfig interfaces.IndexConfig, strictMode bool) error {
	err := ValidateKeywordConfig(ctx, indexConfig.KeywordConfig)
	if err != nil {
		return err
	}
	err = ValidateFulltextConfig(ctx, indexConfig.FulltextConfig)
	if err != nil {
		return err
	}
	err = ValidateVectorConfig(ctx, indexConfig.VectorConfig, strictMode)
	if err != nil {
		return err
	}

	return nil
}

func ValidateKeywordConfig(ctx context.Context, keywordConfig interfaces.KeywordConfig) error {
	if !keywordConfig.Enabled {
		return nil
	}
	if keywordConfig.IgnoreAboveLen <= 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails("KeywordConfig IgnoreAboveLen must be greater than 0")
		return httpErr
	}
	return nil
}

func ValidateFulltextConfig(ctx context.Context, fulltextConfig interfaces.FulltextConfig) error {
	if !fulltextConfig.Enabled {
		return nil
	}
	switch fulltextConfig.Analyzer {
	case "standard", "english", "ik_max_word", "hanlp_standard", "hanlp_index":
	default:
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails("FulltextConfig Analyzer must be standard, english, ik_max_word, hanlp_standard or hanlp_index")
		return httpErr
	}
	return nil
}

func ValidateVectorConfig(ctx context.Context, vectorConfig interfaces.VectorConfig, strictMode bool) error {
	if !vectorConfig.Enabled {
		return nil
	}
	if strictMode && vectorConfig.ModelID == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails("VectorConfig ModelID must be set")
		return httpErr
	}
	return nil
}
