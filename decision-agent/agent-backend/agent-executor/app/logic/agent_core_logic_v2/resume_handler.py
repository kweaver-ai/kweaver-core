# -*- coding:utf-8 -*-
"""Resume 处理器 - 独立的恢复执行逻辑"""

from typing import AsyncGenerator, TYPE_CHECKING

from app.utils.json import json_serialize_async
from app.common.stand_log import StandLogger
from app.utils.interrupt_converter import interrupt_handle_to_resume_handle
from .agent_instance_manager import agent_instance_manager

if TYPE_CHECKING:
    from dolphin.sdk.agent.dolphin_agent import DolphinAgent
    from .agent_core_v2 import AgentCoreV2
    from app.router.agent_controller_pkg.rdto.v2.req.resume_agent import ResumeInfo


async def create_resume_generator(
    agent: "DolphinAgent",
    agent_core: "AgentCoreV2",
    agent_run_id: str,
    resume_info: "ResumeInfo",
) -> AsyncGenerator[str, None]:
    """创建恢复执行的生成器（独立方法）

    Args:
        agent: DolphinAgent 实例
        agent_core: AgentCoreV2 实例
        agent_run_id: Agent 运行 ID
        resume_info: 恢复执行信息

    Yields:
        JSON 序列化的输出字符串
    """

    try:
        # 1. 构造 updates 参数
        tool_args = []
        if resume_info.modified_args:
            for arg in resume_info.modified_args:
                tool_args.append({"key": arg.key, "value": arg.value})

        updates = {
            "tool": {
                "action": resume_info.action,
                "tool_args": tool_args,
            }
        }

        if resume_info.action == "skip":
            updates["__skip_tool__"] = True

        # 2. 调用 SDK resume
        # 将 API 层的 InterruptHandle 转换为 Dolphin SDK 的 ResumeHandle
        resume_handle = interrupt_handle_to_resume_handle(resume_info.resume_handle)
        await agent.resume(updates=updates, resume_handle=resume_handle)

        # 3. 继续执行 arun（使用公共方法）
        from .interrupt_utils import process_arun_loop

        async for output in process_arun_loop(agent, is_debug=False):
            # 添加 resume 特有的字段
            output["status"] = "False"
            output["agent_run_id"] = agent_run_id
            yield await json_serialize_async(output)

        # 4. 完成，清理实例
        # agent_instance_manager.remove(agent_run_id)
        final_output = {
            "answer": {},
            "status": "True",
            "agent_run_id": agent_run_id,
        }
        yield await json_serialize_async(final_output)

    except Exception as e:
        StandLogger.error(f"Resume agent error: {e}")
        error_output = {
            "answer": {},
            "status": "Error",
            "agent_run_id": agent_run_id,
            "error": str(e),
        }
        yield await json_serialize_async(error_output)
        agent_instance_manager.remove(agent_run_id)
