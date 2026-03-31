package cutil

import (
	"encoding/json"
	"fmt"

	"github.com/imdario/mergo"
)

// MergeJSONStrings merges two JSON strings into one.
// json2's content will overwrite json1's content if they share the same keys.
// It supports nested structures and will recursively merge objects.
// Arrays will be directly merged.
func MergeJSONStrings(jsonStr1, jsonStr2 string) (string, error) {
	var map1, map2 map[string]interface{}

	// Parse the JSON strings into maps
	if err := json.Unmarshal([]byte(jsonStr1), &map1); err != nil {
		return "", fmt.Errorf("failed to parse jsonStr1: %w", err)
	}

	if err := json.Unmarshal([]byte(jsonStr2), &map2); err != nil {
		return "", fmt.Errorf("failed to parse jsonStr2: %w", err)
	}

	// Merge map2 into map1
	if err := mergo.Merge(&map1, map2, mergo.WithOverride); err != nil {
		return "", fmt.Errorf("failed to merge maps: %w", err)
	}

	// Convert the merged map back to a JSON string with sorted keys
	mergedJSON, err := json.Marshal(map1)
	if err != nil {
		return "", fmt.Errorf("failed to marshal merged map: %w", err)
	}

	return string(mergedJSON), nil
}

// MergeMapInterface merges a map[string]interface{} with an interface{}.
// It uses the mergo library which allows overriding existing values in the map.
func MergeMapInterface(m map[string]interface{}, i interface{}) error {
	// Use mergo.Merge to merge the interface i into the map m
	// WithOverride will allow i to override m's values.
	// 1. 将interface转换为map[string]interface{}
	var map2 map[string]interface{}

	jsonBytes, err := json.Marshal(i)
	if err != nil {
		return fmt.Errorf("failed to marshal interface into json: %w", err)
	}

	if err = json.Unmarshal(jsonBytes, &map2); err != nil {
		return fmt.Errorf("failed to parse jsonStr2: %w", err)
	}

	// 2. 合并两个map
	if err = mergo.Merge(&m, map2, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge interface into map: %w", err)
	}

	return nil
}
