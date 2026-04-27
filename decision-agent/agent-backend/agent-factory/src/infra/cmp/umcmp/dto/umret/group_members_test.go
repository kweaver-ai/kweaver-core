package umret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGetGroupMembersRetDto(t *testing.T) {
	t.Parallel()

	dto := NewGetGroupMembersRetDto()

	assert.NotNil(t, dto)
	assert.NotNil(t, dto.UserIDs)
	assert.NotNil(t, dto.DepartmentIDs)
	assert.IsType(t, &GetGroupMembersRetDto{}, dto)
}

func TestGetGroupMembersRetDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetGroupMembersRetDto{
		UserIDs:       []string{"user-1", "user-2"},
		DepartmentIDs: []string{"dept-1", "dept-2", "dept-3"},
	}

	assert.Len(t, dto.UserIDs, 2)
	assert.Len(t, dto.DepartmentIDs, 3)
	assert.Equal(t, "user-1", dto.UserIDs[0])
	assert.Equal(t, "user-2", dto.UserIDs[1])
	assert.Equal(t, "dept-1", dto.DepartmentIDs[0])
}

func TestGetGroupMembersRetDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetGroupMembersRetDto{}

	assert.Nil(t, dto.UserIDs)
	assert.Nil(t, dto.DepartmentIDs)
}

func TestGetGroupMembersRetDto_WithNew(t *testing.T) {
	t.Parallel()

	dto := NewGetGroupMembersRetDto()

	assert.NotNil(t, dto.UserIDs)
	assert.NotNil(t, dto.DepartmentIDs)
	assert.Len(t, dto.UserIDs, 0)
	assert.Len(t, dto.DepartmentIDs, 0)
}

func TestGetGroupMembersRetDto_AddUserIDs(t *testing.T) {
	t.Parallel()

	dto := NewGetGroupMembersRetDto()

	dto.UserIDs = append(dto.UserIDs, "user-1")
	dto.UserIDs = append(dto.UserIDs, "user-2")

	assert.Len(t, dto.UserIDs, 2)
	assert.Equal(t, "user-1", dto.UserIDs[0])
	assert.Equal(t, "user-2", dto.UserIDs[1])
}

func TestGetGroupMembersRetDto_AddDepartmentIDs(t *testing.T) {
	t.Parallel()

	dto := NewGetGroupMembersRetDto()

	dto.DepartmentIDs = append(dto.DepartmentIDs, "dept-1")
	dto.DepartmentIDs = append(dto.DepartmentIDs, "dept-2")
	dto.DepartmentIDs = append(dto.DepartmentIDs, "dept-3")

	assert.Len(t, dto.DepartmentIDs, 3)
}

func TestGetGroupMembersRetDto_WithBothTypes(t *testing.T) {
	t.Parallel()

	dto := &GetGroupMembersRetDto{
		UserIDs:       []string{"user-1", "user-2"},
		DepartmentIDs: []string{"dept-1"},
	}

	assert.Len(t, dto.UserIDs, 2)
	assert.Len(t, dto.DepartmentIDs, 1)
}

func TestGetGroupMembersRetDto_WithChineseIDs(t *testing.T) {
	t.Parallel()

	dto := &GetGroupMembersRetDto{
		UserIDs:       []string{"用户-1", "用户-2"},
		DepartmentIDs: []string{"部门-1"},
	}

	assert.Len(t, dto.UserIDs, 2)
	assert.Equal(t, "用户-1", dto.UserIDs[0])
	assert.Equal(t, "部门-1", dto.DepartmentIDs[0])
}

func TestGetGroupMembersRetDto_WithEmptyArrays(t *testing.T) {
	t.Parallel()

	dto := &GetGroupMembersRetDto{
		UserIDs:       []string{},
		DepartmentIDs: []string{},
	}

	assert.NotNil(t, dto.UserIDs)
	assert.NotNil(t, dto.DepartmentIDs)
	assert.Len(t, dto.UserIDs, 0)
	assert.Len(t, dto.DepartmentIDs, 0)
}

func TestGetGroupMembersRetDto_WithDuplicateIDs(t *testing.T) {
	t.Parallel()

	dto := &GetGroupMembersRetDto{
		UserIDs:       []string{"user-1", "user-1", "user-2"},
		DepartmentIDs: []string{"dept-1", "dept-1"},
	}

	assert.Len(t, dto.UserIDs, 3)
	assert.Len(t, dto.DepartmentIDs, 2)
}

func TestGetGroupMembersRetDto_Capacity(t *testing.T) {
	t.Parallel()

	dto := NewGetGroupMembersRetDto()

	// Add many users
	for i := 0; i < 100; i++ {
		dto.UserIDs = append(dto.UserIDs, "user-"+string(rune(i)))
	}

	assert.Len(t, dto.UserIDs, 100)
}

func TestGetGroupMembersRetDto_ModifyArrays(t *testing.T) {
	t.Parallel()

	dto := NewGetGroupMembersRetDto()

	dto.UserIDs = append(dto.UserIDs, "user-1")
	dto.DepartmentIDs = append(dto.DepartmentIDs, "dept-1")

	// Modify arrays
	dto.UserIDs[0] = "user-updated"
	dto.DepartmentIDs[0] = "dept-updated"

	assert.Equal(t, "user-updated", dto.UserIDs[0])
	assert.Equal(t, "dept-updated", dto.DepartmentIDs[0])
}

func TestGetGroupMembersRetDto_WithNilArrays(t *testing.T) {
	t.Parallel()

	dto := &GetGroupMembersRetDto{
		UserIDs:       nil,
		DepartmentIDs: nil,
	}

	assert.Nil(t, dto.UserIDs)
	assert.Nil(t, dto.DepartmentIDs)
}

func TestGetGroupMembersRetDto_NewVsDirect(t *testing.T) {
	t.Parallel()

	dto1 := NewGetGroupMembersRetDto()
	dto2 := &GetGroupMembersRetDto{}

	// NewGetGroupMembersRetDto initializes arrays
	assert.NotNil(t, dto1.UserIDs)
	assert.NotNil(t, dto1.DepartmentIDs)

	// Direct initialization doesn't
	assert.Nil(t, dto2.UserIDs)
	assert.Nil(t, dto2.DepartmentIDs)
}

func TestGetGroupMembersRetDto_SliceOperations(t *testing.T) {
	t.Parallel()

	userIDs := []string{"user-1", "user-2", "user-3"}
	deptIDs := []string{"dept-1", "dept-2"}

	dto := &GetGroupMembersRetDto{
		UserIDs:       userIDs,
		DepartmentIDs: deptIDs,
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
