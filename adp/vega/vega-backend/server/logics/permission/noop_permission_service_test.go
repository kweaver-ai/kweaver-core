// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package permission

import (
	"context"
	"testing"

	"vega-backend/interfaces"
)

func TestNoopPermission_CheckPermission(t *testing.T) {
	svc := NewNoopPermissionService(nil)
	err := svc.CheckPermission(context.Background(), interfaces.PermissionResource{
		Type: "catalog",
		ID:   "all",
	}, []string{"create"})
	if err != nil {
		t.Fatalf("noop should always return nil, got: %v", err)
	}
}

func TestNoopPermission_CreateResources(t *testing.T) {
	svc := NewNoopPermissionService(nil)
	err := svc.CreateResources(context.Background(), []interfaces.PermissionResource{
		{ID: "r1", Type: "catalog", Name: "test"},
	}, []string{"view", "modify"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNoopPermission_DeleteResources(t *testing.T) {
	svc := NewNoopPermissionService(nil)
	err := svc.DeleteResources(context.Background(), "catalog", []string{"r1", "r2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNoopPermission_FilterResources(t *testing.T) {
	svc := NewNoopPermissionService(nil)
	ids := []string{"r1", "r2", "r3"}
	ops := []string{"view_detail", "modify"}

	result, err := svc.FilterResources(context.Background(), "catalog", ids, ops, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 results, got %d", len(result))
	}
	for _, id := range ids {
		r, ok := result[id]
		if !ok {
			t.Errorf("expected result for id '%s'", id)
			continue
		}
		if r.ResourceID != id {
			t.Errorf("expected ResourceID '%s', got '%s'", id, r.ResourceID)
		}
		if len(r.Operations) != 2 {
			t.Errorf("expected 2 operations, got %d", len(r.Operations))
		}
	}
}

func TestNoopPermission_UpdateResource(t *testing.T) {
	svc := NewNoopPermissionService(nil)
	err := svc.UpdateResource(context.Background(), interfaces.PermissionResource{
		ID:   "r1",
		Type: "catalog",
		Name: "updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
