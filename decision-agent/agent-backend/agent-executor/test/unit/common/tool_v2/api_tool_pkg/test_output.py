# -*- coding: utf-8 -*-
"""单元测试 - app/common/tool_v2/api_tool_pkg/output.py"""


class TestAPIToolOutputHandler:
    """测试 APIToolOutputHandler 类"""

    def test_parse_outputs_empty_api_spec(self):
        """测试空 API 规范"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        handler = APIToolOutputHandler()
        result = handler._parse_outputs({})

        assert result == {}

    def test_parse_outputs_no_responses(self):
        """测试无响应定义"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        handler = APIToolOutputHandler()
        result = handler._parse_outputs({"parameters": []})

        assert result == {}

    def test_parse_outputs_with_200_response(self):
        """测试有 200 响应"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        handler = APIToolOutputHandler()
        api_spec = {
            "responses": [
                {
                    "status_code": "200",
                    "content": {
                        "application/json": {
                            "schema": {
                                "type": "object",
                                "properties": {
                                    "result": {
                                        "type": "string",
                                        "description": "Result value",
                                    }
                                },
                            }
                        }
                    },
                }
            ]
        }

        result = handler._parse_outputs(api_spec)

        assert "result" in result
        assert result["result"]["type"] == "string"
        assert result["result"]["description"] == "Result value"

    def test_parse_outputs_with_ref(self):
        """测试带 $ref 引用的响应"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        handler = APIToolOutputHandler()
        api_spec = {
            "responses": [
                {
                    "status_code": "200",
                    "content": {
                        "application/json": {
                            "schema": {"$ref": "#/components/schemas/ResultSchema"}
                        }
                    },
                }
            ],
            "components": {
                "schemas": {
                    "ResultSchema": {
                        "type": "object",
                        "properties": {
                            "id": {"type": "string", "description": "ID"},
                            "name": {"type": "string", "description": "Name"},
                        },
                    }
                }
            },
        }

        result = handler._parse_outputs(api_spec)

        assert "id" in result
        assert "name" in result

    def test_parse_outputs_with_required_fields(self):
        """测试带必填字段的响应"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        handler = APIToolOutputHandler()
        api_spec = {
            "responses": [
                {
                    "status_code": "200",
                    "content": {
                        "application/json": {
                            "schema": {
                                "type": "object",
                                "required": ["id"],
                                "properties": {
                                    "id": {"type": "string"},
                                    "optional": {"type": "string"},
                                },
                            }
                        }
                    },
                }
            ]
        }

        result = handler._parse_outputs(api_spec)

        assert result["id"]["required"] is True
        assert result["optional"]["required"] is False

    def test_parse_outputs_with_example(self):
        """测试带示例值的响应"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        handler = APIToolOutputHandler()
        api_spec = {
            "responses": [
                {
                    "status_code": "200",
                    "content": {
                        "application/json": {
                            "schema": {
                                "type": "object",
                                "properties": {
                                    "status": {"type": "string", "example": "success"}
                                },
                            }
                        }
                    },
                }
            ]
        }

        result = handler._parse_outputs(api_spec)

        assert result["status"]["example"] == "success"

    def test_parse_outputs_ignore_non_200(self):
        """测试忽略非 200 响应"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        handler = APIToolOutputHandler()
        api_spec = {
            "responses": [
                {
                    "status_code": "400",
                    "content": {
                        "application/json": {
                            "schema": {
                                "type": "object",
                                "properties": {"error": {"type": "string"}},
                            }
                        }
                    },
                },
                {
                    "status_code": "500",
                    "content": {
                        "application/json": {
                            "schema": {
                                "type": "object",
                                "properties": {"error": {"type": "string"}},
                            }
                        }
                    },
                },
            ]
        }

        result = handler._parse_outputs(api_spec)

        assert result == {}

    def test_parse_outputs_multiple_content_types(self):
        """测试多种内容类型"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        handler = APIToolOutputHandler()
        api_spec = {
            "responses": [
                {
                    "status_code": "200",
                    "content": {
                        "application/json": {
                            "schema": {
                                "type": "object",
                                "properties": {"json_field": {"type": "string"}},
                            }
                        },
                        "application/xml": {
                            "schema": {
                                "type": "object",
                                "properties": {"xml_field": {"type": "string"}},
                            }
                        },
                    },
                }
            ]
        }

        result = handler._parse_outputs(api_spec)

        # 应该处理所有内容类型
        assert "json_field" in result or "xml_field" in result

    def test_parse_outputs_no_content(self):
        """测试无内容定义的响应"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        handler = APIToolOutputHandler()
        api_spec = {"responses": [{"status_code": "200", "description": "Success"}]}

        result = handler._parse_outputs(api_spec)

        assert result == {}

    def test_parse_outputs_no_schema(self):
        """测试无 schema 的响应"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        handler = APIToolOutputHandler()
        api_spec = {
            "responses": [{"status_code": "200", "content": {"application/json": {}}}]
        }

        result = handler._parse_outputs(api_spec)

        assert result == {}


class TestModuleImports:
    """测试模块导入"""

    def test_import_output_handler(self):
        """测试导入 APIToolOutputHandler"""
        from app.common.tool_v2.api_tool_pkg.output import APIToolOutputHandler

        assert APIToolOutputHandler is not None
