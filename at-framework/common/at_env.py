# -*- coding: utf-8 -*-
"""
统一环境变量配置（优先于 config.ini 同义项）。

静态 token（不含「Bearer 」前缀，若误带前缀会自动去掉），优先级：
  API_ACCESS_TOKEN > [test_data].application_token

鉴权方式（默认 Bearer 用静态 token 还是走 get_token）：
  AT_AUTH_SOURCE      login | static（默认 login）

get_token 账号（仅读取 [test_data]）：
  admin_user
  admin_password

请求基址（缺省读 [server]）：
  AT_SERVER_HOST

模块与其它（与现有变量并存）：
  AT_CASE_FILE        与 CASE_FILE 相同含义，任一存在即可

会话清理：
  AT_CLEAN_UP          1 开启（默认关闭）
  AT_CLEAN_UP_MODULE   限定清理模块名
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
    """静态 access token：环境变量 API_ACCESS_TOKEN > [test_data].application_token。"""
    v = normalize_bearer_token_value(os.environ.get("API_ACCESS_TOKEN", ""))
    if v:
        return v
    td = ini_config.get("test_data") or {}
    return normalize_bearer_token_value(td.get("application_token", ""))


def auth_token_source(ini_config: Dict[str, Dict[str, str]]) -> str:
    """login | static：环境变量优先，默认 login。"""
    v = _strip(os.environ.get("AT_AUTH_SOURCE") or os.environ.get("AUTH_TOKEN_SOURCE"))
    if v:
        return v.lower()
    return "login"


def admin_credentials(ini_config: Dict[str, Dict[str, str]]) -> Tuple[str, str]:
    """仅从 [test_data].admin_user / admin_password 读取账号密码。"""
    td = ini_config.get("test_data") or {}
    u = _strip(td.get("admin_user"))
    p = _strip(td.get("admin_password"))
    return u, p


def server_host_for_requests(ini_config: Dict[str, Dict[str, str]]) -> str:
    """AT_SERVER_HOST > SERVER_HOST > [server].host。"""
    h = _strip(os.environ.get("AT_SERVER_HOST") or os.environ.get("SERVER_HOST"))
    if h:
        return h
    srv = ini_config.get("server") or {}
    return _strip(srv.get("host"))


def request_scheme(ini_config: Dict[str, Dict[str, str]]) -> str:
    """从 [server].base_url 推导协议；无有效 base_url 时默认为 https。"""
    srv = ini_config.get("server") or {}
    base_url = _strip(srv.get("base_url", ""))
    if "://" in base_url:
        return base_url.split("://", 1)[0].rstrip(":/").lower() or "https"
    return "https"


def resolve_request_target(ini_config: Dict[str, Dict[str, str]]) -> Tuple[str, str]:
    """返回 (scheme, host)，用于拼接 API URL。"""
    return request_scheme(ini_config), server_host_for_requests(ini_config)


def default_case_file(ini_config: Dict[str, Dict[str, str]]) -> str:
    """AT_CASE_FILE > CASE_FILE > ./testcase/etrino。"""
    v = _strip(os.environ.get("AT_CASE_FILE") or os.environ.get("CASE_FILE"))
    if v:
        return v
    return "./testcase/etrino"


def clean_up_enabled(ini_config: Dict[str, Dict[str, str]]) -> bool:
    v = _strip(os.environ.get("AT_CLEAN_UP"))
    if v:
        return v == "1"
    return False


def clean_up_module_name(ini_config: Dict[str, Dict[str, str]]) -> str:
    v = _strip(os.environ.get("AT_CLEAN_UP_MODULE"))
    if v:
        return v
    return ""
