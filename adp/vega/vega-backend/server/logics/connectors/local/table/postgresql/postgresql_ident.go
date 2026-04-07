// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package postgresql

import (
	"strings"

	"vega-backend/interfaces"
)

// pgQuoteIdent 双引号转义 PostgreSQL 标识符。
func pgQuoteIdent(s string) string {
	if s == "" {
		return `""`
	}
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// qualTable 返回 schema.table 形式的双引号限定表名。
func qualTable(res *interfaces.Resource) string {
	ident := res.SourceIdentifier
	parts := strings.SplitN(ident, ".", 2)
	for i, p := range parts {
		parts[i] = pgQuoteIdent(strings.TrimSpace(p))
	}
	return strings.Join(parts, ".")
}

// quoteColumnName 列名/别名限定；支持 "alias.col" -> "alias"."col"。
func quoteColumnName(name string) string {
	if name == "" {
		return `""`
	}
	if idx := strings.Index(name, "."); idx >= 0 {
		alias := strings.TrimSpace(name[:idx])
		col := strings.TrimSpace(name[idx+1:])
		return pgQuoteIdent(alias) + "." + pgQuoteIdent(col)
	}
	return pgQuoteIdent(strings.TrimSpace(name))
}

// qualifyFieldForSQL 将 "alias.col" 转为带引号的限定名（用于 SELECT/ORDER BY/JOIN ON）。
func qualifyFieldForSQL(field string) string {
	field = strings.TrimSpace(field)
	if idx := strings.Index(field, "."); idx >= 0 {
		alias := field[:idx]
		col := field[idx+1:]
		return pgQuoteIdent(strings.TrimSpace(alias)) + "." + pgQuoteIdent(strings.TrimSpace(col))
	}
	return pgQuoteIdent(field)
}
