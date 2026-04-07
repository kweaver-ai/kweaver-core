package visithistoryreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisitHistoryReq_StructFields(t *testing.T) {
	t.Parallel()

	req := VisitHistoryReq{
		UserID:    "user-123",
		StartTime: 1640995200000, // 2022-01-01 00:00:00 UTC in milliseconds
		EndTime:   1641081600000, // 2022-01-02 00:00:00 UTC in milliseconds
	}

	assert.Equal(t, "user-123", req.UserID)
	assert.Equal(t, int64(1640995200000), req.StartTime)
	assert.Equal(t, int64(1641081600000), req.EndTime)
}

func TestVisitHistoryReq_EmptyValues(t *testing.T) {
	t.Parallel()

	req := VisitHistoryReq{}

	assert.Empty(t, req.UserID)
	assert.Equal(t, int64(0), req.StartTime)
	assert.Equal(t, int64(0), req.EndTime)
}

func TestVisitHistoryReq_WithPageSize(t *testing.T) {
	t.Parallel()

	req := VisitHistoryReq{
		UserID:    "user-456",
		StartTime: 1640995200000,
		EndTime:   1641081600000,
	}

	// Set PageSize fields (embedded struct)
	req.Size = 20
	req.Page = 2

	assert.Equal(t, "user-456", req.UserID)
	assert.Equal(t, 20, req.Size)
	assert.Equal(t, 2, req.Page)
}

func TestVisitHistoryReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := VisitHistoryReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	// Should contain error messages from embedded PageSize struct
	assert.Contains(t, errMsgMap, "Size.numeric")
	assert.Contains(t, errMsgMap, "Size.max")
	assert.Contains(t, errMsgMap, "Page.numeric")
	assert.Contains(t, errMsgMap, "Page.min")
}

func TestVisitHistoryReq_TimeRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		startTime int64
		endTime   int64
		valid     bool
	}{
		{
			name:      "valid time range",
			startTime: 1640995200000,
			endTime:   1641081600000,
			valid:     true,
		},
		{
			name:      "same start and end time",
			startTime: 1640995200000,
			endTime:   1640995200000,
			valid:     true,
		},
		{
			name:      "zero times",
			startTime: 0,
			endTime:   0,
			valid:     true, // May be valid for "all time" queries
		},
		{
			name:      "end before start",
			startTime: 1641081600000,
			endTime:   1640995200000,
			valid:     false, // Invalid range
		},
		{
			name:      "negative start time",
			startTime: -1000,
			endTime:   1640995200000,
			valid:     true, // End time > start time, so technically valid range
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := VisitHistoryReq{
				StartTime: tt.startTime,
				EndTime:   tt.endTime,
			}

			if tt.valid {
				assert.True(t, req.EndTime >= req.StartTime, "End time should be >= start time for valid ranges")
			} else {
				assert.False(t, req.EndTime >= req.StartTime, "End time should be < start time for invalid ranges")
			}
		})
	}
}

func TestVisitHistoryReq_PageSizeMethods(t *testing.T) {
	t.Parallel()

	req := VisitHistoryReq{
		UserID:    "user-789",
		StartTime: 1640995200000,
		EndTime:   1641081600000,
	}

	// Test GetSize with default
	size := req.GetSize()
	assert.Equal(t, 10, size)

	// Test GetPage with default
	page := req.GetPage()
	assert.Equal(t, 1, page)

	// Test GetOffset
	offset := req.GetOffset()
	assert.Equal(t, 0, offset)

	// Test GetLimit
	limit := req.GetLimit()
	assert.Equal(t, 10, limit)

	// Now set custom values
	req.Size = 50
	req.Page = 3

	assert.Equal(t, 50, req.GetSize())
	assert.Equal(t, 3, req.GetPage())
	assert.Equal(t, 100, req.GetOffset()) // (3-1)*50
	assert.Equal(t, 50, req.GetLimit())
}

func TestVisitHistoryReq_ToLimitOffset(t *testing.T) {
	t.Parallel()

	req := VisitHistoryReq{
		UserID:    "user-999",
		StartTime: 1640995200000,
		EndTime:   1641081600000,
	}
	req.Size = 25
	req.Page = 4

	lo := req.ToLimitOffset()

	assert.Equal(t, 25, lo.Limit)
	assert.Equal(t, 75, lo.Offset) // (4-1)*25
}

func TestVisitHistoryReq_Complete(t *testing.T) {
	t.Parallel()

	req := VisitHistoryReq{
		UserID:    "user-complete",
		StartTime: 1640995200000,
		EndTime:   1672531200000, // One year later
	}
	req.Size = 100
	req.Page = 5

	assert.Equal(t, "user-complete", req.UserID)
	assert.Equal(t, int64(1640995200000), req.StartTime)
	assert.Equal(t, int64(1672531200000), req.EndTime)
	assert.Equal(t, 100, req.GetSize())
	assert.Equal(t, 5, req.GetPage())
	assert.Equal(t, 400, req.GetOffset()) // (5-1)*100
	assert.Equal(t, 100, req.GetLimit())
}

func TestVisitHistoryReq_EmptyWithDefaults(t *testing.T) {
	t.Parallel()

	req := VisitHistoryReq{}

	// PageSize methods should work with defaults
	assert.Equal(t, 10, req.GetSize())
	assert.Equal(t, 1, req.GetPage())
	assert.Equal(t, 0, req.GetOffset())
	assert.Equal(t, 10, req.GetLimit())

	// ToLimitOffset should work
	lo := req.ToLimitOffset()
	assert.Equal(t, 10, lo.Limit)
	assert.Equal(t, 0, lo.Offset)
}

func TestVisitHistoryReq_LargePageNumbers(t *testing.T) {
	t.Parallel()

	req := VisitHistoryReq{
		UserID:    "user-large",
		StartTime: 1640995200000,
		EndTime:   1641081600000,
	}
	req.Page = 1000
	req.Size = 50

	assert.Equal(t, 1000, req.GetPage())
	assert.Equal(t, 50, req.GetSize())
	assert.Equal(t, 49950, req.GetOffset()) // (1000-1)*50
	assert.Equal(t, 50, req.GetLimit())
}
