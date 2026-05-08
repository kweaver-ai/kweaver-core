"""Framework Hooks: dataflow"""
from typing import Dict

import requests
import urllib3

from common import at_env

urllib3.disable_warnings(urllib3.connectionpool.InsecureRequestWarning)


def session_setup(session_id, config):
    """会话开始前的设置"""
    pass


def session_clean_up(config: Dict[str, Dict[str, str]], allure) -> None:
    """
    会话结束后清理本次测试创建的数据
    """

    allure.attach("开始清理 vega-backend 历史数据", name="session_clean_up")

    scheme, host = at_env.resolve_request_target(config)
    headers = {"Content-Type": "application/json", "Accept": "application/json",
               "Authorization": "Bearer %s" % at_env.static_access_token(config)}

    query_catalog_url = "%s://%s/api/vega-backend/v1/catalogs" % (scheme, host)
    response = requests.get(
        query_catalog_url, params={"limit": -1, "offset": 0}, headers=headers, verify=False
    )
    if response.status_code != 200:
        allure.attach("清理 vega-backend 历史数据失败：%s" % response.content, name="session_clean_up")
        return

    catalog_ids = [x["id"] for x in response.json().get("entries", []) if "adp_bkn" not in x["name"]]
    delete_catalog_url = "%s://%s/api/vega-backend/v1/catalogs/%s" % (scheme, host, ','.join(catalog_ids))
    response = requests.delete(delete_catalog_url, headers=headers, verify=False)
    if response.status_code != 204:
        allure.attach("清理 vega-backend 历史数据失败：%s" % response.content, name="session_clean_up")
        return

    allure.attach("vega-backend 历史数据清理完成", name="session_clean_up")


def test_setup(test_id, config):
    pass


def test_teardown(test_id, config):
    pass
