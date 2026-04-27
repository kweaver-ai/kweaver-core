// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// RiskTypeAccess 风险类数据访问接口
//
//go:generate mockgen -source risk_type_access.go -destination mock/mock_risk_type_access.go
type RiskTypeAccess interface {
	CheckRiskTypeExistByID(ctx context.Context, knID string, branch string, rtID string) (string, bool, error)
	CheckRiskTypeExistByName(ctx context.Context, knID string, branch string, rtName string) (string, bool, error)
	CreateRiskType(ctx context.Context, tx *sql.Tx, riskType *RiskType) error
	ListRiskTypes(ctx context.Context, query RiskTypesQueryParams) ([]*RiskType, error)
	GetRiskTypesTotal(ctx context.Context, query RiskTypesQueryParams) (int, error)
	GetRiskTypesByIDs(ctx context.Context, knID string, branch string, rtIDs []string) ([]*RiskType, error)
	UpdateRiskType(ctx context.Context, tx *sql.Tx, riskType *RiskType) error
	DeleteRiskTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, rtIDs []string) (int64, error)
	GetAllRiskTypesByKnID(ctx context.Context, knID string, branch string) ([]*RiskType, error)
	DeleteRiskTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error)
}
