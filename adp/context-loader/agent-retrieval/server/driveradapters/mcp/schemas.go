// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mcp

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed schemas/*.json
var schemasFS embed.FS

// ToolMeta defines tool metadata (name, description).
type ToolMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// loadToolMeta loads tool metadata (name, description) from schemas/tools_meta.json.
func loadToolMeta(toolKey string) (name, description string) {
	data, err := schemasFS.ReadFile("schemas/tools_meta.json")
	if err != nil {
		panic("cannot read tools_meta.json: " + err.Error())
	}
	var meta map[string]ToolMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		panic("invalid tools_meta.json: " + err.Error())
	}
	t, ok := meta[toolKey]
	if !ok {
		panic("tool meta not found: " + toolKey)
	}
	return t.Name, t.Description
}

// toolSchemaFile defines the structure of a merged tool schema JSON file.
type toolSchemaFile struct {
	InputSchema  json.RawMessage `json:"input_schema"`
	OutputSchema json.RawMessage `json:"output_schema"`
}

// loadToolSchemas loads input and output schema for a tool from its merged JSON file.
// File: schemas/<toolKey>.json, containing input_schema and output_schema keys.
func loadToolSchemas(toolKey string) (input, output json.RawMessage) {
	path := fmt.Sprintf("schemas/%s.json", toolKey)
	data, err := schemasFS.ReadFile(path)
	if err != nil {
		panic("cannot read " + path + ": " + err.Error())
	}
	var wrapper toolSchemaFile
	if err := json.Unmarshal(data, &wrapper); err != nil {
		panic("invalid " + path + ": " + err.Error())
	}
	if len(wrapper.InputSchema) == 0 {
		panic(path + ": missing input_schema")
	}
	return wrapper.InputSchema, wrapper.OutputSchema
}
