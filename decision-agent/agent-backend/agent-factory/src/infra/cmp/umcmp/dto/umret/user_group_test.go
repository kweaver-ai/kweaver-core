package umret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserGroupInfo_StructFields(t *testing.T) {
	t.Parallel()

	info := UserGroupInfo{
		ID:   "group-123",
		Name: "Test Group",
	}

	assert.Equal(t, "group-123", info.ID)
	assert.Equal(t, "Test Group", info.Name)
}

func TestUserGroupInfo_Empty(t *testing.T) {
	t.Parallel()

	info := UserGroupInfo{}

	assert.Empty(t, info.ID)
	assert.Empty(t, info.Name)
}

func TestNewUserGroupListResDto(t *testing.T) {
	t.Parallel()

	dto := NewUserGroupListResDto()

	assert.NotNil(t, dto)
	assert.NotNil(t, dto.Entries)
	assert.IsType(t, &UserGroupListResDto{}, dto)
}

func TestUserGroupListResDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := UserGroupListResDto{
		Entries: []*UserGroupInfo{
			{ID: "group-1", Name: "Group 1"},
			{ID: "group-2", Name: "Group 2"},
		},
		TotalCount: 100,
	}

	assert.Len(t, dto.Entries, 2)
	assert.Equal(t, int64(100), dto.TotalCount)
	assert.Equal(t, "group-1", dto.Entries[0].ID)
}

func TestUserGroupListResDto_Empty(t *testing.T) {
	t.Parallel()

	dto := UserGroupListResDto{}

	assert.Nil(t, dto.Entries)
	assert.Equal(t, int64(0), dto.TotalCount)
}

func TestUserGroupListResDto_AddEntries(t *testing.T) {
	t.Parallel()

	dto := NewUserGroupListResDto()

	dto.Entries = append(dto.Entries, &UserGroupInfo{
		ID:   "group-1",
		Name: "Group 1",
	})
	dto.Entries = append(dto.Entries, &UserGroupInfo{
		ID:   "group-2",
		Name: "Group 2",
	})

	assert.Len(t, dto.Entries, 2)
}

func TestUserGroupInfo_WithChineseCharacters(t *testing.T) {
	t.Parallel()

	info := UserGroupInfo{
		ID:   "组-123",
		Name: "测试组名",
	}

	assert.Equal(t, "组-123", info.ID)
	assert.Equal(t, "测试组名", info.Name)
}

func TestUserGroupListResDto_WithTotalCount(t *testing.T) {
	t.Parallel()

	totalCounts := []int64{0, 1, 100, 1000, 999999}

	for _, tc := range totalCounts {
		dto := &UserGroupListResDto{
			TotalCount: tc,
		}
		assert.Equal(t, tc, dto.TotalCount)
	}
}

func TestUserGroupListResDto_WithMultipleEntries(t *testing.T) {
	t.Parallel()

	entries := make([]*UserGroupInfo, 50)
	for i := 0; i < 50; i++ {
		entries[i] = &UserGroupInfo{
			ID:   "group-" + string(rune(i)),
			Name: "Group " + string(rune(i)),
		}
	}

	dto := &UserGroupListResDto{
		Entries:    entries,
		TotalCount: 50,
	}

	assert.Len(t, dto.Entries, 50)
	assert.Equal(t, int64(50), dto.TotalCount)
}

func TestUserGroupListResDto_Iteration(t *testing.T) {
	t.Parallel()

	dto := NewUserGroupListResDto()

	for i := 0; i < 5; i++ {
		dto.Entries = append(dto.Entries, &UserGroupInfo{
			ID:   "group-" + string(rune(i)),
			Name: "Group " + string(rune(i)),
		})
	}

	count := 0

	for _, entry := range dto.Entries {
		assert.NotEmpty(t, entry.ID)
		assert.NotEmpty(t, entry.Name)

		count++
	}

	assert.Equal(t, 5, count)
}
