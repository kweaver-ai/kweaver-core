# -*- coding:UTF-8 -*-
"""
算子调试接口测试

测试目标：
    验证算子调试功能，包括不同HTTP方法（GET/POST/DELETE）的调试、不同参数传递方式、异常场景处理等。

测试覆盖：
    1. GET接口调试（path/query/header传参）
    2. POST接口调试（body传参）
    3. DELETE接口调试（body传参）
    4. 异常场景：算子不存在、缺少必填参数

说明：
    算子调试功能允许开发者在发布前测试算子的执行逻辑，支持不同HTTP方法和参数传递方式。
"""

import allure
import uuid
import pytest

from common.get_content import GetContent
from lib.operator import Operator

operator_list = []

@allure.feature("算子注册与管理接口测试：算子调试")
class TestOperatorDebug:
    """
    算子调试测试类
    
    说明：
        测试算子调试功能的各种场景，包括不同HTTP方法和参数传递方式的调试。
    """
    
    client = Operator()

    def _find_operator_by_name(self, name):
        """
        辅助方法：根据算子名称查找算子ID和版本
        
        参数：
            name: 算子名称
        
        返回：
            tuple: (operator_id, operator_version)，如果未找到返回(None, None)
        """
        global operator_list
        for operator in operator_list:
            if operator.get("name") == name:
                return operator["operator_id"], operator.get("version")
        return None, None

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        """
        测试前置准备
        
        功能：
            注册算子并获取未发布状态的算子列表，为后续测试准备数据
        
        说明：
            - 注册算子后，获取所有未发布状态的算子
            - operator_list用于后续测试用例查找特定算子
        """
        global operator_list

        file = GetContent("./config/env.ini")
        config = file.config()
        host = config["server"]["host"]
        filepath = "./resource/openapi/compliant/operator.json"
        api_data = GetContent(filepath).jsonfile()
        server = api_data["servers"][0]
        server["url"] = f"https://{host}"
        
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }

        result = self.client.RegisterOperator(data, Headers)
        
        # 处理服务器不可用错误（502, 503, 504）
        if result[0] in [502, 503, 504]:
            pytest.skip(f"服务器暂时不可用，状态码: {result[0]}, 跳过测试")
        
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}"
        
        operators = result[1]
        for operator in operators:
            assert operator.get("status") == "success", \
                f"算子注册状态应该是success，实际: {operator.get('status')}"

        # 获取未发布状态的算子列表
        data = {
            "page_size": -1,
            "status": "unpublish"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}"
        operator_list = result[1].get("data", [])
        
        if len(operator_list) == 0:
            print(f"警告: operator_list 为空，某些测试可能会失败")

    @allure.title("算子调试 - GET接口（path传参），调试成功")
    def test_operator_debug_01(self, Headers):
        """
        测试用例1：正常场景 - GET接口调试，使用path参数传递
        
        测试场景：
            - 调试"获取算子信息"算子
            - 通过path参数传递operator_id
        
        验证点：
            - 调试成功，返回200状态码
        
        说明：
            path参数是GET接口常见的传参方式，参数直接放在URL路径中。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, operator_version = self._find_operator_by_name("获取算子信息")
        if not operator_id:
            pytest.skip("未找到名称为'获取算子信息'的算子，跳过此测试")

        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers,
            "path": {
                "operator_id": operator_id
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 200, f"算子调试失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("算子调试 - GET接口（query传参），调试成功")
    def test_operator_debug_02(self, Headers):
        """
        测试用例2：正常场景 - GET接口调试，使用query参数传递
        
        测试场景：
            - 调试"获取算子列表"算子
            - 通过query参数传递分页和排序信息
        
        验证点：
            - 调试成功，返回200状态码
        
        说明：
            query参数通常用于GET请求的分页、排序、过滤等场景。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, operator_version = self._find_operator_by_name("获取算子列表")
        if not operator_id:
            pytest.skip("未找到名称为'获取算子列表'的算子，跳过此测试")

        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers,
            "query": {
                "page_size": "5",
                "sort_by": "name",
                "sort_order": "asc"
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 200, f"算子调试失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("算子调试 - GET接口（header传参），调试成功")
    def test_operator_debug_03(self, Headers):
        """
        测试用例3：正常场景 - GET接口调试，仅使用header参数
        
        测试场景：
            - 调试"获取算子分类"算子
            - 仅通过header传递认证信息，无需额外参数
        
        验证点：
            - 调试成功，返回200状态码
        
        说明：
            某些GET接口只需要认证信息，不需要额外的path或query参数。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, operator_version = self._find_operator_by_name("获取算子分类")
        if not operator_id:
            pytest.skip("未找到名称为'获取算子分类'的算子，跳过此测试")

        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 200, f"算子调试失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("算子调试 - POST接口（body传参），调试成功")
    def test_operator_debug_04(self, Headers):
        """
        测试用例4：正常场景 - POST接口调试，使用body参数传递
        
        测试场景：
            - 调试"更新算子信息"算子
            - 通过body传递完整的更新数据
        
        验证点：
            - 调试成功，返回200状态码
        
        说明：
            POST接口通常用于创建或更新操作，数据通过请求体传递。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, operator_version = self._find_operator_by_name("更新算子信息")
        if not operator_id:
            pytest.skip("未找到名称为'更新算子信息'的算子，跳过此测试")
        
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "data": str(api_data),
                "operator_metadata_type": "openapi"
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 200, f"算子调试失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("算子调试 - DELETE接口（body传参），调试成功")
    def test_operator_debug_05(self, Headers):
        """
        测试用例5：正常场景 - DELETE接口调试，使用body参数传递
        
        测试场景：
            - 调试"删除算子"算子
            - 通过body传递要删除的算子ID和版本信息
        
        验证点：
            - 调试成功，返回200状态码
        
        说明：
            DELETE接口用于删除操作，虽然HTTP规范建议使用path参数，但某些场景下也会使用body传递参数。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, operator_version = self._find_operator_by_name("删除算子")
        if not operator_id:
            pytest.skip("未找到名称为'删除算子'的算子，跳过此测试")

        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 200, f"算子调试失败，状态码: {result[0]}, 响应: {result}"

    @allure.title("算子调试 - 算子不存在，调试失败")
    def test_operator_debug_06(self, Headers):
        """
        测试用例6：异常场景 - 算子版本不存在
        
        测试场景：
            - 使用存在的算子ID但不存在版本号（随机UUID）
            - 调用算子调试接口
        
        验证点：
            - 接口返回404状态码（Not Found）
        
        说明：
            当算子ID存在但版本号不存在时，应该返回404错误，表示资源未找到。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, _ = self._find_operator_by_name("更新算子状态")
        if not operator_id:
            pytest.skip("未找到名称为'更新算子状态'的算子，跳过此测试")
        
        # 使用随机UUID作为不存在的版本号
        fake_version = str(uuid.uuid4())
                
        debug_data = {
            "operator_id": operator_id,
            "version": fake_version,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": fake_version,
                "status": "published"
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 404, f"不存在的版本号应该返回404，实际: {result[0]}, 响应: {result}"

    @allure.title("算子调试 - 缺少必填参数operator_id，调试失败")
    def test_operator_debug_07(self, Headers):
        """
        测试用例7：异常场景 - 缺少必填参数operator_id
        
        测试场景：
            - 调试数据中不包含operator_id字段
            - 调用算子调试接口
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            operator_id是算子调试的必填参数，缺少此参数会导致请求失败。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, operator_version = self._find_operator_by_name("更新算子状态")
        if not operator_id:
            pytest.skip("未找到名称为'更新算子状态'的算子，跳过此测试")
                
        # 不包含operator_id字段
        debug_data = {
            "version": operator_version,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "status": "published"
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 400, f"缺少operator_id应该返回400，实际: {result[0]}, 响应: {result}"

    @allure.title("算子调试 - 缺少必填参数version，调试失败")
    def test_operator_debug_08(self, Headers):
        """
        测试用例8：异常场景 - 缺少必填参数version
        
        测试场景：
            - 调试数据中不包含version字段
            - 调用算子调试接口
        
        验证点：
            - 接口返回400状态码（Bad Request）
        
        说明：
            version是算子调试的必填参数，缺少此参数会导致请求失败。
        """
        global operator_list
        
        if len(operator_list) == 0:
            pytest.skip("operator_list 为空，跳过此测试")
        
        operator_id, operator_version = self._find_operator_by_name("更新算子状态")
        if not operator_id:
            pytest.skip("未找到名称为'更新算子状态'的算子，跳过此测试")
                
        # 不包含version字段
        debug_data = {
            "operator_id": operator_id,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "status": "published"
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 400, f"缺少version应该返回400，实际: {result[0]}, 响应: {result}"
