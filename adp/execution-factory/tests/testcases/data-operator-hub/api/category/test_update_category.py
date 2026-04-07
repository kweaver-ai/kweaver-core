# -*- coding:UTF-8 -*-
"""
算子分类管理接口测试：更新算子分类

测试覆盖：
- 正常更新分类名称
- 分类不存在时的更新
- 分类名称重复校验
- 分类名称格式校验
"""

import allure
import pytest

from lib.operator_internal import InternalOperator


# 测试数据常量
class UpdateCategoryTestData:
    """更新分类测试数据常量"""
    # Setup 中创建的测试分类
    TEST_CATEGORY_TYPE = "update_type"
    TEST_CATEGORY_NAME = "待更新分类"
    UPDATED_NAME = "等待更新"
    
    # 不存在的分类类型
    NON_EXISTENT_TYPE = "update_type_01"
    
    # 已存在的分类名称（系统预置，不能重复）
    EXISTING_NAMES = ["未分类", "系统工具"]
    
    # 不合法的分类名称（与创建测试相同）
    INVALID_NAMES = [
        "invalid name", "name~", "name@", "name`", "name#", "name$", "name%", 
        "name^", "name&", "name*", "name()", "name-", "name+", "name=", 
        "name[]", "name{}", "name|", "name\\", "name:", "name;", "name'", 
        "name,", "name.", "name?", "name/", "name<", "name>",
        "name；", "name\"", "name：", "name'", "name【】", "name《", "name》", 
        "name？", "name·", "name、", "name，", "name。",
        "invalid_name:_more_then_50_characters_aaaaaaaaaaaaa"
    ]


@allure.feature("算子分类管理接口测试：更新算子分类")
class TestUpdateCategory:
    """测试更新算子分类功能"""
    
    client = InternalOperator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, UserHeaders):
        """测试前置准备：创建一个用于更新的测试分类"""
        category_data = {
            "category_type": UpdateCategoryTestData.TEST_CATEGORY_TYPE,
            "name": UpdateCategoryTestData.TEST_CATEGORY_NAME
        }
        result = self.client.CreateCategory(category_data, UserHeaders)
        assert result[0] == 200, f"Setup 失败：创建测试分类失败，状态码: {result[0]}"

    @allure.title("更新算子分类，传参正确，更新成功")
    @allure.description("验证使用合法的分类名称更新分类，应返回200状态码，且分类名称已更新")
    def test_update_category_01(self, UserHeaders):
        """测试正常更新分类名称场景"""
        # 执行更新操作
        update_data = {"name": UpdateCategoryTestData.UPDATED_NAME}
        result = self.client.UpdateCategory(
            UpdateCategoryTestData.TEST_CATEGORY_TYPE, 
            update_data, 
            UserHeaders
        )
        assert result[0] == 200, f"更新分类失败，状态码: {result[0]}"
        
        # 验证更新结果：获取分类列表，确认名称已更新
        result = self.client.GetCategory(UserHeaders)
        assert result[0] == 200, f"获取分类列表失败，状态码: {result[0]}"
        
        # 验证旧名称不存在，新名称存在
        category_list_str = str(result[1])
        assert UpdateCategoryTestData.TEST_CATEGORY_NAME not in category_list_str, \
            "旧分类名称仍存在于分类列表中"
        assert UpdateCategoryTestData.UPDATED_NAME in category_list_str, \
            "新分类名称未出现在分类列表中"

    @allure.title("更新算子分类，分类不存在，更新失败")
    @allure.description("验证更新不存在的分类，应返回404状态码")
    def test_update_category_02(self, UserHeaders):
        """测试更新不存在的分类"""
        update_data = {"name": "更新"}
        result = self.client.UpdateCategory(
            UpdateCategoryTestData.NON_EXISTENT_TYPE, 
            update_data, 
            UserHeaders
        )
        assert result[0] == 404, f"更新不存在的分类应返回404，但返回状态码: {result[0]}"

    @allure.title("更新算子分类，分类名称存在同名，更新失败")
    @allure.description("验证使用已存在的分类名称更新分类，应返回400状态码")
    def test_update_category_03(self, UserHeaders):
        """测试分类名称重复校验"""
        for existing_name in UpdateCategoryTestData.EXISTING_NAMES:
            update_data = {"name": existing_name}
            result = self.client.UpdateCategory(
                UpdateCategoryTestData.TEST_CATEGORY_TYPE, 
                update_data, 
                UserHeaders
            )
            assert result[0] == 400, \
                f"使用已存在的名称 '{existing_name}' 更新分类应失败，但返回状态码: {result[0]}"

    @allure.title("更新算子分类，分类名称不合法，更新失败")
    @allure.description("验证使用不合法的分类名称更新分类，应返回400状态码")
    @pytest.mark.parametrize("name", UpdateCategoryTestData.INVALID_NAMES)
    def test_update_category_04(self, name, UserHeaders):
        """测试分类名称格式校验"""
        update_data = {"name": name}
        result = self.client.UpdateCategory(
            UpdateCategoryTestData.TEST_CATEGORY_TYPE, 
            update_data, 
            UserHeaders
        )
        assert result[0] == 400, \
            f"使用不合法的名称 '{name}' 更新分类应失败，但返回状态码: {result[0]}"