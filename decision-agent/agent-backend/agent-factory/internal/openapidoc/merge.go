package openapidoc

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	pkgerrors "github.com/pkg/errors"
)

// MergeOverlay 将 overlay 中的字段覆盖合并到生成文档上。
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

// MergeMissingFromBaseline 只在当前文档缺字段时，使用 baseline 中的内容进行兜底补齐。
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

// mergeOverwrite 递归覆盖 map 字段，适合应用 overlay 这类“显式覆盖”数据源。
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

// fillMissing 递归补齐缺失字段，不覆盖目标中已有的有效内容。
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

// fillMissingValue 根据值类型决定“补齐”策略，用于 baseline fallback。
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

// normalizeValue 将 YAML 解析出的任意 map/slice 统一成 string-key map 结构。
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

// cloneValue 深拷贝任意 map/slice 值，避免合并时共享底层引用。
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

// isZeroValue 判断一个值是否可视为“空值”，用于 baseline 补齐决策。
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
