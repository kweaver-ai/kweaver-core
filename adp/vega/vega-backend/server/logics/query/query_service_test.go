// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package query

import (
	"context"
	"strings"
	"testing"

	"vega-backend/interfaces"
)

// mockQuerySessionStore 用于测试的 session store
type mockQuerySessionStore struct {
	cursors map[string]map[int]string
}

func (m *mockQuerySessionStore) GetCursor(ctx context.Context, queryID string, offset int) (string, bool) {
	if m.cursors == nil {
		return "", false
	}
	if cursors, ok := m.cursors[queryID]; ok {
		if c, ok := cursors[offset]; ok {
			return c, true
		}
	}
	return "", false
}

func (m *mockQuerySessionStore) SetCursor(ctx context.Context, queryID string, offset int, cursorEncoded string) error {
	if m.cursors == nil {
		m.cursors = make(map[string]map[int]string)
	}
	if m.cursors[queryID] == nil {
		m.cursors[queryID] = make(map[int]string)
	}
	m.cursors[queryID][offset] = cursorEncoded
	return nil
}

func (m *mockQuerySessionStore) Touch(ctx context.Context, queryID string) error {
	return nil
}

func TestValidateRequest_QueryIDRequired(t *testing.T) {
	qs := &queryService{}
	req := &interfaces.QueryExecuteRequest{
		QueryID: "",
		Tables:  []interfaces.TableInQuery{{ResourceID: "r1"}},
		Offset:  100,
		Limit:   100,
	}
	err := qs.validateRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing query_id")
	}
	if !strings.Contains(err.Error(), "QueryIDRequired") && !strings.Contains(err.Error(), "query_id") {
		t.Errorf("expected query_id related error, got %v", err)
	}
}

func TestValidateRequest_QueryIDOptionalForFirstPage(t *testing.T) {
	qs := &queryService{}
	req := &interfaces.QueryExecuteRequest{
		QueryID: "",
		Tables:  []interfaces.TableInQuery{{ResourceID: "r1"}},
		Offset:  0,
		Limit:   100,
	}
	err := qs.validateRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error for missing query_id on first page, got %v", err)
	}
}

func TestValidateRequest_EmptyTables(t *testing.T) {
	qs := &queryService{}
	req := &interfaces.QueryExecuteRequest{
		QueryID: "q1",
		Tables:  nil,
	}
	err := qs.validateRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty tables")
	}
	if !strings.Contains(err.Error(), "tables") {
		t.Errorf("expected tables related error, got %v", err)
	}
}

func TestValidateRequest_LimitExceeded(t *testing.T) {
	qs := &queryService{}
	req := &interfaces.QueryExecuteRequest{
		QueryID: "q1",
		Tables:  []interfaces.TableInQuery{{ResourceID: "r1"}},
		Limit:   10001,
	}
	err := qs.validateRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for limit exceeded")
	}
	if !strings.Contains(err.Error(), "limit") && !strings.Contains(err.Error(), "10000") {
		t.Errorf("expected limit related error, got %v", err)
	}
}

func TestValidateRequest_DefaultLimitAndOffset(t *testing.T) {
	qs := &queryService{}
	req := &interfaces.QueryExecuteRequest{
		QueryID: "q1",
		Tables:  []interfaces.TableInQuery{{ResourceID: "r1"}},
		Limit:   0,
		Offset:  -1,
	}
	err := qs.validateRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.Limit != 100 {
		t.Errorf("expected default limit 100, got %d", req.Limit)
	}
	if req.Offset != 0 {
		t.Errorf("expected normalized offset 0, got %d", req.Offset)
	}
}

func TestValidateRequest_ValidRequest(t *testing.T) {
	qs := &queryService{}
	req := &interfaces.QueryExecuteRequest{
		QueryID: "q1",
		Tables:  []interfaces.TableInQuery{{ResourceID: "r1", Alias: "t1"}},
		Limit:   50,
		Offset:  0,
	}
	err := qs.validateRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestCursorPrevOffsetLogic 验证游标"上一页 offset"计算逻辑
func TestCursorPrevOffsetLogic(t *testing.T) {
	tests := []struct {
		offset     int
		limit      int
		prevOffset int
	}{
		{0, 100, -1},    // 首页，无上一页
		{100, 100, 0},   // 第二页，上一页 offset=0
		{200, 100, 100}, // 第三页，上一页 offset=100
		{100, 50, 50},   // offset=100 limit=50，上一页 offset=50
	}
	for _, tt := range tests {
		prev := tt.offset - tt.limit
		if prev < 0 {
			prev = -1
		}
		if prev != tt.prevOffset {
			t.Errorf("offset=%d limit=%d: expected prevOffset=%d, got %d",
				tt.offset, tt.limit, tt.prevOffset, prev)
		}
	}
}

func TestMockSessionStore(t *testing.T) {
	ctx := context.Background()
	ss := &mockQuerySessionStore{}

	// SetCursor
	_ = ss.SetCursor(ctx, "q1", 0, "cursor0")
	_ = ss.SetCursor(ctx, "q1", 100, "cursor100")

	// GetCursor
	c, ok := ss.GetCursor(ctx, "q1", 0)
	if !ok || c != "cursor0" {
		t.Errorf("expected cursor0, got %q ok=%v", c, ok)
	}
	c, ok = ss.GetCursor(ctx, "q1", 100)
	if !ok || c != "cursor100" {
		t.Errorf("expected cursor100, got %q ok=%v", c, ok)
	}
	_, ok = ss.GetCursor(ctx, "q1", 200)
	if ok {
		t.Error("expected no cursor for 200")
	}
}
