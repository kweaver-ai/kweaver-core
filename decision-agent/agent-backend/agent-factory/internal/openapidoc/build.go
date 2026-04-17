package openapidoc

import (
	"context"
	"os"

	pkgerrors "github.com/pkg/errors"
)

// BuildArtifactsFromFiles 按“读取 Swagger -> 转 OpenAPI -> 合并 overlay/baseline -> 归一化 -> 校验 -> 序列化”的流程构建全部文档产物。
func BuildArtifactsFromFiles(ctx context.Context, opts BuildOptions) (*BuildArtifacts, error) {
	swaggerData, err := os.ReadFile(opts.SwaggerPath)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read swagger input")
	}

	generatedDoc, err := ConvertSwagger2ToOpenAPI3(swaggerData)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "convert swagger2 to openapi3")
	}

	RewriteAgentFactoryPaths(generatedDoc)

	if opts.OverlayPath != "" {
		overlayData, err := os.ReadFile(opts.OverlayPath)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "read overlay")
		}

		if err := MergeOverlay(generatedDoc, overlayData); err != nil {
			return nil, pkgerrors.Wrap(err, "merge overlay")
		}
	}

	NormalizeSecurity(generatedDoc)

	finalDoc, err := CloneOpenAPIDoc(generatedDoc)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "clone generated doc")
	}

	var compareReport string

	if opts.BaselinePath != "" {
		baselineDoc, err := LoadOpenAPIDocFile(opts.BaselinePath)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "load baseline")
		}

		compareReport = BuildComparisonReport(generatedDoc, baselineDoc)

		if opts.ApplyBaselineFallback {
			if err := MergeMissingFromBaseline(finalDoc, baselineDoc); err != nil {
				return nil, pkgerrors.Wrap(err, "merge baseline fallback")
			}
		}
	}

	NormalizePathParameters(finalDoc)
	NormalizeOperationIDs(finalDoc)
	NormalizeSecurity(finalDoc)

	if err := ValidateOpenAPI(ctx, finalDoc); err != nil {
		return nil, pkgerrors.Wrap(err, "validate final openapi")
	}

	finalJSON, err := MarshalOpenAPIJSON(finalDoc)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "marshal final openapi json")
	}

	finalYAML, err := MarshalOpenAPIYAML(finalDoc)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "marshal final openapi yaml")
	}

	return &BuildArtifacts{
		GeneratedDoc:  generatedDoc,
		FinalDoc:      finalDoc,
		CompareReport: compareReport,
		JSON:          finalJSON,
		YAML:          finalYAML,
		HTML:          []byte(RenderScalarStaticHTML(finalJSON)),
		RedocHTML:     []byte(RenderRedocStaticHTML(finalJSON)),
	}, nil
}
