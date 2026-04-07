package openapidoc

import (
	"regexp"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// CountPathsAndOperations 统计文档中的路径数和操作数，用于校验和报告展示。
func CountPathsAndOperations(doc *openapi3.T) (int, int) {
	if doc == nil || doc.Paths == nil {
		return 0, 0
	}

	pathCount := doc.Paths.Len()
	operationCount := 0

	for _, path := range doc.Paths.InMatchingOrder() {
		operationCount += len(pathItemOperations(doc.Paths.Value(path)))
	}

	return pathCount, operationCount
}

// orderedOperation 表示一个按固定顺序遍历出来的 HTTP 方法与接口定义对。
type orderedOperation struct {
	method    string
	operation *openapi3.Operation
}

// orderedOperations 按固定 HTTP 方法顺序返回某条 path 下的所有接口，便于稳定输出。
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

// pathItemOperations 返回某条 path 下已定义的 HTTP 方法列表。
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

// sanitizePathForOperationID 将路径转换为适合拼接到 operationId 中的稳定标识。
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

// pathOperationSet 将文档转换成 path -> methods 的映射，便于集合对比。
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

// diffOperations 计算 base 中存在但 target 中缺失的“方法 + 路径”组合。
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

// operationResponseCodes 提取一个接口定义中的全部响应状态码。
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

// sortedMapKeys 返回 map 的有序 key 列表，保证对比输出稳定。
func sortedMapKeys[V any](input map[string]V) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// difference 计算 left 相对 right 的差集。
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

// writeMarkdownList 将差异列表格式化为 Markdown 二级标题加列表段落。
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
