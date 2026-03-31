package cutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// JSONNormalizer normalizes JSON by sorting the keys alphabetically.
type JSONNormalizer map[string]interface{}

// MarshalJSON implements the json.Marshaler interface for JSONNormalizer
// to sort the keys before marshaling.
func (jn JSONNormalizer) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	keys := make([]string, 0, len(jn))
	for key := range jn {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	buf.WriteString("{")

	for i, key := range keys {
		if i > 0 {
			buf.WriteString(",")
		}

		buf.WriteString(fmt.Sprintf("\"%s\":", key))

		value, err := json.Marshal(jn[key])
		if err != nil {
			return nil, err
		}

		buf.Write(value)
	}

	buf.WriteString("}")

	return buf.Bytes(), nil
}

func JSONStrCompare(jsonStr1, jsonStr2 string) (bool, error) {
	var map1, map2 map[string]interface{}

	// Parse the JSON strings into maps
	if err := json.Unmarshal([]byte(jsonStr1), &map1); err != nil {
		return false, fmt.Errorf("failed to parse jsonStr1: %w", err)
	}

	if err := json.Unmarshal([]byte(jsonStr2), &map2); err != nil {
		return false, fmt.Errorf("failed to parse jsonStr2: %w", err)
	}

	// Compare map1 and map2

	normalized := JSONNormalizer(map1)
	normalized2 := JSONNormalizer(map2)

	// Convert the merged map back to a JSON string with sorted keys
	normalizedJSON, err := json.Marshal(normalized)
	if err != nil {
		return false, fmt.Errorf("failed to marshal merged map: %w", err)
	}

	normalizedJSON2, err := json.Marshal(normalized2)
	if err != nil {
		return false, fmt.Errorf("failed to marshal merged map: %w", err)
	}

	return string(normalizedJSON) == string(normalizedJSON2), nil
}
