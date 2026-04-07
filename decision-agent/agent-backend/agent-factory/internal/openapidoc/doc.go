package openapidoc

import (
	"regexp"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	agentFactoryBasePath = "/api/agent-factory"
	scalarCDNURL         = "https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.34.6"
	apiKeySecurityName   = "ApiKeyAuth"
	bearerSecurityName   = "BearerAuth"
)

var pathParamPattern = regexp.MustCompile(`\{([^}/]+)\}`)

// BuildOptions 描述一次文档构建所需的输入文件和合并策略。
type BuildOptions struct {
	SwaggerPath           string
	OverlayPath           string
	BaselinePath          string
	ApplyBaselineFallback bool
}

// BuildArtifacts 汇总构建过程中产生的中间文档、最终文档与可落盘产物。
type BuildArtifacts struct {
	GeneratedDoc  *openapi3.T
	FinalDoc      *openapi3.T
	CompareReport string
	JSON          []byte
	YAML          []byte
	HTML          []byte
}
