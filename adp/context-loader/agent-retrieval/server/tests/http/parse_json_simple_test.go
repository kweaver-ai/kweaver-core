// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"
)

// QueryStrategy represents the top-level structure
type QueryStrategy struct {
	Strategies []Strategy `json:"query_strategy"`
}

// Strategy represents each strategy in the array
type Strategy struct {
	StrategyType string  `json:"strategy_type"`
	Confidence   float64 `json:"confidence"`
	Filter       Filter  `json:"filter"`
}

// Filter represents the filter object
type Filter struct {
	ConceptType string      `json:"concept_type"`
	Conditions  []Condition `json:"conditions"`
}

// Condition represents each condition in the filter
type Condition struct {
	Field     string `json:"field"`
	Operation string `json:"operation"`
	Value     string `json:"value"`
}

// parseJSONString parses a JSON string that may contain escape characters
func parseJSONString(jsonStr string) (QueryStrategy, error) {
	var queryStrategy QueryStrategy

	start := strings.Index(jsonStr, "{")
	end := strings.LastIndex(jsonStr, "}")
	if start == -1 || end == -1 {
		return queryStrategy, fmt.Errorf("invalid JSON format")
	}

	jsonStr = jsonStr[start : end+1]

	// If the string contains escape characters, unescape them
	if strings.Contains(jsonStr, "\\n") || strings.Contains(jsonStr, "\\\"") {
		jsonStr = strings.ReplaceAll(jsonStr, "\\n", "\n")
		jsonStr = strings.ReplaceAll(jsonStr, "\\\"", "\"")
	}

	err := json.Unmarshal([]byte(jsonStr), &queryStrategy)
	return queryStrategy, err
}

func TestParseJSONString(_ *testing.T) {
	// 测试字符串包含 JSON 块，需要提取并解析
	jsonStr := "```json\n" +
		`{"query_strategy":[{"strategy_type":"concept_discovery","confidence":0.9,` +
		`"filter":{"concept_type":"object_type","conditions":[` +
		`{"field":"name","operation":"like","value":"技能特长"}]}}]}` +
		"\n```"
	queryStrategy, err := parseJSONString(jsonStr)
	if err != nil {
		log.Fatal("Error unmarshaling JSON:", err)
	}

	// Print the parsed data
	fmt.Printf("Parsed Data: %+v\n", queryStrategy)

	// Access individual fields
	if len(queryStrategy.Strategies) > 0 {
		strategy := queryStrategy.Strategies[0]
		fmt.Printf("Strategy Type: %s\n", strategy.StrategyType)
		fmt.Printf("Confidence: %.1f\n", strategy.Confidence)
		fmt.Printf("Concept Type: %s\n", strategy.Filter.ConceptType)

		if len(strategy.Filter.Conditions) > 0 {
			condition := strategy.Filter.Conditions[0]
			fmt.Printf("Condition Field: %s\n", condition.Field)
			fmt.Printf("Condition Operation: %s\n", condition.Operation)
			fmt.Printf("Condition Value: %s\n", condition.Value)
		}
	}
}
