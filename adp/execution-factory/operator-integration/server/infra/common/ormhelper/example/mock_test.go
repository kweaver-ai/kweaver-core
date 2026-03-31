package example_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
)

// MockExecutor 模拟数据库执行器
type MockExecutor struct {
	// 存储预期的查询和响应
	expectedQueries []ExpectedQuery
	currentIndex    int
}

type ExpectedQuery struct {
	SQL       string
	Args      []interface{}
	Result    sql.Result
	Rows      *sql.Rows
	Error     error
	QueryType string // "exec", "query", "queryrow"
}

// MockResult 模拟SQL结果
type MockResult struct {
	lastInsertID int64
	rowsAffected int64
}

func (m *MockResult) LastInsertId() (int64, error) {
	return m.lastInsertID, nil
}

func (m *MockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, nil
}

// NewMockExecutor 创建Mock执行器
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		expectedQueries: make([]ExpectedQuery, 0),
		currentIndex:    0,
	}
}

// ExpectExec 预期一个Exec调用
func (m *MockExecutor) ExpectExec(sql string, args ...interface{}) *ExpectedQuery {
	query := ExpectedQuery{
		SQL:       sql,
		Args:      args,
		QueryType: "exec",
	}
	m.expectedQueries = append(m.expectedQueries, query)
	return &m.expectedQueries[len(m.expectedQueries)-1]
}

// ExpectQuery 预期一个Query调用
func (m *MockExecutor) ExpectQuery(sql string, args ...interface{}) *ExpectedQuery {
	query := ExpectedQuery{
		SQL:       sql,
		Args:      args,
		QueryType: "query",
	}
	m.expectedQueries = append(m.expectedQueries, query)
	return &m.expectedQueries[len(m.expectedQueries)-1]
}

// ExpectQueryRow 预期一个QueryRow调用
func (m *MockExecutor) ExpectQueryRow(sql string, args ...interface{}) *ExpectedQuery {
	query := ExpectedQuery{
		SQL:       sql,
		Args:      args,
		QueryType: "queryrow",
	}
	m.expectedQueries = append(m.expectedQueries, query)
	return &m.expectedQueries[len(m.expectedQueries)-1]
}

// WillReturnResult 设置返回的结果
func (eq *ExpectedQuery) WillReturnResult(lastInsertID, rowsAffected int64) *ExpectedQuery {
	eq.Result = &MockResult{
		lastInsertID: lastInsertID,
		rowsAffected: rowsAffected,
	}
	return eq
}

// WillReturnError 设置返回的错误
func (eq *ExpectedQuery) WillReturnError(err error) *ExpectedQuery {
	eq.Error = err
	return eq
}

// ExecContext 模拟执行SQL
func (m *MockExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if m.currentIndex >= len(m.expectedQueries) {
		return nil, fmt.Errorf("unexpected exec call: %s", query)
	}

	expected := m.expectedQueries[m.currentIndex]
	m.currentIndex++

	if expected.QueryType != "exec" {
		return nil, fmt.Errorf("expected %s but got exec", expected.QueryType)
	}

	if expected.Error != nil {
		return nil, expected.Error
	}

	return expected.Result, nil
}

// QueryContext 模拟查询多行
func (m *MockExecutor) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if m.currentIndex >= len(m.expectedQueries) {
		return nil, fmt.Errorf("unexpected query call: %s", query)
	}

	expected := m.expectedQueries[m.currentIndex]
	m.currentIndex++

	if expected.QueryType != "query" {
		return nil, fmt.Errorf("expected %s but got query", expected.QueryType)
	}

	if expected.Error != nil {
		return nil, expected.Error
	}

	return expected.Rows, nil
}

// QueryRowContext 模拟查询单行
func (m *MockExecutor) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if m.currentIndex >= len(m.expectedQueries) {
		// 返回一个包含错误的Row
		return &sql.Row{}
	}

	expected := m.expectedQueries[m.currentIndex]
	m.currentIndex++

	if expected.QueryType != "queryrow" {
		// 返回一个包含错误的Row
		return &sql.Row{}
	}

	// 这里简化处理，实际应该返回包含预期数据的Row
	return &sql.Row{}
}

// Reset 重置Mock状态
func (m *MockExecutor) Reset() {
	m.expectedQueries = make([]ExpectedQuery, 0)
	m.currentIndex = 0
}

// ExpectationsWereMet 检查所有预期是否都被满足
func (m *MockExecutor) ExpectationsWereMet() error {
	if m.currentIndex < len(m.expectedQueries) {
		return fmt.Errorf("expected %d queries, but only %d were executed", len(m.expectedQueries), m.currentIndex)
	}
	return nil
}

// MockRows 简单的模拟Rows实现
type MockRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func NewMockRows(columns []string, values [][]driver.Value) *MockRows {
	return &MockRows{
		columns: columns,
		values:  values,
		index:   -1,
	}
}
