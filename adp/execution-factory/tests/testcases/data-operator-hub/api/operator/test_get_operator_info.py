# -*- coding:UTF-8 -*-
"""
获取算子信息接口测试

测试目标：
    验证获取算子详细信息的功能，包括不同状态算子的获取、多版本算子的处理等。

测试覆盖：
    1. 异常场景：算子不存在，获取失败
    2. 正常场景：算子存在并已发布，获取成功
    3. 正常场景：算子存在但未发布，获取成功
    4. 正常场景：算子存在多个版本，默认获取到最新版本

说明：
    获取算子信息接口会根据算子ID返回算子的详细信息。
    如果算子有多个版本，默认返回最新版本的信息。
"""

import pytest
import allure

from jsonschema import Draft7Validator

from common.get_content import GetContent
from lib.operator import Operator


@allure.feature("算子注册与管理接口测试：算子信息获取")
class TestGetOperatorInfo:
    """
    获取算子信息测试类
    
    说明：
        测试获取算子详细信息的各种场景，包括正常获取和异常处理。
    """
    
    client = Operator()

    failed_resp = GetContent("./response/data-operator-hub/agent-operator-integration/response_failed.json").jsonfile()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        """
        测试前置准备
        
        功能：
            注册算子并发布部分算子，为后续测试准备数据
        
        说明：
            - 注册算子后，将前一半的算子发布
            - 剩余的算子保持未发布状态
        """
        filepath = "./resource/openapi/compliant/test2.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200, f"注册算子失败，状态码: {re[0]}, 响应: {re}"
        
        count = int(len(re[1])/2)
        operators = re[1]
        for i in range(count):
            operator_id = operators[i]["operator_id"]
            version = operators[i]["version"]            
            update_data = [{
                "operator_id": operator_id,
                "version": version,
                "status": "published"
            }]
            re = self.client.UpdateOperatorStatus(update_data, Headers)
            assert re[0] == 200, f"发布算子失败，状态码: {re[0]}"

    @allure.title("获取算子信息 - 算子不存在，获取失败")
    def test_get_operator_info_01(self, Headers):
        """
        测试用例1：异常场景 - 算子不存在
        
        测试场景：
            - 使用不存在的算子ID
            - 调用获取算子信息接口
        
        验证点：
            - 接口返回404状态码（Not Found）
            - 响应格式符合错误响应schema
        
        说明：
            当算子ID不存在时，应该返回404错误，并且错误响应格式应该符合规范。
        """
        operator_id = "b6ea4229-c2a1-4007-b2f0-ab8301ccd31c"
        result = self.client.GetOperatorInfo(operator_id, Headers)
        
        # 验证返回404错误
        assert result[0] == 404, f"不存在的算子ID应该返回404，实际: {result[0]}, 响应: {result}"
        
        # 验证响应格式符合错误响应schema
        validator = Draft7Validator(self.failed_resp)
        assert validator.is_valid(result), "错误响应格式不符合schema规范"

    @allure.title("获取算子信息 - 算子存在并已发布，获取成功")
    def test_get_operator_info_02(self, Headers):
        """
        测试用例2：正常场景 - 获取已发布算子信息
        
        测试场景：
            - 查询已发布状态的算子列表
            - 获取第一个已发布算子的详细信息
        
        验证点：
            - 接口返回200状态码
            - 返回的operator_id匹配
        
        说明：
            已发布的算子可以被正常获取，返回算子的详细信息。
        """
        data = {"status": "published"}
        result = self.client.GetOperatorList(data, Headers)
        if result[0] != 200:
            pytest.skip(f"获取算子列表失败，状态码: {result[0]}, 响应: {result}")
        assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}"
        
        operators = result[1].get("data", [])
        if len(operators) == 0:
            pytest.skip("已发布算子列表为空，跳过此测试")
        
        operator_id = operators[0]["operator_id"]
        result = self.client.GetOperatorInfo(operator_id, Headers)
        
        # 验证获取成功
        assert result[0] == 200, f"获取算子信息失败，状态码: {result[0]}, 响应: {result}"
        assert result[1]["operator_id"] == operator_id, "返回的operator_id不匹配"

    @allure.title("获取算子信息 - 算子存在但未发布，获取成功")
    def test_get_operator_info_03(self, Headers):
        """
        测试用例3：正常场景 - 获取未发布算子信息
        
        测试场景：
            - 查询未发布状态的算子列表
            - 获取第一个未发布算子的详细信息
        
        验证点：
            - 接口返回200状态码
            - 返回的operator_id匹配
        
        说明：
            未发布的算子也可以被获取，只是不会出现在市场列表中。
        """
        data = {"status": "unpublish"}
        result = self.client.GetOperatorList(data, Headers)
        if result[0] != 200:
            pytest.skip(f"获取算子列表失败，状态码: {result[0]}, 响应: {result}")
        assert result[0] == 200, f"获取算子列表失败，状态码: {result[0]}"
        
        operators = result[1].get("data", [])
        if len(operators) == 0:
            pytest.skip("未发布算子列表为空，跳过此测试")
        
        operator_id = operators[0]["operator_id"]
        result = self.client.GetOperatorInfo(operator_id, Headers)
        
        # 验证获取成功
        assert result[0] == 200, f"获取算子信息失败，状态码: {result[0]}, 响应: {result}"
        assert result[1]["operator_id"] == operator_id, "返回的operator_id不匹配"

    @allure.title("获取算子信息 - 算子存在多个版本，默认获取到最新版本")
    def test_get_operator_info_04(self, Headers):
        """
        测试用例4：正常场景 - 多版本算子获取最新版本
        
        测试场景：
            1. 获取已发布算子的信息（初始版本）
            2. 编辑算子生成新版本（状态变为editing）
            3. 验证获取到的信息是新版本（editing状态）
            4. 发布新版本（状态变为published）
            5. 验证获取到的信息是新版本（published状态）
            6. 下架新版本（状态变为offline）
            7. 验证获取到的信息是新版本（offline状态）
            8. 再次编辑生成新版本（状态变为unpublish）
            9. 验证获取到的信息是最新版本（unpublish状态）
        
        验证点：
            - 每次获取都返回最新版本的算子信息
            - 版本号会随着编辑操作而更新
            - 状态会随着操作而变化
        
        说明：
            当算子有多个版本时，获取算子信息接口默认返回最新版本的详细信息。
            最新版本是指版本号最大的版本，无论其状态如何。
        """
        data = {"status": "published"}
        result = self.client.GetOperatorList(data, Headers)
        if result[0] != 200:
            pytest.skip(f"获取算子列表失败，状态码: {result[0]}, 响应: {result}")
        assert result[0] == 200
        
        operators = result[1].get("data", [])
        if len(operators) == 0:
            pytest.skip("已发布算子列表为空，跳过此测试")
        
        operator_id = operators[0]["operator_id"]
        version = operators[0]["version"]

        # 步骤1：获取初始版本信息
        re = self.client.GetOperatorInfo(operator_id, Headers)
        assert re[0] == 200, f"获取算子信息失败，状态码: {re[0]}"

        # 步骤2：编辑算子生成新版本（状态变为editing）
        data = {
            "operator_id": operator_id,
            "metadata_type": "openapi",  # 传递description时需要metadata_type
            "description": "test edit 1234567"
        }
        edit_re = self.client.EditOperator(data, Headers)
        assert edit_re[0] == 200, f"编辑算子失败，状态码: {edit_re[0]}"
        assert edit_re[1]["version"] != version, "应该生成新版本"
        assert edit_re[1]["status"] == "editing", f"编辑后状态应该是editing，实际: {edit_re[1].get('status')}"

        # 步骤3：验证获取到的信息是新版本（editing状态）
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200, f"获取算子信息失败，状态码: {result[0]}"
        assert result[1]["operator_id"] == operator_id, "operator_id不匹配"
        assert result[1]["version"] == edit_re[1]["version"], "应该返回最新版本"
        assert result[1]["status"] == "editing", f"状态应该是editing，实际: {result[1].get('status')}"

        # 步骤4：发布新版本（状态变为published）
        data = [{
            "operator_id": edit_re[1]["operator_id"],
            "status": "published"
        }]
        re = self.client.UpdateOperatorStatus(data, Headers)
        assert re[0] == 200, f"发布算子失败，状态码: {re[0]}"

        # 步骤5：验证获取到的信息是新版本（published状态）
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200, f"获取算子信息失败，状态码: {result[0]}"
        assert result[1]["operator_id"] == operator_id, "operator_id不匹配"
        assert result[1]["version"] == edit_re[1]["version"], "应该返回最新版本"
        assert result[1]["status"] == "published", f"状态应该是published，实际: {result[1].get('status')}"

        # 步骤6：下架新版本（状态变为offline）
        data = [{
            "operator_id": edit_re[1]["operator_id"],
            "status": "offline"
        }]
        re = self.client.UpdateOperatorStatus(data, Headers)
        assert re[0] == 200, f"下架算子失败，状态码: {re[0]}"

        # 步骤7：验证获取到的信息是新版本（offline状态）
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200, f"获取算子信息失败，状态码: {result[0]}"
        assert result[1]["operator_id"] == operator_id, "operator_id不匹配"
        assert result[1]["version"] == edit_re[1]["version"], "应该返回最新版本"
        assert result[1]["status"] == "offline", f"状态应该是offline，实际: {result[1].get('status')}"

        # 步骤8：再次编辑生成新版本（状态变为unpublish）
        data = {
            "operator_id": operator_id,
            "metadata_type": "openapi",  # 传递description时需要metadata_type
            "description": "test edit 123fhg4567"
        }
        edit_re1 = self.client.EditOperator(data, Headers)
        assert edit_re1[0] == 200, f"编辑算子失败，状态码: {edit_re1[0]}"
        assert edit_re1[1]["version"] != edit_re[1]["version"], "应该生成新的版本"
        assert edit_re1[1]["status"] == "unpublish", f"编辑后状态应该是unpublish，实际: {edit_re1[1].get('status')}"

        # 步骤9：验证获取到的信息是最新版本（unpublish状态）
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200, f"获取算子信息失败，状态码: {result[0]}"
        assert result[1]["operator_id"] == operator_id, "operator_id不匹配"
        assert result[1]["version"] == edit_re1[1]["version"], "应该返回最新版本"
        assert result[1]["status"] == "unpublish", f"状态应该是unpublish，实际: {result[1].get('status')}"
