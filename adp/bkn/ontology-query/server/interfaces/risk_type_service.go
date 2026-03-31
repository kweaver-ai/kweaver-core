// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// RiskTypeEvalResult 风险评估结果
type RiskTypeEvalResult struct {
	Allow   bool
	Message string
}

//go:generate mockgen -source risk_type_service.go -destination mock/mock_risk_type_service.go
type RiskTypeService interface {
	// Evaluate 对 ActionType 进行风险评估
	Evaluate(ctx context.Context, actionType *ActionType, knID string, branch string) (*RiskTypeEvalResult, error)
	// MustAllow 若风险评估返回 disallow 则返回错误
	MustAllow(ctx context.Context, actionType *ActionType, knID string, branch string) error
}
