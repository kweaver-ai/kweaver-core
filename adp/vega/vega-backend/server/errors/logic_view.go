// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package errors Resource 模块错误码
package errors

// Resource 错误码
const (
	// LogicView 校验相关
	VegaBackend_LogicView_InvalidParameter_JoinType          = "VegaBackend.LogicView.InvalidParameter.JoinType"
	VegaBackend_LogicView_InvalidParameter_LogicDefinition   = "VegaBackend.LogicView.InvalidParameter.LogicDefinition"
	VegaBackend_LogicView_Duplicated_NodeID                  = "VegaBackend.LogicView.Duplicated.NodeID"
	VegaBackend_LogicView_InvalidParameter_FieldName         = "VegaBackend.LogicView.InvalidParameter.FieldName"
	VegaBackend_LogicView_LengthExceeded_FieldName           = "VegaBackend.LogicView.LengthExceeded.FieldName"
	VegaBackend_LogicView_LengthExceeded_FieldDisplayName    = "VegaBackend.LogicView.LengthExceeded.FieldDisplayName"
	VegaBackend_LogicView_LengthExceeded_FieldComment        = "VegaBackend.LogicView.LengthExceeded.FieldComment"
	VegaBackend_LogicView_Duplicated_FieldName               = "VegaBackend.LogicView.Duplicated.FieldName"
	VegaBackend_LogicView_Duplicated_FieldDisplayName        = "VegaBackend.LogicView.Duplicated.FieldDisplayName"
	VegaBackend_LogicView_InvalidParameter_FieldFeatureName  = "VegaBackend.LogicView.InvalidParameter.FieldFeatureName"
	VegaBackend_LogicView_LengthExceeded_FieldFeatureName    = "VegaBackend.LogicView.LengthExceeded.FieldFeatureName"
	VegaBackend_LogicView_Duplicated_FieldFeatureName        = "VegaBackend.LogicView.Duplicated.FieldFeatureName"
	VegaBackend_LogicView_LengthExceeded_FieldFeatureComment = "VegaBackend.LogicView.LengthExceeded.FieldFeatureComment"
)

var LogicViewErrCodeList = []string{
	VegaBackend_LogicView_InvalidParameter_JoinType,
	VegaBackend_LogicView_InvalidParameter_LogicDefinition,
	VegaBackend_LogicView_Duplicated_NodeID,
	VegaBackend_LogicView_InvalidParameter_FieldName,
	VegaBackend_LogicView_LengthExceeded_FieldName,
	VegaBackend_LogicView_LengthExceeded_FieldDisplayName,
	VegaBackend_LogicView_LengthExceeded_FieldComment,
	VegaBackend_LogicView_Duplicated_FieldName,
	VegaBackend_LogicView_Duplicated_FieldDisplayName,
	VegaBackend_LogicView_InvalidParameter_FieldFeatureName,
	VegaBackend_LogicView_LengthExceeded_FieldFeatureName,
	VegaBackend_LogicView_Duplicated_FieldFeatureName,
	VegaBackend_LogicView_LengthExceeded_FieldFeatureComment,
}
