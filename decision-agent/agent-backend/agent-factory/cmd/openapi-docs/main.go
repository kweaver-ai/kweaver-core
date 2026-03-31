package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kweaver-ai/decision-agent/agent-factory/internal/openapidoc"
	pkgerrors "github.com/pkg/errors"
)

const (
	defaultSwaggerPath  = "docs/swagger.json"
	defaultOverlayPath  = "docs/api/overlay.yaml"
	defaultBaselinePath = "docs/api/baseline/agent-factory.json"
	defaultOutJSONPath  = "docs/api/agent-factory.json"
	defaultOutYAMLPath  = "docs/api/agent-factory.yaml"
	defaultOutHTMLPath  = "docs/api/agent-factory.html"
	defaultReportPath   = "test_out/openapi_compare_report.md"
	defaultExpectPaths  = 54
	defaultExpectOps    = 68
)

func main() {
	if len(os.Args) < 2 {
		exitWithError(usageError())
	}

	var err error

	switch os.Args[1] {
	case "generate":
		err = runGenerate(os.Args[2:])
	case "compare":
		err = runCompare(os.Args[2:])
	case "validate":
		err = runValidate(os.Args[2:])
	default:
		err = usageError()
	}

	if err != nil {
		exitWithError(err)
	}
}

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

func runCompare(args []string) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	swaggerPath := fs.String("swagger", defaultSwaggerPath, "Swagger 2.0 input path")
	overlayPath := fs.String("overlay", defaultOverlayPath, "OpenAPI overlay path")
	baselinePath := fs.String("baseline", defaultBaselinePath, "Baseline OpenAPI path")
	reportPath := fs.String("out", defaultReportPath, "Compare report output path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	artifacts, err := openapidoc.BuildArtifactsFromFiles(context.Background(), openapidoc.BuildOptions{
		SwaggerPath:           *swaggerPath,
		OverlayPath:           optionalPath(*overlayPath),
		BaselinePath:          *baselinePath,
		ApplyBaselineFallback: false,
	})
	if err != nil {
		return err
	}

	if err := openapidoc.WriteFile(*reportPath, []byte(artifacts.CompareReport)); err != nil {
		return pkgerrors.Wrap(err, "write compare report")
	}

	generatedPaths, generatedOps := openapidoc.CountPathsAndOperations(artifacts.GeneratedDoc)
	fmt.Printf("generated raw spec: %d paths / %d operations\n", generatedPaths, generatedOps)
	fmt.Printf("wrote %s\n", *reportPath)

	return nil
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	inputPath := fs.String("input", defaultOutJSONPath, "OpenAPI document to validate")
	htmlPath := fs.String("html", defaultOutHTMLPath, "Scalar HTML document to validate")
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
		if !strings.Contains(htmlContent, "@scalar/api-reference") || !strings.Contains(htmlContent, "openapi-document") {
			return pkgerrors.New("scalar html is missing expected embedded reference markup")
		}
	}

	fmt.Printf("validated %s: %d paths / %d operations\n", *inputPath, paths, ops)

	return nil
}

func usageError() error {
	return pkgerrors.New(`usage: go run ./cmd/openapi-docs <generate|compare|validate> [flags]`)
}

func optionalPath(path string) string {
	if strings.TrimSpace(path) == "" || path == "-" {
		return ""
	}

	return path
}

func exitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
