# -*- coding:UTF-8 -*-

import allure
import string
import random
import math
import pytest

from lib.tool_box import ToolBox
from common.get_content import GetContent
from common.assert_tools import AssertTools


@allure.feature("工具注册与管理接口测试：获取工具箱列表")
class TestGetToolboxList:
    
    client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        for i in range(10):
            filepath = "./resource/openapi/compliant/toolbox.yaml"
            yaml_data = GetContent(filepath).yamlfile()
            name = "test" + ''.join(random.choice(string.ascii_letters) for i in range(8))
            
            data = {
                "box_name": name,
                "data": yaml_data,
                "metadata_type": "openapi"
            }
            result = self.client.CreateToolbox(data, Headers)
            assert result[0] == 200

    @allure.title("获取工具箱列表，默认从第一页开始，每页10个，按照创建时间倒序排列")
    def test_get_toolbox_list_01(self, Headers):
        params = {}
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 200
        toolboxlist = result[1]
        assert toolboxlist["total"] >= 10
        assert toolboxlist["page"] == 1
        assert toolboxlist["page_size"] == 10
        assert toolboxlist["total_pages"] == math.ceil(toolboxlist["total"]/toolboxlist["page_size"])
        assert len(toolboxlist["data"]) == 10
        if toolboxlist["total_pages"] > toolboxlist["page"]:
            assert toolboxlist["has_next"] == True
        else:
            assert toolboxlist["has_next"] == False
        if toolboxlist["page"] == 1:
            assert toolboxlist["has_prev"] == False
        elif toolboxlist["total_pages"] > 1:
            assert toolboxlist["has_prev"] == True

        create_times = []
        for toolbox in toolboxlist["data"]:
            create_times.append(toolbox["create_time"])
        assert AssertTools.is_descending_str(create_times) == True

    @allure.title("page小于0，获取工具箱列表失败")
    def test_get_toolbox_list_02(self, Headers):
        params = {
            "page": -1
        }
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 400

    @allure.title("page和page_size为0，获取工具箱列表，获取成功，采用默认值")
    def test_get_toolbox_list_03(self, Headers):
        params = {
            "page": 0,
            "page_size": 0
        }
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 200
        assert result[1]["page"] == 1
        assert result[1]["page_size"] == 10

    @allure.title("page_size不在[1-100]内，获取工具箱列表失败")
    @pytest.mark.parametrize("page_size", [-1, -2, 101])
    def test_get_toolbox_list_04(self, page_size, Headers):
        params = {
            "page_size": page_size
        }
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 400

    @allure.title("page_size在[1-100]内，获取工具箱列表成功")
    @pytest.mark.parametrize("page_size", [1, 20, 50, 100])
    def test_get_toolbox_list_05(self, page_size, Headers):
        params = {
            "page_size": page_size
        }
        result = self.client.GetToolboxList(params, Headers)
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

    @allure.title("获取工具箱列表，排序字段不正确，获取失败")
    def test_get_toolbox_list_06(self, Headers):
        params = {
            "page": 1,
            "page_size": 10,
            "sort_by": "invalid_field",
            "sort_order": "desc"
        }
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 400

    @allure.title("获取工具箱列表，排序顺序不正确，获取失败")
    def test_get_toolbox_list_07(self, Headers):
        params = {
            "page": 1,
            "page_size": 10,
            "sort_by": "create_time",
            "sort_order": "invalid_order"
        }
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 400

    @allure.title("获取工具箱列表，all为True获取所有工具箱，获取成功")
    def test_get_toolbox_list_08(self, Headers):
        params = {
            "all": "true"
        }
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 200
        assert "data" in result[1]
        assert "total" in result[1]
        assert len(result[1]["data"]) == result[1]["total"]

    @allure.title("获取工具箱列表，根据名称正序排列，获取成功")
    def test_get_toolbox_list_09(self, Headers):
        params = {
            "page": 1,
            "page_size": 10,
            "sort_by": "name",
            "sort_order": "asc"
        }
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 200
        
        box_names = []
        for toolbox in result[1]["data"]:
            box_names.append(toolbox["box_name"])
        
        assert AssertTools.is_ascending_str(box_names) == True

    @allure.title("获取工具箱列表，根据类型获取工具箱，获取成功")
    def test_get_toolbox_list_10(self, Headers):
        params = {
            "category": "data_process"
        }
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 200
        for box in result[1]["data"]:
            assert box["category_type"] == "data_process"

    @allure.title("获取工具箱列表，根据状态获取工具箱，获取成功")
    def test_get_toolbox_list_11(self, Headers):
        params = {
            "status": "unpublish"
        }
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 200
        for box in result[1]["data"]:
            assert box["status"] == "unpublish"

    @allure.title("获取工具箱列表，根据名称获取工具箱，获取成功")
    def test_get_toolbox_list_12(self, Headers):
        params = {
            "name": "test"
        }
        result = self.client.GetToolboxList(params, Headers)
        assert result[0] == 200
        for box in result[1]["data"]:
            assert "test" in box["box_name"] 