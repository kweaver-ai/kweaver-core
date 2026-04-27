package parsers

import (
	"context"
	"fmt"
	"os"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
	. "github.com/smartystreets/goconvey/convey"
)

// 测试Function转换成OpenAPI3.0规范的Schema

func TestFunctionToOpenAPISchema(t *testing.T) {
	Convey("TestFunctionToOpenAPISchema: 传参为空时", t, func() {
		input := &interfaces.FunctionInput{
			Name:        "传参为空时",
			Description: "test function",
			Inputs:      []*interfaces.ParameterDef{},
			Outputs:     []*interfaces.ParameterDef{},
		}
		schema := convertToPathItemContent(input)
		So(schema, ShouldNotBeNil)
		// 输出到文件中
		data, _ := jsoniter.Marshal(schema)
		filename := input.Name + ".json"
		err := os.WriteFile(filename, data, 0644)
		So(err, ShouldBeNil)
		t.Logf("Successfully wrote API metadata to %s", filename)
		_ = os.Remove(filename)
	})
	Convey("TestFunctionToOpenAPISchema: 传参为nil时", t, func() {
		input := &interfaces.FunctionInput{
			Name:        "传参为nil时",
			Description: "test function",
			Inputs: []*interfaces.ParameterDef{
				{
					Name:        "id",
					Description: "任务ID",
					Type:        "string",
					Required:    true,
				},
				{
					Name:        "params",
					Description: "任务参数",
					Type:        "object",
					Required:    true,
				},
			},
			Outputs: []*interfaces.ParameterDef{
				{
					Name:        "status",
					Description: "任务执行状态",
					Type:        "string",
					Required:    true,
				},
				{
					Name:        "output",
					Description: "任务执行结果",
					Type:        "object",
					Required:    true,
				},
			},
		}
		schema := convertToPathItemContent(input)
		So(schema, ShouldNotBeNil)
		// 输出到文件中
		data, _ := jsoniter.Marshal(schema)
		filename := input.Name + ".json"
		err := os.WriteFile(filename, data, 0644)
		So(err, ShouldBeNil)
		t.Logf("Successfully wrote API metadata to %s", filename)
		_ = os.Remove(filename)
	})
}

func TestCheckHandler(t *testing.T) {
	Convey("TestCheckHandler: 检查是否包含入口函数handler", t, func() {
		code := `def handler(event):
    return {"status": "success"}`
		err := checkRegexpHandler(context.Background(), code)
		So(err, ShouldBeNil)
		code2 := `def handler(event, context):
    return {"status": "success"}`
		err = checkRegexpHandler(context.Background(), code2)
		So(err, ShouldBeNil)
		code3 := `def main(event, context):
    return {"status": "success"}`
		err = checkRegexpHandler(context.Background(), code3)
		So(err, ShouldNotBeNil)
		fmt.Println(err)
		code4 := `def handler(text,event):
    return {"status": "success"}`
		err = checkRegexpHandler(context.Background(), code4)
		So(err, ShouldBeNil)
	})
}

func TestCreateParameterSchema(t *testing.T) {
	Convey("TestCreateParameterSchema: 平铺参数定义", t, func() {
		functionInputJSON := `{
	    "inputs": [
	        {
	            "name": "content",
	            "type": "string",
	            "description": "敏感词库内容，当load_mode=content时必填",
	            "required": false
	        },
	        {
	            "name": "load_mode",
	            "type": "string",
	            "description": "敏感词库加载模式（url/content）",
	            "required": true
	        },
	        {
	            "name": "text",
	            "type": "string",
	            "description": "待识别文本内容",
	            "required": true
	        },
	        {
	            "name": "url",
	            "type": "string",
	            "description": "敏感词库下载地址，当load_mode=url时必填",
	            "required": false
	        }
	    ],
		"name": "测试简单格式参数转换",
	    "outputs": [],
	    "code": "def handler(event) -> dict: \n    return {'is_pass': True}",
	    "script_type": "python"
	}`
		functionInput := &interfaces.FunctionInput{}
		err := utils.StringToObject(functionInputJSON, functionInput)
		So(err, ShouldBeNil)
		schema := convertToPathItemContent(functionInput)
		So(schema, ShouldNotBeNil)
		// 输出到文件中
		data, _ := jsoniter.Marshal(schema)
		filename := functionInput.Name + ".json"
		err = os.WriteFile(filename, data, 0644)
		So(err, ShouldBeNil)
		t.Logf("Successfully wrote API metadata to %s", filename)
		_ = os.Remove(filename)
	})
	Convey("TestCreateParameterSchema: 嵌套参数定义", t, func() {
		functionInputJSON := `{
    "inputs": [
        {
            "name": "content",
            "type": "object",
            "description": "请求对象",
            "required": false,
            "sub_parameters": [
                {
                    "name": "file_info",
                    "type": "object",
                    "description": "文件信息",
                    "required": true,
                    "sub_parameters": [
                        {
                            "name": "name",
                            "type": "string",
                            "description": "文件名",
                            "required": true
                        },
                        {
                            "name": "size",
                            "type": "number",
                            "description": "文件名",
                            "required": true
                        },
                        {
                            "name": "is_file",
                            "type": "boolean",
                            "description": "是否是文件",
                            "required": false
                        }
                    ]
                },
                {
                    "name": "file_list",
                    "type": "array",
                    "description": "列表对象",
                    "required": false,
					"sub_parameters": [
                        {
                            "type": "string"
                        }
                    ]
                }
            ]
        }
    ],
    "name": "测试嵌套参数转换",
    "outputs": [],
    "code": "def handler(event) -> dict: \n    return {'is_pass': True}",
    "script_type": "python"
}`
		functionInput := &interfaces.FunctionInput{}
		err := utils.StringToObject(functionInputJSON, functionInput)
		So(err, ShouldBeNil)
		schema := convertToPathItemContent(functionInput)
		So(schema, ShouldNotBeNil)
		// 输出到文件中
		data, _ := jsoniter.Marshal(schema)
		filename := functionInput.Name + ".json"
		err = os.WriteFile(filename, data, 0644)
		So(err, ShouldBeNil)
		t.Logf("Successfully wrote API metadata to %s", filename)
		_ = os.Remove(filename)
	})
}
