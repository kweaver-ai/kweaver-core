# -*- coding:utf-8 -*-
import asyncio
import gettext
import inspect
import os
import sys
import traceback
from concurrent.futures import ThreadPoolExecutor
from enum import Enum
from typing import Any, Callable, Dict, Tuple
from urllib.parse import urlparse

from fastapi import Request
from pydantic import BaseModel

from app.domain.enum.common.user_account_header_key import get_user_account_id

# 使用延迟导入替代直接导入 Dolphin SDK
# from dolphin.core.context.var_output import VarOutput  # ← 移除直接导入


cur_pwd = os.getcwd()


def _get_var_output_class():
    """延迟获取 VarOutput 类

    只有在需要时才导入 Dolphin SDK，避免强制依赖。
    """
    try:
        from app.common.dependencies import get_dolphin_var_output_class

        return get_dolphin_var_output_class()
    except ImportError:
        # 如果依赖模块不可用，返回 None
        return None


def get_caller_info() -> Tuple[str, int]:
    """获取调用者文件项目相对位置以及行号"""
    caller_frame = inspect.stack()[2]
    caller_filename = caller_frame.filename.split(cur_pwd)[-1][1:]
    caller_lineno = caller_frame.lineno
    return caller_filename, caller_lineno


def is_in_pod() -> bool:
    """检查是否在 Kubernetes Pod 中运行"""
    return (
        "KUBERNETES_SERVICE_HOST" in os.environ
        and "KUBERNETES_SERVICE_PORT" in os.environ
    )


# 触发熔断的失败次数
_failure_threshold = 10


def get_failure_threshold() -> int:
    return _failure_threshold


def set_failure_threshold(threshold: int) -> None:
    global _failure_threshold
    _failure_threshold = threshold


# 熔断触发后的再次重试间隔，单位：秒
_recovery_timeout = 5


def get_recovery_timeout() -> int:
    return _recovery_timeout


def set_recovery_timeout(timeout: int) -> None:
    global _recovery_timeout
    _recovery_timeout = timeout


_l = gettext.gettext


def set_lang(lang: Callable[..., str]) -> None:
    global _l
    _l = lang


def get_lang() -> Callable[..., str]:
    return _l


def get_request_lang_func(request: Request) -> Callable[..., str]:
    return get_request_lang_from_header(request.headers)


def get_request_lang_from_header(header: Dict[str, str]) -> Callable[..., str]:
    lang = header.get("x-language", "en")

    if lang.startswith("zh"):
        return gettext.translation(
            "messages", localedir="app/common/international", languages=["zh"]
        ).gettext

    return gettext.gettext


def get_unknown_error(
    file_name: str, func_name: str, details: str, lang_func: Callable[..., str]
) -> Dict[str, str]:
    return {
        "description": lang_func("Unknown error occurred!"),
        "solution": lang_func("Please check the service."),
        "error_code": "AgentExecutor.InternalServerError.UnknownError",
        "error_details": details,
        "error_link": "",
    }


def convert_to_camel_case(string: str) -> str | None:
    """字符串的下划线格式，小驼峰格式转为大驼峰模式，不支持非下划线全小写(helloworld)转大驼峰"""
    if not isinstance(string, str):
        return None

    words = string.split("_")
    capitalized_words = []
    for word in words:
        if len(word) == 1:
            capitalized_words.append(word.upper())
        elif len(word) > 1:
            capitalized_words.append(word[0].upper() + word[1:])
    return "".join(capitalized_words)


def get_user_id_by_request(request: Request) -> str:
    """从请求中获取用户ID"""
    return get_user_account_id(request.headers) or ""


def convert_to_valid_class_name(name: str) -> str:
    """将字符串转换为合法的类名"""
    if not name:
        return ""
    name = "".join(c if c.isalnum() else "_" for c in name)
    if name[0].isdigit():
        name = "_" + name
    return name


def truncate_by_byte_len(text: str, length: int = 65535) -> str:
    """
    将文本按照指定字节长度进行截断
    Args:
        text: 待截断的文本
        length: 截断的字节长度，默认值65535是数据库text类型的长度
    """
    char_length = min(len(text), length)
    while len(text[:char_length].encode("utf-8")) > length:
        char_length -= 1
    return text[:char_length]


def create_subclass(
    base_class: type, class_name: str, class_attributes: Dict[str, Any]
) -> type:
    """使用type动态创建子类"""
    return type(class_name, (base_class,), class_attributes)


def is_valid_url(url: str) -> bool:
    """判断是否为有效的URL地址"""
    try:
        result = urlparse(url)
        return all([result.scheme, result.netloc])
    except (ValueError, AttributeError):
        return False


def func_judgment(func: object) -> Tuple[bool, bool]:
    """
    判断函数是否流式返回以及是否异步
    Args:
        func: 函数对象
    Returns:
        (asynchronous, stream): 异步标志和流式标志
    """
    asynchronous = False
    stream = False
    if inspect.iscoroutinefunction(func):
        asynchronous = True
        if inspect.isasyncgenfunction(func):
            stream = True
    else:
        if inspect.isgeneratorfunction(func):
            stream = True
    if inspect.isasyncgenfunction(func):
        asynchronous = True
        stream = True
    return asynchronous, stream


def sync_wrapper(async_func: Callable, *args: Any, **kwargs: Any) -> Any:
    """在同步函数中调用异步函数"""
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    try:
        result = loop.run_until_complete(async_func(*args, **kwargs))
        return result
    finally:
        loop.close()


def run_async_in_thread(async_func: Callable, *args: Any, **kwargs: Any) -> Any:
    """在单独的线程中运行异步函数"""
    with ThreadPoolExecutor() as executor:
        future = executor.submit(sync_wrapper, async_func, *args, **kwargs)
        return future.result()


def make_json_serializable(obj: Any) -> Any:
    """将不可json序列化的对象转为可json序列化的对象"""
    if isinstance(obj, list):
        return [make_json_serializable(item) for item in obj]
    elif isinstance(obj, tuple):
        return make_json_serializable(list(obj))
    elif isinstance(obj, dict):
        result = {}
        for k, v in obj.items():
            if k == "embedding" and isinstance(v, list):
                result[k] = None
            else:
                result[k] = make_json_serializable(v)
        return result
    elif isinstance(obj, BaseModel):
        return make_json_serializable(obj.model_dump())
    elif isinstance(obj, Enum):
        return make_json_serializable(obj.value)
    elif isinstance(obj, float):
        import math

        if math.isnan(obj):
            return None
        return obj
    return obj


async def get_format_error_info(
    header: Dict[str, str], exc: Exception
) -> Dict[str, str]:
    """获取格式化的错误信息"""
    lang_func = get_request_lang_from_header(header)
    if hasattr(exc, "FormatHttpError"):
        return exc.FormatHttpError(lang_func)

    traceback_list = traceback.extract_tb(sys.exc_info()[2])
    if traceback_list:
        file_name, line_number, func_name, line_text = traceback_list[-1]
        file_name = os.path.basename(file_name).split(".")[0]
    else:
        file_name = "common"
        func_name = "common"
    return get_unknown_error(file_name, func_name, repr(exc), lang_func)


def is_dolphin_var(var: Any) -> bool:
    """检查是否为 Dolphin 变量"""
    # 使用延迟导入获取 VarOutput 类，保持原始逻辑
    VarOutputClass = _get_var_output_class()
    if VarOutputClass is not None:
        try:
            return VarOutputClass.is_serialized_dict(var)
        except (AttributeError, TypeError):
            pass

    # 如果 VarOutput 不可用，使用与原始实现相同的逻辑
    # 原始实现：isinstance(data, dict) and data.get("__type__") == "VarOutput"
    return isinstance(var, dict) and var.get("__type__") == "VarOutput"


def get_dolphin_var_value(var: Any) -> Any:
    """获取 Dolphin 变量的值"""
    if is_dolphin_var(var):
        if isinstance(var.get("value"), list):
            return [get_dolphin_var_value(item) for item in var.get("value")]
        else:
            return get_dolphin_var_value(var.get("value"))
    return var


def get_dolphin_var_final_value(var: Any) -> Any:
    """获取 Dolphin 变量的最终值"""
    if is_dolphin_var(var):
        source_type = var.get("source_type")
        if source_type == "EXPLORE":
            value = var.get("value")
            if isinstance(value, list) and len(value) > 0:
                return get_dolphin_var_final_value(value[-1]["answer"])
            elif isinstance(value, dict):
                return get_dolphin_var_final_value(value["answer"])
        elif source_type == "LLM":
            return var.get("value", {}).get("answer")
        else:
            return get_dolphin_var_value(var)
    return var


# 兼容老版命名的别名
GetCallerInfo = get_caller_info
IsInPod = is_in_pod
GetFailureThreshold = get_failure_threshold
SetFailureThreshold = set_failure_threshold
GetRecoveryTimeout = get_recovery_timeout
SetRecoveryTimeout = set_recovery_timeout
GetRequestLangFunc = get_request_lang_func
GetRequestLangFromHeader = get_request_lang_from_header
GetUnknowError = get_unknown_error
ConvertToCamelCase = convert_to_camel_case
GetUserIDByRequest = get_user_id_by_request
