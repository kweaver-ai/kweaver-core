# -*- coding: utf-8 -*-
"""
统一环境变量配置（优先于 config.ini 同义项）。

静态 token（不含「Bearer 」前缀，若误带前缀会自动去掉），优先级：
  AT_API_TOKEN / API_ACCESS_TOKEN > [external].token > [test_data].application_token

鉴权方式（默认 Bearer 用静态 token 还是走 get_token）：
  AT_AUTH_SOURCE      login | static（缺省读 [env] auth_token_source，再无则 static）

get_token 账号（缺省读 [test_data]）：
  AT_ADMIN_USER
  AT_ADMIN_PASSWORD

请求基址（缺省读 [env] / [server]）：
  AT_SERVER_HOST
  AT_REQUEST_SCHEME   https | http

模块与其它（与现有变量并存）：
  AT_CASE_FILE        与 CASE_FILE 相同含义，任一存在即可

会话清理：
  AT_CLEAN_UP          1 开启（覆盖 [env].clean_up）
  AT_CLEAN_UP_MODULE   覆盖 [env].clean_up_module
"""
from __future__ import annotations

import os
from typing import Any, Dict, Optional, Tuple


def _strip(s: Any) -> str:
    return (s if s is not None else "").strip()


def normalize_bearer_token_value(raw: str) -> str:
    """去掉首尾空白；若以 Bearer 开头则去掉该前缀，得到 access_token 本体。"""
    s = _strip(raw)
    if s.lower().startswith("bearer "):
        return s[7:].strip()
    return s


def static_access_token(ini_config: Dict[str, Dict[str, str]]) -> str:
    """静态 access token：环境变量 > [external].token > [test_data].application_token。"""
    for key in ("AT_API_TOKEN", "API_ACCESS_TOKEN"):
        v = normalize_bearer_token_value(os.environ.get(key, ""))
        if v:
            return v
    ext = ini_config.get("external") or {}
    v = normalize_bearer_token_value(ext.get("token", ""))
    if v:
        return v
    td = ini_config.get("test_data") or {}
    return normalize_bearer_token_value(td.get("application_token", ""))


def auth_token_source(ini_config: Dict[str, Dict[str, str]]) -> str:
    """login | static：环境变量 > [env].auth_token_source。"""
    v = _strip(os.environ.get("AT_AUTH_SOURCE") or os.environ.get("AUTH_TOKEN_SOURCE"))
    if v:
        return v.lower()
    env_sec = ini_config.get("env") or {}
    return _strip(env_sec.get("auth_token_source") or env_sec.get("token_source") or "static").lower() or "static"


def admin_credentials(ini_config: Dict[str, Dict[str, str]]) -> Tuple[str, str]:
    """AT_ADMIN_USER/PASSWORD > [test_data].admin_user / admin_password。"""
    u = _strip(os.environ.get("AT_ADMIN_USER"))
    p = _strip(os.environ.get("AT_ADMIN_PASSWORD"))
    td = ini_config.get("test_data") or {}
    if not u:
        u = _strip(td.get("admin_user"))
    if not p:
        p = _strip(td.get("admin_password"))
    return u, p


def server_host_for_requests(ini_config: Dict[str, Dict[str, str]]) -> str:
    """AT_SERVER_HOST > [env].host > [server].host。"""
    h = _strip(os.environ.get("AT_SERVER_HOST") or os.environ.get("SERVER_HOST"))
    if h:
        return h
    env_sec = ini_config.get("env") or {}
    srv = ini_config.get("server") or {}
    return _strip(env_sec.get("host")) or _strip(srv.get("host"))


def request_scheme(ini_config: Dict[str, Dict[str, str]]) -> str:
    """AT_REQUEST_SCHEME > [env].request_scheme > https。"""
    s = _strip(os.environ.get("AT_REQUEST_SCHEME"))
    if s:
        return s.rstrip(":/").lower() or "https"
    env_sec = ini_config.get("env") or {}
    s = _strip(env_sec.get("request_scheme") or "https")
    return s.rstrip(":/").lower() or "https"


def resolve_request_target(ini_config: Dict[str, Dict[str, str]]) -> Tuple[str, str]:
    """返回 (scheme, host)，用于拼接 API URL。"""
    return request_scheme(ini_config), server_host_for_requests(ini_config)


def default_case_file(ini_config: Dict[str, Dict[str, str]]) -> str:
    """AT_CASE_FILE > CASE_FILE > [env].case_file > ./testcase/etrino。"""
    v = _strip(os.environ.get("AT_CASE_FILE") or os.environ.get("CASE_FILE"))
    if v:
        return v
    env_sec = ini_config.get("env") or {}
    return _strip(env_sec.get("case_file")) or "./testcase/etrino"


def clean_up_enabled(ini_config: Dict[str, Dict[str, str]]) -> bool:
    v = _strip(os.environ.get("AT_CLEAN_UP"))
    if v:
        return v == "1"
    env_sec = ini_config.get("env") or {}
    try:
        return int(env_sec.get("clean_up", "0")) == 1
    except ValueError:
        return False


def clean_up_module_name(ini_config: Dict[str, Dict[str, str]]) -> str:
    v = _strip(os.environ.get("AT_CLEAN_UP_MODULE"))
    if v:
        return v
    env_sec = ini_config.get("env") or {}
    return _strip(env_sec.get("clean_up_module"))
