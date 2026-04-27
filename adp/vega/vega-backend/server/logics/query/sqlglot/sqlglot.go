// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package sqlglot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"vega-backend/interfaces"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

// ExtractTablesResult 存储表提取结果
type ExtractTablesResult struct {
	Tables  []*Table `json:"tables"`
	SQL     string   `json:"sql"`
	Dialect string   `json:"dialect"`
	Error   string   `json:"error"`
}

type Table struct {
	Catalog string `json:"catalog"`
	Schema  string `json:"schema"`
	Name    string `json:"name"`
}

// SQLParseResult 存储SQL解析结果
type SQLParseResult struct {
	AST     interface{} `json:"ast"`
	SQL     string      `json:"sql"`
	Dialect string      `json:"dialect"`
	Error   string      `json:"error"`
}

// ExtractTables 从SQL中提取所有表名
func ExtractTables(sql string, dialect string) (*ExtractTablesResult, error) {
	cmd := exec.Command("python3", "-c", `
import sys
import json
import sqlglot
from sqlglot.expressions import Table

try:
    sql = sys.argv[1]
    dialect = sys.argv[2]
    tables = [{
        "catalog": t.catalog,
        "schema": t.db,
        "name": t.name
    } for t in sqlglot.parse_one(sql, dialect=dialect).find_all(Table)]
    print(json.dumps({
        "tables": tables,
        "sql": sql,
        "dialect": dialect,
        "error": None
    }))
except Exception as e:
    print(json.dumps({
        "tables": [],
        "sql": sql,
        "dialect": dialect,
        "error": str(e)
    }))
`, sql, dialect)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		logger.Errorf("ExtractTables failed, %s", err.Error())
		return nil, err
	}

	var result ExtractTablesResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		logger.Errorf("ExtractTables failed, %s", err.Error())
		return nil, err
	}

	if result.Error != "" {
		logger.Errorf("ExtractTables failed, %s", result.Error)
		return nil, errors.New(result.Error)
	}

	return &result, nil
}

// MapDataSourceTypeToDialect 将数据源类型映射到sqlglot方言
func MapDataSourceTypeToDialect(dataSourceType string) (string, error) {
	switch strings.ToLower(dataSourceType) {
	case interfaces.ConnectorTypeMySQL:
		return "mysql", nil
	case "postgres":
		return "postgres", nil
	case "maria", interfaces.ConnectorTypeMariaDB:
		return "mysql", nil // MariaDB使用mysql方言
	default:
		logger.Errorf("unsupported dataSourceType: %s", dataSourceType)
		return "", fmt.Errorf("unsupported dataSourceType: %s", dataSourceType)
	}
}

// TranspileSQL 将SQL从一种方言转换为另一种方言
func TranspileSQL(sql string, fromDialect string, dataSourceType string) (*SQLParseResult, error) {

	// 映射数据源类型到sqlglot方言
	toDialect, err := MapDataSourceTypeToDialect(dataSourceType)
	if err != nil {
		logger.Errorf("MapDataSourceTypeToDialect failed, %s", err.Error())
		return nil, err
	}

	cmd := exec.Command("python3", "-c", `
import sys
import json
import sqlglot

try:
    sql = sys.argv[1]
    from_dialect = sys.argv[2]
    to_dialect = sys.argv[3]
    transpiled = sqlglot.transpile(sql, read=from_dialect, write=to_dialect)[0]
    print(json.dumps({
        "ast": None,
        "sql": transpiled,
        "dialect": to_dialect,
        "error": None
    }))
except Exception as e:
    print(json.dumps({
        "ast": None,
        "sql": sql,
        "dialect": from_dialect,
        "error": str(e)
    }))
`, sql, fromDialect, toDialect)

	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		logger.Errorf("TranspileSQL failed, %s", err.Error())
		return nil, err
	}

	var result SQLParseResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		logger.Errorf("TranspileSQL failed, %s", err.Error())
		return nil, err
	}

	if result.Error != "" {
		logger.Errorf("TranspileSQL failed, %s", result.Error)
		return nil, errors.New(result.Error)
	}

	return &result, nil
}
