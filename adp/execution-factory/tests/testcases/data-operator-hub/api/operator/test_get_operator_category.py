# -*- coding:UTF-8 -*-
"""
获取算子分类接口测试

测试目标：
    验证获取算子分类列表的功能，确保返回所有系统预定义的分类。

测试覆盖：
    1. 正常场景：获取算子分类列表，获取成功，包含所有系统分类

说明：
    算子分类是系统预定义的分类类型，用于对算子进行分类管理。
    系统分类包括：other_category、data_process、data_transform、data_store、
    data_analysis、data_query、data_extract、data_split、model_train、system。
"""

import allure

from jsonschema import Draft7Validator

from common.get_content import GetContent
from lib.operator import Operator


@allure.feature("算子注册与管理接口测试：获取算子分类")
class TestGetOperatorCategory:
    """
    获取算子分类测试类
    
    说明：
        测试获取算子分类列表的功能，验证返回的分类是否完整。
    """
    
    client = Operator()
    operator_category_success = GetContent("./response/data-operator-hub/agent-operator-integration/operator_category_response_success.json").jsonfile()
    
    @allure.title("获取算子分类 - 获取成功，包含所有系统分类")
    def test_get_operator_category_01(self, Headers):
        """
        测试用例1：正常场景 - 获取算子分类列表
        
        测试场景：
            - 调用获取算子分类接口
            - 验证返回的分类列表
        
        验证点：
            - 接口返回200状态码
            - 响应格式符合schema规范
            - 包含所有系统预定义的分类类型
        
        说明：
            系统预定义的分类类型包括：
            - other_category: 其他分类
            - data_process: 数据处理
            - data_transform: 数据转换
            - data_store: 数据存储
            - data_analysis: 数据分析
            - data_query: 数据查询
            - data_extract: 数据提取
            - data_split: 数据拆分
            - model_train: 模型训练
            - system: 系统分类
        """
        result = self.client.GetOperatorCategory(Headers)
        
        # 验证接口调用成功
        assert result[0] == 200, f"获取算子分类失败，状态码: {result[0]}, 响应: {result}"
        
        # 验证响应格式符合schema规范
        validator = Draft7Validator(self.operator_category_success)
        assert validator.is_valid(result), "响应格式不符合schema规范"

        # 系统预定义的分类类型
        sys_categorys = [
            "other_category",    # 其他分类
            "data_process",      # 数据处理
            "data_transform",    # 数据转换
            "data_store",        # 数据存储
            "data_analysis",     # 数据分析
            "data_query",        # 数据查询
            "data_extract",      # 数据提取
            "data_split",        # 数据拆分
            "model_train",       # 模型训练
            "system"             # 系统分类
        ]
        
        # 提取返回的分类类型
        categorys = []
        for category in result[1]:
            categorys.append(category["category_type"])

        # 验证所有系统分类都存在
        for category in sys_categorys:
            assert category in categorys, \
                f"系统分类 '{category}' 应该存在于返回的分类列表中，实际返回的分类: {categorys}"
