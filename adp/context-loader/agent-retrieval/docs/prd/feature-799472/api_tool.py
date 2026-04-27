from typing import Any, Dict, List
import json

import aiohttp

from app.utils.dict_util import get_dict_val_by_path
from DolphinLanguageSDK.utils.tools import ToolInterrupt
from DolphinLanguageSDK.context import Context
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
from .common import parse_kwargs, ToolMapInfo, COLORS, APIToolResponse

# Import from api_tool_pkg module using relative import
from .api_tool_pkg.input import APIToolInputHandler


class APITool(APIToolInputHandler):
    def __init__(self, tool_info, tool_config):
        tool_name = tool_info.get("name", tool_info.get("tool_id", ""))

        self.name = tool_name
        self.description = self._parse_description(tool_info)

        # 1. input参数映射
        self.tool_map_list: List[ToolMapInfo] = []

        for item in tool_config.get("tool_input", []):
            # 如果item是Pydantic模型对象，需要转换为字典
            if hasattr(item, "model_dump"):
                item_dict = item.model_dump()
            elif hasattr(item, "dict"):
                item_dict = item.dict()
            elif isinstance(item, dict):
                item_dict = item
            else:
                # 如果都不是，尝试使用__dict__
                item_dict = item.__dict__ if hasattr(item, "__dict__") else item

            self.tool_map_list.append(ToolMapInfo(**item_dict))

        # 2. input参数解析
        # ---inputs start---
        self.unfiltered_inputs = self._parse_inputs(
            tool_info.get("metadata", {}).get("api_spec", {})
        )

        self.inputs_schema = self._filter_exposed_inputs(self.unfiltered_inputs)

        self.inputs = self._parse_inputs_schema(self.inputs_schema)
        # ---inputs end---

        # 3. output参数解析
        self.outputs = self._parse_outputs(
            tool_info.get("metadata", {}).get("api_spec", {})
        )

        self.tool_info = tool_info
        self.tool_config = tool_config
        self.intervention = tool_config.get("intervention", False)

        # 4. result_process_strategies解析（结果处理策略）
        result_process_strategy_cfg = tool_config.get("result_process_strategies", [])

        if result_process_strategy_cfg:
            self.result_process_strategy_cfg = []

            for cfg in result_process_strategy_cfg:
                category_cfg = cfg.get("category", None)
                strategy_cfg = cfg.get("strategy", None)

                if category_cfg and strategy_cfg:
                    tmp_strategy_cfg = {
                        "category": category_cfg.get("id", None),
                        "strategy": strategy_cfg.get("id", None),
                    }

                    self.result_process_strategy_cfg.append(tmp_strategy_cfg)

    def _parse_description(self, tool_info):
        """解析工具描述"""
        description = tool_info.get("description", "")

        # use_rule = tool_info.get("use_rule", "")
        # if use_rule:
        #     description += "\n## Use Rule:\n" + use_rule

        return description

    def _filter_exposed_inputs(
        self, inputs: Dict[str, Any], tool_map_list: List["ToolMapInfo"] = None
    ) -> Dict[str, Any]:
        """
        过滤输入参数，移除设置为不暴露给大模型的参数
        分析tool_map_list:
        1. 如果enabled为False, 则不暴露给大模型
        2. 只有map_type为auto的参数, 才暴露给大模型
        3. 如果有children, 且children中有需要暴露给大模型的参数,则该参数也暴露给大模型

        Args:
            inputs: 原始输入参数字典

        Returns:
            过滤后的输入参数字典，只包含暴露给大模型的参数
        """
        if tool_map_list is None:
            tool_map_list = self.tool_map_list

        def should_expose_param(tool_map_item: ToolMapInfo) -> bool:
            """
            递归判断参数是否应该暴露给大模型

            Args:
                tool_map_item: 工具映射项

            Returns:
                bool: 是否应该暴露给大模型
            """
            # 1. 如果enabled为False，则不暴露
            if not tool_map_item.is_enabled():
                return False

            # 2. 如果有children，递归检查children
            if tool_map_item.children:
                # 检查children中是否有需要暴露的参数
                for child_item in tool_map_item.children:
                    if should_expose_param(child_item):
                        return True

                # 如果所有children都不暴露，则该参数也不暴露
                return False

            # 3. 没有children时，只有map_type为auto的参数才暴露给大模型
            return tool_map_item.get_map_type() == "auto"

        # 遍历所有输入参数
        if "properties" in inputs:
            new_properties = {}
            new_required = inputs.get("required", [])

            for param_name, param_schema in inputs["properties"].items():
                # 查找对应的tool_map_item
                corresponding_tool_map_item = None

                for tool_map_item in tool_map_list:
                    if tool_map_item.input_name == param_name:
                        corresponding_tool_map_item = tool_map_item
                        break

                # 如果找到了对应的tool_map_item，根据规则判断是否暴露
                if corresponding_tool_map_item:
                    if should_expose_param(corresponding_tool_map_item):
                        new_properties[param_name] = param_schema

                        if "properties" in param_schema:
                            new_properties[param_name] = self._filter_exposed_inputs(
                                param_schema, corresponding_tool_map_item.children
                            )
                    else:
                        if param_name in new_required:
                            new_required.remove(param_name)
                else:
                    # 如果没有找到对应的tool_map_item，默认暴露（保持向后兼容）
                    new_properties[param_name] = param_schema

            inputs["properties"] = new_properties

            if inputs.get("required", []):
                if new_required:
                    inputs["required"] = new_required
                else:
                    inputs.pop("required", None)

        return inputs

    async def arun_stream(self, **kwargs):
        tool_input, props = parse_kwargs(**kwargs)
        tool_input.pop("props", None)
        """异步流式执行工具"""
        # 如果工具需要干预，则抛出ToolInterrupt异常
        if isinstance(props, dict) and "intervention" in props:
            intervention = props["intervention"]
        else:
            intervention = self.intervention
        if intervention:
            tool_args = []
            for key, value in tool_input.items():
                tool_args.append(
                    {
                        "key": key,
                        "value": value,
                        "type": self.unfiltered_inputs.get("properties", {})
                        .get(key, {})
                        .get("type"),
                    }
                )
            raise ToolInterrupt(tool_name=self.name, tool_args=tool_args)

        gvp: "Context" = props.get("gvp")

        path_params, query_params, body_params, header_params = self.process_params(
            tool_input,
            self.tool_info.get("metadata", {}).get("api_spec", {}),
            gvp,
        )

        # ?stream=true&mode=sse
        url = "http://{HOST_AGENT_OPERATOR}:{PORT_AGENT_OPERATOR}/api/agent-operator-integration/internal-v1/tool-box/{box_id}/proxy/{tool_id}?stream=true&mode=sse".format(
            HOST_AGENT_OPERATOR=self.tool_config.get(
                "HOST_AGENT_OPERATOR", "agent-operator-integration"
            ),
            PORT_AGENT_OPERATOR=self.tool_config.get("PORT_AGENT_OPERATOR", "9000"),
            box_id=self.tool_config.get("tool_box_id"),
            tool_id=self.tool_config.get("tool_id"),
        )
        body = {
            "header": header_params,
            "body": body_params,
            "query": query_params,
            "path": path_params,
        }

        StandLogger.info(
            f"\n{COLORS['header']}{COLORS['bold']}开始请求工具 {self.name} 的代理接口{COLORS['end']}\n"
            f"{COLORS['blue']}========================================{COLORS['end']}\n"
            f"{COLORS['cyan']}{COLORS['bold']}URL:{COLORS['end']} {url}\n"
            f"{COLORS['green']}{COLORS['bold']}Headers:{COLORS['end']}\n{json.dumps(header_params, indent=2, ensure_ascii=False)}\n"
            f"{COLORS['yellow']}{COLORS['bold']}Body:{COLORS['end']}\n{json.dumps(body, indent=2, ensure_ascii=False)}\n"
            f"{COLORS['blue']}========================================{COLORS['end']}"
        )
        async with aiohttp.ClientSession(
            timeout=aiohttp.ClientTimeout(total=300)
        ) as session:
            async with session.request(
                "POST",
                url,
                headers=header_params,
                json=body,
                verify_ssl=False,
            ) as response:
                async for rt in self.handle_response(response):
                    yield rt

    async def handle_response(self, response):
        # 默认为流式
        # is_stream = False
        is_stream = True

        if response.status != 200:
            error_str = await response.text()
            StandLogger.error(
                f"\n{COLORS['header']}{COLORS['bold']}请求工具 {self.name} 的代理接口失败: {error_str}{COLORS['end']}\n"
            )
            resp = APIToolResponse(answer=error_str)
            yield resp.to_dict()

        else:
            if is_stream:
                try:
                    buffer = bytearray()
                    async for chunk in response.content.iter_chunked(1024):
                        buffer.extend(chunk)

                        lines = buffer.split(b"\n")

                        for line in lines[:-1]:
                            if not line.startswith(b"data"):
                                continue

                            line_decoded = line.decode().split("data:", 1)[1]

                            if "[DONE]" in line_decoded:
                                break

                            try:
                                line_json = json.loads(line_decoded, strict=False)
                                resp = APIToolResponse(
                                    answer=line_json, block_answer=line_json
                                )

                                yield resp.to_dict()
                            except Exception as e:
                                StandLogger.error(
                                    f"APITool Execute, Error parsing line: {line_decoded}, error: {e}"
                                )
                                yield {"answer": line_decoded}

                        buffer = lines[-1]  # 保留最后一个不完整的行，等待下一个块的到来

                except Exception as e:
                    error_str = await response.text()
                    StandLogger.error(
                        f"\n{COLORS['header']}{COLORS['bold']}请求工具 {self.name} 的代理接口失败: {e}{COLORS['end']}\n"
                    )
                    resp = APIToolResponse(answer=error_str)
                    yield resp.to_dict()
            else:
                try:
                    res = await response.json()
                    if res.get("error"):
                        StandLogger.error(
                            f"\n{COLORS['header']}{COLORS['bold']}请求工具 {self.name} 的代理接口失败: {json.dumps(res, ensure_ascii=False)}{COLORS['end']}\n"
                        )
                        resp = APIToolResponse(answer=res["error"])
                    elif res.get("status_code") != 200:
                        StandLogger.error(
                            f"\n{COLORS['header']}{COLORS['bold']}请求工具 {self.name} 的代理接口失败: {json.dumps(res, ensure_ascii=False)}{COLORS['end']}\n"
                        )
                        resp = APIToolResponse(answer=res.get("body", ""))
                    else:
                        StandLogger.info(
                            f"\n{COLORS['header']}{COLORS['bold']}请求工具 {self.name} 的代理接口成功: {json.dumps(res, ensure_ascii=False)}{COLORS['end']}\n"
                        )
                        resp = APIToolResponse(
                            answer=res.get("body", ""), block_answer=res.get("body", "")
                        )

                    yield resp.to_dict()
                except Exception:
                    res = await response.text()

                    StandLogger.error(
                        f"\n{COLORS['header']}{COLORS['bold']}请求工具 {self.name} 的代理接口失败: {json.dumps(res, ensure_ascii=False)}{COLORS['end']}\n"
                    )
                    resp = APIToolResponse(answer=res)

                    yield resp.to_dict()

    def process_params(self, tool_input, api_spec, gvp: "Context"):
        """处理工具输入参数"""
        # 根据self.tool_map_list中的map_type，处理tool_input
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

        def process_tool_map_item(
            tool_map_item: ToolMapInfo,
            current_tool_input: Dict[str, Any],
            input_params: Dict[str, Any],
        ):
            """递归处理单个工具映射项"""
            # 检查是否启用
            if tool_map_item.is_enabled() is False:
                if tool_map_item.input_name in current_tool_input:
                    current_tool_input.pop(tool_map_item.input_name)
                return

            # 处理 children 递归情况
            if tool_map_item.children:
                process_needed = False

                if tool_map_item.input_name not in current_tool_input:
                    process_needed = True
                    current_tool_input[tool_map_item.input_name] = {}

                for child_item in tool_map_item.children:
                    process_tool_map_item(
                        child_item,
                        current_tool_input[tool_map_item.input_name],
                        input_params.get("properties", {}).get(
                            tool_map_item.input_name, {}
                        ),
                    )

                if (
                    process_needed
                    and current_tool_input[tool_map_item.input_name] == {}
                ):
                    current_tool_input.pop(tool_map_item.input_name)

                return

            # 处理 map_type
            map_type = tool_map_item.get_map_type()

            if map_type == "auto":
                return

            elif map_type == "var":
                cite_var = tool_map_item.get_map_value()
                # 递归获取变量值
                cite_var_value = get_dict_val_by_path(gvp.get_all_variables(), cite_var)
                current_tool_input[tool_map_item.input_name] = cite_var_value

            elif map_type == "fixedValue":
                if isinstance(tool_map_item.get_map_value(), str):
                    if (
                        input_params.get("properties", {})
                        .get(tool_map_item.input_name, {})
                        .get("type", "")
                        != "string"
                    ):
                        try:
                            tool_map_item.map_value = json.loads(
                                tool_map_item.get_map_value()
                            )
                        except Exception:
                            StandLogger.warn(
                                f"工具的输入参数{tool_map_item.input_name}的值{tool_map_item.get_map_value()}不是json格式"
                            )
                            tool_map_item.map_value = tool_map_item.get_map_value()
                    else:
                        if tool_map_item.get_map_value().startswith(
                            '"'
                        ) and tool_map_item.get_map_value().endswith('"'):
                            tool_map_item.map_value = json.loads(
                                tool_map_item.get_map_value()
                            )
                current_tool_input[tool_map_item.input_name] = (
                    tool_map_item.get_map_value()
                )
            else:
                current_tool_input[tool_map_item.input_name] = (
                    tool_map_item.get_map_value()
                )

        # 处理 tool_map_list
        for item in self.tool_map_list:
            process_tool_map_item(item, tool_input, self.unfiltered_inputs)

        # 根据api_spec中参数的位置，处理tool_input为各个位置的参数
        path_params, query_params, body, headers = {}, {}, {}, {}
        arg_type = {
            "path": [],
            "query": [],
            "body": [],
            "header": {},
            "cookie": [],
        }

        # 确定各参数的位置
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

        # 处理body参数
        request_body = api_spec.get("request_body", {})

        if request_body and "content" in request_body:
            # 遍历所有content类型
            for content_type, content_info in request_body["content"].items():
                if "schema" in content_info:
                    schema = content_info["schema"]

                    # 处理schema引用
                    if "$ref" in schema:
                        ref_path = schema["$ref"]
                        if ref_path.startswith("#/components/schemas/"):
                            schema_name = ref_path.split("/")[-1]
                            if schema_name in api_spec.get("components", {}).get(
                                "schemas", {}
                            ):
                                schema = api_spec["components"]["schemas"][schema_name]

                    # 解析schema的properties
                    if "properties" in schema:
                        for prop_name, prop_schema in schema["properties"].items():
                            if prop_name in tool_input:
                                body[prop_name] = tool_input[prop_name]

        for key, value in tool_input.items():
            if key in arg_type["path"]:
                path_params[key] = value
            elif key in arg_type["query"]:
                query_params[key] = value
            elif key in arg_type["body"]:
                body[key] = value
            elif key in arg_type["header"]:
                headers[key] = value

        # 从全局变量中提取内部接口鉴权参数
        global_headers = gvp.get_var_value("header")
        if global_headers:
            if has_user_account(global_headers):
                set_user_account_id(headers, get_user_account_id(global_headers))
            if has_user_account_type(global_headers):
                set_user_account_type(headers, get_user_account_type(global_headers))

        return path_params, query_params, body, headers
