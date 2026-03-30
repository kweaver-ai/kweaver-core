# coding: utf-8
"""统一 Token 提供者：读取配置并通过 Login 流程获取 access token（支持按用户缓存与强制刷新）。"""
from typing import Optional, Tuple

from src.auth.login import Login
from src.common.logger import logger
from src.config.setting import config as setting_config

_token_cache: dict[Tuple[str, str, str], str] = {}
_ini_cfg = None


def _framework_ini():
    """与 conftest 同源的 config.ini 字典，用于环境变量优先的 host 解析。"""
    global _ini_cfg
    if _ini_cfg is None:
        from common.func import load_sys_config

        _ini_cfg = load_sys_config("./config/config.ini")
    return _ini_cfg


def _get_hosts() -> Tuple[Optional[str], Optional[str]]:
    """AT_SERVER_HOST > 与 conftest 一致的 ini；再回退 src.config.setting。"""
    try:
        from common import at_env

        host = at_env.server_host_for_requests(_framework_ini())
        if host:
            return host, host
    except Exception as e:
        logger.warning("从环境变量/框架配置读取 host 失败：%s", e)
    try:
        host = setting_config.get("server", "host")
        client_host = setting_config.get("server", "host")
        return host, client_host
    except Exception as e:
        logger.warning("读取 host/client_host 配置失败：%s", e)
        return None, None


def get_token(user: str, password: str, force_refresh: bool = False) -> Optional[str]:
    """
    获取指定用户的 access token。
    - user/password：账号密码（建议配合 AT_ADMIN_USER / AT_ADMIN_PASSWORD）
    - force_refresh：是否强制刷新（例如 401/403 后重试）
    """
    host, client_host = _get_hosts()
    if not host or not client_host:
        logger.warning("未找到登录所需的 host/client_host 配置，跳过获取 token。")
        return None

    cache_key = (host, client_host, user)
    if not force_refresh and cache_key in _token_cache:
        return _token_cache[cache_key]

    try:
        login = Login().UserLogin(host, client_host, user, password)
        if isinstance(login, list) and len(login) >= 3:
            token = login[2]
            _token_cache[cache_key] = token
            logger.info("成功获取 access token 并缓存（user=%s）。", user)
            return token
        logger.error("获取 token 失败：返回结果格式不正确或登录失败。")
        return None
    except Exception as e:
        logger.error("获取 token 发生异常：%s", e)
        return None


def clear_token_cache(user: Optional[str] = None):
    """清理 token 缓存；若传入 user 则只清理该用户对应缓存。"""
    global _token_cache
    if user is None:
        _token_cache.clear()
    else:
        host, client_host = _get_hosts()
        if host and client_host:
            _token_cache.pop((host, client_host, user), None)


def peek_cached_token(user: str) -> Optional[str]:
    """仅从缓存读取 token（不触发登录）。"""
    host, client_host = _get_hosts()
    if not host or not client_host:
        return None
    return _token_cache.get((host, client_host, user))


if __name__ == "__main__":
    t = get_token("test", "111111")
    print(t)
