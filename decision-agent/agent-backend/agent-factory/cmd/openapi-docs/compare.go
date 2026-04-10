package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/internal/openapidoc"
	pkgerrors "github.com/pkg/errors"
)

// runCompare 生成“当前生成结果”与 baseline 之间的差异报告，不写出最终文档文件。
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
