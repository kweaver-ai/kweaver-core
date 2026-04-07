from dolphin.core.utils.tools import Tool


class APIToolOutputHandler(Tool):
    """API工具输出处理器基类"""

    def _parse_outputs(self, api_spec):
        """解析工具输出参数"""
        outputs = {}

        # 解析 responses 中的输出参数
        if api_spec and "responses" in api_spec:
            for response_info in api_spec["responses"]:
                if (
                    response_info.get("status_code") == "200"
                    and "content" in response_info
                ):
                    for content_type, content_info in response_info["content"].items():
                        if "schema" in content_info:
                            schema = content_info["schema"]

                            # 处理schema引用
                            if "$ref" in schema:
                                ref_path = schema["$ref"]

                                if ref_path.startswith("#/components/schemas/"):
                                    schema_name = ref_path.split("/")[-1]

                                    if schema_name in api_spec.get(
                                        "components", {}
                                    ).get("schemas", {}):
                                        schema = api_spec["components"]["schemas"][
                                            schema_name
                                        ]

                            # 解析schema的properties
                            if "properties" in schema:
                                required_fields = schema.get("required", [])

                                for prop_name, prop_schema in schema[
                                    "properties"
                                ].items():
                                    outputs[prop_name] = {
                                        "type": prop_schema.get("type", "string"),
                                        "description": prop_schema.get(
                                            "description", ""
                                        ),
                                        "required": prop_name in required_fields,
                                    }

                                    # 添加示例值
                                    if "example" in prop_schema:
                                        outputs[prop_name]["example"] = prop_schema[
                                            "example"
                                        ]

        return outputs
