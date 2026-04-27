from typing import Any, Dict, Optional, TYPE_CHECKING

from app.common.config import BuiltinIds
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span

from ..trace import span_set_attrs

from .process_skill_pkg import process_skills
from app.domain.enum.common.user_account_header_key import get_user_account_id
from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo, SkillInputVo, SkillVo

if TYPE_CHECKING:
    from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2


@internal_span()
async def build_skills(
    ac: "AgentCoreV2",
    headers: Dict[str, str],
    llm_config: Dict[str, Any],
    context_variables: Dict,
    temp_files: Dict[str, Any] = None,
    span: Optional[Span] = None,
):
    config = ac.agent_config

    span_set_attrs(
        span=span,
        agent_id=config.agent_id or "",
        user_id=get_user_account_id(headers) or "",
    )

    # 确保skills不为None
    if config.skills is None:
        config.skills = SkillVo()

    # 构造skills
    # 文件处理tools
    for file_name, file_info in temp_files.items():
        if file_info:
            process_file_tool = ToolSkillVo(
                tool_id=BuiltinIds.get_tool_id("process_file_intelligent"),
                tool_box_id=BuiltinIds.get_tool_box_id("文件处理工具"),
                tool_input=[
                    SkillInputVo(
                        enable=True,
                        input_name="llm",
                        input_type="object",
                        map_type="model",
                        map_value=llm_config["llms"][llm_config["default"]],
                    ),
                    SkillInputVo(
                        enable=True,
                        input_name="token",
                        input_type="string",
                        map_type="var",
                        map_value="header.token",
                    ),
                ],
            )
            config.skills.tools.append(process_file_tool)
            break

    # memory 工具
    if config.memory and config.memory.get("is_enabled"):
        build_memory_tool = ToolSkillVo(
            tool_id=BuiltinIds.get_tool_id("build_memory"),
            tool_box_id=BuiltinIds.get_tool_box_id("记忆管理"),
            tool_input=[
                SkillInputVo(
                    enable=True,
                    input_name="user_id",
                    input_type="string",
                    map_type="auto",
                    map_value="null",
                ),
                SkillInputVo(
                    enable=True,
                    input_name="agent_id",
                    input_type="string",
                    map_type="auto",
                    map_value="null",
                ),
                SkillInputVo(
                    enable=True,
                    input_name="run_id",
                    input_type="string",
                    map_type="auto",
                    map_value="null",
                ),
                SkillInputVo(
                    enable=True,
                    input_name="messages",
                    input_type="array",
                    map_type="auto",
                    map_value="null",
                ),
                SkillInputVo(
                    enable=True,
                    input_name="metadata",
                    input_type="object",
                    map_type="auto",
                    map_value="null",
                ),
            ],
            intervention=False,
        )
        config.skills.tools.append(build_memory_tool)

        search_memory_tool = ToolSkillVo(
            tool_id=BuiltinIds.get_tool_id("search_memory"),
            tool_box_id=BuiltinIds.get_tool_box_id("记忆管理"),
            tool_input=[
                SkillInputVo(
                    enable=True,
                    input_name="user_id",
                    input_type="string",
                    map_type="auto",
                    map_value="null",
                ),
                SkillInputVo(
                    enable=True,
                    input_name="agent_id",
                    input_type="string",
                    map_type="auto",
                    map_value="null",
                ),
                SkillInputVo(
                    enable=True,
                    input_name="run_id",
                    input_type="string",
                    map_type="auto",
                    map_value="null",
                ),
                SkillInputVo(
                    enable=True,
                    input_name="query",
                    input_type="string",
                    map_type="auto",
                    map_value="null",
                ),
                SkillInputVo(
                    enable=True,
                    input_name="limit",
                    input_type="integer",
                    map_type="auto",
                    map_value="null",
                ),
                SkillInputVo(
                    enable=True,
                    input_name="threshold",
                    input_type="number",
                    map_type="auto",
                    map_value="null",
                ),
            ],
            intervention=False,
        )
        config.skills.tools.append(search_memory_tool)

    # 构造tools实例
    await process_skills(ac, headers, context_variables, temp_files)
