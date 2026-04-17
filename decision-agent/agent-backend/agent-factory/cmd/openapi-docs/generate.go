package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/internal/openapidoc"
	pkgerrors "github.com/pkg/errors"
)

// runGenerate 按约定流程生成最终的 OpenAPI JSON、YAML、HTML 以及可选的对比报告。
func runGenerate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	swaggerPath := fs.String("swagger", defaultSwaggerPath, "Swagger 2.0 input path")
	overlayPath := fs.String("overlay", defaultOverlayPath, "OpenAPI overlay path")
	baselinePath := fs.String("baseline", defaultBaselinePath, "Baseline OpenAPI path for compare/fallback")
	outJSONPath := fs.String("out-json", defaultOutJSONPath, "Final OpenAPI JSON output path")
	outYAMLPath := fs.String("out-yaml", defaultOutYAMLPath, "Final OpenAPI YAML output path")
	outHTMLPath := fs.String("out-html", defaultOutHTMLPath, "Final Scalar HTML output path")
	outRedocHTMLPath := fs.String("out-redoc-html", defaultOutRedocHTMLPath, "Final Redoc HTML output path")
	outFaviconPath := fs.String("out-favicon", defaultPublicFaviconPath, "Public favicon output path")
	outUIDirPath := fs.String("out-ui-dir", defaultPublicUIDirPath, "Public UI assets output path")
	runtimeJSONPath := fs.String("runtime-json", defaultRuntimeJSONPath, "Runtime OpenAPI JSON output path")
	runtimeYAMLPath := fs.String("runtime-yaml", defaultRuntimeYAMLPath, "Runtime OpenAPI YAML output path")
	runtimeHTMLPath := fs.String("runtime-html", defaultRuntimeHTMLPath, "Runtime Scalar HTML output path")
	runtimeRedocHTMLPath := fs.String("runtime-redoc-html", defaultRuntimeRedocHTMLPath, "Runtime Redoc HTML output path")
	runtimeFaviconPath := fs.String("runtime-favicon", defaultRuntimeFaviconPath, "Runtime favicon output path")
	runtimeUIDirPath := fs.String("runtime-ui-dir", defaultRuntimeUIDirPath, "Runtime UI assets output path")
	faviconSourcePath := fs.String("favicon-source", defaultFaviconSourcePath, "Source favicon path")
	uiSourceDirPath := fs.String("ui-source-dir", defaultUISourceDirPath, "Source UI assets path")
	reportPath := fs.String("report", defaultReportPath, "Compare report output path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	artifacts, err := openapidoc.BuildArtifactsFromFiles(context.Background(), openapidoc.BuildOptions{
		SwaggerPath:           *swaggerPath,
		OverlayPath:           optionalPath(*overlayPath),
		BaselinePath:          optionalPath(*baselinePath),
		ApplyBaselineFallback: true,
	})
	if err != nil {
		return err
	}

	if err := writeGeneratedArtifacts(docOutputPaths{
		PublicJSONPath:       *outJSONPath,
		PublicYAMLPath:       *outYAMLPath,
		PublicHTMLPath:       *outHTMLPath,
		PublicRedocHTMLPath:  *outRedocHTMLPath,
		PublicFaviconPath:    *outFaviconPath,
		PublicUIDirPath:      *outUIDirPath,
		RuntimeJSONPath:      *runtimeJSONPath,
		RuntimeYAMLPath:      *runtimeYAMLPath,
		RuntimeHTMLPath:      *runtimeHTMLPath,
		RuntimeRedocHTMLPath: *runtimeRedocHTMLPath,
		RuntimeFaviconPath:   *runtimeFaviconPath,
		RuntimeUIDirPath:     *runtimeUIDirPath,
		FaviconSourcePath:    *faviconSourcePath,
		UISourceDirPath:      *uiSourceDirPath,
	}, artifacts); err != nil {
		return pkgerrors.Wrap(err, "write generated artifacts")
	}

	if report := strings.TrimSpace(artifacts.CompareReport); report != "" && optionalPath(*reportPath) != "" {
		if err := openapidoc.WriteFile(*reportPath, []byte(report)); err != nil {
			return pkgerrors.Wrap(err, "write compare report")
		}
	}

	generatedPaths, generatedOps := openapidoc.CountPathsAndOperations(artifacts.GeneratedDoc)
	finalPaths, finalOps := openapidoc.CountPathsAndOperations(artifacts.FinalDoc)

	fmt.Printf("generated raw spec: %d paths / %d operations\n", generatedPaths, generatedOps)
	fmt.Printf("generated final spec: %d paths / %d operations\n", finalPaths, finalOps)
	fmt.Printf("wrote %s\nwrote %s\nwrote %s\nwrote %s\n", *outJSONPath, *outYAMLPath, *outHTMLPath, *outRedocHTMLPath)
	fmt.Printf("wrote %s\nwrote %s\nwrote %s\nwrote %s\n", *runtimeJSONPath, *runtimeYAMLPath, *runtimeHTMLPath, *runtimeRedocHTMLPath)
	fmt.Printf("mirrored UI assets to %s and %s\n", *outUIDirPath, *runtimeUIDirPath)

	if optionalPath(*reportPath) != "" && strings.TrimSpace(artifacts.CompareReport) != "" {
		fmt.Printf("wrote %s\n", *reportPath)
	}

	return nil
}
