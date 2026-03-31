package productreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListReq_StructFields(t *testing.T) {
	t.Parallel()

	req := ListReq{}
	req.Size = 10
	req.Page = 1

	assert.Equal(t, 10, req.Size)
	assert.Equal(t, 1, req.Page)
}

func TestListReq_Empty(t *testing.T) {
	t.Parallel()

	req := ListReq{}

	assert.Equal(t, 0, req.Size)
	assert.Equal(t, 0, req.Page)
}

func TestListReq_WithPagination(t *testing.T) {
	t.Parallel()

	req := ListReq{}
	req.Size = 20
	req.Page = 2

	offset := req.GetOffset()
	assert.Equal(t, 20, offset)
}

func TestListReq_WithDefaultPagination(t *testing.T) {
	t.Parallel()

	req := ListReq{}
	// PageSize has default values when Size is 0
	req.Size = 0
	req.Page = 0

	offset := req.GetOffset()
	assert.Equal(t, 0, offset)
}

func TestListReq_PaginationEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		page     int
		size     int
		expected int
	}{
		{
			name:     "first page",
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
			name:     "large page number",
			page:     100,
			size:     20,
			expected: 1980,
		},
		{
			name:     "zero page",
			page:     0,
			size:     10,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := ListReq{}
			req.Page = tt.page
			req.Size = tt.size

			offset := req.GetOffset()
			assert.Equal(t, tt.expected, offset)
		})
	}
}

func TestListReq_EmbeddedPageSize(t *testing.T) {
	t.Parallel()

	req := ListReq{}

	// Verify that PageSize is embedded
	assert.IsType(t, req.Size, 0)
	assert.IsType(t, req.Page, 0)

	// Set and verify pagination values
	req.Size = 15
	req.Page = 3

	assert.Equal(t, 15, req.Size)
	assert.Equal(t, 3, req.Page)
	assert.Equal(t, 30, req.GetOffset())
}

func TestListReq_GetOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     int
		page     int
		expected int
	}{
		{
			name:     "page 1 size 10",
			size:     10,
			page:     1,
			expected: 0,
		},
		{
			name:     "page 2 size 10",
			size:     10,
			page:     2,
			expected: 10,
		},
		{
			name:     "page 3 size 25",
			size:     25,
			page:     3,
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := ListReq{}
			req.Size = tt.size
			req.Page = tt.page

			offset := req.GetOffset()
			assert.Equal(t, tt.expected, offset)
		})
	}
}

func TestListReq_WithLargePageSize(t *testing.T) {
	t.Parallel()

	req := ListReq{}
	req.Size = 100
	req.Page = 5

	assert.Equal(t, 100, req.Size)
	assert.Equal(t, 5, req.Page)
	assert.Equal(t, 400, req.GetOffset())
}

func TestListReq_WithZeroPageSize(t *testing.T) {
	t.Parallel()

	req := ListReq{}
	req.Size = 0
	req.Page = 1

	offset := req.GetOffset()
	assert.Equal(t, 0, offset)
}
