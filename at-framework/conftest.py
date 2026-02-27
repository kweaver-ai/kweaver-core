import os
import allure
import pytest
import requests

from common.func import load_sys_config, load_case, get_cases

config = load_sys_config("./config/config.ini")
_case_file = config["env"]["case_file"]
# 当前模块名：case_file 路径最后一段（如 testcase/vega -> vega），用于按模块执行 clean_up 等
_module_name = os.path.basename(os.path.normpath(os.path.abspath(_case_file))) if _case_file else ""
_scope = os.environ.get("SCOPE", "").strip()
_tags = os.environ.get("TAGS", "").strip()
_api_name = os.environ.get("API_NAME", "").strip()
_api_path = os.environ.get("API_PATH", "").strip()
_case_names = os.environ.get("CASE_NAMES", "").strip()  # 智能体根据提交内容选出的用例 name 列表，逗号分隔
if _case_names:
    names_list = [x.strip() for x in _case_names.split(",") if x.strip()]
    case_list = get_cases(_case_file, names=names_list) if names_list else load_case(_case_file)
elif _scope or _tags or _api_name or _api_path:
    case_list = get_cases(_case_file, scope=_scope or None, tags=_tags or None,
                          api_name=_api_name or None, api_path=_api_path or None)
else:
    case_list = load_case(_case_file)
BEARER_AUTH = "Bearer %s" % config["external"]["token"]


def pytest_collection_modifyitems(items) -> None:
    # item表示每个测试用例，解决用例名称中文显示问题
    for item in items:
        item.name = item.name.encode("utf-8").decode("unicode-escape")
        item._nodeid = item.nodeid.encode("utf-8").decode("unicode-escape")


@pytest.fixture(scope="session", autouse=True)
@allure.step("清理历史数据")
def clean_up():
    if int(config["env"].get("clean_up", "0")) != 1:
        return
    clean_up_module = config["env"].get("clean_up_module", "").strip()
    if clean_up_module and clean_up_module != _module_name:
        return  # 仅当配置的模块与当前运行模块一致时执行（框架通用，不同模块可有不同清理逻辑）

    print("清理历史数据")

    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json",
        "Authorization": "Bearer %s" % config["external"]["token"]
    }

    response = requests.get("https://%s/api/data-connection/v1/datasource" % config["env"]["host"],
                            verify=False, headers=headers)
    if response.status_code > 200:
        print("查询数据源列表失败")
        allure.attach(
            body="query catalog failed: %s" % response.json(), name="清理历史数据失败"
        )
        return

    for x in response.json()["entries"]:
        if not x["is_built_in"]:
            response = requests.get("https://%s/api/mdl-data-model/v1/data-views" % config["env"]["host"],
                                    params={"data_source_id": x["id"], "type": "atomic"}, verify=False, headers=headers)
            if response.status_code > 200:
                print("查询视图失败 @ %s" % x["name"])
                allure.attach(
                    body="query view failed: %s" % response.json(), name="清理历史数据失败"
                )
                return

            for y in response.json()["entries"]:
                response = requests.delete("https://%s/api/mdl-data-model/v1/data-views/%s" % (
                    config["env"]["host"], y["id"]), verify=False, headers=headers)
                if response.status_code > 300:
                    print("删除视图失败 @ %s" % x["name"])
                    allure.attach(
                        body="delete view %s failed: %s" % (x["name"], response.json()), name="清理历史数据失败"
                    )

            response = requests.delete("http://%s/api/data-connection/v1/datasource/%s" % (
                config["env"]["host"], x["id"]), verify=False, headers=headers)
            if response.status_code > 200:
                print("删除数据源失败 @ %s" % x["name"])
                allure.attach(
                    body="delete catalog %s failed: %s" % (x["name"], response.json()), name="清理历史数据失败"
                )
