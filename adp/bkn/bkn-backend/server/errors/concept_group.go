// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 概念分组错误码
const (
	// 400
	BknBackend_ConceptGroup_Duplicated_Name                    = "BknBackend.ConceptGroup.Duplicated.Name"
	BknBackend_ConceptGroup_InvalidParameter                   = "BknBackend.ConceptGroup.InvalidParameter"
	BknBackend_ConceptGroup_InvalidParameter_ConceptCondition  = "BknBackend.ConceptGroup.InvalidParameter.ConceptCondition"
	BknBackend_ConceptGroup_InvalidParameter_Direction         = "BknBackend.ConceptGroup.InvalidParameter.Direction"
	BknBackend_ConceptGroup_InvalidParameter_IncludeStatistics = "BknBackend.ConceptGroup.InvalidParameter.IncludeStatistics"
	BknBackend_ConceptGroup_InvalidParameter_IncludeTypeInfo   = "BknBackend.ConceptGroup.InvalidParameter.IncludeTypeInfo"
	BknBackend_ConceptGroup_InvalidParameter_PathLength        = "BknBackend.ConceptGroup.InvalidParameter.PathLength"
	BknBackend_ConceptGroup_ConceptGroupIDExisted              = "BknBackend.ConceptGroup.ConceptGroupIDExisted"
	BknBackend_ConceptGroup_ConceptGroupNameExisted            = "BknBackend.ConceptGroup.ConceptGroupNameExisted"
	BknBackend_ConceptGroup_ConceptGroupRelationExisted        = "BknBackend.ConceptGroup.ConceptGroupRelationExisted"
	BknBackend_ConceptGroup_ConceptGroupRelationNotExisted     = "BknBackend.ConceptGroup.ConceptGroupRelationNotExisted"
	BknBackend_ConceptGroup_LengthExceeded_Name                = "BknBackend.ConceptGroup.LengthExceeded.Name"
	BknBackend_ConceptGroup_NullParameter_Direction            = "BknBackend.ConceptGroup.NullParameter.Direction"
	BknBackend_ConceptGroup_NullParameter_Name                 = "BknBackend.ConceptGroup.NullParameter.Name"
	BknBackend_ConceptGroup_NullParameter_SourceObjectTypeId   = "BknBackend.ConceptGroup.NullParameter.SourceObjectTypeId"

	// 404
	BknBackend_ConceptGroup_ConceptGroupNotFound = "BknBackend.ConceptGroup.ConceptGroupNotFound"
	BknBackend_ConceptGroup_ObjectTypeNotFound   = "BknBackend.ConceptGroup.ObjectTypeNotFound"

	// 500
	BknBackend_ConceptGroup_InternalError                                      = "BknBackend.ConceptGroup.InternalError"
	BknBackend_ConceptGroup_InternalError_AddObjectTypesToConceptGroupFailed   = "BknBackend.ConceptGroup.InternalError.AddObjectTypesToConceptGroupFailed"
	BknBackend_ConceptGroup_InternalError_MissingTransaction                   = "BknBackend.ConceptGroup.InternalError.MissingTransaction"
	BknBackend_ConceptGroup_InternalError_BeginTransactionFailed               = "BknBackend.ConceptGroup.InternalError.BeginTransactionFailed"
	BknBackend_ConceptGroup_InternalError_BindBusinessDomainFailed             = "BknBackend.ConceptGroup.InternalError.BindBusinessDomainFailed"
	BknBackend_ConceptGroup_InternalError_UnbindBusinessDomainFailed           = "BknBackend.ConceptGroup.InternalError.UnbindBusinessDomainFailed"
	BknBackend_ConceptGroup_InternalError_CheckConceptGroupIfExistFailed       = "BknBackend.ConceptGroup.InternalError.CheckConceptGroupIfExistFailed"
	BknBackend_ConceptGroup_InternalError_GetConceptGroupByIDFailed            = "BknBackend.ConceptGroup.InternalError.GetConceptGroupByIDFailed"
	BknBackend_ConceptGroup_InternalError_UpdateConceptGroupFailed             = "BknBackend.ConceptGroup.InternalError.UpdateConceptGroupFailed"
	BknBackend_ConceptGroup_InternalError_CreateConceptGroupFailed             = "BknBackend.ConceptGroup.InternalError.CreateConceptGroupFailed"
	BknBackend_ConceptGroup_InternalError_CreateConceptGroupRelationFailed     = "BknBackend.ConceptGroup.InternalError.CreateConceptGroupRelationFailed"
	BknBackend_ConceptGroup_InternalError_GetActionTypesTotalFailed            = "BknBackend.ConceptGroup.InternalError.GetActionTypesTotalFailed"
	BknBackend_ConceptGroup_InternalError_GetConceptIDsByConceptGroupIDsFailed = "BknBackend.ConceptGroup.InternalError.GetConceptIDsByConceptGroupIDsFailed"
	BknBackend_ConceptGroup_InternalError_GetRelationTypesTotalFailed          = "BknBackend.ConceptGroup.InternalError.GetRelationTypesTotalFailed"
	BknBackend_ConceptGroup_InternalError_GetVectorFailed                      = "BknBackend.ConceptGroup.InternalError.GetVectorFailed"
	BknBackend_ConceptGroup_InternalError_InsertOpenSearchDataFailed           = "BknBackend.ConceptGroup.InternalError.InsertOpenSearchDataFailed"
	BknBackend_ConceptGroup_InternalError_CreateObjectTypesFailed              = "BknBackend.ConceptGroup.InternalError.CreateObjectTypesFailed"
	BknBackend_ConceptGroup_InternalError_CreateRelationTypesFailed            = "BknBackend.ConceptGroup.InternalError.CreateRelationTypesFailed"
	BknBackend_ConceptGroup_InternalError_CreateActionTypesFailed              = "BknBackend.ConceptGroup.InternalError.CreateActionTypesFailed"
)

var (
	ConceptGroupErrCodeList = []string{
		// 400
		BknBackend_ConceptGroup_Duplicated_Name,
		BknBackend_ConceptGroup_InvalidParameter,
		BknBackend_ConceptGroup_InvalidParameter_ConceptCondition,
		BknBackend_ConceptGroup_InvalidParameter_Direction,
		BknBackend_ConceptGroup_InvalidParameter_IncludeStatistics,
		BknBackend_ConceptGroup_InvalidParameter_IncludeTypeInfo,
		BknBackend_ConceptGroup_InvalidParameter_PathLength,
		BknBackend_ConceptGroup_ConceptGroupIDExisted,
		BknBackend_ConceptGroup_ConceptGroupNameExisted,
		BknBackend_ConceptGroup_ConceptGroupRelationExisted,
		BknBackend_ConceptGroup_ConceptGroupRelationNotExisted,
		BknBackend_ConceptGroup_LengthExceeded_Name,
		BknBackend_ConceptGroup_NullParameter_Direction,
		BknBackend_ConceptGroup_NullParameter_Name,
		BknBackend_ConceptGroup_NullParameter_SourceObjectTypeId,

		// 404
		BknBackend_ConceptGroup_ConceptGroupNotFound,
		BknBackend_ConceptGroup_ObjectTypeNotFound,

		// 500
		BknBackend_ConceptGroup_InternalError,
		BknBackend_ConceptGroup_InternalError_AddObjectTypesToConceptGroupFailed,
		BknBackend_ConceptGroup_InternalError_MissingTransaction,
		BknBackend_ConceptGroup_InternalError_CheckConceptGroupIfExistFailed,
		BknBackend_ConceptGroup_InternalError_BeginTransactionFailed,
		BknBackend_ConceptGroup_InternalError_BindBusinessDomainFailed,
		BknBackend_ConceptGroup_InternalError_UnbindBusinessDomainFailed,
		BknBackend_ConceptGroup_InternalError_GetConceptGroupByIDFailed,
		BknBackend_ConceptGroup_InternalError_UpdateConceptGroupFailed,
		BknBackend_ConceptGroup_InternalError_CreateConceptGroupFailed,
		BknBackend_ConceptGroup_InternalError_CreateConceptGroupRelationFailed,
		BknBackend_ConceptGroup_InternalError_GetActionTypesTotalFailed,
		BknBackend_ConceptGroup_InternalError_GetConceptIDsByConceptGroupIDsFailed,
		BknBackend_ConceptGroup_InternalError_GetRelationTypesTotalFailed,
		BknBackend_ConceptGroup_InternalError_GetVectorFailed,
		BknBackend_ConceptGroup_InternalError_InsertOpenSearchDataFailed,
		BknBackend_ConceptGroup_InternalError_CreateObjectTypesFailed,
		BknBackend_ConceptGroup_InternalError_CreateRelationTypesFailed,
		BknBackend_ConceptGroup_InternalError_CreateActionTypesFailed,
	}
)
