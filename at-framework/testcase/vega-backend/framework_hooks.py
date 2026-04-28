"""Framework Hooks: dataflow"""
import os
import sys
import importlib
from pathlib import Path

import allure
import requests
import urllib3
import time

urllib3.disable_warnings(urllib3.connectionpool.InsecureRequestWarning)

try:
    at_env = importlib.import_module("common.at_env")
except ImportError:
    # 兼容从 testcase 子目录执行时，common 不在默认 sys.path 的场景
    framework_root = Path(__file__).resolve().parents[2]
    framework_root_str = str(framework_root)
    if framework_root_str not in sys.path:
        sys.path.insert(0, framework_root_str)
    at_env = importlib.import_module("common.at_env")


def _default_bearer_auth(config):
    """
    默认 Bearer：
    - AT_AUTH_SOURCE=login 时尝试 get_token
    - 否则使用静态 token（API_ACCESS_TOKEN / test_data.application_token）
    """
    source = at_env.auth_token_source(config)
    if source == "login":
        try:
            token_provider = importlib.import_module("src.common.token_provider")
            get_token = token_provider.get_token
            clear_token_cache = token_provider.clear_token_cache

            user, pwd = at_env.admin_credentials(config)
            if user:
                clear_token_cache(user)
                tok = at_env.normalize_bearer_token_value(get_token(user, pwd, force_refresh=True) or "")
                if tok:
                    return "Bearer %s" % tok
        except Exception as ex:
            allure.attach(str(ex), name="get_token 异常")

    static_token = at_env.static_access_token(config)
    if static_token:
        return "Bearer %s" % static_token
    # isf=true 时要求默认都带 Authorization 头，token 缺失时保留 Bearer 空值
    return "Bearer "


def session_setup(session_id, config):
    """会话开始前的设置"""
    pass


def session_clean_up(config, allure):
    """
    会话结束后清理本次测试创建的数据
    """
    print("\n========== 开始清理 vega-backend 测试数据 ==========")

    os.environ["AT_ISF"]

    server_config = config.get("server", {})
    host = server_config.get("host", "")
    port = server_config.get("public_port", "443")
    base_url = f"https://{host}:{port}"

    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json",
    }

    if at_env.isf_enabled(config):
        headers["Authorization"] = _default_bearer_auth(config)

    print("\n========== vega-backend 测试数据清理完成 ==========\n")


def test_setup(test_id, config):
    pass


def test_teardown(test_id, config):
    pass
