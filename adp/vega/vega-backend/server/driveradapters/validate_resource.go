// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/extensions"
)

func ValidateResourceRequest(ctx context.Context, req *interfaces.ResourceRequest) error {
	if err := validateID(ctx, req.ID); err != nil {
		return err
	}

	if err := validateName(ctx, req.Name); err != nil {
		return err
	}
	if err := ValidateTags(ctx, req.Tags); err != nil {
		return err
	}
	if err := validateDescription(ctx, req.Description); err != nil {
		return err
	}

	if req.Extensions != nil {
		if err := extensions.ValidateEntityExtensionsMap(ctx, *req.Extensions); err != nil {
			return err
		}
	}

	switch req.Category {
	case interfaces.ResourceCategoryLogicView:
		return validateLogicViewRequest(ctx, req)
	default:
		if err := extensions.ValidateSchemaPropertiesExtensions(ctx, req.SchemaDefinition); err != nil {
			return err
		}
		return nil
	}
}

func validateLogicViewRequest(ctx context.Context, req *interfaces.ResourceRequest) error {
	outputFields, err := validateLogicDefinition(ctx, req.LogicDefinition)
	if err != nil {
		return err
	}

	// 校验字段
	err = validateViewFields(ctx, outputFields)
	if err != nil {
		return err
	}

	return nil

}

// 校验逻辑视图定义
func validateLogicDefinition(ctx context.Context, nodes []*interfaces.LogicDefinitionNode) (outputFields []*interfaces.ViewProperty, err error) {
	if nodes == nil {
		return nil, nil
	}

	if len(nodes) > 20 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition nodes cannot be more than 20")
	}

	for _, node := range nodes {
		// 检测 nodeType
		if _, ok := interfaces.LogicDefinitionNodeTypeMap[node.Type]; !ok {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition node type is invalid")
		}

		if node.Type == interfaces.LogicDefinitionNodeType_Output {
			outputFields = node.OutputFields
		}
	}

	return outputFields, nil
}

// 校验字段和字段特征
func validateViewFields(ctx context.Context, viewFields []*interfaces.ViewProperty) error {
	fieldsMap := make(map[string]*interfaces.ViewProperty)
	for _, field := range viewFields {
		fieldsMap[field.Name] = field
	}

	// 校验字段名称、显示名是否重复
	nameMap := make(map[string]struct{})
	displayNameMap := make(map[string]struct{})
	for _, field := range viewFields {
		if field.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_FieldName).
				WithErrorDetails("The field name is null")
		}

		// 校验字段名称长度, 长度限制255
		if utf8.RuneCountInString(field.Name) > interfaces.MaxLength_ViewPropertyName {
			errDetails := fmt.Sprintf("The length of the field name %s exceeds %d", field.Name, interfaces.MaxLength_ViewPropertyName)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_LengthExceeded_FieldName).
				WithErrorDetails(errDetails)
		}

		// 如果display_name为 "", 将display_name的值等于field的值
		if field.DisplayName == "" {
			field.DisplayName = field.Name
		}

		// 校验字段显示名长度, 长度限制255
		if utf8.RuneCountInString(field.DisplayName) > interfaces.MaxLength_ViewPropertyDisplayName {
			errDetails := fmt.Sprintf("The length of the field display name %s exceeds %d", field.DisplayName, interfaces.MaxLength_ViewPropertyDisplayName)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_LengthExceeded_FieldDisplayName).
				WithErrorDetails(errDetails)
		}

		// 校验字段备注长度，长度限制1000
		if utf8.RuneCountInString(field.Description) > interfaces.MaxLength_ViewPropertyDescription {
			errDetails := fmt.Sprintf("The length of the field comment %s exceeds %d", field.Description, interfaces.MaxLength_ViewPropertyDescription)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_LengthExceeded_FieldComment).
				WithErrorDetails(errDetails)
		}

		// 校验字段名称是否重复
		if _, ok := nameMap[field.Name]; !ok {
			nameMap[field.Name] = struct{}{}
		} else {
			errDetails := fmt.Sprintf("Logic view field '%s' name '%s' already exists", field.Name, field.Name)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_Duplicated_FieldName).
				WithDescription(map[string]any{"FieldName": field.Name}).
				WithErrorDetails(errDetails)
		}

		if _, ok := displayNameMap[field.DisplayName]; !ok {
			displayNameMap[field.DisplayName] = struct{}{}
		} else {
			errDetails := fmt.Sprintf("Logic view field '%s' display name '%s' already exists", field.Name, field.DisplayName)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_Duplicated_FieldDisplayName).
				WithDescription(map[string]any{"FieldName": field.Name, "DisplayName": field.DisplayName}).
				WithErrorDetails(errDetails)
		}

		// 校验特征
		err := validateFeatures(ctx, fieldsMap, field.Features)
		if err != nil {
			return err
		}

		if len(field.Extensions) > 0 {
			if err := extensions.ValidatePropertyExtensionsMap(ctx, field.Extensions); err != nil {
				return err
			}
		}
	}

	return nil
}

// 校验特征
func validateFeatures(ctx context.Context, fieldsMap map[string]*interfaces.ViewProperty, features []interfaces.PropertyFeature) error {
	enabledMap := make(map[string]bool)
	featureNameMap := make(map[string]struct{})
	for _, f := range features {
		if f.FeatureName == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_FieldFeatureName).
				WithErrorDetails("The field feature name is null")
		}

		// 校验特征名称长度, 长度限制255
		if utf8.RuneCountInString(f.FeatureName) > interfaces.MaxLength_ViewPropertyFeatureName {
			errDetails := fmt.Sprintf("The length of the field feature name %s exceeds %d", f.FeatureName, interfaces.MaxLength_ViewPropertyFeatureName)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_LengthExceeded_FieldFeatureName).
				WithErrorDetails(errDetails)
		}

		// 校验特征名称是否重复
		if _, ok := featureNameMap[f.FeatureName]; !ok {
			featureNameMap[f.FeatureName] = struct{}{}
		} else {
			errDetails := fmt.Sprintf("Logic view field feature '%s' name '%s' already exists", f.FeatureName, f.FeatureName)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_Duplicated_FieldFeatureName).
				WithDescription(map[string]any{"FieldFeatureName": f.FeatureName}).
				WithErrorDetails(errDetails)
		}

		// feature type
		if _, ok := interfaces.PropertyFeatureTypeMap[f.FeatureType]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("The field feature type is invalid")
		}

		// 校验特征备注，长度限制1000
		if utf8.RuneCountInString(f.Description) > interfaces.MaxLength_ViewPropertyFeatureDescription {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_LengthExceeded_FieldFeatureComment).
				WithErrorDetails(fmt.Sprintf("The length of the field feature comment %s exceeds %d", f.Description, interfaces.MaxLength_ViewPropertyFeatureDescription))
		}

		if f.RefProperty == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails("The field feature ref field is null")
		}

		// 校验非原生特征的引用字段
		if !f.IsNative {
			// 引用字段是否在字段列表里
			if _, ok := fieldsMap[f.RefProperty]; !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails(fmt.Sprintf("The field feature ref field '%s' is not in the field list", f.RefProperty))
			}

			// 引用字段的类型是否符合特征类型
			if !IsFeatureSupported(fieldsMap[f.RefProperty].Type, f.FeatureType) {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails(fmt.Sprintf("The field feature ref field '%s' type '%s' is not supported", f.RefProperty, fieldsMap[f.RefProperty].Type))
			}
		}

		// 校验是否已启用
		if f.IsDefault {
			if enabledMap[f.FeatureType] {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
					WithErrorDetails(fmt.Sprintf("Same type features can only have one default feature, current field feature name '%s' type is '%s'",
						f.FeatureName, f.FeatureType))
			}
			enabledMap[f.FeatureType] = true
		}
	}

	return nil
}

func IsFeatureSupported(fieldType string, featureType string) bool {
	switch featureType {
	case interfaces.PropertyFeatureType_Fulltext:
		return fieldType == interfaces.DataType_Text
	case interfaces.PropertyFeatureType_Keyword:
		return fieldType == interfaces.DataType_String
	case interfaces.PropertyFeatureType_Vector:
		return fieldType == interfaces.DataType_Vector
	default:
		return false
	}
}
