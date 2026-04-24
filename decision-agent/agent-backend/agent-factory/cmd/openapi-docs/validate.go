package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	publicUIDirPath := fs.String("public-ui-dir", defaultPublicUIDirPath, "Public UI assets directory to validate (optional)")
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

		if err := validatePublicStaticHTML("scalar html", string(htmlData), publicScalarScriptRef()); err != nil {
			return err
		}
	}

	if optionalPath(*redocHTMLPath) != "" {
		htmlData, err := os.ReadFile(*redocHTMLPath)
		if err != nil {
			return pkgerrors.Wrap(err, "read redoc html")
		}

		if err := validatePublicStaticHTML("redoc html", string(htmlData), publicRedocScriptRef()); err != nil {
			return err
		}
	}

	if optionalPath(*runtimeHTMLPath) != "" {
		htmlData, err := os.ReadFile(*runtimeHTMLPath)
		if err != nil {
			return pkgerrors.Wrap(err, "read runtime scalar html")
		}

		if err := validateRuntimeStaticHTML("runtime scalar html", string(htmlData), "ui/scalar-api-reference.js"); err != nil {
			return err
		}
	}

	if optionalPath(*runtimeRedocHTMLPath) != "" {
		htmlData, err := os.ReadFile(*runtimeRedocHTMLPath)
		if err != nil {
			return pkgerrors.Wrap(err, "read runtime redoc html")
		}

		if err := validateRuntimeStaticHTML("runtime redoc html", string(htmlData), "ui/redoc.standalone.js"); err != nil {
			return err
		}
	}

	if err := validateUIDirectory(*publicUIDirPath, false); err != nil {
		return err
	}

	if err := validateUIDirectory(*runtimeUIDirPath, true); err != nil {
		return err
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

func validatePublicStaticHTML(label string, htmlContent string, scriptRef string) error {
	if !strings.Contains(htmlContent, "openapi-document") || !strings.Contains(htmlContent, scriptRef) {
		return pkgerrors.Errorf("%s is missing expected CDN-backed reference markup", label)
	}

	if strings.Contains(htmlContent, "ui/scalar-api-reference.js") || strings.Contains(htmlContent, "ui/redoc.standalone.js") {
		return pkgerrors.Errorf("%s still references local UI assets", label)
	}

	if containsUnsupportedExternalUIReference(htmlContent) {
		return pkgerrors.Errorf("%s references unsupported external UI assets", label)
	}

	return nil
}

func validateRuntimeStaticHTML(label string, htmlContent string, scriptRef string) error {
	if !strings.Contains(htmlContent, "openapi-document") || !strings.Contains(htmlContent, scriptRef) {
		return pkgerrors.Errorf("%s is missing expected embedded reference markup", label)
	}

	if containsAnyExternalUIReference(htmlContent) {
		return pkgerrors.Errorf("%s still references external UI assets", label)
	}

	return nil
}

func validateUIDirectory(dirPath string, required bool) error {
	if optionalPath(dirPath) == "" {
		return nil
	}

	requiredFiles := []string{"scalar-api-reference.js", "redoc.standalone.js"}
	for _, name := range requiredFiles {
		fullPath := filepath.Join(dirPath, name)
		info, err := os.Stat(fullPath)
		if err != nil {
			return pkgerrors.Wrapf(err, "stat ui asset %s", fullPath)
		}
		if required && info.Size() == 0 {
			return pkgerrors.Errorf("ui asset %s is empty", fullPath)
		}
	}

	return nil
}

func containsUnsupportedExternalUIReference(htmlContent string) bool {
	return strings.Contains(htmlContent, "cdn.jsdelivr.net") ||
		strings.Contains(htmlContent, "cdn.redocly.com") ||
		strings.Contains(htmlContent, "fonts.googleapis.com") ||
		strings.Contains(htmlContent, "fonts.gstatic.com")
}

func containsAnyExternalUIReference(htmlContent string) bool {
	return strings.Contains(htmlContent, "cdn.jsdmirror.com") || containsUnsupportedExternalUIReference(htmlContent)
}

func publicScalarScriptRef() string {
	return "https://cdn.jsdmirror.com/npm/@scalar/api-reference@1.34.6/dist/browser/standalone.js"
}

func publicRedocScriptRef() string {
	return "https://cdn.jsdmirror.com/npm/redoc@2.5.1/bundles/redoc.standalone.js"
}
