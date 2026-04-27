package cutil

import "github.com/shopspring/decimal"

// DecimalSum 计算多个float64数字的精确和
func DecimalSum(numbers ...float64) decimal.Decimal {
	sum := decimal.NewFromFloat(0)
	for _, num := range numbers {
		sum = sum.Add(decimal.NewFromFloat(num))
	}

	return sum
}

func DecimalSumEqualOne(numbers ...float64) bool {
	sum := DecimalSum(numbers...)
	return sum.Equal(decimal.NewFromFloat(1))
}
