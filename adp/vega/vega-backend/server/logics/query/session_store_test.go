// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package query

import (
	"context"
	"testing"
	"time"
)

func TestMemorySessionStore_GetSetCursor(t *testing.T) {
	ctx := context.Background()
	store := NewMemorySessionStore(30 * time.Minute).(*memorySessionStore)

	queryID := "test-query-1"

	// 初始无游标
	cursor, ok := store.GetCursor(ctx, queryID, 0)
	if ok || cursor != "" {
		t.Errorf("expected no cursor, got ok=%v cursor=%q", ok, cursor)
	}

	// 设置游标
	err := store.SetCursor(ctx, queryID, 0, "cursor-encoded-0")
	if err != nil {
		t.Fatalf("SetCursor failed: %v", err)
	}

	// 获取游标
	cursor, ok = store.GetCursor(ctx, queryID, 0)
	if !ok || cursor != "cursor-encoded-0" {
		t.Errorf("expected cursor-encoded-0, got ok=%v cursor=%q", ok, cursor)
	}

	// 不同 offset 的游标
	_ = store.SetCursor(ctx, queryID, 100, "cursor-encoded-100")
	cursor, ok = store.GetCursor(ctx, queryID, 100)
	if !ok || cursor != "cursor-encoded-100" {
		t.Errorf("expected cursor-encoded-100, got ok=%v cursor=%q", ok, cursor)
	}

	// 不存在的 offset
	_, ok = store.GetCursor(ctx, queryID, 200)
	if ok {
		t.Error("expected no cursor for offset 200")
	}
}

func TestMemorySessionStore_Touch(t *testing.T) {
	ctx := context.Background()
	store := NewMemorySessionStore(100 * time.Millisecond).(*memorySessionStore)

	queryID := "test-query-touch"
	_ = store.SetCursor(ctx, queryID, 0, "cursor")

	// Touch 不应报错
	err := store.Touch(ctx, queryID)
	if err != nil {
		t.Errorf("Touch failed: %v", err)
	}

	// 不存在的 queryID Touch 也不应报错
	err = store.Touch(ctx, "non-existent")
	if err != nil {
		t.Errorf("Touch non-existent failed: %v", err)
	}
}

func TestMemorySessionStore_MultipleQueryIDs(t *testing.T) {
	ctx := context.Background()
	store := NewMemorySessionStore(30 * time.Minute).(*memorySessionStore)

	_ = store.SetCursor(ctx, "q1", 0, "c1")
	_ = store.SetCursor(ctx, "q2", 0, "c2")

	c1, ok1 := store.GetCursor(ctx, "q1", 0)
	c2, ok2 := store.GetCursor(ctx, "q2", 0)

	if !ok1 || c1 != "c1" {
		t.Errorf("q1: expected c1, got %q ok=%v", c1, ok1)
	}
	if !ok2 || c2 != "c2" {
		t.Errorf("q2: expected c2, got %q ok=%v", c2, ok2)
	}
}
