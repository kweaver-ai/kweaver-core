from typing import Any, Dict
import json

from app.utils.dict_util import get_dict_val_by_path
from dolphin.core.context.context import Context
from app.common.stand_log import StandLogger

from app.utils.common import (
    get_dolphin_var_value,
)

# Import from common module using relative import
from ..common import ToolMapInfo
from .process_fixed_value import process_fixed_value


def process_children(
    tool_map_item: ToolMapInfo,
    current_tool_input: Dict[str, Any],
    input_params: Dict[str, Any],
    gvp: Context,
):
    """处理工具映射项的子节点"""
    process_needed = False

    input_name = tool_map_item.input_name

    if input_name not in current_tool_input:
        process_needed = True
        current_tool_input[input_name] = {}

    for child_item in tool_map_item.children:
        process_tool_map_item(
            child_item,
            current_tool_input[input_name],
            input_params.get("properties", {}).get(input_name, {}),
            gvp,
        )

    if process_needed and current_tool_input[input_name] == {}:
        current_tool_input.pop(input_name)


def old_process_fixed_value(
    tool_map_item: ToolMapInfo,
    current_tool_input: Dict[str, Any],
    input_params: Dict[str, Any],
):
    """处理 fixedValue 类型的映射值"""

    val = tool_map_item.get_map_value()
    input_type = tool_map_item.input_type

    # 之前的判断方式（可能不对，暂时保留代码，并注释）
    # input_params.get("properties", {})
    # .get(tool_map_item.input_name, {})
    # .get("type", "")
    # != "string"

    if isinstance(val, str):
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
            else:
                tool_map_item.map_value = val

    current_tool_input[tool_map_item.input_name] = tool_map_item.get_map_value()


def process_map_type(
    tool_map_item: ToolMapInfo,
    current_tool_input: Dict[str, Any],
    input_params: Dict[str, Any],
    gvp: Context,
):
    """处理工具映射项的 map_type"""
    map_type = tool_map_item.get_map_type()

    if map_type == "auto":
        return

    elif map_type == "var":
        cite_var = tool_map_item.get_map_value()
        # 递归获取变量值
        all_variables = gvp.get_all_variables()

        cite_var_value = get_dict_val_by_path(all_variables, cite_var)

        cite_var_value = get_dolphin_var_value(cite_var_value)

        current_tool_input[tool_map_item.input_name] = cite_var_value

    elif map_type == "fixedValue":
        process_fixed_value(tool_map_item, current_tool_input, input_params)
    else:
        current_tool_input[tool_map_item.input_name] = tool_map_item.get_map_value()


def process_tool_map_item(
    tool_map_item: ToolMapInfo,
    current_tool_input: Dict[str, Any],
    input_params: Dict[str, Any],
    gvp: Context,
):
    """递归处理单个工具映射项"""

    input_name = tool_map_item.input_name
    # 1. 检查是否启用
    if tool_map_item.is_enabled() is False:
        if input_name in current_tool_input:
            current_tool_input.pop(input_name)
        return

    # 2. 处理 children 递归情况
    if tool_map_item.children:
        process_children(tool_map_item, current_tool_input, input_params, gvp)
        return

    # 3. 处理 map_type
    process_map_type(tool_map_item, current_tool_input, input_params, gvp)
