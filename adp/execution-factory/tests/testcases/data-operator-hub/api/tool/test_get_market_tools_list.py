# -*- coding:UTF-8 -*-

import allure
import pytest
import math
import string
import random
import time

from lib.tool_box import ToolBox
from common.assert_tools import AssertTools
from common.get_content import GetContent

box_id = ""

@allure.feature("工具注册与管理接口测试：查询市场工具列表")
class TestGetMarketToolsList:
    '''为优化性能，tool_name设置为必填字段'''
    
    client = ToolBox()
    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        # 创建工具箱并发布，添加重试机制处理503错误
        box_ids = []
        max_retries = 3
        
        for i in range(15):
            filepath = "./resource/openapi/compliant/toolbox.yaml"
            yaml_data = GetContent(filepath).yamlfile()
            name = ''.join(random.choice(string.ascii_letters) for i in range(8)) 
            data = {
                "box_name": name,
                "data": yaml_data,
                "metadata_type": "openapi"
            }
            
            # 重试创建工具箱
            result = None
            for attempt in range(max_retries):
                result = self.client.CreateToolbox(data, Headers)
                if result[0] == 200:
                    break
                elif result[0] == 503 and attempt < max_retries - 1:
                    # 503错误时等待后重试
                    wait_time = min(2 ** attempt, 5)  # 最多等待5秒
                    time.sleep(wait_time)
                    continue
                else:
                    # 其他错误或最后一次重试失败
                    break
            
            # 如果创建失败，根据错误类型处理
            if result[0] == 503:
                # 503服务不可用，跳过setup
                pytest.skip(f"后端服务不可用(503)，无法创建工具箱进行测试。响应: {result[1]}")
            elif result[0] != 200:
                # 其他错误，断言失败
                assert False, f"创建工具箱失败，状态码: {result[0]}, 响应: {result[1]}"
            
            # 确保result[1]是字典类型
            assert isinstance(result[1], dict), f"响应格式错误，期望字典，实际: {type(result[1])}, 内容: {result[1]}"
            box_id = result[1]["box_id"]
            box_ids.append(box_id)
        
        # 发布工具箱
        for box_id in box_ids:
            update_data = {
                "status": "published"
            }
            # 发布时也添加重试机制
            result = None
            for attempt in range(max_retries):
                result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
                if result[0] == 200:
                    break
                elif result[0] == 503 and attempt < max_retries - 1:
                    wait_time = min(2 ** attempt, 5)
                    time.sleep(wait_time)
                    continue
                else:
                    break
            
            if result[0] == 503:
                pytest.skip(f"后端服务不可用(503)，无法发布工具箱。响应: {result[1]}")
            elif result[0] != 200:
                assert False, f"发布工具箱失败，状态码: {result[0]}, 响应: {result[1]}"

    @allure.title("获取市场工具列表，默认从第一页开始，每页10个，按照更新时间倒序排列")
    def test_get_market_tools_list_01(self, Headers):
        result = self.client.GetMarketToolsList({"tool_name": "工具"}, Headers)
        assert result[0] == 200
        toolbox_list = result[1]["data"]
        assert result[1]["page"] == 1
        assert result[1]["page_size"] == 10
        assert result[1]["total_pages"] == math.ceil(result[1]["total"]/result[1]["page_size"])
        assert len(toolbox_list) == 10
        if result[1]["total_pages"] > result[1]["page"]:
            assert result[1]["has_next"] == True
        else:
            assert result[1]["has_next"] == False
        if result[1]["page"] == 1:
            assert result[1]["has_prev"] == False
        elif result[1]["total_pages"] > 1:
            assert result[1]["has_prev"] == True

        box_update_times = []
        tool_update_times = []
        for toolbox in toolbox_list:
            assert toolbox["status"] == "published"
            assert "box_id" in toolbox
            assert "box_name" in toolbox
            assert "box_desc" in toolbox
            box_update_times.append(toolbox["update_time"]) 
            for tool in toolbox["tools"]:
                assert "tool_id" in tool
                assert "name" in tool
                assert "description" in tool
                assert "status" in tool
                tool_update_times.append(tool["update_time"]) 
            assert AssertTools.is_descending_str(tool_update_times) == True
        assert AssertTools.is_descending_str(box_update_times) == True

    @allure.title("获取市场工具列表，page小于0，获取失败")
    def test_get_market_tools_list_02(self, Headers):
        params = {
            "page": -1,
            "tool_name": "工具"
        }
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 400, f"期望返回400，实际返回{result[0]}，响应: {result[1]}"

    @allure.title("获取市场工具列表，page和page_size为0，获取成功，采用默认值")
    def test_get_market_tools_list_03(self, Headers):
        params = {
            "page": 0,
            "page_size": 0,
            "tool_name": "工具"
        }
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 200
        assert result[1]["page"] == 1
        assert result[1]["page_size"] == 10

    @allure.title("获取市场工具列表，page_size不在[1-100]内，获取失败")
    @pytest.mark.parametrize("page_size", [-1, -2, 101])
    def test_get_market_tools_list_04(self, Headers, page_size):
        params = {
            "page_size": page_size,
            "tool_name": "工具"
        }
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 400, f"page_size={page_size}时，期望返回400，实际返回{result[0]}，响应: {result[1]}"

    @allure.title("page_size在[1-100]内，获取市场工具列表成功")
    @pytest.mark.parametrize("page_size", [1, 20, 50, 100])
    def test_get_market_tools_list_05(self, Headers, page_size):
        params = {
            "page_size": page_size,
            "tool_name": "工具"
        }
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 200, f"page_size={page_size}时，期望返回200，实际返回{result[0]}，响应: {result[1]}"
        assert isinstance(result[1], dict), f"响应格式错误，期望字典，实际: {type(result[1])}"
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

    @allure.title("获取市场工具列表，排序字段不正确，获取失败")
    def test_get_market_tools_list_06(self, Headers):
        params = {
            "sort_by": "invalid_field",
            "tool_name": "工具"
        }
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 400, f"期望返回400，实际返回{result[0]}，响应: {result[1]}"

    @allure.title("获取市场工具列表，排序顺序不正确，获取失败")
    def test_get_market_tools_list_07(self, Headers):
        params = {
            "sort_order": "invalid_order",
            "tool_name": "工具"
        }
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 400

    @allure.title("获取市场工具列表，all为True获取所有工具箱及工具，获取成功")
    def test_get_market_tools_list_08(self, Headers):
        params = {
            "all": True,
            "tool_name": "工具"
        }
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) == result[1]["total"]

    @allure.title("获取市场工具列表，根据工具名称查询，获取成功")
    def test_get_market_tools_list_09(self, Headers):
        params = {
            "tool_name": "获取"
        }
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 200
        for box in result[1]["data"]:
            for tool in box["tools"]:
                assert "获取" in tool["name"]

        params = {
            "tool_name": "test"
        }
        result = self.client.GetMarketToolsList(params, Headers)
        assert result[0] == 200
        assert result[1]["data"] == []

       