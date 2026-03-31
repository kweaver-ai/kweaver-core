# -*- coding:UTF-8 -*-
"""
算子分类管理接口测试：删除算子分类

测试覆盖：
- 删除不存在的分类
- 删除存在的分类
- 删除后再次删除的校验
"""

import allure
import pytest

from lib.operator_internal import InternalOperator


# 测试数据常量
class DeleteCategoryTestData:
    """删除分类测试数据常量"""
    # Setup 中创建的测试分类
    TEST_CATEGORY_TYPE = "delete_type"
    TEST_CATEGORY_NAME = "待删除分类"
    
    # 不存在的分类类型
    NON_EXISTENT_TYPE = "delete_type_01"


@allure.feature("算子分类管理接口测试：删除算子分类")
class TestDeleteCategory:
    """测试删除算子分类功能"""
    
    client = InternalOperator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, UserHeaders):
        """测试前置准备：创建一个用于删除的测试分类"""
        category_data = {
            "category_type": DeleteCategoryTestData.TEST_CATEGORY_TYPE,
            "name": DeleteCategoryTestData.TEST_CATEGORY_NAME
        }
        result = self.client.CreateCategory(category_data, UserHeaders)
        assert result[0] == 200, f"Setup 失败：创建测试分类失败，状态码: {result[0]}"

    @allure.title("删除算子分类，分类不存在，删除失败")
    @allure.description("验证删除不存在的分类，应返回404状态码")
    def test_delete_category_01(self, UserHeaders):
        """测试删除不存在的分类"""
        result = self.client.DeleteCategory(
            DeleteCategoryTestData.NON_EXISTENT_TYPE, 
            UserHeaders
        )
        assert result[0] == 404, \
            f"删除不存在的分类应返回404，但返回状态码: {result[0]}"

    @allure.title("删除算子分类，分类存在，删除成功")
    @allure.description("验证删除存在的分类，应返回200状态码；删除后再次删除应返回404")
    def test_delete_category_02(self, UserHeaders):
        """测试删除存在的分类"""
        # 第一次删除：应成功
        result = self.client.DeleteCategory(
            DeleteCategoryTestData.TEST_CATEGORY_TYPE, 
            UserHeaders
        )
        assert result[0] == 200, \
            f"删除存在的分类应返回200，但返回状态码: {result[0]}"

        # 第二次删除：分类已不存在，应返回404
        result = self.client.DeleteCategory(
            DeleteCategoryTestData.TEST_CATEGORY_TYPE, 
            UserHeaders
        )
        assert result[0] == 404, \
            f"删除已删除的分类应返回404，但返回状态码: {result[0]}"