# -*- coding: utf-8 -*-
"""单元测试 - app/common/tool_v2/api_tool_pkg/input.py"""


class TestAPIToolInputHandler:
    """测试 APIToolInputHandler 类"""

    def test_parse_inputs_empty_api_spec(self):
        """测试空 API 规范"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        result = handler._parse_inputs({})

        assert result == {"type": "object", "properties": {}, "required": []}

    def test_parse_inputs_with_parameters(self):
        """测试带参数的 API 规范"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        api_spec = {
            "parameters": [
                {
                    "name": "id",
                    "in": "path",
                    "required": True,
                    "schema": {"type": "string", "description": "ID parameter"},
                },
                {
                    "name": "filter",
                    "in": "query",
                    "required": False,
                    "schema": {"type": "string"},
                },
            ]
        }

        result = handler._parse_inputs(api_spec)

        assert "properties" in result
        assert "id" in result["properties"]
        assert "filter" in result["properties"]
        assert "id" in result["required"]

    def test_parse_inputs_with_request_body(self):
        """测试带请求体的 API 规范"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        api_spec = {
            "request_body": {
                "content": {
                    "application/json": {
                        "schema": {
                            "type": "object",
                            "required": ["name"],
                            "properties": {
                                "name": {"type": "string", "description": "Name"},
                                "age": {"type": "integer"},
                            },
                        }
                    }
                }
            }
        }

        result = handler._parse_inputs(api_spec)

        assert "properties" in result
        assert "name" in result["properties"]
        assert "age" in result["properties"]
        assert "name" in result["required"]

    def test_parse_inputs_with_ref(self):
        """测试带 $ref 引用的参数"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        api_spec = {
            "request_body": {
                "content": {
                    "application/json": {
                        "schema": {"$ref": "#/components/schemas/RequestBody"}
                    }
                }
            },
            "components": {
                "schemas": {
                    "RequestBody": {
                        "type": "object",
                        "properties": {
                            "username": {"type": "string"},
                            "password": {"type": "string"},
                        },
                        "required": ["username"],
                    }
                }
            },
        }

        result = handler._parse_inputs(api_spec)

        assert "username" in result["properties"]
        assert "password" in result["properties"]

    def test_parse_inputs_schema(self):
        """测试解析 inputs_schema"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        inputs_schema = {
            "type": "object",
            "properties": {
                "name": {"type": "string", "description": "Name"},
                "count": {"type": "integer"},
            },
            "required": ["name"],
        }

        result = handler._parse_inputs_schema(inputs_schema)

        assert "name" in result
        assert "count" in result
        assert result["name"]["required"] is True
        assert result["count"].get("required") is None

    def test_resolve_refs_recursively_simple(self):
        """测试递归解析简单引用"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        schema = {"type": "string", "description": "Simple string"}
        api_spec = {}

        result = handler._resolve_refs_recursively(schema, api_spec)

        assert result == {"type": "string", "description": "Simple string"}

    def test_resolve_refs_recursively_with_ref(self):
        """测试递归解析 $ref 引用"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        schema = {"$ref": "#/components/schemas/User"}
        api_spec = {
            "components": {
                "schemas": {
                    "User": {
                        "type": "object",
                        "properties": {
                            "id": {"type": "string"},
                            "name": {"type": "string"},
                        },
                    }
                }
            }
        }

        result = handler._resolve_refs_recursively(schema, api_spec)

        assert "properties" in result
        assert "id" in result["properties"]

    def test_resolve_refs_recursively_nested(self):
        """测试递归解析嵌套结构"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        schema = {
            "type": "object",
            "properties": {
                "user": {"$ref": "#/components/schemas/User"},
                "metadata": {
                    "type": "object",
                    "properties": {"created": {"type": "string"}},
                },
            },
        }
        api_spec = {
            "components": {
                "schemas": {
                    "User": {"type": "object", "properties": {"id": {"type": "string"}}}
                }
            }
        }

        result = handler._resolve_refs_recursively(schema, api_spec)

        assert "user" in result["properties"]
        assert "id" in result["properties"]["user"]["properties"]

    def test_resolve_refs_recursively_circular(self):
        """测试处理循环引用"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()

        # 创建循环引用结构
        circular_schema = {"type": "object"}
        circular_schema["properties"] = {"self": circular_schema}

        result = handler._resolve_refs_recursively(circular_schema, {})

        # 应该能处理循环引用而不崩溃
        assert result is not None

    def test_resolve_refs_with_list(self):
        """测试解析列表中的引用"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        schema = {
            "type": "array",
            "items": [{"$ref": "#/components/schemas/Item"}, {"type": "string"}],
        }
        api_spec = {
            "components": {
                "schemas": {
                    "Item": {
                        "type": "object",
                        "properties": {"name": {"type": "string"}},
                    }
                }
            }
        }

        result = handler._resolve_refs_recursively(schema, api_spec)

        assert "items" in result

    def test_parse_inputs_parameter_without_schema(self):
        """测试无 schema 的参数"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        api_spec = {
            "parameters": [
                {
                    "name": "simple_param",
                    "in": "query",
                    "type": "string",
                    "description": "Simple param without schema",
                }
            ]
        }

        result = handler._parse_inputs(api_spec)

        assert "simple_param" in result["properties"]
        assert result["properties"]["simple_param"]["type"] == "string"

    def test_parse_inputs_parameter_partial_schema(self):
        """测试部分 schema 的参数"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        handler = APIToolInputHandler()
        api_spec = {
            "parameters": [
                {
                    "name": "partial_param",
                    "in": "query",
                    "schema": {"description": "Partial schema"},
                }
            ]
        }

        result = handler._parse_inputs(api_spec)

        assert "partial_param" in result["properties"]
        assert result["properties"]["partial_param"]["type"] == "string"


class TestModuleImports:
    """测试模块导入"""

    def test_import_input_handler(self):
        """测试导入 APIToolInputHandler"""
        from app.common.tool_v2.api_tool_pkg.input import APIToolInputHandler

        assert APIToolInputHandler is not None
