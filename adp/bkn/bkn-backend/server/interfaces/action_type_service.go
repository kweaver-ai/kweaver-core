// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/action_type_service.go -destination ../interfaces/mock/mock_action_type_service.go
type ActionTypeService interface {
	CheckActionTypeExistByID(ctx context.Context, knID string, branch string, atID string) (string, bool, error)
	CheckActionTypeExistByName(ctx context.Context, knID string, branch string, atName string) (string, bool, error)
	CreateActionTypes(ctx context.Context, tx *sql.Tx, actionTypes []*ActionType, mode string, strictMode bool) ([]string, error)
	ListActionTypes(ctx context.Context, query ActionTypesQueryParams) ([]*ActionType, int, error)
	GetActionTypesByIDs(ctx context.Context, knID string, branch string, atIDs []string) ([]*ActionType, error)
	UpdateActionType(ctx context.Context, tx *sql.Tx, actionType *ActionType, strictMode bool) error
	DeleteActionTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, atIDs []string) error

	GetActionTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error)
	DeleteActionTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error

	SearchActionTypes(ctx context.Context, query *ConceptsQuery) (ActionTypes, error)

	// 写行动类到索引
	InsertDatasetData(ctx context.Context, actionTypes []*ActionType) error

	// ValidateActionTypes 仅校验依赖存在性，不写库
	ValidateActionTypes(ctx context.Context, knID string, branch string, actionTypes []*ActionType, strictMode bool, batch *BatchIDIndex, mode string) error
}
