package openapidoc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	agentFactoryBasePath = "/api/agent-factory"
	scalarCDNURL         = "https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.34.6"
	apiKeySecurityName   = "ApiKeyAuth"
	bearerSecurityName   = "BearerAuth"
)

var pathParamPattern = regexp.MustCompile(`\{([^}/]+)\}`)

type BuildOptions struct {
	SwaggerPath           string
	OverlayPath           string
	BaselinePath          string
	ApplyBaselineFallback bool
}

type BuildArtifacts struct {
	GeneratedDoc  *openapi3.T
	FinalDoc      *openapi3.T
	CompareReport string
	JSON          []byte
	YAML          []byte
	HTML          []byte
}

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
	}, nil
}

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

func RewriteAgentFactoryPaths(doc *openapi3.T) {
	if doc == nil || doc.Paths == nil {
		return
	}

	rewritten := openapi3.NewPathsWithCapacity(doc.Paths.Len())

	for _, path := range doc.Paths.InMatchingOrder() {
		item := doc.Paths.Value(path)
		rewritten.Set(prefixAgentFactoryPath(path), item)
	}

	doc.Paths = rewritten
}

func MergeOverlay(doc *openapi3.T, overlayData []byte) error {
	if doc == nil || len(strings.TrimSpace(string(overlayData))) == 0 {
		return nil
	}

	baseMap, err := docToMap(doc)
	if err != nil {
		return err
	}

	overlayMap, err := loadMap(overlayData)
	if err != nil {
		return pkgerrors.Wrap(err, "parse overlay")
	}

	mergeOverwrite(baseMap, overlayMap)

	mergedDoc, err := mapToDoc(baseMap)
	if err != nil {
		return err
	}

	*doc = *mergedDoc

	return nil
}

func MergeMissingFromBaseline(doc *openapi3.T, baseline *openapi3.T) error {
	if doc == nil || baseline == nil {
		return nil
	}

	baseMap, err := docToMap(doc)
	if err != nil {
		return err
	}

	baselineMap, err := docToMap(baseline)
	if err != nil {
		return err
	}

	fillMissing(baseMap, baselineMap)

	mergedDoc, err := mapToDoc(baseMap)
	if err != nil {
		return err
	}

	*doc = *mergedDoc

	return nil
}

func CloneOpenAPIDoc(doc *openapi3.T) (*openapi3.T, error) {
	if doc == nil {
		return nil, nil
	}

	data, err := MarshalOpenAPIJSON(doc)
	if err != nil {
		return nil, err
	}

	return loadOpenAPIDoc(data)
}

func LoadOpenAPIDocFile(path string) (*openapi3.T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read openapi file")
	}

	return loadOpenAPIDoc(data)
}

func MarshalOpenAPIJSON(doc *openapi3.T) ([]byte, error) {
	raw, err := doc.MarshalJSON()
	if err != nil {
		return nil, pkgerrors.Wrap(err, "marshal openapi json")
	}

	var pretty any
	if err := json.Unmarshal(raw, &pretty); err != nil {
		return nil, pkgerrors.Wrap(err, "decode openapi json for pretty print")
	}

	prettyJSON, err := json.MarshalIndent(pretty, "", "  ")
	if err != nil {
		return nil, pkgerrors.Wrap(err, "pretty print openapi json")
	}

	return append(prettyJSON, '\n'), nil
}

func MarshalOpenAPIYAML(doc *openapi3.T) ([]byte, error) {
	rawMap, err := docToMap(doc)
	if err != nil {
		return nil, err
	}

	rawYAML, err := yaml.Marshal(rawMap)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "marshal openapi yaml")
	}

	return rawYAML, nil
}

func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return pkgerrors.Wrap(err, "create parent directory")
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return pkgerrors.Wrap(err, "write file")
	}

	return nil
}

func RenderScalarStaticHTML(specJSON []byte) string {
	escapedSpec := strings.ReplaceAll(string(specJSON), "</script", `<\/script`)

	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Agent Factory API Reference</title>
  <style>
    body {
      margin: 0;
      padding: 0;
    }
  </style>
</head>
<body>
  <noscript>Scalar requires JavaScript to render the API reference.</noscript>
  <script type="application/json" id="openapi-document">` + escapedSpec + `</script>
  <script>
    (() => {
      const specElement = document.getElementById("openapi-document");
      const referenceElement = document.createElement("script");
      const specBlob = new Blob([specElement.textContent], { type: "application/json" });
      const specURL = URL.createObjectURL(specBlob);
      referenceElement.id = "api-reference";
      referenceElement.dataset.url = specURL;
      document.body.appendChild(referenceElement);
      window.addEventListener("pagehide", () => URL.revokeObjectURL(specURL), { once: true });
    })();
  </script>
  <script src="` + scalarCDNURL + `"></script>
</body>
</html>
`
}

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

func CountPathsAndOperations(doc *openapi3.T) (int, int) {
	if doc == nil || doc.Paths == nil {
		return 0, 0
	}

	pathCount := doc.Paths.Len()
	operationCount := 0

	for _, path := range doc.Paths.InMatchingOrder() {
		item := doc.Paths.Value(path)
		operationCount += len(pathItemOperations(item))
	}

	return pathCount, operationCount
}

func ValidateOpenAPI(ctx context.Context, doc *openapi3.T) error {
	if doc == nil {
		return pkgerrors.New("openapi document is nil")
	}

	return doc.Validate(
		ctx,
		openapi3.DisableExamplesValidation(),
		openapi3.AllowExtraSiblingFields(
			"content",
			"description",
			"summary",
			"properties",
			"required",
			"type",
			"items",
			"example",
			"default",
			"nullable",
			"enum",
			"format",
			"allOf",
			"oneOf",
			"anyOf",
		),
	)
}

func NormalizePathParameters(doc *openapi3.T) {
	if doc == nil || doc.Paths == nil {
		return
	}

	if doc.Components != nil {
		for _, parameterRef := range doc.Components.Parameters {
			normalizePathParameterRef(parameterRef)
		}
	}

	for _, path := range doc.Paths.InMatchingOrder() {
		pathItem := doc.Paths.Value(path)
		normalizeParameters(pathItem.Parameters)
		pathLevelParams := collectPathParameterNames(pathItem.Parameters)

		for _, operation := range pathItem.Operations() {
			normalizeParameters(operation.Parameters)
		}

		for _, name := range extractPathParameterNames(path) {
			if _, exists := pathLevelParams[name]; exists {
				continue
			}

			pathItem.Parameters = append(pathItem.Parameters, &openapi3.ParameterRef{
				Value: openapi3.NewPathParameter(name).WithSchema(openapi3.NewStringSchema()),
			})
			pathLevelParams[name] = struct{}{}
		}
	}
}

func NormalizeOperationIDs(doc *openapi3.T) {
	if doc == nil || doc.Paths == nil {
		return
	}

	counts := map[string]int{}

	for _, path := range doc.Paths.InMatchingOrder() {
		for _, pair := range orderedOperations(doc.Paths.Value(path)) {
			if pair.operation.OperationID == "" {
				continue
			}

			counts[pair.operation.OperationID]++
		}
	}

	for _, path := range doc.Paths.InMatchingOrder() {
		for _, pair := range orderedOperations(doc.Paths.Value(path)) {
			if pair.operation.OperationID == "" || counts[pair.operation.OperationID] <= 1 {
				continue
			}

			pair.operation.OperationID = fmt.Sprintf(
				"%s_%s_%s",
				pair.operation.OperationID,
				strings.ToLower(pair.method),
				sanitizePathForOperationID(path),
			)
		}
	}
}

func NormalizeSecurity(doc *openapi3.T) {
	if doc == nil {
		return
	}

	if doc.Components == nil {
		components := openapi3.NewComponents()
		doc.Components = &components
	}

	if doc.Components.SecuritySchemes == nil {
		doc.Components.SecuritySchemes = openapi3.SecuritySchemes{}
	}

	if legacyScheme, ok := doc.Components.SecuritySchemes[bearerSecurityName]; ok {
		if _, exists := doc.Components.SecuritySchemes[apiKeySecurityName]; !exists {
			doc.Components.SecuritySchemes[apiKeySecurityName] = legacyScheme
		}

		delete(doc.Components.SecuritySchemes, bearerSecurityName)
	}

	doc.Security = normalizeSecurityRequirementsRef(doc.Security)

	if doc.Paths == nil {
		return
	}

	for _, path := range doc.Paths.InMatchingOrder() {
		pathItem := doc.Paths.Value(path)
		for _, pair := range orderedOperations(pathItem) {
			if pair.operation.Security == nil {
				continue
			}

			normalized := normalizeSecurityRequirementsRef(*pair.operation.Security)
			if securityRequirementsEqual(normalized, doc.Security) {
				pair.operation.Security = nil
				continue
			}

			pair.operation.Security = &normalized
		}
	}
}

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

func normalizeParameters(parameters openapi3.Parameters) {
	for _, parameterRef := range parameters {
		normalizePathParameterRef(parameterRef)
	}
}

func normalizePathParameterRef(parameterRef *openapi3.ParameterRef) {
	if parameterRef == nil || parameterRef.Value == nil {
		return
	}

	if strings.EqualFold(parameterRef.Value.In, "path") {
		parameterRef.Value.Required = true
	}
}

func collectPathParameterNames(parameters openapi3.Parameters) map[string]struct{} {
	names := map[string]struct{}{}

	for _, parameterRef := range parameters {
		if parameterRef == nil || parameterRef.Value == nil {
			continue
		}

		if !strings.EqualFold(parameterRef.Value.In, "path") {
			continue
		}

		names[parameterRef.Value.Name] = struct{}{}
	}

	return names
}

func extractPathParameterNames(path string) []string {
	matches := pathParamPattern.FindAllStringSubmatch(path, -1)
	names := make([]string, 0, len(matches))

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		names = append(names, match[1])
	}

	return names
}

type orderedOperation struct {
	method    string
	operation *openapi3.Operation
}

func orderedOperations(pathItem *openapi3.PathItem) []orderedOperation {
	if pathItem == nil {
		return nil
	}

	pairs := make([]orderedOperation, 0, 8)
	if pathItem.Get != nil {
		pairs = append(pairs, orderedOperation{method: "GET", operation: pathItem.Get})
	}

	if pathItem.Post != nil {
		pairs = append(pairs, orderedOperation{method: "POST", operation: pathItem.Post})
	}

	if pathItem.Put != nil {
		pairs = append(pairs, orderedOperation{method: "PUT", operation: pathItem.Put})
	}

	if pathItem.Delete != nil {
		pairs = append(pairs, orderedOperation{method: "DELETE", operation: pathItem.Delete})
	}

	if pathItem.Patch != nil {
		pairs = append(pairs, orderedOperation{method: "PATCH", operation: pathItem.Patch})
	}

	if pathItem.Options != nil {
		pairs = append(pairs, orderedOperation{method: "OPTIONS", operation: pathItem.Options})
	}

	if pathItem.Head != nil {
		pairs = append(pairs, orderedOperation{method: "HEAD", operation: pathItem.Head})
	}

	return pairs
}

func sanitizePathForOperationID(path string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"{", "",
		"}", "",
		"-", "_",
	)
	sanitized := replacer.Replace(strings.Trim(path, "/"))
	sanitized = regexp.MustCompile(`_+`).ReplaceAllString(sanitized, "_")

	return strings.Trim(sanitized, "_")
}

func loadOpenAPIDoc(data []byte) (*openapi3.T, error) {
	normalizedMap, err := loadMap(data)
	if err != nil {
		return nil, err
	}

	return mapToDoc(normalizedMap)
}

func loadMap(data []byte) (map[string]any, error) {
	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, pkgerrors.Wrap(err, "unmarshal yaml/json map")
	}

	normalized, ok := normalizeValue(raw).(map[string]any)
	if !ok {
		return nil, pkgerrors.New("document root must be an object")
	}

	sanitizeOpenAPIMap(normalized)

	return normalized, nil
}

func docToMap(doc *openapi3.T) (map[string]any, error) {
	raw, err := doc.MarshalJSON()
	if err != nil {
		return nil, pkgerrors.Wrap(err, "marshal openapi doc to map")
	}

	return loadMap(raw)
}

func mapToDoc(raw map[string]any) (*openapi3.T, error) {
	jsonData, err := json.Marshal(raw)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "marshal normalized map")
	}

	loader := openapi3.NewLoader()

	doc, err := loader.LoadFromData(jsonData)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "load openapi from normalized map")
	}

	if doc.Components == nil {
		components := openapi3.NewComponents()
		doc.Components = &components
	}

	if doc.Paths == nil {
		doc.Paths = openapi3.NewPaths()
	}

	return doc, nil
}

func mergeOverwrite(dst map[string]any, src map[string]any) {
	for key, srcValue := range src {
		srcMap, srcIsMap := srcValue.(map[string]any)
		dstMap, dstIsMap := dst[key].(map[string]any)

		if srcIsMap && dstIsMap {
			mergeOverwrite(dstMap, srcMap)
			dst[key] = dstMap

			continue
		}

		dst[key] = cloneValue(srcValue)
	}
}

func fillMissing(dst map[string]any, src map[string]any) {
	for key, srcValue := range src {
		existing, exists := dst[key]
		if !exists {
			dst[key] = cloneValue(srcValue)
			continue
		}

		dst[key] = fillMissingValue(existing, srcValue)
	}
}

func fillMissingValue(dst any, src any) any {
	src = normalizeValue(src)
	dst = normalizeValue(dst)

	srcMap, srcIsMap := src.(map[string]any)
	dstMap, dstIsMap := dst.(map[string]any)

	if srcIsMap && dstIsMap {
		fillMissing(dstMap, srcMap)
		return dstMap
	}

	if srcIsMap && dst == nil {
		return cloneValue(srcMap)
	}

	srcSlice, srcIsSlice := src.([]any)
	dstSlice, dstIsSlice := dst.([]any)

	if srcIsSlice {
		if !dstIsSlice || len(dstSlice) == 0 {
			return cloneValue(srcSlice)
		}

		return dstSlice
	}

	if isZeroValue(dst) {
		return cloneValue(src)
	}

	return dst
}

func normalizeValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, nested := range v {
			out[key] = normalizeValue(nested)
		}

		return out
	case map[any]any:
		out := make(map[string]any, len(v))
		for key, nested := range v {
			out[fmt.Sprint(key)] = normalizeValue(nested)
		}

		return out
	case []any:
		out := make([]any, len(v))
		for idx, nested := range v {
			out[idx] = normalizeValue(nested)
		}

		return out
	default:
		return v
	}
}

func cloneValue(value any) any {
	switch v := normalizeValue(value).(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, nested := range v {
			out[key] = cloneValue(nested)
		}

		return out
	case []any:
		out := make([]any, len(v))
		for idx, nested := range v {
			out[idx] = cloneValue(nested)
		}

		return out
	default:
		return v
	}
}

func sanitizeOpenAPIMap(node map[string]any) {
	sanitizeOpenAPIMapWithContext(node, false, false)
}

func sanitizeOpenAPIMapWithContext(node map[string]any, inSchema bool, inExample bool) {
	for key, value := range node {
		if !inExample && key == "required" {
			if _, isBool := value.(bool); isBool {
				delete(node, key)
				continue
			}
		}

		if !inExample && key == "description" {
			if descriptionLines, ok := value.([]any); ok {
				node[key] = joinStringList(descriptionLines)
				continue
			}
		}

		if inSchema && !inExample && key == "type" {
			if _, isList := value.([]any); isList {
				delete(node, key)
				continue
			}

			if normalizedType, keep := normalizeSchemaType(value); keep {
				node[key] = normalizedType
			} else {
				delete(node, key)
				continue
			}
		}

		sanitizeOpenAPIValue(value, inSchema || isSchemaContextKey(key), inExample || key == "example" || key == "examples")
	}

	if inSchema && !inExample {
		if schemaType, ok := node["type"].(string); ok {
			if schemaType == "array" {
				if _, hasItems := node["items"]; !hasItems {
					node["items"] = map[string]any{}
				}
			}

			if sanitizedExample, exists := node["example"]; exists {
				if normalizedExample, keep := sanitizeExampleForSchemaType(schemaType, sanitizedExample); keep {
					node["example"] = normalizedExample
				} else {
					delete(node, "example")
				}
			}
		}
	}
}

func sanitizeOpenAPIValue(value any, inSchema bool, inExample bool) {
	switch child := value.(type) {
	case map[string]any:
		sanitizeOpenAPIMapWithContext(child, inSchema, inExample)
	case []any:
		for _, item := range child {
			sanitizeOpenAPIValue(item, inSchema, inExample)
		}
	}
}

func isSchemaContextKey(key string) bool {
	switch key {
	case "schema", "schemas", "properties", "items", "additionalProperties", "allOf", "anyOf", "oneOf", "not":
		return true
	default:
		return false
	}
}

func normalizeSchemaType(value any) (string, bool) {
	schemaType, ok := value.(string)
	if !ok {
		return "", false
	}

	switch strings.TrimSpace(schemaType) {
	case "string", "number", "integer", "boolean", "array", "object":
		return schemaType, true
	case "int":
		return "integer", true
	case "float", "double":
		return "number", true
	case "any", "null":
		return "", false
	default:
		return schemaType, true
	}
}

func joinStringList(items []any) string {
	lines := make([]string, 0, len(items))

	for _, item := range items {
		line := strings.TrimSpace(fmt.Sprint(item))
		if line == "" {
			continue
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func sanitizeExampleForSchemaType(schemaType string, example any) (any, bool) {
	switch schemaType {
	case "string":
		return fmt.Sprint(example), true
	case "integer":
		switch v := example.(type) {
		case int:
			return v, true
		case int64:
			return v, true
		case float64:
			if v == float64(int64(v)) {
				return int64(v), true
			}
		case string:
			parsed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
			if err == nil {
				return parsed, true
			}
		}

		return nil, false
	case "number":
		switch v := example.(type) {
		case int:
			return float64(v), true
		case int64:
			return float64(v), true
		case float64:
			return v, true
		case string:
			parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err == nil {
				return parsed, true
			}
		}

		return nil, false
	case "boolean":
		switch v := example.(type) {
		case bool:
			return v, true
		case string:
			parsed, err := strconv.ParseBool(strings.TrimSpace(v))
			if err == nil {
				return parsed, true
			}
		}

		return nil, false
	case "array":
		if values, ok := example.([]any); ok {
			return values, true
		}

		return nil, false
	case "object":
		if valueMap, ok := example.(map[string]any); ok {
			return valueMap, true
		}

		return nil, false
	default:
		return example, true
	}
}

func isZeroValue(value any) bool {
	switch v := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(v) == ""
	case []any:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	default:
		return false
	}
}

func pathOperationSet(doc *openapi3.T) map[string][]string {
	out := map[string][]string{}
	if doc == nil || doc.Paths == nil {
		return out
	}

	for _, path := range doc.Paths.InMatchingOrder() {
		methods := pathItemOperations(doc.Paths.Value(path))
		sort.Strings(methods)
		out[path] = methods
	}

	return out
}

func pathItemOperations(item *openapi3.PathItem) []string {
	if item == nil {
		return nil
	}

	operations := make([]string, 0, 8)
	if item.Get != nil {
		operations = append(operations, "GET")
	}

	if item.Post != nil {
		operations = append(operations, "POST")
	}

	if item.Put != nil {
		operations = append(operations, "PUT")
	}

	if item.Delete != nil {
		operations = append(operations, "DELETE")
	}

	if item.Patch != nil {
		operations = append(operations, "PATCH")
	}

	if item.Options != nil {
		operations = append(operations, "OPTIONS")
	}

	if item.Head != nil {
		operations = append(operations, "HEAD")
	}

	return operations
}

func diffOperations(base map[string][]string, target map[string][]string) []string {
	var diffs []string

	for _, path := range sortedMapKeys(base) {
		baseOps := base[path]
		targetOps := target[path]
		targetSet := make(map[string]struct{}, len(targetOps))

		for _, method := range targetOps {
			targetSet[method] = struct{}{}
		}

		for _, method := range baseOps {
			if _, ok := targetSet[method]; ok {
				continue
			}

			diffs = append(diffs, method+" "+path)
		}
	}

	return diffs
}

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

func diffOperationField(generated *openapi3.T, baseline *openapi3.T, diffFn func(path string, method string, generatedOp *openapi3.Operation, baselineOp *openapi3.Operation) string) []string {
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

func normalizeSecurityRequirements(global openapi3.SecurityRequirements, override *openapi3.SecurityRequirements) []string {
	requirements := global
	if override != nil {
		requirements = *override
	}

	values := make([]string, 0, len(requirements))

	for _, requirement := range requirements {
		keys := make([]string, 0, len(requirement))
		for key := range requirement {
			keys = append(keys, key)
		}

		sort.Strings(keys)
		values = append(values, strings.Join(keys, "+"))
	}

	return normalizeStringSlice(values)
}

func normalizeSecurityRequirementsRef(requirements openapi3.SecurityRequirements) openapi3.SecurityRequirements {
	if requirements == nil {
		return nil
	}

	normalized := make(openapi3.SecurityRequirements, 0, len(requirements))

	for _, requirement := range requirements {
		if requirement == nil {
			normalized = append(normalized, openapi3.SecurityRequirement{})
			continue
		}

		next := openapi3.SecurityRequirement{}

		for key, scopes := range requirement {
			targetKey := key
			if key == bearerSecurityName {
				targetKey = apiKeySecurityName
			}

			next[targetKey] = append([]string(nil), scopes...)
		}

		normalized = append(normalized, next)
	}

	return normalized
}

func securityRequirementsEqual(left openapi3.SecurityRequirements, right openapi3.SecurityRequirements) bool {
	return stringSlicesEqual(normalizeSecurityRequirementNames(left), normalizeSecurityRequirementNames(right))
}

func normalizeSecurityRequirementNames(requirements openapi3.SecurityRequirements) []string {
	values := make([]string, 0, len(requirements))

	for _, requirement := range requirements {
		keys := make([]string, 0, len(requirement))
		for key := range requirement {
			keys = append(keys, key)
		}

		sort.Strings(keys)
		values = append(values, strings.Join(keys, "+"))
	}

	return normalizeStringSlice(values)
}

func operationResponseCodes(operation *openapi3.Operation) []string {
	if operation == nil || operation.Responses == nil {
		return nil
	}

	responseMap := operation.Responses.Map()
	codes := make([]string, 0, len(responseMap))

	for code := range responseMap {
		codes = append(codes, code)
	}

	return codes
}

func sortedMapKeys[V any](input map[string]V) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func difference(left []string, right []string) []string {
	rightSet := make(map[string]struct{}, len(right))
	for _, item := range right {
		rightSet[item] = struct{}{}
	}

	var diff []string

	for _, item := range left {
		if _, ok := rightSet[item]; ok {
			continue
		}

		diff = append(diff, item)
	}

	return diff
}

func writeMarkdownList(builder *strings.Builder, title string, items []string) {
	builder.WriteString("## " + title + "\n")

	if len(items) == 0 {
		builder.WriteString("- none\n\n")
		return
	}

	for _, item := range items {
		builder.WriteString("- " + item + "\n")
	}

	builder.WriteString("\n")
}
