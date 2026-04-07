package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/internal/openapidoc"
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

	if err := openapidoc.WriteFile(*outJSONPath, artifacts.JSON); err != nil {
		return pkgerrors.Wrap(err, "write final json")
	}

	if err := openapidoc.WriteFile(*outYAMLPath, artifacts.YAML); err != nil {
		return pkgerrors.Wrap(err, "write final yaml")
	}

	if err := openapidoc.WriteFile(*outHTMLPath, artifacts.HTML); err != nil {
		return pkgerrors.Wrap(err, "write final html")
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
	fmt.Printf("wrote %s\nwrote %s\nwrote %s\n", *outJSONPath, *outYAMLPath, *outHTMLPath)

	if optionalPath(*reportPath) != "" && strings.TrimSpace(artifacts.CompareReport) != "" {
		fmt.Printf("wrote %s\n", *reportPath)
	}

	return nil
}
