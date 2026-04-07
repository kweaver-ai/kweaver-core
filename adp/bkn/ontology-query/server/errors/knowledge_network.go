// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// 指标模型
const (
	// 400
	OntologyQuery_KnowledgeNetwork_NullParameter_Direction             = "OntologyQuery.KnowledgeNetwork.NullParameter.Direction"
	OntologyQuery_KnowledgeNetwork_NullParameter_SourceObjectTypeId    = "OntologyQuery.KnowledgeNetwork.NullParameter.SourceObjectTypeId"
	OntologyQuery_KnowledgeNetwork_NullParameter_TypePathObjectTypes   = "OntologyQuery.KnowledgeNetwork.NullParameter.TypePathObjectTypes"
	OntologyQuery_KnowledgeNetwork_NullParameter_TypePathRelationTypes = "OntologyQuery.KnowledgeNetwork.NullParameter.TypePathRelationTypes"
	OntologyQuery_KnowledgeNetwork_InvalidParameter                    = "OntologyQuery.KnowledgeNetwork.InvalidParameter"
	OntologyQuery_KnowledgeNetwork_InvalidParameter_Direction          = "OntologyQuery.KnowledgeNetwork.InvalidParameter.Direction"
	OntologyQuery_KnowledgeNetwork_InvalidParameter_IncludeTypeInfo    = "OntologyQuery.KnowledgeNetwork.InvalidParameter.IncludeTypeInfo"
	OntologyQuery_KnowledgeNetwork_InvalidParameter_PathLength         = "OntologyQuery.KnowledgeNetwork.InvalidParameter.PathLength"
	OntologyQuery_KnowledgeNetwork_InvalidParameter_TypePath           = "OntologyQuery.KnowledgeNetwork.InvalidParameter.TypePath"
	// OntologyQuery_KnowledgeNetwork_UnsupportLogicPropertyType       = "OntologyQuery.KnowledgeNetwork.UnsupportLogicPropertyType"

	//404
	OntologyQuery_KnowledgeNetwork_KnowledgeNetworkNotFound = "OntologyQuery.KnowledgeNetwork.KnowledgeNetworkNotFound"
	OntologyQuery_KnowledgeNetwork_RelationTypeNotFound     = "OntologyQuery.KnowledgeNetwork.RelationTypeNotFound"

	// 500
	OntologyQuery_KnowledgeNetwork_InternalError_GetViewDataByIDFailed          = "OntologyQuery.KnowledgeNetwork.InternalError.GetViewDataByIDFailed"
	OntologyQuery_KnowledgeNetwork_InternalError_GetKnowledgeNetworksByIDFailed = "OntologyQuery.KnowledgeNetwork.InternalError.GetKnowledgeNetworksByIDFailed"
	OntologyQuery_KnowledgeNetwork_InternalError_GetRelationTypeFailed          = "OntologyQuery.KnowledgeNetwork.InternalError.GetRelationTypeFailed"
)

var (
	knowledgeNetworkErrCodeList = []string{
		// 400
		OntologyQuery_KnowledgeNetwork_NullParameter_Direction,
		OntologyQuery_KnowledgeNetwork_NullParameter_SourceObjectTypeId,
		OntologyQuery_KnowledgeNetwork_NullParameter_TypePathObjectTypes,
		OntologyQuery_KnowledgeNetwork_NullParameter_TypePathRelationTypes,
		OntologyQuery_KnowledgeNetwork_InvalidParameter,
		OntologyQuery_KnowledgeNetwork_InvalidParameter_Direction,
		OntologyQuery_KnowledgeNetwork_InvalidParameter_IncludeTypeInfo,
		OntologyQuery_KnowledgeNetwork_InvalidParameter_PathLength,
		OntologyQuery_KnowledgeNetwork_InvalidParameter_TypePath,

		// 404
		OntologyQuery_KnowledgeNetwork_KnowledgeNetworkNotFound,
		OntologyQuery_KnowledgeNetwork_RelationTypeNotFound,

		// 500
		OntologyQuery_KnowledgeNetwork_InternalError_GetViewDataByIDFailed,
		OntologyQuery_KnowledgeNetwork_InternalError_GetKnowledgeNetworksByIDFailed,
		OntologyQuery_KnowledgeNetwork_InternalError_GetRelationTypeFailed,
	}
)
