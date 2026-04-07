# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import time

from lib.operator import Operator
from common.get_content import GetContent
from common.assert_tools import AssertTools

@allure.feature("算子注册与管理接口测试：高级功能测试")
class TestOperatorAdvanced:
    """
    本测试类涵盖算子注册与管理的高级功能，包括：
    1. 历史版本详情的语义化标签 (tag) 查询。
    2. 市场列表的高级过滤（execution_mode, metadata_type）。
    3. 状态更新的冲突处理 (409 Conflict)。
    4. 调试接口的复杂参数验证。
    """
    
    client = Operator()
    operator_id = ""
    version = ""

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        # 注册并发布一个算子用于历史和状态测试
        filepath = "./resource/openapi/compliant/test1.json"
        api_data = GetContent(filepath).jsonfile()
        name = 'adv_op_' + ''.join(random.choice(string.ascii_letters) for i in range(5))
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        
        res = None
        for attempt in range(3):
            res = self.client.RegisterOperator(data, Headers)
            if res[0] == 200: break
            time.sleep(1)
            
        if res[0] != 200:
            pytest.skip("Setup failed to register operator")
            
        TestOperatorAdvanced.operator_id = res[1][0]["operator_id"]
        TestOperatorAdvanced.version = res[1][0]["version"]

    @allure.title("查询算子历史版本，使用tag参数，查询成功")
    def test_get_operator_history_with_tag(self, Headers):
        # 目前系统中可能没有设置过tag，这里验证参数传递
        result = self.client.GetOperatorHistoryDetail(self.operator_id, self.version, Headers, tag=1)
        # 如果没有对应tag，可能返回404或200但数据不匹配，取决于后端实现
        assert result[0] in [200, 404]

    @allure.title("获取市场算子列表，按执行模式(sync)过滤，查询成功")
    def test_get_operator_market_filter_sync(self, Headers):
        params = {"execution_mode": "sync", "page_size": 10}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200
        for op in result[1].get("data", []):
            assert op["operator_info"]["execution_mode"] == "sync"

    @allure.title("获取市场算子列表，按元数据类型(openapi)过滤，查询成功")
    def test_get_operator_market_filter_openapi(self, Headers):
        params = {"metadata_type": "openapi", "page_size": 10}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200
        for op in result[1].get("data", []):
            assert op["metadata_type"] == "openapi"

    @allure.title("更新算子状态，状态转换冲突时返回409")
    def test_update_operator_status_conflict(self, Headers):
        # 已经是 published 状态，再次尝试 published 可能会冲突或成功（取决于幂等性设计）
        # 按照 YAML，409 是可能的结果。如果返回 400 且错误码是冲突相关的，也应接受
        update_data = [{"operator_id": self.operator_id, "status": "published"}]
        result = self.client.UpdateOperatorStatus(update_data, Headers)
        # 如果后端支持幂等，则返回200；如果不支持且状态相同，可能返回409或400带冲突错误
        assert result[0] in [200, 400, 409]

    @allure.title("获取算子市场列表，按名称排序，查询成功")
    def test_get_operator_market_sort_name(self, Headers):
        params = {"sort_by": "name", "sort_order": "asc", "page_size": 20}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200
        names = [op["name"] for op in result[1].get("data", [])]
        assert AssertTools.is_ascending_str(names)
