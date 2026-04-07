// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package table provides base implementation for relational database connectors.
package table

import (
	"database/sql"

	"vega-backend/interfaces"
)

// convertValue converts []byte to string for MariaDB driver compatibility
func convertValue(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

// ScanRows scans SQL rows into QueryResult.
func ScanRows(rows *sql.Rows) (*interfaces.QueryResult, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	result := &interfaces.QueryResult{
		Columns: columns,
		Rows:    make([]map[string]any, 0),
	}

	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = convertValue(values[i])
		}
		result.Rows = append(result.Rows, row)
	}

	result.Total = int64(len(result.Rows))
	return result, nil
}
