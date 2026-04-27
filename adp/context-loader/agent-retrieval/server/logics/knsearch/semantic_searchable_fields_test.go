// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knsearch

import (
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func TestFindSemanticSearchableFields_Empty(t *testing.T) {
	// 无 DataProperties 或 非文本/无可搜操作符 时返回空
	objType := &interfaces.KnSearchObjectType{ConceptID: "ot1"}
	out := findSemanticSearchableFields(objType)
	if len(out) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(out))
	}
}

func TestFindSemanticSearchableFields_TextWithOps(t *testing.T) {
	// 文本类型且 condition_operations 含 knn/match/== 时入选
	objType := &interfaces.KnSearchObjectType{
		ConceptID: "ot1",
		DataProperties: []*interfaces.KnSearchDataProperty{
			{Name: "title", Type: "text", ConditionOperations: []interfaces.KnOperationType{interfaces.KnOperationTypeKnn, interfaces.KnOperationTypeMatch}},
			{Name: "name", Type: "string", ConditionOperations: []interfaces.KnOperationType{interfaces.KnOperationTypeEqual}},
			{Name: "count", Type: "int", ConditionOperations: []interfaces.KnOperationType{interfaces.KnOperationTypeKnn}}, // 非文本，不选
		},
	}
	out := findSemanticSearchableFields(objType)
	if len(out) != 2 {
		t.Fatalf("Expected 2 fields (title, name), got %d", len(out))
	}
	if out[0].Name != "title" || !out[0].HasKnn || !out[0].HasMatch || out[0].HasExactMatch {
		t.Errorf("Expected title with knn+match, got %+v", out[0])
	}
	if out[1].Name != "name" || !out[1].HasExactMatch || out[1].HasKnn || out[1].HasMatch {
		t.Errorf("Expected name with == only, got %+v", out[1])
	}
}

func TestBuildSemanticSearchConditionStruct_WithSearchableFields(t *testing.T) {
	svc := &localSearchImpl{}
	config := &interfaces.KnSearchSemanticInstanceRetrievalConfig{
		PerTypeInstanceLimit:     5,
		MaxSemanticSubConditions: 10,
	}
	objType := &interfaces.KnSearchObjectType{
		ConceptID: "ot1",
		DataProperties: []*interfaces.KnSearchDataProperty{
			{Name: "title", Type: "text", ConditionOperations: []interfaces.KnOperationType{interfaces.KnOperationTypeKnn, interfaces.KnOperationTypeMatch}},
		},
	}
	cond := svc.buildSemanticSearchConditionStruct("query", objType, config)
	if cond.Operation != interfaces.KnOperationTypeOr {
		t.Errorf("Expected OR operation, got %s", cond.Operation)
	}
	// 有 1 个可搜字段 title：knn + match => 至少 2 条
	if len(cond.SubConditions) < 2 {
		t.Errorf("Expected at least 2 subconditions (knn+match for title), got %d", len(cond.SubConditions))
	}
	// knn 的 limit_value 应为 PerTypeInstanceLimit（5）
	first := cond.SubConditions[0]
	if first.Operation == interfaces.KnOperationTypeKnn && first.Field == "title" {
		if first.LimitValue != 5 {
			t.Errorf("Expected knn limit_value=5 (PerTypeInstanceLimit), got %v", first.LimitValue)
		}
	}
}
