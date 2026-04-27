// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package errors

const (
	BknBackend_Metric_InvalidParameter                       = "BknBackend.Metric.InvalidParameter"
	BknBackend_Metric_NullParameter_Name                     = "BknBackend.Metric.NullParameter.Name"
	BknBackend_Metric_LengthExceeded_Name                    = "BknBackend.Metric.LengthExceeded.Name"
	BknBackend_Metric_InternalError                          = "BknBackend.Metric.InternalError"
	BknBackend_Metric_InternalError_BeginTransactionFailed   = "BknBackend.Metric.InternalError.BeginTransactionFailed"
	BknBackend_Metric_InternalError_CheckMetricIfExistFailed = "BknBackend.Metric.InternalError.CheckMetricIfExistFailed"
	BknBackend_Metric_InternalError_GetMetricsByIDsFailed    = "BknBackend.Metric.InternalError.GetMetricsByIDsFailed"
	BknBackend_Metric_NotFound                               = "BknBackend.Metric.NotFound"
	BknBackend_Metric_Duplicated_Name                        = "BknBackend.Metric.Duplicated.Name"
	BknBackend_Metric_InvalidMetricType                      = "BknBackend.Metric.InvalidMetricType"
)

var MetricErrCodeList = []string{
	BknBackend_Metric_InvalidParameter,
	BknBackend_Metric_NullParameter_Name,
	BknBackend_Metric_LengthExceeded_Name,
	BknBackend_Metric_InternalError,
	BknBackend_Metric_InternalError_BeginTransactionFailed,
	BknBackend_Metric_InternalError_CheckMetricIfExistFailed,
	BknBackend_Metric_InternalError_GetMetricsByIDsFailed,
	BknBackend_Metric_NotFound,
	BknBackend_Metric_Duplicated_Name,
	BknBackend_Metric_InvalidMetricType,
}
