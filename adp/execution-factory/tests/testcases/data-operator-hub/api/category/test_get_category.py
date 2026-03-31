# -*- coding:UTF-8 -*-
"""
算子分类管理接口测试：获取算子分类

测试覆盖：
- 内置分类的排序规则（other_category 默认排在最前面）
- 自定义分类的排序规则（按更新时间降序排列）
"""

import allure

from lib.operator_internal import InternalOperator


# 测试数据常量
class GetCategoryTestData:
    """获取分类测试数据常量"""
    # 内置分类类型（系统预置）
    BUILTIN_CATEGORY_OTHER = "other_category"
    BUILTIN_CATEGORY_SYSTEM = "system"
    
    # 自定义分类数据
    CUSTOM_CATEGORY_TYPE_PREFIX = "custom_"
    CUSTOM_CATEGORY_NAME_PREFIX = "自定义分类_"
    CUSTOM_CATEGORY_COUNT = 3


@allure.feature("算子分类管理接口测试：获取算子分类")
class TestGetCategory:
    """测试获取算子分类功能"""
    
    client = InternalOperator()

    @allure.title("获取算子分类，内置分类(other_category)默认排在最前面")
    @allure.description("验证获取分类列表时，内置分类 other_category 和 system 应排在最前面")
    def test_get_category_01(self, UserHeaders):
        """测试内置分类的排序规则"""
        # 创建一个自定义分类用于验证排序
        custom_category = {
            "category_type": GetCategoryTestData.CUSTOM_CATEGORY_TYPE_PREFIX + "test",
            "name": "自定义分类"
        }
        result = self.client.CreateCategory(custom_category, UserHeaders)
        assert result[0] == 200, f"创建自定义分类失败，状态码: {result[0]}"
        
        # 获取分类列表
        result = self.client.GetCategory(UserHeaders)
        assert result[0] == 200, f"获取分类列表失败，状态码: {result[0]}"
        
        categories = result[1]
        assert isinstance(categories, list), "分类列表应为列表类型"
        assert len(categories) > 0, "分类列表不应为空"
        
        # 验证内置分类排在最前面
        assert categories[0]["category_type"] == GetCategoryTestData.BUILTIN_CATEGORY_OTHER, \
            f"第一个分类应为 '{GetCategoryTestData.BUILTIN_CATEGORY_OTHER}'，实际为 '{categories[0]['category_type']}'"
        assert categories[1]["category_type"] == GetCategoryTestData.BUILTIN_CATEGORY_SYSTEM, \
            f"第二个分类应为 '{GetCategoryTestData.BUILTIN_CATEGORY_SYSTEM}'，实际为 '{categories[1]['category_type']}'"

    @allure.title("获取算子分类，自定义分类按照更新时间降序排列")
    @allure.description("验证获取分类列表时，自定义分类应按更新时间降序排列（最近更新的排在最前面）")
    def test_get_category_02(self, UserHeaders):
        """测试自定义分类的排序规则"""
        # 创建多个自定义分类
        created_categories = []
        for i in range(GetCategoryTestData.CUSTOM_CATEGORY_COUNT):
            category_data = {
                "category_type": f"{GetCategoryTestData.CUSTOM_CATEGORY_TYPE_PREFIX}{i:02d}",
                "name": f"{GetCategoryTestData.CUSTOM_CATEGORY_NAME_PREFIX}{i:02d}"
            }
            result = self.client.CreateCategory(category_data, UserHeaders)
            assert result[0] == 200, f"创建分类 '{category_data['category_type']}' 失败，状态码: {result[0]}"
            created_categories.append(category_data["category_type"])
        
        # 更新中间的分类（custom_01），使其更新时间变为最新
        update_category_type = f"{GetCategoryTestData.CUSTOM_CATEGORY_TYPE_PREFIX}01"
        update_data = {"name": "测试分类"}
        result = self.client.UpdateCategory(update_category_type, update_data, UserHeaders)
        assert result[0] == 200, f"更新分类 '{update_category_type}' 失败，状态码: {result[0]}"
        
        # 获取分类列表
        result = self.client.GetCategory(UserHeaders)
        assert result[0] == 200, f"获取分类列表失败，状态码: {result[0]}"
        
        categories = result[1]
        
        # 找到自定义分类在列表中的位置（跳过内置分类）
        # 假设前两个是内置分类，从索引2开始是自定义分类
        custom_category_start_index = 2
        
        # 验证排序：custom_01（最近更新）应排在最前面，然后是 custom_02，最后是 custom_00
        assert categories[custom_category_start_index]["category_type"] == f"{GetCategoryTestData.CUSTOM_CATEGORY_TYPE_PREFIX}01", \
            f"最近更新的分类应排在最前面，期望: '{GetCategoryTestData.CUSTOM_CATEGORY_TYPE_PREFIX}01'"
        assert categories[custom_category_start_index + 1]["category_type"] == f"{GetCategoryTestData.CUSTOM_CATEGORY_TYPE_PREFIX}02", \
            f"第二个自定义分类应为 '{GetCategoryTestData.CUSTOM_CATEGORY_TYPE_PREFIX}02'"
        assert categories[custom_category_start_index + 2]["category_type"] == f"{GetCategoryTestData.CUSTOM_CATEGORY_TYPE_PREFIX}00", \
            f"最早创建的分类应排在最后，期望: '{GetCategoryTestData.CUSTOM_CATEGORY_TYPE_PREFIX}00'"