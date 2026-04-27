package openapidoc

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// BuildComparisonReport 生成当前文档与 baseline 之间的结构和语义差异报告。
func BuildComparisonReport(generated *openapi3.T, baseline *openapi3.T) string {
	generatedPaths, generatedOps := CountPathsAndOperations(generated)
	baselinePaths, baselineOps := CountPathsAndOperations(baseline)

	generatedPathSet := pathOperationSet(generated)
	baselinePathSet := pathOperationSet(baseline)

	missingPaths := difference(sortedMapKeys(baselinePathSet), sortedMapKeys(generatedPathSet))
	extraPaths := difference(sortedMapKeys(generatedPathSet), sortedMapKeys(baselinePathSet))
	missingOps := diffOperations(baselinePathSet, generatedPathSet)
	extraOps := diffOperations(generatedPathSet, baselinePathSet)
	summaryMismatches := diffOperationSummaries(generated, baseline)
	tagMismatches := diffOperationTags(generated, baseline)
	securityMismatches := diffOperationSecurity(generated, baseline)
	requestBodyMismatches := diffOperationRequestBodies(generated, baseline)
	responseCodeMismatches := diffOperationResponseCodes(generated, baseline)

	var builder strings.Builder

	builder.WriteString("# OpenAPI Compare Report\n\n")
	builder.WriteString(fmt.Sprintf("Generated paths: %d\n", generatedPaths))
	builder.WriteString(fmt.Sprintf("Generated operations: %d\n", generatedOps))
	builder.WriteString(fmt.Sprintf("Baseline paths: %d\n", baselinePaths))
	builder.WriteString(fmt.Sprintf("Baseline operations: %d\n\n", baselineOps))

	writeMarkdownList(&builder, "Missing paths in generated", missingPaths)
	writeMarkdownList(&builder, "Extra paths in generated", extraPaths)
	writeMarkdownList(&builder, "Missing operations in generated", missingOps)
	writeMarkdownList(&builder, "Extra operations in generated", extraOps)
	writeMarkdownList(&builder, "Summary mismatches", summaryMismatches)
	writeMarkdownList(&builder, "Tag mismatches", tagMismatches)
	writeMarkdownList(&builder, "Security mismatches", securityMismatches)
	writeMarkdownList(&builder, "Request body mismatches", requestBodyMismatches)
	writeMarkdownList(&builder, "Response code mismatches", responseCodeMismatches)

	return builder.String()
}

// diffOperationSummaries 对比同一接口在 generated 与 baseline 中的 summary 是否一致。
func diffOperationSummaries(generated *openapi3.T, baseline *openapi3.T) []string {
	return diffOperationField(generated, baseline, func(path string, method string, generatedOp *openapi3.Operation, baselineOp *openapi3.Operation) string {
		if generatedOp == nil || baselineOp == nil {
			return ""
		}

		if strings.TrimSpace(generatedOp.Summary) == strings.TrimSpace(baselineOp.Summary) {
			return ""
		}

		return fmt.Sprintf(
			`%s %s :: generated=%q baseline=%q`,
			method,
			path,
			strings.TrimSpace(generatedOp.Summary),
			strings.TrimSpace(baselineOp.Summary),
		)
	})
}

// diffOperationTags 对比同一接口的 tag 列表是否一致。
func diffOperationTags(generated *openapi3.T, baseline *openapi3.T) []string {
	return diffOperationField(generated, baseline, func(path string, method string, generatedOp *openapi3.Operation, baselineOp *openapi3.Operation) string {
		generatedTags := normalizeStringSlice(generatedOp.Tags)
		baselineTags := normalizeStringSlice(baselineOp.Tags)

		if stringSlicesEqual(generatedTags, baselineTags) {
			return ""
		}

		return fmt.Sprintf(
			"%s %s :: generated=%s baseline=%s",
			method,
			path,
			strings.Join(generatedTags, ","),
			strings.Join(baselineTags, ","),
		)
	})
}

// diffOperationSecurity 对比同一接口的鉴权声明是否一致。
func diffOperationSecurity(generated *openapi3.T, baseline *openapi3.T) []string {
	return diffOperationField(generated, baseline, func(path string, method string, generatedOp *openapi3.Operation, baselineOp *openapi3.Operation) string {
		generatedSecurity := normalizeSecurityRequirements(generated.Security, generatedOp.Security)
		baselineSecurity := normalizeSecurityRequirements(baseline.Security, baselineOp.Security)

		if stringSlicesEqual(generatedSecurity, baselineSecurity) {
			return ""
		}

		return fmt.Sprintf(
			"%s %s :: generated=%s baseline=%s",
			method,
			path,
			strings.Join(generatedSecurity, ","),
			strings.Join(baselineSecurity, ","),
		)
	})
}

// diffOperationRequestBodies 对比同一接口是否声明了请求体。
func diffOperationRequestBodies(generated *openapi3.T, baseline *openapi3.T) []string {
	return diffOperationField(generated, baseline, func(path string, method string, generatedOp *openapi3.Operation, baselineOp *openapi3.Operation) string {
		generatedHasBody := generatedOp.RequestBody != nil
		baselineHasBody := baselineOp.RequestBody != nil

		if generatedHasBody == baselineHasBody {
			return ""
		}

		return fmt.Sprintf("%s %s :: generated=%t baseline=%t", method, path, generatedHasBody, baselineHasBody)
	})
}

// diffOperationResponseCodes 对比同一接口的响应状态码集合是否一致。
func diffOperationResponseCodes(generated *openapi3.T, baseline *openapi3.T) []string {
	return diffOperationField(generated, baseline, func(path string, method string, generatedOp *openapi3.Operation, baselineOp *openapi3.Operation) string {
		generatedCodes := normalizeStringSlice(operationResponseCodes(generatedOp))
		baselineCodes := normalizeStringSlice(operationResponseCodes(baselineOp))

		if stringSlicesEqual(generatedCodes, baselineCodes) {
			return ""
		}

		return fmt.Sprintf(
			"%s %s :: generated=%s baseline=%s",
			method,
			path,
			strings.Join(generatedCodes, ","),
			strings.Join(baselineCodes, ","),
		)
	})
}

// diffOperationField 遍历 baseline 中存在的接口，并使用回调抽取某一维度的差异项。
func diffOperationField(
	generated *openapi3.T,
	baseline *openapi3.T,
	diffFn func(path string, method string, generatedOp *openapi3.Operation, baselineOp *openapi3.Operation) string,
) []string {
	var diffs []string
	if generated == nil || generated.Paths == nil || baseline == nil || baseline.Paths == nil {
		return diffs
	}

	for _, path := range baseline.Paths.InMatchingOrder() {
		baselineItem := baseline.Paths.Value(path)
		generatedItem := generated.Paths.Value(path)

		if baselineItem == nil || generatedItem == nil {
			continue
		}

		for _, pair := range orderedOperations(baselineItem) {
			generatedOp := generatedItem.GetOperation(strings.ToUpper(pair.method))
			if generatedOp == nil {
				continue
			}

			if diff := diffFn(path, pair.method, generatedOp, pair.operation); diff != "" {
				diffs = append(diffs, diff)
			}
		}
	}

	sort.Strings(diffs)

	return diffs
}

// normalizeStringSlice 去除空白、排序后返回稳定的字符串切片，便于比较和展示。
func normalizeStringSlice(values []string) []string {
	out := make([]string, 0, len(values))

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		out = append(out, trimmed)
	}

	sort.Strings(out)

	return out
}

// stringSlicesEqual 判断两个已经归一化的字符串切片是否完全一致。
func stringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for idx := range left {
		if left[idx] != right[idx] {
			return false
		}
	}

	return true
}
