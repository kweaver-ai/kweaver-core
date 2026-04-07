// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

var half = 0.5

func GetSQEmbeddingVector(origin []float64) (target []int64) {
	if len(origin) == 0 {
		return []int64{}
	}
	maxVal := origin[0]
	minVal := origin[1]
	for i := range origin {
		if origin[i] > maxVal {
			maxVal = origin[i]
		}
		if origin[i] < minVal {
			minVal = origin[i]
		}
	}
	minVal = -minVal
	target = make([]int64, 0)
	for i := range origin {
		if origin[i] > 0 {
			target = append(target, GetPositiveNumber(maxVal, 0, origin[i]))
		} else if origin[i] < 0 {
			target = append(target, GetPositiveNumber(minVal, 0, origin[i]))
		} else {
			target = append(target, 0)
		}
	}
	return target
}

func GetPositiveNumber(maxValue, minValue, val float64) int64 {
	B := 127
	val = (val - minValue) / (maxValue - minValue)
	val *= float64(B)
	intPart := int64(val)
	fracPart := val - float64(intPart)
	if fracPart > half {
		return intPart + 1
	}
	return intPart
}
