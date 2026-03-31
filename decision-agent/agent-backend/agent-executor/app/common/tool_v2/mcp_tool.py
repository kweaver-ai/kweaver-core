from typing import Dict
import json

import aiohttp

from dolphin.core.utils.tools import Tool
from dolphin.core.context.context import Context
from app.common.stand_log import StandLogger
from app.domain.enum.common.user_account_header_key import (
    get_user_account_id,
    get_user_account_type,
    set_user_account_id,
    set_user_account_type,
    has_user_account,
    has_user_account_type,
)

# Import from common module using relative import
from .common import parse_kwargs


class MCPTool(Tool):
    def __init__(self, mcp_tool_info, mcp_config):
        self.name = mcp_tool_info.get("name", "unknown")
        self.description = mcp_tool_info.get("description", "")
        self.inputs = self._parse_mcp_inputs(mcp_tool_info.get("inputSchema", {}))
        self.outputs = {"result": {"type": "object", "description": "MCP工具执行结果"}}
        self.mcp_config = mcp_config
        # self.intervention = mcp_config.get("intervention", False)
        self.mcp_server_id = mcp_config.get("mcp_server_id", "")
        # self.tool_map_list = mcp_config.get("tool_map_list", [])

    def _parse_mcp_inputs(self, input_schema):
        """解析MCP工具输入参数"""
        # """解析MCP工具输入参数，支持$ref引用和复杂嵌套结构"""
        inputs = {}
        properties = input_schema.get("properties", {})
        required = input_schema.get("required", [])

        for prop_name, prop_schema in properties.items():
            # 老的，等后面统一调整（2025年09月02日10:32:43）
            inputs[prop_name] = {
                "type": prop_schema.get("type", "string"),
                "description": prop_schema.get("description", ""),
                "required": prop_name in required,
            }

            # note: 暂时注释掉这块，使用上面的老的，等后面统一调整（2025年09月02日10:32:43）
            # # 递归解析$ref引用
            # resolved_schema = self._resolve_mcp_refs_recursively(
            #     prop_schema, input_schema
            # )
            #
            # # 保留所有schema属性，然后添加required字段
            # param_def = resolved_schema.copy()
            # param_def["required"] = prop_name in required
            #
            # # 确保基本字段有默认值
            # if "type" not in param_def:
            #     param_def["type"] = "string"
            # if "description" not in param_def:
            #     param_def["description"] = ""
            #
            # inputs[prop_name] = param_def

        return inputs

    def _resolve_mcp_refs_recursively(self, schema, input_schema, visited=None):
        """递归解析MCP schema中的所有$ref引用，支持$defs定义"""
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
            if ref_path.startswith("#/$defs/"):
                schema_name = ref_path.split("/")[-1]
                defs = input_schema.get("$defs", {})
                if schema_name in defs:
                    # 递归解析引用的schema
                    resolved_schema = self._resolve_mcp_refs_recursively(
                        defs[schema_name], input_schema, visited
                    )
                    return resolved_schema

        # 递归处理所有嵌套的字典和列表
        result = {}
        for key, value in schema.items():
            if isinstance(value, dict):
                result[key] = self._resolve_mcp_refs_recursively(
                    value, input_schema, visited
                )
            elif isinstance(value, list):
                result[key] = [
                    (
                        self._resolve_mcp_refs_recursively(item, input_schema, visited)
                        if isinstance(item, dict)
                        else item
                    )
                    for item in value
                ]
            else:
                result[key] = value

        return result

    async def arun_stream(self, **kwargs):
        tool_input, props = parse_kwargs(**kwargs)
        tool_input.pop("props", None)

        gvp: "Context" = props.get("gvp")
        """异步流式执行MCP工具"""
        # # 如果工具需要干预，则抛出ToolInterrupt异常
        # if self.intervention:
        #     tool_args = []
        #     for key, value in tool_input.items():
        #         tool_args.append(
        #             {
        #                 "key": key,
        #                 "value": value,
        #                 "type": self.inputs.get(key, {}).get("type"),
        #             }
        #         )
        #     raise ToolInterrupt(tool_name=self.name, tool_args=tool_args)
        # # 根据self.tool_map_list中的map_type，处理tool_input
        # for item in self.tool_map_list:
        #     if item.get("enable", True) is False:
        #         if item.get("input_name", "") in tool_input:
        #             tool_input.pop(item.get("input_name", ""))
        #         continue
        #     if item.get("map_type", "") == "auto":
        #         continue
        #     elif item.get("map_type", "") == "var":
        #         cite_var = item["map_value"]
        #         # 递归获取变量值

        #         cite_var_value = get_dict_val_by_path(gvp.get_all_variables(), cite_var)
        #         tool_input[item["input_name"]] = cite_var_value
        #     elif item.get("map_type", "") == "fixedValue":
        #         if self.inputs.get(item["input_name"], {}).get("type", "") != "string":
        #             if not isinstance(item["map_value"], str):
        #                 try:
        #                     item["map_value"] = json.loads(item["map_value"])
        #                 except Exception:
        #                     StandLogger.warn(
        #                         f"工具的输入参数{item['input_name']}的值{item['map_value']}不是json格式"
        #                     )
        #                     item["map_value"] = item["map_value"]
        #         tool_input[item["input_name"]] = item["map_value"]
        #     else:
        #         tool_input[item["input_name"]] = item["map_value"]

        url = "http://{HOST_AGENT_OPERATOR}:{PORT_AGENT_OPERATOR}/api/agent-operator-integration/internal-v1/mcp/proxy/{mcp_server_id}/tool/call".format(
            HOST_AGENT_OPERATOR=self.mcp_config.get(
                "HOST_AGENT_OPERATOR", "agent-operator-integration"
            ),
            PORT_AGENT_OPERATOR=self.mcp_config.get("PORT_AGENT_OPERATOR", "9000"),
            mcp_server_id=self.mcp_server_id,
        )
        method = "POST"

        headers = {
            "Content-Type": "application/json",
        }
        required_headers = ["authorization", "token"]
        for header in required_headers:
            if header in gvp.get_var_value("header"):
                headers[header] = gvp.get_var_value("header")[header]

        # 从全局变量中提取账号信息
        global_headers = gvp.get_var_value("header")
        if global_headers:
            if has_user_account(global_headers):
                set_user_account_id(headers, get_user_account_id(global_headers))
            if has_user_account_type(global_headers):
                set_user_account_type(headers, get_user_account_type(global_headers))

        body = {
            "tool_name": self.name,
            "parameters": tool_input,
        }
        StandLogger.info(
            f"""开始请求MCP工具 {self.name}
url: {url}
body: {json.dumps(body, ensure_ascii=False)}
"""
        )
        async with aiohttp.ClientSession(
            timeout=aiohttp.ClientTimeout(total=300)
        ) as session:
            async with session.request(
                method, url, headers=headers, json=body
            ) as response:
                if response.status != 200:
                    error_str = await response.text()
                    StandLogger.error(f"请求MCP工具 {self.name} 失败: {error_str}")
                    yield {"answer": error_str}
                else:
                    try:
                        res = await response.json()
                        yield {
                            "answer": res["content"][0]["text"],
                            "block_answer": res["content"][0]["text"],
                        }
                    except Exception:
                        res = await response.text()
                        StandLogger.error(f"请求MCP工具 {self.name} 失败: {res}")
                        yield {
                            "answer": res,
                        }


def get_mock_mcp_tools():
    with open(".local/mcp_tools.json", "r") as f:
        return json.load(f)


async def get_mcp_tools(mcp_info: dict) -> Dict[str, MCPTool]:
    mcp_server_id = mcp_info.get("mcp_server_id")
    HOST_AGENT_OPERATOR = mcp_info.get("HOST_AGENT_OPERATOR")
    PORT_AGENT_OPERATOR = mcp_info.get("PORT_AGENT_OPERATOR")

    # 获取指定MCP服务下的工具列表
    url = f"http://{HOST_AGENT_OPERATOR}:{PORT_AGENT_OPERATOR}/api/agent-operator-integration/internal-v1/mcp/proxy/{mcp_server_id}/tools"
    method = "GET"

    async with aiohttp.ClientSession(
        timeout=aiohttp.ClientTimeout(total=300)
    ) as session:
        async with session.request(method, url) as response:
            if response.status != 200:
                error_str = await response.text()
                StandLogger.error(f"failed to request {url}:{error_str}")
                return {}

            res = await response.json()
            tool_infos = res.get("tools", [])

    mcp_tools = {}
    for tool_info in tool_infos:
        mcp_tools[tool_info.get("name")] = MCPTool(tool_info, mcp_info)

    return mcp_tools
