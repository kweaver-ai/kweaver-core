package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/internal/openapidoc"
	pkgerrors "github.com/pkg/errors"
)

// runValidate 校验生成后的 OpenAPI 文档是否合法，并检查路径数、操作数与静态 HTML 标记。
func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	inputPath := fs.String("input", defaultOutJSONPath, "OpenAPI document to validate")
	yamlPath := fs.String("yaml", defaultOutYAMLPath, "Public OpenAPI YAML document to validate")
	htmlPath := fs.String("html", defaultOutHTMLPath, "Scalar HTML document to validate")
	redocHTMLPath := fs.String("redoc-html", defaultOutRedocHTMLPath, "Redoc HTML document to validate")
	publicFaviconPath := fs.String("public-favicon", defaultPublicFaviconPath, "Public favicon document to validate")
	publicUIDirPath := fs.String("public-ui-dir", defaultPublicUIDirPath, "Public UI assets directory to validate")
	runtimeJSONPath := fs.String("runtime-json", defaultRuntimeJSONPath, "Runtime OpenAPI JSON document to compare")
	runtimeYAMLPath := fs.String("runtime-yaml", defaultRuntimeYAMLPath, "Runtime OpenAPI YAML document to compare")
	runtimeHTMLPath := fs.String("runtime-html", defaultRuntimeHTMLPath, "Runtime Scalar HTML document to compare")
	runtimeRedocHTMLPath := fs.String("runtime-redoc-html", defaultRuntimeRedocHTMLPath, "Runtime Redoc HTML document to compare")
	runtimeFaviconPath := fs.String("runtime-favicon", defaultRuntimeFaviconPath, "Runtime favicon document to compare")
	runtimeUIDirPath := fs.String("runtime-ui-dir", defaultRuntimeUIDirPath, "Runtime UI assets directory to compare")
	expectPaths := fs.Int("expect-paths", defaultExpectPaths, "Expected path count (0 to skip)")
	expectOps := fs.Int("expect-ops", defaultExpectOps, "Expected operation count (0 to skip)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	doc, err := openapidoc.LoadOpenAPIDocFile(*inputPath)
	if err != nil {
		return err
	}

	openapidoc.NormalizePathParameters(doc)
	openapidoc.NormalizeOperationIDs(doc)

	if err := openapidoc.ValidateOpenAPI(context.Background(), doc); err != nil {
		return err
	}

	paths, ops := openapidoc.CountPathsAndOperations(doc)
	if *expectPaths > 0 && paths != *expectPaths {
		return pkgerrors.Errorf("unexpected path count: got %d want %d", paths, *expectPaths)
	}

	if *expectOps > 0 && ops != *expectOps {
		return pkgerrors.Errorf("unexpected operation count: got %d want %d", ops, *expectOps)
	}

	if optionalPath(*htmlPath) != "" {
		htmlData, err := os.ReadFile(*htmlPath)
		if err != nil {
			return pkgerrors.Wrap(err, "read scalar html")
		}

		htmlContent := string(htmlData)
		if !strings.Contains(htmlContent, "ui/scalar-api-reference.js") || !strings.Contains(htmlContent, "openapi-document") {
			return pkgerrors.New("scalar html is missing expected embedded reference markup")
		}
		if containsExternalUIReference(htmlContent) {
			return pkgerrors.New("scalar html still references external UI assets")
		}
	}

	if optionalPath(*redocHTMLPath) != "" {
		htmlData, err := os.ReadFile(*redocHTMLPath)
		if err != nil {
			return pkgerrors.Wrap(err, "read redoc html")
		}

		htmlContent := string(htmlData)
		if !strings.Contains(htmlContent, "ui/redoc.standalone.js") || !strings.Contains(htmlContent, "openapi-document") {
			return pkgerrors.New("redoc html is missing expected embedded reference markup")
		}
		if containsExternalUIReference(htmlContent) {
			return pkgerrors.New("redoc html still references external UI assets")
		}
	}

	if err := validateMirroredArtifacts(mirroredArtifactPaths{
		PublicJSONPath:       *inputPath,
		PublicYAMLPath:       *yamlPath,
		PublicHTMLPath:       *htmlPath,
		PublicRedocHTMLPath:  *redocHTMLPath,
		PublicFaviconPath:    *publicFaviconPath,
		PublicUIDirPath:      *publicUIDirPath,
		RuntimeJSONPath:      *runtimeJSONPath,
		RuntimeYAMLPath:      *runtimeYAMLPath,
		RuntimeHTMLPath:      *runtimeHTMLPath,
		RuntimeRedocHTMLPath: *runtimeRedocHTMLPath,
		RuntimeFaviconPath:   *runtimeFaviconPath,
		RuntimeUIDirPath:     *runtimeUIDirPath,
	}); err != nil {
		return err
	}

	fmt.Printf("validated %s: %d paths / %d operations\n", *inputPath, paths, ops)

	return nil
}

func containsExternalUIReference(htmlContent string) bool {
	return strings.Contains(htmlContent, "cdn.jsdelivr.net") ||
		strings.Contains(htmlContent, "cdn.redocly.com") ||
		strings.Contains(htmlContent, "fonts.googleapis.com") ||
		strings.Contains(htmlContent, "fonts.gstatic.com")
}
