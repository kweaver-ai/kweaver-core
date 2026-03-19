import os

import allure
import pytest

from common.func import load_sys_config, load_case, get_cases
from common.hooks import load_session_clean_up

config = load_sys_config("./config/config.ini")
_case_file = config["env"]["case_file"]
_module_name = os.path.basename(os.path.normpath(os.path.abspath(_case_file))) if _case_file else ""

_case_list_cache = None


def compute_case_list():
    """延迟加载用例列表：在 test_run 收集时执行，避免仅跑 tests/ 单元测试时加载业务模块。"""
    global _case_list_cache
    if _case_list_cache is not None:
        return _case_list_cache
    _scope = os.environ.get("SCOPE", "").strip()
    _tags = os.environ.get("TAGS", "").strip()
    _api_name = os.environ.get("API_NAME", "").strip()
    _api_path = os.environ.get("API_PATH", "").strip()
    _case_names = os.environ.get("CASE_NAMES", "").strip()
    if _case_names:
        names_list = [x.strip() for x in _case_names.split(",") if x.strip()]
        _case_list_cache = get_cases(_case_file, names=names_list) if names_list else load_case(_case_file)
    elif _scope or _tags or _api_name or _api_path:
        _case_list_cache = get_cases(_case_file, scope=_scope or None, tags=_tags or None,
                                     api_name=_api_name or None, api_path=_api_path or None)
    else:
        _case_list_cache = load_case(_case_file)
    return _case_list_cache


BEARER_AUTH = "Bearer %s" % config["external"].get("token", "")


def pytest_collection_modifyitems(items) -> None:
    for item in items:
        item.name = item.name.encode("utf-8").decode("unicode-escape")
        item._nodeid = item.nodeid.encode("utf-8").decode("unicode-escape")


@pytest.fixture(scope="session", autouse=True)
@allure.step("会话清理（模块可选）")
def clean_up(request):
    # 仅当本会话实际收集到 test_run 中的 API 用例时才执行（避免只跑 tests/ 时误连环境）
    try:
        items = request.session.items or []
        if not any("test_run.py::" in getattr(i, "nodeid", "") or i.nodeid.startswith("test_run.py") for i in items):
            return
    except Exception:
        pass
    if int(config["env"].get("clean_up", "0")) != 1:
        return
    clean_up_module = config["env"].get("clean_up_module", "").strip()
    if clean_up_module and clean_up_module != _module_name:
        return

    case_dir = os.path.abspath(_case_file)
    fn = load_session_clean_up(case_dir)
    if fn is None:
        print(
            "未注册会话清理：模块 %s 无 framework_hooks.session_clean_up，跳过（adp 其他模块可不加此文件）"
            % _module_name
        )
        return
    fn(config, allure)
