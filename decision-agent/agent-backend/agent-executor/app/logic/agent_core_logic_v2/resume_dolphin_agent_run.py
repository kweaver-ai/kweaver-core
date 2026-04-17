from typing import Any, AsyncGenerator, Dict, Optional, TYPE_CHECKING
from dolphin.sdk.agent.dolphin_agent import DolphinAgent
from opentelemetry.trace import Span

# from DolphinLanguageSDK.context_engineer.core.context_manager import (
#     ContextManager,
# )

from app.domain.vo.agentvo import AgentConfigVo
from app.utils.observability.trace_wrapper import internal_span
from .dialog_log import DialogLogHandler
from .trace import span_set_attrs


from app.utils.interrupt_converter import interrupt_handle_to_resume_handle

if TYPE_CHECKING:
    from .agent_core_v2 import AgentCoreV2
    from dolphin.sdk.agent.dolphin_agent import DolphinAgent
    from app.router.agent_controller_pkg.rdto.v2.req.resume_agent import ResumeInfo


@internal_span()
async def resume_dolphin_agent_run(
    ac: "AgentCoreV2",
    agent: "DolphinAgent",
    agent_run_id: str,
    resume_info: "ResumeInfo",
    config: AgentConfigVo,
    context_variables: Dict[str, Any],
    headers: Dict[str, str],
    is_debug: bool = False,
    span: Optional[Span] = None,
) -> AsyncGenerator[Dict[str, Any], None]:
    """运行Dolphin引擎处理请求

    Args:
        config: Agent配置
        context_variables: 上下文变量
        headers: HTTP请求头

    Yields:
        Dict[str, Any]: Dolphin引擎生成的响应
    """

    span_set_attrs(
        span=span,
        agent_run_id=agent_run_id or "",
        agent_id=config.agent_id or "",
        conversation_id=config.conversation_id or "",
    )

    # 1. 验证 data 必须是 dict
    data = resume_info.data
    if not isinstance(data, dict):
        raise ValueError(f"resume_info.data must be dict, got {type(data)}")

    # 2. 从 data 中提取信息
    tool_name = data.get("tool_name", "")
    tool_args = data.get("tool_args", [])

    # 3. 合并 modified_args（参考 test_tool_interrupt.py 实现）
    final_args = [dict(arg) for arg in tool_args]  # 复制原始参数
    if resume_info.modified_args:
        for mod_arg in resume_info.modified_args:
            for i, arg in enumerate(final_args):
                if arg["key"] == mod_arg.key:
                    final_args[i]["value"] = mod_arg.value
                    break

    # 4. 构造 updates 参数
    updates = {
        "tool": {
            "tool_name": tool_name,
            "tool_args": final_args,
            "action": resume_info.action,
        }
    }

    if resume_info.action == "skip":
        updates["__skip_tool__"] = True

    # 4. 调用 SDK resume
    # 将 API 层的 InterruptHandle 转换为 Dolphin SDK 的 ResumeHandle
    resume_handle = interrupt_handle_to_resume_handle(resume_info.resume_handle)
    await agent.resume(updates=updates, resume_handle=resume_handle)

    # 12. 执行agent
    output = {}

    # 使用公共的 arun 循环处理方法
    from .interrupt_utils import process_arun_loop

    async for output in process_arun_loop(agent, is_debug):
        yield output

    yield output

    # 使用新的日志生成函数
    dialog_log_handler = DialogLogHandler(agent, config, headers)
    dialog_log_handler.save_dialog_logs()

    ac.memory_handler.start_memory_build_thread(
        config, context_variables, headers, output
    )
