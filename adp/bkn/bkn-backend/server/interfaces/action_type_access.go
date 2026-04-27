// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/action_type_access.go -destination ../interfaces/mock/mock_action_type_access.go
type ActionTypeAccess interface {
	CheckActionTypeExistByID(ctx context.Context, knID string, branch string, atID string) (string, bool, error)
	CheckActionTypeExistByName(ctx context.Context, knID string, branch string, atName string) (string, bool, error)

	CreateActionType(ctx context.Context, tx *sql.Tx, actionType *ActionType) error
	ListActionTypes(ctx context.Context, query ActionTypesQueryParams) ([]*ActionType, error)
	GetActionTypesTotal(ctx context.Context, query ActionTypesQueryParams) (int, error)
	GetActionTypesByIDs(ctx context.Context, knID string, branch string, atIDs []string) ([]*ActionType, error)
	UpdateActionType(ctx context.Context, tx *sql.Tx, actionType *ActionType) error
	DeleteActionTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, atIDs []string) (int64, error)

	GetAllActionTypesByKnID(ctx context.Context, knID string, branch string) (map[string]*ActionType, error)
	GetActionTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error)
	DeleteActionTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error)
}
