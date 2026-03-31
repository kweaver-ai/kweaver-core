"""
控制台日志格式化器
"""

import json
import traceback
from structlog.types import EventDict, WrappedLogger

from .constants import COLORS, RESET, BOLD, DIM, LEVEL_EMOJI
from .utils import safe_json_serialize


def _parse_event_content(raw_event):
    """
    解析事件内容,提取关键信息

    Args:
        raw_event: 原始事件内容

    Returns:
        tuple: (event, extracted_info) 事件描述和提取的额外信息
    """
    event = raw_event
    extracted_info = {}

    # 处理字典类型的事件
    if isinstance(raw_event, dict):
        event = (
            raw_event.get("description") or raw_event.get("message") or str(raw_event)
        )
        for k, v in raw_event.items():
            if k not in ("description", "message"):
                extracted_info[k] = v
    # 尝试将字符串事件解析为字典
    elif isinstance(raw_event, str) and (
        raw_event.startswith("{") or len(raw_event) > 200
    ):
        try:
            import re

            json_match = re.search(r"\{.*\}", raw_event)
            if json_match:
                parsed = json.loads(json_match.group())
                if isinstance(parsed, dict):
                    event = parsed.get("description", parsed.get("message", ""))
                    for k, v in parsed.items():
                        if k not in ("description", "message"):
                            extracted_info[k] = v
        except Exception:
            pass

    # 如果事件太长，截断
    if len(str(event)) > 150:
        event = str(event)[:150] + "..."

    return event, extracted_info


def _format_log_header(level, event, timestamp):
    """
    格式化日志头部(级别、事件、时间戳)

    Args:
        level: 日志级别
        event: 事件描述
        timestamp: 时间戳

    Returns:
        str: 格式化后的头部
    """
    level_color = COLORS.get(level.lower(), COLORS["info"])
    emoji = LEVEL_EMOJI.get(level, "📝")

    return (
        f"{emoji} {BOLD}{level_color}[{level}]{RESET} "
        f"{BOLD}{event}{RESET} "
        f"{COLORS['timestamp']}({timestamp}){RESET}"
    )


def _format_location_and_type(caller, log_type):
    """
    格式化调用位置和日志类型

    Args:
        caller: 调用位置
        log_type: 日志类型

    Returns:
        list: 格式化后的行列表
    """
    lines = []

    if caller:
        lines.append(f'  {COLORS["caller"]}📍 Location: "{caller}"{RESET}')

    if log_type:
        lines.append(
            f"  {COLORS['key']}🏷️  Type:{RESET} {COLORS['value']}{log_type}{RESET}"
        )

    return lines


def _format_call_stack(call_stack):
    """
    格式化 call_stack 字段 - 多行显示方便跳转

    Args:
        call_stack: 堆栈信息 (字符串或数组格式)

    Returns:
        list: 格式化后的行列表
    """
    lines = []
    lines.append(f"    {BOLD}{COLORS['exception_type']}call_stack:{RESET}")

    if isinstance(call_stack, str):
        # 字符串格式的堆栈
        for line in call_stack.strip().split("\n"):
            if "File" in line:
                # 文件路径行使用特殊颜色，方便点击跳转
                lines.append(f"      {COLORS['caller']}{line}{RESET}")
            elif any(
                line.strip().startswith(x) for x in ["Traceback", "Error", "Exception"]
            ):
                # 异常标题行
                lines.append(f"      {BOLD}{COLORS['exception_type']}{line}{RESET}")
            elif ":" in line and not line.startswith(" "):
                # 异常类型和消息行
                lines.append(f"      {COLORS['exception_type']}{line}{RESET}")
            else:
                # 代码行和其他
                lines.append(f"      {COLORS['traceback']}{line}{RESET}")
    elif isinstance(call_stack, list):
        # 数组格式的堆栈 (如你提供的示例)
        for stack_item in call_stack:
            if isinstance(stack_item, dict):
                # 格式化每个堆栈项
                file_path = stack_item.get("file", "")
                line_num = stack_item.get("line", "")
                function_name = stack_item.get("function", "")

                if file_path and line_num and function_name:
                    # 文件路径和行号 - 使用蓝色，方便点击跳转
                    lines.append(
                        f'      {COLORS["caller"]}File "{file_path}", line {line_num}, in {function_name}{RESET}'
                    )
                else:
                    # 备用格式
                    lines.append(
                        f"      {COLORS['caller']}File {file_path}:{line_num} in {function_name}{RESET}"
                    )
            else:
                # 备用格式
                lines.append(f"      {COLORS['value']}{str(stack_item)}{RESET}")
    else:
        # 其他格式，使用默认显示
        lines.append(f"      {COLORS['value']}{str(call_stack)}{RESET}")

    return lines


def _format_context_fields(event_dict):
    """
    格式化上下文字段

    Args:
        event_dict: 事件字典

    Returns:
        list: 格式化后的行列表
    """
    lines = []

    if not event_dict:
        return lines

    # 先收集所有要显示的字段，只有在有实际内容时才添加标题
    content_lines = []

    for key, value in event_dict.items():
        # 跳过已经在其他地方显示的字段
        if key in ("logger", "level", "timestamp", "stack"):
            continue

        # 判断是否是错误相关字段
        error_related_keys = [
            "error",
            "error_message",
            "error_code",
            "error_details",
            "validation_errors",
            "exception",
            "failure",
            "failed",
        ]
        is_error_field = any(err_key in key.lower() for err_key in error_related_keys)
        value_color = COLORS["error_value"] if is_error_field else COLORS["value"]

        # 特殊处理 call_stack 字段 - 多行显示方便跳转
        if key == "call_stack":
            content_lines.extend(_format_call_stack(value))
        # 特殊处理 validation_errors 字段 - 单行展示
        elif key == "validation_errors" and isinstance(value, list):
            try:
                # 压缩 JSON，单行显示
                value_str = json.dumps(value, ensure_ascii=False, separators=(",", ":"))
            except (TypeError, ValueError):
                try:
                    safe_value = safe_json_serialize(value)
                    value_str = json.dumps(
                        safe_value, ensure_ascii=False, separators=(",", ":")
                    )
                except Exception:
                    value_str = str(value)

            content_lines.append(
                f"    {COLORS['key']}{key}:{RESET} {value_color}{value_str}{RESET}"
            )
        # 格式化其他类型的值
        elif isinstance(value, (dict, list)):
            try:
                value_str = json.dumps(value, ensure_ascii=False, indent=4)
            except (TypeError, ValueError):
                try:
                    safe_value = safe_json_serialize(value)
                    value_str = json.dumps(safe_value, ensure_ascii=False, indent=4)
                except Exception:
                    value_str = str(value)

            # 多行值缩进
            value_lines = value_str.split("\n")
            content_lines.append(
                f"    {COLORS['key']}{key}:{RESET} {value_color}{value_lines[0]}{RESET}"
            )
            for vl in value_lines[1:]:
                content_lines.append(f"      {value_color}{vl}{RESET}")
        else:
            value_str = str(value)
            content_lines.append(
                f"    {COLORS['key']}{key}:{RESET} {value_color}{value_str}{RESET}"
            )

    # 只有在有实际内容时才添加标题
    if content_lines:
        lines.append(f"  {DIM}Context Fields:{RESET}")
        lines.extend(content_lines)

    return lines


def _format_error_details(error_details):
    """
    格式化错误详情

    Args:
        error_details: 错误详情

    Returns:
        list: 格式化后的行列表
    """
    lines = []

    if not error_details:
        return lines

    lines.append(f"  {BOLD}{COLORS['exception_type']}Error Details:{RESET}")

    if isinstance(error_details, dict):
        for k, v in error_details.items():
            lines.append(
                f"    {COLORS['key']}{k}:{RESET} {COLORS['exception_msg']}{v}{RESET}"
            )
    else:
        lines.append(f"    {COLORS['exception_msg']}{error_details}{RESET}")

    return lines


def _format_traceback_line(line):
    """
    格式化单行堆栈信息

    Args:
        line: 堆栈信息的一行

    Returns:
        str: 格式化后的行
    """
    line_stripped = line.strip()

    # 文件路径行
    if line_stripped.startswith("File") or "File" in line:
        return f"    {COLORS['caller']}{line}{RESET}"
    # 异常类型或错误信息（非缩进的行）
    elif line_stripped and not line.startswith(" "):
        return f"    {COLORS['exception_type']}{line}{RESET}"
    # 其他堆栈信息
    else:
        return f"    {COLORS['traceback']}{line}{RESET}"


def _format_exception_traceback(exc_info, exc_str, stack_str):
    """
    格式化异常堆栈信息

    Args:
        exc_info: 异常对象或异常字符串（可能已被 format_exc_info processor 转换）
        exc_str: 异常字符串
        stack_str: 堆栈字符串

    Returns:
        list: 格式化后的行列表
    """
    lines = []

    if not (exc_info or exc_str or stack_str):
        return lines

    lines.append(f"  {BOLD}{COLORS['exception_type']}Exception Traceback:{RESET}")

    # 获取堆栈文本
    stack_text = None
    if exc_info:
        if isinstance(exc_info, str):
            # 如果是字符串（已被 format_exc_info processor 处理）
            stack_text = exc_info
        elif isinstance(exc_info, BaseException):
            # 如果是异常对象，转换为字符串
            tb_lines = traceback.format_exception(
                type(exc_info), exc_info, exc_info.__traceback__
            )
            stack_text = "".join(tb_lines)
    elif stack_str or exc_str:
        stack_text = stack_str or exc_str

    # 统一格式化堆栈文本
    if stack_text:
        for line in stack_text.rstrip().split("\n"):
            lines.append(_format_traceback_line(line))

    return lines


def format_console_log(
    logger: WrappedLogger, method_name: str, event_dict: EventDict
) -> str:
    """
    增强的控制台日志格式化器
    - 使用颜色突出不同信息
    - trace 分行展示，方便编辑器识别和跳转
    - 添加边界标识区分不同日志
    - 分类展示不同类型的信息

    Args:
        logger: 日志记录器
        method_name: 日志方法名
        event_dict: 事件字典

    Returns:
        格式化后的日志字符串
    """
    # 1. 提取基本信息
    timestamp = event_dict.pop("timestamp", "")
    level = event_dict.pop("level", "").upper()
    raw_event = event_dict.pop("event", "")
    caller = event_dict.pop("caller", "")
    log_type = event_dict.pop("log_type", "")

    # 2. 解析事件内容
    event, extracted_info = _parse_event_content(raw_event)

    # 3. 提取异常信息
    exc_info = event_dict.pop("exception", None)
    error_details = event_dict.pop("error_details", None)
    exc_str = event_dict.pop("exc_info", None)
    stack_str = event_dict.pop("stack", None)

    # 4. 合并提取的信息
    if extracted_info:
        event_dict.update(extracted_info)

    # 5. 再次尝试提取 stack
    if not stack_str and "stack" in event_dict:
        stack_str = event_dict.pop("stack")

    # 6. 移除重复字段
    event_dict.pop("caller", None)

    # 7. 构建输出
    lines = []

    # 7.1 顶部边界
    border_char = "═" if level in ("ERROR", "CRITICAL") else "─"
    lines.append(f"{COLORS['border']}{border_char * 120}{RESET}")

    # 7.2 日志头部
    lines.append(_format_log_header(level, event, timestamp))

    # 7.3 位置和类型
    location_lines = _format_location_and_type(caller, log_type)
    lines.extend(location_lines)

    # 7.4 分隔线
    if location_lines:
        lines.append(f"{COLORS['border']}{'  ' + '┄' * 116}{RESET}")

    # 7.5 上下文字段
    lines.extend(_format_context_fields(event_dict))

    # 7.6 错误详情
    lines.extend(_format_error_details(error_details))

    # 7.7 异常堆栈
    lines.extend(_format_exception_traceback(exc_info, exc_str, stack_str))

    # 7.8 底部边界
    lines.append(f"{COLORS['border']}{border_char * 120}{RESET}\n")

    return "\n".join(lines)
