import ast
import json
from typing import Dict, Optional, Any

from app.domain.vo.agentvo import AgentConfigVo, AgentInputVo
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span

from ..trace import span_set_attrs
from app.domain.enum.common.user_account_header_key import get_user_account_id


@internal_span()
async def process_input(
    agent_config: AgentConfigVo,
    agent_input: AgentInputVo,
    headers: Dict[str, str],
    is_debug: bool = False,
    span: Optional[Span] = None,
) -> Dict[str, Any]:
    """处理不同类型的输入"""
    span_set_attrs(
        span=span,
        agent_run_id=agent_config.agent_run_id,
        agent_id=agent_config.agent_id,
        user_id=get_user_account_id(headers) or "",
    )

    temp_files = {}

    for input_field in agent_config.input.get("fields", []):
        if input_field.get("type") == "string":
            var_name = input_field.get("name")
            var_value = agent_input.get_value(var_name)
            if var_value:
                agent_input.set_value(var_name, str(var_value))
            else:
                agent_input.set_value(var_name, "")

        elif input_field.get("type") == "object":
            var_name = input_field.get("name")
            var_value = agent_input.get_value(var_name)
            if var_value:
                # 非string类型的保持原样
                if not isinstance(var_value, str):
                    continue
                # string类型传入，则解析
                try:
                    var_value = json.loads(var_value)
                except (json.JSONDecodeError, TypeError):
                    try:
                        var_value = ast.literal_eval(var_value)
                    except (ValueError, SyntaxError):
                        var_value = str(var_value)
                agent_input.set_value(var_name, var_value)
            else:
                agent_input.set_value(var_name, {})

        elif input_field.get("type") == "file":
            # 如果debug为True，调试页面的临时区传的是文件的元信息
            # 否则是文件内容
            var_name = input_field.get("name")

            file_infos = agent_input.get_value(var_name)
            if not file_infos:
                file_infos = []

            agent_input.set_value(var_name, file_infos)

            temp_files[var_name] = file_infos

    # 为内置变量赋值

    if not agent_input.get_value("history"):
        agent_input.set_value("history", [])

    agent_input.header = headers
    agent_input.self_config = agent_config.model_dump()

    return temp_files
