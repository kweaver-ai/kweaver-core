import asyncio
import json
import uuid
from typing import TYPE_CHECKING

import aiohttp

from dolphin.lib.utils.handle_progress import handle_progress, cleanup_progress
from app.utils.dict_util import get_dict_val_by_path
from dolphin.core.utils.tools import Tool, ToolInterrupt
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
from app.domain.vo.agentvo.agent_config_vos import AgentSkillVo
from app.domain.constant.agent_version import AGENT_VERSION_V0

from app.utils.common import get_dolphin_var_value

# Import from common module using relative import
from .common import parse_kwargs
from ..config import Config

if TYPE_CHECKING:
    from ...logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2


class AgentTool(Tool):
    def __init__(self, ac: "AgentCoreV2", agent_skill: AgentSkillVo):
        self.ac = ac

        agent_info = agent_skill.inner_dto.agent_info or {}
        self.name = agent_info.get("name", "")
        self.description = agent_info.get("profile", "")
        self.inputs = self._parse_agent_inputs(agent_skill)
        self.outputs = {"answer": {"type": "object", "description": "agent输出结果"}}
        self.agent_info = agent_info
        self.agent_skill = agent_skill
        self.intervention = agent_skill.intervention or False
        self.tool_map_list = agent_skill.agent_input or []
        self.agent_options = agent_skill.inner_dto.agent_options or {}

        # 根据 intervention 配置生成 interrupt_config
        # 供 Dolphin SDK 使用，在 SDK 内部触发中断
        if self.intervention:
            intervention_message = getattr(
                agent_skill,
                "intervention_confirmation_message",
                f"Agent工具 {self.name} 需要确认执行",  # 默认值
            )
            self.interrupt_config = {
                "requires_confirmation": True,
                "confirmation_message": intervention_message,
            }
        else:
            self.interrupt_config = None

    def _parse_agent_inputs(self, agent_skill: AgentSkillVo):
        """解析Agent输入参数"""
        inputs = {}
        fields = agent_skill.agent_input or []
        for field in fields:
            if field.enable and field.map_type == "auto":
                inputs[field.input_name] = {
                    "type": field.input_type,
                    "description": field.input_desc or "",
                    "required": True,
                }
        return inputs

    def _is_enable_dependency_cache(self):
        return self.ac.run_options_vo.enable_dependency_cache

    async def arun_stream(self, **kwargs):
        tool_input, props = parse_kwargs(**kwargs)
        tool_input.pop("props", None)

        # 注意：中断逻辑已移至 Dolphin SDK 内部，通过 interrupt_config 触发
        # 此处不再主动抛出 ToolInterrupt 异常

        gvp: "Context" = props.get("gvp")

        # 处理agent_input
        for item in self.tool_map_list:
            if not item.enable:
                if item.input_name in tool_input:
                    tool_input.pop(item.input_name)
                continue

            if item.map_type == "auto":
                continue
            elif item.map_type == "var":
                cite_var = item.map_value
                cite_var_value = get_dict_val_by_path(gvp.get_all_variables(), cite_var)

                cite_var_value = get_dolphin_var_value(cite_var_value)
                tool_input[item.input_name] = cite_var_value
            elif item.map_type == "fixedValue":
                if self.inputs.get(item.input_name, {}).get("type", "") != "string":
                    if not isinstance(item.map_value, str):
                        try:
                            map_value = json.loads(item.map_value)
                        except Exception:
                            StandLogger.warn(
                                f"Agent的输入参数{item.input_name}的值{item.map_value}不是json格式"
                            )
                            map_value = item.map_value
                    else:
                        map_value = item.map_value
                else:
                    map_value = item.map_value
                tool_input[item.input_name] = map_value
            else:
                tool_input[item.input_name] = item.map_value

        url = "http://{HOST_AGENT_EXECUTOR}:{PORT_AGENT_EXECUTOR}/api/agent-executor/v2/agent/run".format(
            HOST_AGENT_EXECUTOR=self.agent_skill.inner_dto.HOST_AGENT_EXECUTOR
            or "agent-executor",
            PORT_AGENT_EXECUTOR=self.agent_skill.inner_dto.PORT_AGENT_EXECUTOR
            or "30778",
        )

        headers = {}
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

        method = "POST"
        if "tool" in gvp.get_all_variables():
            tool_input["tool"] = gvp.get_var_value("tool")

        if tool_input.get("_options"):
            self.agent_options.update(tool_input.pop("_options"))

        body = {
            "agent_id": self.agent_info.get("id", ""),
            "agent_config": self.agent_info.get("config", {}),
            "agent_version": AGENT_VERSION_V0,  # agent作为技能使用时，agent_version固定为v0（最新的）
            "agent_input": tool_input,
            "_options": self.agent_options,
        }

        if self._is_enable_dependency_cache():
            body["_options"]["enable_dependency_cache"] = True

        if Config.features.is_skill_agent_need_progress:
            body["_options"]["is_need_progress"] = True

        StandLogger.info(
            f"""开始请求agent工具 {self.name}
url: {url}
body: {json.dumps(body, ensure_ascii=False)}
headers: {json.dumps(headers, ensure_ascii=False)}
"""
        )

        agent_timeout = self.agent_skill.agent_timeout
        if agent_timeout <= 0:
            agent_timeout = 1800
        # 请求agent工具超时时间
        agent_tool_timeout = aiohttp.ClientTimeout(
            total=agent_timeout,  # 总超时
            sock_connect=30,  # 建立连接超时
            sock_read=agent_timeout,  # 读取数据超时
        )

        line_json = {}
        session_id = self.agent_info.get("config", {}).get(
            "session_id", str(uuid.uuid4())
        )

        async with aiohttp.ClientSession(timeout=agent_tool_timeout) as session:
            async with session.request(
                method,
                url,
                headers=headers,
                json=body,
                verify_ssl=False,
            ) as response:
                if response.status != 200:
                    error_str = await response.text()
                    StandLogger.error(f"failed to request {url}:{error_str}")
                    yield {"answer": error_str}

                # try:
                #     async for line in response.content:
                #         line = line.strip()
                #         if not line.startswith(b"data"):
                #             continue
                #         line_decoded = line.decode().split("data:", 1)[1]
                #         try:
                #             line_json = json.loads(line_decoded, strict=False)
                #             yield self.get_output(line_json, session_id)
                #         except Exception:
                #             yield line_decoded
                #         self.handle_tool_interrupt(line_json, tool_input)
                # finally:
                #     cleanup_progress(session_id)

                # 避免报错 chunk too big 使用缓冲机制
                try:
                    buffer = bytearray()
                    start_time = asyncio.get_event_loop().time()

                    while True:
                        # 检查总时间是否超时
                        current_time = asyncio.get_event_loop().time()
                        elapsed_time = current_time - start_time
                        remaining_time = agent_timeout - elapsed_time

                        if remaining_time <= 0:
                            StandLogger.error(
                                f"请求Agent工具 {self.name} 的总时长超时 ({agent_timeout}秒)，已耗时 {elapsed_time:.2f} 秒"
                            )
                            yield {
                                "answer": f"请求Agent工具超时: 总时长超过 {agent_timeout} 秒"
                            }
                            break

                        # 使用asyncio.wait_for包装chunk读取，防止阻塞
                        try:
                            chunk = await asyncio.wait_for(
                                response.content.read(1024),
                                timeout=min(
                                    agent_timeout, remaining_time
                                ),  # 最大agent_timeout，或剩余时间
                            )
                        except asyncio.TimeoutError:
                            StandLogger.error(
                                f"请求Agent工具 {self.name} 的chunk读取超时"
                            )
                            yield {"answer": "请求Agent工具超时: chunk读取超时"}
                            break

                        if not chunk:  # 流结束
                            break

                        buffer.extend(chunk)
                        lines = buffer.split(b"\n")

                        for line in lines[:-1]:
                            if not line.startswith(b"data"):
                                continue

                            line_decoded = line.decode().split("data:", 1)[1]

                            try:
                                line_json = json.loads(line_decoded, strict=False)
                                yield self.get_output(line_json, session_id)
                            except Exception as e:
                                StandLogger.error(
                                    f"AgentTool Execute, Error parsing line: {line_decoded}, error: {e}"
                                )
                                yield {"answer": line_decoded}

                            self.handle_tool_interrupt(line_json, tool_input)

                        buffer = lines[-1]  # 保留最后一个不完整的行，等待下一个块的到来

                finally:
                    cleanup_progress(session_id)

    def handle_tool_interrupt(self, line_json, tool_input):
        if isinstance(line_json, dict) and line_json.get("ask"):
            tool_name = self.name
            tool_args = [
                {
                    "key": "tool",
                    "value": line_json["ask"],
                    "type": "object",
                }
            ]
            for input_name, input_value in tool_input.items():
                if input_name == "tool":
                    continue
                tool_args.append(
                    {
                        "key": input_name,
                        "value": input_value,
                        "type": self.inputs.get(input_name, {}).get("type", "object"),
                    }
                )

            raise ToolInterrupt(
                tool_name=tool_name,
                tool_args=tool_args,
            )

    def get_output(self, line_json, session_id):
        # 已指明需要agent输出的哪个字段，则直接返回
        if "output_vars" in self.agent_options:
            return line_json

        # 处理错误
        if line_json.get("error"):
            return {"error": line_json.get("error")}

        agent_answer = line_json["answer"]
        # 处理progress
        if "_progress" in agent_answer:
            agent_answer["_progress"] = handle_progress(
                session_id, agent_answer.get("_progress", [])
            )
        # 取出最终结果放入result字段，其余结果放入full_result字段
        answer_var = self.agent_info["config"]["output"]["variables"]["answer_var"]
        final_answer = self._get_dolphin_var_value(agent_answer.get(answer_var))
        return {
            "answer": final_answer,
            # "block_answer": agent_answer,
        }

    def _get_dolphin_var_value(self, var):
        if self._is_explore_var(var):
            return var[-1]["answer"]
        elif self._is_llm_var(var):
            return var.get("answer")
        else:
            return var

    def _is_explore_var(self, var):
        executor_keys = [
            "agent_name",
            "stage",
            "answer",
            "think",
            "status",
            "skill_info",
            "block_answer",
            "input_message",
            "interrupted",
        ]
        # 如果var是list of dict 且每个dict的key和executor_keys一致，则返回True
        if isinstance(var, list) and all(
            isinstance(item, dict) and set(item.keys()) == set(executor_keys)
            for item in var
        ):
            return True
        return False

    def _is_llm_var(self, var):
        llm_keys = ["answer", "think"]
        if isinstance(var, dict) and set(var.keys()) == set(llm_keys):
            return True
        return False
