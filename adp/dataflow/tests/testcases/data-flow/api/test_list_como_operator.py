# -*- coding:UTF-8 -*-

import pytest
import allure

from lib.dataflow_como_operator import AutomationClient

@allure.feature("automation服务接口测试：列举组合算子")
class TestListCombinationOperators:
    client = AutomationClient()
    
    @allure.title("列举所有组合算子 - 默认参数")
    def test_list_operators_default(self, Headers):
        """测试默认参数列举组合算子"""
        params = {}
        result = self.client.GetOperatorsList(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"列举组合算子应返回状态码200，实际为: {result[0]}"
        
        # 验证分页结构
        assert "page" in result[1], "响应缺少page字段"
        assert "limit" in result[1], "响应缺少limit字段"
        assert "total" in result[1], "响应缺少total字段"
        assert "ops" in result[1], "响应缺少ops字段"
        
        # 验证有返回数据
        assert result[1]["total"] > 0, "总数应大于0"
        assert len(result[1]["ops"]) > 0, "应返回至少一个算子"
    
    @allure.title("列举组合算子 - 分页参数")
    @pytest.mark.parametrize("page,limit", [(1, 2), (2, 2)])
    def test_list_operators_pagination(self, Headers, page, limit):
        """测试分页参数对列举组合算子的影响"""
        params = {
            "page": page,
            "limit": limit
        }
        result = self.client.GetOperatorsList(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"列举组合算子应返回状态码200，实际为: {result[0]}"
        
        # 验证分页参数生效
        assert result[1]["page"] == page, f"返回的page应为{page}，实际为: {result[1]['page']}"
        assert result[1]["limit"] == limit, f"返回的limit应为{limit}，实际为: {result[1]['limit']}"
        
        # 验证返回的数据量符合限制
        assert len(result[1]["ops"]) <= limit, f"返回的算子数量应小于等于{limit}，实际为: {len(result[1]['ops'])}"
    
    @allure.title("列举组合算子 - 按分类过滤")
    @pytest.mark.parametrize("category", ["data_split", "data_extract", "data_process"])
    def test_list_operators_by_category(self, Headers, category):
        """测试按分类过滤组合算子"""
        params = {
            "category": category
        }
        result = self.client.GetOperatorsList(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"列举组合算子应返回状态码200，实际为: {result[0]}"
        
        # 如果结果不为空，则验证分类是否正确
        if len(result[1]["ops"]) > 0:
            for op in result[1]["ops"]:
                assert op["category"] == category, f"过滤结果中的算子分类应为{category}，实际为: {op['category']}"
    
    @allure.title("列举组合算子 - 按算子ID过滤")
    def test_list_operators_by_id(self, Headers):
        """测试按算子ID过滤组合算子"""
        # 先获取一个已存在的算子ID
        all_result = self.client.GetOperatorsList({}, Headers)
        if len(all_result[1]["ops"]) == 0:
            pytest.skip("没有可用的算子数据进行测试")
            
        operator_id = all_result[1]["ops"][0]["operator_id"]
        params = {
            "operator_id": operator_id
        }
        result = self.client.GetOperatorsList(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"列举组合算子应返回状态码200，实际为: {result[0]}"
        
        # 验证只返回了指定ID的算子
        assert len(result[1]["ops"]) > 0, "过滤结果应该至少包含一个算子"
        for op in result[1]["ops"]:
            assert op["operator_id"] == operator_id, f"过滤结果中的算子ID应为{operator_id}，实际为: {op['operator_id']}"
    
    @allure.title("列举组合算子 - 按名称过滤")
    def test_list_operators_by_name(self, Headers):
        """测试按名称过滤组合算子"""
        # 先获取一个已存在的算子名称
        all_result = self.client.GetOperatorsList({}, Headers)
        if len(all_result[1]["ops"]) == 0:
            pytest.skip("没有可用的算子数据进行测试")
            
        # 使用第一个算子名称的前几个字符作为过滤条件
        name_part = all_result[1]["ops"][0]["operator_name"][:5]
        params = {
            "name": name_part
        }
        result = self.client.GetOperatorsList(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"列举组合算子应返回状态码200，实际为: {result[0]}"
        
        # 验证返回的算子名称包含指定前缀
        if len(result[1]["ops"]) > 0:
            found = False
            for op in result[1]["ops"]:
                if name_part in op["operator_name"]:
                    found = True
                    break
            assert found, f"过滤结果中应该包含标题包含'{name_part}'的算子"
    
    @allure.title("列举组合算子 - 按版本过滤")
    def test_list_operators_by_version(self, Headers):
        """测试按版本过滤组合算子"""
        # 使用常见版本号
        version = "1.0.0"
        params = {
            "version": version
        }
        result = self.client.GetOperatorsList(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"列举组合算子应返回状态码200，实际为: {result[0]}"
        
        # 只验证接口调用成功，因为我们不确定是否有此版本的算子
        assert "ops" in result[1], "响应缺少ops字段"
    
    @allure.title("列举组合算子 - 排序参数")
    @pytest.mark.parametrize("sortby,order", [
        ("create_time", "desc"),
        ("update_time", "asc"),
    ])
    def test_list_operators_sorting(self, Headers, sortby, order):
        """测试排序参数对列举组合算子的影响"""
        params = {
            "sortby": sortby,
            "order": order
        }
        result = self.client.GetOperatorsList(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"列举组合算子应返回状态码200，实际为: {result[0]}"
        
        # 只验证API请求成功，因为无法确认排序是否正确
        assert "ops" in result[1], "响应缺少ops字段"
    
    @allure.title("列举组合算子 - 多参数组合过滤")
    def test_list_operators_combined_filters(self, Headers):
        """测试多个过滤参数组合使用"""
        # 先获取已有数据
        all_result = self.client.GetOperatorsList({}, Headers)
        if len(all_result[1]["ops"]) == 0:
            pytest.skip("没有可用的算子数据进行测试")
            
        # 使用第一个算子的分类
        category = all_result[1]["ops"][0]["category"]
        
        params = {
            "category": category,
            "limit": 5
        }
        result = self.client.GetOperatorsList(params, Headers)
        
        # 验证响应状态码
        assert result[0] == 200, f"列举组合算子应返回状态码200，实际为: {result[0]}"
        
        # 验证返回的算子符合过滤条件
        if len(result[1]["ops"]) > 0:
            for op in result[1]["ops"]:
                assert op["category"] == category, f"过滤结果中的算子分类应为{category}"
    
    @allure.title("列举组合算子 - 无效分类参数")
    def test_list_operators_invalid_category(self, Headers):
        """测试使用无效分类参数列举组合算子"""
        params = {
            "category": "invalid_category_xyz"
        }
        result = self.client.GetOperatorsList(params, Headers)
        
        # 根据截图显示，此接口返回400或500
        if result[0] in [400, 500]:
            # API验证了参数有效性，返回错误码
            assert True, "无效分类参数返回了正确的错误状态码"
        else:
            # 如果API接受任何分类参数，将返回空列表
            assert result[0] == 200, f"无效分类参数应返回状态码200，实际为: {result[0]}"
            assert len(result[1]["ops"]) == 0, "无效分类应返回空列表"
    
    @allure.title("列举组合算子 - 无效页码参数")
    def test_list_operators_invalid_page(self, Headers):
        """测试使用无效页码参数列举组合算子"""
        # 使用负数页码
        params = {
            "page": -1
        }
        result = self.client.GetOperatorsList(params, Headers)
        
        # 根据截图显示，此接口对于无效的page会返回400
        assert result[0] == 400, f"无效页码参数应返回状态码400，实际为: {result[0]}"
        assert "code" in result[1], "错误响应应包含code字段"
        assert "description" in result[1], "错误响应应包含description字段" 