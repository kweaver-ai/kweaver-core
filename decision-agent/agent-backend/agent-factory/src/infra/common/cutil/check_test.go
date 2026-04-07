package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckInRange_Int(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    int
		min      int
		max      int
		expected bool
	}{
		{
			name:     "value in range",
			value:    5,
			min:      1,
			max:      10,
			expected: true,
		},
		{
			name:     "value at min boundary",
			value:    1,
			min:      1,
			max:      10,
			expected: true,
		},
		{
			name:     "value at max boundary",
			value:    10,
			min:      1,
			max:      10,
			expected: true,
		},
		{
			name:     "value below min",
			value:    0,
			min:      1,
			max:      10,
			expected: false,
		},
		{
			name:     "value above max",
			value:    11,
			min:      1,
			max:      10,
			expected: false,
		},
		{
			name:     "negative values",
			value:    -5,
			min:      -10,
			max:      0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CheckInRange(tt.value, tt.min, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckInRange_Float64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    float64
		min      float64
		max      float64
		expected bool
	}{
		{
			name:     "value in range",
			value:    5.5,
			min:      1.0,
			max:      10.0,
			expected: true,
		},
		{
			name:     "value at min boundary",
			value:    1.0,
			min:      1.0,
			max:      10.0,
			expected: true,
		},
		{
			name:     "value at max boundary",
			value:    10.0,
			min:      1.0,
			max:      10.0,
			expected: true,
		},
		{
			name:     "value below min",
			value:    0.9,
			min:      1.0,
			max:      10.0,
			expected: false,
		},
		{
			name:     "value above max",
			value:    10.1,
			min:      1.0,
			max:      10.0,
			expected: false,
		},
		{
			name:     "negative values",
			value:    -5.5,
			min:      -10.0,
			max:      0.0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CheckInRange(tt.value, tt.min, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckInRange_Uint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    uint
		min      uint
		max      uint
		expected bool
	}{
		{
			name:     "value in range",
			value:    5,
			min:      1,
			max:      10,
			expected: true,
		},
		{
			name:     "value at boundaries",
			value:    0,
			min:      0,
			max:      100,
			expected: true,
		},
		{
			name:     "value above max",
			value:    101,
			min:      0,
			max:      100,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CheckInRange(tt.value, tt.min, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckMin_Int(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    int
		min      int
		expected bool
	}{
		{
			name:     "value above min",
			value:    10,
			min:      5,
			expected: true,
		},
		{
			name:     "value at min",
			value:    5,
			min:      5,
			expected: true,
		},
		{
			name:     "value below min",
			value:    4,
			min:      5,
			expected: false,
		},
		{
			name:     "negative values",
			value:    -5,
			min:      -10,
			expected: true,
		},
		{
			name:     "value equal to zero",
			value:    0,
			min:      0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CheckMin(tt.value, tt.min)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckMin_Float64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    float64
		min      float64
		expected bool
	}{
		{
			name:     "value above min",
			value:    10.5,
			min:      5.0,
			expected: true,
		},
		{
			name:     "value at min",
			value:    5.0,
			min:      5.0,
			expected: true,
		},
		{
			name:     "value below min",
			value:    4.9,
			min:      5.0,
			expected: false,
		},
		{
			name:     "negative values",
			value:    -5.0,
			min:      -10.0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := CheckMin(tt.value, tt.min)
			assert.Equal(t, tt.expected, result)
		})
	}
}
