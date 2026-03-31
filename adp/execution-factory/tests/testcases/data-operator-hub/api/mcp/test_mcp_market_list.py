# -*- coding:UTF-8 -*-

import allure
import pytest
import math
import string
import random

from lib.mcp import MCP
from common.assert_tools import AssertTools


@allure.feature("MCP服务市场接口测试：获取MCP服务市场列表")
class TestMarketList:
    
    client = MCP()
    mcp_id = ""

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        # 注册并发布mcp
        for i in range(30):
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            data = {
                "name": "test_" + name,
                "description": "test mcp server",
                "mode": "sse",
                "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
                "headers": {
                    "Content-Type": "application/json"
                },
                "category": "other_category"
            }
            result = self.client.RegisterMCP(data, Headers)
            assert result[0] == 200
            self.mcp_id = result[1]["mcp_id"]
            data = {
                "status": "published"
            }
            result = self.client.MCPReleaseAction(self.mcp_id, data, Headers)
            assert result[0] == 200

    @allure.title("获取已发布的MCP服务列表，默认参数，获取成功， 仅返回mcp最新已发布版本，默认每页10个，按照更新时间倒序排列")
    def test_market_list_01(self, Headers):
        result = self.client.GetMCPMarketList(None, Headers)
        assert result[0] == 200
        assert result[1]["page"] == 1
        assert result[1]["page_size"] == 10
        mcp_list = result[1]["data"]
        assert len(mcp_list) <= 10
        assert result[1]["total_pages"] == math.ceil(result[1]["total"]/result[1]["page_size"])
        if result[1]["total_pages"] > result[1]["page"]:
            assert result[1]["has_next"] == True
        else:
            assert result[1]["has_next"] == False
        if result[1]["page"] == 1:
            assert result[1]["has_prev"] == False
        elif result[1]["total_pages"] > 1:
            assert result[1]["has_prev"] == True

        mcp_ids = []
        update_times = []
        for mcp in mcp_list:
            mcp_ids.append(mcp["mcp_id"])
            update_times.append(mcp["update_time"])
            assert mcp["status"] == "published"

        assert AssertTools.has_duplicates(mcp_ids) == False
        assert AssertTools.is_descending_str(update_times) == True

    @allure.title("获取MCP服务市场列表，分页参数为0，获取成功，采用默认值")
    def test_market_list_02(self, Headers):
        params = {
            "page": 0,
            "page_size": 0
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 200
        assert result[1]["page"] == 1
        assert result[1]["page_size"] == 10
        assert len(result[1]["data"]) <= 10

    @allure.title("获取MCP服务市场列表，page参数小于0，获取失败")
    def test_market_list_03(self, Headers):
        params = {
            "page": -1
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 400

    @allure.title("获取MCP服务市场列表，page_size参数超出范围，获取失败")
    @pytest.mark.parametrize("page_size", [-1, -2, 101])
    def test_market_list_04(self, page_size, Headers):
        params = {
            "page_size": page_size
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 400

    @allure.title("获取MCP服务市场列表，page_size在[1-100]范围内，获取成功")
    @pytest.mark.parametrize("page_size", [1, 20, 50, 100])
    def test_market_list_05(self, page_size, Headers):
        params = {
            "page_size": page_size
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 200
        assert result[1]["page_size"] == page_size
        assert len(result[1]["data"]) <= page_size
        if result[1]["total_pages"] > result[1]["page"]:
            assert result[1]["has_next"] == True
        else:
            assert result[1]["has_next"] == False
        if result[1]["page"] == 1:
            assert result[1]["has_prev"] == False
        elif result[1]["total_pages"] > 1:
            assert result[1]["has_prev"] == True

        for mcp in result[1]["data"]:
            assert mcp["status"] == "published"

    @allure.title("获取MCP服务市场列表，mode参数不正确，获取失败")
    def test_market_list_06(self, Headers):
        params = {
            "mode": "invalid_mode"
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 400

    @allure.title("获取MCP服务市场列表，根据name获取，获取成功")
    def test_market_list_07(self, Headers):
        params = {
            "name": "test"
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) > 0
        for mcp in result[1]["data"]:
            assert "test" in mcp["name"]

    @allure.title("获取MCP服务市场列表，根据category获取，获取成功")
    def test_market_list_08(self, Headers):
        params = {
            "category": "other_category"
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 200
        for item in result[1]["data"]:
            assert item["category"] == "other_category" 

    @allure.title("获取MCP服务市场列表，排序字段不正确，获取失败")
    def test_market_list_09(self, Headers):
        params = {
            "sort_by": "invalid_field"
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 400

    @allure.title("获取MCP服务市场列表，排序顺序不正确，获取失败")
    def test_market_list_10(self, Headers):
        params = {
            "sort_order": "invalid_order"
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 400

    @allure.title("获取MCP服务市场列表，all为True获取所有工具箱，获取成功")
    def test_market_list_11(self, Headers):
        params = {
            "all": True
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) == result[1]["total"]
        for mcp in result[1]["data"]:
            assert mcp["status"] == "published"

    @allure.title("获取MCP服务市场列表，按照名称正序排列，获取成功")
    def test_market_list_12(self, Headers):
        params = {
            "sort_by": "name",
            "sort_order": "asc",
            "all": True
        }
        result = self.client.GetMCPMarketList(params, Headers)
        assert result[0] == 200
        mcp_list = result[1]["data"]
        names = []
        for mcp in mcp_list:
            names.append(mcp["name"])

        assert AssertTools.is_ascending_str(names)