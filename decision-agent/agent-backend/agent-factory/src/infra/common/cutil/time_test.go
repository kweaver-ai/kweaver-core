package cutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNowStr(t *testing.T) {
	t.Parallel()

	result := NowStr()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "-")
	assert.Contains(t, result, ":")
}

func TestGetCurrentMSTimestamp(t *testing.T) {
	t.Parallel()

	timestamp := GetCurrentMSTimestamp()
	assert.Greater(t, timestamp, int64(0))
	assert.Less(t, timestamp, int64(9999999999999))
}

func TestGetCurrentTimestamp(t *testing.T) {
	t.Parallel()

	timestamp := GetCurrentTimestamp()
	assert.Greater(t, timestamp, int64(1700000000)) // After 2023
	assert.Less(t, timestamp, int64(9999999999))
}

func TestFormatTime(t *testing.T) {
	t.Parallel()

	testTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	result := FormatTime(testTime)
	assert.Equal(t, "2024-01-15 14:30:45", result)
}

func TestFormatTimeUnix(t *testing.T) {
	t.Parallel()

	// Just test that it returns a valid formatted string
	timestamp := GetCurrentTimestamp()
	result := FormatTimeUnix(timestamp)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "-") // Has date separators
	assert.Contains(t, result, ":") // Has time separators
}

func TestParseTime_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		timeStr   string
		expectedH int
		expectedM int
		expectedS int
	}{
		{
			name:      "midnight",
			timeStr:   "00:00:00",
			expectedH: 0,
			expectedM: 0,
			expectedS: 0,
		},
		{
			name:      "end of day",
			timeStr:   "23:59:59",
			expectedH: 23,
			expectedM: 59,
			expectedS: 59,
		},
		{
			name:      "noon",
			timeStr:   "12:00:00",
			expectedH: 12,
			expectedM: 0,
			expectedS: 0,
		},
		{
			name:      "with spaces",
			timeStr:   " 14 : 30 : 45 ",
			expectedH: 14,
			expectedM: 30,
			expectedS: 45,
		},
		{
			name:      "regular time",
			timeStr:   "14:30:45",
			expectedH: 14,
			expectedM: 30,
			expectedS: 45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h, m, s, err := ParseTime(tt.timeStr)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedH, h)
			assert.Equal(t, tt.expectedM, m)
			assert.Equal(t, tt.expectedS, s)
		})
	}
}

func TestParseTime_InvalidFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		timeStr string
	}{
		{
			name:    "missing seconds",
			timeStr: "14:30",
		},
		{
			name:    "missing minutes and seconds",
			timeStr: "14",
		},
		{
			name:    "too many parts",
			timeStr: "14:30:45:00",
		},
		{
			name:    "empty string",
			timeStr: "",
		},
		{
			name:    "wrong delimiter",
			timeStr: "14-30-45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h, m, s, err := ParseTime(tt.timeStr)
			assert.Error(t, err)
			assert.Equal(t, 0, h)
			assert.Equal(t, 0, m)
			assert.Equal(t, 0, s)
		})
	}
}

func TestParseTime_InvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		timeStr string
	}{
		{
			name:    "invalid hour",
			timeStr: "24:00:00",
		},
		{
			name:    "invalid hour text",
			timeStr: "ab:30:45",
		},
		{
			name:    "invalid minute",
			timeStr: "14:60:00",
		},
		{
			name:    "invalid minute text",
			timeStr: "14:cd:45",
		},
		{
			name:    "invalid second",
			timeStr: "14:30:60",
		},
		{
			name:    "invalid second text",
			timeStr: "14:30:ef",
		},
		{
			name:    "negative hour",
			timeStr: "-1:30:45",
		},
		{
			name:    "negative minute",
			timeStr: "14:-1:45",
		},
		{
			name:    "negative second",
			timeStr: "14:30:-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h, m, s, err := ParseTime(tt.timeStr)
			assert.Error(t, err)
			assert.Equal(t, 0, h)
			assert.Equal(t, 0, m)
			assert.Equal(t, 0, s)
		})
	}
}
