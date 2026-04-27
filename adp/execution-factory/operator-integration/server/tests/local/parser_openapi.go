package local

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

// APIMetadata API元数据
type APIMetadata struct {
	ID          string       `json:"id" validate:"required"`          // 主键ID
	Hash        string       `json:"hash" validate:"required"`        // 哈希值
	Version     string       `json:"version" validate:"required"`     // 版本
	Title       string       `json:"title" validate:"required"`       // 标题
	Summary     string       `json:"summary" validate:"required"`     // 摘要
	Description string       `json:"description" validate:"required"` // 描述
	Path        string       `json:"path" validate:"required"`        // 路径
	Method      string       `json:"method" validate:"required"`      // 方法
	Parameters  []*Parameter `json:"parameters" validate:"required"`  // 结构化参数
	RequestBody *RequestBody `json:"request_body"`                    // 请求体结构
	Responses   []*Response  `json:"responses"`                       // 响应结构
	CreateTime  int64        `json:"create_time" validate:"required"` // 创建时间
	UpdateTime  int64        `json:"update_time" validate:"required"` // 更新时间
	CreateUser  string       `json:"create_user" validate:"required"` // 创建人
	UpdateUser  string       `json:"update_user" validate:"required"` // 更新人
	IsDeleted   bool         `json:"is_deleted" validate:"required"`  // 是否删除
	// APISpec     string      `json:"api_spec" validate:"required"`    // OpenAPI 格式
}

// type APISpec

// Parameter 参数类型
type Parameter struct {
	Name        string                 `json:"name"`
	In          string                 `json:"in"` // path/query/header/cookie
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	Ref         string                 `json:"$ref,omitempty"` // 新增引用字段
	Schema      map[string]interface{} `json:"schema"`         // 参数schema
}

// RequestBody 请求体结构
type RequestBody struct {
	Description string             `json:"description"`
	Content     map[string]Content `json:"content"` // 按媒体类型分类
}

// Response 响应结构
type Response struct {
	StatusCode  string             `json:"status_code"` // 200/400等
	Description string             `json:"description"`
	Content     map[string]Content `json:"content"`
}

// Content 内容结构
type Content struct {
	Ref     string                 `json:"$ref,omitempty"` // 新增引用字段
	Schema  map[string]interface{} `json:"schema"`         // 完整schema
	Example interface{}            `json:"example,omitempty"`
}

// 新增schema转换方法
func schemaRefToMap(ref *openapi3.SchemaRef, components *openapi3.Components,
	visited map[string]bool) (refPath string, schema map[string]interface{}) {
	if ref == nil {
		return "", nil
	}
	// 处理引用
	if ref.Ref != "" {
		refPath = ref.Ref
		refKey := strings.TrimPrefix(ref.Ref, "#/components/schemas/")
		if visited[refKey] {
			return refPath, map[string]interface{}{"$ref": refPath}
		}
		visited[refKey] = true
		defer delete(visited, refKey)

		if schemaDef, exists := components.Schemas[refKey]; exists {
			// 保留原始引用
			schema = map[string]interface{}{"$ref": refPath}
			// 同时保留解析后的schema
			_, resolved := schemaRefToMap(schemaDef, components, visited)
			for k, v := range resolved {
				if k != "$ref" { // 避免覆盖原始引用
					schema[k] = v
				}
			}
			return refPath, schema
		}
	}

	// 解析当前schema
	schema = make(map[string]interface{})
	if ref.Value == nil {
		return refPath, schema
	}
	// 保留原始引用信息
	if refPath != "" {
		schema["$ref"] = refPath
	}

	// 基本类型处理
	if ref.Value.Type != nil {
		schema["type"] = ref.Value.Type
	}

	if ref.Value.Format != "" {
		schema["format"] = ref.Value.Format
	}

	// 处理嵌套结构
	if ref.Value.Properties != nil {
		props := make(map[string]interface{})
		for name, prop := range ref.Value.Properties {
			refPath, resolved := schemaRefToMap(prop, components, visited)
			if refPath != "" {
				props[name] = map[string]interface{}{"$ref": refPath}
				components.Schemas[strings.TrimPrefix(refPath, "#/components/schemas/")] = prop
			} else {
				props[name] = resolved
			}
		}
		schema["properties"] = props
	}

	// 处理数组类型
	if ref.Value.Items != nil {
		refPath, resolved := schemaRefToMap(ref.Value.Items, components, visited)
		if refPath != "" {
			schema["items"] = map[string]interface{}{"$ref": refPath}
			components.Schemas[strings.TrimPrefix(refPath, "#/components/schemas/")] = ref.Value.Items
		} else {
			schema["items"] = resolved
		}
	}

	// 处理组合类型
	handleComposition := func(schemas openapi3.SchemaRefs) []interface{} {
		var result []interface{}
		for _, s := range schemas {
			// 生成引用路径并保留原始结构
			refPath, resolvedSchema := schemaRefToMap(s, components, visited)
			if refPath != "" {
				result = append(result, map[string]interface{}{"$ref": refPath})
				components.Schemas[strings.TrimPrefix(refPath, "#/components/schemas/")] = s
			} else {
				result = append(result, resolvedSchema)
			}
		}
		return result
	}
	if len(ref.Value.AllOf) > 0 {
		schema["allOf"] = handleComposition(ref.Value.AllOf)
	}
	if len(ref.Value.AnyOf) > 0 {
		schema["anyOf"] = handleComposition(ref.Value.AnyOf)
	}
	if len(ref.Value.OneOf) > 0 {
		schema["oneOf"] = handleComposition(ref.Value.OneOf)
	}

	return refPath, schema
}

// type ParameterIn string

// const (
// 	ParameterInPath   ParameterIn = "path"   // 路径参数
// 	ParameterInQuery  ParameterIn = "query"  // 查询参数
// 	ParameterInHeader ParameterIn = "header" // 头部参数
// 	ParameterInCookie ParameterIn = "cookie" // Cookie 参数
// 	ParameterInBody   ParameterIn = "body"   // 请求体参数
// )

type OpenAPIDataType string

const (
	ContentDataType OpenAPIDataType = "content"
	FileDataType    OpenAPIDataType = "file"
)

// GetHash 获取哈希
// func GetHash(path, method string) (hash string, err error) {
// 	type hashGenerator struct {
// 		Path    string `json:"path"`
// 		Method  string `json:"method"`

// 	}
// 	hash, err = utils.ObjectMD5Hash(&hashGenerator{
// 		Path:   path,
// 		Method: method,
// 	})
// 	return
// }

// Summary string `json:"summary"` na

// GetVersion 获取版本
// func GetVersion(path, method, title, summary string) (version string, err error) {
// 	type versionGenerator struct {
// 		Path    string `json:"path"`
// 		Method  string `json:"method"`
// 		Summary string `json:"summary"`
// 	}
// 	version, err = utils.ObjectMD5Hash(&versionGenerator{
// 		Path:    path,
// 		Method:  method,
// 		Title:   title,
// 		Summary: summary,
// 	})
// 	return
// }

type openAPIParser struct {
	Loader    *openapi3.Loader  // 加载器
	Doc       *openapi3.T       // OpenAPI文档
	DataType  string            // 数据类型
	DataValue interface{}       // 数据值
	SubParser []*openapi3.T     // 子解析器
	Logger    interfaces.Logger // 日志器
}

// LoadOpenAPIMetadata 加载OpenAPI元数据
func LoadOpenAPIMetadata(ctx context.Context, dataType string, dataValue interface{}, logger interfaces.Logger) (metadatas []*APIMetadata, err error) {
	p := &openAPIParser{
		Loader:    openapi3.NewLoader(),
		DataType:  dataType,
		DataValue: dataValue,
		Logger:    logger,
	}
	// ParseAndValidateOpenAPI 解析并校验注入的OpenAPI数据
	err = p.parseAndValidateOpenAPI(ctx)
	if err != nil {
		return
	}
	// 拆分OpenAPI文档
	err = p.splitOpenAPIDocument(ctx)
	if err != nil {
		return
	}
	fmt.Println(len(p.SubParser))
	metadatas = make([]*APIMetadata, 0, len(p.SubParser))
	for _, doc := range p.SubParser {
		// 解析OpenAPI文档
		metadata, err := p.getAPIMetadata(doc)
		if err != nil {
			return nil, err
		}
		metadatas = append(metadatas, metadata)
	}
	data, _ := jsoniter.Marshal(metadatas)
	fmt.Println(string(data))
	return
}

// getAPIMetadatas 获取API元数据
func (p *openAPIParser) getAPIMetadata(doc *openapi3.T) (metadata *APIMetadata, err error) {
	for path, pathItem := range doc.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			// 转换结构化参数
			parameters := make([]*Parameter, 0)
			for _, param := range operation.Parameters {
				ref, paramSchema := schemaRefToMap(param.Value.Schema, doc.Components, make(map[string]bool))
				parameters = append(parameters, &Parameter{
					Name:        param.Value.Name,
					In:          param.Value.In,
					Description: param.Value.Description,
					Required:    param.Value.Required,
					Ref:         ref,
					Schema:      paramSchema,
				})
			}
			// 处理请求体
			var requestBody *RequestBody
			if operation.RequestBody != nil {
				reqContent := make(map[string]Content)
				for contentType, content := range operation.RequestBody.Value.Content {
					ref, contentSchema := schemaRefToMap(content.Schema, doc.Components, make(map[string]bool))
					reqContent[contentType] = Content{
						Ref:     ref,
						Schema:  contentSchema,
						Example: content.Example,
					}
				}
				requestBody = &RequestBody{
					Description: operation.RequestBody.Value.Description,
					Content:     reqContent,
				}
			}

			// 处理响应
			var responses []*Response
			for statusCode, resp := range operation.Responses.Map() {
				respContent := make(map[string]Content)
				for contentType, content := range resp.Value.Content {
					ref, contentSchema := schemaRefToMap(content.Schema, doc.Components, make(map[string]bool))
					respContent[contentType] = Content{
						Ref:     ref,
						Schema:  contentSchema,
						Example: content.Example,
					}
				}
				responses = append(responses, &Response{
					StatusCode:  statusCode,
					Description: *resp.Value.Description,
					Content:     respContent,
				})
			}
			metadata = &APIMetadata{
				Title:       doc.Info.Title,
				Summary:     operation.Summary,
				Description: operation.Description,
				Path:        path,
				Method:      method,
				CreateTime:  time.Now().UnixNano(),
				UpdateTime:  time.Now().UnixNano(),
				IsDeleted:   false,
				Parameters:  parameters,
				RequestBody: requestBody,
				Responses:   responses,
			}
			// metadata.Hash, err = GetHash(metadata.Path, metadata.Method)
			// if err != nil {
			// 	return
			// }
			// metadata.Version, err = GetVersion(metadata.Path, metadata.Method, metadata.Title, metadata.Summary)
			// if err != nil {
			// 	return
			// }
		}
	}
	return
}

// ParseOpenAPIFromData 解析OpenAPI数据
func (p *openAPIParser) parseAndValidateOpenAPI(ctx context.Context) (err error) {
	switch p.DataType {
	case string(ContentDataType):
		p.Doc, err = p.Loader.LoadFromData(p.DataValue.([]byte))
	case string(FileDataType):
		p.Doc, err = p.Loader.LoadFromFile(p.DataValue.(string))
	default:
		err = fmt.Errorf("unsupported data type: %s", p.DataType)
	}
	if err != nil {
		p.Logger.WithContext(ctx).Warnf("Failed to load OpenAPI document: %v", err)
		return
	}
	err = p.Doc.Validate(p.Loader.Context)
	if err != nil {
		p.Logger.WithContext(ctx).Warnf("Failed to validate OpenAPI document: %v", err)
	}
	return
}

// SplitOpenAPIDocument 拆分OpenAPI文档
func (p *openAPIParser) splitOpenAPIDocument(ctx context.Context) (err error) {
	if p.Doc == nil {
		err = fmt.Errorf("OpenAPI document is nil")
		return
	}
	// 将批量导入的OpenAPI分割成多个
	for path, pathItem := range p.Doc.Paths.Map() {
		// 创建新的精简版OpenAPI文档
		newDoc := &openapi3.T{
			OpenAPI: p.Doc.OpenAPI,
			Info:    p.Doc.Info,
			Servers: p.Doc.Servers,
			Components: &openapi3.Components{
				SecuritySchemes: p.Doc.Components.SecuritySchemes,
				Schemas:         make(map[string]*openapi3.SchemaRef),
			},
			Paths:    openapi3.NewPaths(openapi3.WithPath(path, pathItem)),
			Security: p.Doc.Security,
		}
		// 自动收集依赖的schema
		for _, op := range pathItem.Operations() {
			if op.RequestBody != nil {
				collectSchemas(p.Doc.Components, op.RequestBody.Value.Content, newDoc.Components.Schemas, make(map[string]bool))
			}
			for _, resp := range op.Responses.Map() {
				collectSchemas(p.Doc.Components, resp.Value.Content, newDoc.Components.Schemas, make(map[string]bool))
			}
		}
		err = newDoc.Validate(p.Loader.Context)
		if err != nil {
			p.Logger.WithContext(ctx).Warnf("Failed to validate OpenAPI document: %v", err)
			return
		}
		p.SubParser = append(p.SubParser, newDoc)
	}
	return
}

// 收集所有嵌套schema引用（添加visited参数）
func collectSchemas(docComponents *openapi3.Components, content openapi3.Content, schemas map[string]*openapi3.SchemaRef, visited map[string]bool) {
	for _, mediaType := range content {
		if mediaType.Schema != nil {
			// 保留原始schema引用
			if mediaType.Schema.Ref != "" {
				refKey := strings.TrimPrefix(mediaType.Schema.Ref, "#/components/schemas/")
				if _, exists := schemas[refKey]; !exists {
					schemas[refKey] = mediaType.Schema
				}
			}
			traverseSchema(docComponents, mediaType.Schema, schemas, visited)
		}
	}
}

// 递归遍历schema（添加visited参数跟踪引用路径）
func traverseSchema(docComponents *openapi3.Components, schemaRef *openapi3.SchemaRef, schemas map[string]*openapi3.SchemaRef, visited map[string]bool) {
	if schemaRef == nil || schemaRef.Value == nil {
		return
	}

	// 处理直接引用
	if schemaRef.Ref != "" {
		refKey := strings.TrimPrefix(schemaRef.Ref, "#/components/schemas/")
		if visited[refKey] {
			return
		}

		if _, exists := schemas[refKey]; !exists {
			if origSchema, exists := docComponents.Schemas[refKey]; exists {
				visited[refKey] = true
				schemas[refKey] = origSchema
				// 确保保留所有嵌套引用
				if origSchema.Ref != "" {
					nestedRefKey := strings.TrimPrefix(origSchema.Ref, "#/components/schemas/")
					if _, exists := schemas[nestedRefKey]; !exists {
						schemas[nestedRefKey] = origSchema
					}
				}
				traverseSchema(docComponents, origSchema, schemas, visited)
				delete(visited, refKey)
			}
		}
		return
	}

	// 在处理组合结构、对象属性和数组项时传递visited参数
	schema := schemaRef.Value
	for _, s := range schema.AllOf {
		traverseSchema(docComponents, s, schemas, visited)
	}
	for _, s := range schema.AnyOf {
		traverseSchema(docComponents, s, schemas, visited)
	}
	for _, s := range schema.OneOf {
		traverseSchema(docComponents, s, schemas, visited)
	}
	for _, prop := range schema.Properties {
		traverseSchema(docComponents, prop, schemas, visited)
	}
	if schema.Items != nil {
		traverseSchema(docComponents, schema.Items, schemas, visited)
	}
}
