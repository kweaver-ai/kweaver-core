from typing import Any, Dict


# Import from output module using relative import
from .output import APIToolOutputHandler


class APIToolInputHandler(APIToolOutputHandler):
    def _resolve_refs_recursively(self, schema, api_spec, visited=None):
        """递归解析schema中的所有$ref引用，包括嵌套字段中的引用"""
        if visited is None:
            visited = set()

        if not isinstance(schema, dict):
            return schema

        # 防止循环引用
        schema_id = id(schema)

        if schema_id in visited:
            return schema

        visited.add(schema_id)

        # 处理直接的$ref
        if "$ref" in schema:
            ref_path = schema["$ref"]

            if ref_path.startswith("#/components/schemas/"):
                schema_name = ref_path.split("/")[-1]
                schemas = api_spec.get("components", {}).get("schemas", {})

                if schema_name in schemas:
                    # 递归解析引用的schema
                    resolved_schema = self._resolve_refs_recursively(
                        schemas[schema_name], api_spec, visited
                    )

                    return resolved_schema

        # 递归处理所有嵌套的字典和列表
        result = {}
        for key, value in schema.items():
            if isinstance(value, dict):
                result[key] = self._resolve_refs_recursively(value, api_spec, visited)

            elif isinstance(value, list):
                result[key] = [
                    (
                        self._resolve_refs_recursively(item, api_spec, visited)
                        if isinstance(item, dict)
                        else item
                    )
                    for item in value
                ]
            else:
                result[key] = value

        return result

    def _parse_inputs(self, api_spec):
        """解析工具输入参数, 用于提供给大模型
        如果self.tool_map_list中设置为不暴露给大模型,则不进行展示"""
        """
        "api_spec": {
            "parameters": [
                {
                    "name": "Authorization",
                    "in": "header",
                    "description": "认证信息",
                    "required": true
                },
                {
                    "name": "box_id",
                    "in": "path",
                    "description": "工具箱ID",
                    "required": true
                }
            ],
            "request_body": {
                "description": "",
                "content": {
                    "application/json": {
                        "schema": {
                            "$ref": "#/components/schemas/UpdateToolBoxRequest"
                        }
                    }
                },
                "required": false
            },
            "responses": [
                {
                    "status_code": "404",
                    "description": "Not Found",
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            }
                        }
                    }
                },
                {
                    "status_code": "500",
                    "description": "Internal Server Error",
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            }
                        }
                    }
                },
                {
                    "status_code": "200",
                    "description": "更新工具箱成功",
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/UpdateToolBoxResult"
                            }
                        }
                    }
                },
                {
                    "status_code": "400",
                    "description": "Bad Request",
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            }
                        }
                    }
                },
                {
                    "status_code": "401",
                    "description": "Unauthorized",
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/Error"
                            }
                        }
                    }
                }
            ],
            "schemas": {
                "UpdateToolBoxRequest": {
                    "type": "object",
                    "description": "更新工具箱请求",
                    "required": [
                        "box_name",
                        "box_desc",
                        "box_svc_url",
                        "box_icon",
                        "box_category"
                    ],
                    "properties": {
                        "box_desc": {
                            "type": "string",
                            "description": "工具箱描述"
                        },
                        "box_icon": {
                            "description": "工具箱图标",
                            "type": "string"
                        },
                        "box_name": {
                            "type": "string",
                            "description": "工具箱名称"
                        },
                        "box_svc_url": {
                            "type": "string",
                            "description": "工具箱服务地址"
                        },
                        "extend_info": {
                            "type": "object",
                            "description": "扩展信息"
                        },
                        "box_category": {
                            "description": "工具箱分类",
                            "type": "string"
                        }
                    }
                },
                "Error": {
                    "type": "object",
                    "description": "错误信息",
                    "properties": {
                        "description": {
                            "type": "string",
                            "description": "错误信息"
                        },
                        "detail": {
                            "type": "object",
                            "description": "错误详情"
                        },
                        "link": {
                            "description": "错误链接",
                            "type": "string"
                        },
                        "solution": {
                            "type": "string",
                            "description": "错误解决方案"
                        },
                        "code": {
                            "type": "string",
                            "description": "错误码"
                        }
                    }
                },
                "UpdateToolBoxResult": {
                    "properties": {
                        "box_id": {
                            "description": "工具箱ID",
                            "type": "string"
                        }
                    },
                    "type": "object",
                    "description": "工具箱结果"
                }
            },
            "callbacks": null,
            "security": null,
            "tags": null,
            "external_docs": null
        }
        """

        inputs = {}
        required = []

        # 1. 解析 parameters 中的输入参数
        if api_spec and "parameters" in api_spec:
            for param in api_spec["parameters"]:
                param_name = param.get("name", "")
                param_schema = param.get("schema", {})

                if not param_schema:
                    param_schema = {
                        "type": param.get("type", "string"),
                        "description": param.get("description", ""),
                    }
                else:
                    # 检查是否存在 type 和 description，不存在则赋值
                    param_schema["type"] = param_schema.get(
                        "type", param.get("type", "string")
                    )
                    param_schema["description"] = param_schema.get(
                        "description", param.get("description", "")
                    )

                # 递归解析schema中的$ref引用
                resolved_schema = self._resolve_refs_recursively(param_schema, api_spec)
                inputs[param_name] = resolved_schema

                if param.get("required", False):
                    required.append(param_name)

        # 2. 解析 request_body 中的输入参数
        if api_spec and "request_body" in api_spec:
            request_body = api_spec["request_body"]

            for content_type, content_info in request_body["content"].items():
                if "schema" in content_info:
                    schema = content_info["schema"]

                    # 递归解析schema中的所有$ref引用
                    resolved_schema = self._resolve_refs_recursively(schema, api_spec)

                    # 解析schema的properties
                    if "properties" in resolved_schema:
                        required_fields = resolved_schema.get("required", [])

                        required.extend(required_fields)

                        for prop_name, prop_schema in resolved_schema[
                            "properties"
                        ].items():
                            # 确保每个属性也被递归解析
                            inputs[prop_name] = self._resolve_refs_recursively(
                                prop_schema, api_spec
                            )

        # 3. 构建OpenAI的参数schema
        parameters_schema_openai = {
            "type": "object",
            "properties": inputs,
            "required": required,
        }

        return parameters_schema_openai

    def _parse_inputs_schema(self, inputs_schema: Dict[str, Any]) -> Dict[str, Any]:
        """解析工具输入参数的schema"""
        inputs = {}

        for param_name, param_schema in inputs_schema["properties"].items():
            inputs[param_name] = param_schema.copy()

            if param_name in inputs_schema["required"]:
                inputs[param_name]["required"] = True

        return inputs
