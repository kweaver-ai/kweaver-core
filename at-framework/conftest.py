import os

import allure
import pytest

from common import at_env
from common.func import load_sys_config, load_case, get_cases
from common.hooks import load_session_clean_up

config = load_sys_config("./config/config.ini")


def _default_bearer_auth():
    """默认 Authorization：AT_AUTH_SOURCE 为 login 时走 get_token；否则使用静态 token。"""
    src = at_env.auth_token_source(config)
    static_tok = at_env.static_access_token(config)
    if src in ("login", "get_token", "token_provider"):
        user, pwd = "", ""
        try:
            from src.common.token_provider import get_token

            user, pwd = at_env.admin_credentials(config)
            if user and pwd:
                tok = get_token(user, pwd)
                print("tok", tok)
                if tok:
                    return "Bearer %s" % tok
        except Exception as ex:
            print("user", user, "pwd", pwd)
            print("获取 token 异常:", ex)
    return "Bearer %s" % static_tok if static_tok else "Bearer "


_default_case_file = at_env.default_case_file(config).strip()
_case_file = _default_case_file.strip()
_module_name = os.path.basename(os.path.normpath(os.path.abspath(_case_file))) if _case_file else ""

_case_list_cache = None


def pytest_addoption(parser):
    """支持运行时覆盖模块与筛选范围，避免频繁改 config.ini。"""
    parser.addoption("--case-file", action="store", default=None, help="模块用例目录，如 ./testcase/vega")
    parser.addoption("--scope", action="store", default=None, help="scope id")
    parser.addoption("--tags", action="store", default=None, help="逗号分隔 tags")
    parser.addoption("--api-name", action="store", default=None, help="API 名（apis.yaml 中的 name）")
    parser.addoption("--api-path", action="store", default=None, help="API 路径")
    parser.addoption("--case-names", action="store", default=None, help="逗号分隔 case 名列表")
    parser.addoption("--suite", action="store", default=None, help="套件 story 或 suites 下文件名（不含 .yaml）")


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
    _suite = os.environ.get("SUITE", "").strip()
    if _case_names:
        names_list = [x.strip() for x in _case_names.split(",") if x.strip()]
        _case_list_cache = get_cases(_case_file, names=names_list) if names_list else load_case(_case_file)
    elif _scope or _tags or _api_name or _api_path or _suite:
        _case_list_cache = get_cases(
            _case_file,
            scope=_scope or None,
            tags=_tags or None,
            suite=_suite or None,
            api_name=_api_name or None,
            api_path=_api_path or None,
        )
    else:
        _case_list_cache = load_case(_case_file)
    return _case_list_cache


BEARER_AUTH = _default_bearer_auth()


def pytest_collection_modifyitems(items) -> None:
    for item in items:
        item.name = item.name.encode("utf-8").decode("unicode-escape")
        item._nodeid = item.nodeid.encode("utf-8").decode("unicode-escape")


def pytest_configure(config):
    """运行时参数优先级：pytest 参数 > 环境变量 > 默认值。"""
    global _case_file, _module_name, _case_list_cache
    _case_list_cache = None
    opt_case_file = (config.getoption("--case-file") or "").strip()
    if opt_case_file:
        _case_file = opt_case_file
    else:
        _case_file = at_env.default_case_file(config).strip()
    _module_name = os.path.basename(os.path.normpath(os.path.abspath(_case_file))) if _case_file else ""

    opt_scope = config.getoption("--scope")
    opt_tags = config.getoption("--tags")
    opt_api_name = config.getoption("--api-name")
    opt_api_path = config.getoption("--api-path")
    opt_case_names = config.getoption("--case-names")
    opt_suite = config.getoption("--suite")

    if opt_scope is not None:
        os.environ["SCOPE"] = opt_scope.strip()
    if opt_tags is not None:
        os.environ["TAGS"] = opt_tags.strip()
    if opt_api_name is not None:
        os.environ["API_NAME"] = opt_api_name.strip()
    if opt_api_path is not None:
        os.environ["API_PATH"] = opt_api_path.strip()
    if opt_case_names is not None:
        os.environ["CASE_NAMES"] = opt_case_names.strip()
    if opt_suite is not None:
        os.environ["SUITE"] = opt_suite.strip()


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
    if not at_env.clean_up_enabled(config):
        return
    clean_up_module = at_env.clean_up_module_name(config)
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
