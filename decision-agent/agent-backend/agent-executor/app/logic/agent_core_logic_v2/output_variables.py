from typing import TYPE_CHECKING, List

from app.infra.common.infra_constant.const import FINAL_ANSWER_DEFAULT_VAR

if TYPE_CHECKING:
    from .agent_core_v2 import AgentCoreV2


def get_output_variables(ac: "AgentCoreV2") -> List[str]:
    """获取输出变量配置"""

    # 1. 先获取配置中的所有输出变量
    output_vars: List[str] = ac.agent_config.output.get_all_vars()

    # 2. 除了上面那个以外其他需要输出的字段 如果有其他需要输出的字段，可以在这里添加
    # 说明：
    # a. interventions可能没有用到(agent-app/src/domain/valueobject/agentrespvo/answer_data.go有，暂时保留)
    to_add = [
        FINAL_ANSWER_DEFAULT_VAR,
        "interventions",
        "intervention_judge_block_vars",
        "intervention_tool_block_vars",
        "tool",
    ]

    # 3. is_need_progress
    if ac.run_options_vo.is_need_progress:
        to_add.append("_progress")

    # 4. 合并
    output_vars.extend([var for var in to_add if var not in output_vars])

    # 5. 返回
    return output_vars
