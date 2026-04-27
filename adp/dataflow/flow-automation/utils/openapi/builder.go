package openapi

import (
	"fmt"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

type OpenAPI struct {
	doc *openapi3.T
}

func NewOpenAPI() *OpenAPI {
	return &OpenAPI{
		doc: &openapi3.T{
			OpenAPI: "3.0.0",
		},
	}
}

func (o *OpenAPI) AddOpenAPIInfo(title, description string) *OpenAPI {
	o.doc.Info = &openapi3.Info{
		Title:       title,
		Description: description,
		Version:     "1.0.0",
	}

	return o
}

func (o *OpenAPI) AddServers(URL string) *OpenAPI {
	o.doc.Servers = openapi3.Servers{
		&openapi3.Server{
			URL: URL,
		},
	}
	return o
}

func (o *OpenAPI) AddPaths(paths *openapi3.Paths) *OpenAPI {
	o.doc.Paths = paths
	return o
}

func (o *OpenAPI) Build() *openapi3.T {
	return o.doc
}

// OpenAPIDocumentBuilder 实现文档构建
type OpenAPIDocumentBuilder struct {
	paths map[string]*openapi3.PathItem
}

func NewOpenAPIDocumentBuilder() *OpenAPIDocumentBuilder {
	return &OpenAPIDocumentBuilder{
		paths: make(map[string]*openapi3.PathItem),
	}
}

func (b *OpenAPIDocumentBuilder) AddPath(path string) *OpenAPIDocumentBuilder {
	if _, exists := b.paths[path]; !exists {
		b.paths[path] = &openapi3.PathItem{}
	}
	return b
}

func (b *OpenAPIDocumentBuilder) AddOperation(path, method string, operation *openapi3.Operation) *OpenAPIDocumentBuilder {
	b.AddPath(path)

	switch strings.ToUpper(method) {
	case "GET":
		b.paths[path].Get = operation
	case "POST":
		b.paths[path].Post = operation
	case "PUT":
		b.paths[path].Put = operation
	case "DELETE":
		b.paths[path].Delete = operation
	case "PATCH":
		b.paths[path].Patch = operation
	}
	return b
}

func (b *OpenAPIDocumentBuilder) Build() *openapi3.Paths {
	paths := &openapi3.Paths{}
	for path, item := range b.paths {
		paths.Set(path, item)
	}
	return paths
}

type OperationBuilder struct {
	operation *openapi3.Operation
}

// Header 表示 OpenAPI 中的 Header 定义
type Header struct {
	Description string
	Type        string
	Example     interface{}
}

// HeaderRef 是 Header 的引用
type HeaderRef struct {
	Value *Header
}

func NewOperationBuilder() *OperationBuilder {
	return &OperationBuilder{
		operation: &openapi3.Operation{},
	}
}

func (b *OperationBuilder) Build() *openapi3.Operation {
	return b.operation
}

func (b *OperationBuilder) WithOperatorMeta(summary, description string) *OperationBuilder {
	b.operation.Summary = summary
	b.operation.Description = description

	return b
}

func (b *OperationBuilder) WithRequestParam(name, description, schemaType, in string, required bool) *OperationBuilder {
	if b.operation.Parameters == nil {
		b.operation.Parameters = openapi3.Parameters{}
	}

	param := &openapi3.Parameter{
		Name:        name,
		In:          in,
		Description: description,
		Required:    required,
		Schema: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: &openapi3.Types{schemaType},
			},
		},
	}
	b.operation.Parameters = append(b.operation.Parameters, &openapi3.ParameterRef{Value: param})
	return b
}

func (b *OperationBuilder) WithResponseHeader(statusCode int, name, description, schemaType string) *OperationBuilder {
	response := b.operation.Responses.Status(statusCode)
	if response == nil {
		response = &openapi3.ResponseRef{
			Value: &openapi3.Response{},
		}
		b.operation.Responses.Set(fmt.Sprintf("%v", statusCode), response)
	}

	if response.Value.Headers == nil {
		response.Value.Headers = make(map[string]*openapi3.HeaderRef)
	}

	response.Value.Headers[name] = &openapi3.HeaderRef{
		Value: &openapi3.Header{
			Parameter: openapi3.Parameter{
				Description: description,
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{schemaType},
					},
				},
			},
		},
	}
	return b
}

func (b *OperationBuilder) WithRequestBody(contentType string, schema interface{}, schemaParse func(interface{}) *openapi3.SchemaRef) *OperationBuilder {
	b.operation.RequestBody = &openapi3.RequestBodyRef{
		Value: &openapi3.RequestBody{
			Content: openapi3.Content{
				contentType: &openapi3.MediaType{
					Schema: schemaParse(schema),
				},
			},
		},
	}
	return b
}

func (b *OperationBuilder) WithResponse(statusCode, description string) *OperationBuilder {
	if b.operation.Responses == nil {
		b.operation.Responses = &openapi3.Responses{}
	}

	b.operation.Responses.Set(statusCode, &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: &description,
		},
	})
	return b
}

func (b *OperationBuilder) WithResponseContent(statusCode int, contentType string, schema interface{}, schemaParse func(interface{}) *openapi3.SchemaRef) *OperationBuilder {
	response := b.operation.Responses.Status(statusCode)
	if response == nil {
		b.WithResponse(fmt.Sprintf("%v", statusCode), "")
		response = b.operation.Responses.Status(statusCode)
	}

	if response.Value.Content == nil {
		response.Value.Content = make(openapi3.Content)
	}

	response.Value.Content[contentType] = &openapi3.MediaType{
		Schema: schemaParse(schema),
	}
	return b
}

func ConvertFlowParamsToSchema(data interface{}) *openapi3.SchemaRef {
	_data, ok := data.(map[string]interface{})
	if !ok {
		return &openapi3.SchemaRef{}
	}
	schema := &openapi3.Schema{
		Type:       &openapi3.Types{"object"},
		Properties: make(map[string]*openapi3.SchemaRef),
	}
	var required []string

	fields, ok := _data["fields"].([]interface{})
	if !ok {
		return &openapi3.SchemaRef{Value: schema}
	}

	for _, field := range fields {
		param, ok := field.(map[string]interface{})
		if !ok {
			continue
		}

		key := fmt.Sprintf("%v", param["key"])	
		// 默认使用key作为描述
		desc := key

		// 先查看是否存在描述信息
		if descMap, ok := param["description"].(map[string]interface{}); ok {
			dt :=descMap["text"]
			if dt == nil {
				dt = ""
			}
			desc = fmt.Sprintf("%v", dt)
		} else if name, ok := param["name"].(string); ok {
			desc = name
		}

		if r, ok := param["required"]; ok && r.(bool) {
			required = append(required, key)
		}

		fieldSchema := createFieldSchema(param["type"], desc)
		if fieldSchema != nil {
			schema.Properties[key] = fieldSchema
		}
	}

	schema.Required = required
	return &openapi3.SchemaRef{Value: schema}
}

func createFieldSchema(fieldType interface{}, description string) *openapi3.SchemaRef {
	typeStr := fmt.Sprintf("%v", fieldType)

	switch typeStr {
	case "string", "long_string", "radio", "asFile":
		return createBasicSchema("string", description)
	case "datetime":
		return &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:        &openapi3.Types{"string"},
				Format:      "date-time",
				Description: description,
			},
		}
	case "number":
		return createBasicSchema("number", description)
	case "multipleFiles":
		return createArraySchema(createBasicSchema("string", description), description)
	case "asPerm":
		return createObjectSchema(map[string]*openapi3.SchemaRef{
			"allow": createBasicSchema("string", ""),
			"deny":  createBasicSchema("string", ""),
		}, description)
	case "asUsers", "asDepartments":
		itemSchema := createObjectSchema(map[string]*openapi3.SchemaRef{
			"id":   createBasicSchema("string", ""),
			"name": createBasicSchema("string", ""),
			"type": createBasicSchema("string", ""),
		}, "")
		return createArraySchema(itemSchema, description)
	case "object":
		return createObjectSchema(map[string]*openapi3.SchemaRef{}, description)
	case "array":
		return createArraySchema(createObjectSchema(map[string]*openapi3.SchemaRef{}, ""), description)
	default:
		return nil
	}
}

func createBasicSchema(typeStr, description string) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        &openapi3.Types{typeStr},
			Description: description,
		},
	}
}

func createArraySchema(items *openapi3.SchemaRef, description string) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        &openapi3.Types{"array"},
			Description: description,
			Items:       items,
		},
	}
}

func createObjectSchema(properties map[string]*openapi3.SchemaRef, description string) *openapi3.SchemaRef {
	return &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type:        &openapi3.Types{"object"},
			Description: description,
			Properties:  properties,
		},
	}
}

// 增强版类型转换
func ConvertMapToSchema(data interface{}) *openapi3.SchemaRef {
	schema := &openapi3.Schema{}

	switch v := data.(type) {
	case map[string]interface{}:
		schema.Type = &openapi3.Types{"object"}
		schema.Properties = make(map[string]*openapi3.SchemaRef)
		for key, val := range v {
			schema.Properties[key] = ConvertMapToSchema(val)
		}
	case []interface{}:
		if len(v) > 0 {
			schema.Type = &openapi3.Types{"array"}
			schema.Items = ConvertMapToSchema(v[0])
		}
	case string:
		schema.Type = &openapi3.Types{"string"}
		// 识别特殊格式
		if _, err := time.Parse(time.RFC3339, v); err == nil {
			schema.Format = "date-time"
		} else if len(v) > 7 && v[:6] == "gns://" {
			schema.Pattern = `^gns://([A-Z0-9]{32}/?)+$`
		}
	case float64:
		if v == float64(int(v)) {
			schema.Type = &openapi3.Types{"integer"}
		} else {
			schema.Type = &openapi3.Types{"number"}
		}
	case bool:
		schema.Type = &openapi3.Types{"boolean"}
	case nil:
		schema.Type = &openapi3.Types{"null"}
	}

	return &openapi3.SchemaRef{Value: schema}
}
