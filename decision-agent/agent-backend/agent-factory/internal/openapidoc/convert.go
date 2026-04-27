package openapidoc

import (
	"encoding/json"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	pkgerrors "github.com/pkg/errors"
)

// ConvertSwagger2ToOpenAPI3 将 swagger2 JSON 转为 OpenAPI 3 文档，并补齐基础结构。
func ConvertSwagger2ToOpenAPI3(data []byte) (*openapi3.T, error) {
	var doc2 openapi2.T
	if err := json.Unmarshal(data, &doc2); err != nil {
		return nil, pkgerrors.Wrap(err, "unmarshal swagger2 json")
	}

	doc3, err := openapi2conv.ToV3(&doc2)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "convert swagger2 document")
	}

	doc3.OpenAPI = "3.0.2"
	if doc3.Paths == nil {
		doc3.Paths = openapi3.NewPaths()
	}

	if doc3.Components == nil {
		components := openapi3.NewComponents()
		doc3.Components = &components
	}

	return doc3, nil
}

// RewriteAgentFactoryPaths 为所有路径统一补上 /api/agent-factory 前缀。
func RewriteAgentFactoryPaths(doc *openapi3.T) {
	if doc == nil || doc.Paths == nil {
		return
	}

	rewritten := openapi3.NewPathsWithCapacity(doc.Paths.Len())

	for _, path := range doc.Paths.InMatchingOrder() {
		rewritten.Set(prefixAgentFactoryPath(path), doc.Paths.Value(path))
	}

	doc.Paths = rewritten
}

// prefixAgentFactoryPath 计算单条路径的最终展示路径，并避免重复加前缀。
func prefixAgentFactoryPath(path string) string {
	if path == "" {
		return agentFactoryBasePath
	}

	if strings.HasPrefix(path, agentFactoryBasePath) {
		return path
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return agentFactoryBasePath + path
}
