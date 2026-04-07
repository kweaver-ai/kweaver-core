# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import math

from lib.mcp import MCP
from lib.tool_box import ToolBox
from common.get_content import GetContent
from common.assert_tools import AssertTools

@allure.feature("MCP服务管理接口测试：获取MCP_Server列表")
class TestGetMCPList:   
    client = MCP()
    tool_client = ToolBox()
    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        for i in range(15):
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            data = {
                "name": name,
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

        for i in range(5):
            filepath = "./resource/openapi/compliant/test.json"
            json_data = GetContent(filepath).jsonfile()
            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            data = {
                "box_name": name,
                "data": json_data,
                "metadata_type": "openapi"
            }
            result = self.tool_client.CreateToolbox(data, Headers)
            box_id = result[1]["box_id"]
            result = self.tool_client.GetBoxToolsList(box_id, None, Headers)
            tools = result[1]["tools"]
            update_data = []
            tools_id = []
            tool_configs = []
            for tool in tools:
                data = {
                    "tool_id": tool["tool_id"],
                    "status": "enabled"
                }
                update_data.append(data) 
                tools_id.append(tool["tool_id"])  
                tool_config = {
                    "box_id": box_id,
                    "box_name": name,
                    "tool_id": tool["tool_id"],
                    "tool_name": tool["name"],
                    "description": tool["description"],
                    "use_rule": "all"
                }
                tool_configs.append(tool_config)       
            result = self.tool_client.UpdateToolStatus(box_id, update_data, Headers)
            assert result[0] == 200
            result = self.tool_client.UpdateToolboxStatus(box_id, {"status": "published"}, Headers)
            assert result[0] == 200

            name = ''.join(random.choice(string.ascii_letters) for i in range(8))
            data = {
                "name": name,
                "description": "add mcp server config",
                "mode": "sse",
                "url": "http://localhost:8080/api/v1/tools",
                "category": "data_analysis",
                "creation_type": "tool_imported",
                "tool_configs": tool_configs
            }
            result = self.client.RegisterMCP(data, Headers)
            assert result[0] == 200

    @allure.title("获取MCP_Server列表，默认从第一页开始，每页10个，按照创建时间倒序排列")
    def test_get_mcp_list_01(self, Headers):
        result = self.client.GetMCPList(None, Headers)
        assert result[0] == 200
        mcp_list = result[1]["data"]
        assert result[1]["page"] == 1
        assert result[1]["page_size"] == 10
        assert result[1]["total_pages"] == math.ceil(result[1]["total"]/result[1]["page_size"])
        assert len(mcp_list) == 10
        if result[1]["total_pages"] > result[1]["page"]:
            assert result[1]["has_next"] == True
        else:
            assert result[1]["has_next"] == False
        if result[1]["page"] == 1:
            assert result[1]["has_prev"] == False
        elif result[1]["total_pages"] > 1:
            assert result[1]["has_prev"] == True

        create_times = []
        for mcp in mcp_list:
            if mcp["creation_type"] == "custom":
                assert "tool_configs" not in mcp
            elif mcp["creation_type"] == "tool_imported":
                assert len(mcp["tool_configs"]) > 0

            create_times.append(mcp["create_time"])
        
        assert AssertTools.is_descending_str(create_times) == True

    @allure.title("获取MCP_Server列表，page小于0，获取失败")
    def test_get_mcp_list_02(self, Headers):
        params = {
            "page": -1
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 400

    @allure.title("获取MCP_Server列表，page和page_size为0，获取成功，采用默认值")
    def test_get_mcp_list_03(self, Headers):
        params = {
            "page": 0,
            "page_size": 0
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 200
        assert result[1]["page"] == 1
        assert result[1]["page_size"] == 10

    @allure.title("获取MCP_Server列表，page_size不在[1-100]内，获取失败")
    @pytest.mark.parametrize("page_size", [-1, -2, 101])
    def test_get_mcp_list_04(self, Headers, page_size):
        params = {
            "page_size": page_size
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 400

    @allure.title("page_size在[1-100]内，获取MCP_Server列表成功")
    @pytest.mark.parametrize("page_size", [1, 20, 50, 100])
    def test_get_mcp_list_05(self, Headers, page_size):
        params = {
            "page_size": page_size
        }
        result = self.client.GetMCPList(params, Headers)
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

    @allure.title("获取MCP_Server列表，按status搜索，获取成功")
    def test_get_mcp_list_06(self, Headers):
        params = {
            "status": "unpublish"
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 200
        for mcp in result[1]["data"]:
            assert mcp["status"] == "unpublish"

    @allure.title("获取MCP_Server列表，按mode搜索，获取成功")
    def test_get_mcp_list_07(self, Headers):
        params = {
            "mode": "sse"
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 200

    @allure.title("获取MCP_Server列表，按name搜索，获取成功")
    def test_get_mcp_list_08(self, Headers):
        # 先获取一个MCP Server的name
        result = self.client.GetMCPList(None, Headers)
        assert result[0] == 200
        name = result[1]["data"][0]["name"]

        # 按name搜索
        params = {
            "name": name
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) > 0
        for mcp in result[1]["data"]:
            assert name in mcp["name"]

    @allure.title("获取MCP_Server列表，按category搜索，获取成功")
    def test_get_mcp_list_09(self, Headers):
        # 先获取一个MCP Server的category
        result = self.client.GetMCPList(None, Headers)
        assert result[0] == 200
        category = result[1]["data"][0]["category"]

        # 按category搜索
        params = {
            "category": category
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) > 0
        for mcp in result[1]["data"]:
            assert mcp["category"] == category 

    @allure.title("获取MCP_Server列表，排序字段不正确，获取失败")
    def test_get_mcp_list_10(self, Headers):
        params = {
            "sort_by": "invalid_field"
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 400

    @allure.title("获取MCP_Server列表，排序顺序不正确，获取失败")
    def test_get_mcp_list_11(self, Headers):
        params = {
            "sort_order": "invalid_order"
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 400

    @allure.title("获取MCP_Server列表，all为True获取所有MCP，获取成功")
    def test_get_mcp_list_12(self, Headers):
        params = {
            "all": True
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) == result[1]["total"]

    @allure.title("获取MCP_Server列表，按照名称正序排列，获取成功")
    def test_get_mcp_list_13(self, Headers):
        params = {
            "sort_by": "name",
            "sort_order": "asc",
            "all": True
        }
        result = self.client.GetMCPList(params, Headers)
        assert result[0] == 200
        mcp_list = result[1]["data"]
        names = []
        for mcp in mcp_list:
            names.append(mcp["name"])

        assert AssertTools.is_ascending_str(names)