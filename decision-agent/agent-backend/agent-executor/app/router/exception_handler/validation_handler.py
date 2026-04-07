"""
请求参数验证异常处理器
"""

import inspect
from fastapi import Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse


from app.common.struct_logger import struct_logger
from app.utils.common import GetRequestLangFunc
from app.utils.regex_rules import GetErrorMessageByRegex


def handle_param_error(request: Request, exc: RequestValidationError):
    """
    处理请求参数验证错误

    Args:
        request: FastAPI 请求对象
        exc: 验证错误异常

    Returns:
        JSONResponse: 错误响应
    """
    _l = GetRequestLangFunc(request)

    # 获取请求路径和方法，用于定位错误来源
    request_path = request.url.path
    request_method = request.method

    errorMessage = ""
    for error in exc.errors():
        paramName = " ".join(map(str, error["loc"][1:]))
        if error["type"] == "missing":
            errorMessage += _l(" parameter: ") + paramName + _l(" is required!")
        elif error["type"] == "string_type":
            errorMessage += (
                _l(" parameter: ") + paramName + _l(" must be a string type!")
            )
        elif error["type"] == "int_type":
            errorMessage += _l(" parameter: ") + paramName + _l(" must be an int type!")
        elif error["type"] == "float_type":
            errorMessage += (
                _l(" parameter: ") + paramName + _l(" must be a float type!")
            )
        elif error["type"] == "list_type":
            errorMessage += _l(" parameter: ") + paramName + _l(" must be a list!")
        elif error["type"] == "dict_type":
            errorMessage += _l(" parameter: ") + paramName + _l(" must be a dict!")
        elif error["type"] == "bool_type":
            errorMessage += _l(" parameter: ") + paramName + _l(" must be a bool type!")
        elif error["type"] == "bytes_type":
            errorMessage += _l(" parameter: ") + paramName + _l(" must be a byte type!")
        elif error["type"] == "string_too_short":
            errorMessage += (
                _l(" parameter: ")
                + paramName
                + _l(" length is at least ")
                + str(error["ctx"]["min_length"])
                + _l("!")
            )
        elif error["type"] == "string_too_long":
            errorMessage += (
                _l(" parameter: ")
                + paramName
                + _l(" length is up to ")
                + str(error["ctx"]["max_length"])
                + _l("!")
            )
        elif error["type"] == "greater_than_equal":
            errorMessage += (
                _l(" parameter: ")
                + paramName
                + _l(" is at least ")
                + str(error["ctx"]["ge"])
                + _l("!")
            )
        elif error["type"] == "less_than_equal":
            errorMessage += (
                _l(" parameter: ")
                + paramName
                + _l(" is up to ")
                + str(error["ctx"]["le"])
                + _l("!")
            )
        elif error["type"] == "string_pattern_mismatch":
            errorMessage += (
                _l(" parameter: ")
                + paramName
                + _l(GetErrorMessageByRegex(str(error["ctx"]["pattern"])))
            )
        elif error["type"] == "too_long":
            errorMessage += (
                _l(" parameter: ")
                + paramName
                + _l(" length is up to ")
                + str(error["ctx"]["max_length"])
                + _l("!")
            )
        elif error["type"] == "too_short":
            errorMessage += (
                _l(" parameter: ")
                + paramName
                + _l(" length is at least ")
                + str(error["ctx"]["min_length"])
                + _l("!")
            )
        elif error["type"] == "value_error.list.unique_items":
            # pydantic2 不支持unique_items 需要额外校验
            errorMessage += (
                _l(" parameter: ") + paramName + _l(" has duplicated items!")
            )
        elif error["type"] == "json_invalid":
            errorMessage += _l(" json format error!")
        else:
            # 提供更详细的错误信息
            error_type = error.get("type", "unknown")
            error_msg = error.get("msg", "")
            error_ctx = error.get("ctx", {})

            errorMessage += _l(" parameter: ") + paramName + _l(" is invalid!")
            errorMessage += f" (type: {error_type}"

            # 添加具体的错误信息
            if error_msg:
                errorMessage += f", msg: {error_msg}"

            # 添加上下文信息
            if error_ctx:
                ctx_str = ", ".join([f"{k}={v}" for k, v in error_ctx.items()])
                errorMessage += f", context: {ctx_str}"

            # 添加输入值信息
            if "input" in error:
                input_value = error["input"]
                # 限制输入值的长度，避免日志过长
                input_str = str(input_value)
                if len(input_str) > 100:
                    input_str = input_str[:100] + "..."
                errorMessage += f", input: {input_str}"

            errorMessage += ")"

    # 使用结构化日志记录错误详情
    # 获取堆栈跟踪信息 - 只保留项目相关代码
    stack_trace = []
    try:
        # 获取更多层调用栈（15层），然后只保留项目相关的
        for frame_info in inspect.stack()[1:16]:  # 跳过当前函数，获取后续15层
            filename = frame_info.filename
            # 只记录项目相关的文件
            if "/agent-executor/" in filename:
                # 简化路径，只保留项目内的相对路径
                simplified_path = filename.split("/agent-executor/")[-1]
                stack_trace.append(
                    {
                        "file": simplified_path,
                        "line": frame_info.lineno,
                        "function": frame_info.function,
                    }
                )

        # 如果没有找到项目代码（参数验证在框架层就失败了），添加请求信息作为上下文
        if not stack_trace:
            stack_trace = None  # 设为 None，避免记录空数组
    except Exception as e:
        struct_logger.warning("Failed to get stack trace", error=str(e))
        stack_trace = None

    # 记录结构化错误日志
    # 注意: struct_logger 已内置安全序列化机制,但这里预处理可以提高性能
    from app.common.struct_logger import _safe_json_serialize

    struct_logger.error(
        "Request validation error",
        request_method=request_method,
        request_path=request_path,
        error_message=errorMessage,
        validation_errors=_safe_json_serialize(exc.errors()),
        call_stack=stack_trace if stack_trace else None,
    )

    content = {
        "Description": errorMessage,
        "ErrorCode": "AgentExecutor.BadRequest.ParamError",
        "ErrorDetails": errorMessage,
        "ErrorLink": "",
        "Solution": _l("Please check your parameter again."),
    }

    return JSONResponse(status_code=400, content=content)
