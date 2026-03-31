from typing import Any, Dict
import json

from app.common.stand_log import StandLogger

# Import from common module using relative import
from ..common import ToolMapInfo


def process_fixed_value(
    tool_map_item: ToolMapInfo,
    current_tool_input: Dict[str, Any],
    input_params: Dict[str, Any],
):
    """处理 fixedValue 类型的映射值"""

    val = tool_map_item.get_map_value()
    input_type = tool_map_item.input_type

    if input_type != "string":
        try:
            tool_map_item.map_value = json.loads(val)
        except Exception:
            StandLogger.warn(
                f"工具的输入参数{tool_map_item.input_name}的值{val}不是json格式"
            )
            tool_map_item.map_value = val
    else:
        if val.startswith('"') and val.endswith('"'):
            tool_map_item.map_value = json.loads(val)

    current_tool_input[tool_map_item.input_name] = tool_map_item.get_map_value()
