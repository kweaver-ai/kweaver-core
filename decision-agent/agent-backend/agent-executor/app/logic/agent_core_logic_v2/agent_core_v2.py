from typing import Any, AsyncGenerator, Dict, Optional, TYPE_CHECKING
from dolphin.core import flags
from app.common.exceptions.tool_interrupt import ToolInterruptException
from dolphin.core.common.constants import KEY_SESSION_ID, KEY_USER_ID
from dolphin.core.common.exceptions import (
    ModelException,
    SkillException,
    DolphinException,
)

from app.common.config import Config
from app.common.stand_log import StandLogger
from app.common.errors import DolphinSDKException
from app.domain.vo.agentvo import AgentInputVo, AgentConfigVo, AgentRunOptionsVo

from .exception import ExceptionHandler
from .interrupt import InterruptHandler
from .memory import MemoryHandler
from .output import OutputHandler
from .warm_up import WarmUpHandler
from .cache_handler import CacheHandler

from app.utils.snow_id import snow_id
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span
from app.utils.observability.observability_log import get_logger as o11y_logger

from .trace import span_set_attrs
from app.domain.enum.common.user_account_header_key import (
    get_user_account_id,
    get_user_account_type,
    set_user_account_id,
    set_user_account_type,
)
from .input_handler_pkg import (
    process_input,
    process_tool_input,
)

from .run_dolphin import run_dolphin
from .agent_instance_manager import agent_instance_manager

if TYPE_CHECKING:
    from dolphin.sdk.agent.dolphin_agent import DolphinAgent
    from app.router.agent_controller_pkg.rdto.v2.req.resume_agent import ResumeInfo


# AgentConfigVo

from dolphin.core.executor.executor import Executor as DolphinExecutor


class AgentCoreV2:
    is_warmup: bool = False

    agent_config: AgentConfigVo = None

    tool_dict: Dict[str, Any] = {}

    temp_files: Dict[str, Any] = {}

    agent_input: AgentInputVo = None

    agent_output: Dict[str, Any] = {}

    agent_run_id: str = ""

    memory_handler: MemoryHandler = None

    output_handler: OutputHandler = None

    cache_handler: CacheHandler = None

    warmup_handler: WarmUpHandler = None

    def __init__(self, agent_config: AgentConfigVo = None, is_warmup: bool = False):
        self.executor: DolphinExecutor = None
        self.tool_dict = {}
        self.temp_files = {}
        self.agent_config: AgentConfigVo = agent_config

        self.is_warmup = is_warmup

        self.run_options_vo: AgentRunOptionsVo = AgentRunOptionsVo()

        self.memory_handler: MemoryHandler = MemoryHandler()
        self.output_handler: OutputHandler = OutputHandler(self)

        self.cache_handler: CacheHandler = CacheHandler(self)
        self.warmup_handler: WarmUpHandler = WarmUpHandler(self)

    def cleanup(self):
        if self.executor:
            # self.executor.shutdown()
            self.executor = None
        if self.tool_dict:
            for tool_name, tool_instance in self.tool_dict.items():
                del tool_instance
            self.tool_dict.clear()

    @staticmethod
    def remove_context_from_response(response: Dict[str, Any]) -> Dict[str, Any]:
        """
        从响应中移除context键的公共方法

        Args:
            response: 原始响应字典

        Returns:
            移除了context键的响应字典
        """
        if "context" in response:
            response = response.copy()
            del response["context"]
        return response

    def set_run_options(self, run_options_vo: AgentRunOptionsVo):
        self.run_options_vo = run_options_vo

    def _get_resume_info_from_options(self) -> Optional["ResumeInfo"]:
        """从 run_options 中获取 resume_info，并转换为 ResumeInfo 类型"""
        if (
            hasattr(self.run_options_vo, "resume_info")
            and self.run_options_vo.resume_info
        ):
            resume_info_data = self.run_options_vo.resume_info

            # 如果已经是 ResumeInfo 类型，直接返回
            from app.router.agent_controller_pkg.rdto.v2.req.resume_agent import (
                ResumeInfo,
            )

            if isinstance(resume_info_data, ResumeInfo):
                return resume_info_data

            # 如果是 dict 类型（JSON 反序列化后），转换为 ResumeInfo 对象
            if isinstance(resume_info_data, dict):
                return ResumeInfo(**resume_info_data)

            # 其他情况尝试直接返回
            return resume_info_data
        return None

    def _get_registered_agent_instance(self, agent_run_id: str) -> "DolphinAgent":
        """根据 agent_run_id 获取已注册的 DolphinAgent 实例

        Args:
            agent_run_id: Agent 运行 ID

        Returns:
            DolphinAgent 实例

        Raises:
            ValueError: 当实例不存在时
        """
        result = agent_instance_manager.get(agent_run_id)
        if result is None:
            raise ValueError(
                f"Agent instance not found for agent_run_id: {agent_run_id}"
            )

        agent, _ = result
        return agent

    @internal_span(name="invoke_agent")
    async def run(
        self,
        agent_config: AgentConfigVo,
        agent_input: AgentInputVo,
        headers: Dict[str, str],
        is_debug: bool = False,
        span: Optional[Span] = None,
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """运行Agent的核心函数

        Args:
            config: Agent配置字典
            inputs: 输入参数字典
            headers: HTTP请求头

        Yields:
            Dict[str, Any]: 包含状态和答案的字典
        """

        span_set_attrs(
            span=span,
            agent_run_id=agent_config.agent_run_id,
            agent_id=agent_config.agent_id,
            user_id=get_user_account_id(headers) or "",
            conversation_id=agent_config.conversation_id,
            is_root_span=True,
        )

        self.agent_config = agent_config
        self.agent_input = agent_input

        event_key = agent_config.agent_run_id or "agent-session-" + str(snow_id())

        # 保存到实例变量，供外部访问
        self.agent_run_id = event_key

        res = {}

        try:
            # 处理不同类型的输入
            temp_files = await process_input(
                agent_config, agent_input, headers, is_debug
            )

            # 处理工具输入
            (
                context_variables,
                new_event_key,
            ) = await process_tool_input(agent_input)

            # 更新event_key
            if new_event_key:
                event_key = new_event_key

            # 将 sessionid 添加到 context_variables 中
            context_variables[KEY_SESSION_ID] = event_key
            context_variables[KEY_USER_ID] = get_user_account_id(headers) or "unknown"

            # 根据配置启用 dophin explore v2
            if Config.features.use_explore_block_v2:
                # context_variables["explore_block_v2"] = "true"
                flags.set_flag(flags.EXPLORE_BLOCK_V2, True)
            else:
                flags.set_flag(flags.EXPLORE_BLOCK_V2, False)

            # 禁用dolphin sdk llm缓存
            if Config.features.disable_dolphin_sdk_llm_cache:
                flags.set_flag(flags.DISABLE_LLM_CACHE, True)
            else:
                flags.set_flag(flags.DISABLE_LLM_CACHE, False)

            # 将认证信息添加到 context_variables 中
            set_user_account_id(context_variables, get_user_account_id(headers) or "")
            set_user_account_type(
                context_variables, get_user_account_type(headers) or ""
            )

            if Config.llm_message_logging and Config.llm_message_logging.enabled:
                flags.set_flag(flags.LLM_MESSAGE_LOGGING, True)
                flags.set_param(
                    flags.LLM_MESSAGE_LOG_DIR,
                    Config.llm_message_logging.log_dir,
                )
            else:
                flags.set_flag(flags.LLM_MESSAGE_LOGGING, False)
                flags.set_param(flags.LLM_MESSAGE_LOG_DIR, "")

            # 获取输出变量
            output_vars = agent_config.output_vars or []

            try:
                # 判断是否为恢复执行场景
                resume_info = self._get_resume_info_from_options()

                if resume_info:
                    # 恢复执行：使用 resume_dolphin_agent_run
                    from .resume_dolphin_agent_run import resume_dolphin_agent_run

                    # 获取已注册的 DolphinAgent 实例
                    agent = self._get_registered_agent_instance(event_key)

                    StandLogger.info(f"Resume agent with agent_run_id: {event_key}")
                    output_generator = resume_dolphin_agent_run(
                        self,
                        agent,
                        event_key,
                        resume_info,
                        agent_config,
                        context_variables,
                        headers,
                        is_debug,
                    )
                else:
                    # 首次执行：走原有 run_dolphin 流程
                    output_generator = run_dolphin(
                        self,
                        agent_config,
                        context_variables,
                        headers,
                        is_debug,
                        temp_files,
                    )

                if output_vars:
                    output_generator = self.output_handler.partial_output(
                        output_generator, output_vars
                    )

                async for res in output_generator:
                    # 在yield前移除context键并添加 agent_run_id
                    res_with_run_id = self.remove_context_from_response(res)
                    res_with_run_id["agent_run_id"] = event_key
                    yield res_with_run_id

                StandLogger.info("AgentCore run end")

            except ToolInterruptException as tool_interrupt:
                # 处理工具中断

                await InterruptHandler.handle_tool_interrupt(
                    tool_interrupt, res, context_variables
                )
                # 在yield前移除context键并添加 agent_run_id
                res_with_run_id = self.remove_context_from_response(res)
                res_with_run_id["agent_run_id"] = event_key
                yield res_with_run_id

            except (ModelException, SkillException, DolphinException) as e:
                dolphin_except = DolphinSDKException(
                    raw_exception=e,
                    agent_id=agent_config.agent_id,
                    session_id=agent_config.agent_run_id,
                    user_id=get_user_account_id(headers) or "",
                )
                await ExceptionHandler.handle_exception(dolphin_except, res, headers)
                o11y_logger().error(f"agent run failed: {e}")
                # 在yield前移除context键并添加 agent_run_id
                res_with_run_id = self.remove_context_from_response(res)
                res_with_run_id["agent_run_id"] = event_key
                yield res_with_run_id
            except Exception as e:
                # 处理其他异常
                await ExceptionHandler.handle_exception(e, res, headers)
                o11y_logger().error(f"agent run failed: {e}")
                # 在yield前移除context键并添加 agent_run_id
                res_with_run_id = self.remove_context_from_response(res)
                res_with_run_id["agent_run_id"] = event_key
                yield res_with_run_id

        except Exception as e:
            # 处理整体异常
            await ExceptionHandler.handle_exception(e, res, headers)
            o11y_logger().error(f"agent run failed: {e}")
            # 在yield前移除context键并添加 agent_run_id
            res_with_run_id = self.remove_context_from_response(res)
            res_with_run_id["agent_run_id"] = event_key
            yield res_with_run_id
