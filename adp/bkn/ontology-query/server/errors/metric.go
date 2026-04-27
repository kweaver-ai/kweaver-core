// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

const (
	// 400
	OntologyQuery_Metric_InvalidParameter  = "OntologyQuery.Metric.InvalidParameter"
	OntologyQuery_Metric_InvalidDataSource = "OntologyQuery.Metric.InvalidDataSource"
	OntologyQuery_Metric_UnsupportedScope  = "OntologyQuery.Metric.UnsupportedScope"

	// 404
	OntologyQuery_Metric_NotFound           = "OntologyQuery.Metric.NotFound"
	OntologyQuery_Metric_ObjectTypeNotFound = "OntologyQuery.Metric.ObjectTypeNotFound"

	// 500
	OntologyQuery_Metric_InternalError_QueryFailed = "OntologyQuery.Metric.InternalError.QueryFailed"
)

var metricErrCodeList = []string{
	OntologyQuery_Metric_InvalidParameter,
	OntologyQuery_Metric_InvalidDataSource,
	OntologyQuery_Metric_UnsupportedScope,
	OntologyQuery_Metric_NotFound,
	OntologyQuery_Metric_ObjectTypeNotFound,
	OntologyQuery_Metric_InternalError_QueryFailed,
}
