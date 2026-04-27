// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

// RiskType evaluation and related errors (ontology-query)
const (
	// 403 — policy outcome: evaluated risk level exceeds configured maximum
	OntologyQuery_RiskType_RiskLevelExceedsMaxAcceptable = "OntologyQuery.RiskType.RiskLevelExceedsMaxAcceptable"

	// 404
	OntologyQuery_RiskType_RiskTypeNotFound = "OntologyQuery.RiskType.RiskTypeNotFound"

	// 500
	OntologyQuery_RiskType_InternalError_GetRiskTypesByIDsFailed         = "OntologyQuery.RiskType.InternalError.GetRiskTypesByIDsFailed"
	OntologyQuery_RiskType_InternalError_ExecuteRiskAssessmentToolFailed = "OntologyQuery.RiskType.InternalError.ExecuteRiskAssessmentToolFailed"
)

var (
	riskTypeErrCodeList = []string{
		OntologyQuery_RiskType_RiskLevelExceedsMaxAcceptable,
		OntologyQuery_RiskType_RiskTypeNotFound,
		OntologyQuery_RiskType_InternalError_GetRiskTypesByIDsFailed,
		OntologyQuery_RiskType_InternalError_ExecuteRiskAssessmentToolFailed,
	}
)
