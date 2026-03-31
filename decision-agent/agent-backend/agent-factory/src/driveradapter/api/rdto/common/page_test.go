package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageSize_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	p := PageSize{}

	errMsgMap := p.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"size"的值必须是数字`, errMsgMap["Size.numeric"])
	assert.Equal(t, `"size"的值不能大于1000`, errMsgMap["Size.max"])
	assert.Equal(t, `"page"的值必须是数字`, errMsgMap["Page.numeric"])
	assert.Equal(t, `"page"的值不能小于1`, errMsgMap["Page.min"])
}

func TestPageSize_GetSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{
			name:     "zero size returns default",
			size:     0,
			expected: 10,
		},
		{
			name:     "non-zero size returns itself",
			size:     50,
			expected: 50,
		},
		{
			name:     "size of 1 returns 1",
			size:     1,
			expected: 1,
		},
		{
			name:     "size of 1000 returns 1000",
			size:     1000,
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := PageSize{Size: tt.size}
			result := p.GetSize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPageSize_GetPage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		page     int
		expected int
	}{
		{
			name:     "zero page returns default",
			page:     0,
			expected: 1,
		},
		{
			name:     "non-zero page returns itself",
			page:     5,
			expected: 5,
		},
		{
			name:     "page of 1 returns 1",
			page:     1,
			expected: 1,
		},
		{
			name:     "page of 100 returns 100",
			page:     100,
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := PageSize{Page: tt.page}
			result := p.GetPage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPageSize_GetOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		page     int
		size     int
		expected int
	}{
		{
			name:     "first page with default size",
			page:     1,
			size:     10,
			expected: 0,
		},
		{
			name:     "second page",
			page:     2,
			size:     10,
			expected: 10,
		},
		{
			name:     "third page with custom size",
			page:     3,
			size:     25,
			expected: 50,
		},
		{
			name:     "zero page uses default",
			page:     0,
			size:     10,
			expected: 0, // (1-1)*10 = 0
		},
		{
			name:     "zero size uses default",
			page:     5,
			size:     0,
			expected: 40, // (5-1)*10 = 40
		},
		{
			name:     "both zero use defaults",
			page:     0,
			size:     0,
			expected: 0, // (1-1)*10 = 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := PageSize{Page: tt.page, Size: tt.size}
			result := p.GetOffset()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPageSize_GetLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{
			name:     "zero size returns default limit",
			size:     0,
			expected: 10,
		},
		{
			name:     "non-zero size returns itself as limit",
			size:     50,
			expected: 50,
		},
		{
			name:     "size of 1 returns limit 1",
			size:     1,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := PageSize{Size: tt.size}
			result := p.GetLimit()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPageSize_ToLimitOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		page          int
		size          int
		expectedLimit int
		expectedOff   int
	}{
		{
			name:          "first page",
			page:          1,
			size:          10,
			expectedLimit: 10,
			expectedOff:   0,
		},
		{
			name:          "second page",
			page:          2,
			size:          20,
			expectedLimit: 20,
			expectedOff:   20,
		},
		{
			name:          "zero values use defaults",
			page:          0,
			size:          0,
			expectedLimit: 10,
			expectedOff:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := PageSize{Page: tt.page, Size: tt.size}
			result := p.ToLimitOffset()

			assert.Equal(t, tt.expectedLimit, result.Limit)
			assert.Equal(t, tt.expectedOff, result.Offset)
		})
	}
}

func TestPageSize_EmptyStruct(t *testing.T) {
	t.Parallel()

	p := PageSize{}

	assert.Equal(t, 0, p.Size)
	assert.Equal(t, 0, p.Page)
	assert.Equal(t, 10, p.GetSize())
	assert.Equal(t, 1, p.GetPage())
}

func TestPageSize_WithValues(t *testing.T) {
	t.Parallel()

	p := PageSize{
		Page: 5,
		Size: 50,
	}

	assert.Equal(t, 5, p.Page)
	assert.Equal(t, 50, p.Size)
	assert.Equal(t, 50, p.GetSize())
	assert.Equal(t, 5, p.GetPage())
	assert.Equal(t, 200, p.GetOffset())
	assert.Equal(t, 50, p.GetLimit())
}

func TestPageByLastIntID_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	p := PageByLastIntID{}
	errMsgMap := p.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"size"的值必须是数字`, errMsgMap["Size.numeric"])
	assert.Equal(t, `"size"的值不能大于1000`, errMsgMap["Size.max"])
	assert.Equal(t, `"last_id"的值必须是数字`, errMsgMap["LastID.numeric"])
}

func TestPageByLastIntID_GetSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{
			name:     "zero size returns default",
			size:     0,
			expected: 10,
		},
		{
			name:     "non-zero size returns itself",
			size:     50,
			expected: 50,
		},
		{
			name:     "size of 1 returns 1",
			size:     1,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := PageByLastIntID{Size: tt.size}
			result := p.GetSize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPageByLastIntID_GetLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{
			name:     "zero size returns default",
			size:     0,
			expected: 10,
		},
		{
			name:     "non-zero size returns itself",
			size:     25,
			expected: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := PageByLastIntID{Size: tt.size}
			result := p.GetLimit()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPageByLastIntID_Fields(t *testing.T) {
	t.Parallel()

	p := PageByLastIntID{
		Size:   20,
		LastID: 123,
	}

	assert.Equal(t, 20, p.Size)
	assert.Equal(t, 123, p.LastID)
}

func TestPageByStrID_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	p := PageByStrID{}
	errMsgMap := p.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"size"的值必须是数字`, errMsgMap["Size.numeric"])
	assert.Equal(t, `"size"的值不能大于1000`, errMsgMap["Size.max"])
}

func TestPageByStrID_GetSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{
			name:     "zero size returns default",
			size:     0,
			expected: 10,
		},
		{
			name:     "non-zero size returns itself",
			size:     50,
			expected: 50,
		},
		{
			name:     "size of 1 returns 1",
			size:     1,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := PageByStrID{Size: tt.size}
			result := p.GetSize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPageByStrID_GetLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     int
		expected int
	}{
		{
			name:     "zero size returns default",
			size:     0,
			expected: 10,
		},
		{
			name:     "non-zero size returns itself",
			size:     25,
			expected: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := PageByStrID{Size: tt.size}
			result := p.GetLimit()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPageByStrID_Fields(t *testing.T) {
	t.Parallel()

	p := PageByStrID{
		Size:   20,
		LastID: "abc123",
	}

	assert.Equal(t, 20, p.Size)
	assert.Equal(t, "abc123", p.LastID)
}
