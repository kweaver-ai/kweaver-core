package mcp

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestExtractParameters 测试参数提取功能
func TestExtractParameters(t *testing.T) {
	Convey("TestExtractParameters: 测试参数提取和描述", t, func() {
		converter := NewSimpleConverter()

		Convey("提取 header 参数并添加描述", func() {
			params := []Parameter{
				{
					Name:        "Authorization",
					In:          "header",
					Description: "认证令牌",
					Required:    true,
					Schema: map[string]any{
						"type": "string",
					},
				},
				{
					Name:     "Content-Type",
					In:       "header",
					Required: false,
					Schema: map[string]any{
						"type": "string",
					},
				},
			}

			result := converter.extractParameters(params, "header", "HTTP 请求头参数")

			So(result, ShouldNotBeNil)
			So(result["type"], ShouldEqual, "object")
			So(result["description"], ShouldEqual, "HTTP 请求头参数")

			props := result["properties"].(map[string]any)
			So(props, ShouldContainKey, "Authorization")
			So(props, ShouldContainKey, "Content-Type")

			// 验证参数描述
			authSchema := props["Authorization"].(map[string]any)
			So(authSchema["description"], ShouldEqual, "认证令牌")

			// 验证必填字段
			required := result["required"].([]string)
			So(required, ShouldContain, "Authorization")
			So(required, ShouldNotContain, "Content-Type")
		})

		Convey("提取 query 参数并添加描述", func() {
			params := []Parameter{
				{
					Name:        "page",
					In:          "query",
					Description: "页码",
					Required:    false,
					Schema: map[string]any{
						"type": "integer",
					},
				},
			}

			result := converter.extractParameters(params, "query", "URL 查询字符串参数")

			So(result, ShouldNotBeNil)
			So(result["description"], ShouldEqual, "URL 查询字符串参数")

			props := result["properties"].(map[string]any)
			So(props, ShouldContainKey, "page")
		})

		Convey("提取 path 参数并添加描述", func() {
			params := []Parameter{
				{
					Name:        "id",
					In:          "path",
					Description: "资源ID",
					Required:    true,
					Schema: map[string]any{
						"type": "string",
					},
				},
			}

			result := converter.extractParameters(params, "path", "URL 路径参数")

			So(result, ShouldNotBeNil)
			So(result["description"], ShouldEqual, "URL 路径参数")

			props := result["properties"].(map[string]any)
			So(props, ShouldContainKey, "id")

			required := result["required"].([]string)
			So(required, ShouldContain, "id")
		})

		Convey("参数列表为空时返回 nil", func() {
			params := []Parameter{}
			result := converter.extractParameters(params, "header", "HTTP 请求头参数")
			So(result, ShouldBeNil)
		})

		Convey("无匹配类型参数时返回 nil", func() {
			params := []Parameter{
				{
					Name: "Authorization",
					In:   "header",
					Schema: map[string]any{
						"type": "string",
					},
				},
			}
			result := converter.extractParameters(params, "query", "URL 查询字符串参数")
			So(result, ShouldBeNil)
		})
	})
}

// TestExtractRequestBody 测试请求体提取功能
func TestExtractRequestBody(t *testing.T) {
	Convey("TestExtractRequestBody: 测试请求体提取和默认描述", t, func() {
		converter := NewSimpleConverter()

		Convey("提取 JSON 请求体，无 description 时补充默认描述", func() {
			reqBody := &RequestBody{
				Content: map[string]Content{
					"application/json": {
						Schema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"name": map[string]any{
									"type": "string",
								},
							},
						},
					},
				},
			}

			result := converter.extractRequestBody(reqBody)

			So(result, ShouldNotBeNil)
			So(result["description"], ShouldEqual, "Request Body 参数")
			So(result["type"], ShouldEqual, "object")
		})

		Convey("提取 JSON 请求体，已有 description 时保留原描述", func() {
			reqBody := &RequestBody{
				Description: "用户信息",
				Content: map[string]Content{
					"application/json": {
						Schema: map[string]any{
							"type":        "object",
							"description": "用户对象",
							"properties": map[string]any{
								"name": map[string]any{
									"type": "string",
								},
							},
						},
					},
				},
			}

			result := converter.extractRequestBody(reqBody)

			So(result, ShouldNotBeNil)
			// 应该保留原有的 description
			So(result["description"], ShouldEqual, "用户对象")
		})

		Convey("RequestBody 为 nil 时返回 nil", func() {
			result := converter.extractRequestBody(nil)
			So(result, ShouldBeNil)
		})

		Convey("无 application/json 内容时返回 nil", func() {
			reqBody := &RequestBody{
				Content: map[string]Content{
					"application/xml": {
						Schema: map[string]any{
							"type": "object",
						},
					},
				},
			}

			result := converter.extractRequestBody(reqBody)
			So(result, ShouldBeNil)
		})
	})
}

// TestConvertSimpleOpenAPI 测试完整转换流程
func TestConvertSimpleOpenAPI(t *testing.T) {
	Convey("TestConvertSimpleOpenAPI: 测试完整 OpenAPI 到 MCP JSON Schema 转换", t, func() {
		converter := NewSimpleConverter()

		Convey("包含所有参数类型的完整转换", func() {
			simple := &SimpleOpenAPI{
				Parameters: []Parameter{
					{
						Name:        "Authorization",
						In:          "header",
						Description: "认证令牌",
						Required:    true,
						Schema: map[string]any{
							"type": "string",
						},
					},
					{
						Name:        "page",
						In:          "query",
						Description: "页码",
						Required:    false,
						Schema: map[string]any{
							"type": "integer",
						},
					},
					{
						Name:        "id",
						In:          "path",
						Description: "资源ID",
						Required:    true,
						Schema: map[string]any{
							"type": "string",
						},
					},
				},
				RequestBody: &RequestBody{
					Content: map[string]Content{
						"application/json": {
							Schema: map[string]any{
								"type": "object",
								"properties": map[string]any{
									"name": map[string]any{
										"type": "string",
									},
								},
							},
						},
					},
				},
			}

			result := converter.convertSimpleOpenAPI(simple)

			So(result.Success, ShouldBeTrue)
			So(result.Data, ShouldNotBeNil)
			So(result.Data["type"], ShouldEqual, "object")

			// 验证 properties
			properties := result.Data["properties"].(map[string]any)
			So(properties, ShouldContainKey, "header")
			So(properties, ShouldContainKey, "query")
			So(properties, ShouldContainKey, "path")
			So(properties, ShouldContainKey, "body")

			// 验证各参数组的 description
			header := properties["header"].(map[string]any)
			So(header["description"], ShouldEqual, "HTTP 请求头参数")

			query := properties["query"].(map[string]any)
			So(query["description"], ShouldEqual, "URL 查询字符串参数")

			path := properties["path"].(map[string]any)
			So(path["description"], ShouldEqual, "URL 路径参数")

			body := properties["body"].(map[string]any)
			So(body["description"], ShouldEqual, "Request Body 参数")
		})

		Convey("验证 $defs 和 $ref 转换", func() {
			simple := &SimpleOpenAPI{
				RequestBody: &RequestBody{
					Content: map[string]Content{
						"application/json": {
							Schema: map[string]any{
								"$ref": "#/components/schemas/User",
							},
						},
					},
				},
				Components: &Components{
					Schemas: map[string]map[string]any{
						"User": {
							"type": "object",
							"properties": map[string]any{
								"id": map[string]any{
									"type": "integer",
								},
							},
						},
					},
				},
			}

			result := converter.convertSimpleOpenAPI(simple)

			So(result.Success, ShouldBeTrue)

			// 验证 $defs 存在
			defs := result.Data["$defs"].(map[string]any)
			So(defs, ShouldContainKey, "User")

			// 验证 $ref 路径转换
			properties := result.Data["properties"].(map[string]any)
			body := properties["body"].(map[string]any)
			So(body["$ref"], ShouldEqual, "#/$defs/User")
		})

		Convey("空参数时正常处理", func() {
			simple := &SimpleOpenAPI{}

			result := converter.convertSimpleOpenAPI(simple)

			So(result.Success, ShouldBeTrue)
			So(result.Data["type"], ShouldEqual, "object")

			properties := result.Data["properties"].(map[string]any)
			So(len(properties), ShouldEqual, 0)
		})
	})
}

// TestConvertFromBytes 测试从字节数组转换
func TestConvertFromBytes(t *testing.T) {
	Convey("TestConvertFromBytes: 测试从字节数组解析", t, func() {
		converter := NewSimpleConverter()

		Convey("有效 JSON 解析", func() {
			jsonData := []byte(`{
				"parameters": [
					{
						"name": "id",
						"in": "path",
						"description": "ID",
						"required": true,
						"schema": {"type": "string"}
					}
				]
			}`)

			result := converter.ConvertFromBytes(jsonData)

			So(result.Success, ShouldBeTrue)
			So(result.Data, ShouldNotBeNil)
		})

		Convey("无效 JSON 解析失败", func() {
			jsonData := []byte(`{invalid json}`)

			result := converter.ConvertFromBytes(jsonData)

			So(result.Success, ShouldBeFalse)
			So(result.Error, ShouldContainSubstring, "解析JSON失败")
		})
	})
}

// TestConvertFromString 测试从字符串转换
func TestConvertFromString(t *testing.T) {
	Convey("TestConvertFromString: 测试从字符串解析", t, func() {
		converter := NewSimpleConverter()

		Convey("有效 JSON 字符串解析", func() {
			jsonStr := `{
				"parameters": [
					{
						"name": "token",
						"in": "header",
						"required": true,
						"schema": {"type": "string"}
					}
				]
			}`

			result := converter.ConvertFromString(jsonStr)

			So(result.Success, ShouldBeTrue)
			So(result.Data, ShouldNotBeNil)
		})

		Convey("无效 JSON 字符串解析失败", func() {
			jsonStr := `not a json`

			result := converter.ConvertFromString(jsonStr)

			So(result.Success, ShouldBeFalse)
			So(result.Error, ShouldNotBeEmpty)
		})
	})
}

// TestToJSONString 测试转换为 JSON 字符串
func TestToJSONString(t *testing.T) {
	Convey("TestToJSONString: 测试转换结果序列化", t, func() {
		converter := NewSimpleConverter()

		Convey("成功结果转换为 JSON 字符串", func() {
			result := &Result{
				Success: true,
				Data: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{
							"type": "string",
						},
					},
				},
			}

			jsonStr, err := converter.ToJSONString(result)

			So(err, ShouldBeNil)
			So(jsonStr, ShouldContainSubstring, "\"type\": \"object\"")
			So(jsonStr, ShouldContainSubstring, "properties")
		})

		Convey("失败结果转换报错", func() {
			result := &Result{
				Success: false,
				Error:   "转换失败",
			}

			jsonStr, err := converter.ToJSONString(result)

			So(err, ShouldNotBeNil)
			So(jsonStr, ShouldBeEmpty)
			So(err.Error(), ShouldContainSubstring, "转换失败")
		})
	})
}

// TestGetSchemaInfo 测试获取 Schema 信息
func TestGetSchemaInfo(t *testing.T) {
	Convey("TestGetSchemaInfo: 测试获取 Schema 摘要信息", t, func() {
		converter := NewSimpleConverter()

		Convey("获取成功结果的信息", func() {
			result := &Result{
				Success: true,
				Data: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"header": map[string]any{},
						"query":  map[string]any{},
						"body":   map[string]any{},
					},
					"$defs": map[string]any{
						"User":    map[string]any{},
						"Product": map[string]any{},
					},
				},
			}

			info := converter.GetSchemaInfo(result)

			So(info["type"], ShouldEqual, "object")
			So(info["property_count"], ShouldEqual, 3)
			So(info["schema_count"], ShouldEqual, 2)

			properties := info["properties"].([]string)
			So(properties, ShouldContain, "header")
			So(properties, ShouldContain, "query")
			So(properties, ShouldContain, "body")

			schemas := info["schemas"].([]string)
			So(schemas, ShouldContain, "User")
			So(schemas, ShouldContain, "Product")
		})

		Convey("获取失败结果的信息", func() {
			result := &Result{
				Success: false,
				Error:   "测试错误",
			}

			info := converter.GetSchemaInfo(result)

			So(info["error"], ShouldEqual, "测试错误")
		})

		Convey("无 $defs 时的信息", func() {
			result := &Result{
				Success: true,
				Data: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			}

			info := converter.GetSchemaInfo(result)

			So(info["schema_count"], ShouldEqual, 0)
			So(info["schemas"], ShouldResemble, []string{})
		})
	})
}
