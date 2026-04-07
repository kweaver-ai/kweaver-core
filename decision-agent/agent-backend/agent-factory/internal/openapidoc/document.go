package openapidoc

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// CloneOpenAPIDoc 通过序列化再反序列化的方式深拷贝 OpenAPI 文档。
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

// LoadOpenAPIDocFile 从 JSON 或 YAML 文件读取并加载 OpenAPI 文档。
func LoadOpenAPIDocFile(path string) (*openapi3.T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "read openapi file")
	}

	return loadOpenAPIDoc(data)
}

// MarshalOpenAPIJSON 将文档格式化为带缩进的 JSON，便于提交和审阅。
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

// MarshalOpenAPIYAML 将文档转为 YAML，用于对外分发和人工阅读。
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

// WriteFile 确保父目录存在后再写文件，统一处理文档输出。
func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return pkgerrors.Wrap(err, "create parent directory")
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return pkgerrors.Wrap(err, "write file")
	}

	return nil
}

// loadOpenAPIDoc 将原始文件内容标准化后加载为 kin-openapi 文档对象。
func loadOpenAPIDoc(data []byte) (*openapi3.T, error) {
	normalizedMap, err := loadMap(data)
	if err != nil {
		return nil, err
	}

	return mapToDoc(normalizedMap)
}

// loadMap 将 YAML/JSON 解析为 map，并执行兼容性清洗与结构标准化。
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

// docToMap 将 OpenAPI 文档转为 map 结构，便于做深度合并和归一化处理。
func docToMap(doc *openapi3.T) (map[string]any, error) {
	raw, err := doc.MarshalJSON()
	if err != nil {
		return nil, pkgerrors.Wrap(err, "marshal openapi doc to map")
	}

	return loadMap(raw)
}

// mapToDoc 将标准化后的 map 重新加载为 OpenAPI 文档对象。
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
