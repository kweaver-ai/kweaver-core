package cutil

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestDecimalSum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		numbers  []float64
		wantStr  string
		wantSign int
	}{
		{
			name:     "两个正数",
			numbers:  []float64{1.1, 2.2},
			wantStr:  "3.3",
			wantSign: 1,
		},
		{
			name:     "多个正数",
			numbers:  []float64{1.1, 2.2, 3.3, 4.4},
			wantStr:  "11",
			wantSign: 1,
		},
		{
			name:     "包含负数",
			numbers:  []float64{5.5, -2.2, 1.1},
			wantStr:  "4.4",
			wantSign: 1,
		},
		{
			name:     "全是负数",
			numbers:  []float64{-1.1, -2.2, -3.3},
			wantStr:  "-6.6",
			wantSign: -1,
		},
		{
			name:     "单个数",
			numbers:  []float64{5.5},
			wantStr:  "5.5",
			wantSign: 1,
		},
		{
			name:     "零",
			numbers:  []float64{0},
			wantStr:  "0",
			wantSign: 0,
		},
		{
			name:     "正负抵消",
			numbers:  []float64{5.5, -5.5},
			wantStr:  "0",
			wantSign: 0,
		},
		{
			name:     "浮点精度问题",
			numbers:  []float64{0.1, 0.2},
			wantStr:  "0.3",
			wantSign: 1,
		},
		{
			name:     "空参数",
			numbers:  []float64{},
			wantStr:  "0",
			wantSign: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := DecimalSum(tt.numbers...)
			if got.Sign() != tt.wantSign {
				t.Errorf("DecimalSum(%v).Sign() = %d, want %d", tt.numbers, got.Sign(), tt.wantSign)
			}

			if got.String() != tt.wantStr {
				t.Errorf("DecimalSum(%v) = %s, want %s", tt.numbers, got.String(), tt.wantStr)
			}
		})
	}
}

func TestDecimalSumEqualOne(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		numbers []float64
		want    bool
	}{
		{
			name:    "和等于1",
			numbers: []float64{0.5, 0.5},
			want:    true,
		},
		{
			name:    "多个数和等于1",
			numbers: []float64{0.1, 0.2, 0.3, 0.4},
			want:    true,
		},
		{
			name:    "单个数等于1",
			numbers: []float64{1.0},
			want:    true,
		},
		{
			name:    "和小于1",
			numbers: []float64{0.3, 0.3},
			want:    false,
		},
		{
			name:    "和大于1",
			numbers: []float64{0.6, 0.6},
			want:    false,
		},
		{
			name:    "等于1但使用浮点精度",
			numbers: []float64{0.1, 0.2, 0.3, 0.4},
			want:    true,
		},
		{
			name:    "包含负数和等于1",
			numbers: []float64{1.5, -0.5},
			want:    true,
		},
		{
			name:    "空数组",
			numbers: []float64{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := DecimalSumEqualOne(tt.numbers...)
			if got != tt.want {
				t.Errorf("DecimalSumEqualOne(%v) = %v, want %v", tt.numbers, got, tt.want)
			}
		})
	}
}

func TestDecimalSum_Precision(t *testing.T) {
	t.Parallel()

	sum := DecimalSum(0.1, 0.2, 0.3)
	expected := decimal.NewFromFloat(0.6)

	if !sum.Equal(expected) {
		t.Errorf("DecimalSum(0.1, 0.2, 0.3) = %s, want %s", sum.String(), expected.String())
	}
}

func TestDecimalSumEqualOne_Precision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		numbers []float64
		want    bool
	}{
		{
			name:    "0.1 + 0.2 + 0.3 + 0.4 = 1.0",
			numbers: []float64{0.1, 0.2, 0.3, 0.4},
			want:    true,
		},
		{
			name:    "0.333... + 0.666... = 1.0",
			numbers: []float64{0.333333, 0.666667},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := DecimalSumEqualOne(tt.numbers...)
			if got != tt.want {
				t.Errorf("DecimalSumEqualOne(%v) = %v, want %v", tt.numbers, got, tt.want)
			}
		})
	}
}
