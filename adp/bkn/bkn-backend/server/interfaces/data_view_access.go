// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

type DataView struct {
	ViewID    string                `json:"id"`
	ViewName  string                `json:"name"`
	QueryType string                `json:"query_type"`
	Fields    []*ViewField          `json:"fields"`
	FieldsMap map[string]*ViewField `json:"-"`
}

// 数据视图字段
type ViewField struct {
	Name              string       `json:"name"`
	Type              string       `json:"type"`
	Comment           string       `json:"comment"`
	DisplayName       string       `json:"display_name"`
	OriginalName      string       `json:"original_name"`
	DataLength        int32        `json:"data_length"`
	DataAccuracy      int32        `json:"data_accuracy"`
	Status            string       `json:"status"`
	IsNullable        string       `json:"is_nullable"`
	BusinessTimestamp bool         `json:"business_timestamp"`
	PrimaryKey        sql.NullBool `json:"-"`
}

type ViewQueryResult struct {
	View        *DataView        `json:"view"`
	TotalCount  int64            `json:"total_count"`
	SearchAfter []any            `json:"search_after"`
	Entries     []map[string]any `json:"entries"`
}

//go:generate mockgen -source ../interfaces/data_view_access.go -destination ../interfaces/mock/mock_data_view_access.go
type DataViewAccess interface {
	GetDataViewByID(ctx context.Context, id string) (*DataView, error)
	GetDataStart(ctx context.Context, id string, incKey string, incValue any, limit int) (*ViewQueryResult, error)
	GetDataNext(ctx context.Context, id string, searchAfter []any, limit int) (*ViewQueryResult, error)
}
