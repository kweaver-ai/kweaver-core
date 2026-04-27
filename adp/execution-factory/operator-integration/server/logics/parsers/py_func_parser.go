package parsers

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// pythonFunctionParser Python 函数解析器
type pythonFunctionParser struct {
	Logger    interfaces.Logger
	Validator interfaces.Validator
}

func (p *pythonFunctionParser) Type() interfaces.MetadataType {
	return interfaces.MetadataTypeFunc
}

func (p *pythonFunctionParser) validate(ctx context.Context, inputValue any) (input *interfaces.FunctionInput, err error) {
	input, ok := inputValue.(*interfaces.FunctionInput)
	if !ok {
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "input value is not *interfaces.FunctionInput")
		return
	}
	if input == nil {
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "input content is empty")
		return
	}
	// Code 校验
	if input.Code == "" {
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "python function code is empty")
		return
	}
	// 校验参数定义
	err = p.Validator.ValidatorStruct(ctx, input)
	if err != nil {
		return
	}
	if input.Inputs == nil {
		input.Inputs = make([]*interfaces.ParameterDef, 0)
	}
	for _, param := range input.Inputs {
		err = p.Validator.VisitorParameterDef(ctx, param)
		if err != nil {
			return
		}
	}
	return
}

// 检查是否包含入口函数handler
func checkRegexpHandler(ctx context.Context, code string) (err error) {
	// 使用正则表达式检查是否包含 handler 函数定义
	pattern := `def\s+handler\s*\(`
	matched, err := regexp.MatchString(pattern, code)
	if err != nil {
		return errors.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("check handler regexp failed: %v", err))
	}
	if !matched {
		// 必须包含 handler 函数定义
		return errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtFunctionNoHandlerFound, "python function must have a handler(event) function")
	}
	return nil
}

// func checAstkHandler(ctx context.Context, code string) (err error) {
// 	// 解析Python代码
// 	mod, err := parser.ParseString(code, py.ExecMode)
// 	if err != nil {
// 		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("parse python code failed: %v", err))
// 		return
// 	}
// 	// 检查是否包含入口函数handler
// 	var hasHandler bool
// 	ast.Walk(mod, func(node ast.Ast) bool {
// 		n, ok := node.(*ast.FunctionDef)
// 		if ok && n.Name == "handler" {
// 			hasHandler = true
// 		}
// 		return true
// 	})
// 	if !hasHandler {
// 		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "python function must have a handler function")
// 	}
// 	return
// }

// Parse 解析 Python 函数
func (p *pythonFunctionParser) Parse(ctx context.Context, inputValue any) (metadatas []interfaces.IMetadataDB, err error) {
	// 记录可观测性
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	input, err := p.validate(ctx, inputValue)
	if err != nil {
		return nil, err
	}
	err = checkRegexpHandler(ctx, input.Code)
	if err != nil {
		return nil, err
	}
	pathItem := convertToPathItemContent(input)
	desc := pathItem.Description
	if desc == "" {
		desc = pathItem.Summary
	}
	metadatas = make([]interfaces.IMetadataDB, 0)
	metadataDB := &model.FunctionMetadataDB{
		ScriptType:      string(input.ScriptType),
		Code:            input.Code,
		Dependencies:    utils.ObjectToJSON(input.Dependencies),
		DependenciesURL: input.DependenciesURL,
		Summary:         pathItem.Summary,
		Description:     desc,
		Path:            pathItem.Path,
		ServerURL:       pathItem.ServerURL,
		Method:          pathItem.Method,
		APISpec:         pathItem.APISpec.ToJSON(),
	}
	metadatas = append(metadatas, metadataDB)
	return
}

// GetAllContent 获取所有内容
func (p *pythonFunctionParser) GetAllContent(ctx context.Context, inputValue any) (content any, err error) {
	input, err := p.validate(ctx, inputValue)
	if err != nil {
		return nil, err
	}
	// 解析Python代码
	mod, err := parser.ParseString(input.Code, "exec")
	if err != nil {
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, fmt.Sprintf("解析Python代码失败: %v", err))
		return
	} // 检查是否包含入口函数handler
	var hasHandler bool
	ast.Walk(mod, func(node ast.Ast) bool {
		n, ok := node.(*ast.FunctionDef)
		if ok && n.Name == "handler" {
			hasHandler = true
		}
		return true
	})
	if !hasHandler {
		err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "python function must have a handler function")
		return
	}
	content = convertToPathItemContent(input)
	return
}

// 将input\output转换成 PathItemContent
func convertToPathItemContent(input *interfaces.FunctionInput) (result *interfaces.PathItemContent) {
	result = &interfaces.PathItemContent{
		Path:        interfaces.GetAOIFuncExecPath(),
		Method:      http.MethodPost,
		ServerURL:   interfaces.AOIServerURL,
		Summary:     input.Name,
		Description: input.Description,
		APISpec:     &interfaces.APISpec{},
	}
	// 添加超时时间参数
	result.APISpec.Parameters = createParameter()
	// 根据处理输入参数创建请求体
	result.APISpec.RequestBody = createRequestBody(input.Inputs)
	// 处理输出参数
	result.APISpec.Responses = createResponseBody(input.Outputs)
	return
}

// 构造Parameter参数
func createParameter() []*interfaces.Parameter {
	parameters := make([]*interfaces.Parameter, 0)
	// 超时时间
	timeoutParam := &interfaces.Parameter{
		Name:        "timeout",
		In:          "query",
		Description: "函数执行超时时间，单位毫秒",
		Required:    false,
		Schema: openapi3.NewSchemaRef("", &openapi3.Schema{
			Type:        &openapi3.Types{openapi3.TypeNumber},
			Description: "函数执行超时时间，单位毫秒",
		}),
	}
	parameters = append(parameters, timeoutParam)
	return parameters
}

// 构造请求体结构
func createRequestBody(inputs []*interfaces.ParameterDef) *interfaces.RequestBody {
	// 创建schema定义
	requestBodySchema := openapi3.NewObjectSchema()
	if len(inputs) > 0 {
		for _, input := range inputs {
			propertySchema := createParameterSchema(input)
			requestBodySchema.Properties[input.Name] = openapi3.NewSchemaRef("", propertySchema)
			// 设置必填字段
			if input.Required {
				requestBodySchema.Required = append(requestBodySchema.Required, input.Name)
			}
		}
	}
	// 创建请求体
	requestBody := &interfaces.RequestBody{
		Description: "函数输入参数",
		Content:     openapi3.NewContentWithJSONSchema(requestBodySchema),
		Required:    true,
	}
	return requestBody
}

// 处理输出参数
// 根据处理输出参数创建响应体
func createResponseBody(outputs []*interfaces.ParameterDef) []*interfaces.Response {
	// 创建schema定义
	responseSchema := openapi3.NewObjectSchema()
	responseSchema.Properties["stdout"] = openapi3.NewSchemaRef("", &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeString},
		Description: "标准输出流内容",
	})
	responseSchema.Properties["stderr"] = openapi3.NewSchemaRef("", &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeString},
		Description: "标准错误流内容",
	})

	resultSchema := &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeObject},
		Description: "Handler 函数返回的业务结果: any or null",
		Properties:  make(openapi3.Schemas),
	}
	for _, output := range outputs {
		propertySchema := createParameterSchema(output)
		resultSchema.Properties[output.Name] = openapi3.NewSchemaRef("", propertySchema)
		// 设置必填字段
		if output.Required {
			resultSchema.Required = append(resultSchema.Required, output.Name)
		}
	}
	responseSchema.Properties["result"] = openapi3.NewSchemaRef("", resultSchema)
	// 添加指标
	metricsSchema := &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeObject},
		Description: "指标",
		Properties:  make(openapi3.Schemas),
	}
	metricsSchema.Properties["duration_ms"] = openapi3.NewSchemaRef("", &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeNumber},
		Description: "执行总耗时（毫秒）",
	})
	metricsSchema.Properties["memory_peak_mb"] = openapi3.NewSchemaRef("", &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeNumber},
		Description: "峰值内存占用（MB）",
	})
	metricsSchema.Properties["cpu_time_ms"] = openapi3.NewSchemaRef("", &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeNumber},
		Description: "CPU 时间（毫秒）",
	})
	responseSchema.Properties["metrics"] = openapi3.NewSchemaRef("", metricsSchema)
	// 添加错误响应体
	errSchema := &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeObject},
		Description: "失败详情",
		Properties:  map[string]*openapi3.SchemaRef{},
	}
	errSchema.Properties["code"] = openapi3.NewSchemaRef("", &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeString},
		Description: "错误码",
	})
	errSchema.Properties["description"] = openapi3.NewSchemaRef("", &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeString},
		Description: "错误描述",
	})
	errSchema.Properties["detail"] = openapi3.NewSchemaRef("", &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeObject},
		Description: "错误详情",
	})
	errSchema.Properties["solution"] = openapi3.NewSchemaRef("", &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeString},
		Description: "错误解决办法",
	})
	errSchema.Properties["link"] = openapi3.NewSchemaRef("", &openapi3.Schema{
		Type:        &openapi3.Types{openapi3.TypeString},
		Description: "错误链接",
	})
	// 创建响应体
	responseBody := []*interfaces.Response{
		{
			StatusCode:  "200",
			Description: "成功",
			Content:     openapi3.NewContentWithJSONSchema(responseSchema),
		},
		{
			StatusCode:  "400",
			Description: "参数校验失败",
			Content:     openapi3.NewContentWithJSONSchema(errSchema),
		},
		{
			StatusCode:  "404",
			Description: "资源不存在",
			Content:     openapi3.NewContentWithJSONSchema(errSchema),
		},
		{
			StatusCode:  "500",
			Description: "函数执行失败",
			Content:     openapi3.NewContentWithJSONSchema(errSchema),
		},
	}
	return responseBody
}

// mapTypeToOpenAPI 将参数类型映射到OpenAPI类型
func mapTypeToOpenAPI(paramType string) *openapi3.Types {
	switch strings.ToLower(paramType) {
	case "string":
		return &openapi3.Types{openapi3.TypeString}
	case "int", "integer", "number":
		return &openapi3.Types{openapi3.TypeNumber}
	case "float", "double":
		return &openapi3.Types{openapi3.TypeNumber}
	case "bool", "boolean":
		return &openapi3.Types{openapi3.TypeBoolean}
	case "array":
		return &openapi3.Types{openapi3.TypeArray}
	case "object":
		return &openapi3.Types{openapi3.TypeObject}
	default:
		return &openapi3.Types{openapi3.TypeString}
	}
}

func createParameterSchema(param *interfaces.ParameterDef) *openapi3.Schema {
	if param.Description == "" {
		param.Description = param.Name
	}
	propertySchema := &openapi3.Schema{
		Type:        mapTypeToOpenAPI(string(param.Type)),
		Description: param.Description,
	}

	// 设置默认值
	if param.Default != nil {
		propertySchema.Default = param.Default
	}
	// 设置枚举值
	if len(param.Enum) > 0 {
		propertySchema.Enum = param.Enum
	}
	// 设置示例值
	if param.Example != nil {
		propertySchema.Example = param.Example
	}
	// 处理嵌套参数
	if len(param.SubParameters) > 0 {
		switch param.Type {
		case interfaces.ParameterTypeObject:
			// Object类型：SubParameters定义对象的属性
			propertySchema.Properties = make(openapi3.Schemas)
			for _, subParam := range param.SubParameters {
				subPropertySchema := createParameterSchema(subParam)
				propertySchema.Properties[subParam.Name] = openapi3.NewSchemaRef("", subPropertySchema)
				// 子对象的必填字段需要添加到父对象的Required数组中
				if subParam.Required {
					propertySchema.Required = append(propertySchema.Required, subParam.Name)
				}
			}

		case interfaces.ParameterTypeArray:
			// Array类型：SubParameters只包含一个元素，定义数组元素的结构
			subParam := param.SubParameters[0]
			if subParam.Description == "" {
				subParam.Description = param.Description
			}
			itemsSchema := createParameterSchema(subParam)
			propertySchema.Items = openapi3.NewSchemaRef("", itemsSchema)

		case interfaces.ParameterTypeString, interfaces.ParameterTypeNumber, interfaces.ParameterTypeBoolean:
		}
	}
	return propertySchema
}
