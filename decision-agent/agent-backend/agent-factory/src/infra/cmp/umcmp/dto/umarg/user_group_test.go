package umarg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserGroupListArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetUserGroupListArgDto{
		Limit:   10,
		Offset:  5,
		Keyword: "test",
	}

	assert.Equal(t, 10, dto.Limit)
	assert.Equal(t, 5, dto.Offset)
	assert.Equal(t, "test", dto.Keyword)
}

func TestGetUserGroupListArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetUserGroupListArgDto{}

	assert.Equal(t, 0, dto.Limit)
	assert.Equal(t, 0, dto.Offset)
	assert.Empty(t, dto.Keyword)
}

func TestGetUserGroupListArgDto_WithLimit(t *testing.T) {
	t.Parallel()

	limits := []int{1, 10, 100, 1000}

	for _, limit := range limits {
		dto := GetUserGroupListArgDto{
			Limit: limit,
		}
		assert.Equal(t, limit, dto.Limit)
	}
}

func TestGetUserGroupListArgDto_WithOffset(t *testing.T) {
	t.Parallel()

	offsets := []int{0, 10, 100, 999}

	for _, offset := range offsets {
		dto := GetUserGroupListArgDto{
			Offset: offset,
		}
		assert.Equal(t, offset, dto.Offset)
	}
}

func TestGetUserGroupListArgDto_WithKeyword(t *testing.T) {
	t.Parallel()

	keywords := []string{
		"test group",
		"测试组",
		"",
		"group-123",
	}

	for _, keyword := range keywords {
		dto := GetUserGroupListArgDto{
			Keyword: keyword,
		}
		assert.Equal(t, keyword, dto.Keyword)
	}
}

func TestGetUserGroupListArgDto_WithChineseKeyword(t *testing.T) {
	t.Parallel()

	dto := GetUserGroupListArgDto{
		Keyword: "测试组名",
	}

	assert.Equal(t, "测试组名", dto.Keyword)
}

func TestGetUserGroupListArgDto_WithOnlyLimit(t *testing.T) {
	t.Parallel()

	dto := GetUserGroupListArgDto{
		Limit: 50,
	}

	assert.Equal(t, 50, dto.Limit)
	assert.Equal(t, 0, dto.Offset)
	assert.Empty(t, dto.Keyword)
}

func TestGetUserGroupListArgDto_WithOnlyOffset(t *testing.T) {
	t.Parallel()

	dto := GetUserGroupListArgDto{
		Offset: 100,
	}

	assert.Equal(t, 0, dto.Limit)
	assert.Equal(t, 100, dto.Offset)
	assert.Empty(t, dto.Keyword)
}

func TestGetUserGroupListArgDto_BoundaryValues(t *testing.T) {
	t.Parallel()

	dto := GetUserGroupListArgDto{
		Limit:   1,
		Offset:  0,
		Keyword: "",
	}

	assert.Equal(t, 1, dto.Limit)
	assert.Equal(t, 0, dto.Offset)
	assert.Empty(t, dto.Keyword)
}

func TestGetUserGroupListArgDto_Limits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		limit     int
		wantValid bool
	}{
		{name: "minimum limit", limit: 1, wantValid: true},
		{name: "maximum limit", limit: 1000, wantValid: true},
		{name: "zero limit", limit: 0, wantValid: true},
		{name: "negative limit", limit: -1, wantValid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dto := GetUserGroupListArgDto{
				Limit: tt.limit,
			}
			assert.Equal(t, tt.limit, dto.Limit)
		})
	}
}

func TestGetUserGroupListArgDto_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	dto := GetUserGroupListArgDto{
		Keyword: "group@#$%",
	}

	assert.Equal(t, "group@#$%", dto.Keyword)
}

func TestGetUserGroupListArgDto_AllFieldsSet(t *testing.T) {
	t.Parallel()

	dto := GetUserGroupListArgDto{
		Limit:   100,
		Offset:  50,
		Keyword: "search term",
	}

	assert.Equal(t, 100, dto.Limit)
	assert.Equal(t, 50, dto.Offset)
	assert.Equal(t, "search term", dto.Keyword)
}

func TestGetUserGroupListArgDto_WithDifferentKeywords(t *testing.T) {
	t.Parallel()

	keywords := []string{
		"admin",
		"user",
		"测试",
		"",
		"123",
		"!@#$%",
	}

	for _, keyword := range keywords {
		dto := GetUserGroupListArgDto{
			Keyword: keyword,
		}
		assert.Equal(t, keyword, dto.Keyword)
	}
}
