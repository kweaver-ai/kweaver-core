// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package filter_condition

import (
	"context"
	"strings"
	"testing"

	"vega-backend/interfaces"
)

// 测试用字段映射
func testFieldsMap() map[string]*interfaces.Property {
	return map[string]*interfaces.Property{
		"name":       {Name: "name", Type: interfaces.DataType_String},
		"age":        {Name: "age", Type: interfaces.DataType_Integer},
		"score":      {Name: "score", Type: interfaces.DataType_Float},
		"created_at": {Name: "created_at", Type: interfaces.DataType_Datetime},
		"is_active":  {Name: "is_active", Type: interfaces.DataType_Boolean},
		"tags":       {Name: "tags", Type: interfaces.DataType_Text},
		"other_name": {Name: "other_name", Type: interfaces.DataType_String},
	}
}

func constCfg(name, op string, value any) *interfaces.FilterCondCfg {
	return &interfaces.FilterCondCfg{
		Name:      name,
		Operation: op,
		ValueOptCfg: interfaces.ValueOptCfg{
			ValueFrom: interfaces.ValueFrom_Const,
			Value:     value,
		},
	}
}

// ===== NewFilterCondition 工厂函数 =====

func TestNewFilterCondition_NilConfig(t *testing.T) {
	cond, err := NewFilterCondition(context.Background(), nil, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cond != nil {
		t.Fatal("expected nil condition for nil config")
	}
}

func TestNewFilterCondition_EmptyConfig(t *testing.T) {
	cfg := &interfaces.FilterCondCfg{}
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cond != nil {
		t.Fatal("expected nil condition for empty config")
	}
}

func TestNewFilterCondition_UnsupportedOperation(t *testing.T) {
	cfg := constCfg("name", "unknown_op", "test")
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for unsupported operation")
	}
	if !strings.Contains(err.Error(), "unsupported operation") {
		t.Errorf("expected 'unsupported operation' error, got: %v", err)
	}
}

func TestNewFilterCondition_ValidEqual(t *testing.T) {
	cfg := constCfg("name", "==", "alice")
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cond == nil {
		t.Fatal("expected non-nil condition")
	}
	if cond.GetOperation() != OperationEqual {
		t.Errorf("expected operation '==', got '%s'", cond.GetOperation())
	}
}

// ===== IsSlice / IsSameType 工具函数 =====

func TestIsSlice(t *testing.T) {
	if !IsSlice([]int{1, 2, 3}) {
		t.Error("expected true for int slice")
	}
	if !IsSlice([]string{"a"}) {
		t.Error("expected true for string slice")
	}
	if IsSlice("not a slice") {
		t.Error("expected false for string")
	}
	if IsSlice(42) {
		t.Error("expected false for int")
	}
}

func TestIsSameType(t *testing.T) {
	if !IsSameType([]any{}) {
		t.Error("expected true for empty slice")
	}
	if !IsSameType([]any{1, 2, 3}) {
		t.Error("expected true for same-type slice")
	}
	if !IsSameType([]any{"a", "b"}) {
		t.Error("expected true for string slice")
	}
	if IsSameType([]any{1, "two", 3}) {
		t.Error("expected false for mixed-type slice")
	}
}

// ===== EqualCond =====

func TestEqualCond_Valid(t *testing.T) {
	cfg := constCfg("name", "==", "alice")
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	eq := cond.(*EqualCond)
	if eq.Lfield.Name != "name" {
		t.Errorf("expected left field 'name', got '%s'", eq.Lfield.Name)
	}
	if eq.Value != "alice" {
		t.Errorf("expected value 'alice', got '%v'", eq.Value)
	}
}

func TestEqualCond_EmptyFieldName(t *testing.T) {
	cfg := constCfg("", "==", "alice")
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for empty field name")
	}
	if !strings.Contains(err.Error(), "left field is empty") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEqualCond_FieldNotFound(t *testing.T) {
	cfg := constCfg("nonexistent", "==", "alice")
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEqualCond_RejectsArrayValue(t *testing.T) {
	cfg := constCfg("name", "==", []any{"a", "b"})
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for array value in equal condition")
	}
	if !strings.Contains(err.Error(), "single value") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEqualCond_FieldToField(t *testing.T) {
	cfg := &interfaces.FilterCondCfg{
		Name:      "name",
		Operation: "==",
		ValueOptCfg: interfaces.ValueOptCfg{
			ValueFrom: interfaces.ValueFrom_Field,
			Value:     "other_name",
		},
	}
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	eq := cond.(*EqualCond)
	if eq.Rfield == nil {
		t.Fatal("expected right field to be set")
	}
	if eq.Rfield.Name != "other_name" {
		t.Errorf("expected right field 'other_name', got '%s'", eq.Rfield.Name)
	}
}

func TestEqualCond_FieldToField_RightNotFound(t *testing.T) {
	cfg := &interfaces.FilterCondCfg{
		Name:      "name",
		Operation: "==",
		ValueOptCfg: interfaces.ValueOptCfg{
			ValueFrom: interfaces.ValueFrom_Field,
			Value:     "nonexistent",
		},
	}
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for unknown right field")
	}
	if !strings.Contains(err.Error(), "right field") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEqualCond_AliasOperations(t *testing.T) {
	// "eq" 是 "==" 的别名
	cfg := constCfg("name", "eq", "alice")
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cond.GetOperation() != OperationEqual {
		t.Errorf("expected operation '==', got '%s'", cond.GetOperation())
	}
}

// ===== InCond =====

func TestInCond_Valid(t *testing.T) {
	cfg := constCfg("name", "in", []any{"alice", "bob"})
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	in := cond.(*InCond)
	if len(in.Value) != 2 {
		t.Errorf("expected 2 values, got %d", len(in.Value))
	}
}

func TestInCond_EmptyArray(t *testing.T) {
	cfg := constCfg("name", "in", []any{})
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for empty array")
	}
	if !strings.Contains(err.Error(), "length >= 1") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInCond_NonArrayValue(t *testing.T) {
	cfg := constCfg("name", "in", "single_value")
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for non-array value")
	}
	if !strings.Contains(err.Error(), "should be an array") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInCond_RejectsFieldValueFrom(t *testing.T) {
	cfg := &interfaces.FilterCondCfg{
		Name:      "name",
		Operation: "in",
		ValueOptCfg: interfaces.ValueOptCfg{
			ValueFrom: interfaces.ValueFrom_Field,
			Value:     []any{"alice"},
		},
	}
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for non-const value_from")
	}
	if !strings.Contains(err.Error(), "does not support value_from") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ===== LikeCond =====

func TestLikeCond_Valid(t *testing.T) {
	cfg := constCfg("name", "like", "ali%")
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	like := cond.(*LikeCond)
	if like.Value != "ali%" {
		t.Errorf("expected value 'ali%%', got '%s'", like.Value)
	}
}

func TestLikeCond_NonStringField(t *testing.T) {
	cfg := constCfg("age", "like", "test")
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for non-string field")
	}
	if !strings.Contains(err.Error(), "not a string field") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLikeCond_NonStringValue(t *testing.T) {
	cfg := constCfg("name", "like", 123)
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for non-string value")
	}
	if !strings.Contains(err.Error(), "not a string value") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ===== RangeCond =====

func TestRangeCond_Valid(t *testing.T) {
	cfg := constCfg("age", "range", []any{18, 65})
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := cond.(*RangeCond)
	if len(r.Value) != 2 {
		t.Errorf("expected 2 values, got %d", len(r.Value))
	}
}

func TestRangeCond_WrongArrayLength(t *testing.T) {
	cfg := constCfg("age", "range", []any{18})
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for wrong array length")
	}
	if !strings.Contains(err.Error(), "length 2") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRangeCond_NonNumericField(t *testing.T) {
	cfg := constCfg("name", "range", []any{1, 2})
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for non-numeric field")
	}
	if !strings.Contains(err.Error(), "not a date/number field") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRangeCond_DateField(t *testing.T) {
	cfg := constCfg("created_at", "range", []any{"2024-01-01", "2024-12-31"})
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ===== AndCond =====

func TestAndCond_Valid(t *testing.T) {
	cfg := &interfaces.FilterCondCfg{
		Operation: "and",
		SubConds: []*interfaces.FilterCondCfg{
			constCfg("name", "==", "alice"),
			constCfg("age", ">", 18),
		},
	}
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	and := cond.(*AndCond)
	if len(and.SubConds) != 2 {
		t.Errorf("expected 2 sub-conditions, got %d", len(and.SubConds))
	}
}

func TestAndCond_EmptySubConds(t *testing.T) {
	cfg := &interfaces.FilterCondCfg{
		Operation: "and",
		SubConds:  []*interfaces.FilterCondCfg{},
	}
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for empty sub-conditions")
	}
	if !strings.Contains(err.Error(), "size is 0") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAndCond_InvalidSubCond(t *testing.T) {
	cfg := &interfaces.FilterCondCfg{
		Operation: "and",
		SubConds: []*interfaces.FilterCondCfg{
			constCfg("nonexistent", "==", "test"),
		},
	}
	_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err == nil {
		t.Fatal("expected error for invalid sub-condition")
	}
}

func TestAndCond_NestedAndOr(t *testing.T) {
	cfg := &interfaces.FilterCondCfg{
		Operation: "and",
		SubConds: []*interfaces.FilterCondCfg{
			constCfg("name", "==", "alice"),
			{
				Operation: "or",
				SubConds: []*interfaces.FilterCondCfg{
					constCfg("age", ">", 18),
					constCfg("age", "<", 65),
				},
			},
		},
	}
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	and := cond.(*AndCond)
	if len(and.SubConds) != 2 {
		t.Errorf("expected 2 sub-conditions, got %d", len(and.SubConds))
	}
	if and.SubConds[1].GetOperation() != OperationOr {
		t.Errorf("expected nested 'or', got '%s'", and.SubConds[1].GetOperation())
	}
}

// ===== 所有操作符注册验证 =====

func TestAllOperationsRegistered(t *testing.T) {
	expectedOps := []string{
		"and", "or",
		"==", "eq", "!=", "not_eq",
		">", "gt", ">=", "gte", "<", "lt", "<=", "lte",
		"in", "not_in",
		"like", "not_like",
		"contain", "not_contain",
		"range", "out_range",
		"exist", "not_exist",
		"empty", "not_empty",
		"regex", "match", "match_phrase",
		"prefix", "not_prefix",
		"null", "not_null",
		"true", "false",
		"before", "current", "between",
		"knn_vector", "multi_match",
	}

	for _, op := range expectedOps {
		if _, exists := OperationMap[op]; !exists {
			t.Errorf("operation '%s' not registered in OperationMap", op)
		}
	}
}

// ===== 比较运算符共通测试 =====

func TestComparisonOps_EmptyField(t *testing.T) {
	ops := []string{"==", "!=", ">", ">=", "<", "<="}
	for _, op := range ops {
		cfg := constCfg("", op, "test")
		_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
		if err == nil {
			t.Errorf("operation '%s': expected error for empty field name", op)
		}
	}
}

func TestComparisonOps_FieldNotFound(t *testing.T) {
	ops := []string{"==", "!=", ">", ">=", "<", "<="}
	for _, op := range ops {
		cfg := constCfg("nonexistent", op, "test")
		_, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
		if err == nil {
			t.Errorf("operation '%s': expected error for unknown field", op)
		}
	}
}

// ===== Null/NotNull/Exist/NotExist (无需 value) =====

func TestNullCond_Valid(t *testing.T) {
	cfg := &interfaces.FilterCondCfg{
		Name:      "name",
		Operation: "null",
	}
	cond, err := NewFilterCondition(context.Background(), cfg, testFieldsMap())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cond.GetOperation() != OperationNull {
		t.Errorf("expected operation 'null', got '%s'", cond.GetOperation())
	}
	if cond.NeedValue() {
		t.Error("null condition should not need value")
	}
}
