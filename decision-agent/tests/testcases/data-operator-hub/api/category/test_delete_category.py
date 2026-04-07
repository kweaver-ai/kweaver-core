# -*- coding:UTF-8 -*-

import allure
import pytest

from lib.operator_internal import InternalOperator

@allure.feature("MCP服务管理接口测试：删除算子分类")
class TestDeleteCategory:
    
    client = InternalOperator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, UserHeaders):
        category = {
            "category_type": "delete_type",
            "name": "待删除分类"
        }
        result = self.client.CreateCategory(category, UserHeaders)
        assert result[0] == 200

    @allure.title("删除算子分类，分类不存在，删除失败")
    def test_delete_category_01(self, UserHeaders):
        result = self.client.DeleteCategory("delete_type_01", UserHeaders)
        assert result[0] == 404

    @allure.title("删除算子分类，分类存在，删除成功")
    def test_delete_category_02(self, UserHeaders):
        result = self.client.DeleteCategory("delete_type", UserHeaders)
        assert result[0] == 200

        result = self.client.DeleteCategory("delete_type", UserHeaders)
        assert result[0] == 404