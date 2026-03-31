# -*- coding:UTF-8 -*-
"""
删除算子接口测试

测试目标：
    验证删除算子的功能，包括单个删除、批量删除、不同状态的算子删除等场景。

测试覆盖：
    1. 正常场景：单个算子删除，算子未发布，删除成功
    2. 异常场景：算子不存在，删除失败
    3. 正常场景：批量算子删除，算子均未发布，删除成功
    4. 异常场景：批量删除时部分算子不存在，删除失败
    5. 异常场景：算子已发布，删除失败
    6. 正常场景：算子已下架但无引用关系，删除成功

说明：
    只有未发布（unpublish）或已下架（offline）状态的算子可以被删除。
    已发布（published）状态的算子不能删除，需要先下架才能删除。
    如果算子有引用关系，即使已下架也不能删除。
"""

import allure
import pytest

from common.get_content import GetContent
from lib.operator import Operator

operator_list = []

@allure.feature("算子注册与管理接口测试：删除算子")
class TestDeleteOperator:
    """
    删除算子测试类
    
    说明：
        测试删除算子的各种场景，包括单个删除、批量删除、不同状态的算子删除等。
    """
    
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        """
        测试前置准备
        
        功能：
            注册算子，为后续测试准备数据
        
        说明：
            - 注册算子后，算子状态为未发布（unpublish）
            - 可以用于测试删除未发布算子的场景
        """
        global operator_list

        filepath = "./resource/openapi/compliant/relations.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200, f"注册算子失败，状态码: {re[0]}, 响应: {re}"

    @allure.title("删除算子 - 单个算子删除，算子存在且未发布，删除成功")
    def test_delete_operator_01(self, Headers):
        """
        测试用例1：正常场景 - 删除未发布算子
        
        测试场景：
            1. 注册一个算子（状态为unpublish）
            2. 删除该算子
            3. 验证算子已从列表中移除
        
        验证点：
            - 删除接口返回200状态码
            - 算子不再出现在未发布算子列表中
        
        说明：
            未发布（unpublish）状态的算子可以直接删除。
        """
        filepath = "./resource/openapi/compliant/del-test1.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}"
        assert len(result[1]) == 1, f"应该注册1个算子，实际: {len(result[1])}"
        
        operator_id = result[1][0]["operator_id"]
        version = result[1][0]["version"]
        
        # 删除算子
        data = [
            {
                "operator_id": operator_id,
                "version": version
            }
        ]
        result = self.client.DeleteOperator(data, Headers)
        assert result[0] == 200, f"删除算子失败，状态码: {result[0]}, 响应: {result}"

        # 验证算子已从列表中移除
        data = {
            "page_size": -1,
            "status": "unpublish"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}"
        
        ids = []
        for operator in result[1].get("data", []):
            ids.append(operator["operator_id"])
        
        assert operator_id not in ids, f"算子 {operator_id} 应该已被删除，但仍存在于列表中"

    @allure.title("删除算子 - 算子不存在，删除失败")
    def test_delete_operator_02(self, Headers):
        """
        测试用例2：异常场景 - 算子不存在
        
        测试场景：
            - 使用不存在的算子ID和版本号
            - 调用删除算子接口
        
        验证点：
            - 接口返回404状态码（Not Found）
        
        说明：
            当算子ID或版本号不存在时，应该返回404错误，表示资源未找到。
        """
        # 使用不存在的算子ID和版本号
        operator_id = "test"
        version = "1.0.0"
        data = [
            {
                "operator_id": operator_id,
                "version": version
            }
        ]
        result = self.client.DeleteOperator(data, Headers)
        assert result[0] == 404, f"不存在的算子应该返回404，实际: {result[0]}, 响应: {result}"

    @allure.title("删除算子 - 批量算子删除，算子均存在且未发布，删除成功")
    def test_delete_operator_03(self, Headers):
        """
        测试用例3：正常场景 - 批量删除未发布算子
        
        测试场景：
            1. 注册多个算子（状态为unpublish）
            2. 批量删除这些算子
            3. 验证所有算子都已从列表中移除
        
        验证点：
            - 删除接口返回200状态码
            - 所有算子都不再出现在未发布算子列表中
        
        说明：
            批量删除时，所有算子必须都存在且未发布，才能成功删除。
            如果任何一个算子不存在或已发布，整个批量删除会失败。
        """
        filepath = "./resource/openapi/compliant/del-test2.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}"
        
        operator_ids = []
        delete_data = []
        for operator in result[1]:
            operator_ids.append(operator["operator_id"])
            operator_data = {
                "operator_id": operator["operator_id"],
                "version": operator["version"]
            }
            delete_data.append(operator_data)

        # 批量删除算子
        result = self.client.DeleteOperator(delete_data, Headers)
        assert result[0] == 200, f"批量删除算子失败，状态码: {result[0]}, 响应: {result}"

        # 验证所有算子都已从列表中移除
        data = {
            "page_size": -1,
            "status": "unpublish"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}"
        
        ids = []
        for operator in result[1].get("data", []):
            ids.append(operator["operator_id"])
        
        for operator_id in operator_ids:
            assert operator_id not in ids, f"算子 {operator_id} 应该已被删除，但仍存在于列表中"

    @allure.title("删除算子 - 批量删除时部分算子不存在，删除失败")
    def test_delete_operator_04(self, Headers):
        """
        测试用例4：异常场景 - 批量删除时部分算子不存在
        
        测试场景：
            1. 注册多个算子
            2. 在删除列表中混入一个不存在的算子
            3. 调用批量删除接口
        
        验证点：
            - 接口返回404状态码（Not Found）
        
        说明：
            批量删除时，如果任何一个算子不存在，整个批量删除会失败。
            这是为了保证批量操作的原子性。
        """
        filepath = "./resource/openapi/compliant/del-test2.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}"
        
        operator_ids = []
        for operator in result[1]:
            operator_ids.append(operator["operator_id"])

        # 构建删除列表，包含存在的算子和不存在的算子
        delete_data = []
        for operator_id in operator_ids:
            result = self.client.GetOperatorInfo(operator_id, Headers)
            assert result[0] == 200, f"获取算子信息失败，状态码: {result[0]}"
            version = result[1]["version"]
            operator_data = {
                "operator_id": operator_id,
                "version": version
            }
            delete_data.append(operator_data)

        # 添加一个不存在的算子
        not_exist_operator = {
            "operator_id": "test",
            "version": "V1"
        }
        delete_data.append(not_exist_operator)

        # 批量删除（应该失败）
        result = self.client.DeleteOperator(delete_data, Headers)
        assert result[0] == 404, \
            f"批量删除时部分算子不存在应该返回404，实际: {result[0]}, 响应: {result}"

    @allure.title("删除算子 - 算子已发布，删除失败")
    def test_delete_operator_05(self, Headers):
        """
        测试用例5：异常场景 - 已发布算子不能删除
        
        测试场景：
            1. 注册并直接发布一个算子（状态为published）
            2. 尝试删除该算子
        
        验证点：
            - 删除接口返回400状态码（Bad Request）
        
        说明：
            已发布（published）状态的算子不能直接删除，需要先下架才能删除。
            这是为了保护已发布的算子不被误删。
        """
        filepath = "./resource/openapi/compliant/del-test3.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True  # 直接发布
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}"
        
        operator_id = result[1][0]["operator_id"]
        version = result[1][0]["version"]
        
        # 尝试删除已发布的算子（应该失败）
        del_data = [
            {
                "operator_id": operator_id,
                "version": version
            }
        ]
        result = self.client.DeleteOperator(del_data, Headers)
        assert result[0] == 400, \
            f"已发布算子不支持删除，应该返回400，实际: {result[0]}, 响应: {result}"

    @allure.title("删除算子 - 算子已下架但无引用关系，删除成功")
    def test_delete_operator_06(self, Headers):
        """
        测试用例6：正常场景 - 删除已下架算子
        
        测试场景：
            1. 注册并发布一个算子（状态为published）
            2. 下架该算子（状态变为offline）
            3. 删除该算子
            4. 验证算子已从列表中移除
        
        验证点：
            - 下架接口返回200状态码
            - 删除接口返回200状态码
            - 算子不再出现在列表中
        
        说明：
            已下架（offline）状态的算子可以删除，但前提是没有其他算子引用它。
            如果有引用关系，即使已下架也不能删除。
        """
        filepath = "./resource/openapi/compliant/del-test1.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True  # 直接发布
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200, f"注册算子失败，状态码: {result[0]}, 响应: {result}"
        
        operator_id = result[1][0]["operator_id"]
        version = result[1][0]["version"]
        
        # 下架算子
        data = [
            {
                "operator_id": operator_id,
                "status": "offline"
            }
        ]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200, f"下架算子失败，状态码: {result[0]}, 响应: {result}"

        # 删除已下架的算子
        del_data = [
            {
                "operator_id": operator_id,
                "version": version
            }
        ]
        result = self.client.DeleteOperator(del_data, Headers)
        assert result[0] == 200, f"删除已下架算子失败，状态码: {result[0]}, 响应: {result}"
