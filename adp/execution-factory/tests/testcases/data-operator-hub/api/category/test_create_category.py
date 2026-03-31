# -*- coding:UTF-8 -*-
"""
算子分类管理接口测试：新建算子分类

测试覆盖：
- 正常创建分类
- 分类名称重复校验
- 分类类型重复校验
- 分类名称格式校验（特殊字符、长度限制等）
"""

import allure
import pytest

from lib.operator_internal import InternalOperator


# 测试数据常量
class CategoryTestData:
    """测试数据常量定义"""
    # 合法的分类类型和名称
    VALID_CATEGORY_TYPE = "debug"
    VALID_CATEGORY_NAME = "调试"
    
    # 已存在的分类名称（系统预置）
    EXISTING_NAMES = ["数据处理", "未分类", "系统工具"]
    
    # 已存在的分类类型（系统预置）
    EXISTING_CATEGORY_TYPES = ["data_process", "other_category", "system"]
    
    # 不合法的分类名称（特殊字符、中文标点、超长等）
    INVALID_NAMES = [
        # 英文特殊字符
        "invalid name", "name~", "name@", "name`", "name#", "name$", "name%", 
        "name^", "name&", "name*", "name()", "name-", "name+", "name=", 
        "name[]", "name{}", "name|", "name\\", "name:", "name;", "name'", 
        "name,", "name.", "name?", "name/", "name<", "name>",
        # 中文标点符号
        "name；", "name\"", "name：", "name'", "name【】", "name《", "name》", 
        "name？", "name·", "name、", "name，", "name。",
        # 超长名称（超过50个字符）
        "invalid_name:_more_then_50_characters_aaaaaaaaaaaaa"
    ]


@allure.feature("算子分类管理接口测试：新建算子分类")
class TestCreateCategory:
    """测试新建算子分类功能"""
    
    client = InternalOperator()

    @allure.title("新建算子分类，传参正确，新建成功")
    @allure.description("验证使用合法的分类类型和名称创建分类，应返回200状态码")
    def test_create_category_01(self, UserHeaders):
        """测试正常创建分类场景"""
        # 准备测试数据
        category_data = {
            "category_type": CategoryTestData.VALID_CATEGORY_TYPE,
            "name": CategoryTestData.VALID_CATEGORY_NAME
        }
        
        # 执行创建操作
        result = self.client.CreateCategory(category_data, UserHeaders)
        
        # 验证结果
        assert result[0] == 200, f"创建分类失败，状态码: {result[0]}"

    @allure.title("新建算子分类，分类名称已存在，新建失败")
    @allure.description("验证使用已存在的分类名称创建分类，应返回400状态码")
    @pytest.mark.parametrize("name", CategoryTestData.EXISTING_NAMES)
    def test_create_category_02(self, name, UserHeaders):
        """测试分类名称重复校验"""
        # 准备测试数据
        category_data = {
            "category_type": "process",
            "name": name
        }
        
        # 执行创建操作
        result = self.client.CreateCategory(category_data, UserHeaders)
        
        # 验证结果：应返回400（Bad Request）
        assert result[0] == 400, f"使用已存在的名称 '{name}' 创建分类应失败，但返回状态码: {result[0]}"

    @allure.title("新建算子分类，分类类型已存在，新建失败")
    @allure.description("验证使用已存在的分类类型创建分类，应返回400状态码")
    @pytest.mark.parametrize("category_type", CategoryTestData.EXISTING_CATEGORY_TYPES)
    def test_create_category_03(self, category_type, UserHeaders):
        """测试分类类型重复校验"""
        # 准备测试数据
        category_data = {
            "category_type": category_type,
            "name": "测试数据处理"
        }
        
        # 执行创建操作
        result = self.client.CreateCategory(category_data, UserHeaders)
        
        # 验证结果：应返回400（Bad Request）
        assert result[0] == 400, f"使用已存在的类型 '{category_type}' 创建分类应失败，但返回状态码: {result[0]}"

    @allure.title("新建算子分类，分类名称不合法，新建失败")
    @allure.description("验证使用不合法的分类名称（包含特殊字符、超长等）创建分类，应返回400状态码")
    @pytest.mark.parametrize("name", CategoryTestData.INVALID_NAMES)
    def test_create_category_04(self, name, UserHeaders):
        """测试分类名称格式校验"""
        # 准备测试数据
        category_data = {
            "category_type": "data_process",
            "name": name
        }
        
        # 执行创建操作
        result = self.client.CreateCategory(category_data, UserHeaders)
        
        # 验证结果：应返回400（Bad Request）
        assert result[0] == 400, f"使用不合法的名称 '{name}' 创建分类应失败，但返回状态码: {result[0]}"