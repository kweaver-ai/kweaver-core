package umret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchOrgRetDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := SearchOrgRetDto{
		UserIDs:       []string{"user-1", "user-2"},
		DepartmentIDs: []string{"dept-1", "dept-2", "dept-3"},
	}

	assert.Len(t, dto.UserIDs, 2)
	assert.Len(t, dto.DepartmentIDs, 3)
	assert.Equal(t, "user-1", dto.UserIDs[0])
	assert.Equal(t, "dept-1", dto.DepartmentIDs[0])
}

func TestSearchOrgRetDto_Empty(t *testing.T) {
	t.Parallel()

	dto := SearchOrgRetDto{}

	assert.Nil(t, dto.UserIDs)
	assert.Nil(t, dto.DepartmentIDs)
}

func TestSearchOrgRetDto_WithOnlyUserIDs(t *testing.T) {
	t.Parallel()

	dto := SearchOrgRetDto{
		UserIDs: []string{"user-1", "user-2", "user-3"},
	}

	assert.Len(t, dto.UserIDs, 3)
	assert.Nil(t, dto.DepartmentIDs)
}

func TestSearchOrgRetDto_WithOnlyDepartmentIDs(t *testing.T) {
	t.Parallel()

	dto := SearchOrgRetDto{
		DepartmentIDs: []string{"dept-1", "dept-2"},
	}

	assert.Nil(t, dto.UserIDs)
	assert.Len(t, dto.DepartmentIDs, 2)
}

func TestSearchOrgRetDto_WithChineseIDs(t *testing.T) {
	t.Parallel()

	dto := SearchOrgRetDto{
		UserIDs:       []string{"用户-1", "用户-2"},
		DepartmentIDs: []string{"部门-1"},
	}

	assert.Equal(t, "用户-1", dto.UserIDs[0])
	assert.Equal(t, "部门-1", dto.DepartmentIDs[0])
}

func TestSearchOrgRetDto_WithEmptyArrays(t *testing.T) {
	t.Parallel()

	dto := SearchOrgRetDto{
		UserIDs:       []string{},
		DepartmentIDs: []string{},
	}

	assert.NotNil(t, dto.UserIDs)
	assert.NotNil(t, dto.DepartmentIDs)
	assert.Len(t, dto.UserIDs, 0)
	assert.Len(t, dto.DepartmentIDs, 0)
}

func TestSearchOrgRetDto_AppendUserIDs(t *testing.T) {
	t.Parallel()

	dto := &SearchOrgRetDto{}
	dto.UserIDs = append(dto.UserIDs, "user-1")
	dto.UserIDs = append(dto.UserIDs, "user-2")

	assert.Len(t, dto.UserIDs, 2)
}

func TestSearchOrgRetDto_AppendDepartmentIDs(t *testing.T) {
	t.Parallel()

	dto := &SearchOrgRetDto{}
	dto.DepartmentIDs = append(dto.DepartmentIDs, "dept-1")
	dto.DepartmentIDs = append(dto.DepartmentIDs, "dept-2")

	assert.Len(t, dto.DepartmentIDs, 2)
}

func TestSearchOrgRetDto_WithMultipleIDs(t *testing.T) {
	t.Parallel()

	userIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		userIDs[i] = "user-" + string(rune(i))
	}

	deptIDs := make([]string, 50)
	for i := 0; i < 50; i++ {
		deptIDs[i] = "dept-" + string(rune(i))
	}

	dto := SearchOrgRetDto{
		UserIDs:       userIDs,
		DepartmentIDs: deptIDs,
	}

	assert.Len(t, dto.UserIDs, 100)
	assert.Len(t, dto.DepartmentIDs, 50)
}

func TestSearchOrgRetDto_SliceOperations(t *testing.T) {
	t.Parallel()

	dto := SearchOrgRetDto{
		UserIDs:       []string{"user-1", "user-2", "user-3"},
		DepartmentIDs: []string{"dept-1", "dept-2"},
	}

	// Test slicing
	subUserIDs := dto.UserIDs[1:3]
	assert.Len(t, subUserIDs, 2)
	assert.Equal(t, "user-2", subUserIDs[0])

	// Test iteration
	count := 0

	for _, userID := range dto.UserIDs {
		assert.NotEmpty(t, userID)

		count++
	}

	assert.Equal(t, 3, count)
}

func TestSearchOrgRetDto_WithBothTypes(t *testing.T) {
	t.Parallel()

	dto := &SearchOrgRetDto{
		UserIDs:       []string{"user-1"},
		DepartmentIDs: []string{"dept-1"},
	}

	assert.NotNil(t, dto.UserIDs)
	assert.NotNil(t, dto.DepartmentIDs)
}
