// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package helpers provides logical catalog specific helpers.
package helpers

import (
	cataloghelpers "vega-backend-tests/at/catalog/helpers"
)

// BuildLogicalCatalogPayload 构建logical catalog创建payload
func BuildLogicalCatalogPayload() map[string]any {
	return map[string]any{
		"name":        cataloghelpers.GenerateUniqueName("logical-catalog"),
		"type":        "logical",
		"description": "逻辑Catalog测试",
		"tags":        []string{"test", "logical"},
	}
}

// BuildLogicalCatalogPayloadWithName 构建带指定名称的logical catalog
func BuildLogicalCatalogPayloadWithName(name string) map[string]any {
	return map[string]any{
		"name":        name,
		"type":        "logical",
		"description": "逻辑Catalog测试",
		"tags":        []string{"test", "logical"},
	}
}

// BuildFullLogicalCatalogPayload 构建完整字段的logical catalog
func BuildFullLogicalCatalogPayload() map[string]any {
	return map[string]any{
		"name":        cataloghelpers.GenerateUniqueName("full-logical-catalog"),
		"type":        "logical",
		"description": "完整的逻辑Catalog测试，包含所有可选字段",
		"tags":        []string{"test", "logical", "full"},
	}
}
