// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// RiskTypeService 风险类服务接口
//
//go:generate mockgen -source risk_type_service.go -destination mock/mock_risk_type_service.go
type RiskTypeService interface {
	CheckRiskTypeExistByID(ctx context.Context, knID string, branch string, rtID string) (string, bool, error)
	CheckRiskTypeExistByName(ctx context.Context, knID string, branch string, rtName string) (string, bool, error)
	CreateRiskTypes(ctx context.Context, tx *sql.Tx, riskTypes []*RiskType, mode string) ([]string, error)
	ListRiskTypes(ctx context.Context, query RiskTypesQueryParams) ([]*RiskType, int, error)
	GetRiskTypesByIDs(ctx context.Context, knID string, branch string, rtIDs []string) ([]*RiskType, error)
	UpdateRiskType(ctx context.Context, tx *sql.Tx, riskType *RiskType) error
	DeleteRiskTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, rtIDs []string) error
	GetAllRiskTypesByKnID(ctx context.Context, knID string, branch string) ([]*RiskType, error)
	DeleteRiskTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error

	SearchRiskTypes(ctx context.Context, query *ConceptsQuery) (RiskTypes, error)
	InsertDatasetData(ctx context.Context, riskTypes []*RiskType) error
}
