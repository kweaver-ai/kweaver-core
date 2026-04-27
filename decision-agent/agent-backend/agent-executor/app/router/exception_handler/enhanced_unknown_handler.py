# -*- coding:utf-8 -*-
"""
增强版未知异常处理器

使用 exception_logger 提供更好的异常日志输出：
1. 控制台：仅项目代码的 Traceback，带颜色高亮
2. 简单文件日志：仅项目代码的 Traceback
3. 详细文件日志：完整的 Traceback（包含第三方库）
"""

import os
import sys
import traceback
from datetime import datetime
from typing import Dict, Any

from fastapi import Request
from fastapi.responses import JSONResponse

from app.common.exception_logger import exception_logger
from app.utils.common import GetRequestLangFunc, GetUnknowError
from app.domain.enum.common.user_account_header_key import (
    get_user_account_id,
    get_user_account_type,
    get_biz_domain_id,
)


def cache_request_body(request: Request, body: Any) -> None:
    """
    缓存请求体到 request.state（在中间件中调用）

    Args:
        request: FastAPI 请求对象
        body: 请求体
    """
    request.state.cached_body = body


def _extract_request_info(request: Request) -> Dict[str, Any]:
    """
    从请求中提取详细信息

    Args:
        request: FastAPI 请求对象

    Returns:
        Dict: 请求信息（包含 URL、method、headers、body 等）
    """
    info = {
        "method": request.method,
        "path": request.url.path,
        "url": str(request.url),
    }

    # 查询字符串
    if request.url.query:
        info["query_string"] = request.url.query

    # 客户端 IP
    if request.client:
        info["client_ip"] = request.client.host

    # 用户信息（使用统一的 header 获取方法）
    headers_dict_for_account = dict(request.headers)
    try:
        account_id = get_user_account_id(headers_dict_for_account)
        if account_id:
            info["account_id"] = account_id

        account_type = get_user_account_type(headers_dict_for_account)
        if account_type:
            info["account_type"] = account_type
    except (KeyError, Exception):
        pass  # 忽略获取失败

    # 业务域
    biz_domain = get_biz_domain_id(headers_dict_for_account)
    if biz_domain:
        info["biz_domain"] = biz_domain

    # 提取所有 headers
    headers_dict = {}
    for key, value in request.headers.items():
        # 过滤敏感信息
        if key.lower() in {"authorization", "cookie", "x-api-key"}:
            headers_dict[key] = "***REDACTED***"
        else:
            headers_dict[key] = value
    info["headers"] = headers_dict

    # 尝试获取缓存的请求体（从 request.state 中获取）
    try:
        if hasattr(request.state, "cached_body"):
            info["body"] = request.state.cached_body
    except Exception:
        pass  # 忽略获取失败

    return info


def _get_actual_exception(exc: Exception) -> Exception:
    """
    获取实际的异常（处理 ExceptionGroup）

    Args:
        exc: 原始异常

    Returns:
        Exception: 实际异常
    """
    if hasattr(exc, "__class__") and exc.__class__.__name__ == "ExceptionGroup":
        if hasattr(exc, "exceptions") and exc.exceptions:
            return exc.exceptions[0]
    return exc


def handle_enhanced_unknown_exception(request: Request, exc: Exception) -> JSONResponse:
    """
    增强版未知异常处理器

    Args:
        request: FastAPI 请求对象
        exc: 未知异常（可能是 ExceptionGroup）

    Returns:
        JSONResponse: 错误响应
    """
    timestamp = datetime.now()

    # 提取请求信息
    request_info = _extract_request_info(request)

    # 使用 exception_logger 记录异常
    exception_logger.log_exception(exc, request_info, timestamp)

    # 获取实际异常用于响应
    actual_exc = _get_actual_exception(exc)

    # 获取翻译函数
    tr = GetRequestLangFunc(request)

    # 获取异常位置信息
    traceback_list = traceback.extract_tb(sys.exc_info()[2])
    if traceback_list:
        file_name, line_number, func_name, line_text = traceback_list[-1]
        file_name = os.path.basename(file_name).split(".")[0]
    else:
        file_name = "common"
        func_name = "common"

    # 返回错误响应
    return JSONResponse(
        status_code=500,
        content=GetUnknowError(file_name, func_name, repr(actual_exc), tr),
    )
