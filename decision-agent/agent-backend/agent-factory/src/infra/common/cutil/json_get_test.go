package cutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitJSONArray(t *testing.T) {
	t.Parallel()

	jsonData := `[1,2,3,4,5,6,7,8,9,10]`
	expected := []string{
		`[1,2,3]`,
		`[4,5,6]`,
		`[7,8,9]`,
		`[10]`,
	}

	result := SplitJSONArray(jsonData, 3)

	assert.Equal(t, expected, result, "The arrays should be split correctly into chunks of size 3.")
}

func TestSplitJSONArrayWithStrings(t *testing.T) {
	t.Parallel()

	jsonData := `["a2ss","b","c","d","e","f","g","h","i","j"]`
	expected := []string{
		`["a2ss","b","c"]`,
		`["d","e","f"]`,
		`["g","h","i"]`,
		`["j"]`,
	}

	result := SplitJSONArray(jsonData, 3)

	assert.Equal(t, expected, result, "The arrays should be split correctly into chunks of size 3.")
}

func TestSplitJSONArrayBys(t *testing.T) {
	t.Parallel()

	jsonData := []byte(`["a2ss","b","c","d","e","f","g","h","i","j"]`)
	expected := [][]byte{
		[]byte(`["a2ss","b","c"]`),
		[]byte(`["d","e","f"]`),
		[]byte(`["g","h","i"]`),
		[]byte(`["j"]`),
	}

	result := SplitJSONArrayBys(jsonData, 3)

	assert.Equal(t, expected, result, "The arrays should be split correctly into chunks of size 3.")
}
