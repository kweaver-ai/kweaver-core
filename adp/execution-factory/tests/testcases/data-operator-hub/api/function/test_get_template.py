# -*- coding:UTF-8 -*-

import allure
import pytest

from lib.tool_box import ToolBox


@allure.feature("函数相关接口测试：获取代码模板")
class TestGetTemplate:
    
    client = ToolBox()

    @allure.title("获取代码模板，template_type为python，获取成功")
    def test_get_template_01(self, Headers):
        result = self.client.GetTemplate("python", Headers)
        assert result[0] == 200
        assert "template_type" in result[1]
        assert "code_template" in result[1]
        assert result[1]["template_type"] == "python"
        assert len(result[1]["code_template"]) > 0

    @allure.title("获取代码模板，template_type不存在，获取失败")
    def test_get_template_02(self, Headers):
        result = self.client.GetTemplate("invalid_type", Headers)
        assert result[0] == 400 or result[0] == 404

    @allure.title("获取代码模板，template_type为空，获取失败")
    def test_get_template_03(self, Headers):
        result = self.client.GetTemplate("", Headers)
        assert result[0] == 404

    @pytest.mark.parametrize("template_type", ["python"])
    @allure.title("获取代码模板，template_type为有效值，获取成功")
    def test_get_template_04(self, template_type, Headers):
        result = self.client.GetTemplate(template_type, Headers)
        assert result[0] == 200
        assert result[1]["template_type"] == template_type
        assert "code_template" in result[1]
        assert isinstance(result[1]["code_template"], str)
