# -*- coding:UTF-8 -*-

import allure
import string
import random
import math
import pytest

from lib.tool_box import ToolBox
from lib.operator import Operator
from common.get_content import GetContent
from common.assert_tools import AssertTools

box_id = ""

@allure.feature("工具注册与管理接口测试：获取工具箱中的工具列表")
class TestGetBoxToolsList:
    
    client = ToolBox()
    client1 = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id

        # 创建工具箱
        filepath = "./resource/openapi/compliant/toolbox.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]

    @allure.title("获取工具箱中的工具列表，默认从第一页开始，每页10个，按照创建时间倒序排列")
    def test_get_box_tools_list_01(self, Headers):
        global box_id

        result = self.client.GetBoxToolsList(box_id, None, Headers)
        assert result[0] == 200
        tool_list = result[1]["tools"]

        assert result[1]["total"] == 15
        assert result[1]["page"] == 1
        assert result[1]["page_size"] == 10
        assert result[1]["total_pages"] == math.ceil(result[1]["total"]/result[1]["page_size"])
        assert len(tool_list) == 10
        if result[1]["total_pages"] > result[1]["page"]:
            assert result[1]["has_next"] == True
        else:
            assert result[1]["has_next"] == False
        if result[1]["page"] == 1:
            assert result[1]["has_prev"] == False
        elif result[1]["total_pages"] > 1:
            assert result[1]["has_prev"] == True

        create_times = []
        for tool in tool_list:
            create_times.append(tool["create_time"])
        
        assert AssertTools.is_descending_str(create_times) == True

    @allure.title("获取工具箱中的工具列表，page小于0，获取失败")
    def test_get_box_tools_list_02(self, Headers):
        global box_id

        params = {
            "page": -1
        }

        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 400

    @allure.title("page和page_size为0，获取工具箱中的工具列表，获取成功，采用默认值")
    def test_get_box_tools_list_03(self, Headers):
        global box_id

        params = {
            "page": 0,
            "page_size": 0
        }

        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 200
        assert result[1]["page"] == 1
        assert result[1]["page_size"] == 10

    @allure.title("page_size不在[1-100]内，获取工具箱中的工具列表失败")
    @pytest.mark.parametrize("page_size", [-1, -2, 101])
    def test_get_box_tools_list_04(self, Headers, page_size):
        global box_id

        params = {
            "page_size": page_size
        }

        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 400

    @allure.title("page_size在[1-100]内，获取工具箱中的工具列表成功")
    @pytest.mark.parametrize("page_size", [1, 20, 50, 100])
    def test_get_box_tools_list_05(self, Headers, page_size):
        global box_id

        params = {
            "page_size": page_size
        }

        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 200
        assert result[1]["page_size"] == page_size
        assert len(result[1]["tools"]) <= page_size
        if result[1]["total_pages"] > result[1]["page"]:
            assert result[1]["has_next"] == True
        else:
            assert result[1]["has_next"] == False
        if result[1]["page"] == 1:
            assert result[1]["has_prev"] == False
        elif result[1]["total_pages"] > 1:
            assert result[1]["has_prev"] == True

    @allure.title("获取工具箱中的工具列表，排序字段不正确，获取失败")
    def test_get_box_tools_list_06(self, Headers):
        global box_id

        params = {
            "page": 1,
            "page_size": 10,
            "sort_by": "invalid_field",
            "sort_order": "desc"
        }

        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 400

    @allure.title("获取工具箱中的工具列表，排序顺序不正确，获取失败")
    def test_get_box_tools_list_07(self, Headers):
        global box_id

        params = {
            "page": 1,
            "page_size": 10,
            "sort_by": "create_time",
            "sort_order": "invalid_order"
        }

        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 400

    @allure.title("获取工具箱中的工具列表，all为True获取所有工具箱，获取成功")
    def test_get_box_tools_list_08(self, Headers):
        global box_id

        params = {
            "all": "true"
        }

        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 200
        assert "tools" in result[1]
        assert "total" in result[1]
        assert len(result[1]["tools"]) == result[1]["total"]

    @allure.title("获取工具箱中的工具列表，根据名称正序排列，获取成功")
    def test_get_box_tools_list_09(self, Headers):
        global box_id

        params = {
            "page": 1,
            "page_size": 10,
            "sort_by": "tool_name",
            "sort_order": "asc"
        }

        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 200
        
        names = []
        for tool in result[1]["tools"]:
            names.append(tool["name"])
        
        assert AssertTools.is_ascending_str(names) == True

    @allure.title("获取工具箱中的工具列表，工具箱内包含有转化为工具的算子，获取成功")
    def test_get_box_tools_list_10(self, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()

        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }

        result = self.client1.RegisterOperator(data, Headers)
        assert result[0] == 200
        operator = result[1][0]
        assert operator["status"] == "success"
        operator_id = operator["operator_id"]
        operator_version = operator["version"]

        # 转换算子为工具
        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 200
        tool_id = result[1]["tool_id"]

        params = {
            "all": "true"
        }

        # 获取工具列表
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 200
        assert len(result[1]["tools"]) == result[1]["total"]
        tool_ids = []
        for tool in result[1]["tools"]:
            tool_ids.append(tool["tool_id"])
        assert tool_id in tool_ids

        # 下架算子后再次获取工具列表
        update_data = [
            {
                "operator_id": operator_id,
                "version": operator_version,
                "status": "offline"
            }
        ] 
        result = self.client1.UpdateOperatorStatus(update_data, Headers)
        assert result[0] == 200

        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 200
        assert len(result[1]["tools"]) == result[1]["total"]
        tool_ids = []
        for tool in result[1]["tools"]:
            tool_ids.append(tool["tool_id"])
        assert tool_id in tool_ids
