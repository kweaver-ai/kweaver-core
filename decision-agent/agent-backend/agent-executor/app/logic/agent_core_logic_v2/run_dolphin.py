import json
from typing import Any, AsyncGenerator, Dict, Optional, TYPE_CHECKING
from dolphin.sdk.agent.dolphin_agent import DolphinAgent
from dolphin.core.config.global_config import GlobalConfig
from dolphin.sdk.skill.traditional_toolkit import TriditionalToolkit
from dolphin.core.common.constants import KEY_HISTORY

from app.common.config import Config
from app.common.stand_log import StandLogger
from app.common.struct_logger import struct_logger
from app.domain.vo.agentvo import AgentConfigVo
from app.logic.agent_core_logic_v2.output_variables import get_output_variables
from app.logic.agent_core_logic_v2.prompt_builder import PromptBuilder
from app.common.tool_v2.tool import build_tools
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span
from app.utils.observability.observability_log import get_logger as o11y_logger
from .dialog_log import DialogLogHandler

from .trace import span_set_attrs
from .input_handler_pkg import (
    build_llm_config,
    build_skills,
)
from app.domain.enum.common.user_account_header_key import (
    get_user_account_id,
    get_user_account_type,
)

if TYPE_CHECKING:
    from .agent_core_v2 import AgentCoreV2

from .agent_instance_manager import agent_instance_manager


@internal_span()
async def run_dolphin(
    ac: "AgentCoreV2",
    config: AgentConfigVo,
    context_variables: Dict[str, Any],
    headers: Dict[str, str],
    is_debug: bool = False,
    temp_files: Dict[str, Any] = None,
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
        agent_run_id=config.agent_run_id or "",
        agent_id=config.agent_id or "",
        user_id=get_user_account_id(headers) or "",
        conversation_id=config.conversation_id or "",
    )

    # 从headers中提取user_id和visitor_type
    user_id = get_user_account_id(headers) or ""
    visitor_type = get_user_account_type(headers) or ""

    # 1. 构造dolphin使用的LLM参数
    llm_config = await build_llm_config(ac, user_id, visitor_type)

    # 2. 构造dolphin_prompt
    prompt_builder = PromptBuilder(config, temp_files)
    dolphin_prompt = await prompt_builder.build()

    # 3. 构造skills
    await build_skills(ac, headers, llm_config, context_variables, temp_files)

    # 3.1 构造tool_dict
    from app.domain.vo.agentvo.agent_config_vos import SkillVo

    skills = ac.agent_config.skills if ac.agent_config is not None else SkillVo()

    # warmup模式下从这里就可以返回了，不实际run
    if ac.is_warmup:
        return

    tool_dict = await build_tools(ac, skills, request_headers=headers)

    # 3.2 构造toolkit
    toolkit = TriditionalToolkit.buildFromTooldict(tool_dict)

    # 4. 记录信息
    # ANSI颜色码定义
    COLORS = {
        "header": "\033[95m",  # 紫色
        "blue": "\033[94m",  # 蓝色
        "cyan": "\033[96m",  # 青色
        "green": "\033[92m",  # 绿色
        "yellow": "\033[93m",  # 黄色
        "red": "\033[91m",  # 红色
        "bold": "\033[1m",  # 粗体
        "underline": "\033[4m",  # 下划线
        "end": "\033[0m",  # 结束符
    }

    from enum import Enum

    class EnumEncoder(json.JSONEncoder):
        def default(self, o):
            if isinstance(o, Enum):
                return o.value
            return super().default(o)

    StandLogger.info_log(
        f"{COLORS['header']}{COLORS['bold']}Agent execution details:{COLORS['end']}\n"
        f"{COLORS['blue']}========================================{COLORS['end']}\n"
        f"{COLORS['cyan']}{COLORS['bold']}Dolphin Language Prompt:{COLORS['end']}\n{dolphin_prompt}\n"
        f"{COLORS['blue']}----------------------------------------{COLORS['end']}\n"
        f"{COLORS['green']}{COLORS['bold']}Context Variables:{COLORS['end']} {json.dumps(context_variables, indent=2, ensure_ascii=False)}\n"
        f"{COLORS['blue']}----------------------------------------{COLORS['end']}\n"
        f"{COLORS['yellow']}{COLORS['bold']}Skill Kit Tolls:{COLORS['end']} {json.dumps(toolkit.tools, indent=2, ensure_ascii=False, default=str)}\n"
        f"{COLORS['blue']}----------------------------------------{COLORS['end']}\n"
        f"{COLORS['red']}{COLORS['bold']}LLM Config:{COLORS['end']} {json.dumps(llm_config, indent=2, ensure_ascii=False, cls=EnumEncoder)}\n"
        f"{COLORS['blue']}========================================{COLORS['end']}"
    )


    # 8. 从llm_config字典创建GlobalConfig对象
    global_config = GlobalConfig.from_dict(llm_config)
    # 只启用内置日期函数
    global_config.skill_config.enabled_skills = [
        "system_functions._date",
    ]

    # 9. 创建DolphinAgent
    # 9.1 根据配置设置output_variables
    output_variables = None
    if Config.features.enable_dolphin_agent_output_variables_ctrl:
        output_variables = get_output_variables(ac)
        # 本地配置覆盖
        if Config.local_dev.dolphin_agent_output_variables:
            output_variables = Config.local_dev.dolphin_agent_output_variables
        # 打印output_variables
        struct_logger.console_logger.debug(
            f"[run_dolphin] output_variables: {output_variables}"
        )

    # ctx_manager = ContextManager()

    # 9.2 创建trace listener（如果启用）
    trace_listener = None
    if Config.is_dolphin_trace_enabled():
        try:
            from dolphin.core.observability.otel_listener import OTelTraceListener

            trace_listener = OTelTraceListener(
                agent_id=config.agent_id or "",
                conversation_id=config.conversation_id or "",
                user_id=user_id or "",
            )
        except Exception as e:
            StandLogger.warn(f"[run_dolphin] Failed to create trace listener: {e}")
            o11y_logger().warn(f"[run_dolphin] Failed to create trace listener: {e}")
            trace_listener = None

    # 适配 Dolphin SDK 的 history 变量名：
    # SDK 内部使用 KEY_HISTORY 常量（当前值为 "_history"）而不是 "history"
    # 这里引用 SDK 常量避免硬编码，自动跟随 SDK 的变化
    if "history" in context_variables:
        context_variables[KEY_HISTORY] = context_variables.pop("history")

    agent = DolphinAgent(
        content=dolphin_prompt,
        name=f"agent_core_v2_{config.agent_id}",
        skillkit=toolkit,
        variables=context_variables,
        global_config=global_config,
        verbose=Config.app.enable_dolphin_agent_verbose,  # 启用详细输出模式
        log_level=Config.app.get_stdlib_log_level(),
        output_variables=output_variables,
        trace_listener=trace_listener,
        # context_manager=ctx_manager,
    )
    # todo control by config
    # from DolphinLanguageSDK import flags

    # flags.set_flag(flags.DISABLE_LLM_CACHE, True)
    # 10. 初始化agent
    await agent.initialize()

    # 11. 保存到ac中以便后续使用
    ac.agent = agent
    ac.executor = agent.executor  # 保持向后兼容，某些代码可能需要访问executor

    # 11.1 注册到实例管理器（用于后续 resume/terminate）
    agent_run_id = ac.agent_run_id
    if agent_run_id:
        agent_instance_manager.register(agent_run_id, agent, ac)

    # 12. 执行agent
    output = {}

    # 使用公共的 arun 循环处理方法
    from .interrupt_utils import process_arun_loop

    async for output in process_arun_loop(agent, is_debug):
        yield output

    yield output
    # StandLogger.debug_log(
    #     f"LLM块专家模式输出: {json.dumps(item_value, ensure_ascii=False)}"
    # )

    # 使用新的日志生成函数
    dialog_log_handler = DialogLogHandler(agent, config, headers)
    dialog_log_handler.save_dialog_logs()

    ac.memory_handler.start_memory_build_thread(
        config, context_variables, headers, output
    )
