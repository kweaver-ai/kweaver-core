package umarg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAppListArgDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{
		Limit:   10,
		Offset:  5,
		Keyword: "test",
	}

	assert.Equal(t, 10, dto.Limit)
	assert.Equal(t, 5, dto.Offset)
	assert.Equal(t, "test", dto.Keyword)
}

func TestGetAppListArgDto_Empty(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{}

	assert.Equal(t, 0, dto.Limit)
	assert.Equal(t, 0, dto.Offset)
	assert.Empty(t, dto.Keyword)
}

func TestGetAppListArgDto_WithLimit(t *testing.T) {
	t.Parallel()

	limits := []int{1, 10, 100, 1000}

	for _, limit := range limits {
		dto := GetAppListArgDto{
			Limit: limit,
		}
		assert.Equal(t, limit, dto.Limit)
	}
}

func TestGetAppListArgDto_WithOffset(t *testing.T) {
	t.Parallel()

	offsets := []int{0, 10, 100, 999}

	for _, offset := range offsets {
		dto := GetAppListArgDto{
			Offset: offset,
		}
		assert.Equal(t, offset, dto.Offset)
	}
}

func TestGetAppListArgDto_WithKeyword(t *testing.T) {
	t.Parallel()

	keywords := []string{
		"test app",
		"测试应用",
		"",
		"app-123",
	}

	for _, keyword := range keywords {
		dto := GetAppListArgDto{
			Keyword: keyword,
		}
		assert.Equal(t, keyword, dto.Keyword)
	}
}

func TestGetAppListArgDto_WithChineseKeyword(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{
		Keyword: "测试应用名",
	}

	assert.Equal(t, "测试应用名", dto.Keyword)
}

func TestGetAppListArgDto_WithOnlyLimit(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{
		Limit: 50,
	}

	assert.Equal(t, 50, dto.Limit)
	assert.Equal(t, 0, dto.Offset)
	assert.Empty(t, dto.Keyword)
}

func TestGetAppListArgDto_WithOnlyOffset(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{
		Offset: 100,
	}

	assert.Equal(t, 0, dto.Limit)
	assert.Equal(t, 100, dto.Offset)
	assert.Empty(t, dto.Keyword)
}

func TestGetAppListArgDto_WithOnlyKeyword(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{
		Keyword: "search",
	}

	assert.Equal(t, 0, dto.Limit)
	assert.Equal(t, 0, dto.Offset)
	assert.Equal(t, "search", dto.Keyword)
}

func TestGetAppListArgDto_BoundaryValues(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{
		Limit:   1,
		Offset:  0,
		Keyword: "",
	}

	assert.Equal(t, 1, dto.Limit)
	assert.Equal(t, 0, dto.Offset)
	assert.Empty(t, dto.Keyword)
}

func TestGetAppListArgDto_Limits(t *testing.T) {
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

			dto := GetAppListArgDto{
				Limit: tt.limit,
			}
			assert.Equal(t, tt.limit, dto.Limit)
		})
	}
}

func TestGetAppListArgDto_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{
		Keyword: "app@#$%",
	}

	assert.Equal(t, "app@#$%", dto.Keyword)
}

func TestGetAppListArgDto_AllFieldsSet(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{
		Limit:   100,
		Offset:  50,
		Keyword: "search term",
	}

	assert.Equal(t, 100, dto.Limit)
	assert.Equal(t, 50, dto.Offset)
	assert.Equal(t, "search term", dto.Keyword)
}

func TestGetAppListArgDto_WithDifferentKeywords(t *testing.T) {
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
		dto := GetAppListArgDto{
			Keyword: keyword,
		}
		assert.Equal(t, keyword, dto.Keyword)
	}
}

func TestGetAppListArgDto_WithMaximumLimit(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{
		Limit: 1000,
	}

	assert.Equal(t, 1000, dto.Limit)
}

func TestGetAppListArgDto_WithLargeOffset(t *testing.T) {
	t.Parallel()

	dto := GetAppListArgDto{
		Offset: 9999,
	}

	assert.Equal(t, 9999, dto.Offset)
}
