package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitOffset_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	l := LimitOffset{}

	errMsgMap := l.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"limit"的值必须是数字`, errMsgMap["Limit.numeric"])
	assert.Equal(t, `"limit"的值不能大于1000`, errMsgMap["Limit.max"])
}

func TestLimitOffset_GetLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		limit          int
		expectedResult int
	}{
		{
			name:           "zero limit returns default 10",
			limit:          0,
			expectedResult: 10,
		},
		{
			name:           "non-zero limit returns itself",
			limit:          50,
			expectedResult: 50,
		},
		{
			name:           "limit 1 returns 1",
			limit:          1,
			expectedResult: 1,
		},
		{
			name:           "limit 1000 returns 1000",
			limit:          1000,
			expectedResult: 1000,
		},
		{
			name:           "negative limit returns negative (no validation in getter)",
			limit:          -10,
			expectedResult: -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l := LimitOffset{Limit: tt.limit}
			result := l.GetLimit()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestLimitOffset_GetOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		offset         int
		expectedResult int
	}{
		{
			name:           "zero offset",
			offset:         0,
			expectedResult: 0,
		},
		{
			name:           "positive offset",
			offset:         10,
			expectedResult: 10,
		},
		{
			name:           "negative offset",
			offset:         -5,
			expectedResult: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			l := LimitOffset{Offset: tt.offset}
			result := l.GetOffset()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestLimitOffset_DefaultValues(t *testing.T) {
	t.Parallel()

	l := LimitOffset{}

	assert.Equal(t, 0, l.Limit)
	assert.Equal(t, 0, l.Offset)
	assert.Equal(t, 10, l.GetLimit())
}

func TestLimitOffset_WithValues(t *testing.T) {
	t.Parallel()

	l := LimitOffset{
		Limit:  50,
		Offset: 100,
	}

	assert.Equal(t, 50, l.Limit)
	assert.Equal(t, 100, l.Offset)
	assert.Equal(t, 50, l.GetLimit())
	assert.Equal(t, 100, l.GetOffset())
}
