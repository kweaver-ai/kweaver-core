from typing import Any, AsyncGenerator, Dict, List, Optional, TYPE_CHECKING
import time
from datetime import datetime

from app.domain.vo.agentvo import AgentConfigVo, AgentInputVo
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span

from app.utils.json import json_serialize_async
from app.utils.increment_json import incremental_async_generator

from app.utils.common import (
    is_dolphin_var,
)

from .trace import span_set_attrs
from app.domain.enum.common.user_account_header_key import get_user_account_id

import asyncio

if TYPE_CHECKING:
    from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2


class OutputHandler:
    def __init__(self, agent_core: "AgentCoreV2"):
        self.agent_core = agent_core

    async def string_output(
        self,
        generator: AsyncGenerator[Dict[str, Any], None],
    ) -> AsyncGenerator[str, None]:
        """将字典输出转换为JSON字符串

        Args:
            generator: 原始生成器

        Yields:
            str: JSON字符串
        """
        _loop = asyncio.get_event_loop()

        async for chunk in generator:
            yield await json_serialize_async(chunk)

    async def add_status(
        self,
        generator: AsyncGenerator[Dict[str, Any], None],
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """添加状态字段到输出中

        Args:
            generator: 原始生成器

        Yields:
            Dict[str, Any]: 添加了状态字段的输出
        """
        chunk = None
        async for chunk in generator:
            if "status" not in chunk:
                chunk["status"] = "False"
            yield chunk

        if chunk and chunk.get("status") == "False":
            chunk["status"] = "True"
            yield chunk

    async def add_ttft(
        self,
        generator: AsyncGenerator[Dict[str, Any], None],
        start_time: float,
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """添加TTFT（Time To First Token）字段到输出中

        Args:
            generator: 原始生成器
            start_time: 请求开始时间

        Yields:
            Dict[str, Any]: 添加了TTFT字段的输出
        """
        # 计算TTFT（毫秒）
        _ttft_ms = int((time.time() - start_time) * 1000)

        async for chunk in generator:
            # 在最外层添加ttft字段
            # chunk["ttft"] = ttft_ms
            yield chunk

    async def add_datetime(
        self,
        generator: AsyncGenerator[Dict[str, Any], None],
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """添加日期时间字段到输出中

        Args:
            generator: 原始生成器

        Yields:
            Dict[str, Any]: 添加了日期时间字段的输出
        """
        # 生成格式化的日期时间（精确到毫秒）
        _dt = datetime.now().strftime("%Y-%m-%d %H:%M:%S.%f")[:-3]

        async for chunk in generator:
            # 在最外层添加datetime字段
            # chunk["datetime"] = dt
            yield chunk

    @internal_span()
    async def result_output(
        self,
        agent_config: AgentConfigVo,
        agent_input: AgentInputVo,
        headers: Dict[str, str],
        is_debug_mode: bool = False,
        span: Optional[Span] = None,
        start_time: Optional[float] = None,
    ) -> AsyncGenerator[str, None]:
        """处理最终输出结果

        Args:
            config: Agent配置
            inputs: 输入参数
            headers: HTTP请求头
            is_debug_mode: 是否调试模式
            span: OpenTelemetry span
            start_time: 请求开始时间（用于计算TTFT）

        Yields:
            str: 最终输出的JSON字符串
        """

        span_set_attrs(
            span=span,
            agent_run_id=agent_config.agent_run_id or "",
            agent_id=agent_config.agent_id or "",
            user_id=get_user_account_id(headers) or "",
        )

        try:
            # 1. 执行agent
            output_generator = self.agent_core.run(
                agent_config, agent_input, headers, is_debug_mode
            )

            # 2. 处理输出
            if agent_config.incremental_output:
                output_generator = incremental_async_generator(output_generator)
            elif not agent_config.output_vars:
                output_generator = self.add_status(output_generator)

            # 2.1 如果提供了start_time，添加TTFT字段
            if start_time is not None:
                output_generator = self.add_ttft(output_generator, start_time)

            # 2.2 添加日期时间字段
            output_generator = self.add_datetime(output_generator)

            # 2.3 转换为JSON字符串
            output_generator = self.string_output(output_generator)

            # 2.4 遍历输出
            async for chunk in output_generator:
                # print(chunk)
                yield chunk

        finally:
            self.agent_core.cleanup()

    @internal_span()
    async def partial_output(
        self,
        dolphin_output: AsyncGenerator[Dict[str, Any], None],
        output_vars: List[str],
        span: Optional[Span] = None,
    ) -> AsyncGenerator[Any, None]:
        """处理输出变量

        Args:
            dolphin_output: Dolphin输出生成器
            output_vars: 输出变量列表

        Yields:
            Any: 处理后的输出
        """
        res = {}
        async for output in dolphin_output:
            if len(output_vars) == 1:
                fields = output_vars[0].split(".")
                value = output
                for field in fields:
                    try:
                        value = value[field]
                        if is_dolphin_var(value):
                            value = value.get("value")
                    except (KeyError, TypeError):
                        pass
                res = value
            elif len(output_vars) > 1:
                for output_var in output_vars:
                    if not output_var:
                        continue
                    fields = output_var.split(".")
                    value = output
                    has_value = True
                    for field in fields:
                        try:
                            value = value[field]
                            if is_dolphin_var(value):
                                value = value.get("value")
                        except (KeyError, TypeError):
                            has_value = False
                    if has_value:
                        res[field] = value
            else:
                res = output
            yield res
