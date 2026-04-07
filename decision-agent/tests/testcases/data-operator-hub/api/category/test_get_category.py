# -*- coding:UTF-8 -*-

import allure

from lib.operator_internal import InternalOperator

@allure.feature("MCP服务管理接口测试：获取算子分类")
class TestGetCategory:
    
    client = InternalOperator()

    @allure.title("获取算子分类，内置分类(other_category)默认排在最前面")
    def test_get_category_01(self, UserHeaders):
        category = {
            "category_type": "custom",
            "name": "自定义分类"
        }
        result = self.client.CreateCategory(category, UserHeaders)
        assert result[0] == 200
        result = self.client.GetCategory(UserHeaders)
        assert result[0] == 200
        assert result[1][0]["category_type"] == "other_category"
        assert result[1][1]["category_type"] == "system"

    @allure.title("获取算子分类，自定义分类按照更新时间降序排列")
    def test_get_category_02(self, UserHeaders):
        # 创建自定义分类
        for i in range(3):
            category = {
                "category_type": "custom_0" + str(i),
                "name": "自定义分类_0" + str(i)
            }
            result = self.client.CreateCategory(category, UserHeaders)
            assert result[0] == 200
        # 更新自定义分类
        result = self.client.UpdateCategory("custom_01", {"name": "测试分类"}, UserHeaders)
        assert result[0] == 200
        # 获取分类
        result = self.client.GetCategory(UserHeaders)
        assert result[0] == 200
        assert result[1][2]["category_type"] == "custom_01"
        assert result[1][3]["category_type"] == "custom_02"
        assert result[1][4]["category_type"] == "custom_00"