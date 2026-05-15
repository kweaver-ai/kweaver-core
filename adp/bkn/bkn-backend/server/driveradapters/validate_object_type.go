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
	"unicode/utf8"

	"github.com/dlclark/regexp2"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	cond "bkn-backend/common/condition"
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

// ValidateObjectType 对象类必要创建参数的合法性校验。
// 校验顺序：基础信息 → 数据来源 → 数据属性 → 键 → 逻辑属性 → 过滤条件
func ValidateObjectType(ctx context.Context, objectType *interfaces.ObjectType, strictMode bool) error {
	// 1. 校验基础信息：id、name、tags
	if err := validateObjectTypeBasicInfo(ctx, objectType); err != nil {
		return err
	}

	// 2. 校验数据来源
	if err := validateObjectTypeDataSource(ctx, objectType); err != nil {
		return err
	}

	// 3. 校验数据属性
	if err := validateObjectTypeDataProperties(ctx, objectType, strictMode); err != nil {
		return err
	}

	// 4. 构建数据属性索引，校验键（依赖数据属性）
	dataPropMap := buildDataPropMap(objectType.DataProperties)
	if err := validateObjectTypeKeys(ctx, objectType, dataPropMap, strictMode); err != nil {
		return err
	}

	// 5. 校验逻辑属性
	if err := validateObjectTypeLogicProperties(ctx, objectType, strictMode); err != nil {
		return err
	}

	// 6. 校验过滤条件
	if err := validateObjectTypeCondition(ctx, objectType, dataPropMap); err != nil {
		return err
	}

	return nil
}

// validateObjectTypeBasicInfo 校验对象类基础信息：id、name、tags。
func validateObjectTypeBasicInfo(ctx context.Context, objectType *interfaces.ObjectType) error {
	// 校验 id 合法性
	if err := validateID(ctx, objectType.OTID); err != nil {
		return err
	}

	// 去掉名称前后空格后校验合法性
	objectType.OTName = strings.TrimSpace(objectType.OTName)
	if err := validateObjectName(ctx, objectType.OTName, interfaces.MODULE_TYPE_OBJECT_TYPE); err != nil {
		return err
	}

	// 校验 tags 合法性
	if err := ValidateTags(ctx, objectType.Tags); err != nil {
		return err
	}
	// 去掉 tag 前后空格并去重
	objectType.Tags = libCommon.TagSliceTransform(objectType.Tags)

	return nil
}

// validateObjectTypeDataSource 校验对象类数据来源：type 只支持 data_view、resource。
func validateObjectTypeDataSource(ctx context.Context, objectType *interfaces.ObjectType) error {
	if objectType.DataSource == nil || objectType.DataSource.Type == "" {
		return nil
	}
	if objectType.DataSource.Type != interfaces.DATA_SOURCE_TYPE_DATA_VIEW &&
		objectType.DataSource.Type != interfaces.DATA_SOURCE_TYPE_RESOURCE {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("对象类[%s]数据来源类型[%s]不支持, 只支持 data_view、resource", objectType.OTName, objectType.DataSource.Type))
	}
	return nil
}

// buildDataPropMap 将数据属性列表转为以属性名为键的 map，纯构建无副作用。
func buildDataPropMap(dataProperties []*interfaces.DataProperty) map[string]*interfaces.DataProperty {
	m := make(map[string]*interfaces.DataProperty, len(dataProperties))
	for _, prop := range dataProperties {
		m[prop.Name] = prop
	}
	return m
}

// validateObjectTypeDataProperties 校验数据属性数量上限及每个属性的合法性。
func validateObjectTypeDataProperties(ctx context.Context, objectType *interfaces.ObjectType, strictMode bool) error {
	if len(objectType.DataProperties) > interfaces.MAX_PROPERTY_NUM {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("对象类[%s]数据属性数[%d]超过最大限制[%d]", objectType.OTName, len(objectType.DataProperties), interfaces.MAX_PROPERTY_NUM))
	}

	for _, prop := range objectType.DataProperties {
		if err := ValidateDataProperty(ctx, prop, strictMode); err != nil {
			return err
		}
	}
	return nil
}

// validateObjectTypeKeys 校验主键、显示键、增量键的合法性（依赖 dataPropMap）。
// 严格模式下：primary_keys 和 display_key 不能为空。
// 非严格模式下：可不配置，但若配置了必须是已存在的合法字段。
func validateObjectTypeKeys(ctx context.Context, objectType *interfaces.ObjectType, dataPropMap map[string]*interfaces.DataProperty, strictMode bool) error {
	// 校验主键：严格模式下不能为空；若配置了则校验存在性和类型
	if len(objectType.PrimaryKeys) == 0 {
		if strictMode {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_NullParameter_PrimaryKeys)
		}
	} else {
		for _, pKey := range objectType.PrimaryKeys {
			prop, ok := dataPropMap[pKey]
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]主键[%s]不存在", objectType.OTName, pKey))
			}
			// primary_keys：主键属性类型只能是 integer, unsigned integer, string, text
			if !interfaces.ValidPrimaryKeyTypes[prop.Type] {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]主键[%s]类型[%s]无效，只支持 integer, unsigned integer, string, text", objectType.OTName, pKey, prop.Type))
			}
		}
	}

	// 校验显示键：严格模式下不能为空；若配置了则校验存在性和类型
	if objectType.DisplayKey == "" {
		if strictMode {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_NullParameter_DisplayKey)
		}
	} else {
		prop, ok := dataPropMap[objectType.DisplayKey]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]显示键[%s]不存在", objectType.OTName, objectType.DisplayKey))
		}
		// display_key：类型支持 integer, unsigned integer, float, decimal, string, text, date, timestamp, time, datetime, boolean
		if !interfaces.ValidDisplayKeyTypes[prop.Type] {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]显示键[%s]类型[%s]无效，只支持 integer, unsigned integer, float, decimal, string, text, date, timestamp, time, datetime, boolean", objectType.OTName, objectType.DisplayKey, prop.Type))
		}
	}

	// 校验增量键：始终可选；若配置了则校验存在性和类型
	if objectType.IncrementalKey != "" {
		field, ok := dataPropMap[objectType.IncrementalKey]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]增量键[%s]不存在", objectType.OTName, objectType.IncrementalKey))
		}
		switch field.Type {
		case "integer", "datetime", "timestamp":
		default:
			return rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("不支持的对象类[%s]增量键[%s]类型[%s]", objectType.OTName, field.Name, field.Type))
		}
	}

	return nil
}

// validateObjectTypeLogicProperties 校验逻辑属性数量上限及每个逻辑属性的合法性，
// 并为指标类型属性自动补全系统参数（instant、start、end、step）。
func validateObjectTypeLogicProperties(ctx context.Context, objectType *interfaces.ObjectType, strictMode bool) error {
	if len(objectType.LogicProperties) > interfaces.MAX_PROPERTY_NUM {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性数[%d]超过最大限制[%d]", objectType.OTName, len(objectType.LogicProperties), interfaces.MAX_PROPERTY_NUM))
	}

	ifSystemGen := true
	dataPropMap := buildDataPropMap(objectType.DataProperties)
	for i, prop := range objectType.LogicProperties {
		// 校验属性名合法性（支持大写字母，规则与 id 不同）
		if err := ValidatePropertyName(ctx, prop.Name); err != nil {
			return err
		}

		// 校验 displayName
		if prop.DisplayName == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的显示名称不能为空", objectType.OTName, prop.Name))
		}
		if utf8.RuneCountInString(prop.DisplayName) > interfaces.OBJECT_NAME_MAX_LENGTH {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的显示名称长度不能超过%d个字符", objectType.OTName, prop.Name, interfaces.OBJECT_NAME_MAX_LENGTH))
		}

		// type 非空且只支持 metric / operator / rid
		if prop.Type == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]类型不能为空", objectType.OTName, prop.Name))
		}
		if prop.Type != interfaces.LOGIC_PROPERTY_TYPE_METRIC &&
			prop.Type != interfaces.LOGIC_PROPERTY_TYPE_OPERATOR &&
			prop.Type != interfaces.LOGIC_PROPERTY_TYPE_RID {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]类型[%s]无效，只支持 metric, operator, rid", objectType.OTName, prop.Name, prop.Type))
		}

		// rid 类型特殊校验：需要 kind 和 field
		if prop.Type == interfaces.LOGIC_PROPERTY_TYPE_RID {
			if prop.Kind == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的 kind 不能为空", objectType.OTName, prop.Name))
			}
			if !interfaces.IsValidRidKind(prop.Kind) {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的 kind[%s]无效，只支持 %s",
						objectType.OTName, prop.Name, prop.Kind, strings.Join(interfaces.ValidRidKindList(), ", ")))
			}
			if prop.Field == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的 field 不能为空", objectType.OTName, prop.Name))
			}
			if _, ok := dataPropMap[prop.Field]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的 field[%s]不存在", objectType.OTName, prop.Name, prop.Field))
			}
			// rid 类型不需要 data_source，跳过后续 data_source 校验
			continue
		}

		// 校验 data_source：
		// - 严格模式：必须存在（type 和 id 均非空）
		// - 非严格模式：可不存在；若存在，type 和 id 必须同时有效，不允许半填
		if prop.DataSource != nil && (prop.DataSource.Type != "" || prop.DataSource.ID != "") {
			// DataSource 存在，校验 type 和 id 必须同时填写
			if prop.DataSource.Type == "" || prop.DataSource.ID == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的数据来源 type 和 id 必须同时填写", objectType.OTName, prop.Name))
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
		} else if strictMode {
			// DataSource 不存在（nil 或 type/id 均为空），严格模式下报错
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的数据来源(type、id)不能为空", objectType.OTName, prop.Name))
		}

		// 校验参数名称非空
		for _, param := range prop.Parameters {
			if param.Name == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("对象类[%s]逻辑属性[%s]的参数名称不能为空", objectType.OTName, prop.Name))
			}
		}

		// 指标类型：自动补全系统参数 instant、start、end、step
		if prop.Type == interfaces.LOGIC_PROPERTY_TYPE_METRIC {
			paramMap := make(map[string]struct{}, len(prop.Parameters))
			for _, param := range prop.Parameters {
				paramMap[param.Name] = struct{}{}
			}
			var extra []interfaces.Parameter
			for _, name := range []string{"instant", "start", "end", "step"} {
				if _, exists := paramMap[name]; exists {
					continue
				}
				p := interfaces.Parameter{
					Operation:   "==",
					ValueFrom:   interfaces.VALUE_FROM_INPUT,
					IfSystemGen: &ifSystemGen,
					Name:        name,
				}
				switch name {
				case "instant":
					p.Type = "boolean"
				case "start", "end":
					p.Type = "integer"
				case "step":
					p.Type = "string"
				}
				extra = append(extra, p)
			}
			objectType.LogicProperties[i].Parameters = append(objectType.LogicProperties[i].Parameters, extra...)
		}
	}

	return nil
}

// validateObjectTypeCondition 校验对象类的过滤条件 Condition 合法性。
func validateObjectTypeCondition(ctx context.Context, objectType *interfaces.ObjectType, dataPropMap map[string]*interfaces.DataProperty) error {
	if objectType.Condition == nil {
		return nil
	}

	// 构建字段映射用于校验
	fieldsMap := make(map[string]*cond.FieldCfg, len(dataPropMap))
	for name, prop := range dataPropMap {
		fieldsMap[name] = &cond.FieldCfg{
			Name:         name,
			Type:         prop.Type,
			DisplayName:  prop.DisplayName,
			OriginalName: name,
		}
	}

	// 校验 condition 结构合法性
	if err := validateCondCfg(ctx, objectType.Condition, fieldsMap, objectType.OTName, ""); err != nil {
		return err
	}

	return nil
}

// validateCondCfg 递归校验 CondCfg 的合法性
func validateCondCfg(ctx context.Context, cfg *cond.CondCfg, fieldsMap map[string]*cond.FieldCfg, otName string, prefix string) error {
	if cfg == nil {
		return nil
	}

	// 校验 operation 不能为空
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("对象类[%s]过滤条件%s的 operation 不能为空", otName, prefix))
	}

	// 校验 operation 是否有效
	if _, ok := cond.OperationMap[cfg.Operation]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("对象类[%s]过滤条件%s的 operation[%s]无效", otName, prefix, cfg.Operation))
	}

	// and/or 操作需要校验子条件
	if cfg.Operation == cond.OperationAnd || cfg.Operation == cond.OperationOr {
		if len(cfg.SubConds) == 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]过滤条件%s的子条件(sub_conditions)不能为空", otName, prefix))
		}

		// 校验子条件数量限制
		if len(cfg.SubConds) > cond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]过滤条件%s的子条件数量[%d]超过最大限制[%d]", otName, prefix, len(cfg.SubConds), cond.MaxSubCondition))
		}

		// 递归校验每个子条件
		for i, subCond := range cfg.SubConds {
			subPrefix := fmt.Sprintf("%s.sub_conditions[%d]", prefix, i)
			if err := validateCondCfg(ctx, subCond, fieldsMap, otName, subPrefix); err != nil {
				return err
			}
		}
		return nil
	}

	// 非 and/or 操作需要校验 field
	if cfg.Field == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("对象类[%s]过滤条件%s的 field 不能为空", otName, prefix))
	}

	// 校验 field 是否存在于数据属性中（* 表示所有字段，特殊处理）
	if cfg.Field != cond.AllField {
		if _, ok := fieldsMap[cfg.Field]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]过滤条件%s的字段[%s]不存在", otName, prefix, cfg.Field))
		}
	}

	// 校验 value：非特殊操作需要 value
	if _, ok := cond.NotRequiredValueOperationMap[cfg.Operation]; !ok {
		if cfg.Value == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("对象类[%s]过滤条件%s的值(value)不能为空", otName, prefix))
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

	if dataProperty.DisplayName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("数据属性[%s]的显示名称不能为空", dataProperty.Name))
	}
	if utf8.RuneCountInString(dataProperty.DisplayName) > interfaces.OBJECT_NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("数据属性[%s]的显示名称长度不能超过%d个字符", dataProperty.Name, interfaces.OBJECT_NAME_MAX_LENGTH))
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
