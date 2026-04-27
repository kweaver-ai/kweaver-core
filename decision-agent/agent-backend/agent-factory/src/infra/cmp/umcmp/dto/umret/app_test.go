package umret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppInfo_StructFields(t *testing.T) {
	t.Parallel()

	info := AppInfo{
		ID:   "app-123",
		Name: "Test App",
	}

	assert.Equal(t, "app-123", info.ID)
	assert.Equal(t, "Test App", info.Name)
}

func TestAppInfo_Empty(t *testing.T) {
	t.Parallel()

	info := AppInfo{}

	assert.Empty(t, info.ID)
	assert.Empty(t, info.Name)
}

func TestAppInfo_WithChineseName(t *testing.T) {
	t.Parallel()

	info := AppInfo{
		ID:   "应用-123",
		Name: "测试应用",
	}

	assert.Equal(t, "应用-123", info.ID)
	assert.Equal(t, "测试应用", info.Name)
}

func TestAppListResDto_StructFields(t *testing.T) {
	t.Parallel()

	dto := AppListResDto{
		Entries: []*AppInfo{
			{ID: "app-1", Name: "App 1"},
			{ID: "app-2", Name: "App 2"},
		},
		TotalCount: 100,
	}

	assert.Len(t, dto.Entries, 2)
	assert.Equal(t, int64(100), dto.TotalCount)
	assert.Equal(t, "app-1", dto.Entries[0].ID)
}

func TestAppListResDto_Empty(t *testing.T) {
	t.Parallel()

	dto := AppListResDto{}

	assert.Nil(t, dto.Entries)
	assert.Equal(t, int64(0), dto.TotalCount)
}

func TestAppListResDto_AddEntries(t *testing.T) {
	t.Parallel()

	dto := AppListResDto{}

	dto.Entries = append(dto.Entries, &AppInfo{
		ID:   "app-1",
		Name: "App 1",
	})
	dto.Entries = append(dto.Entries, &AppInfo{
		ID:   "app-2",
		Name: "App 2",
	})

	assert.Len(t, dto.Entries, 2)
	assert.Equal(t, "app-1", dto.Entries[0].ID)
	assert.Equal(t, "app-2", dto.Entries[1].ID)
}

func TestAppListResDto_WithTotalCount(t *testing.T) {
	t.Parallel()

	totalCounts := []int64{0, 1, 100, 1000, 999999}

	for _, tc := range totalCounts {
		dto := &AppListResDto{
			TotalCount: tc,
		}
		assert.Equal(t, tc, dto.TotalCount)
	}
}

func TestAppListResDto_WithMultipleEntries(t *testing.T) {
	t.Parallel()

	entries := make([]*AppInfo, 50)
	for i := 0; i < 50; i++ {
		entries[i] = &AppInfo{
			ID:   "app-" + string(rune(i)),
			Name: "App " + string(rune(i)),
		}
	}

	dto := &AppListResDto{
		Entries:    entries,
		TotalCount: 50,
	}

	assert.Len(t, dto.Entries, 50)
	assert.Equal(t, int64(50), dto.TotalCount)
}

func TestAppListResDto_Iteration(t *testing.T) {
	t.Parallel()

	dto := AppListResDto{}

	for i := 0; i < 5; i++ {
		dto.Entries = append(dto.Entries, &AppInfo{
			ID:   "app-" + string(rune(i)),
			Name: "App " + string(rune(i)),
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

func TestAppInfo_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	info := AppInfo{
		ID:   "app@#$%",
		Name: "App Name (Test)",
	}

	assert.Equal(t, "app@#$%", info.ID)
	assert.Equal(t, "App Name (Test)", info.Name)
}

func TestAppListResDto_WithOnlyTotalCount(t *testing.T) {
	t.Parallel()

	dto := AppListResDto{
		TotalCount: 10,
	}

	assert.Equal(t, int64(10), dto.TotalCount)
	assert.Nil(t, dto.Entries)
}

func TestAppListResDto_WithOnlyEntries(t *testing.T) {
	t.Parallel()

	dto := AppListResDto{
		Entries: []*AppInfo{
			{ID: "app-1", Name: "App 1"},
		},
	}

	assert.Len(t, dto.Entries, 1)
	assert.Equal(t, int64(0), dto.TotalCount)
}
