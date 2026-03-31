package umtypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParentDepInfo_StructFields(t *testing.T) {
	t.Parallel()

	info := ParentDepInfo{
		ID:   "dept-123",
		Name: "Parent Department",
		Type: "department",
	}

	assert.Equal(t, "dept-123", info.ID)
	assert.Equal(t, "Parent Department", info.Name)
	assert.Equal(t, "department", info.Type)
}

func TestParentDepInfo_Empty(t *testing.T) {
	t.Parallel()

	info := ParentDepInfo{}

	assert.Empty(t, info.ID)
	assert.Empty(t, info.Name)
	assert.Empty(t, info.Type)
}

func TestParentDepInfo_WithChineseName(t *testing.T) {
	t.Parallel()

	info := ParentDepInfo{
		ID:   "部门-123",
		Name: "父部门",
		Type: "department",
	}

	assert.Equal(t, "部门-123", info.ID)
	assert.Equal(t, "父部门", info.Name)
	assert.Equal(t, "department", info.Type)
}

func TestParentDepInfo_WithOnlyID(t *testing.T) {
	t.Parallel()

	info := ParentDepInfo{
		ID: "dept-123",
	}

	assert.Equal(t, "dept-123", info.ID)
	assert.Empty(t, info.Name)
	assert.Empty(t, info.Type)
}

func TestParentDepInfo_WithOnlyName(t *testing.T) {
	t.Parallel()

	info := ParentDepInfo{
		Name: "Test Department",
	}

	assert.Empty(t, info.ID)
	assert.Equal(t, "Test Department", info.Name)
	assert.Empty(t, info.Type)
}

func TestParentDepInfo_TypeIsAlwaysDepartment(t *testing.T) {
	t.Parallel()

	info := ParentDepInfo{
		ID:   "dept-123",
		Name: "Department",
		Type: "department",
	}

	assert.Equal(t, "department", info.Type)
}

func TestDepartmentInfo_StructFields(t *testing.T) {
	t.Parallel()

	parentDeps := []*ParentDepInfo{
		{ID: "parent-1", Name: "Parent 1", Type: "department"},
		{ID: "parent-2", Name: "Parent 2", Type: "department"},
	}

	info := DepartmentInfo{
		DepartmentId: "dept-123",
		Name:         "Test Department",
		ParentDeps:   parentDeps,
	}

	assert.Equal(t, "dept-123", info.DepartmentId)
	assert.Equal(t, "Test Department", info.Name)
	assert.Len(t, info.ParentDeps, 2)
	assert.Equal(t, "parent-1", info.ParentDeps[0].ID)
}

func TestDepartmentInfo_Empty(t *testing.T) {
	t.Parallel()

	info := DepartmentInfo{}

	assert.Empty(t, info.DepartmentId)
	assert.Empty(t, info.Name)
	assert.Nil(t, info.ParentDeps)
}

func TestDepartmentInfo_WithChineseName(t *testing.T) {
	t.Parallel()

	info := DepartmentInfo{
		DepartmentId: "部门-123",
		Name:         "测试部门",
	}

	assert.Equal(t, "部门-123", info.DepartmentId)
	assert.Equal(t, "测试部门", info.Name)
}

func TestDepartmentInfo_WithOnlyDepartmentId(t *testing.T) {
	t.Parallel()

	info := DepartmentInfo{
		DepartmentId: "dept-123",
	}

	assert.Equal(t, "dept-123", info.DepartmentId)
	assert.Empty(t, info.Name)
	assert.Nil(t, info.ParentDeps)
}

func TestDepartmentInfo_WithOnlyName(t *testing.T) {
	t.Parallel()

	info := DepartmentInfo{
		Name: "Test Department",
	}

	assert.Empty(t, info.DepartmentId)
	assert.Equal(t, "Test Department", info.Name)
	assert.Nil(t, info.ParentDeps)
}

func TestDepartmentInfo_WithSingleParentDep(t *testing.T) {
	t.Parallel()

	parentDeps := []*ParentDepInfo{
		{ID: "parent-1", Name: "Parent 1", Type: "department"},
	}

	info := DepartmentInfo{
		DepartmentId: "dept-123",
		Name:         "Test Department",
		ParentDeps:   parentDeps,
	}

	assert.Len(t, info.ParentDeps, 1)
	assert.Equal(t, "parent-1", info.ParentDeps[0].ID)
}

func TestDepartmentInfo_WithMultipleParentDeps(t *testing.T) {
	t.Parallel()

	parentDeps := make([]*ParentDepInfo, 10)
	for i := 0; i < 10; i++ {
		parentDeps[i] = &ParentDepInfo{
			ID:   "parent-" + string(rune(i)),
			Name: "Parent " + string(rune(i)),
			Type: "department",
		}
	}

	info := DepartmentInfo{
		DepartmentId: "dept-123",
		Name:         "Test Department",
		ParentDeps:   parentDeps,
	}

	assert.Len(t, info.ParentDeps, 10)
}

func TestDepartmentInfo_WithEmptyParentDeps(t *testing.T) {
	t.Parallel()

	parentDeps := []*ParentDepInfo{}

	info := DepartmentInfo{
		DepartmentId: "dept-123",
		Name:         "Test Department",
		ParentDeps:   parentDeps,
	}

	assert.NotNil(t, info.ParentDeps)
	assert.Len(t, info.ParentDeps, 0)
}

func TestDepartmentInfo_WithNilParentDeps(t *testing.T) {
	t.Parallel()

	info := DepartmentInfo{
		DepartmentId: "dept-123",
		Name:         "Test Department",
		ParentDeps:   nil,
	}

	assert.Nil(t, info.ParentDeps)
}

func TestParentDepInfo_AllFieldsSet(t *testing.T) {
	t.Parallel()

	info := ParentDepInfo{
		ID:   "test-id",
		Name: "test-name",
		Type: "department",
	}

	assert.Equal(t, "test-id", info.ID)
	assert.Equal(t, "test-name", info.Name)
	assert.Equal(t, "department", info.Type)
}

func TestDepartmentInfo_AllFieldsSet(t *testing.T) {
	t.Parallel()

	parentDeps := []*ParentDepInfo{
		{ID: "parent-1", Name: "Parent 1", Type: "department"},
	}

	info := DepartmentInfo{
		DepartmentId: "test-dept-id",
		Name:         "test-dept-name",
		ParentDeps:   parentDeps,
	}

	assert.Equal(t, "test-dept-id", info.DepartmentId)
	assert.Equal(t, "test-dept-name", info.Name)
	assert.Len(t, info.ParentDeps, 1)
}
