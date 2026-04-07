package cutil

import "testing"

func TestMinInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a小于b", 1, 5, 1},
		{"a大于b", 5, 1, 1},
		{"a等于b", 5, 5, 5},
		{"负数a", -5, 1, -5},
		{"负数b", 1, -5, -5},
		{"都是负数", -5, -10, -10},
		{"零值", 0, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := MinInt(tt.a, tt.b); got != tt.want {
				t.Errorf("MinInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMaxInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"a小于b", 1, 5, 5},
		{"a大于b", 5, 1, 5},
		{"a等于b", 5, 5, 5},
		{"负数a", -5, 1, 1},
		{"负数b", 1, -5, 1},
		{"都是负数", -5, -10, -5},
		{"零值", 0, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := MaxInt(tt.a, tt.b); got != tt.want {
				t.Errorf("MaxInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
