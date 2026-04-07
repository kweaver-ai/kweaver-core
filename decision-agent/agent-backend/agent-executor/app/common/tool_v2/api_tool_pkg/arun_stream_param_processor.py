from typing import Any, Dict, List, Tuple

from dolphin.core.context.context import Context
from app.domain.enum.common.user_account_header_key import (
    get_user_account_id,
    get_user_account_type,
    set_user_account_id,
    set_user_account_type,
    has_user_account,
    has_user_account_type,
)

# Import from common module using relative import
from ..common import ToolMapInfo

# Import from process_tool_map_item module using relative import
from .process_tool_map_item import process_tool_map_item


def _process_request_body(
    request_body: Dict[str, Any],
    api_spec: Dict[str, Any],
    tool_input: Dict[str, Any],
    body: Dict[str, Any],
):
    """处理请求体参数

    Args:
        request_body: 请求体配置
        api_spec: API规范
        tool_input: 工具输入参数
        body: 输出的body参数字典
    """
    if not request_body or "content" not in request_body:
        return

    # 遍历所有content类型
    for content_type, content_info in request_body["content"].items():
        if "schema" not in content_info:
            continue

        schema = content_info["schema"]

        # 处理schema引用
        if "$ref" in schema:
            ref_path = schema["$ref"]
            if ref_path.startswith("#/components/schemas/"):
                schema_name = ref_path.split("/")[-1]
                if schema_name in api_spec.get("components", {}).get("schemas", {}):
                    schema = api_spec["components"]["schemas"][schema_name]

        # 解析schema的properties
        if "properties" in schema:
            for prop_name, prop_schema in schema["properties"].items():
                if prop_name in tool_input:
                    body[prop_name] = tool_input[prop_name]


def process_params(
    tool_input: Dict[str, Any],
    api_spec: Dict[str, Any],
    gvp: Context,
    tool_map_list: List[ToolMapInfo],
    unfiltered_inputs: Dict[str, Any],
) -> Tuple[Dict[str, Any], Dict[str, Any], Dict[str, Any], Dict[str, Any]]:
    """处理工具输入参数

    Args:
        tool_input: 工具输入参数
        api_spec: API规范
        gvp: 全局变量上下文
        tool_map_list: 工具映射列表
        unfiltered_inputs: 未过滤的输入参数

    Returns:
        Tuple[path_params, query_params, body, headers]: 处理后的各类参数
    """
    # 根据tool_map_list中的map_type，处理tool_input
    """
    tool_input = [
        {
            "input_name": "input_name",
            "input_type": "string",
            "map_type": "fixedValue",  # fixedValue（固定值）、var（变量）、model（选择模型）、auto（模型自动生成）
            "map_value": "map_value"
        }
    ]
    """

    # 1. 处理 tool_map_list
    for item in tool_map_list:
        process_tool_map_item(item, tool_input, unfiltered_inputs, gvp)

    # 2. 根据api_spec中参数的位置，处理tool_input为各个位置的参数
    path_params, query_params, body, headers = {}, {}, {}, {}
    arg_type = {
        "path": [],
        "query": [],
        "body": [],
        "header": {},
        "cookie": [],
    }

    # 3. 确定各参数的位置
    for item in api_spec.get("parameters", []):
        if item.get("in", "") == "path":
            arg_type["path"].append(item.get("name", ""))
        elif item.get("in", "") == "query":
            arg_type["query"].append(item.get("name", ""))
        elif item.get("in", "") == "body":
            arg_type["body"].append(item.get("name", ""))
        elif item.get("in", "") == "header":
            arg_type["header"][item.get("name", "")] = item.get("schema", {}).get(
                "type", ""
            )

    # 4. 处理body参数
    request_body = api_spec.get("request_body", {})
    _process_request_body(request_body, api_spec, tool_input, body)

    # 5. 处理path、query、header参数
    for key, value in tool_input.items():
        if key in arg_type["path"]:
            path_params[key] = value
        elif key in arg_type["query"]:
            query_params[key] = value
        elif key in arg_type["body"]:
            body[key] = value
        elif key in arg_type["header"]:
            headers[key] = value

    # 6. 从全局变量中提取内部接口鉴权参数
    global_headers = gvp.get_var_value("header")
    if global_headers:
        if has_user_account(global_headers):
            set_user_account_id(headers, get_user_account_id(global_headers))
        if has_user_account_type(global_headers):
            set_user_account_type(headers, get_user_account_type(global_headers))

    return path_params, query_params, body, headers
