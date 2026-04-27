// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/mock/gomock"

	"vega-backend/interfaces"
	mock_interfaces "vega-backend/interfaces/mock"
)

// newTestService 使用 mockgen 生成的 mock 构建 resourceService
func newTestService(t *testing.T) (*resourceService, *mock_interfaces.MockResourceAccess, *mock_interfaces.MockPermissionService,
	*mock_interfaces.MockDatasetService, *mock_interfaces.MockUserMgmtService, *mock_interfaces.MockCatalogService) {
	ctrl := gomock.NewController(t)
	mockRA := mock_interfaces.NewMockResourceAccess(ctrl)
	mockPS := mock_interfaces.NewMockPermissionService(ctrl)
	mockDS := mock_interfaces.NewMockDatasetService(ctrl)
	mockUMS := mock_interfaces.NewMockUserMgmtService(ctrl)
	mockCS := mock_interfaces.NewMockCatalogService(ctrl)

	rs := &resourceService{
		ra:  mockRA,
		ps:  mockPS,
		ds:  mockDS,
		ums: mockUMS,
		cs:  mockCS,
	}
	return rs, mockRA, mockPS, mockDS, mockUMS, mockCS
}

// ===== CheckExistByID =====

func TestCheckExistByID_Found(t *testing.T) {
	rs, mockRA, _, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "r1").
		Return(&interfaces.Resource{ID: "r1"}, nil)

	exists, err := rs.CheckExistByID(context.Background(), "r1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected resource to exist")
	}
}

func TestCheckExistByID_NotFound(t *testing.T) {
	rs, mockRA, _, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "missing").
		Return(nil, nil)

	exists, err := rs.CheckExistByID(context.Background(), "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected resource to not exist")
	}
}

func TestCheckExistByID_Error(t *testing.T) {
	rs, mockRA, _, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "r1").
		Return(nil, fmt.Errorf("db error"))

	_, err := rs.CheckExistByID(context.Background(), "r1")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== CheckExistByName =====

func TestCheckExistByName_Found(t *testing.T) {
	rs, mockRA, _, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByName(gomock.Any(), "cat1", "test").
		Return(&interfaces.Resource{Name: "test"}, nil)

	exists, err := rs.CheckExistByName(context.Background(), "cat1", "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected resource to exist")
	}
}

func TestCheckExistByName_NotFound(t *testing.T) {
	rs, mockRA, _, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByName(gomock.Any(), "cat1", "missing").
		Return(nil, nil)

	exists, err := rs.CheckExistByName(context.Background(), "cat1", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected resource to not exist")
	}
}

// ===== GetByID =====

func TestGetByID_Success(t *testing.T) {
	rs, mockRA, mockPS, _, mockUMS, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "r1").
		Return(&interfaces.Resource{ID: "r1", Name: "test"}, nil)
	mockPS.EXPECT().FilterResources(gomock.Any(), interfaces.RESOURCE_TYPE_RESOURCE,
		[]string{"r1"}, gomock.Any(), true, gomock.Any()).
		Return(map[string]interfaces.PermissionResourceOps{
			"r1": {ResourceID: "r1", Operations: []string{"view_detail"}},
		}, nil)
	mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

	resource, err := rs.GetByID(context.Background(), "r1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resource.ID != "r1" {
		t.Errorf("expected ID 'r1', got '%s'", resource.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	rs, mockRA, _, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "missing").
		Return(nil, nil)

	_, err := rs.GetByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for not found resource")
	}
}

func TestGetByID_DBError(t *testing.T) {
	rs, mockRA, _, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByID(gomock.Any(), "r1").
		Return(nil, fmt.Errorf("db error"))

	_, err := rs.GetByID(context.Background(), "r1")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== GetByIDs =====

func TestGetByIDs_Success(t *testing.T) {
	rs, mockRA, mockPS, _, mockUMS, _ := newTestService(t)
	mockRA.EXPECT().GetByIDs(gomock.Any(), []string{"r1", "r2"}).
		Return([]*interfaces.Resource{{ID: "r1"}, {ID: "r2"}}, nil)
	mockPS.EXPECT().FilterResources(gomock.Any(), interfaces.RESOURCE_TYPE_RESOURCE,
		[]string{"r1", "r2"}, gomock.Any(), true, gomock.Any()).
		Return(map[string]interfaces.PermissionResourceOps{
			"r1": {ResourceID: "r1", Operations: []string{"view_detail"}},
			"r2": {ResourceID: "r2", Operations: []string{"view_detail"}},
		}, nil)
	mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

	resources, err := rs.GetByIDs(context.Background(), []string{"r1", "r2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
}

// ===== GetByCatalogID =====

func TestGetByCatalogID_Success(t *testing.T) {
	rs, mockRA, _, _, _, _ := newTestService(t)
	mockRA.EXPECT().GetByCatalogID(gomock.Any(), "cat1").
		Return([]*interfaces.Resource{{ID: "r1", CatalogID: "cat1"}}, nil)

	resources, err := rs.GetByCatalogID(context.Background(), "cat1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(resources))
	}
}

// ===== List 分页逻辑 =====

func TestList_Pagination(t *testing.T) {
	rs, mockRA, mockPS, _, mockUMS, _ := newTestService(t)
	ids := []string{"c1", "c2", "c3", "c4"}
	mockRA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
	mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
		Return(map[string]interfaces.PermissionResourceOps{
			"c1": {ResourceID: "c1"}, "c2": {ResourceID: "c2"}, "c3": {ResourceID: "c3"}, "c4": {ResourceID: "c4"},
		}, nil)
	resources := []*interfaces.Resource{{ID: "r2"}, {ID: "r3"}}
	mockRA.EXPECT().GetByIDsBasic(gomock.Any(), gomock.Any()).Return(resources, nil)
	mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

	result, total, err := rs.List(context.Background(), interfaces.ResourcesQueryParams{
		PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 1, Limit: 2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 4 {
		t.Errorf("expected total 4, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
	if result[0].ID != "r2" {
		t.Errorf("expected first item 'r2', got '%s'", result[0].ID)
	}
}

func TestList_ReturnAll(t *testing.T) {
	rs, mockRA, mockPS, _, mockUMS, _ := newTestService(t)
	ids := []string{"c1", "c2"}
	resources := []*interfaces.Resource{{ID: "r1"}, {ID: "r2"}}
	mockRA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
	mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
		Return(map[string]interfaces.PermissionResourceOps{
			"c1": {ResourceID: "c1"}, "c2": {ResourceID: "c2"},
		}, nil)
	mockRA.EXPECT().GetByIDsBasic(gomock.Any(), gomock.Any()).Return(resources, nil)
	mockUMS.EXPECT().GetAccountNames(gomock.Any(), gomock.Any()).Return(nil)

	result, total, err := rs.List(context.Background(), interfaces.ResourcesQueryParams{
		PaginationQueryParams: interfaces.PaginationQueryParams{Limit: -1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

func TestList_OffsetBeyondTotal(t *testing.T) {
	rs, mockRA, mockPS, _, _, _ := newTestService(t)

	ids := []string{"c1"}
	mockRA.EXPECT().ListIDs(gomock.Any(), gomock.Any()).Return(ids, nil)
	mockPS.EXPECT().FilterResources(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), true, gomock.Any()).
		Return(map[string]interfaces.PermissionResourceOps{"c1": {ResourceID: "c1"}}, nil)

	result, total, err := rs.List(context.Background(), interfaces.ResourcesQueryParams{
		PaginationQueryParams: interfaces.PaginationQueryParams{Offset: 10, Limit: 5},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

func TestCreate_DatasetCategory(t *testing.T) {
	rs, mockRA, mockPS, mockDS, _, mockCS := newTestService(t)
	mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockCS.EXPECT().CheckExistByID(gomock.Any(), gomock.Any()).Return(true, nil)
	mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	mockDS.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	resource, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
		Name:     "test-dataset",
		Category: interfaces.ResourceCategoryDataset,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resource == nil {
		t.Error("expected non-empty ID")
	}
}

// ===== DeleteByIDs =====

func TestDeleteByIDs_Empty(t *testing.T) {
	rs, _, _, _, _, _ := newTestService(t)
	err := rs.DeleteByIDs(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteByIDs_Success(t *testing.T) {
	rs, mockRA, mockPS, _, _, _ := newTestService(t)
	mockPS.EXPECT().FilterResources(gomock.Any(), interfaces.RESOURCE_TYPE_RESOURCE,
		[]string{"r1"}, gomock.Any(), true, gomock.Any()).
		Return(map[string]interfaces.PermissionResourceOps{
			"r1": {ResourceID: "r1"},
		}, nil)
	mockRA.EXPECT().GetByIDs(gomock.Any(), []string{"r1"}).
		Return([]*interfaces.Resource{{ID: "r1", Category: "table"}}, nil)
	mockRA.EXPECT().DeleteByIDs(gomock.Any(), []string{"r1"}).Return(nil)
	mockPS.EXPECT().DeleteResources(gomock.Any(), interfaces.RESOURCE_TYPE_RESOURCE, []string{"r1"}).Return(nil)

	err := rs.DeleteByIDs(context.Background(), []string{"r1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ===== Create =====

func TestCreate_Success(t *testing.T) {
	rs, mockRA, mockPS, _, _, mockCS := newTestService(t)
	mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockCS.EXPECT().CheckExistByID(gomock.Any(), gomock.Any()).Return(true, nil)
	mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	resource, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
		Name:     "test-resource",
		Category: "table",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resource == nil {
		t.Error("expected non-empty ID")
	}
}

func TestCreate_WithExplicitID(t *testing.T) {
	rs, mockRA, mockPS, _, _, mockCS := newTestService(t)
	mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockCS.EXPECT().CheckExistByID(gomock.Any(), gomock.Any()).Return(true, nil)
	mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
	mockPS.EXPECT().CreateResources(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	resource, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
		ID:       "custom-id",
		Name:     "test-resource",
		Category: "table",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resource == nil || resource.ID != "custom-id" {
		t.Errorf("expected 'custom-id', got '%s'", resource.ID)
	}
}

func TestCreate_DBError(t *testing.T) {
	rs, mockRA, mockPS, _, _, mockCS := newTestService(t)
	mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockCS.EXPECT().CheckExistByID(gomock.Any(), gomock.Any()).Return(true, nil)
	mockRA.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("db error"))

	_, err := rs.Create(context.Background(), &interfaces.ResourceRequest{
		Name: "test-resource",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== UpdateStatus =====

func TestUpdateStatus_Success(t *testing.T) {
	rs, mockRA, _, _, _, _ := newTestService(t)
	mockRA.EXPECT().UpdateStatus(gomock.Any(), "r1", "active", "").Return(nil)

	err := rs.UpdateStatus(context.Background(), "r1", "active", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateStatus_Error(t *testing.T) {
	rs, mockRA, _, _, _, _ := newTestService(t)
	mockRA.EXPECT().UpdateStatus(gomock.Any(), "r1", "active", "").
		Return(fmt.Errorf("db error"))

	err := rs.UpdateStatus(context.Background(), "r1", "active", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// ===== Update =====

func TestUpdate_NilOrigin(t *testing.T) {
	rs, _, _, _, _, _ := newTestService(t)
	err := rs.Update(context.Background(), "r1", &interfaces.ResourceRequest{
		OriginResource: nil,
	})
	if err == nil {
		t.Fatal("expected error for nil origin resource")
	}
}

func TestUpdate_Success(t *testing.T) {
	rs, mockRA, mockPS, _, _, mockCS := newTestService(t)
	mockPS.EXPECT().CheckPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockCS.EXPECT().CheckExistByID(gomock.Any(), gomock.Any()).Return(true, nil)
	mockRA.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	err := rs.Update(context.Background(), "r1", &interfaces.ResourceRequest{
		OriginResource: &interfaces.Resource{ID: "r1"},
		Name:           "updated",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
