from typing import Any, Dict, List
import json

import asyncio
import aiohttp

from dolphin.core.context.context import Context
from app.common.config import Config
from app.common.stand_log import StandLogger

# Import from common module using relative import
from .common import parse_kwargs, ToolMapInfo, COLORS, APIToolResponse

# Import from api_tool_pkg module using relative import
from .api_tool_pkg.input import APIToolInputHandler
from .api_tool_pkg.arun_stream_param_processor import (
    process_params as process_params_func,
)
from .api_tool_pkg.evidence_extractor import (
    is_evidence_extraction_enabled,
    extract_evidence,
)


class APITool(APIToolInputHandler):
    TRACEAI_EVIDENCE_HEADER = "X-TraceAi-Enable-Evidence"

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

        # 5. 根据 intervention 配置生成 interrupt_config
        # 供 Dolphin SDK 使用，在 SDK 内部触发中断
        if self.intervention:
            intervention_message = tool_config.get(
                "intervention_confirmation_message",
                f"工具 {self.name} 需要确认执行",  # 默认值
            )
            self.interrupt_config = {
                "requires_confirmation": True,
                "confirmation_message": intervention_message,
            }
        else:
            self.interrupt_config = None

    def _parse_description(self, tool_info):
        """解析工具描述"""
        description = tool_info.get("description", "")

        use_rule = tool_info.get("use_rule", "")
        if use_rule:
            description += "\n## Use Rule:\n" + use_rule

        return description

    def _apply_global_feature_headers(self, header_params: Dict[str, Any]) -> None:
        """将全局特性开关注入到 proxy 请求头中。"""
        header_params[self.TRACEAI_EVIDENCE_HEADER] = str(
            Config.features.enable_traceai_evidence
        ).lower()

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

        # 注意：中断逻辑已移至 Dolphin SDK 内部，通过 interrupt_config 触发
        # 此处不再主动抛出 ToolInterrupt 异常

        # Mock: is_aaron_local_dev() 时直接使用 mock 数据
        # if is_aaron_local_dev():
        #     async for rt in self._mock_kn_search_stream():
        #         yield rt
        #     return

        # 1. 获取gvp
        gvp: "Context" = props.get("gvp")

        # 2. 处理参数
        path_params, query_params, body_params, header_params = self.process_params(
            tool_input,
            self.tool_info.get("metadata", {}).get("api_spec", {}),
            gvp,
        )
        self._apply_global_feature_headers(header_params)

        # 4. 构建请求URL
        # 工具超时时间 （读取完整数据的超时时间）
        toolTimeout = self.tool_config.get("tool_timeout", 300)
        if toolTimeout <= 0:
            toolTimeout = 300
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
            "timeout": toolTimeout,
        }

        toolTimeout = toolTimeout + 30

        # 5. 打印请求信息
        StandLogger.info(
            f"\n{COLORS['header']}{COLORS['bold']}开始请求工具 {self.name} 的代理接口{COLORS['end']}\n"
            f"{COLORS['blue']}========================================{COLORS['end']}\n"
            f"{COLORS['cyan']}{COLORS['bold']}URL:{COLORS['end']} {url}\n"
            f"{COLORS['green']}{COLORS['bold']}Headers:{COLORS['end']}\n{json.dumps(header_params, indent=2, ensure_ascii=False)}\n"
            f"{COLORS['yellow']}{COLORS['bold']}Body:{COLORS['end']}\n{json.dumps(body, indent=2, ensure_ascii=False)}\n"
            f"{COLORS['blue']}========================================{COLORS['end']}"
        )

        # 6. 发送请求
        async with aiohttp.ClientSession(
            timeout=aiohttp.ClientTimeout(
                total=toolTimeout,
                sock_connect=30,  # 建立连接超时
                sock_read=toolTimeout,  # 读取数据超时
            )
        ) as session:
            async with session.request(
                "POST",
                url,
                headers=header_params,
                json=body,
                verify_ssl=False,
            ) as response:
                # 传递超时时间给handle_response
                try:
                    last_rt = None
                    async for rt in self.handle_response(response, toolTimeout):
                        yield rt
                        # 记录最后一条有效 rt
                        if rt.get("answer"):
                            last_rt = rt

                    # 流式响应结束后，修改原 rt 追加 evidence 并重新 yield
                    if self._try_append_evidence(last_rt):
                        yield last_rt
                except asyncio.TimeoutError:
                    StandLogger.error(
                        f"\n{COLORS['header']}{COLORS['bold']}请求工具 {self.name} 的总时长超时 ({toolTimeout}秒){COLORS['end']}\n"
                    )
                    resp = APIToolResponse(
                        answer=f"工具请求超时: 总时长超过 {toolTimeout} 秒"
                    )
                    yield resp.to_dict()

    async def handle_response(self, response, total_timeout=None):
        """
        处理响应，支持总时长超时控制

        Args:
            response: HTTP响应对象
            total_timeout: 总超时时间（秒），如果为None则不启用总时长控制
        """
        import time

        # 记录开始时间
        start_time = time.time()

        # 默认为流式
        # is_stream = False
        is_stream = True

        def check_total_timeout():
            """检查总时长是否超时"""
            if total_timeout is not None:
                elapsed_time = time.time() - start_time
                if elapsed_time > total_timeout:
                    StandLogger.error(
                        f"\n{COLORS['header']}{COLORS['bold']}请求工具 {self.name} 的总时长超时 ({total_timeout}秒)，已耗时 {elapsed_time:.2f} 秒{COLORS['end']}\n"
                    )
                    raise asyncio.TimeoutError(f"总时长超过 {total_timeout} 秒")

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
                        # 检查总时长是否超时
                        check_total_timeout()

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

                except asyncio.TimeoutError:
                    # 总时长超时异常，直接抛出
                    raise
                except Exception as e:
                    error_str = await response.text()
                    StandLogger.error(
                        f"\n{COLORS['header']}{COLORS['bold']}请求工具 {self.name} 的代理接口失败: {e}{COLORS['end']}\n"
                    )
                    resp = APIToolResponse(answer=error_str)
                    yield resp.to_dict()
            else:
                try:
                    # 检查总时长是否超时
                    check_total_timeout()

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
                except asyncio.TimeoutError:
                    # 总时长超时异常，直接抛出
                    raise
                except Exception:
                    # 检查总时长是否超时
                    check_total_timeout()

                    res = await response.text()

                    StandLogger.error(
                        f"\n{COLORS['header']}{COLORS['bold']}请求工具 {self.name} 的代理接口失败: {json.dumps(res, ensure_ascii=False)}{COLORS['end']}\n"
                    )
                    resp = APIToolResponse(answer=res)

                    yield resp.to_dict()

    def process_params(self, tool_input, api_spec, gvp: "Context"):
        """处理工具输入参数"""
        return process_params_func(
            tool_input, api_spec, gvp, self.tool_map_list, self.unfiltered_inputs
        )

    def _try_append_evidence(self, last_rt):
        """
        尝试从最后一条 rt 中提取 evidence，直接修改原 rt 的 answer/block_answer。
        仅当 name=='kn_search' 且环境变量开启时触发。

        Args:
            last_rt: 流式响应中最后一条有效的 rt 字典

        Returns:
            bool: 是否成功追加了 evidence
        """
        if self.name != "kn_search":
            return False
        if not is_evidence_extraction_enabled():
            return False
        if not last_rt:
            return False

        answer_data = last_rt.get("answer")
        if not answer_data or not isinstance(answer_data, dict):
            return False

        evidence = extract_evidence(answer_data)
        if evidence:
            # 直接修改原 rt 的 answer 和 block_answer
            answer_data["_evidence"] = evidence
            block_answer = last_rt.get("block_answer")
            if block_answer and isinstance(block_answer, dict):
                block_answer["_evidence"] = evidence

            StandLogger.info(
                f"\n{COLORS['green']}{COLORS['bold']}工具 {self.name} 提取到 evidence, "
                f"包含 {len(evidence.get('evidences', [{}])[0].get('content', {}).get('object_instances', []))} 个 object_instances"
                f"{COLORS['end']}\n"
            )
            return True

        return False

    async def _mock_kn_search_stream(self):
        """
        本地开发环境 mock 流式响应。
        使用 .local/tmp/evidence/kn_serach_res.json 作为 mock 数据。
        """
        import os

        mock_file = os.path.join(
            os.path.dirname(os.path.abspath(__file__)),
            "..",
            "..",
            "..",
            ".local",
            "tmp",
            "evidence",
            "kn_serach_res.json",
        )

        try:
            with open(mock_file, "r", encoding="utf-8") as f:
                mock_data = json.load(f)
            StandLogger.info(
                f"\n{COLORS['cyan']}{COLORS['bold']}[Mock] 工具 {self.name} 使用本地 mock 数据: {mock_file}{COLORS['end']}\n"
            )
        except FileNotFoundError:
            StandLogger.error(
                f"\n{COLORS['red']}{COLORS['bold']}[Mock] Mock 文件不存在: {mock_file}{COLORS['end']}\n"
            )
            mock_data = {
                "nodes": [],
                "relation_types": [],
                "action_types": [],
                "object_types": [],
            }

        resp = APIToolResponse(answer=mock_data, block_answer=mock_data)
        rt = resp.to_dict()

        # mock 也触发 evidence 提取（直接修改 rt）
        self.name = "kn_search"
        self._try_append_evidence(rt)

        yield rt
