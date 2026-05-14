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
  AT_CLEAN_UP          1 开启（默认启用）
  AT_CLEAN_UP_MODULE   限定清理模块名
"""
from __future__ import annotations

import os
from typing import Any, Dict, Optional, Tuple
from urllib.parse import urlparse


def _strip(s: Any) -> str:
    return (s if s is not None else "").strip()


def _to_bool(v: Any, default: bool = False) -> bool:
    s = _strip(v).lower()
    if not s:
        return default
    if s in ("1", "true", "yes", "on", "y"):
        return True
    if s in ("0", "false", "no", "off", "n"):
        return False
    return default


def isf_enabled(ini_config: Dict[str, Dict[str, str]]) -> bool:
    """
    是否启用 ISF（是否默认携带 Authorization）。
    优先读取运行参数 AT_ISF（由 --isf 注入），
    其次 [server].isf，再次 [server].isf_default；都未配置时默认 True。
    """
    env_v = _strip(os.environ.get("AT_ISF"))
    if env_v:
        return _to_bool(env_v, default=True)
    srv = ini_config.get("server") or {}
    if "isf" in srv:
        return _to_bool(srv.get("isf"), default=True)
    if "isf_default" in srv:
        return _to_bool(srv.get("isf_default"), default=True)
    return True


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
    """AT_SERVER_HOST > SERVER_HOST > [server].host(+public_port)。"""
    h = _strip(os.environ.get("AT_SERVER_HOST") or os.environ.get("SERVER_HOST"))
    if h:
        return h
    srv = ini_config.get("server") or {}
    host = _strip(srv.get("host"))
    if not host:
        return ""
    if ":" in host:
        return host
    public_port = _strip(srv.get("public_port"))
    if public_port:
        return "%s:%s" % (host, public_port)
    return host


def server_public_port(ini_config: Dict[str, Dict[str, str]]) -> str:
    """读取 [server].public_port。"""
    srv = ini_config.get("server") or {}
    return _strip(srv.get("public_port"))


def _host_has_port(host: str) -> bool:
    """判断 host 是否已包含端口（支持常见 host:port 写法）。"""
    if not host:
        return False
    try:
        # 借助 urlparse 统一解析 host:port / ipv4:port 场景
        p = urlparse("http://%s" % host)
        return p.port is not None
    except Exception:
        return False


def request_scheme(ini_config: Dict[str, Dict[str, str]]) -> str:
    """AT_REQUEST_SCHEME > REQUEST_SCHEME > base_url 协议；默认 https。"""
    v = _strip(os.environ.get("AT_REQUEST_SCHEME") or os.environ.get("REQUEST_SCHEME"))
    if v:
        return v.lower()
    # 从 base_url 中解析协议
    base_url = _strip(ini_config.get("server", {}).get("base_url", ""))
    if base_url:
        parsed = urlparse(base_url)
        if parsed.scheme:
            return parsed.scheme
    return "https"


def resolve_request_target(ini_config: Dict[str, Dict[str, str]]) -> Tuple[str, str]:
    """返回 (scheme, host)，用于拼接 API URL。"""
    scheme = request_scheme(ini_config)
    host = server_host_for_requests(ini_config)
    port = server_public_port(ini_config)
    # 仅当显式配置了非 443 端口时，自动拼接到请求 host。
    if host and port and port != "443" and not _host_has_port(host):
        host = "%s:%s" % (host, port)
    return scheme, host


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
    return True  # 默认启用清理


def clean_up_module_name(ini_config: Dict[str, Dict[str, str]]) -> str:
    v = _strip(os.environ.get("AT_CLEAN_UP_MODULE"))
    if v:
        return v
    return ""
