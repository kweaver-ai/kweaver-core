// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 业务知识网络错误码
const (
	// 400
	BknBackend_KnowledgeNetwork_Duplicated_Name                    = "BknBackend.KnowledgeNetwork.Duplicated.Name"
	BknBackend_KnowledgeNetwork_InvalidParameter                   = "BknBackend.KnowledgeNetwork.InvalidParameter"
	BknBackend_KnowledgeNetwork_InvalidParameter_BusinessDomain    = "BknBackend.KnowledgeNetwork.InvalidParameter.BusinessDomain"
	BknBackend_KnowledgeNetwork_InvalidParameter_ConceptCondition  = "BknBackend.KnowledgeNetwork.InvalidParameter.ConceptCondition"
	BknBackend_KnowledgeNetwork_InvalidParameter_Direction         = "BknBackend.KnowledgeNetwork.InvalidParameter.Direction"
	BknBackend_KnowledgeNetwork_InvalidParameter_IncludeStatistics = "BknBackend.KnowledgeNetwork.InvalidParameter.IncludeStatistics"
	BknBackend_KnowledgeNetwork_InvalidParameter_IncludeTypeInfo   = "BknBackend.KnowledgeNetwork.InvalidParameter.IncludeTypeInfo"
	BknBackend_KnowledgeNetwork_InvalidParameter_PathLength        = "BknBackend.KnowledgeNetwork.InvalidParameter.PathLength"
	BknBackend_KnowledgeNetwork_KNIDExisted                        = "BknBackend.KnowledgeNetwork.KNIDExisted"
	BknBackend_KnowledgeNetwork_KNNameExisted                      = "BknBackend.KnowledgeNetwork.KNNameExisted"
	BknBackend_KnowledgeNetwork_LengthExceeded_Name                = "BknBackend.KnowledgeNetwork.LengthExceeded.Name"
	BknBackend_KnowledgeNetwork_NullParameter_Branch               = "BknBackend.KnowledgeNetwork.NullParameter.Branch"
	BknBackend_KnowledgeNetwork_NullParameter_Direction            = "BknBackend.KnowledgeNetwork.NullParameter.Direction"
	BknBackend_KnowledgeNetwork_NullParameter_Name                 = "BknBackend.KnowledgeNetwork.NullParameter.Name"
	BknBackend_KnowledgeNetwork_NullParameter_SourceObjectTypeId   = "BknBackend.KnowledgeNetwork.NullParameter.SourceObjectTypeId"

	// 403
	BknBackend_KnowledgeNetwork_Forbidden_HasRunningJob = "BknBackend.KnowledgeNetwork.Forbidden.HasRunningJob"

	// 404
	BknBackend_KnowledgeNetwork_NotFound = "BknBackend.KnowledgeNetwork.NotFound"

	// 500
	BknBackend_KnowledgeNetwork_InternalError                             = "BknBackend.KnowledgeNetwork.InternalError"
	BknBackend_KnowledgeNetwork_InternalError_BeginTransactionFailed      = "BknBackend.KnowledgeNetwork.InternalError.BeginTransactionFailed"
	BknBackend_KnowledgeNetwork_InternalError_BindBusinessDomainFailed    = "BknBackend.KnowledgeNetwork.InternalError.BindBusinessDomainFailed"
	BknBackend_KnowledgeNetwork_InternalError_UnbindBusinessDomainFailed  = "BknBackend.KnowledgeNetwork.InternalError.UnbindBusinessDomainFailed"
	BknBackend_KnowledgeNetwork_InternalError_CheckKNIfExistFailed        = "BknBackend.KnowledgeNetwork.InternalError.CheckKNIfExistFailed"
	BknBackend_KnowledgeNetwork_InternalError_GetKNByIDFailed             = "BknBackend.KnowledgeNetwork.InternalError.GetKNByIDFailed"
	BknBackend_KnowledgeNetwork_InternalError_UpdateKNFailed              = "BknBackend.KnowledgeNetwork.InternalError.UpdateKNFailed"
	BknBackend_KnowledgeNetwork_InternalError_CreateKNFailed              = "BknBackend.KnowledgeNetwork.InternalError.CreateKNFailed"
	BknBackend_KnowledgeNetwork_InternalError_CreateResourcesFailed       = "BknBackend.KnowledgeNetwork.InternalError.CreateResourcesFailed"
	BknBackend_KnowledgeNetwork_InternalError_GetActionTypesTotalFailed   = "BknBackend.KnowledgeNetwork.InternalError.GetActionTypesTotalFailed"
	BknBackend_KnowledgeNetwork_InternalError_GetObjectTypesTotalFailed   = "BknBackend.KnowledgeNetwork.InternalError.GetObjectTypesTotalFailed"
	BknBackend_KnowledgeNetwork_InternalError_GetRelationTypesTotalFailed = "BknBackend.KnowledgeNetwork.InternalError.GetRelationTypesTotalFailed"
	BknBackend_KnowledgeNetwork_InternalError_GetRiskTypesTotalFailed     = "BknBackend.KnowledgeNetwork.InternalError.GetRiskTypesTotalFailed"
	BknBackend_KnowledgeNetwork_InternalError_GetVectorFailed             = "BknBackend.KnowledgeNetwork.InternalError.GetVectorFailed"
	BknBackend_KnowledgeNetwork_InternalError_InsertOpenSearchDataFailed  = "BknBackend.KnowledgeNetwork.InternalError.InsertOpenSearchDataFailed"
	BknBackend_KnowledgeNetwork_InternalError_CreateObjectTypesFailed     = "BknBackend.KnowledgeNetwork.InternalError.CreateObjectTypesFailed"
	BknBackend_KnowledgeNetwork_InternalError_CreateRelationTypesFailed   = "BknBackend.KnowledgeNetwork.InternalError.CreateRelationTypesFailed"
	BknBackend_KnowledgeNetwork_InternalError_CreateActionTypesFailed     = "BknBackend.KnowledgeNetwork.InternalError.CreateActionTypesFailed"
	BknBackend_KnowledgeNetwork_InternalError_DeleteObjectTypesFailed     = "BknBackend.KnowledgeNetwork.InternalError.DeleteObjectTypesFailed"
	BknBackend_KnowledgeNetwork_InternalError_DeleteRelationTypesFailed   = "BknBackend.KnowledgeNetwork.InternalError.DeleteRelationTypesFailed"
	BknBackend_KnowledgeNetwork_InternalError_DeleteActionTypesFailed     = "BknBackend.KnowledgeNetwork.InternalError.DeleteActionTypesFailed"
)

var (
	KNErrCodeList = []string{
		// 400
		BknBackend_KnowledgeNetwork_Duplicated_Name,
		BknBackend_KnowledgeNetwork_InvalidParameter,
		BknBackend_KnowledgeNetwork_InvalidParameter_BusinessDomain,
		BknBackend_KnowledgeNetwork_InvalidParameter_ConceptCondition,
		BknBackend_KnowledgeNetwork_InvalidParameter_Direction,
		BknBackend_KnowledgeNetwork_InvalidParameter_IncludeStatistics,
		BknBackend_KnowledgeNetwork_InvalidParameter_IncludeTypeInfo,
		BknBackend_KnowledgeNetwork_InvalidParameter_PathLength,
		BknBackend_KnowledgeNetwork_KNIDExisted,
		BknBackend_KnowledgeNetwork_KNNameExisted,
		BknBackend_KnowledgeNetwork_LengthExceeded_Name,
		BknBackend_KnowledgeNetwork_NullParameter_Branch,
		BknBackend_KnowledgeNetwork_NullParameter_Direction,
		BknBackend_KnowledgeNetwork_NullParameter_Name,
		BknBackend_KnowledgeNetwork_NullParameter_SourceObjectTypeId,

		// 403
		BknBackend_KnowledgeNetwork_Forbidden_HasRunningJob,

		// 404
		BknBackend_KnowledgeNetwork_NotFound,

		// 500
		BknBackend_KnowledgeNetwork_InternalError,
		BknBackend_KnowledgeNetwork_InternalError_CheckKNIfExistFailed,
		BknBackend_KnowledgeNetwork_InternalError_BeginTransactionFailed,
		BknBackend_KnowledgeNetwork_InternalError_BindBusinessDomainFailed,
		BknBackend_KnowledgeNetwork_InternalError_UnbindBusinessDomainFailed,
		BknBackend_KnowledgeNetwork_InternalError_GetKNByIDFailed,
		BknBackend_KnowledgeNetwork_InternalError_UpdateKNFailed,
		BknBackend_KnowledgeNetwork_InternalError_CreateKNFailed,
		BknBackend_KnowledgeNetwork_InternalError_CreateResourcesFailed,
		BknBackend_KnowledgeNetwork_InternalError_GetActionTypesTotalFailed,
		BknBackend_KnowledgeNetwork_InternalError_GetObjectTypesTotalFailed,
		BknBackend_KnowledgeNetwork_InternalError_GetRelationTypesTotalFailed,
		BknBackend_KnowledgeNetwork_InternalError_GetRiskTypesTotalFailed,
		BknBackend_KnowledgeNetwork_InternalError_GetVectorFailed,
		BknBackend_KnowledgeNetwork_InternalError_InsertOpenSearchDataFailed,
		BknBackend_KnowledgeNetwork_InternalError_CreateObjectTypesFailed,
		BknBackend_KnowledgeNetwork_InternalError_CreateRelationTypesFailed,
		BknBackend_KnowledgeNetwork_InternalError_CreateActionTypesFailed,
		BknBackend_KnowledgeNetwork_InternalError_DeleteObjectTypesFailed,
		BknBackend_KnowledgeNetwork_InternalError_DeleteRelationTypesFailed,
		BknBackend_KnowledgeNetwork_InternalError_DeleteActionTypesFailed,
	}
)
