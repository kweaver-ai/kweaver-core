package cutil

import (
	"strings"

	"github.com/tidwall/gjson"
)

func SplitJSONArray(json string, size int) []string {
	var result []string
	// Using gjson to parse the JSON array
	values := gjson.Parse(json).Array()
	for i := 0; i < len(values); i += size {
		end := i + size
		if end > len(values) {
			end = len(values)
		}

		subArray := values[i:end]

		var elements []string

		for _, v := range subArray {
			elements = append(elements, v.Raw)
		}

		result = append(result, "["+strings.Join(elements, ",")+"]")
	}

	return result
}

func SplitJSONArrayBys(json []byte, size int) [][]byte {
	var result [][]byte

	rs := SplitJSONArray(string(json), size)

	for _, v := range rs {
		result = append(result, []byte(v))
	}

	return result
}
